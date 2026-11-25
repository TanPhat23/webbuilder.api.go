package repositories

import (
	"context"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"
	"time"

	"my-go-app/internal/models"
)

var (
	projectBoom = errors.New("boom")
	projectCols = []string{"Id", "Name", "OwnerId"}
)

func projectRow(id, owner string) []driver.Value { return []driver.Value{id, "Name " + id, owner} }

func projectHasArg(f *fakeDB, want any) bool {
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

func projectCount(f *fakeDB, n int64) *fakeRule {
	cols, rows := countRow(n)
	return f.on("count(", cols, rows...)
}

func assertProjectErr(t *testing.T, err, target error, prefix string) {
	t.Helper()
	if !errors.Is(err, target) {
		t.Fatalf("want error %v, got %v", target, err)
	}
	if prefix != "" && !strings.HasPrefix(err.Error(), prefix) {
		t.Fatalf("want prefix %q, got %v", prefix, err)
	}
}

func assertProjectArgs(t *testing.T, f *fakeDB, wants ...any) {
	t.Helper()
	for _, want := range wants {
		if !projectHasArg(f, want) {
			t.Fatalf("arg %v not bound: %v", want, f.args)
		}
	}
}

func assertProjectQueries(t *testing.T, f *fakeDB, fragments ...string) {
	t.Helper()
	if !f.sawQuery(fragments...) {
		t.Fatalf("missing query fragments %v in %v", fragments, f.queries)
	}
}

type projectLookupCase struct {
	name   string
	errMsg string
	call   func(r *ProjectRepository) (*models.Project, error)
}

func projectLookupCases(ctx context.Context) []projectLookupCase {
	return []projectLookupCase{
		{"owned", "failed to get project", func(r *ProjectRepository) (*models.Project, error) { return r.GetByID(ctx, "p1", "u1") }},
		{"public", "failed to get public project", func(r *ProjectRepository) (*models.Project, error) { return r.GetPublicByID(ctx, "p1") }},
		{"locked", "failed to get project with lock", func(r *ProjectRepository) (*models.Project, error) { return r.GetWithLock(ctx, "p1", "u1") }},
	}
}

func assertGotProject(t *testing.T, name string, got *models.Project, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: unexpected error %v", name, err)
	}
	if got == nil || got.ID != "p1" {
		t.Fatalf("%s: unexpected project %+v", name, got)
	}
	if got.OwnerId != "u1" || got.Name != "Name p1" {
		t.Fatalf("%s: unexpected owner/name %+v", name, got)
	}
}

func TestProjectRepositoryLookupFound(t *testing.T) {
	ctx := context.Background()
	for _, tc := range projectLookupCases(ctx) {
		db, f := newFakeDB(t)
		f.on("SELECT", projectCols, projectRow("p1", "u1"))
		got, err := tc.call(&ProjectRepository{db: db})
		assertGotProject(t, tc.name, got, err)
		assertProjectArgs(t, f, "p1")
	}
}

func TestProjectRepositoryLookupMissingAndError(t *testing.T) {
	ctx := context.Background()
	for _, tc := range projectLookupCases(ctx) {
		db, _ := newFakeDB(t)
		if got, err := tc.call(&ProjectRepository{db: db}); got != nil || !errors.Is(err, ErrProjectNotFound) {
			t.Fatalf("%s missing: got %+v, %v", tc.name, got, err)
		}
		dbErr, fErr := newFakeDB(t)
		fErr.onErr("SELECT", projectBoom)
		got, err := tc.call(&ProjectRepository{db: dbErr})
		if got != nil || errors.Is(err, ErrProjectNotFound) {
			t.Fatalf("%s error: got %+v, %v", tc.name, got, err)
		}
		assertProjectErr(t, err, projectBoom, tc.errMsg)
	}
}

func TestProjectRepositoryLookupQueryShapes(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.on("SELECT", projectCols, projectRow("p1", "u1"))
	if _, err := (&ProjectRepository{db: db}).GetByID(ctx, "p1", "u1"); err != nil {
		t.Fatal(err)
	}
	assertProjectQueries(t, f, "OwnerId", "DeletedAt", "IS NULL")
	assertProjectArgs(t, f, "u1")

	dbPub, fPub := newFakeDB(t)
	assertProjectNotFound(t, func() error { _, err := (&ProjectRepository{db: dbPub}).GetPublicByID(ctx, "p1"); return err })
	assertProjectQueries(t, fPub, "Published", "true")

	dbLock, fLock := newFakeDB(t)
	assertProjectNotFound(t, func() error { _, err := (&ProjectRepository{db: dbLock}).GetWithLock(ctx, "p1", "u1"); return err })
	assertProjectQueries(t, fLock, "FOR UPDATE")
}

func assertProjectNotFound(t *testing.T, call func() error) {
	t.Helper()
	assertProjectErr(t, call(), ErrProjectNotFound, "")
}

func TestProjectRepositoryGetProjectWithAccessSuccessAndMissing(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.on("SELECT", projectCols, projectRow("p1", "u9"))
	got, err := (&ProjectRepository{db: db}).GetWithAccess(ctx, "p1", "u1")
	if err != nil || got.ID != "p1" || f.queryCount() != 1 {
		t.Fatalf("got %+v, %v, queries=%v", got, err, f.queries)
	}

	dbCollab, fCollab := newFakeDB(t)
	fCollab.on(`FROM "public"."Collaborator"`, []string{"Id", "UserId", "ProjectId"}, []driver.Value{"c1", "u1", "p1"})
	gotCollab, err := (&ProjectRepository{db: dbCollab}).GetWithAccess(ctx, "p1", "u1")
	if gotCollab != nil || !errors.Is(err, ErrProjectNotFound) {
		t.Fatalf("got %+v, %v", gotCollab, err)
	}

	dbStranger, _ := newFakeDB(t)
	if gotStranger, err := (&ProjectRepository{db: dbStranger}).GetWithAccess(ctx, "p1", "u1"); gotStranger != nil || !errors.Is(err, ErrProjectUnauthorized) {
		t.Fatalf("got %+v, %v", gotStranger, err)
	}
}

func TestProjectRepositoryGetProjectWithAccessErrors(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.onErr("SELECT", projectBoom)
	_, err := (&ProjectRepository{db: db}).GetWithAccess(ctx, "p1", "u1")
	assertProjectErr(t, err, projectBoom, "failed to get project")
	if f.queryCount() != 1 {
		t.Fatalf("must not fall through to the collaborator check: %v", f.queries)
	}

	dbCollab, fCollab := newFakeDB(t)
	fCollab.onErr(`FROM "public"."Collaborator"`, projectBoom)
	_, err = (&ProjectRepository{db: dbCollab}).GetWithAccess(ctx, "p1", "u1")
	if errors.Is(err, ErrProjectUnauthorized) {
		t.Fatalf("must not mask collaborator query error: %v", err)
	}
	assertProjectErr(t, err, projectBoom, "failed to check collaborator access")
}

func TestProjectRepositoryListQueries(t *testing.T) {
	ctx := context.Background()
	lists := []struct {
		name   string
		errMsg string
		call   func(r *ProjectRepository) ([]models.Project, error)
	}{
		{"by user", "failed to get projects", func(r *ProjectRepository) ([]models.Project, error) { return r.ListByUser(ctx, "u1") }},
		{"collaborator", "failed to get collaborator projects", func(r *ProjectRepository) ([]models.Project, error) { return r.ListByCollaborator(ctx, "u1") }},
	}
	for _, tc := range lists {
		assertProjectListResult(t, tc.name, tc.call)
		db, f := newFakeDB(t)
		f.onErr("SELECT", projectBoom)
		got, err := tc.call(&ProjectRepository{db: db})
		if got != nil {
			t.Fatalf("%s error returned rows: %+v", tc.name, got)
		}
		assertProjectErr(t, err, projectBoom, tc.errMsg)
	}
}

func assertProjectListResult(t *testing.T, name string, call func(r *ProjectRepository) ([]models.Project, error)) {
	t.Helper()
	db, f := newFakeDB(t)
	f.on("SELECT", projectCols, projectRow("p1", "u1"), projectRow("p2", "u2"))
	got, err := call(&ProjectRepository{db: db})
	if err != nil {
		t.Fatalf("%s: unexpected error %v", name, err)
	}
	if len(got) != 2 || got[0].ID != "p1" || got[1].OwnerId != "u2" {
		t.Fatalf("%s: got %+v", name, got)
	}
	assertOmitsHeavyProjectColumns(t, f.lastQuery())
	assertProjectArgs(t, f, "u1")
	assertProjectQueries(t, f, "ORDER BY")
}

func assertOmitsHeavyProjectColumns(t *testing.T, q string) {
	t.Helper()
	selectPart := strings.SplitN(q, "FROM", 2)[0]
	for _, heavy := range []string{`"Logic"`, `"Styles"`, `"Header"`} {
		if strings.Contains(selectPart, heavy) {
			t.Fatalf("list must omit %s: %s", heavy, q)
		}
	}
}

func TestProjectRepositoryListQuerySQLDetails(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	if _, err := (&ProjectRepository{db: db}).ListByCollaborator(ctx, "u1"); err != nil {
		t.Fatal(err)
	}
	assertProjectQueries(t, f, "JOIN", "Collaborator", "UserId")

	dbOwned, fOwned := newFakeDB(t)
	if _, err := (&ProjectRepository{db: dbOwned}).ListByUser(ctx, "u1"); err != nil {
		t.Fatal(err)
	}
	assertProjectQueries(t, fOwned, "ORDER BY", "OwnerId", "DESC", "CreatedAt")
}

func TestProjectRepositoryGetProjectPages(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.on(`FROM "public"."Project"`, projectCols, projectRow("p1", "u1"))
	f.on(`FROM "public"."Page"`, []string{"Id", "ProjectId", "Name"}, []driver.Value{"a", "p1", "Home"}, []driver.Value{"b", "p1", "About"})
	got, err := (&ProjectRepository{db: db}).ListPages(ctx, "p1", "u1")
	if err != nil || len(got) != 2 || got[0].Name != "Home" {
		t.Fatalf("got %+v, %v", got, err)
	}

	dbUnauth, fUnauth := newFakeDB(t)
	assertProjectErr(t, func() error { _, err := (&ProjectRepository{db: dbUnauth}).ListPages(ctx, "p1", "u1"); return err }(), ErrProjectUnauthorized, "")
	if fUnauth.sawQuery(`"public"."Page"`) {
		t.Fatalf("pages must not be read: %v", fUnauth.queries)
	}

	dbErr, fErr := newFakeDB(t)
	fErr.on(`FROM "public"."Project"`, projectCols, projectRow("p1", "u1"))
	fErr.onErr(`FROM "public"."Page"`, projectBoom)
	pages, err := (&ProjectRepository{db: dbErr}).ListPages(ctx, "p1", "u1")
	if pages != nil {
		t.Fatalf("unexpected pages: %+v", pages)
	}
	assertProjectErr(t, err, projectBoom, "failed to get project pages")
}

func TestProjectRepositoryCreateProject(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	p := &models.Project{ID: "p1", Name: "N", OwnerId: "u1", AuditFields: &models.AuditFields{}}
	if err := (&ProjectRepository{db: db}).Create(ctx, p); err != nil {
		t.Fatal(err)
	}
	if p.CreatedAt.IsZero() || !p.CreatedAt.Equal(p.UpdatedAt) {
		t.Fatalf("timestamps not stamped: %+v", p)
	}
	assertProjectQueries(t, f, "INSERT INTO", "Project")
	assertProjectArgs(t, f, "p1")

	dbErr, fErr := newFakeDB(t)
	fErr.onErr("INSERT", projectBoom)
	err := (&ProjectRepository{db: dbErr}).Create(ctx, &models.Project{ID: "p1", AuditFields: &models.AuditFields{}})
	assertProjectErr(t, err, projectBoom, "failed to create project")
}

func TestProjectRepositoryUpdateProjectSuccess(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	projectCount(f, 1)
	f.on("SELECT *", projectCols, projectRow("p1", "u1"))
	f.onExec("UPDATE", 1)
	updates := map[string]any{"Name": "Renamed"}
	got, err := (&ProjectRepository{db: db}).Update(ctx, "p1", "u1", updates)
	if err != nil || got.ID != "p1" {
		t.Fatalf("got %+v, %v", got, err)
	}
	if _, ok := updates["UpdatedAt"].(time.Time); !ok {
		t.Fatalf("UpdatedAt not stamped: %v", updates)
	}
	assertProjectQueries(t, f, "UPDATE", "Name", "UpdatedAt", "OwnerId")
	assertProjectArgs(t, f, "Renamed")
}

func TestProjectRepositoryUpdateProjectErrors(t *testing.T) {
	ctx := context.Background()
	dbUnauth, fUnauth := newFakeDB(t)
	projectCount(fUnauth, 0)
	got, err := (&ProjectRepository{db: dbUnauth}).Update(ctx, "p1", "u2", map[string]any{"Name": "x"})
	if got != nil || fUnauth.sawQuery("UPDATE") {
		t.Fatalf("got %+v, %v", got, fUnauth.queries)
	}
	assertProjectErr(t, err, ErrProjectUnauthorized, "")

	for _, tc := range []struct {
		rule   func(*fakeDB)
		want   error
		prefix string
	}{
		{func(f *fakeDB) { f.onErr("count(", projectBoom) }, projectBoom, "failed to verify project ownership"},
		{func(f *fakeDB) { projectCount(f, 1); f.onErr("UPDATE", projectBoom) }, projectBoom, "failed to update project"},
		{func(f *fakeDB) { projectCount(f, 1); f.onExec("UPDATE", 1) }, ErrProjectNotFound, ""},
	} {
		db, f := newFakeDB(t)
		tc.rule(f)
		_, err := (&ProjectRepository{db: db}).Update(ctx, "p1", "u1", map[string]any{"Name": "x"})
		assertProjectErr(t, err, tc.want, tc.prefix)
	}
}

func TestProjectRepositoryDeleteProject(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	projectCount(f, 1)
	f.onExec("UPDATE", 1)
	if err := (&ProjectRepository{db: db}).Delete(ctx, "p1", "u1"); err != nil {
		t.Fatal(err)
	}
	assertProjectQueries(t, f, "UPDATE", "DeletedAt")
	if f.sawQuery("DELETE FROM") {
		t.Fatalf("expected a soft delete: %v", f.queries)
	}

	dbUnauth, fUnauth := newFakeDB(t)
	projectCount(fUnauth, 0)
	err := (&ProjectRepository{db: dbUnauth}).Delete(ctx, "p1", "u2")
	assertProjectErr(t, err, ErrProjectUnauthorized, "")
	if fUnauth.sawQuery("UPDATE") {
		t.Fatalf("must not update: %v", fUnauth.queries)
	}
}

func TestProjectRepositoryDeleteProjectErrors(t *testing.T) {
	ctx := context.Background()
	db1, f1 := newFakeDB(t)
	f1.onErr("count(", projectBoom)
	assertProjectErr(t, (&ProjectRepository{db: db1}).Delete(ctx, "p1", "u1"), projectBoom, "failed to verify project ownership")

	db2, f2 := newFakeDB(t)
	projectCount(f2, 1)
	f2.onErr("UPDATE", projectBoom)
	assertProjectErr(t, (&ProjectRepository{db: db2}).Delete(ctx, "p1", "u1"), projectBoom, "failed to delete project")
}

func TestProjectRepositoryHardDeleteProject(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		affected int64
		dbErr    error
		wantErr  error
	}{
		{1, nil, nil},
		{0, nil, ErrProjectUnauthorized},
		{0, projectBoom, projectBoom},
	}
	for _, tc := range tests {
		db, f := newFakeDB(t)
		if tc.dbErr != nil {
			f.onErr("DELETE", tc.dbErr)
		} else {
			f.onExec("DELETE", tc.affected)
		}
		assertProjectErr(t, (&ProjectRepository{db: db}).HardDelete(ctx, "p1", "u1"), tc.wantErr, "")
		assertProjectQueries(t, f, "DELETE FROM", "Project")
		assertProjectArgs(t, f, "u1")
	}
}

func TestProjectRepositoryRestoreProject(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		affected int64
		dbErr    error
		wantErr  error
		prefix   string
	}{
		{1, nil, nil, ""},
		{0, nil, ErrProjectNotFound, ""},
		{0, projectBoom, projectBoom, "failed to restore project"},
	}
	for _, tc := range tests {
		db, f := newFakeDB(t)
		if tc.dbErr != nil {
			f.onErr("UPDATE", tc.dbErr)
		} else {
			f.onExec("UPDATE", tc.affected)
		}
		err := (&ProjectRepository{db: db}).Restore(ctx, "p1", "u1")
		assertProjectErr(t, err, tc.wantErr, tc.prefix)
	}
}

func TestProjectRepositoryExistsForUser(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		n    int64
		want bool
	}{{0, false}, {1, true}} {
		db, f := newFakeDB(t)
		projectCount(f, tc.n)
		got, err := (&ProjectRepository{db: db}).ExistsForUser(ctx, "p1", "u1")
		if err != nil || got != tc.want {
			t.Fatalf("count %d: %v %v", tc.n, got, err)
		}
	}
	db, f := newFakeDB(t)
	f.onErr("count(", projectBoom)
	got, err := (&ProjectRepository{db: db}).ExistsForUser(ctx, "p1", "u1")
	if got {
		t.Fatal("expected false on error")
	}
	assertProjectErr(t, err, projectBoom, "failed to check project existence")
}

func newValidProjectAndPage() (*models.Project, *models.Page) {
	return &models.Project{ID: "p1", Name: "N", OwnerId: "u1"}, &models.Page{Id: "pg", Name: "Home", Type: "page"}
}

func TestProjectRepositoryCreateProjectWithPageSuccess(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	p, pg := newValidProjectAndPage()
	if err := (&ProjectRepository{db: db}).CreateWithPage(ctx, p, pg); err != nil {
		t.Fatal(err)
	}
	assertSingleCommittedTx(t, f)
	if p.AuditFields == nil || p.CreatedAt.IsZero() {
		t.Fatalf("not stamped: %+v", p)
	}
	if !p.CreatedAt.Equal(pg.CreatedAt) || pg.ProjectId != "p1" {
		t.Fatalf("not linked: %+v %+v", p, pg)
	}
	assertProjectQueries(t, f, "INSERT INTO", "Project")
	assertProjectQueries(t, f, "INSERT INTO", "Page")
}

func assertSingleCommittedTx(t *testing.T, f *fakeDB) {
	t.Helper()
	if f.begins != 1 || f.commits != 1 || f.aborts != 0 {
		t.Fatalf("tx begins=%d commits=%d aborts=%d", f.begins, f.commits, f.aborts)
	}
}

func TestProjectRepositoryCreateProjectWithPageRollbacks(t *testing.T) {
	ctx := context.Background()
	db1, f1 := newFakeDB(t)
	f1.onErr(`INSERT INTO "public"."Project"`, projectBoom)
	p1, pg1 := newValidProjectAndPage()
	assertProjectErr(t, (&ProjectRepository{db: db1}).CreateWithPage(ctx, p1, pg1), projectBoom, "failed to create project")
	if f1.aborts != 1 || f1.commits != 0 || f1.sawQuery(`INSERT INTO "public"."Page"`) {
		t.Fatalf("aborts=%d commits=%d %v", f1.aborts, f1.commits, f1.queries)
	}

	db2, f2 := newFakeDB(t)
	f2.onErr(`INSERT INTO "public"."Page"`, projectBoom)
	p2, pg2 := newValidProjectAndPage()
	assertProjectErr(t, (&ProjectRepository{db: db2}).CreateWithPage(ctx, p2, pg2), projectBoom, "failed to create first page")
	if f2.aborts != 1 || f2.commits != 0 {
		t.Fatalf("aborts=%d commits=%d", f2.aborts, f2.commits)
	}
}

func TestProjectRepositoryCreateProjectWithPageValidation(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	repo := &ProjectRepository{db: db}
	if err := repo.CreateWithPage(ctx, nil, &models.Page{}); err == nil || err.Error() != "project and page are required" {
		t.Fatal(err)
	}
	for _, mutate := range []func(*models.Project){
		func(p *models.Project) { p.ID = "" },
		func(p *models.Project) { p.Name = "" },
		func(p *models.Project) { p.OwnerId = "" },
	} {
		p, pg := newValidProjectAndPage()
		mutate(p)
		if err := repo.CreateWithPage(ctx, p, pg); err == nil || err.Error() != "project id, name and owner are required" {
			t.Fatal(err)
		}
	}
	if f.begins != 0 || f.queryCount() != 0 {
		t.Fatalf("invalid input must not start a transaction: %v", f.queries)
	}
}

func TestProjectRepositoryGetProjectAccessRoles(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.on(`FROM "public"."Project"`, []string{"Id", "OwnerId"}, []driver.Value{"p1", "u1"})
	got, err := (&ProjectRepository{db: db}).GetAccess(ctx, "p1", "u1")
	assertOwnerAccessResult(t, got, err)
	if f.sawQuery(`"public"."Collaborator"`) {
		t.Fatalf("owner should not query collaborators: %v", f.queries)
	}

	dbC, fC := newFakeDB(t)
	fC.on(`FROM "public"."Project"`, []string{"Id", "OwnerId"}, []driver.Value{"p1", "u9"})
	fC.on(`FROM "public"."Collaborator"`, []string{"Role"}, []driver.Value{"viewer"})
	gotC, err := (&ProjectRepository{db: dbC}).GetAccess(ctx, "p1", "u1")
	assertCollaboratorAccessResult(t, gotC, err)
}

func assertOwnerAccessResult(t *testing.T, got *models.ProjectAccess, err error) {
	t.Helper()
	if err != nil || got == nil {
		t.Fatalf("got %+v, %v", got, err)
	}
	if !got.IsOwner || got.IsCollaborator || got.Role != models.RoleOwner {
		t.Fatalf("unexpected role flags: %+v", got)
	}
	if got.UserId != "u1" || got.ProjectId != "p1" {
		t.Fatalf("unexpected ids: %+v", got)
	}
}

func assertCollaboratorAccessResult(t *testing.T, got *models.ProjectAccess, err error) {
	t.Helper()
	if err != nil || got == nil {
		t.Fatalf("got %+v, %v", got, err)
	}
	if got.IsOwner || !got.IsCollaborator || got.Role != models.CollaboratorRole("viewer") {
		t.Fatalf("unexpected collaborator access: %+v", got)
	}
}

func TestProjectRepositoryGetProjectAccessNotFoundAndErrors(t *testing.T) {
	ctx := context.Background()
	dbMissing, _ := newFakeDB(t)
	assertProjectNotFound(t, func() error { _, err := (&ProjectRepository{db: dbMissing}).GetAccess(ctx, "p1", "u1"); return err })

	dbNonCollab, fNonCollab := newFakeDB(t)
	fNonCollab.on(`FROM "public"."Project"`, []string{"Id", "OwnerId"}, []driver.Value{"p1", "u9"})
	assertProjectNotFound(t, func() error {
		_, err := (&ProjectRepository{db: dbNonCollab}).GetAccess(ctx, "p1", "u1")
		return err
	})

	dbProjErr, fProjErr := newFakeDB(t)
	fProjErr.onErr(`FROM "public"."Project"`, projectBoom)
	_, err := (&ProjectRepository{db: dbProjErr}).GetAccess(ctx, "p1", "u1")
	assertProjectErr(t, err, projectBoom, "failed to get project")

	dbCollabErr, fCollabErr := newFakeDB(t)
	fCollabErr.on(`FROM "public"."Project"`, []string{"Id", "OwnerId"}, []driver.Value{"p1", "u9"})
	fCollabErr.onErr(`FROM "public"."Collaborator"`, projectBoom)
	_, err = (&ProjectRepository{db: dbCollabErr}).GetAccess(ctx, "p1", "u1")
	assertProjectErr(t, err, projectBoom, "failed to get collaborator")
}
