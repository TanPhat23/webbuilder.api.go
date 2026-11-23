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

func TestEventWorkflowCreate(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.onExec("INSERT INTO", 1)
	repo := eventWorkflowNew(db)
	got, err := repo.CreateEventWorkflow(ctx, &models.EventWorkflow{ProjectId: "p1", Name: "flow"})
	if err != nil || got.Id == "" || got.Name != "flow" {
		t.Fatalf("got %+v, %v", got, err)
	}
	got, err = repo.CreateEventWorkflow(ctx, &models.EventWorkflow{Id: "mine", ProjectId: "p1", Name: "flow"})
	if err != nil || got.Id != "mine" {
		t.Errorf("supplied id changed: %+v, %v", got, err)
	}
	if !f.sawQuery("INSERT INTO", "EventWorkflow") {
		t.Error("no insert")
	}

	bad, fb := newFakeDB(t)
	fb.onErr("INSERT INTO", errEventWorkflowBoom)
	if got, err := eventWorkflowNew(bad).CreateEventWorkflow(ctx, &models.EventWorkflow{}); got != nil || !errors.Is(err, errEventWorkflowBoom) {
		t.Errorf("got %+v, %v", got, err)
	}
}

func TestEventWorkflowGetByID(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.on(eventWorkflowSelect, eventWorkflowCols, []driver.Value{"w1", "p1", "flow", true})
	got, err := eventWorkflowNew(db).GetEventWorkflowByID(ctx, "w1")
	if err != nil || got == nil || got.Id != "w1" || !got.Enabled {
		t.Fatalf("got %+v, %v", got, err)
	}
	empty, _ := newFakeDB(t)
	if got, err := eventWorkflowNew(empty).GetEventWorkflowByID(ctx, "x"); got != nil || err != nil {
		t.Errorf("missing should be nil,nil; got %+v, %v", got, err)
	}
	bad, fb := newFakeDB(t)
	fb.onErr("SELECT", errEventWorkflowBoom)
	if got, err := eventWorkflowNew(bad).GetEventWorkflowByID(ctx, "x"); got != nil || !errors.Is(err, errEventWorkflowBoom) {
		t.Errorf("got %+v, %v", got, err)
	}
}

func TestEventWorkflowListQueries(t *testing.T) {
	ctx := context.Background()
	type listFn func(*EventWorkflowRepository) ([]models.EventWorkflow, error)
	cases := []struct {
		name      string
		call      listFn
		fragments []string
		arg       any
	}{
		{"by project", func(r *EventWorkflowRepository) ([]models.EventWorkflow, error) {
			return r.GetEventWorkflowsByProjectID(ctx, "p1")
		}, []string{`"ProjectId" =`}, "p1"},
		{"by project with elements", func(r *EventWorkflowRepository) ([]models.EventWorkflow, error) {
			return r.GetEventWorkflowsByProjectIDWithElements(ctx, "p1")
		}, []string{`"ProjectId" =`}, "p1"},
		{"enabled by project", func(r *EventWorkflowRepository) ([]models.EventWorkflow, error) {
			return r.GetEnabledEventWorkflowsByProjectID(ctx, "p1")
		}, []string{`"Enabled" =`}, true},
		{"by name", func(r *EventWorkflowRepository) ([]models.EventWorkflow, error) {
			return r.GetEventWorkflowsByName(ctx, "p1", "fl")
		}, []string{`"Name" ILIKE`}, "%fl%"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db, f := newFakeDB(t)
			f.on(eventWorkflowSelect, eventWorkflowCols, []driver.Value{"w1", "p1", "flow", true}, []driver.Value{"w2", "p1", "other", false})
			got, err := tc.call(eventWorkflowNew(db))
			if err != nil || len(got) != 2 || got[0].Id != "w1" || got[1].Name != "other" {
				t.Fatalf("got %+v, %v", got, err)
			}
			if !f.sawQuery(tc.fragments...) || !eventWorkflowArg(f, tc.arg) {
				t.Errorf("query %q args %v", f.queries[0], f.args[0])
			}
			bad, fb := newFakeDB(t)
			fb.onErr("SELECT", errEventWorkflowBoom)
			if got, err := tc.call(eventWorkflowNew(bad)); got != nil || !errors.Is(err, errEventWorkflowBoom) {
				t.Errorf("error case: %+v, %v", got, err)
			}
		})
	}
}

func TestEventWorkflowUpdate(t *testing.T) {
	ctx := context.Background()
	desc := "d"
	t.Run("includes json columns only when provided", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onExec("UPDATE", 1)
		f.on(eventWorkflowSelect, eventWorkflowCols, []driver.Value{"w1", "p1", "renamed", false})
		got, err := eventWorkflowNew(db).UpdateEventWorkflow(ctx, "w1", &models.EventWorkflow{
			Name: "renamed", Description: &desc, CanvasData: []byte(`{"a":1}`), Handlers: []byte(`[]`),
		})
		if err != nil || got == nil || got.Name != "renamed" {
			t.Fatalf("got %+v, %v", got, err)
		}
		if !f.sawQuery("UPDATE", `"CanvasData"`, `"Handlers"`, `"Enabled"`, `"Name"`) {
			t.Errorf("update: %v", f.queries)
		}

		db2, f2 := newFakeDB(t)
		f2.onExec("UPDATE", 1)
		if _, err := eventWorkflowNew(db2).UpdateEventWorkflow(ctx, "w1", &models.EventWorkflow{Name: "n"}); err != nil {
			t.Fatal(err)
		}
		if f2.sawQuery(`"CanvasData"`) || f2.sawQuery(`"Handlers"`) {
			t.Errorf("empty json must not be written: %q", f2.queries[0])
		}
	})
	t.Run("update error", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onErr("UPDATE", errEventWorkflowBoom)
		if got, err := eventWorkflowNew(db).UpdateEventWorkflow(ctx, "w1", &models.EventWorkflow{}); got != nil || !errors.Is(err, errEventWorkflowBoom) {
			t.Errorf("got %+v, %v", got, err)
		}
	})
}

func TestEventWorkflowEnabledDeleteAndCount(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.onExec("UPDATE", 1)
	f.onExec("DELETE FROM", 1)
	f.on("count(*)", []string{"count"}, []driver.Value{int64(4)})
	repo := eventWorkflowNew(db)

	if err := repo.UpdateEventWorkflowEnabled(ctx, "w1", true); err != nil {
		t.Fatal(err)
	}
	if !f.sawQuery("UPDATE", `"Enabled"`) || !eventWorkflowArg(f, true) {
		t.Errorf("enabled update: %v", f.queries)
	}
	if err := repo.DeleteEventWorkflow(ctx, "w1"); err != nil {
		t.Fatal(err)
	}
	if err := repo.DeleteEventWorkflowsByProjectID(ctx, "p1"); err != nil {
		t.Fatal(err)
	}
	if !f.sawQuery("DELETE FROM", `"Id" =`) || !f.sawQuery("DELETE FROM", `"ProjectId" =`) {
		t.Errorf("deletes: %v", f.queries)
	}
	if n, err := repo.CountEventWorkflowsByProjectID(ctx, "p1"); err != nil || n != 4 {
		t.Errorf("count = %d, %v", n, err)
	}

	bad, fb := newFakeDB(t)
	fb.onErr("UPDATE", errEventWorkflowBoom)
	fb.onErr("DELETE FROM", errEventWorkflowBoom)
	fb.onErr("count(*)", errEventWorkflowBoom)
	br := eventWorkflowNew(bad)
	if err := br.UpdateEventWorkflowEnabled(ctx, "w1", true); !errors.Is(err, errEventWorkflowBoom) {
		t.Errorf("enabled: %v", err)
	}
	if err := br.DeleteEventWorkflow(ctx, "w1"); !errors.Is(err, errEventWorkflowBoom) {
		t.Errorf("delete: %v", err)
	}
	if err := br.DeleteEventWorkflowsByProjectID(ctx, "p1"); !errors.Is(err, errEventWorkflowBoom) {
		t.Errorf("delete by project: %v", err)
	}
	if n, err := br.CountEventWorkflowsByProjectID(ctx, "p1"); n != 0 || !errors.Is(err, errEventWorkflowBoom) {
		t.Errorf("count: %d, %v", n, err)
	}
}

func TestEventWorkflowCheckIfNameExists(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		name    string
		n       int64
		exclude string
		want    bool
	}{
		{"taken", 1, "", true},
		{"free", 0, "", false},
		{"taken by another workflow", 2, "w1", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, f := newFakeDB(t)
			f.on("count(*)", []string{"count"}, []driver.Value{tc.n})
			got, err := eventWorkflowNew(db).CheckIfWorkflowNameExists(ctx, "p1", "flow", tc.exclude)
			if err != nil || got != tc.want {
				t.Errorf("got %v, %v", got, err)
			}
			if f.sawQuery(`"Id" !=`) != (tc.exclude != "") {
				t.Errorf("exclusion clause mismatch: %q", f.queries[0])
			}
		})
	}
	bad, fb := newFakeDB(t)
	fb.onErr("count(*)", errEventWorkflowBoom)
	if ok, err := eventWorkflowNew(bad).CheckIfWorkflowNameExists(ctx, "p1", "n", ""); ok || !errors.Is(err, errEventWorkflowBoom) {
		t.Errorf("got %v, %v", ok, err)
	}
}

func TestEventWorkflowFilters(t *testing.T) {
	ctx := context.Background()
	enabled := true
	t.Run("no filters", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on(eventWorkflowSelect, eventWorkflowCols, []driver.Value{"w1", "p1", "a", true})
		got, err := eventWorkflowNew(db).GetEventWorkflowsWithFilters(ctx, "", nil, "")
		if err != nil || len(got) != 1 {
			t.Fatalf("got %+v, %v", got, err)
		}
		if f.sawQuery(`"ProjectId" =`) || f.sawQuery(`"Enabled" =`) || f.sawQuery("ILIKE") {
			t.Errorf("unexpected filters: %q", f.queries[0])
		}
	})
	t.Run("all filters", func(t *testing.T) {
		db, f := newFakeDB(t)
		if _, err := eventWorkflowNew(db).GetEventWorkflowsWithFilters(ctx, "p1", &enabled, "fl"); err != nil {
			t.Fatal(err)
		}
		if !f.sawQuery(`"ProjectId" =`, `"Enabled" =`, `"Name" ILIKE`) || !eventWorkflowArg(f, "%fl%") || !eventWorkflowArg(f, "p1") || !eventWorkflowArg(f, true) {
			t.Errorf("query %q args %v", f.queries[0], f.args[0])
		}
	})
	t.Run("error", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onErr("SELECT", errEventWorkflowBoom)
		if got, err := eventWorkflowNew(db).GetEventWorkflowsWithFilters(ctx, "", nil, ""); got != nil || !errors.Is(err, errEventWorkflowBoom) {
			t.Errorf("got %+v, %v", got, err)
		}
	})
}
