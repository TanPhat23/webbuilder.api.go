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

type migrationSnapshot struct {
	input        Input
	settings     map[string]json.RawMessage
	projectLogic map[string]json.RawMessage
	pageLogic    map[string]json.RawMessage
	pageProject  map[string]string
}

func planInTx(ctx context.Context, tx *sql.Tx) (Plan, error) {
	snap, err := loadMigrationSnapshot(ctx, tx)
	if err != nil {
		return Plan{}, err
	}
	plan := BuildPlan(snap.input)
	failures := map[string]string{}
	completed := markCompletedWorkflows(plan.Interactions, snap.settings, failures)
	validateInteractionTargets(plan.Interactions, snap.settings, completed, failures)
	validateFlowTargets(plan.Flows, snap, completed, failures)
	return finalizePlan(plan, completed, failures), nil
}

func queryTxRows(ctx context.Context, tx *sql.Tx, q string, scan func(*sql.Rows) error) error {
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

func loadMigrationSnapshot(ctx context.Context, tx *sql.Tx) (*migrationSnapshot, error) {
	s := &migrationSnapshot{
		settings:     map[string]json.RawMessage{},
		projectLogic: map[string]json.RawMessage{},
		pageLogic:    map[string]json.RawMessage{},
		pageProject:  map[string]string{},
	}
	if err := queryTxRows(ctx, tx, `SELECT "Id", "Logic" FROM "Project"`, func(r *sql.Rows) error {
		var id string
		var raw []byte
		if err := r.Scan(&id, &raw); err != nil {
			return err
		}
		s.projectLogic[id] = raw
		return nil
	}); err != nil {
		return nil, fmt.Errorf("read projects: %w", err)
	}
	if err := queryTxRows(ctx, tx, `SELECT "Id", "ProjectId", "Logic" FROM "Page"`, func(r *sql.Rows) error {
		var id, p string
		var raw []byte
		if err := r.Scan(&id, &p, &raw); err != nil {
			return err
		}
		s.pageLogic[id] = raw
		s.pageProject[id] = p
		return nil
	}); err != nil {
		return nil, fmt.Errorf("read pages: %w", err)
	}
	if err := queryTxRows(ctx, tx, `SELECT "Id", "PageId", "Settings" FROM "Element"`, func(r *sql.Rows) error {
		var e Element
		var p sql.NullString
		var raw []byte
		if err := r.Scan(&e.ID, &p, &raw); err != nil {
			return err
		}
		e.PageID = p.String
		e.ProjectID = s.pageProject[e.PageID]
		s.input.Elements = append(s.input.Elements, e)
		s.settings[e.ID] = raw
		return nil
	}); err != nil {
		return nil, fmt.Errorf("read elements: %w", err)
	}
	if err := queryTxRows(ctx, tx, `SELECT "Id", "ProjectId", "Enabled", "Handlers" FROM "EventWorkflow"`, func(r *sql.Rows) error {
		var w Workflow
		if err := r.Scan(&w.ID, &w.ProjectID, &w.Enabled, &w.Handlers); err != nil {
			return err
		}
		s.input.Workflows = append(s.input.Workflows, w)
		return nil
	}); err != nil {
		return nil, fmt.Errorf("read workflows: %w", err)
	}
	if err := queryTxRows(ctx, tx, `SELECT "Id", "WorkflowId", "ElementId", "EventName" FROM "ElementEventWorkflow"`, func(r *sql.Rows) error {
		var l Link
		if err := r.Scan(&l.ID, &l.WorkflowID, &l.ElementID, &l.Event); err != nil {
			return err
		}
		s.input.Links = append(s.input.Links, l)
		return nil
	}); err != nil {
		return nil, fmt.Errorf("read workflow links: %w", err)
	}
	return s, nil
}

func markCompletedWorkflows(writes []InteractionWrite, settings map[string]json.RawMessage, failures map[string]string) map[string]bool {
	total := map[string]int{}
	marked := map[string]int{}
	for _, write := range writes {
		total[write.WorkflowID]++
		checkCompletedWriteMark(write, settings[write.ElementID], marked, failures)
	}
	completed := map[string]bool{}
	for id, count := range marked {
		if count == total[id] {
			completed[id] = true
		} else {
			failures[id] = "workflow has only partially migrated links"
		}
	}
	return completed
}

func checkCompletedWriteMark(write InteractionWrite, raw json.RawMessage, marked map[string]int, failures map[string]string) {
	m, err := object(raw)
	if err != nil {
		failures[write.WorkflowID] = "element settings are not an object"
		return
	}
	value, exists := m["migratedWorkflowLinks"]
	if !exists {
		return
	}
	marks, ok := value.(map[string]any)
	if !ok {
		failures[write.WorkflowID] = "migration markers are not an object"
		return
	}
	if mark, exists := marks[write.LinkID]; exists {
		if mark != true {
			failures[write.WorkflowID] = "invalid migration marker"
			return
		}
		marked[write.WorkflowID]++
	}
}

func validateInteractionTargets(writes []InteractionWrite, settings map[string]json.RawMessage, completed map[string]bool, failures map[string]string) {
	planned := map[string]string{}
	for _, write := range writes {
		if completed[write.WorkflowID] {
			continue
		}
		recordPlannedInteraction(write, planned, failures)
		validateExistingElementInteraction(write, settings[write.ElementID], failures)
	}
}

func recordPlannedInteraction(write InteractionWrite, planned, failures map[string]string) {
	key := write.ElementID + "\x00" + write.ID
	if other, exists := planned[key]; exists && other != write.WorkflowID {
		reason := fmt.Sprintf("multiple workflows generate interaction %s on element %s", write.ID, write.ElementID)
		failures[other], failures[write.WorkflowID] = reason, reason
		return
	}
	planned[key] = write.WorkflowID
}

func validateExistingElementInteraction(write InteractionWrite, raw json.RawMessage, failures map[string]string) {
	m, err := object(raw)
	if err != nil {
		failures[write.WorkflowID] = "element settings are not an object"
		return
	}
	value, exists := m["interactions"]
	if !exists {
		return
	}
	existing, ok := value.(map[string]any)
	if !ok {
		failures[write.WorkflowID] = "element interactions are not an object"
		return
	}
	if _, exists := existing[write.ID]; exists {
		failures[write.WorkflowID] = fmt.Sprintf("element %s already has interaction %s", write.ElementID, write.ID)
	}
}

func validateFlowTargets(flows []FlowWrite, snap *migrationSnapshot, completed map[string]bool, failures map[string]string) {
	for _, flow := range flows {
		if completed[flow.WorkflowID] {
			continue
		}
		checkFlowLocation(flow, snap.projectLogic[flow.ProjectID], "project", failures)
		for page, project := range snap.pageProject {
			if project == flow.ProjectID {
				checkFlowLocation(flow, snap.pageLogic[page], "page "+page, failures)
			}
		}
	}
}

func checkFlowLocation(flow FlowWrite, raw json.RawMessage, location string, failures map[string]string) {
	m, err := object(raw)
	if err != nil {
		failures[flow.WorkflowID] = location + " logic is not an object"
		return
	}
	if err := validateExistingLocationLogic(m, raw, location); err != nil {
		failures[flow.WorkflowID] = err.Error()
		return
	}
	flows, err := extractLocationFlows(m, location)
	if err != nil {
		failures[flow.WorkflowID] = err.Error()
		return
	}
	if _, exists := flows[flow.Name]; exists {
		failures[flow.WorkflowID] = fmt.Sprintf("flow %s already exists in %s", flow.Name, location)
	}
}

func validateExistingLocationLogic(m map[string]any, raw json.RawMessage, location string) error {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	if err := logic.Validate(logic.Document(m)); err != nil {
		return fmt.Errorf("%s logic is invalid: %w", location, err)
	}
	return nil
}

func extractLocationFlows(m map[string]any, location string) (map[string]any, error) {
	value, exists := m["flows"]
	if !exists || value == nil {
		return nil, nil
	}
	flows, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s flows are not an object", location)
	}
	return flows, nil
}

func finalizePlan(plan Plan, completed map[string]bool, failures map[string]string) Plan {
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
	return plan
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
