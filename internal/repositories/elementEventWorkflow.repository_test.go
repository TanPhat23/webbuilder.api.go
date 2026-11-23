package repositories

import (
	"context"
	"database/sql/driver"
	"errors"
	"testing"

	"gorm.io/gorm"

	"my-go-app/internal/models"
)

var errElementEventWorkflowBoom = errors.New("element event workflow boom")

const elementEventWorkflowSelect = `SELECT * FROM "public"."ElementEventWorkflow"`

var elementEventWorkflowCols = []string{"Id", "ElementId", "WorkflowId", "EventName"}

func elementEventWorkflowNew(db *gorm.DB) *ElementEventWorkflowRepository {
	return &ElementEventWorkflowRepository{db: db}
}

func elementEventWorkflowArg(f *fakeDB, want any) bool {
	for _, q := range f.args {
		for _, a := range q {
			if a.Value == want {
				return true
			}
		}
	}
	return false
}

func TestElementEventWorkflowCreate(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.onExec("INSERT INTO", 1)
	repo := elementEventWorkflowNew(db)
	got, err := repo.CreateElementEventWorkflow(ctx, &models.ElementEventWorkflow{ElementId: "e1", WorkflowId: "w1", EventName: "onClick"})
	if err != nil || got.Id == "" || got.EventName != "onClick" {
		t.Fatalf("got %+v, %v", got, err)
	}
	got, err = repo.CreateElementEventWorkflow(ctx, &models.ElementEventWorkflow{Id: "mine"})
	if err != nil || got.Id != "mine" {
		t.Errorf("supplied id changed: %+v, %v", got, err)
	}
	bad, fb := newFakeDB(t)
	fb.onErr("INSERT INTO", errElementEventWorkflowBoom)
	if got, err := elementEventWorkflowNew(bad).CreateElementEventWorkflow(ctx, &models.ElementEventWorkflow{}); got != nil || !errors.Is(err, errElementEventWorkflowBoom) {
		t.Errorf("got %+v, %v", got, err)
	}
}

func TestElementEventWorkflowGetByID(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.on(elementEventWorkflowSelect, elementEventWorkflowCols, []driver.Value{"l1", "e1", "w1", "onClick"})
	got, err := elementEventWorkflowNew(db).GetElementEventWorkflowByID(ctx, "l1")
	if err != nil || got == nil || got.ElementId != "e1" || got.WorkflowId != "w1" {
		t.Fatalf("got %+v, %v", got, err)
	}
	empty, _ := newFakeDB(t)
	if got, err := elementEventWorkflowNew(empty).GetElementEventWorkflowByID(ctx, "x"); got != nil || err != nil {
		t.Errorf("missing should be nil,nil: %+v, %v", got, err)
	}
	bad, fb := newFakeDB(t)
	fb.onErr("SELECT", errElementEventWorkflowBoom)
	if got, err := elementEventWorkflowNew(bad).GetElementEventWorkflowByID(ctx, "x"); got != nil || !errors.Is(err, errElementEventWorkflowBoom) {
		t.Errorf("got %+v, %v", got, err)
	}
}

func TestElementEventWorkflowListQueries(t *testing.T) {
	ctx := context.Background()
	type listFn func(*ElementEventWorkflowRepository) ([]models.ElementEventWorkflow, error)
	cases := []struct {
		name     string
		call     listFn
		fragment string
		arg      any
		query    string
	}{
		{"all", func(r *ElementEventWorkflowRepository) ([]models.ElementEventWorkflow, error) {
			return r.GetAllElementEventWorkflows(ctx)
		}, "", nil, elementEventWorkflowSelect},
		{"by element", func(r *ElementEventWorkflowRepository) ([]models.ElementEventWorkflow, error) {
			return r.GetElementEventWorkflowsByElementID(ctx, "e1")
		}, `"ElementId" =`, "e1", elementEventWorkflowSelect},
		{"by workflow", func(r *ElementEventWorkflowRepository) ([]models.ElementEventWorkflow, error) {
			return r.GetElementEventWorkflowsByWorkflowID(ctx, "w1")
		}, `"WorkflowId" =`, "w1", elementEventWorkflowSelect},
		{"by event name", func(r *ElementEventWorkflowRepository) ([]models.ElementEventWorkflow, error) {
			return r.GetElementEventWorkflowsByEventName(ctx, "onClick")
		}, `"EventName" =`, "onClick", elementEventWorkflowSelect},
		{"by page", func(r *ElementEventWorkflowRepository) ([]models.ElementEventWorkflow, error) {
			return r.GetElementEventWorkflowsByPageID(ctx, "pg1")
		}, `"Element"."PageId" =`, "pg1", `JOIN "Element"`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db, f := newFakeDB(t)
			f.on(tc.query, elementEventWorkflowCols, []driver.Value{"l1", "e1", "w1", "onClick"}, []driver.Value{"l2", "e2", "w2", "onHover"})
			got, err := tc.call(elementEventWorkflowNew(db))
			if err != nil || len(got) != 2 || got[0].Id != "l1" || got[1].EventName != "onHover" {
				t.Fatalf("got %+v, %v", got, err)
			}
			if tc.fragment != "" && (!f.sawQuery(tc.fragment) || !elementEventWorkflowArg(f, tc.arg)) {
				t.Errorf("query %q args %v", f.queries[0], f.args[0])
			}
			bad, fb := newFakeDB(t)
			fb.onErr("SELECT", errElementEventWorkflowBoom)
			if got, err := tc.call(elementEventWorkflowNew(bad)); got != nil || !errors.Is(err, errElementEventWorkflowBoom) {
				t.Errorf("error case: %+v, %v", got, err)
			}
		})
	}
}

func TestElementEventWorkflowFilters(t *testing.T) {
	ctx := context.Background()
	t.Run("none", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on(elementEventWorkflowSelect, elementEventWorkflowCols, []driver.Value{"l1", "e1", "w1", "onClick"})
		got, err := elementEventWorkflowNew(db).GetElementEventWorkflowsByFilters(ctx, "", "", "")
		if err != nil || len(got) != 1 {
			t.Fatalf("got %+v, %v", got, err)
		}
		if f.sawQuery(`"ElementId" =`) || f.sawQuery(`"WorkflowId" =`) || f.sawQuery(`"EventName" =`) {
			t.Errorf("unexpected filter: %q", f.queries[0])
		}
	})
	t.Run("all", func(t *testing.T) {
		db, f := newFakeDB(t)
		if _, err := elementEventWorkflowNew(db).GetElementEventWorkflowsByFilters(ctx, "e1", "w1", "onClick"); err != nil {
			t.Fatal(err)
		}
		if !f.sawQuery(`"ElementId" =`, `"WorkflowId" =`, `"EventName" =`) {
			t.Errorf("filters missing: %q", f.queries[0])
		}
	})
	t.Run("error", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onErr("SELECT", errElementEventWorkflowBoom)
		if got, err := elementEventWorkflowNew(db).GetElementEventWorkflowsByFilters(ctx, "", "", ""); got != nil || !errors.Is(err, errElementEventWorkflowBoom) {
			t.Errorf("got %+v, %v", got, err)
		}
	})
}

func TestElementEventWorkflowUpdate(t *testing.T) {
	ctx := context.Background()
	t.Run("updates then refetches", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onExec("UPDATE", 1)
		f.on(elementEventWorkflowSelect, elementEventWorkflowCols, []driver.Value{"l1", "e1", "w2", "onHover"})
		got, err := elementEventWorkflowNew(db).UpdateElementEventWorkflow(ctx, "l1", &models.ElementEventWorkflow{WorkflowId: "w2", EventName: "onHover"})
		if err != nil || got == nil || got.WorkflowId != "w2" || got.EventName != "onHover" {
			t.Fatalf("got %+v, %v", got, err)
		}
		if !f.sawQuery("UPDATE", "ElementEventWorkflow", `"Id" =`) {
			t.Errorf("queries: %v", f.queries)
		}
	})
	t.Run("error", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onErr("UPDATE", errElementEventWorkflowBoom)
		if got, err := elementEventWorkflowNew(db).UpdateElementEventWorkflow(ctx, "l1", &models.ElementEventWorkflow{EventName: "x"}); got != nil || !errors.Is(err, errElementEventWorkflowBoom) {
			t.Errorf("got %+v, %v", got, err)
		}
	})
}

func TestElementEventWorkflowDeletes(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.onExec("DELETE FROM", 1)
	repo := elementEventWorkflowNew(db)
	calls := []struct {
		name     string
		call     func(*ElementEventWorkflowRepository) error
		fragment string
	}{
		{"by id", func(r *ElementEventWorkflowRepository) error { return r.DeleteElementEventWorkflow(ctx, "l1") }, `"Id" =`},
		{"by element", func(r *ElementEventWorkflowRepository) error {
			return r.DeleteElementEventWorkflowsByElementID(ctx, "e1")
		}, `"ElementId" =`},
		{"by workflow", func(r *ElementEventWorkflowRepository) error {
			return r.DeleteElementEventWorkflowsByWorkflowID(ctx, "w1")
		}, `"WorkflowId" =`},
	}
	for _, tc := range calls {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.call(repo); err != nil {
				t.Fatal(err)
			}
			if !f.sawQuery("DELETE FROM", "ElementEventWorkflow", tc.fragment) {
				t.Errorf("query: %q", f.lastQuery())
			}
			bad, fb := newFakeDB(t)
			fb.onErr("DELETE FROM", errElementEventWorkflowBoom)
			if err := tc.call(elementEventWorkflowNew(bad)); !errors.Is(err, errElementEventWorkflowBoom) {
				t.Errorf("error case: %v", err)
			}
		})
	}
}

func TestElementEventWorkflowCheckIfLinked(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		n    int64
		want bool
	}{{0, false}, {1, true}} {
		db, f := newFakeDB(t)
		f.on("count(*)", []string{"count"}, []driver.Value{tc.n})
		got, err := elementEventWorkflowNew(db).CheckIfWorkflowLinkedToElement(ctx, "e1", "w1", "onClick")
		if err != nil || got != tc.want {
			t.Errorf("count=%d: got %v, %v", tc.n, got, err)
		}
		if !f.sawQuery(`"ElementId" =`, `"WorkflowId" =`, `"EventName" =`) {
			t.Errorf("query: %q", f.queries[0])
		}
	}
	bad, fb := newFakeDB(t)
	fb.onErr("count(*)", errElementEventWorkflowBoom)
	if ok, err := elementEventWorkflowNew(bad).CheckIfWorkflowLinkedToElement(ctx, "e", "w", "n"); ok || !errors.Is(err, errElementEventWorkflowBoom) {
		t.Errorf("got %v, %v", ok, err)
	}
}
