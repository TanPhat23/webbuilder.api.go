package repositories

import (
	"context"
	"database/sql/driver"
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"

	"my-go-app/internal/models"
)

var errCollaboratorBoom = errors.New("collaborator boom")

const (
	collaboratorTable = `FROM "public"."Collaborator"`
	collaboratorUser  = `FROM "public"."User"`
	collaboratorProj  = `FROM "public"."Project"`
)

func TestCollaboratorCreateGeneratesIDAndKeepsSuppliedID(t *testing.T) {
	db, f := newFakeDB(t)
	f.onExec("INSERT INTO", 1)
	repo := NewCollaboratorRepository(db)

	got, err := repo.CreateCollaborator(context.Background(), &models.Collaborator{UserId: "u1", ProjectId: "p1", Role: models.RoleEditor})
	if err != nil {
		t.Fatal(err)
	}
	if got.Id == "" || got.UserId != "u1" || got.ProjectId != "p1" {
		t.Errorf("unexpected collaborator: %+v", got)
	}
	got, err = repo.CreateCollaborator(context.Background(), &models.Collaborator{Id: "fixed", UserId: "u1", ProjectId: "p1"})
	if err != nil || got.Id != "fixed" {
		t.Errorf("supplied id changed: %+v, %v", got, err)
	}
	if !f.sawQuery("INSERT INTO", "Collaborator") {
		t.Errorf("no insert issued: %q", f.lastQuery())
	}
}

func TestCollaboratorCreatePropagatesError(t *testing.T) {
	db, f := newFakeDB(t)
	f.onErr("INSERT INTO", errCollaboratorBoom)
	got, err := NewCollaboratorRepository(db).CreateCollaborator(context.Background(), &models.Collaborator{UserId: "u1", ProjectId: "p1"})
	if got != nil || !errors.Is(err, errCollaboratorBoom) {
		t.Errorf("got %+v, %v", got, err)
	}
}

func TestCollaboratorGetByProjectAddsNormalisedOwner(t *testing.T) {
	db, f := newFakeDB(t)
	created := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	f.on(collaboratorTable, []string{"Id", "UserId", "ProjectId", "Role"}, []driver.Value{"c1", "u1", "p1", "editor"})
	// The first User query is the preload of the collaborator, the second the owner lookup.
	f.on(collaboratorUser, []string{"Id", "Email", "FirstName", "LastName"}, []driver.Value{"u1", "a@example.com", nil, nil}).once()
	f.on(collaboratorUser, []string{"Id", "Email", "FirstName", "LastName"}, []driver.Value{"o1", "owner@example.com", "Olly", nil})
	f.on(collaboratorProj, []string{"Id", "OwnerId", "CreatedAt", "UpdatedAt"}, []driver.Value{"p1", "o1", created, created})

	got, err := NewCollaboratorRepository(db).GetCollaboratorsByProject(context.Background(), "p1")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("want collaborator plus owner, got %+v", got)
	}
	member := got[0]
	if member.Id != "c1" || member.User.FirstName == nil || *member.User.FirstName != "a@example.com" || member.User.LastName == nil || *member.User.LastName != "" {
		t.Errorf("member names not normalised: %+v", member.User)
	}
	owner := got[1]
	if owner.Id != "owner-p1" || owner.Role != models.RoleOwner || owner.UserId != "o1" || owner.ProjectId != "p1" || !owner.CreatedAt.Equal(created) {
		t.Errorf("unexpected owner entry: %+v", owner)
	}
	if owner.User.FirstName == nil || *owner.User.FirstName != "Olly" || owner.User.LastName == nil || *owner.User.LastName != "" {
		t.Errorf("owner names not normalised: %+v", owner.User)
	}
}

func TestCollaboratorGetByProjectNormalisesOwnerWithoutNames(t *testing.T) {
	db, f := newFakeDB(t)
	now := time.Now()
	f.on(collaboratorProj, []string{"Id", "OwnerId", "CreatedAt", "UpdatedAt"}, []driver.Value{"p1", "o1", now, now})
	f.on(collaboratorUser, []string{"Id", "Email", "FirstName", "LastName"}, []driver.Value{"o1", "owner@example.com", nil, nil})
	got, err := NewCollaboratorRepository(db).GetCollaboratorsByProject(context.Background(), "p1")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].User.FirstName == nil || *got[0].User.FirstName != "owner@example.com" {
		t.Fatalf("owner first name should fall back to email: %+v", got)
	}
}

func TestCollaboratorGetByProjectOmitsMissingOwner(t *testing.T) {
	db, f := newFakeDB(t)
	now := time.Now()
	f.on(collaboratorProj, []string{"Id", "OwnerId", "CreatedAt", "UpdatedAt"}, []driver.Value{"p1", "ghost", now, now})
	got, err := NewCollaboratorRepository(db).GetCollaboratorsByProject(context.Background(), "p1")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("owner without a user row must be omitted, got %+v", got)
	}
}

func TestCollaboratorGetByProjectErrors(t *testing.T) {
	t.Run("list fails", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onErr(collaboratorTable, errCollaboratorBoom)
		if _, err := NewCollaboratorRepository(db).GetCollaboratorsByProject(context.Background(), "p1"); !errors.Is(err, errCollaboratorBoom) {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("project missing", func(t *testing.T) {
		db, _ := newFakeDB(t)
		got, err := NewCollaboratorRepository(db).GetCollaboratorsByProject(context.Background(), "p1")
		if got != nil || !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Errorf("got %+v, %v", got, err)
		}
	})
}

func TestCollaboratorGetByID(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on(collaboratorTable, []string{"Id", "UserId", "ProjectId", "Role"}, []driver.Value{"c1", "u1", "p1", "viewer"})
		got, err := NewCollaboratorRepository(db).GetCollaboratorByID(context.Background(), "c1")
		if err != nil || got == nil || got.Id != "c1" || got.Role != "viewer" {
			t.Fatalf("got %+v, %v", got, err)
		}
	})
	t.Run("missing returns nil without error", func(t *testing.T) {
		db, _ := newFakeDB(t)
		got, err := NewCollaboratorRepository(db).GetCollaboratorByID(context.Background(), "nope")
		if got != nil || err != nil {
			t.Errorf("got %+v, %v", got, err)
		}
	})
	t.Run("error", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onErr(collaboratorTable, errCollaboratorBoom)
		got, err := NewCollaboratorRepository(db).GetCollaboratorByID(context.Background(), "c1")
		if got != nil || !errors.Is(err, errCollaboratorBoom) {
			t.Errorf("got %+v, %v", got, err)
		}
	})
}

func TestCollaboratorUpdateRole(t *testing.T) {
	db, f := newFakeDB(t)
	f.onExec("UPDATE", 1)
	repo := NewCollaboratorRepository(db)
	if err := repo.UpdateCollaboratorRole(context.Background(), "c1", models.RoleViewer); err != nil {
		t.Fatal(err)
	}
	if !f.sawQuery("UPDATE", "\"Role\"", "\"Id\" =") {
		t.Errorf("unexpected update: %q", f.lastQuery())
	}
	foundRole := false
	for _, a := range f.args[len(f.args)-1] {
		if v, ok := a.Value.(string); ok && v == string(models.RoleViewer) {
			foundRole = true
		}
	}
	if !foundRole {
		t.Errorf("role not bound: %+v", f.args[len(f.args)-1])
	}

	db2, f2 := newFakeDB(t)
	f2.onErr("UPDATE", errCollaboratorBoom)
	if err := NewCollaboratorRepository(db2).UpdateCollaboratorRole(context.Background(), "c1", models.RoleViewer); !errors.Is(err, errCollaboratorBoom) {
		t.Errorf("err = %v", err)
	}
}

func TestCollaboratorDelete(t *testing.T) {
	db, f := newFakeDB(t)
	f.onExec("DELETE", 1)
	if err := NewCollaboratorRepository(db).DeleteCollaborator(context.Background(), "c1"); err != nil {
		t.Fatal(err)
	}
	if !f.sawQuery("DELETE FROM", "Collaborator", "\"Id\" =") {
		t.Errorf("unexpected delete: %q", f.lastQuery())
	}
	db2, f2 := newFakeDB(t)
	f2.onErr("DELETE", errCollaboratorBoom)
	if err := NewCollaboratorRepository(db2).DeleteCollaborator(context.Background(), "c1"); !errors.Is(err, errCollaboratorBoom) {
		t.Errorf("err = %v", err)
	}
}

func TestCollaboratorIsCollaborator(t *testing.T) {
	for _, tc := range []struct {
		name string
		n    int64
		want bool
	}{{"member", 1, true}, {"not a member", 0, false}} {
		t.Run(tc.name, func(t *testing.T) {
			db, f := newFakeDB(t)
			f.on("count(*)", []string{"count"}, []driver.Value{tc.n})
			got, err := NewCollaboratorRepository(db).IsCollaborator(context.Background(), "p1", "u1")
			if err != nil || got != tc.want {
				t.Errorf("got %v, %v want %v", got, err, tc.want)
			}
			if !f.sawQuery("\"ProjectId\" =", "\"UserId\" =") {
				t.Errorf("filters missing: %q", f.lastQuery())
			}
		})
	}
	db, f := newFakeDB(t)
	f.onErr("count(*)", errCollaboratorBoom)
	if ok, err := NewCollaboratorRepository(db).IsCollaborator(context.Background(), "p1", "u1"); ok || !errors.Is(err, errCollaboratorBoom) {
		t.Errorf("got %v, %v", ok, err)
	}
}
