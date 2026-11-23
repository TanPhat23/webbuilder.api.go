package workflowmigration

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func applyFakeDecode(t *testing.T, raw []byte) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("invalid json %q: %v", raw, err)
	}
	return m
}

func TestApplyFakeWritesInteractionsAndPageFlow(t *testing.T) {
	f := applyFakeSeed()
	plan, err := Apply(context.Background(), f.db(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Issues) != 0 || len(plan.Interactions) != 2 || len(plan.Flows) != 1 {
		t.Fatalf("unexpected plan: %#v", plan)
	}
	if f.commits != 1 {
		t.Fatalf("commits = %d", f.commits)
	}
	if f.beginOpts.ReadOnly || f.beginOpts.Isolation == 0 {
		t.Fatalf("apply must use a writable non-default isolation level: %#v", f.beginOpts)
	}
	flow := stableName("flow_", "w", 16)
	s1 := applyFakeDecode(t, f.element("e1").settings)
	if s1["keep"] != true {
		t.Fatalf("existing settings lost: %v", s1)
	}
	ix := s1["interactions"].(map[string]any)[stableName("ix_", "a", 4)].(map[string]any)
	if ix["on"] != "click" || ix["enabled"] != true {
		t.Fatalf("bad interaction: %v", ix)
	}
	do := ix["do"].([]any)
	if len(do) != 1 || do[0].(map[string]any)["run"] != flow {
		t.Fatalf("interaction must call the shared flow: %v", ix)
	}
	if s1["migratedWorkflowLinks"].(map[string]any)["a"] != true {
		t.Fatalf("link marker missing: %v", s1)
	}
	s2 := applyFakeDecode(t, f.element("e2").settings)
	if s2["migratedWorkflowLinks"].(map[string]any)["b"] != true {
		t.Fatalf("second marker missing: %v", s2)
	}
	logic := applyFakeDecode(t, f.page("page").logic)
	flows := logic["flows"].(map[string]any)
	steps := flows[flow].(map[string]any)["do"].([]any)
	if logic["v"] != float64(1) || len(steps) != 1 || steps[0].(map[string]any)["reset"] != "f" {
		t.Fatalf("bad page logic: %v", logic)
	}
	if f.project("p").logic != nil {
		t.Fatalf("project logic must stay untouched: %s", f.project("p").logic)
	}
}

func TestApplyFakeIsIdempotent(t *testing.T) {
	f := applyFakeSeed()
	db := f.db(t)
	if _, err := Apply(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	before := len(f.execs)
	plan, err := Apply(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Issues) != 0 || len(plan.Interactions) != 0 || len(plan.Flows) != 0 {
		t.Fatalf("second run must be empty: %#v", plan)
	}
	if len(f.execs)-before != 1 {
		t.Fatalf("second run must only take the lock, executed %v", f.execs[before:])
	}
}

func TestApplyFakeSingleElementWritesNoFlow(t *testing.T) {
	f := applyFakeSeed()
	f.links = f.links[:1]
	plan, err := Apply(context.Background(), f.db(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Interactions) != 1 || len(plan.Flows) != 0 {
		t.Fatalf("unexpected plan: %#v", plan)
	}
	s := applyFakeDecode(t, f.element("e1").settings)
	do := s["interactions"].(map[string]any)[stableName("ix_", "a", 4)].(map[string]any)["do"].([]any)
	if len(do) != 1 || do[0].(map[string]any)["reset"] != "f" {
		t.Fatalf("steps must be inlined: %v", do)
	}
	if f.page("page").logic != nil {
		t.Fatal("no flow expected")
	}
}

func TestApplyFakeWritesProjectFlowWhenElementsSpanPages(t *testing.T) {
	f := applyFakeSeed()
	f.pages = append(f.pages, applyFakePage{id: "page2", project: "p"})
	f.element("e2").page = "page2"
	plan, err := Apply(context.Background(), f.db(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Flows) != 1 || plan.Flows[0].PageID != "" {
		t.Fatalf("unexpected plan: %#v", plan)
	}
	logic := applyFakeDecode(t, f.project("p").logic)
	if _, ok := logic["flows"].(map[string]any)[stableName("flow_", "w", 16)]; !ok {
		t.Fatalf("flow missing from project logic: %v", logic)
	}
	if f.page("page").logic != nil || f.page("page2").logic != nil {
		t.Fatal("page logic must stay untouched")
	}
}

func TestApplyFakePreservesExistingFlowsAndInteractions(t *testing.T) {
	f := applyFakeSeed()
	f.element("e1").settings = []byte(`{"interactions":{"other":{"on":"click","do":[]}},"migratedWorkflowLinks":{"zz":true}}`)
	f.page("page").logic = []byte(`{"v":1,"variables":{},"flows":{"mine":{"do":[]}}}`)
	if _, err := Apply(context.Background(), f.db(t)); err != nil {
		t.Fatal(err)
	}
	s := applyFakeDecode(t, f.element("e1").settings)
	if len(s["interactions"].(map[string]any)) != 2 || len(s["migratedWorkflowLinks"].(map[string]any)) != 2 {
		t.Fatalf("existing entries lost: %v", s)
	}
	if len(applyFakeDecode(t, f.page("page").logic)["flows"].(map[string]any)) != 2 {
		t.Fatalf("existing flow lost: %s", f.page("page").logic)
	}
}

func TestApplyFakeReturnsIssuesWithoutWriting(t *testing.T) {
	f := applyFakeSeed()
	f.workflows = append(f.workflows, applyFakeWorkflow{id: "bad", project: "p", handlers: []byte(`[{"config":{"type":"unknown"}}]`)})
	f.links = append(f.links, applyFakeLink{"c", "bad", "e1", "onClick"})
	plan, err := Apply(context.Background(), f.db(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Issues) != 1 || plan.Issues[0].WorkflowID != "bad" {
		t.Fatalf("unexpected issues: %#v", plan.Issues)
	}
	if f.commits != 0 || f.rollbacks == 0 {
		t.Fatalf("commits=%d rollbacks=%d", f.commits, f.rollbacks)
	}
	for _, q := range f.execs {
		if strings.HasPrefix(q, "UPDATE") {
			t.Fatalf("unexpected write: %s", q)
		}
	}
}

func TestApplyFakeErrorPropagation(t *testing.T) {
	boom := errors.New("boom")
	cases := map[string]func(f *applyFake){
		"begin":          func(f *applyFake) { f.beginErr = boom },
		"lock":           func(f *applyFake) { f.failExec = map[string]error{"LOCK TABLE": boom} },
		"plan read":      func(f *applyFake) { f.failQuery = map[string]error{`FROM "Project"`: boom} },
		"element select": func(f *applyFake) { f.failQuery = map[string]error{`SELECT "Settings" FROM "Element" WHERE`: boom} },
		"element update": func(f *applyFake) { f.failExec = map[string]error{`UPDATE "Element"`: boom} },
		"flow select":    func(f *applyFake) { f.failQuery = map[string]error{`SELECT "Logic" FROM "Page" WHERE`: boom} },
		"flow update":    func(f *applyFake) { f.failExec = map[string]error{`UPDATE "Page"`: boom} },
		"commit":         func(f *applyFake) { f.commitErr = boom },
	}
	for name, setup := range cases {
		t.Run(name, func(t *testing.T) {
			f := applyFakeSeed()
			setup(f)
			plan, err := Apply(context.Background(), f.db(t))
			if err == nil || !strings.Contains(err.Error(), "boom") {
				t.Fatalf("err = %v", err)
			}
			if len(plan.Interactions) != 0 || len(plan.Flows) != 0 || len(plan.Issues) != 0 {
				t.Fatalf("a failed apply must return an empty plan: %#v", plan)
			}
		})
	}
}

// The hook changes data after planning, simulating drift between the planning
// read and the write phase; Apply re-reads each row and must reject bad shapes.
func TestApplyFakeRejectsDataThatChangedAfterPlanning(t *testing.T) {
	cases := map[string]struct {
		mutate func(f *applyFake)
		query  string
		want   string
	}{
		"settings not an object": {func(f *applyFake) { f.element("e1").settings = []byte(`[1]`) }, `SELECT "Settings"`, "cannot unmarshal"},
		"interactions invalid":   {func(f *applyFake) { f.element("e1").settings = []byte(`{"interactions":5}`) }, `SELECT "Settings"`, "invalid interactions for e1"},
		"markers invalid":        {func(f *applyFake) { f.element("e1").settings = []byte(`{"migratedWorkflowLinks":5}`) }, `SELECT "Settings"`, "invalid migration markers for e1"},
		"element vanished":       {func(f *applyFake) { f.elements = f.elements[1:] }, `SELECT "Settings"`, "no rows"},
		"logic is not json":      {func(f *applyFake) { f.page("page").logic = []byte(`{bad`) }, `SELECT "Logic"`, "invalid character"},
		"logic invalid":          {func(f *applyFake) { f.page("page").logic = []byte(`{"v":2}`) }, `SELECT "Logic"`, "invalid logic document"},
		"page vanished":          {func(f *applyFake) { f.pages = nil }, `SELECT "Logic"`, "no rows"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			f := applyFakeSeed()
			armed := false
			f.hook = func(f *applyFake, q string) {
				// Only drift once the planning reads (which have no WHERE) are done.
				if strings.Contains(q, "WHERE") && strings.HasPrefix(q, c.query) && !armed {
					armed = true
					c.mutate(f)
				}
			}
			_, err := Apply(context.Background(), f.db(t))
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("err = %v, want containing %q", err, c.want)
			}
			if f.commits != 0 {
				t.Fatal("must not commit")
			}
		})
	}
}
