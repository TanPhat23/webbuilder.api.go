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

// projectHasArg reports whether any recorded statement was sent with the given bind value.
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

var projectBoom = errors.New("boom")

var projectCols = []string{"Id", "Name", "OwnerId"}

func projectRow(id, owner string) []driver.Value { return []driver.Value{id, "Name " + id, owner} }

func projectCount(f *fakeDB, n int64) *fakeRule {
	cols, rows := countRow(n)
	return f.on("count(", cols, rows...)
}

func TestProjectRepositoryGetProjectByIDAndPublic(t *testing.T) {
	ctx := context.Background()
	calls := []struct {
		name   string
		errMsg string
		call   func(r *ProjectRepository) (*models.Project, error)
	}{
		{"owned", "failed to get project", func(r *ProjectRepository) (*models.Project, error) { return r.GetProjectByID(ctx, "p1", "u1") }},
		{"public", "failed to get public project", func(r *ProjectRepository) (*models.Project, error) { return r.GetPublicProjectByID(ctx, "p1") }},
		{"locked", "failed to get project with lock", func(r *ProjectRepository) (*models.Project, error) { return r.GetProjectWithLock(ctx, "p1", "u1") }},
	}
	for _, tc := range calls {
		t.Run(tc.name+" found", func(t *testing.T) {
			db, f := newFakeDB(t)
			f.on("SELECT", projectCols, projectRow("p1", "u1"))
			got, err := tc.call(&ProjectRepository{db: db})
			if err != nil || got.ID != "p1" || got.OwnerId != "u1" || got.Name != "Name p1" {
				t.Fatalf("got %+v, %v", got, err)
			}
			if !projectHasArg(f, "p1") {
				t.Fatalf("project id not bound: %v", f.args)
			}
		})
		t.Run(tc.name+" missing", func(t *testing.T) {
			db, _ := newFakeDB(t)
			if got, err := tc.call(&ProjectRepository{db: db}); got != nil || !errors.Is(err, ErrProjectNotFound) {
				t.Fatalf("got %+v, %v", got, err)
			}
		})
		t.Run(tc.name+" error", func(t *testing.T) {
			db, f := newFakeDB(t)
			f.onErr("SELECT", projectBoom)
			got, err := tc.call(&ProjectRepository{db: db})
			if got != nil || !errors.Is(err, projectBoom) || errors.Is(err, ErrProjectNotFound) || !strings.HasPrefix(err.Error(), tc.errMsg) {
				t.Fatalf("got %+v, %v", got, err)
			}
		})
	}
	t.Run("owner scoping", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on("SELECT", projectCols, projectRow("p1", "u1"))
		if _, err := (&ProjectRepository{db: db}).GetProjectByID(ctx, "p1", "u1"); err != nil {
			t.Fatal(err)
		}
		if !f.sawQuery("OwnerId", "DeletedAt", "IS NULL") || !projectHasArg(f, "u1") {
			t.Fatalf("query must scope by owner: %v", f.queries)
		}
	})
	t.Run("public only published", func(t *testing.T) {
		db, f := newFakeDB(t)
		_, _ = (&ProjectRepository{db: db}).GetPublicProjectByID(ctx, "p1")
		if !f.sawQuery("Published", "true") {
			t.Fatalf("query must require Published: %v", f.queries)
		}
	})
	t.Run("lock uses FOR UPDATE", func(t *testing.T) {
		db, f := newFakeDB(t)
		_, _ = (&ProjectRepository{db: db}).GetProjectWithLock(ctx, "p1", "u1")
		if !f.sawQuery("FOR UPDATE") {
			t.Fatalf("expected row lock: %v", f.queries)
		}
	})
}

func TestProjectRepositoryGetProjectWithAccess(t *testing.T) {
	ctx := context.Background()
	t.Run("owner or collaborator is one query", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on("SELECT", projectCols, projectRow("p1", "u9"))
		got, err := (&ProjectRepository{db: db}).GetProjectWithAccess(ctx, "p1", "u1")
		if err != nil || got.ID != "p1" {
			t.Fatalf("got %+v, %v", got, err)
		}
		if f.queryCount() != 1 {
			t.Fatalf("want a single query, got %v", f.queries)
		}
	})
	t.Run("collaborator of a missing project gets not found", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on("FROM \"public\".\"Collaborator\"", []string{"Id", "UserId", "ProjectId"}, []driver.Value{"c1", "u1", "p1"})
		got, err := (&ProjectRepository{db: db}).GetProjectWithAccess(ctx, "p1", "u1")
		if got != nil || !errors.Is(err, ErrProjectNotFound) {
			t.Fatalf("got %+v, %v", got, err)
		}
	})
	t.Run("stranger is unauthorized", func(t *testing.T) {
		db, _ := newFakeDB(t)
		got, err := (&ProjectRepository{db: db}).GetProjectWithAccess(ctx, "p1", "u1")
		if got != nil || !errors.Is(err, ErrProjectUnauthorized) {
			t.Fatalf("got %+v, %v", got, err)
		}
	})
	t.Run("project query error", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onErr("SELECT", projectBoom)
		_, err := (&ProjectRepository{db: db}).GetProjectWithAccess(ctx, "p1", "u1")
		if !errors.Is(err, projectBoom) || !strings.HasPrefix(err.Error(), "failed to get project") {
			t.Fatal(err)
		}
		if f.queryCount() != 1 {
			t.Fatalf("must not fall through to the collaborator check: %v", f.queries)
		}
	})
	t.Run("collaborator query error", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onErr("FROM \"public\".\"Collaborator\"", projectBoom)
		_, err := (&ProjectRepository{db: db}).GetProjectWithAccess(ctx, "p1", "u1")
		if !errors.Is(err, projectBoom) || errors.Is(err, ErrProjectUnauthorized) || !strings.HasPrefix(err.Error(), "failed to check collaborator access") {
			t.Fatal(err)
		}
	})
}

func TestProjectRepositoryListQueries(t *testing.T) {
	ctx := context.Background()
	lists := []struct {
		name   string
		errMsg string
		call   func(r *ProjectRepository) ([]models.Project, error)
	}{
		{"by user", "failed to get projects", func(r *ProjectRepository) ([]models.Project, error) { return r.GetProjectsByUserID(ctx, "u1") }},
		{"collaborator", "failed to get collaborator projects", func(r *ProjectRepository) ([]models.Project, error) { return r.GetCollaboratorProjects(ctx, "u1") }},
	}
	for _, tc := range lists {
		t.Run(tc.name, func(t *testing.T) {
			db, f := newFakeDB(t)
			f.on("SELECT", projectCols, projectRow("p1", "u1"), projectRow("p2", "u2"))
			got, err := tc.call(&ProjectRepository{db: db})
			if err != nil || len(got) != 2 || got[0].ID != "p1" || got[1].OwnerId != "u2" {
				t.Fatalf("got %+v, %v", got, err)
			}
			q := f.lastQuery()
			for _, heavy := range []string{`"Logic"`, `"Styles"`, `"Header"`} {
				if strings.Contains(strings.SplitN(q, "FROM", 2)[0], heavy) {
					t.Fatalf("list must omit %s: %s", heavy, q)
				}
			}
			if !projectHasArg(f, "u1") || !strings.Contains(q, "ORDER BY") {
				t.Fatalf("unexpected query %s %v", q, f.args)
			}
		})
		t.Run(tc.name+" error", func(t *testing.T) {
			db, f := newFakeDB(t)
			f.onErr("SELECT", projectBoom)
			got, err := tc.call(&ProjectRepository{db: db})
			if got != nil || !errors.Is(err, projectBoom) || !strings.HasPrefix(err.Error(), tc.errMsg) {
				t.Fatalf("got %+v, %v", got, err)
			}
		})
	}
	t.Run("collaborator list joins the collaborator table", func(t *testing.T) {
		db, f := newFakeDB(t)
		_, _ = (&ProjectRepository{db: db}).GetCollaboratorProjects(ctx, "u1")
		if !f.sawQuery("JOIN", "Collaborator", "UserId") {
			t.Fatalf("expected join: %v", f.queries)
		}
	})
	t.Run("owned projects sort before shared ones", func(t *testing.T) {
		db, f := newFakeDB(t)
		_, _ = (&ProjectRepository{db: db}).GetProjectsByUserID(ctx, "u1")
		if !f.sawQuery("ORDER BY", "OwnerId", "DESC", "CreatedAt") {
			t.Fatalf("unexpected ordering: %v", f.queries)
		}
	})
}

func TestProjectRepositoryGetProjectPages(t *testing.T) {
	ctx := context.Background()
	t.Run("returns pages for a project the user can access", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on("FROM \"public\".\"Project\"", projectCols, projectRow("p1", "u1"))
		f.on("FROM \"public\".\"Page\"", []string{"Id", "ProjectId", "Name"}, []driver.Value{"a", "p1", "Home"}, []driver.Value{"b", "p1", "About"})
		got, err := (&ProjectRepository{db: db}).GetProjectPages(ctx, "p1", "u1")
		if err != nil || len(got) != 2 || got[0].Name != "Home" {
			t.Fatalf("got %+v, %v", got, err)
		}
	})
	t.Run("access error is returned and pages are not read", func(t *testing.T) {
		db, f := newFakeDB(t)
		_, err := (&ProjectRepository{db: db}).GetProjectPages(ctx, "p1", "u1")
		if !errors.Is(err, ErrProjectUnauthorized) {
			t.Fatal(err)
		}
		if f.sawQuery("\"public\".\"Page\"") {
			t.Fatalf("pages must not be read: %v", f.queries)
		}
	})
	t.Run("page query error", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on("FROM \"public\".\"Project\"", projectCols, projectRow("p1", "u1"))
		f.onErr("FROM \"public\".\"Page\"", projectBoom)
		got, err := (&ProjectRepository{db: db}).GetProjectPages(ctx, "p1", "u1")
		if got != nil || !errors.Is(err, projectBoom) || !strings.HasPrefix(err.Error(), "failed to get project pages") {
			t.Fatalf("got %+v, %v", got, err)
		}
	})
}

func TestProjectRepositoryCreateProject(t *testing.T) {
	ctx := context.Background()
	t.Run("nil audit fields are created and stamped", func(t *testing.T) {
		db, f := newFakeDB(t)
		p := &models.Project{ID: "p1", Name: "N", OwnerId: "u1", AuditFields: &models.AuditFields{}}
		if err := (&ProjectRepository{db: db}).CreateProject(ctx, p); err != nil {
			t.Fatal(err)
		}
		if p.CreatedAt.IsZero() || !p.CreatedAt.Equal(p.UpdatedAt) || !f.sawQuery("INSERT INTO", "Project") || !projectHasArg(f, "p1") {
			t.Fatalf("unexpected: %+v %v", p, f.queries)
		}
	})
	t.Run("database error", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onErr("INSERT", projectBoom)
		err := (&ProjectRepository{db: db}).CreateProject(ctx, &models.Project{ID: "p1", AuditFields: &models.AuditFields{}})
		if !errors.Is(err, projectBoom) || !strings.HasPrefix(err.Error(), "failed to create project") {
			t.Fatal(err)
		}
	})
}

func TestProjectRepositoryUpdateProject(t *testing.T) {
	ctx := context.Background()
	t.Run("verifies ownership, updates and returns the fresh project", func(t *testing.T) {
		db, f := newFakeDB(t)
		projectCount(f, 1)
		f.on("SELECT *", projectCols, projectRow("p1", "u1"))
		f.onExec("UPDATE", 1)
		updates := map[string]any{"Name": "Renamed"}
		got, err := (&ProjectRepository{db: db}).UpdateProject(ctx, "p1", "u1", updates)
		if err != nil || got.ID != "p1" {
			t.Fatalf("got %+v, %v", got, err)
		}
		if _, ok := updates["UpdatedAt"].(time.Time); !ok {
			t.Fatalf("UpdatedAt not stamped: %v", updates)
		}
		if !f.sawQuery("UPDATE", "Name", "UpdatedAt", "OwnerId") || !projectHasArg(f, "Renamed") {
			t.Fatalf("unexpected update: %v", f.queries)
		}
	})
	t.Run("non owner is unauthorized without writing", func(t *testing.T) {
		db, f := newFakeDB(t)
		projectCount(f, 0)
		got, err := (&ProjectRepository{db: db}).UpdateProject(ctx, "p1", "u2", map[string]any{"Name": "x"})
		if got != nil || !errors.Is(err, ErrProjectUnauthorized) || f.sawQuery("UPDATE") {
			t.Fatalf("got %+v, %v, %v", got, err, f.queries)
		}
	})
	t.Run("ownership check error", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onErr("count(", projectBoom)
		_, err := (&ProjectRepository{db: db}).UpdateProject(ctx, "p1", "u1", map[string]any{"Name": "x"})
		if !errors.Is(err, projectBoom) || !strings.HasPrefix(err.Error(), "failed to verify project ownership") {
			t.Fatal(err)
		}
	})
	t.Run("update error", func(t *testing.T) {
		db, f := newFakeDB(t)
		projectCount(f, 1)
		f.onErr("UPDATE", projectBoom)
		_, err := (&ProjectRepository{db: db}).UpdateProject(ctx, "p1", "u1", map[string]any{"Name": "x"})
		if !errors.Is(err, projectBoom) || !strings.HasPrefix(err.Error(), "failed to update project") {
			t.Fatal(err)
		}
	})
	t.Run("refetch error is returned", func(t *testing.T) {
		db, f := newFakeDB(t)
		projectCount(f, 1)
		f.onExec("UPDATE", 1)
		_, err := (&ProjectRepository{db: db}).UpdateProject(ctx, "p1", "u1", map[string]any{"Name": "x"})
		if !errors.Is(err, ErrProjectNotFound) {
			t.Fatal(err)
		}
	})
}

func TestProjectRepositoryDeleteProject(t *testing.T) {
	ctx := context.Background()
	t.Run("soft deletes an owned project", func(t *testing.T) {
		db, f := newFakeDB(t)
		projectCount(f, 1)
		f.onExec("UPDATE", 1)
		if err := (&ProjectRepository{db: db}).DeleteProject(ctx, "p1", "u1"); err != nil {
			t.Fatal(err)
		}
		if !f.sawQuery("UPDATE", "DeletedAt") || f.sawQuery("DELETE FROM") {
			t.Fatalf("expected a soft delete: %v", f.queries)
		}
	})
	t.Run("non owner is unauthorized", func(t *testing.T) {
		db, f := newFakeDB(t)
		projectCount(f, 0)
		if err := (&ProjectRepository{db: db}).DeleteProject(ctx, "p1", "u2"); !errors.Is(err, ErrProjectUnauthorized) || f.sawQuery("UPDATE") {
			t.Fatalf("%v %v", err, f.queries)
		}
	})
	t.Run("ownership check error", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onErr("count(", projectBoom)
		err := (&ProjectRepository{db: db}).DeleteProject(ctx, "p1", "u1")
		if !errors.Is(err, projectBoom) || !strings.HasPrefix(err.Error(), "failed to verify project ownership") {
			t.Fatal(err)
		}
	})
	t.Run("update error", func(t *testing.T) {
		db, f := newFakeDB(t)
		projectCount(f, 1)
		f.onErr("UPDATE", projectBoom)
		err := (&ProjectRepository{db: db}).DeleteProject(ctx, "p1", "u1")
		if !errors.Is(err, projectBoom) || !strings.HasPrefix(err.Error(), "failed to delete project") {
			t.Fatal(err)
		}
	})
}

func TestProjectRepositoryHardDeleteProject(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name     string
		affected int64
		dbErr    error
		wantErr  error
	}{
		{"deleted", 1, nil, nil},
		{"not the owner or missing", 0, nil, ErrProjectUnauthorized},
		{"database error", 0, projectBoom, projectBoom},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, f := newFakeDB(t)
			if tc.dbErr != nil {
				f.onErr("DELETE", tc.dbErr)
			} else {
				f.onExec("DELETE", tc.affected)
			}
			if err := (&ProjectRepository{db: db}).HardDeleteProject(ctx, "p1", "u1"); !errors.Is(err, tc.wantErr) {
				t.Fatalf("want %v, got %v", tc.wantErr, err)
			}
			if !f.sawQuery("DELETE FROM", "Project") || !projectHasArg(f, "u1") {
				t.Fatalf("unexpected statements: %v", f.queries)
			}
		})
	}
}

func TestProjectRepositoryRestoreProject(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name     string
		affected int64
		dbErr    error
		wantErr  error
	}{
		{"restored", 1, nil, nil},
		{"nothing to restore", 0, nil, ErrProjectNotFound},
		{"database error", 0, projectBoom, projectBoom},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, f := newFakeDB(t)
			if tc.dbErr != nil {
				f.onErr("UPDATE", tc.dbErr)
			} else {
				f.onExec("UPDATE", tc.affected)
			}
			err := (&ProjectRepository{db: db}).RestoreProject(ctx, "p1", "u1")
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("want %v, got %v", tc.wantErr, err)
			}
			if tc.dbErr != nil && !strings.HasPrefix(err.Error(), "failed to restore project") {
				t.Fatal(err)
			}
		})
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
	if got || !errors.Is(err, projectBoom) || !strings.HasPrefix(err.Error(), "failed to check project existence") {
		t.Fatalf("%v %v", got, err)
	}
}

func TestProjectRepositoryCreateProjectWithPage(t *testing.T) {
	ctx := context.Background()
	valid := func() (*models.Project, *models.Page) {
		return &models.Project{ID: "p1", Name: "N", OwnerId: "u1"}, &models.Page{Id: "pg", Name: "Home", Type: "page"}
	}
	t.Run("creates both in one committed transaction", func(t *testing.T) {
		db, f := newFakeDB(t)
		p, pg := valid()
		if err := (&ProjectRepository{db: db}).CreateProjectWithPage(ctx, p, pg); err != nil {
			t.Fatal(err)
		}
		if f.begins != 1 || f.commits != 1 || f.aborts != 0 {
			t.Fatalf("tx begins=%d commits=%d aborts=%d", f.begins, f.commits, f.aborts)
		}
		if p.AuditFields == nil || p.CreatedAt.IsZero() || !p.CreatedAt.Equal(pg.CreatedAt) || pg.ProjectId != "p1" {
			t.Fatalf("not stamped/linked: %+v %+v", p, pg)
		}
		if !f.sawQuery("INSERT INTO", "Project") || !f.sawQuery("INSERT INTO", "Page") {
			t.Fatalf("missing inserts: %v", f.queries)
		}
	})
	t.Run("project insert failure rolls back and skips the page", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onErr("INSERT INTO \"public\".\"Project\"", projectBoom)
		p, pg := valid()
		err := (&ProjectRepository{db: db}).CreateProjectWithPage(ctx, p, pg)
		if !errors.Is(err, projectBoom) || !strings.HasPrefix(err.Error(), "failed to create project") {
			t.Fatal(err)
		}
		if f.aborts != 1 || f.commits != 0 || f.sawQuery("INSERT INTO \"public\".\"Page\"") {
			t.Fatalf("aborts=%d commits=%d %v", f.aborts, f.commits, f.queries)
		}
	})
	t.Run("page insert failure rolls back", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onErr("INSERT INTO \"public\".\"Page\"", projectBoom)
		p, pg := valid()
		err := (&ProjectRepository{db: db}).CreateProjectWithPage(ctx, p, pg)
		if !errors.Is(err, projectBoom) || !strings.HasPrefix(err.Error(), "failed to create first page") {
			t.Fatal(err)
		}
		if f.aborts != 1 || f.commits != 0 {
			t.Fatalf("aborts=%d commits=%d", f.aborts, f.commits)
		}
	})
	t.Run("validation", func(t *testing.T) {
		db, f := newFakeDB(t)
		repo := &ProjectRepository{db: db}
		if err := repo.CreateProjectWithPage(ctx, nil, &models.Page{}); err == nil || err.Error() != "project and page are required" {
			t.Fatal(err)
		}
		for _, mutate := range []func(p *models.Project){
			func(p *models.Project) { p.ID = "" },
			func(p *models.Project) { p.Name = "" },
			func(p *models.Project) { p.OwnerId = "" },
		} {
			p, pg := valid()
			mutate(p)
			if err := repo.CreateProjectWithPage(ctx, p, pg); err == nil || err.Error() != "project id, name and owner are required" {
				t.Fatal(err)
			}
		}
		if f.begins != 0 || f.queryCount() != 0 {
			t.Fatalf("invalid input must not start a transaction: %v", f.queries)
		}
	})
}

func TestProjectRepositoryGetProjectAccess(t *testing.T) {
	ctx := context.Background()
	t.Run("owner", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on("FROM \"public\".\"Project\"", []string{"Id", "OwnerId"}, []driver.Value{"p1", "u1"})
		got, err := (&ProjectRepository{db: db}).GetProjectAccess(ctx, "p1", "u1")
		if err != nil || !got.IsOwner || got.IsCollaborator || got.Role != models.RoleOwner || got.UserId != "u1" || got.ProjectId != "p1" {
			t.Fatalf("got %+v, %v", got, err)
		}
		if f.sawQuery("\"public\".\"Collaborator\"") {
			t.Fatalf("owner should not query collaborators: %v", f.queries)
		}
	})
	t.Run("collaborator", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on("FROM \"public\".\"Project\"", []string{"Id", "OwnerId"}, []driver.Value{"p1", "u9"})
		f.on("FROM \"public\".\"Collaborator\"", []string{"Role"}, []driver.Value{"viewer"})
		got, err := (&ProjectRepository{db: db}).GetProjectAccess(ctx, "p1", "u1")
		if err != nil || got.IsOwner || !got.IsCollaborator || got.Role != models.CollaboratorRole("viewer") {
			t.Fatalf("got %+v, %v", got, err)
		}
	})
	t.Run("unknown project", func(t *testing.T) {
		db, _ := newFakeDB(t)
		if got, err := (&ProjectRepository{db: db}).GetProjectAccess(ctx, "p1", "u1"); got != nil || !errors.Is(err, ErrProjectNotFound) {
			t.Fatalf("got %+v, %v", got, err)
		}
	})
	t.Run("not a collaborator is reported as not found", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on("FROM \"public\".\"Project\"", []string{"Id", "OwnerId"}, []driver.Value{"p1", "u9"})
		if got, err := (&ProjectRepository{db: db}).GetProjectAccess(ctx, "p1", "u1"); got != nil || !errors.Is(err, ErrProjectNotFound) {
			t.Fatalf("got %+v, %v", got, err)
		}
	})
	t.Run("project query error", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onErr("FROM \"public\".\"Project\"", projectBoom)
		_, err := (&ProjectRepository{db: db}).GetProjectAccess(ctx, "p1", "u1")
		if !errors.Is(err, projectBoom) || !strings.HasPrefix(err.Error(), "failed to get project") {
			t.Fatal(err)
		}
	})
	t.Run("collaborator query error", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on("FROM \"public\".\"Project\"", []string{"Id", "OwnerId"}, []driver.Value{"p1", "u9"})
		f.onErr("FROM \"public\".\"Collaborator\"", projectBoom)
		_, err := (&ProjectRepository{db: db}).GetProjectAccess(ctx, "p1", "u1")
		if !errors.Is(err, projectBoom) || !strings.HasPrefix(err.Error(), "failed to get collaborator") {
			t.Fatal(err)
		}
	})
}
