package workflowmigration

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func dryrunFakeIssue(plan Plan, workflow string) string {
	for _, i := range plan.Issues {
		if i.WorkflowID == workflow {
			return i.Reason
		}
	}
	return ""
}

func TestDryRunFakeCleanPlanIsReadOnly(t *testing.T) {
	f := applyFakeSeed()
	plan, err := DryRun(context.Background(), f.db(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Issues) != 0 || len(plan.Interactions) != 2 || len(plan.Flows) != 1 {
		t.Fatalf("unexpected plan: %#v", plan)
	}
	if !f.beginOpts.ReadOnly || f.commits != 1 {
		t.Fatalf("dry run must be a read-only tx that commits: %#v commits=%d", f.beginOpts, f.commits)
	}
	if plan.Interactions[0].ProjectID != "p" || plan.Interactions[0].PageID != "page" {
		t.Fatalf("element project must come from its page: %#v", plan.Interactions[0])
	}
	for _, q := range f.execs {
		t.Fatalf("dry run executed %s", q)
	}
}

func TestDryRunFakeAfterApplyReportsNothingToDo(t *testing.T) {
	f := applyFakeSeed()
	db := f.db(t)
	if _, err := Apply(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	plan, err := DryRun(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Issues) != 0 || len(plan.Interactions) != 0 || len(plan.Flows) != 0 {
		t.Fatalf("migrated data must be treated as complete: %#v", plan)
	}
}

func TestDryRunFakeErrorPropagation(t *testing.T) {
	boom := errors.New("boom")
	tables := map[string]string{
		"projects":  `FROM "Project"`,
		"pages":     `FROM "Page"`,
		"elements":  `FROM "Element"`,
		"workflows": `FROM "EventWorkflow"`,
		"links":     `FROM "ElementEventWorkflow"`,
	}
	for what, sub := range tables {
		t.Run("query "+what, func(t *testing.T) {
			f := applyFakeSeed()
			f.failQuery = map[string]error{sub: boom}
			_, err := DryRun(context.Background(), f.db(t))
			if err == nil || !strings.HasPrefix(err.Error(), "read ") || !errors.Is(err, boom) {
				t.Fatalf("err = %v", err)
			}
		})
		t.Run("scan "+what, func(t *testing.T) {
			f := applyFakeSeed()
			f.nullFirst = map[string]bool{sub: true}
			_, err := DryRun(context.Background(), f.db(t))
			if err == nil || !strings.HasPrefix(err.Error(), "read ") {
				t.Fatalf("err = %v", err)
			}
		})
		t.Run("rows "+what, func(t *testing.T) {
			f := applyFakeSeed()
			f.rowsErr = map[string]error{sub: boom}
			_, err := DryRun(context.Background(), f.db(t))
			if err == nil || !errors.Is(err, boom) {
				t.Fatalf("err = %v", err)
			}
		})
	}
	t.Run("begin", func(t *testing.T) {
		f := applyFakeSeed()
		f.beginErr = boom
		if _, err := DryRun(context.Background(), f.db(t)); !errors.Is(err, boom) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("commit", func(t *testing.T) {
		f := applyFakeSeed()
		f.commitErr = boom
		plan, err := DryRun(context.Background(), f.db(t))
		if !errors.Is(err, boom) || len(plan.Interactions) != 0 {
			t.Fatalf("err = %v plan = %#v", err, plan)
		}
	})
}

func TestDryRunFakeConflicts(t *testing.T) {
	aID := stableName("ix_", "a", 4)
	cases := map[string]struct {
		setup func(f *applyFake)
		want  string
	}{
		"settings not an object": {func(f *applyFake) { f.element("e1").settings = []byte(`[1]`) }, "element settings are not an object"},
		"markers not an object":  {func(f *applyFake) { f.element("e1").settings = []byte(`{"migratedWorkflowLinks":[]}`) }, "migration markers are not an object"},
		"invalid marker":         {func(f *applyFake) { f.element("e1").settings = []byte(`{"migratedWorkflowLinks":{"a":"yes"}}`) }, "invalid migration marker"},
		"partially migrated":     {func(f *applyFake) { f.element("e1").settings = []byte(`{"migratedWorkflowLinks":{"a":true}}`) }, "workflow has only partially migrated links"},
		"interactions not obj":   {func(f *applyFake) { f.element("e1").settings = []byte(`{"interactions":"x"}`) }, "element interactions are not an object"},
		"interaction exists": {func(f *applyFake) {
			f.element("e1").settings = []byte(`{"interactions":{"` + aID + `":{}}}`)
		}, "element e1 already has interaction " + aID},
		"project logic not an object": {func(f *applyFake) { f.project("p").logic = []byte(`[1]`) }, "project logic is not an object"},
		"project logic invalid":       {func(f *applyFake) { f.project("p").logic = []byte(`{"v":2}`) }, "project logic is invalid: "},
		"page logic invalid":          {func(f *applyFake) { f.page("page").logic = []byte(`{"v":1}`) }, "page page logic is invalid: "},
		"flows not an object": {func(f *applyFake) {
			f.project("p").logic = []byte(`{"v":1,"variables":{},"flows":[]}`)
		}, "logic is invalid: "},
		"flow exists in page": {func(f *applyFake) {
			f.page("page").logic = []byte(`{"v":1,"variables":{},"flows":{"` + stableName("flow_", "w", 16) + `":{"do":[]}}}`)
		}, "already exists in page page"},
		"flow exists in project": {func(f *applyFake) {
			f.project("p").logic = []byte(`{"v":1,"variables":{},"flows":{"` + stableName("flow_", "w", 16) + `":{"do":[]}}}`)
		}, "already exists in project"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			f := applyFake{}
			f = *applyFakeSeed()
			c.setup(&f)
			plan, err := DryRun(context.Background(), f.db(t))
			if err != nil {
				t.Fatal(err)
			}
			got := dryrunFakeIssue(plan, "w")
			if got == "" || !strings.Contains(got, c.want) {
				t.Fatalf("issue = %q, want containing %q (plan %#v)", got, c.want, plan)
			}
			if len(plan.Interactions) != 0 || len(plan.Flows) != 0 {
				t.Fatalf("conflicting workflow must not be planned: %#v", plan)
			}
		})
	}
}

// Pins current behaviour: Validate rejects null flows, so planInTx's own
// "flows missing or null" early return is unreachable for non-empty logic.
func TestDryRunFakeNullFlowsAreRejectedByValidation(t *testing.T) {
	f := applyFakeSeed()
	f.project("p").logic = []byte(`{"v":1,"variables":{},"flows":null}`)
	plan, err := DryRun(context.Background(), f.db(t))
	if err != nil {
		t.Fatal(err)
	}
	if got := dryrunFakeIssue(plan, "w"); !strings.Contains(got, "flows must be an object") || len(plan.Flows) != 0 {
		t.Fatalf("unexpected plan: %#v", plan)
	}
}

func TestDryRunFakeFlowCollisionInSiblingPageOnly(t *testing.T) {
	f := applyFakeSeed()
	f.pages = append(f.pages, applyFakePage{id: "other", project: "p", logic: []byte(`{"v":1,"variables":{},"flows":{"` + stableName("flow_", "w", 16) + `":{"do":[]}}}`)})
	// a page of another project must be ignored
	f.projects = append(f.projects, applyFakeProject{id: "q"})
	f.pages = append(f.pages, applyFakePage{id: "foreign", project: "q", logic: []byte(`garbage`)})
	plan, err := DryRun(context.Background(), f.db(t))
	if err != nil {
		t.Fatal(err)
	}
	if got := dryrunFakeIssue(plan, "w"); !strings.Contains(got, "already exists in page other") || len(plan.Issues) != 1 {
		t.Fatalf("plan: %#v", plan)
	}
}

func TestDryRunFakeDuplicateGeneratedInteractionIDs(t *testing.T) {
	f := applyFakeSeed()
	f.elements = f.elements[:1]
	f.workflows = []applyFakeWorkflow{
		{id: "w1", project: "p", handlers: []byte(applyFakeResetHandlers)},
		{id: "w2", project: "p", handlers: []byte(applyFakeResetHandlers)},
	}
	f.links = []applyFakeLink{{"69235", "w1", "e1", "onClick"}, {"95303", "w2", "e1", "onClick"}}
	plan, err := DryRun(context.Background(), f.db(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"w1", "w2"} {
		if got := dryrunFakeIssue(plan, id); !strings.Contains(got, "multiple workflows generate interaction") {
			t.Fatalf("%s issue = %q", id, got)
		}
	}
	if len(plan.Interactions) != 0 {
		t.Fatalf("colliding interactions must be dropped: %#v", plan.Interactions)
	}
}

func TestDryRunFakeKeepsHealthyWorkflowsBesideFailingOnes(t *testing.T) {
	f := applyFakeSeed()
	f.elements = append(f.elements, applyFakeElement{id: "e3", page: "page", settings: []byte(`[1]`)})
	f.workflows = append(f.workflows, applyFakeWorkflow{id: "x", project: "p", enabled: true, handlers: []byte(applyFakeResetHandlers)})
	f.links = append(f.links, applyFakeLink{"c", "x", "e3", "onClick"})
	plan, err := DryRun(context.Background(), f.db(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Issues) != 1 || plan.Issues[0].WorkflowID != "x" {
		t.Fatalf("issues: %#v", plan.Issues)
	}
	if len(plan.Interactions) != 2 || len(plan.Flows) != 1 || plan.Flows[0].WorkflowID != "w" {
		t.Fatalf("healthy workflow lost: %#v", plan)
	}
}

func TestDryRunFakeNullPageIDElementIsAnIssue(t *testing.T) {
	f := applyFakeSeed()
	f.element("e2").page = nil
	plan, err := DryRun(context.Background(), f.db(t))
	if err != nil {
		t.Fatal(err)
	}
	if got := dryrunFakeIssue(plan, "w"); !strings.Contains(got, "outside the workflow project") {
		t.Fatalf("plan: %#v", plan)
	}
}

func TestObject(t *testing.T) {
	for _, raw := range []string{"", "null"} {
		m, err := object(json.RawMessage(raw))
		if err != nil || m == nil || len(m) != 0 {
			t.Errorf("object(%q) = %v, %v; want empty map", raw, m, err)
		}
	}
	m, err := object(json.RawMessage(`{"a":1}`))
	if err != nil || m["a"] != float64(1) {
		t.Fatalf("object = %v, %v", m, err)
	}
	for _, raw := range []string{`[1]`, `"x"`, `{bad`, `5`} {
		if _, err := object(json.RawMessage(raw)); err == nil {
			t.Errorf("object(%s) must fail", raw)
		}
	}
}
