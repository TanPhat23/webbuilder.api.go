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

var (
	pageBoom = errors.New("boom")
	pageCols = []string{"Id", "ProjectId", "Name", "Type"}
)

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

func pageRow(id, project, name string) []driver.Value {
	return []driver.Value{id, project, name, "page"}
}

func assertPageNoErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func assertPageErrPrefix(t *testing.T, err, want error, prefix string) {
	t.Helper()
	if !errors.Is(err, want) {
		t.Fatalf("want %v, got %v", want, err)
	}
	if prefix != "" && !strings.HasPrefix(err.Error(), prefix) {
		t.Fatalf("want prefix %q, got %v", prefix, err)
	}
}

func TestPageRepositoryGetPagesByProjectID(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.on("SELECT", pageCols, pageRow("a", "p1", "Home"), pageRow("b", "p1", "About"))
	got, err := (&PageRepository{db: db}).GetPagesByProjectID(ctx, "p1")
	assertPageNoErr(t, err)
	if len(got) != 2 || got[0].Name != "Home" || got[1].Id != "b" {
		t.Fatalf("got %+v", got)
	}
	if !f.sawQuery("DeletedAt", "IS NULL", "ORDER BY", "ASC") || !pageHasArg(f, "p1") {
		t.Fatalf("unexpected query: %v", f.queries)
	}

	dbErr, fErr := newFakeDB(t)
	fErr.onErr("SELECT", pageBoom)
	_, err = (&PageRepository{db: dbErr}).GetPagesByProjectID(ctx, "p1")
	assertPageErrPrefix(t, err, pageBoom, "failed to get pages by project ID")
}

func TestPageRepositoryGetPageByID(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.on("SELECT", pageCols, pageRow("a", "p1", "Home"))
	got, err := (&PageRepository{db: db}).GetPageByID(ctx, "a", "p1")
	assertPageNoErr(t, err)
	if got.Id != "a" || got.ProjectId != "p1" || !f.sawQuery("ProjectId") || !pageHasArg(f, "p1") {
		t.Fatalf("got=%+v queries=%v", got, f.queries)
	}

	dbNoProj, fNoProj := newFakeDB(t)
	fNoProj.on("SELECT", pageCols, pageRow("a", "p1", "Home"))
	_, err = (&PageRepository{db: dbNoProj}).GetPageByID(ctx, "a", "")
	assertPageNoErr(t, err)
	if fNoProj.sawQuery("ProjectId") {
		t.Fatalf("unexpected project filter: %v", fNoProj.queries)
	}

	dbMiss, _ := newFakeDB(t)
	_, err = (&PageRepository{db: dbMiss}).GetPageByID(ctx, "a", "p1")
	assertPageErrPrefix(t, err, ErrPageNotFound, "")

	dbErr, fErr := newFakeDB(t)
	fErr.onErr("SELECT", pageBoom)
	_, err = (&PageRepository{db: dbErr}).GetPageByID(ctx, "a", "p1")
	assertPageErrPrefix(t, err, pageBoom, "failed to get page")
}

func validTestPage() *models.Page {
	return &models.Page{Id: "pg", ProjectId: "p1", Name: "Home", Type: "page"}
}

func TestPageRepositoryCreatePageValidation(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	repo := &PageRepository{db: db}
	noName, noProject, noType := validTestPage(), validTestPage(), validTestPage()
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
}

func TestPageRepositoryCreatePageSuccess(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	p := validTestPage()
	assertPageNoErr(t, (&PageRepository{db: db}).CreatePage(ctx, p))
	if string(p.Styles) != "{}" || p.CreatedAt.IsZero() || !p.CreatedAt.Equal(p.UpdatedAt) || !f.sawQuery("INSERT INTO", "Page") {
		t.Fatalf("defaults/insert missing: %+v %v", p, f.queries)
	}

	dbStyled, _ := newFakeDB(t)
	styled := validTestPage()
	styled.Styles = []byte(`{"a":1}`)
	assertPageNoErr(t, (&PageRepository{db: dbStyled}).CreatePage(ctx, styled))
	if string(styled.Styles) != `{"a":1}` {
		t.Fatalf("got %s", styled.Styles)
	}
}

func TestPageRepositoryCreatePageErrors(t *testing.T) {
	ctx := context.Background()
	dbDup, fDup := newFakeDB(t)
	fDup.onErr("INSERT", &pgconn.PgError{Code: "23505", ConstraintName: "Page_ProjectId_Name_key"})
	assertPageErrPrefix(t, (&PageRepository{db: dbDup}).CreatePage(ctx, validTestPage()), ErrPageNameTaken, "")

	dbOther, fOther := newFakeDB(t)
	pgErr := &pgconn.PgError{Code: "23505", ConstraintName: "Page_pkey"}
	fOther.onErr("INSERT", pgErr)
	err := (&PageRepository{db: dbOther}).CreatePage(ctx, validTestPage())
	if errors.Is(err, ErrPageNameTaken) {
		t.Fatalf("must not map other constraint: %v", err)
	}
	assertPageErrPrefix(t, err, pgErr, "failed to create page")

	dbErr, fErr := newFakeDB(t)
	fErr.onErr("INSERT", pageBoom)
	assertPageErrPrefix(t, (&PageRepository{db: dbErr}).CreatePage(ctx, validTestPage()), pageBoom, "")
}

func TestPageRepositoryUpdatePage(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.onExec("UPDATE", 1)
	p := &models.Page{Id: "pg", Name: "New"}
	assertPageNoErr(t, (&PageRepository{db: db}).UpdatePage(ctx, p))
	if time.Since(p.UpdatedAt) > time.Minute || !f.sawQuery("UPDATE", "Page", "Name") || !pageHasArg(f, "New") {
		t.Fatalf("unexpected update: %v %v", f.queries, f.args)
	}

	dbMiss, fMiss := newFakeDB(t)
	fMiss.onExec("UPDATE", 0)
	assertPageErrPrefix(t, (&PageRepository{db: dbMiss}).UpdatePage(ctx, &models.Page{Id: "pg", Name: "x"}), ErrPageNotFound, "")

	dbErr, fErr := newFakeDB(t)
	fErr.onErr("UPDATE", pageBoom)
	assertPageErrPrefix(t, (&PageRepository{db: dbErr}).UpdatePage(ctx, &models.Page{Id: "pg", Name: "x"}), pageBoom, "failed to update page")
}

func TestPageRepositoryUpdatePageFields(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.onExec("UPDATE", 1)
	updates := map[string]any{"Name": "Renamed"}
	assertPageNoErr(t, (&PageRepository{db: db}).UpdatePageFields(ctx, "pg", updates))
	if _, ok := updates["UpdatedAt"].(time.Time); !ok {
		t.Fatalf("UpdatedAt not stamped: %v", updates)
	}
	if !f.sawQuery("UPDATE", "Name", "UpdatedAt", "DeletedAt") || !pageHasArg(f, "Renamed") {
		t.Fatalf("unexpected update: %v", f.queries)
	}

	dbMiss, fMiss := newFakeDB(t)
	fMiss.onExec("UPDATE", 0)
	assertPageErrPrefix(t, (&PageRepository{db: dbMiss}).UpdatePageFields(ctx, "pg", map[string]any{"Name": "x"}), ErrPageNotFound, "")

	dbErr, fErr := newFakeDB(t)
	fErr.onErr("UPDATE", pageBoom)
	assertPageErrPrefix(t, (&PageRepository{db: dbErr}).UpdatePageFields(ctx, "pg", map[string]any{"Name": "x"}), pageBoom, "failed to update page fields")
}

// pageWrite runs the shared not-found / database-error matrix for the page statements.
func pageWrite(t *testing.T, verb string, call func(db *fakeDB, r *PageRepository) error, prefix string) {
	t.Helper()
	dbOk, fOk := newFakeDB(t)
	fOk.onExec(verb, 1)
	assertPageNoErr(t, call(fOk, &PageRepository{db: dbOk}))

	dbMiss, fMiss := newFakeDB(t)
	fMiss.onExec(verb, 0)
	assertPageErrPrefix(t, call(fMiss, &PageRepository{db: dbMiss}), ErrPageNotFound, "")

	dbErr, fErr := newFakeDB(t)
	fErr.onErr(verb, pageBoom)
	assertPageErrPrefix(t, call(fErr, &PageRepository{db: dbErr}), pageBoom, prefix)
}

func TestPageRepositoryDeleteHardDeleteAndRestore(t *testing.T) {
	ctx := context.Background()
	pageWrite(t, "UPDATE", func(f *fakeDB, r *PageRepository) error {
		err := r.DeletePage(ctx, "pg")
		if !f.sawQuery("UPDATE", "DeletedAt") || !pageHasArg(f, "pg") {
			t.Errorf("soft delete expected: %v", f.queries)
		}
		return err
	}, "failed to delete page")
	pageWrite(t, "DELETE", func(f *fakeDB, r *PageRepository) error {
		err := r.HardDeletePage(ctx, "pg")
		if !f.sawQuery("DELETE FROM", "Page") {
			t.Errorf("hard delete expected: %v", f.queries)
		}
		return err
	}, "failed to hard delete page")
	pageWrite(t, "UPDATE", func(f *fakeDB, r *PageRepository) error {
		err := r.RestorePage(ctx, "pg")
		if !f.sawQuery("UPDATE", "DeletedAt", "NULL") {
			t.Errorf("restore expected: %v", f.queries)
		}
		return err
	}, "failed to restore page")
}

func TestPageRepositoryFiltersDeletedPages(t *testing.T) {
	db, f := newFakeDB(t)
	f.onExec("UPDATE", 1)
	r := &PageRepository{db: db}
	assertPageNoErr(t, r.RestorePage(context.Background(), "pg"))
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

func setPageOwnerCount(f *fakeDB, n int64) {
	cols, rows := countRow(n)
	f.on("count(", cols, rows...)
}

func TestPageRepositoryDeletePageByProjectID(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	setPageOwnerCount(f, 1)
	f.onExec("UPDATE", 1)
	assertPageNoErr(t, (&PageRepository{db: db}).DeletePageByProjectID(ctx, "pg", "p1", "u1"))
	if !pageHasArg(f, "u1") || !pageHasArg(f, "pg") || !f.sawQuery("UPDATE", "DeletedAt") {
		t.Fatalf("unexpected statements: %v %v", f.queries, f.args)
	}

	dbUnauth, fUnauth := newFakeDB(t)
	setPageOwnerCount(fUnauth, 0)
	assertPageErrPrefix(t, (&PageRepository{db: dbUnauth}).DeletePageByProjectID(ctx, "pg", "p1", "u2"), ErrPageUnauthorized, "")
	if fUnauth.sawQuery("UPDATE") {
		t.Fatalf("must not update: %v", fUnauth.queries)
	}
}

func TestPageRepositoryDeletePageByProjectIDErrors(t *testing.T) {
	ctx := context.Background()
	db1, f1 := newFakeDB(t)
	f1.onErr("count(", pageBoom)
	assertPageErrPrefix(t, (&PageRepository{db: db1}).DeletePageByProjectID(ctx, "pg", "p1", "u1"), pageBoom, "failed to verify project ownership")

	db2, f2 := newFakeDB(t)
	setPageOwnerCount(f2, 1)
	f2.onExec("UPDATE", 0)
	assertPageErrPrefix(t, (&PageRepository{db: db2}).DeletePageByProjectID(ctx, "pg", "p1", "u1"), ErrPageNotFound, "")

	db3, f3 := newFakeDB(t)
	setPageOwnerCount(f3, 1)
	f3.onErr("UPDATE", pageBoom)
	assertPageErrPrefix(t, (&PageRepository{db: db3}).DeletePageByProjectID(ctx, "pg", "p1", "u1"), pageBoom, "failed to delete page")
}

func TestPageRepositoryExistsAndCount(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		n    int64
		want bool
	}{{0, false}, {1, true}, {3, true}} {
		db, f := newFakeDB(t)
		setPageOwnerCount(f, tc.n)
		got, err := (&PageRepository{db: db}).ExistsInProject(ctx, "pg", "p1")
		assertPageNoErr(t, err)
		if got != tc.want || !pageHasArg(f, "pg") || !pageHasArg(f, "p1") {
			t.Fatalf("count %d: got %v args=%v", tc.n, got, f.args)
		}
	}
	dbErr, fErr := newFakeDB(t)
	fErr.onErr("count(", pageBoom)
	_, err := (&PageRepository{db: dbErr}).ExistsInProject(ctx, "pg", "p1")
	assertPageErrPrefix(t, err, pageBoom, "failed to check page existence")

	dbCnt, fCnt := newFakeDB(t)
	setPageOwnerCount(fCnt, 7)
	n, err := (&PageRepository{db: dbCnt}).CountPagesByProjectID(ctx, "p1")
	assertPageNoErr(t, err)
	if n != 7 || !pageHasArg(fCnt, "p1") {
		t.Fatalf("got %d", n)
	}
	_, err = (&PageRepository{db: dbErr}).CountPagesByProjectID(ctx, "p1")
	assertPageErrPrefix(t, err, pageBoom, "failed to count pages")
}
