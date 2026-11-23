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

// invitationHasArg reports whether any recorded statement was sent with the given bind value.
func invitationHasArg(f *fakeDB, want any) bool {
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

var invitationBoom = errors.New("boom")

var invitationCols = []string{"Id", "ProjectId", "Token", "Role", "Status", "ExpiresAt"}

func invitationRow(id, status string, expires time.Time) []driver.Value {
	return []driver.Value{id, "p1", "tok-" + id, "editor", status, expires}
}

func TestInvitationRepositoryCreateInvitation(t *testing.T) {
	ctx := context.Background()
	t.Run("fills id and pending status", func(t *testing.T) {
		db, f := newFakeDB(t)
		in := &models.Invitation{ProjectId: "p1", Email: "a@b.c", Token: "t", Role: "editor"}
		got, err := (&InvitationRepository{db: db}).CreateInvitation(ctx, in)
		if err != nil {
			t.Fatal(err)
		}
		if got.Id == "" || got.Status != models.InvitationStatusPending {
			t.Fatalf("defaults not applied: %+v", got)
		}
		if !f.sawQuery("INSERT INTO", "Invitation") || !invitationHasArg(f, got.Id) {
			t.Fatalf("missing insert: %v", f.queries)
		}
	})
	t.Run("keeps caller id and status", func(t *testing.T) {
		db, _ := newFakeDB(t)
		in := &models.Invitation{Id: "mine", Status: models.InvitationStatusCancelled, ProjectId: "p1", Token: "t"}
		got, err := (&InvitationRepository{db: db}).CreateInvitation(ctx, in)
		if err != nil || got.Id != "mine" || got.Status != models.InvitationStatusCancelled {
			t.Fatalf("got %+v, %v", got, err)
		}
	})
	t.Run("wraps database errors", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onErr("INSERT", invitationBoom)
		got, err := (&InvitationRepository{db: db}).CreateInvitation(ctx, &models.Invitation{ProjectId: "p1"})
		if got != nil || !errors.Is(err, invitationBoom) || !strings.HasPrefix(err.Error(), "failed to create invitation") {
			t.Fatalf("got %+v, %v", got, err)
		}
	})
}

func TestInvitationRepositoryListQueries(t *testing.T) {
	ctx := context.Background()
	future := time.Now().Add(time.Hour)
	lists := []struct {
		name   string
		errMsg string
		call   func(r *InvitationRepository) ([]models.Invitation, error)
	}{
		{"by project", "failed to get invitations by project", func(r *InvitationRepository) ([]models.Invitation, error) {
			return r.GetInvitationsByProject(ctx, "p1")
		}},
		{"pending by project", "failed to get pending invitations", func(r *InvitationRepository) ([]models.Invitation, error) {
			return r.GetPendingInvitationsByProject(ctx, "p1")
		}},
	}
	for _, tc := range lists {
		t.Run(tc.name, func(t *testing.T) {
			db, f := newFakeDB(t)
			f.on("SELECT", invitationCols, invitationRow("i1", "pending", future), invitationRow("i2", "pending", future))
			got, err := tc.call(&InvitationRepository{db: db})
			if err != nil || len(got) != 2 || got[0].Id != "i1" || got[1].Token != "tok-i2" || got[0].Role != "editor" {
				t.Fatalf("got %+v, %v", got, err)
			}
			if !invitationHasArg(f, "p1") {
				t.Fatalf("project filter missing: %v", f.args)
			}
		})
		t.Run(tc.name+" error", func(t *testing.T) {
			db, f := newFakeDB(t)
			f.onErr("SELECT", invitationBoom)
			got, err := tc.call(&InvitationRepository{db: db})
			if got != nil || !errors.Is(err, invitationBoom) || !strings.HasPrefix(err.Error(), tc.errMsg) {
				t.Fatalf("got %+v, %v", got, err)
			}
		})
	}
}

func TestInvitationRepositoryPendingFilterUsesStatus(t *testing.T) {
	db, f := newFakeDB(t)
	if _, err := (&InvitationRepository{db: db}).GetPendingInvitationsByProject(context.Background(), "p1"); err != nil {
		t.Fatal(err)
	}
	if !invitationHasArg(f, string(models.InvitationStatusPending)) && !invitationHasArg(f, models.InvitationStatusPending) {
		t.Fatalf("pending status not bound: %v", f.args)
	}
}

func TestInvitationRepositorySingleLookups(t *testing.T) {
	ctx := context.Background()
	future := time.Now().Add(time.Hour)
	lookups := []struct {
		name   string
		arg    string
		errMsg string
		call   func(r *InvitationRepository) (*models.Invitation, error)
	}{
		{"by id", "i1", "failed to get invitation by ID", func(r *InvitationRepository) (*models.Invitation, error) {
			return r.GetInvitationByID(ctx, "i1")
		}},
		{"by token", "tok-i1", "failed to get invitation by token", func(r *InvitationRepository) (*models.Invitation, error) {
			return r.GetInvitationByToken(ctx, "tok-i1")
		}},
	}
	for _, tc := range lookups {
		t.Run(tc.name+" found", func(t *testing.T) {
			db, f := newFakeDB(t)
			f.on("SELECT", invitationCols, invitationRow("i1", "pending", future))
			got, err := tc.call(&InvitationRepository{db: db})
			if err != nil || got.Id != "i1" || got.Status != models.InvitationStatusPending {
				t.Fatalf("got %+v, %v", got, err)
			}
			if !invitationHasArg(f, tc.arg) {
				t.Fatalf("lookup key missing: %v", f.args)
			}
		})
		t.Run(tc.name+" missing", func(t *testing.T) {
			db, _ := newFakeDB(t)
			if got, err := tc.call(&InvitationRepository{db: db}); got != nil || !errors.Is(err, ErrInvitationNotFound) {
				t.Fatalf("got %+v, %v", got, err)
			}
		})
		t.Run(tc.name+" error", func(t *testing.T) {
			db, f := newFakeDB(t)
			f.onErr("SELECT", invitationBoom)
			got, err := tc.call(&InvitationRepository{db: db})
			if got != nil || !errors.Is(err, invitationBoom) || errors.Is(err, ErrInvitationNotFound) || !strings.HasPrefix(err.Error(), tc.errMsg) {
				t.Fatalf("got %+v, %v", got, err)
			}
		})
	}
}

func TestInvitationRepositoryAcceptInvitation(t *testing.T) {
	ctx := context.Background()
	future := time.Now().Add(time.Hour)
	past := time.Now().Add(-time.Hour)

	t.Run("creates the collaborator and marks the invitation accepted", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on("SELECT", invitationCols, invitationRow("i1", "pending", future))
		if err := (&InvitationRepository{db: db}).AcceptInvitation(ctx, "tok-i1", "u9"); err != nil {
			t.Fatal(err)
		}
		if !f.sawQuery("INSERT INTO", "Collaborator") || !invitationHasArg(f, "u9") || !invitationHasArg(f, "p1") {
			t.Fatalf("collaborator not created for user and project: %v", f.queries)
		}
		if !f.sawQuery("UPDATE", "Invitation", "Status", "AcceptedAt") || !invitationHasArg(f, string(models.InvitationStatusAccepted)) && !invitationHasArg(f, models.InvitationStatusAccepted) {
			t.Fatalf("invitation not marked accepted: %v", f.queries)
		}
	})
	t.Run("collaborator carries the invited role", func(t *testing.T) {
		db, f := newFakeDB(t)
		row := invitationRow("i1", "pending", future)
		row[3] = "viewer"
		f.on("SELECT", invitationCols, row)
		if err := (&InvitationRepository{db: db}).AcceptInvitation(ctx, "tok-i1", "u9"); err != nil {
			t.Fatal(err)
		}
		if !invitationHasArg(f, "viewer") {
			t.Fatalf("role not propagated: %v", f.args)
		}
	})
	t.Run("expired", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on("SELECT", invitationCols, invitationRow("i1", "pending", past))
		if err := (&InvitationRepository{db: db}).AcceptInvitation(ctx, "tok-i1", "u9"); !errors.Is(err, ErrInvitationExpired) {
			t.Fatal(err)
		}
		if f.sawQuery("INSERT") || f.sawQuery("UPDATE ", "Invitation") {
			t.Fatalf("an expired invitation must not write: %v", f.queries)
		}
	})
	t.Run("not pending", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on("SELECT", invitationCols, invitationRow("i1", "accepted", future))
		if err := (&InvitationRepository{db: db}).AcceptInvitation(ctx, "tok-i1", "u9"); !errors.Is(err, ErrInvitationInvalid) {
			t.Fatal(err)
		}
		if f.sawQuery("INSERT") {
			t.Fatalf("must not write: %v", f.queries)
		}
	})
	t.Run("unknown token", func(t *testing.T) {
		db, _ := newFakeDB(t)
		if err := (&InvitationRepository{db: db}).AcceptInvitation(ctx, "nope", "u9"); !errors.Is(err, ErrInvitationNotFound) {
			t.Fatal(err)
		}
	})
	t.Run("collaborator insert fails", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on("SELECT", invitationCols, invitationRow("i1", "pending", future))
		f.onErr("INSERT", invitationBoom)
		err := (&InvitationRepository{db: db}).AcceptInvitation(ctx, "tok-i1", "u9")
		if !errors.Is(err, invitationBoom) || !strings.HasPrefix(err.Error(), "failed to create collaborator from invitation") {
			t.Fatal(err)
		}
		if f.sawQuery("UPDATE ", "Invitation") {
			t.Fatalf("invitation must stay pending: %v", f.queries)
		}
	})
	t.Run("status update fails", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on("SELECT", invitationCols, invitationRow("i1", "pending", future))
		f.onErr("UPDATE", invitationBoom)
		err := (&InvitationRepository{db: db}).AcceptInvitation(ctx, "tok-i1", "u9")
		if !errors.Is(err, invitationBoom) || !strings.HasPrefix(err.Error(), "failed to mark invitation as accepted") {
			t.Fatal(err)
		}
	})
}

func TestInvitationRepositoryDeleteInvitation(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name     string
		affected int64
		dbErr    error
		wantErr  error
	}{
		{"deleted", 1, nil, nil},
		{"unknown id", 0, nil, ErrInvitationNotFound},
		{"database error", 0, invitationBoom, invitationBoom},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, f := newFakeDB(t)
			if tc.dbErr != nil {
				f.onErr("DELETE", tc.dbErr)
			} else {
				f.onExec("DELETE", tc.affected)
			}
			if err := (&InvitationRepository{db: db}).DeleteInvitation(ctx, "i1"); !errors.Is(err, tc.wantErr) {
				t.Fatalf("want %v, got %v", tc.wantErr, err)
			}
			if !invitationHasArg(f, "i1") {
				t.Fatalf("id not bound: %v", f.args)
			}
		})
	}
}

func TestInvitationRepositoryUpdateAndCancel(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name     string
		affected int64
		dbErr    error
		wantErr  error
	}{
		{"updated", 1, nil, nil},
		{"unknown id", 0, nil, ErrInvitationNotFound},
		{"database error", 0, invitationBoom, invitationBoom},
	}
	for _, tc := range tests {
		t.Run("update "+tc.name, func(t *testing.T) {
			db, f := newFakeDB(t)
			if tc.dbErr != nil {
				f.onErr("UPDATE", tc.dbErr)
			} else {
				f.onExec("UPDATE", tc.affected)
			}
			err := (&InvitationRepository{db: db}).UpdateInvitationStatus(ctx, "i1", models.InvitationStatusAccepted)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("want %v, got %v", tc.wantErr, err)
			}
			if tc.dbErr != nil && !strings.HasPrefix(err.Error(), "failed to update invitation status") {
				t.Fatalf("unexpected message %v", err)
			}
		})
	}
	t.Run("cancel writes the cancelled status", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onExec("UPDATE", 1)
		if err := (&InvitationRepository{db: db}).CancelInvitation(ctx, "i1"); err != nil {
			t.Fatal(err)
		}
		if !invitationHasArg(f, string(models.InvitationStatusCancelled)) && !invitationHasArg(f, models.InvitationStatusCancelled) {
			t.Fatalf("cancelled status not bound: %v", f.args)
		}
	})
}
