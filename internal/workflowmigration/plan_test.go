package workflowmigration

import (
	"encoding/json"
	"testing"
)

func TestPlanSharedWorkflowUsesPageOrProjectFlow(t *testing.T) {
	handlers := json.RawMessage(`[{"config":{"type":"showNotification","message":"Saved"}}]`)
	for _, tc := range []struct {
		name       string
		secondPage string
		scope      string
	}{{"same page", "p1", "p1"}, {"different pages", "p2", ""}} {
		t.Run(tc.name, func(t *testing.T) {
			input := Input{Workflows: []Workflow{{ID: "w1", ProjectID: "project", Enabled: true, Handlers: handlers}}, Links: []Link{{ID: "l1", WorkflowID: "w1", ElementID: "e1", Event: "onClick"}, {ID: "l2", WorkflowID: "w1", ElementID: "e2", Event: "onScroll"}}, Elements: []Element{{ID: "e1", PageID: "p1", ProjectID: "project"}, {ID: "e2", PageID: tc.secondPage, ProjectID: "project"}}}
			plan := BuildPlan(input)
			if len(plan.Issues) != 0 || len(plan.Flows) != 1 || len(plan.Interactions) != 2 {
				t.Fatalf("unexpected plan: %#v", plan)
			}
			if plan.Flows[0].PageID != tc.scope || plan.Interactions[0].Interaction.Do[0]["run"] != plan.Flows[0].Name || plan.Interactions[1].Interaction.Once == nil || *plan.Interactions[1].Interaction.Once {
				t.Fatalf("wrong scope or interaction: %#v", plan)
			}
		})
	}
}

func TestPlanRejectsWholeWorkflowForUntranslatableLink(t *testing.T) {
	input := Input{Workflows: []Workflow{{ID: "w", ProjectID: "p", Enabled: true, Handlers: json.RawMessage(`[{"config":{"type":"resetForm","formElementId":"f"}}]`)}}, Links: []Link{{ID: "a", WorkflowID: "w", ElementID: "e", Event: "onClick"}, {ID: "b", WorkflowID: "w", ElementID: "e", Event: "onUnknown"}}, Elements: []Element{{ID: "e", PageID: "page", ProjectID: "p"}}}
	plan := BuildPlan(input)
	if len(plan.Issues) != 1 || plan.Issues[0].WorkflowID != "w" || len(plan.Interactions) != 0 || len(plan.Flows) != 0 {
		t.Fatalf("unsafe partial plan: %#v", plan)
	}
}

func TestPlanSingleLinkAndDisabledWorkflow(t *testing.T) {
	input := Input{Workflows: []Workflow{{ID: "w", ProjectID: "p", Enabled: false, Handlers: json.RawMessage(`[{"config":{"type":"resetForm","formElementId":"f"}}]`)}}, Links: []Link{{ID: "a", WorkflowID: "w", ElementID: "e", Event: "onClick"}}, Elements: []Element{{ID: "e", PageID: "page", ProjectID: "p"}}}
	plan := BuildPlan(input)
	if len(plan.Issues) != 0 || len(plan.Flows) != 0 || len(plan.Interactions) != 1 || plan.Interactions[0].Interaction.Enabled || plan.Interactions[0].Interaction.Do[0]["reset"] != "f" {
		t.Fatalf("unexpected plan: %#v", plan)
	}
}

func TestMultipleEventsOnOneElementDoNotCreateSharedFlow(t *testing.T) {
	input := Input{Workflows: []Workflow{{ID: "w", ProjectID: "p", Enabled: true, Handlers: json.RawMessage(`[{"config":{"type":"resetForm","formElementId":"f"}}]`)}}, Links: []Link{{ID: "a", WorkflowID: "w", ElementID: "e", Event: "onClick"}, {ID: "b", WorkflowID: "w", ElementID: "e", Event: "onFocus"}}, Elements: []Element{{ID: "e", PageID: "page", ProjectID: "p"}}}
	plan := BuildPlan(input)
	if len(plan.Issues) != 0 || len(plan.Flows) != 0 || len(plan.Interactions) != 2 {
		t.Fatalf("unexpected shared flow: %#v", plan)
	}
}

func TestPlanReportsTranslatedStepsRejectedByLogicSchema(t *testing.T) {
	input := Input{Workflows: []Workflow{{ID: "w", ProjectID: "p", Enabled: true, Handlers: json.RawMessage(`[{"config":{"type":"addClass","elementId":"e","className":"two classes"}}]`)}}, Links: []Link{{ID: "a", WorkflowID: "w", ElementID: "e", Event: "onClick"}}, Elements: []Element{{ID: "e", PageID: "page", ProjectID: "p"}}}
	plan := BuildPlan(input)
	if len(plan.Issues) != 1 || len(plan.Interactions) != 0 {
		t.Fatalf("invalid translated step accepted: %#v", plan)
	}
}

func TestPlanReportsUnlinkedWorkflow(t *testing.T) {
	plan := BuildPlan(Input{Workflows: []Workflow{{ID: "orphan", ProjectID: "p", Enabled: true, Handlers: json.RawMessage(`[]`)}}})
	if len(plan.Issues) != 1 || plan.Issues[0].WorkflowID != "orphan" || len(plan.Interactions) != 0 {
		t.Fatalf("unlinked workflow omitted: %#v", plan)
	}
}
