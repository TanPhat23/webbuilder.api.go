package repositories

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"my-go-app/internal/models"
)

func openMigratedTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("UUID_GORM_TEST_DSN")
	if dsn == "" {
		t.Skip("requires dedicated migrated PostgreSQL database")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { sqlDB.Close() })
	return db
}

func mustExec(t *testing.T, db *gorm.DB, sql string, args ...any) {
	t.Helper()
	if err := db.Exec(sql, args...).Error; err != nil {
		t.Fatal(err)
	}
}

func TestPostgresProjectAccessAndLists(t *testing.T) {
	db := openMigratedTestDB(t)
	ctx := context.Background()
	repo := NewProjectRepository(db)

	owner, collab, stranger := uuid.NewString(), uuid.NewString(), uuid.NewString()
	for _, id := range []string{owner, collab, stranger} {
		mustExec(t, db, `INSERT INTO "User" ("Id","Email","CreatedAt","UpdatedAt") VALUES (?,?,NOW(),NOW())`, id, id+"@example.invalid")
	}
	t.Cleanup(func() {
		// Projects, pages, elements and collaborators cascade from the users.
		mustExec(t, db, `DELETE FROM "User" WHERE "Id" IN (?,?,?)`, owner, collab, stranger)
	})

	newProject := func(name string, ageHours int, deleted bool) string {
		id := uuid.NewString()
		mustExec(t, db, `INSERT INTO "Project" ("Id","Name","OwnerId","CreatedAt","UpdatedAt","DeletedAt","Styles","Logic")
			VALUES (?, ?, ?, NOW() - make_interval(hours => ?), NOW(),
			        CASE WHEN ? THEN NOW() ELSE NULL END, '{"big":true}'::jsonb, '{"flows":[1]}'::jsonb)`,
			id, name, owner, ageHours, deleted)
		return id
	}
	older := newProject("older", 10, false)
	newer := newProject("newer", 1, false)
	gone := newProject("gone", 5, true)
	for _, id := range []string{older, gone} {
		mustExec(t, db, `INSERT INTO "Collaborator" ("Id","UserId","ProjectId","UpdatedAt") VALUES (?,?,?,NOW())`, uuid.NewString(), collab, id)
	}

	t.Run("owner and collaborator get the project", func(t *testing.T) {
		for _, user := range []string{owner, collab} {
			if p, err := repo.GetProjectWithAccess(ctx, older, user); err != nil || p.ID != older {
				t.Fatalf("user %s: %v %v", user, p, err)
			}
		}
	})
	t.Run("a stranger is unauthorized, a missing project too", func(t *testing.T) {
		if _, err := repo.GetProjectWithAccess(ctx, older, stranger); !errors.Is(err, ErrProjectUnauthorized) {
			t.Fatalf("stranger: %v", err)
		}
		if _, err := repo.GetProjectWithAccess(ctx, uuid.NewString(), collab); !errors.Is(err, ErrProjectUnauthorized) {
			t.Fatalf("unknown project: %v", err)
		}
	})
	t.Run("a deleted project is not found for its collaborator", func(t *testing.T) {
		if _, err := repo.GetProjectWithAccess(ctx, gone, collab); !errors.Is(err, ErrProjectNotFound) {
			t.Fatalf("collaborator of a deleted project: %v", err)
		}
		if _, err := repo.GetProjectWithAccess(ctx, gone, owner); !errors.Is(err, ErrProjectUnauthorized) {
			t.Fatalf("owner of a deleted project: %v", err)
		}
	})
	t.Run("lists hold owned then shared projects, newest first, without heavy columns", func(t *testing.T) {
		mine, err := repo.GetProjectsByUserID(ctx, owner)
		if err != nil || len(mine) != 2 || mine[0].ID != newer || mine[1].ID != older {
			t.Fatalf("owner list = %v, %v; want [newer=%s older=%s]", projectIDs(mine), err, newer, older)
		}
		shared, err := repo.GetProjectsByUserID(ctx, collab)
		if err != nil || len(shared) != 1 || shared[0].ID != older {
			t.Fatalf("collaborator list = %v, %v", projectIDs(shared), err)
		}
		for _, p := range append(mine, shared...) {
			if p.Styles != nil || len(p.Logic) != 0 {
				t.Fatalf("list returned heavy columns: %+v", p)
			}
		}
	})
}

func TestPostgresElementsByPageIdsAreScopedToProject(t *testing.T) {
	db := openMigratedTestDB(t)
	ctx := context.Background()
	repo := NewElementRepository(db)

	owner := uuid.NewString()
	mustExec(t, db, `INSERT INTO "User" ("Id","Email","CreatedAt","UpdatedAt") VALUES (?,?,NOW(),NOW())`, owner, owner+"@example.invalid")
	t.Cleanup(func() { mustExec(t, db, `DELETE FROM "User" WHERE "Id" = ?`, owner) })

	pageOf := map[string]string{}
	for _, name := range []string{"a", "b"} {
		project, page := uuid.NewString(), uuid.NewString()
		mustExec(t, db, `INSERT INTO "Project" ("Id","Name","OwnerId","CreatedAt","UpdatedAt") VALUES (?,?,?,NOW(),NOW())`, project, name, owner)
		mustExec(t, db, `INSERT INTO "Page" ("Id","Name","Type","Styles","ProjectId","CreatedAt","UpdatedAt") VALUES (?, 'home', 'sp', '{}', ?, NOW(), NOW())`, page, project)
		mustExec(t, db, `INSERT INTO "Element" ("Id","PageId","Type") VALUES (?,?,'Section')`, "root-"+page, page)
		pageOf[project] = page
	}
	var projectA, projectB string
	for project := range pageOf {
		if projectA == "" {
			projectA = project
		} else {
			projectB = project
		}
	}

	got, err := repo.GetElementsByPageIds(ctx, projectA, []string{pageOf[projectA]})
	if err != nil || len(got) != 1 {
		t.Fatalf("own page: %d elements, %v", len(got), err)
	}
	got, err = repo.GetElementsByPageIds(ctx, projectA, []string{pageOf[projectB]})
	if err != nil || len(got) != 0 {
		t.Fatalf("another project's page must read as empty, got %d elements, %v", len(got), err)
	}
}

func projectIDs(projects []models.Project) []string {
	out := make([]string, 0, len(projects))
	for _, p := range projects {
		out = append(out, p.ID)
	}
	return out
}
