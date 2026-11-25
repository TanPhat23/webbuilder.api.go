package repositories

import (
	"context"
	"database/sql/driver"
	"errors"
	"testing"

	"gorm.io/gorm"

	"my-go-app/internal/models"
)

var errEventWorkflowBoom = errors.New("event workflow boom")

const eventWorkflowSelect = `SELECT * FROM "public"."EventWorkflow"`

var eventWorkflowCols = []string{"Id", "ProjectId", "Name", "Enabled"}

func eventWorkflowNew(db *gorm.DB) *EventWorkflowRepository { return &EventWorkflowRepository{db: db} }

func eventWorkflowArg(f *fakeDB, want any) bool {
	for _, q := range f.args {
		for _, a := range q {
			if a.Value == want {
				return true
			}
		}
	}
	return false
}

func assertWorkflowListQuery(t *testing.T, f *fakeDB, got []models.EventWorkflow, err error, fragments []string, arg any) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 || got[0].Id != "w1" {
		t.Fatalf("unexpected rows: %+v", got)
	}
	if got[1].Name != "other" {
		t.Fatalf("unexpected second row: %+v", got[1])
	}
	if !f.sawQuery(fragments...) || !eventWorkflowArg(f, arg) {
		t.Errorf("query %q args %v", f.queries[0], f.args[0])
	}
}

func TestEventWorkflowCreate(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.onExec("INSERT INTO", 1)
	repo := eventWorkflowNew(db)
	got, err := repo.Create(ctx, &models.EventWorkflow{ProjectId: "p1", Name: "flow"})
	if err != nil || got.Id == "" {
		t.Fatalf("got %+v, %v", got, err)
	}
	if got.Name != "flow" {
		t.Fatalf("got name=%q, want flow", got.Name)
	}
	got, err = repo.Create(ctx, &models.EventWorkflow{Id: "mine", ProjectId: "p1", Name: "flow"})
	if err != nil || got.Id != "mine" {
		t.Errorf("supplied id changed: %+v, %v", got, err)
	}
	if !f.sawQuery("INSERT INTO", "EventWorkflow") {
		t.Error("no insert")
	}

	bad, fb := newFakeDB(t)
	fb.onErr("INSERT INTO", errEventWorkflowBoom)
	if got, err := eventWorkflowNew(bad).Create(ctx, &models.EventWorkflow{}); got != nil || !errors.Is(err, errEventWorkflowBoom) {
		t.Errorf("got %+v, %v", got, err)
	}
}

func TestEventWorkflowGetByID(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.on(eventWorkflowSelect, eventWorkflowCols, []driver.Value{"w1", "p1", "flow", true})
	got, err := eventWorkflowNew(db).GetByID(ctx, "w1")
	if err != nil || got == nil {
		t.Fatalf("got %+v, %v", got, err)
	}
	if got.Id != "w1" || !got.Enabled {
		t.Fatalf("unexpected row: %+v", got)
	}
	empty, emptyFake := newFakeDB(t)
	if emptyFake == nil {
		t.Fatal("missing fake db")
	}
	if got, err := eventWorkflowNew(empty).GetByID(ctx, "x"); got != nil || err != nil {
		t.Errorf("missing should be nil,nil; got %+v, %v", got, err)
	}
	bad, fb := newFakeDB(t)
	fb.onErr("SELECT", errEventWorkflowBoom)
	if got, err := eventWorkflowNew(bad).GetByID(ctx, "x"); got != nil || !errors.Is(err, errEventWorkflowBoom) {
		t.Errorf("got %+v, %v", got, err)
	}
}

func TestEventWorkflowListByProject(t *testing.T) {
	runWorkflowListCase(t, func(ctx context.Context, r *EventWorkflowRepository) ([]models.EventWorkflow, error) {
		return r.ListByProject(ctx, "p1")
	}, []string{`"ProjectId" =`}, "p1")
}

func TestEventWorkflowListByProjectWithElements(t *testing.T) {
	runWorkflowListCase(t, func(ctx context.Context, r *EventWorkflowRepository) ([]models.EventWorkflow, error) {
		return r.ListByProjectWithElements(ctx, "p1")
	}, []string{`"ProjectId" =`}, "p1")
}

func TestEventWorkflowListEnabled(t *testing.T) {
	runWorkflowListCase(t, func(ctx context.Context, r *EventWorkflowRepository) ([]models.EventWorkflow, error) {
		return r.ListEnabledByProject(ctx, "p1")
	}, []string{`"Enabled" =`}, true)
}

func TestEventWorkflowListByName(t *testing.T) {
	runWorkflowListCase(t, func(ctx context.Context, r *EventWorkflowRepository) ([]models.EventWorkflow, error) {
		return r.ListByName(ctx, "p1", "fl")
	}, []string{`"Name" ILIKE`}, "%fl%")
}

func runWorkflowListCase(t *testing.T, call func(context.Context, *EventWorkflowRepository) ([]models.EventWorkflow, error), fragments []string, arg any) {
	t.Helper()
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.on(eventWorkflowSelect, eventWorkflowCols, []driver.Value{"w1", "p1", "flow", true}, []driver.Value{"w2", "p1", "other", false})
	got, err := call(ctx, eventWorkflowNew(db))
	assertWorkflowListQuery(t, f, got, err, fragments, arg)

	bad, fb := newFakeDB(t)
	fb.onErr("SELECT", errEventWorkflowBoom)
	if got, err := call(ctx, eventWorkflowNew(bad)); got != nil || !errors.Is(err, errEventWorkflowBoom) {
		t.Errorf("error case: %+v, %v", got, err)
	}
}

func TestEventWorkflowUpdateIncludesJSONWhenProvided(t *testing.T) {
	ctx := context.Background()
	desc := "d"
	db, f := newFakeDB(t)
	f.onExec("UPDATE", 1)
	f.on(eventWorkflowSelect, eventWorkflowCols, []driver.Value{"w1", "p1", "renamed", false})
	got, err := eventWorkflowNew(db).Update(ctx, "w1", &models.EventWorkflow{
		Name: "renamed", Description: &desc, CanvasData: []byte(`{"a":1}`), Handlers: []byte(`[]`),
	})
	if err != nil || got == nil {
		t.Fatalf("got %+v, %v", got, err)
	}
	if got.Name != "renamed" {
		t.Fatalf("got name=%q, want renamed", got.Name)
	}
	if !f.sawQuery("UPDATE", `"CanvasData"`, `"Handlers"`, `"Enabled"`, `"Name"`) {
		t.Errorf("update: %v", f.queries)
	}

	db2, f2 := newFakeDB(t)
	f2.onExec("UPDATE", 1)
	if _, err := eventWorkflowNew(db2).Update(ctx, "w1", &models.EventWorkflow{Name: "n"}); err != nil {
		t.Fatal(err)
	}
	if f2.sawQuery(`"CanvasData"`) || f2.sawQuery(`"Handlers"`) {
		t.Errorf("empty json must not be written: %q", f2.queries[0])
	}
}

func TestEventWorkflowUpdateError(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.onErr("UPDATE", errEventWorkflowBoom)
	if got, err := eventWorkflowNew(db).Update(ctx, "w1", &models.EventWorkflow{}); got != nil || !errors.Is(err, errEventWorkflowBoom) {
		t.Errorf("got %+v, %v", got, err)
	}
}

func TestEventWorkflowEnabledDeleteAndCount(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.onExec("UPDATE", 1)
	f.onExec("DELETE FROM", 1)
	f.on("count(*)", []string{"count"}, []driver.Value{int64(4)})
	repo := eventWorkflowNew(db)

	if err := repo.SetEnabled(ctx, "w1", true); err != nil {
		t.Fatal(err)
	}
	if !f.sawQuery("UPDATE", `"Enabled"`) || !eventWorkflowArg(f, true) {
		t.Errorf("enabled update: %v", f.queries)
	}
	if err := repo.Delete(ctx, "w1"); err != nil {
		t.Fatal(err)
	}
	if err := repo.DeleteByProject(ctx, "p1"); err != nil {
		t.Fatal(err)
	}
	if !f.sawQuery("DELETE FROM", `"Id" =`) || !f.sawQuery("DELETE FROM", `"ProjectId" =`) {
		t.Errorf("deletes: %v", f.queries)
	}
	if n, err := repo.CountByProject(ctx, "p1"); err != nil || n != 4 {
		t.Errorf("count = %d, %v", n, err)
	}
}

func TestEventWorkflowEnabledDeleteAndCountErrors(t *testing.T) {
	ctx := context.Background()
	bad, fb := newFakeDB(t)
	fb.onErr("UPDATE", errEventWorkflowBoom)
	fb.onErr("DELETE FROM", errEventWorkflowBoom)
	fb.onErr("count(*)", errEventWorkflowBoom)
	br := eventWorkflowNew(bad)
	if err := br.SetEnabled(ctx, "w1", true); !errors.Is(err, errEventWorkflowBoom) {
		t.Errorf("enabled: %v", err)
	}
	if err := br.Delete(ctx, "w1"); !errors.Is(err, errEventWorkflowBoom) {
		t.Errorf("delete: %v", err)
	}
	if err := br.DeleteByProject(ctx, "p1"); !errors.Is(err, errEventWorkflowBoom) {
		t.Errorf("delete by project: %v", err)
	}
	if n, err := br.CountByProject(ctx, "p1"); n != 0 || !errors.Is(err, errEventWorkflowBoom) {
		t.Errorf("count: %d, %v", n, err)
	}
}

func assertWorkflowNameExistsCase(t *testing.T, n int64, exclude string, want bool) {
	t.Helper()
	db, f := newFakeDB(t)
	f.on("count(*)", []string{"count"}, []driver.Value{n})
	got, err := eventWorkflowNew(db).ExistsByName(context.Background(), "p1", "flow", exclude)
	if err != nil || got != want {
		t.Errorf("got %v, %v", got, err)
	}
	if f.sawQuery(`"Id" !=`) != (exclude != "") {
		t.Errorf("exclusion clause mismatch: %q", f.queries[0])
	}
}

func TestEventWorkflowCheckIfNameExists(t *testing.T) {
	assertWorkflowNameExistsCase(t, 1, "", true)
	assertWorkflowNameExistsCase(t, 0, "", false)
	assertWorkflowNameExistsCase(t, 2, "w1", true)

	bad, fb := newFakeDB(t)
	fb.onErr("count(*)", errEventWorkflowBoom)
	if ok, err := eventWorkflowNew(bad).ExistsByName(context.Background(), "p1", "n", ""); ok || !errors.Is(err, errEventWorkflowBoom) {
		t.Errorf("got %v, %v", ok, err)
	}
}

func TestEventWorkflowFiltersNone(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.on(eventWorkflowSelect, eventWorkflowCols, []driver.Value{"w1", "p1", "a", true})
	got, err := eventWorkflowNew(db).ListWithFilters(ctx, "", nil, "")
	if err != nil || len(got) != 1 {
		t.Fatalf("got %+v, %v", got, err)
	}
	if f.sawQuery(`"ProjectId" =`) || f.sawQuery(`"Enabled" =`) {
		t.Errorf("unexpected filters: %q", f.queries[0])
	}
	if f.sawQuery("ILIKE") {
		t.Errorf("unexpected ILIKE filter: %q", f.queries[0])
	}
}

func TestEventWorkflowFiltersAllAndError(t *testing.T) {
	ctx := context.Background()
	enabled := true
	db, f := newFakeDB(t)
	if _, err := eventWorkflowNew(db).ListWithFilters(ctx, "p1", &enabled, "fl"); err != nil {
		t.Fatal(err)
	}
	if !f.sawQuery(`"ProjectId" =`, `"Enabled" =`, `"Name" ILIKE`) {
		t.Errorf("query %q", f.queries[0])
	}
	if !eventWorkflowArg(f, "%fl%") || !eventWorkflowArg(f, "p1") {
		t.Errorf("args %v", f.args[0])
	}
	if !eventWorkflowArg(f, true) {
		t.Errorf("missing enabled arg: %v", f.args[0])
	}

	bad, fb := newFakeDB(t)
	fb.onErr("SELECT", errEventWorkflowBoom)
	if got, err := eventWorkflowNew(bad).ListWithFilters(ctx, "", nil, ""); got != nil || !errors.Is(err, errEventWorkflowBoom) {
		t.Errorf("got %+v, %v", got, err)
	}
}
