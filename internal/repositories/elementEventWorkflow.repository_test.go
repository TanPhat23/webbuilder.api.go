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
	got, err := repo.Create(ctx, &models.ElementEventWorkflow{ElementId: "e1", WorkflowId: "w1", EventName: "onClick"})
	if err != nil || got.Id == "" {
		t.Fatalf("got %+v, %v", got, err)
	}
	if got.EventName != "onClick" {
		t.Fatalf("got EventName=%q", got.EventName)
	}
	got, err = repo.Create(ctx, &models.ElementEventWorkflow{Id: "mine"})
	if err != nil || got.Id != "mine" {
		t.Errorf("supplied id changed: %+v, %v", got, err)
	}
	bad, fb := newFakeDB(t)
	fb.onErr("INSERT INTO", errElementEventWorkflowBoom)
	if got, err := elementEventWorkflowNew(bad).Create(ctx, &models.ElementEventWorkflow{}); got != nil || !errors.Is(err, errElementEventWorkflowBoom) {
		t.Errorf("got %+v, %v", got, err)
	}
}

func TestElementEventWorkflowGetByID(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.on(elementEventWorkflowSelect, elementEventWorkflowCols, []driver.Value{"l1", "e1", "w1", "onClick"})
	got, err := elementEventWorkflowNew(db).GetByID(ctx, "l1")
	if err != nil || got == nil {
		t.Fatalf("got %+v, %v", got, err)
	}
	if got.ElementId != "e1" || got.WorkflowId != "w1" {
		t.Fatalf("unexpected row: %+v", got)
	}
	empty, emptyFake := newFakeDB(t)
	if emptyFake == nil {
		t.Fatal("missing fake db")
	}
	if got, err := elementEventWorkflowNew(empty).GetByID(ctx, "x"); got != nil || err != nil {
		t.Errorf("missing should be nil,nil: %+v, %v", got, err)
	}
	bad, fb := newFakeDB(t)
	fb.onErr("SELECT", errElementEventWorkflowBoom)
	if got, err := elementEventWorkflowNew(bad).GetByID(ctx, "x"); got != nil || !errors.Is(err, errElementEventWorkflowBoom) {
		t.Errorf("got %+v, %v", got, err)
	}
}

func assertElementEventWorkflowListCase(
	t *testing.T,
	call func(*ElementEventWorkflowRepository) ([]models.ElementEventWorkflow, error),
	fragment string,
	arg any,
	query string,
) {
	t.Helper()
	db, f := newFakeDB(t)
	f.on(query, elementEventWorkflowCols, []driver.Value{"l1", "e1", "w1", "onClick"}, []driver.Value{"l2", "e2", "w2", "onHover"})
	got, err := call(elementEventWorkflowNew(db))
	if err != nil || len(got) != 2 {
		t.Fatalf("got %+v, %v", got, err)
	}
	if got[0].Id != "l1" || got[1].EventName != "onHover" {
		t.Fatalf("unexpected items: %+v", got)
	}
	if fragment != "" && (!f.sawQuery(fragment) || !elementEventWorkflowArg(f, arg)) {
		t.Errorf("query %q args %v", f.queries[0], f.args[0])
	}
	bad, fb := newFakeDB(t)
	fb.onErr("SELECT", errElementEventWorkflowBoom)
	if got, err := call(elementEventWorkflowNew(bad)); got != nil || !errors.Is(err, errElementEventWorkflowBoom) {
		t.Errorf("error case: %+v, %v", got, err)
	}
}

func TestElementEventWorkflowListQueries(t *testing.T) {
	ctx := context.Background()
	assertElementEventWorkflowListCase(t, func(r *ElementEventWorkflowRepository) ([]models.ElementEventWorkflow, error) {
		return r.ListAll(ctx)
	}, "", nil, elementEventWorkflowSelect)
	assertElementEventWorkflowListCase(t, func(r *ElementEventWorkflowRepository) ([]models.ElementEventWorkflow, error) {
		return r.ListByElement(ctx, "e1")
	}, `"ElementId" =`, "e1", elementEventWorkflowSelect)
	assertElementEventWorkflowListCase(t, func(r *ElementEventWorkflowRepository) ([]models.ElementEventWorkflow, error) {
		return r.ListByWorkflow(ctx, "w1")
	}, `"WorkflowId" =`, "w1", elementEventWorkflowSelect)
	assertElementEventWorkflowListCase(t, func(r *ElementEventWorkflowRepository) ([]models.ElementEventWorkflow, error) {
		return r.ListByEvent(ctx, "onClick")
	}, `"EventName" =`, "onClick", elementEventWorkflowSelect)
	assertElementEventWorkflowListCase(t, func(r *ElementEventWorkflowRepository) ([]models.ElementEventWorkflow, error) {
		return r.ListByPage(ctx, "pg1")
	}, `"Element"."PageId" =`, "pg1", `JOIN "Element"`)
}

func TestElementEventWorkflowFiltersNone(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.on(elementEventWorkflowSelect, elementEventWorkflowCols, []driver.Value{"l1", "e1", "w1", "onClick"})
	got, err := elementEventWorkflowNew(db).ListByFilters(ctx, "", "", "")
	if err != nil || len(got) != 1 {
		t.Fatalf("got %+v, %v", got, err)
	}
	if f.sawQuery(`"ElementId" =`) || f.sawQuery(`"WorkflowId" =`) {
		t.Errorf("unexpected filter: %q", f.queries[0])
	}
	if f.sawQuery(`"EventName" =`) {
		t.Errorf("unexpected event filter: %q", f.queries[0])
	}
}

func TestElementEventWorkflowFiltersAllAndErr(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	if _, err := elementEventWorkflowNew(db).ListByFilters(ctx, "e1", "w1", "onClick"); err != nil {
		t.Fatal(err)
	}
	if !f.sawQuery(`"ElementId" =`, `"WorkflowId" =`, `"EventName" =`) {
		t.Errorf("filters missing: %q", f.queries[0])
	}

	dbErr, fErr := newFakeDB(t)
	fErr.onErr("SELECT", errElementEventWorkflowBoom)
	if got, err := elementEventWorkflowNew(dbErr).ListByFilters(ctx, "", "", ""); got != nil || !errors.Is(err, errElementEventWorkflowBoom) {
		t.Errorf("got %+v, %v", got, err)
	}
}

func TestElementEventWorkflowUpdate(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.onExec("UPDATE", 1)
	f.on(elementEventWorkflowSelect, elementEventWorkflowCols, []driver.Value{"l1", "e1", "w2", "onHover"})
	got, err := elementEventWorkflowNew(db).Update(ctx, "l1", &models.ElementEventWorkflow{WorkflowId: "w2", EventName: "onHover"})
	if err != nil || got == nil {
		t.Fatalf("got %+v, %v", got, err)
	}
	if got.WorkflowId != "w2" || got.EventName != "onHover" {
		t.Fatalf("got %+v", got)
	}
	if !f.sawQuery("UPDATE", "ElementEventWorkflow", `"Id" =`) {
		t.Errorf("queries: %v", f.queries)
	}

	dbErr, fErr := newFakeDB(t)
	fErr.onErr("UPDATE", errElementEventWorkflowBoom)
	if got, err := elementEventWorkflowNew(dbErr).Update(ctx, "l1", &models.ElementEventWorkflow{EventName: "x"}); got != nil || !errors.Is(err, errElementEventWorkflowBoom) {
		t.Errorf("got %+v, %v", got, err)
	}
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
		{"by id", func(r *ElementEventWorkflowRepository) error { return r.Delete(ctx, "l1") }, `"Id" =`},
		{"by element", func(r *ElementEventWorkflowRepository) error {
			return r.DeleteByElement(ctx, "e1")
		}, `"ElementId" =`},
		{"by workflow", func(r *ElementEventWorkflowRepository) error {
			return r.DeleteByWorkflow(ctx, "w1")
		}, `"WorkflowId" =`},
	}
	for _, tc := range calls {
		if err := tc.call(repo); err != nil {
			t.Fatal(err)
		}
		if !f.sawQuery("DELETE FROM", "ElementEventWorkflow", tc.fragment) {
			t.Errorf("%s query: %q", tc.name, f.lastQuery())
		}
		bad, fb := newFakeDB(t)
		fb.onErr("DELETE FROM", errElementEventWorkflowBoom)
		if err := tc.call(elementEventWorkflowNew(bad)); !errors.Is(err, errElementEventWorkflowBoom) {
			t.Errorf("%s error case: %v", tc.name, err)
		}
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
		got, err := elementEventWorkflowNew(db).IsLinked(ctx, "e1", "w1", "onClick")
		if err != nil || got != tc.want {
			t.Errorf("count=%d: got %v, %v", tc.n, got, err)
		}
		if !f.sawQuery(`"ElementId" =`, `"WorkflowId" =`, `"EventName" =`) {
			t.Errorf("query: %q", f.queries[0])
		}
	}
	bad, fb := newFakeDB(t)
	fb.onErr("count(*)", errElementEventWorkflowBoom)
	if ok, err := elementEventWorkflowNew(bad).IsLinked(ctx, "e", "w", "n"); ok || !errors.Is(err, errElementEventWorkflowBoom) {
		t.Errorf("got %v, %v", ok, err)
	}
}
