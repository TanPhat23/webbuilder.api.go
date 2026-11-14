package workflowmigration

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"

	"my-go-app/internal/realtime/domain/logic"
)

// DryRun reads one consistent database snapshot. The transaction is explicitly
// read-only; callers must explicitly opt into any future write operation.
func DryRun(ctx context.Context, db *sql.DB) (Plan, error) {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return Plan{}, err
	}
	defer tx.Rollback()
	plan, err := planInTx(ctx, tx)
	if err != nil {
		return Plan{}, err
	}
	if err := tx.Commit(); err != nil {
		return Plan{}, err
	}
	return plan, nil
}

func planInTx(ctx context.Context, tx *sql.Tx) (Plan, error) {
	var err error
	input := Input{}
	settings := map[string]json.RawMessage{}
	projectLogic := map[string]json.RawMessage{}
	pageLogic := map[string]json.RawMessage{}
	pageProject := map[string]string{}
	query := func(q string, scan func(*sql.Rows) error) error {
		rows, err := tx.QueryContext(ctx, q)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			if err := scan(rows); err != nil {
				return err
			}
		}
		return rows.Err()
	}
	if err = query(`SELECT "Id", "Logic" FROM "Project"`, func(r *sql.Rows) error {
		var id string
		var raw []byte
		if err := r.Scan(&id, &raw); err != nil {
			return err
		}
		projectLogic[id] = raw
		return nil
	}); err != nil {
		return Plan{}, fmt.Errorf("read projects: %w", err)
	}
	if err = query(`SELECT "Id", "ProjectId", "Logic" FROM "Page"`, func(r *sql.Rows) error {
		var id, p string
		var raw []byte
		if err := r.Scan(&id, &p, &raw); err != nil {
			return err
		}
		pageLogic[id] = raw
		pageProject[id] = p
		return nil
	}); err != nil {
		return Plan{}, fmt.Errorf("read pages: %w", err)
	}
	if err = query(`SELECT "Id", "PageId", "Settings" FROM "Element"`, func(r *sql.Rows) error {
		var e Element
		var p sql.NullString
		var raw []byte
		if err := r.Scan(&e.ID, &p, &raw); err != nil {
			return err
		}
		e.PageID = p.String
		e.ProjectID = pageProject[e.PageID]
		input.Elements = append(input.Elements, e)
		settings[e.ID] = raw
		return nil
	}); err != nil {
		return Plan{}, fmt.Errorf("read elements: %w", err)
	}
	if err = query(`SELECT "Id", "ProjectId", "Enabled", "Handlers" FROM "EventWorkflow"`, func(r *sql.Rows) error {
		var w Workflow
		if err := r.Scan(&w.ID, &w.ProjectID, &w.Enabled, &w.Handlers); err != nil {
			return err
		}
		input.Workflows = append(input.Workflows, w)
		return nil
	}); err != nil {
		return Plan{}, fmt.Errorf("read workflows: %w", err)
	}
	if err = query(`SELECT "Id", "WorkflowId", "ElementId", "EventName" FROM "ElementEventWorkflow"`, func(r *sql.Rows) error {
		var l Link
		if err := r.Scan(&l.ID, &l.WorkflowID, &l.ElementID, &l.Event); err != nil {
			return err
		}
		input.Links = append(input.Links, l)
		return nil
	}); err != nil {
		return Plan{}, fmt.Errorf("read workflow links: %w", err)
	}
	plan := BuildPlan(input)
	failures := map[string]string{}
	total := map[string]int{}
	marked := map[string]int{}
	for _, write := range plan.Interactions {
		total[write.WorkflowID]++
		m, err := object(settings[write.ElementID])
		if err != nil {
			failures[write.WorkflowID] = "element settings are not an object"
			continue
		}
		if value, exists := m["migratedWorkflowLinks"]; exists {
			marks, ok := value.(map[string]any)
			if !ok {
				failures[write.WorkflowID] = "migration markers are not an object"
				continue
			}
			if mark, exists := marks[write.LinkID]; exists {
				if mark != true {
					failures[write.WorkflowID] = "invalid migration marker"
					continue
				}
				marked[write.WorkflowID]++
			}
		}
	}
	completed := map[string]bool{}
	for id, count := range marked {
		if count == total[id] {
			completed[id] = true
		} else {
			failures[id] = "workflow has only partially migrated links"
		}
	}
	plannedInteractions := map[string]string{}
	for _, write := range plan.Interactions {
		if completed[write.WorkflowID] {
			continue
		}
		key := write.ElementID + "\x00" + write.ID
		if other, exists := plannedInteractions[key]; exists && other != write.WorkflowID {
			reason := fmt.Sprintf("multiple workflows generate interaction %s on element %s", write.ID, write.ElementID)
			failures[other], failures[write.WorkflowID] = reason, reason
		} else {
			plannedInteractions[key] = write.WorkflowID
		}
		m, err := object(settings[write.ElementID])
		if err != nil {
			failures[write.WorkflowID] = "element settings are not an object"
			continue
		}
		if value, exists := m["interactions"]; exists {
			existing, ok := value.(map[string]any)
			if !ok {
				failures[write.WorkflowID] = "element interactions are not an object"
				continue
			}
			if _, exists := existing[write.ID]; exists {
				failures[write.WorkflowID] = fmt.Sprintf("element %s already has interaction %s", write.ElementID, write.ID)
			}
		}
	}
	for _, flow := range plan.Flows {
		if completed[flow.WorkflowID] {
			continue
		}
		check := func(raw json.RawMessage, location string) {
			m, err := object(raw)
			if err != nil {
				failures[flow.WorkflowID] = location + " logic is not an object"
				return
			}
			if len(raw) > 0 && string(raw) != "null" {
				if err := logic.Validate(logic.Document(m)); err != nil {
					failures[flow.WorkflowID] = location + " logic is invalid: " + err.Error()
					return
				}
			}
			value, exists := m["flows"]
			if !exists || value == nil {
				return
			}
			flows, ok := value.(map[string]any)
			if !ok {
				failures[flow.WorkflowID] = location + " flows are not an object"
				return
			}
			if _, exists := flows[flow.Name]; exists {
				failures[flow.WorkflowID] = fmt.Sprintf("flow %s already exists in %s", flow.Name, location)
			}
		}
		check(projectLogic[flow.ProjectID], "project")
		for page, project := range pageProject {
			if project == flow.ProjectID {
				check(pageLogic[page], "page "+page)
			}
		}
	}
	interactions := plan.Interactions[:0]
	for _, w := range plan.Interactions {
		if _, bad := failures[w.WorkflowID]; !bad && !completed[w.WorkflowID] {
			interactions = append(interactions, w)
		}
	}
	plan.Interactions = interactions
	flows := plan.Flows[:0]
	for _, w := range plan.Flows {
		if _, bad := failures[w.WorkflowID]; !bad && !completed[w.WorkflowID] {
			flows = append(flows, w)
		}
	}
	plan.Flows = flows
	for id, reason := range failures {
		plan.Issues = append(plan.Issues, Issue{WorkflowID: id, Reason: reason})
	}
	sort.Slice(plan.Issues, func(i, j int) bool { return plan.Issues[i].WorkflowID < plan.Issues[j].WorkflowID })
	return plan, nil
}
func object(raw json.RawMessage) (map[string]any, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return map[string]any{}, nil
	}
	var value map[string]any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, err
	}
	if value == nil {
		return nil, fmt.Errorf("expected object")
	}
	return value, nil
}
