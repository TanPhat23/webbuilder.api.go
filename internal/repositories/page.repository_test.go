package repositories

import (
	"context"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"my-go-app/internal/models"
)

// pageHasArg reports whether any recorded statement was sent with the given bind value.
func pageHasArg(f *fakeDB, want any) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, args := range f.args {
		for _, a := range args {
			if a.Value == want {
				return true
			}
		}
	}
	return false
}

var pageBoom = errors.New("boom")

var pageCols = []string{"Id", "ProjectId", "Name", "Type"}

func pageRow(id, project, name string) []driver.Value {
	return []driver.Value{id, project, name, "page"}
}

func TestPageRepositoryGetPagesByProjectID(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.on("SELECT", pageCols, pageRow("a", "p1", "Home"), pageRow("b", "p1", "About"))
	got, err := (&PageRepository{db: db}).GetPagesByProjectID(ctx, "p1")
	if err != nil || len(got) != 2 || got[0].Name != "Home" || got[1].Id != "b" {
		t.Fatalf("got %+v, %v", got, err)
	}
	if !f.sawQuery("DeletedAt", "IS NULL", "ORDER BY", "ASC") || !pageHasArg(f, "p1") {
		t.Fatalf("unexpected query: %v", f.queries)
	}

	db, f = newFakeDB(t)
	f.onErr("SELECT", pageBoom)
	got, err = (&PageRepository{db: db}).GetPagesByProjectID(ctx, "p1")
	if got != nil || !errors.Is(err, pageBoom) || !strings.HasPrefix(err.Error(), "failed to get pages by project ID") {
		t.Fatalf("got %+v, %v", got, err)
	}
}

func TestPageRepositoryGetPageByID(t *testing.T) {
	ctx := context.Background()
	t.Run("scoped to the project when given", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on("SELECT", pageCols, pageRow("a", "p1", "Home"))
		got, err := (&PageRepository{db: db}).GetPageByID(ctx, "a", "p1")
		if err != nil || got.Id != "a" || got.ProjectId != "p1" {
			t.Fatalf("got %+v, %v", got, err)
		}
		if !f.sawQuery("ProjectId") || !pageHasArg(f, "p1") {
			t.Fatalf("project filter missing: %v", f.queries)
		}
	})
	t.Run("without a project id only the page is matched", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on("SELECT", pageCols, pageRow("a", "p1", "Home"))
		if _, err := (&PageRepository{db: db}).GetPageByID(ctx, "a", ""); err != nil {
			t.Fatal(err)
		}
		if f.sawQuery("ProjectId") {
			t.Fatalf("unexpected project filter: %v", f.queries)
		}
	})
	t.Run("missing", func(t *testing.T) {
		db, _ := newFakeDB(t)
		if got, err := (&PageRepository{db: db}).GetPageByID(ctx, "a", "p1"); got != nil || !errors.Is(err, ErrPageNotFound) {
			t.Fatalf("got %+v, %v", got, err)
		}
	})
	t.Run("database error", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onErr("SELECT", pageBoom)
		_, err := (&PageRepository{db: db}).GetPageByID(ctx, "a", "p1")
		if !errors.Is(err, pageBoom) || errors.Is(err, ErrPageNotFound) || !strings.HasPrefix(err.Error(), "failed to get page") {
			t.Fatal(err)
		}
	})
}

func TestPageRepositoryCreatePage(t *testing.T) {
	ctx := context.Background()
	valid := func() *models.Page { return &models.Page{Id: "pg", ProjectId: "p1", Name: "Home", Type: "page"} }
	t.Run("rejects invalid pages before querying", func(t *testing.T) {
		db, f := newFakeDB(t)
		repo := &PageRepository{db: db}
		noName, noProject, noType := valid(), valid(), valid()
		noName.Name, noProject.ProjectId, noType.Type = "", "", ""
		for want, p := range map[string]*models.Page{
			"page name is required":  noName,
			"project ID is required": noProject,
			"page type is required":  noType,
		} {
			if err := repo.CreatePage(ctx, p); err == nil || err.Error() != want {
				t.Fatalf("want %q, got %v", want, err)
			}
		}
		if f.queryCount() != 0 {
			t.Fatalf("queried: %v", f.queries)
		}
	})
	t.Run("defaults styles and stamps timestamps", func(t *testing.T) {
		db, f := newFakeDB(t)
		p := valid()
		if err := (&PageRepository{db: db}).CreatePage(ctx, p); err != nil {
			t.Fatal(err)
		}
		if string(p.Styles) != "{}" || p.CreatedAt.IsZero() || !p.CreatedAt.Equal(p.UpdatedAt) {
			t.Fatalf("defaults not applied: %+v", p)
		}
		if !f.sawQuery("INSERT INTO", "Page") {
			t.Fatalf("missing insert: %v", f.queries)
		}
	})
	t.Run("keeps given styles", func(t *testing.T) {
		db, _ := newFakeDB(t)
		p := valid()
		p.Styles = []byte(`{"a":1}`)
		if err := (&PageRepository{db: db}).CreatePage(ctx, p); err != nil || string(p.Styles) != `{"a":1}` {
			t.Fatalf("got %s, %v", p.Styles, err)
		}
	})
	t.Run("duplicate name maps to ErrPageNameTaken", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onErr("INSERT", &pgconn.PgError{Code: "23505", ConstraintName: "Page_ProjectId_Name_key"})
		if err := (&PageRepository{db: db}).CreatePage(ctx, valid()); !errors.Is(err, ErrPageNameTaken) {
			t.Fatal(err)
		}
	})
	t.Run("other unique violations are wrapped, not mapped", func(t *testing.T) {
		db, f := newFakeDB(t)
		pgErr := &pgconn.PgError{Code: "23505", ConstraintName: "Page_pkey"}
		f.onErr("INSERT", pgErr)
		err := (&PageRepository{db: db}).CreatePage(ctx, valid())
		if errors.Is(err, ErrPageNameTaken) || !errors.Is(err, pgErr) || !strings.HasPrefix(err.Error(), "failed to create page") {
			t.Fatal(err)
		}
	})
	t.Run("generic error is wrapped", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onErr("INSERT", pageBoom)
		if err := (&PageRepository{db: db}).CreatePage(ctx, valid()); !errors.Is(err, pageBoom) {
			t.Fatal(err)
		}
	})
}

func TestPageRepositoryUpdatePage(t *testing.T) {
	ctx := context.Background()
	t.Run("updates and bumps UpdatedAt", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onExec("UPDATE", 1)
		p := &models.Page{Id: "pg", Name: "New"}
		if err := (&PageRepository{db: db}).UpdatePage(ctx, p); err != nil {
			t.Fatal(err)
		}
		if time.Since(p.UpdatedAt) > time.Minute || !f.sawQuery("UPDATE", "Page", "Name") || !pageHasArg(f, "New") {
			t.Fatalf("unexpected update: %v %v", f.queries, f.args)
		}
	})
	t.Run("no matching row", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onExec("UPDATE", 0)
		if err := (&PageRepository{db: db}).UpdatePage(ctx, &models.Page{Id: "pg", Name: "x"}); !errors.Is(err, ErrPageNotFound) {
			t.Fatal(err)
		}
	})
	t.Run("database error", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onErr("UPDATE", pageBoom)
		err := (&PageRepository{db: db}).UpdatePage(ctx, &models.Page{Id: "pg", Name: "x"})
		if !errors.Is(err, pageBoom) || !strings.HasPrefix(err.Error(), "failed to update page") {
			t.Fatal(err)
		}
	})
}

func TestPageRepositoryUpdatePageFields(t *testing.T) {
	ctx := context.Background()
	t.Run("applies the fields and stamps UpdatedAt", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onExec("UPDATE", 1)
		updates := map[string]any{"Name": "Renamed"}
		if err := (&PageRepository{db: db}).UpdatePageFields(ctx, "pg", updates); err != nil {
			t.Fatal(err)
		}
		if _, ok := updates["UpdatedAt"].(time.Time); !ok {
			t.Fatalf("UpdatedAt not stamped: %v", updates)
		}
		if !f.sawQuery("UPDATE", "Name", "UpdatedAt", "DeletedAt") || !pageHasArg(f, "Renamed") {
			t.Fatalf("unexpected update: %v", f.queries)
		}
	})
	t.Run("no matching row", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onExec("UPDATE", 0)
		if err := (&PageRepository{db: db}).UpdatePageFields(ctx, "pg", map[string]any{"Name": "x"}); !errors.Is(err, ErrPageNotFound) {
			t.Fatal(err)
		}
	})
	t.Run("database error", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onErr("UPDATE", pageBoom)
		err := (&PageRepository{db: db}).UpdatePageFields(ctx, "pg", map[string]any{"Name": "x"})
		if !errors.Is(err, pageBoom) || !strings.HasPrefix(err.Error(), "failed to update page fields") {
			t.Fatal(err)
		}
	})
}

// pageWrite runs the shared not-found / database-error matrix for the page statements.
func pageWrite(t *testing.T, verb string, call func(db *fakeDB, r *PageRepository) error, prefix string, setup func(f *fakeDB)) {
	t.Helper()
	t.Run("affects a row", func(t *testing.T) {
		db, f := newFakeDB(t)
		if setup != nil {
			setup(f)
		}
		f.onExec(verb, 1)
		if err := call(f, &PageRepository{db: db}); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("affects no row", func(t *testing.T) {
		db, f := newFakeDB(t)
		if setup != nil {
			setup(f)
		}
		f.onExec(verb, 0)
		if err := call(f, &PageRepository{db: db}); !errors.Is(err, ErrPageNotFound) {
			t.Fatal(err)
		}
	})
	t.Run("database error", func(t *testing.T) {
		db, f := newFakeDB(t)
		if setup != nil {
			setup(f)
		}
		f.onErr(verb, pageBoom)
		err := call(f, &PageRepository{db: db})
		if !errors.Is(err, pageBoom) || !strings.HasPrefix(err.Error(), prefix) {
			t.Fatal(err)
		}
	})
}

func TestPageRepositoryDeletePage(t *testing.T) {
	ctx := context.Background()
	pageWrite(t, "UPDATE", func(f *fakeDB, r *PageRepository) error {
		err := r.DeletePage(ctx, "pg")
		if !f.sawQuery("UPDATE", "DeletedAt") || !pageHasArg(f, "pg") {
			t.Errorf("soft delete expected: %v", f.queries)
		}
		return err
	}, "failed to delete page", nil)
}

func TestPageRepositoryHardDeletePage(t *testing.T) {
	ctx := context.Background()
	pageWrite(t, "DELETE", func(f *fakeDB, r *PageRepository) error {
		err := r.HardDeletePage(ctx, "pg")
		if !f.sawQuery("DELETE FROM", "Page") {
			t.Errorf("hard delete expected: %v", f.queries)
		}
		return err
	}, "failed to hard delete page", nil)
}

func TestPageRepositoryRestorePage(t *testing.T) {
	ctx := context.Background()
	pageWrite(t, "UPDATE", func(f *fakeDB, r *PageRepository) error {
		err := r.RestorePage(ctx, "pg")
		if !f.sawQuery("UPDATE", "DeletedAt", "NULL") {
			t.Errorf("restore expected: %v", f.queries)
		}
		return err
	}, "failed to restore page", nil)
}

func TestPageRepositoryFiltersDeletedPages(t *testing.T) {
	db, f := newFakeDB(t)
	f.onExec("UPDATE", 1)
	r := &PageRepository{db: db}
	if err := r.RestorePage(context.Background(), "pg"); err != nil {
		t.Fatal(err)
	}
	if !f.sawQuery(`"DeletedAt" IS NOT NULL`) {
		t.Fatalf("missing quoted predicate: %v", f.queries)
	}
	if _, err := r.CountPagesByProjectID(context.Background(), "p"); err != nil {
		t.Fatal(err)
	}
	if !f.sawQuery("count(", `"DeletedAt" IS NULL`) {
		t.Fatalf("missing active filter: %v", f.queries)
	}
}

func TestPageRepositoryDeletePageByProjectID(t *testing.T) {
	ctx := context.Background()
	ownerCount := func(n int64) func(f *fakeDB) {
		return func(f *fakeDB) {
			cols, rows := countRow(n)
			f.on("count(", cols, rows...)
		}
	}
	t.Run("owner deletes the page", func(t *testing.T) {
		db, f := newFakeDB(t)
		ownerCount(1)(f)
		f.onExec("UPDATE", 1)
		if err := (&PageRepository{db: db}).DeletePageByProjectID(ctx, "pg", "p1", "u1"); err != nil {
			t.Fatal(err)
		}
		if !pageHasArg(f, "u1") || !pageHasArg(f, "pg") || !f.sawQuery("UPDATE", "DeletedAt") {
			t.Fatalf("unexpected statements: %v %v", f.queries, f.args)
		}
	})
	t.Run("non owner is unauthorized and nothing is deleted", func(t *testing.T) {
		db, f := newFakeDB(t)
		ownerCount(0)(f)
		if err := (&PageRepository{db: db}).DeletePageByProjectID(ctx, "pg", "p1", "u2"); !errors.Is(err, ErrPageUnauthorized) {
			t.Fatal(err)
		}
		if f.sawQuery("UPDATE") {
			t.Fatalf("must not update: %v", f.queries)
		}
	})
	t.Run("ownership check error", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onErr("count(", pageBoom)
		err := (&PageRepository{db: db}).DeletePageByProjectID(ctx, "pg", "p1", "u1")
		if !errors.Is(err, pageBoom) || !strings.HasPrefix(err.Error(), "failed to verify project ownership") {
			t.Fatal(err)
		}
	})
	t.Run("page not in project", func(t *testing.T) {
		db, f := newFakeDB(t)
		ownerCount(1)(f)
		f.onExec("UPDATE", 0)
		if err := (&PageRepository{db: db}).DeletePageByProjectID(ctx, "pg", "p1", "u1"); !errors.Is(err, ErrPageNotFound) {
			t.Fatal(err)
		}
	})
	t.Run("update error", func(t *testing.T) {
		db, f := newFakeDB(t)
		ownerCount(1)(f)
		f.onErr("UPDATE", pageBoom)
		err := (&PageRepository{db: db}).DeletePageByProjectID(ctx, "pg", "p1", "u1")
		if !errors.Is(err, pageBoom) || !strings.HasPrefix(err.Error(), "failed to delete page") {
			t.Fatal(err)
		}
	})
}

func TestPageRepositoryExistsAndCount(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		n    int64
		want bool
	}{{0, false}, {1, true}, {3, true}} {
		db, f := newFakeDB(t)
		cols, rows := countRow(tc.n)
		f.on("count(", cols, rows...)
		got, err := (&PageRepository{db: db}).ExistsInProject(ctx, "pg", "p1")
		if err != nil || got != tc.want {
			t.Fatalf("count %d: got %v, %v", tc.n, got, err)
		}
		if !pageHasArg(f, "pg") || !pageHasArg(f, "p1") {
			t.Fatalf("binds missing: %v", f.args)
		}
	}
	db, f := newFakeDB(t)
	f.onErr("count(", pageBoom)
	got, err := (&PageRepository{db: db}).ExistsInProject(ctx, "pg", "p1")
	if got || !errors.Is(err, pageBoom) || !strings.HasPrefix(err.Error(), "failed to check page existence") {
		t.Fatalf("got %v, %v", got, err)
	}

	db, f = newFakeDB(t)
	cols, rows := countRow(7)
	f.on("count(", cols, rows...)
	n, err := (&PageRepository{db: db}).CountPagesByProjectID(ctx, "p1")
	if err != nil || n != 7 || !pageHasArg(f, "p1") {
		t.Fatalf("got %d, %v", n, err)
	}
	db, f = newFakeDB(t)
	f.onErr("count(", pageBoom)
	n, err = (&PageRepository{db: db}).CountPagesByProjectID(ctx, "p1")
	if n != 0 || !errors.Is(err, pageBoom) || !strings.HasPrefix(err.Error(), "failed to count pages") {
		t.Fatalf("got %d, %v", n, err)
	}
}
