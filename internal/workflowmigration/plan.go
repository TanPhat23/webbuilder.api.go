package workflowmigration

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
)

// Input is an immutable snapshot of legacy records. Element project IDs must be
// derived from their pages, not from the workflow that references them.
type Input struct {
	Workflows []Workflow
	Links     []Link
	Elements  []Element
}
type Workflow struct {
	ID, ProjectID string
	Enabled       bool
	Handlers      json.RawMessage
}
type Link struct{ ID, WorkflowID, ElementID, Event string }
type Element struct{ ID, PageID, ProjectID string }
type Interaction struct {
	On      string `json:"on"`
	Enabled bool   `json:"enabled"`
	Once    *bool  `json:"once,omitempty"`
	Do      []Step `json:"do"`
}
type InteractionWrite struct {
	WorkflowID, ProjectID, PageID, ElementID, ID, LinkID string
	Interaction                                          Interaction
}
type FlowWrite struct {
	WorkflowID, ProjectID, PageID, Name string
	Do                                  []Step
}
type Issue struct{ WorkflowID, Reason string }
type Plan struct {
	Interactions []InteractionWrite
	Flows        []FlowWrite
	Issues       []Issue
}

func stableName(prefix, source string, bytes int) string {
	sum := sha256.Sum256([]byte(source))
	return prefix + hex.EncodeToString(sum[:bytes])
}

// BuildPlan never mutates input and never emits a partial workflow. Persistence
// must detect existing-name conflicts and apply each plan transactionally.
func BuildPlan(input Input) Plan {
	var plan Plan
	elements := map[string]Element{}
	for _, e := range input.Elements {
		elements[e.ID] = e
	}
	links := map[string][]Link{}
	for _, l := range input.Links {
		links[l.WorkflowID] = append(links[l.WorkflowID], l)
	}
	workflows := append([]Workflow(nil), input.Workflows...)
	sort.Slice(workflows, func(i, j int) bool { return workflows[i].ID < workflows[j].ID })
	seen := map[string]bool{}
	for _, w := range workflows {
		seen[w.ID] = true
		group := links[w.ID]
		sort.Slice(group, func(i, j int) bool { return group[i].ID < group[j].ID })
		if len(group) == 0 {
			plan.Issues = append(plan.Issues, Issue{w.ID, "workflow has no element links"})
			continue
		}
		steps, err := TranslateHandlers(w.Handlers)
		if err != nil {
			plan.Issues = append(plan.Issues, Issue{w.ID, err.Error()})
			continue
		}
		writes := make([]InteractionWrite, 0, len(group))
		pages := map[string]bool{}
		linkedElements := map[string]bool{}
		ids := map[string]bool{}
		for _, l := range group {
			e, exists := elements[l.ElementID]
			event, once, ok := TranslateEvent(l.Event)
			if !exists || e.PageID == "" || e.ProjectID != w.ProjectID {
				err = fmt.Errorf("link %s references an element outside the workflow project", l.ID)
				break
			}
			if !ok {
				err = fmt.Errorf("link %s has unsupported event %q", l.ID, l.Event)
				break
			}
			id := stableName("ix_", l.ID, 4)
			key := e.ID + "/" + id
			if ids[key] {
				err = fmt.Errorf("duplicate interaction id for link %s", l.ID)
				break
			}
			ids[key] = true
			pages[e.PageID] = true
			linkedElements[e.ID] = true
			writes = append(writes, InteractionWrite{WorkflowID: w.ID, ProjectID: w.ProjectID, PageID: e.PageID, ElementID: e.ID, ID: id, LinkID: l.ID, Interaction: Interaction{On: event, Enabled: w.Enabled, Once: once, Do: steps}})
		}
		if err != nil {
			plan.Issues = append(plan.Issues, Issue{w.ID, err.Error()})
			continue
		}
		if len(linkedElements) > 1 {
			name := stableName("flow_", w.ID, 16)
			page := ""
			if len(pages) == 1 {
				for id := range pages {
					page = id
				}
			}
			plan.Flows = append(plan.Flows, FlowWrite{WorkflowID: w.ID, ProjectID: w.ProjectID, PageID: page, Name: name, Do: steps})
			for i := range writes {
				writes[i].Interaction.Do = []Step{{"run": name}}
			}
		}
		plan.Interactions = append(plan.Interactions, writes...)
	}
	for workflowID := range links {
		if !seen[workflowID] {
			plan.Issues = append(plan.Issues, Issue{workflowID, "linked workflow does not exist"})
		}
	}
	sort.Slice(plan.Issues, func(i, j int) bool { return plan.Issues[i].WorkflowID < plan.Issues[j].WorkflowID })
	return plan
}
