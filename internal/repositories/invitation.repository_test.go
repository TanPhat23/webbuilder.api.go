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
	invitationBoom = errors.New("boom")
	invitationCols = []string{"Id", "ProjectId", "Token", "Role", "Status", "ExpiresAt"}
)

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

func invitationRow(id, status string, expires time.Time) []driver.Value {
	return []driver.Value{id, "p1", "tok-" + id, "editor", status, expires}
}

func assertInvitationErrPrefix(t *testing.T, err, want error, prefix string) {
	t.Helper()
	if !errors.Is(err, want) {
		t.Fatalf("want %v, got %v", want, err)
	}
	if prefix != "" && !strings.HasPrefix(err.Error(), prefix) {
		t.Fatalf("want prefix %q, got %v", prefix, err)
	}
}

func TestInvitationRepositoryCreateInvitation(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	in := &models.Invitation{ProjectId: "p1", Email: "a@b.c", Token: "t", Role: "editor"}
	got, err := (&InvitationRepository{db: db}).Create(ctx, in)
	if err != nil || got.Id == "" || got.Status != models.InvitationStatusPending {
		t.Fatalf("defaults not applied: %+v, %v", got, err)
	}
	if !f.sawQuery("INSERT INTO", "Invitation") || !invitationHasArg(f, got.Id) {
		t.Fatalf("missing insert: %v", f.queries)
	}

	dbKeep, _ := newFakeDB(t)
	keepIn := &models.Invitation{Id: "mine", Status: models.InvitationStatusCancelled, ProjectId: "p1", Token: "t"}
	keepGot, err := (&InvitationRepository{db: dbKeep}).Create(ctx, keepIn)
	if err != nil || keepGot.Id != "mine" || keepGot.Status != models.InvitationStatusCancelled {
		t.Fatalf("got %+v, %v", keepGot, err)
	}

	dbErr, fErr := newFakeDB(t)
	fErr.onErr("INSERT", invitationBoom)
	errGot, err := (&InvitationRepository{db: dbErr}).Create(ctx, &models.Invitation{ProjectId: "p1"})
	if errGot != nil {
		t.Fatalf("expected nil, got %+v", errGot)
	}
	assertInvitationErrPrefix(t, err, invitationBoom, "failed to create invitation")
}

func TestInvitationRepositoryListQueries(t *testing.T) {
	ctx := context.Background()
	future := time.Now().Add(time.Hour)
	lists := []struct {
		errMsg string
		call   func(r *InvitationRepository) ([]models.Invitation, error)
	}{
		{"failed to get invitations by project", func(r *InvitationRepository) ([]models.Invitation, error) {
			return r.ListByProject(ctx, "p1")
		}},
		{"failed to get pending invitations", func(r *InvitationRepository) ([]models.Invitation, error) {
			return r.ListPendingByProject(ctx, "p1")
		}},
	}
	for _, tc := range lists {
		db, f := newFakeDB(t)
		f.on("SELECT", invitationCols, invitationRow("i1", "pending", future), invitationRow("i2", "pending", future))
		got, err := tc.call(&InvitationRepository{db: db})
		assertListedInvitations(t, f, got, err)

		dbErr, fErr := newFakeDB(t)
		fErr.onErr("SELECT", invitationBoom)
		errGot, err := tc.call(&InvitationRepository{db: dbErr})
		if errGot != nil {
			t.Fatalf("expected nil, got %+v", errGot)
		}
		assertInvitationErrPrefix(t, err, invitationBoom, tc.errMsg)
	}
}

func assertListedInvitations(t *testing.T, f *fakeDB, got []models.Invitation, err error) {
	t.Helper()
	if err != nil || len(got) != 2 || got[0].Id != "i1" {
		t.Fatalf("got %+v, %v", got, err)
	}
	if got[1].Token != "tok-i2" || got[0].Role != "editor" || !invitationHasArg(f, "p1") {
		t.Fatalf("got %+v, args=%v", got, f.args)
	}
}

func TestInvitationRepositoryPendingFilterUsesStatus(t *testing.T) {
	db, f := newFakeDB(t)
	if _, err := (&InvitationRepository{db: db}).ListPendingByProject(context.Background(), "p1"); err != nil {
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
		arg    string
		errMsg string
		call   func(r *InvitationRepository) (*models.Invitation, error)
	}{
		{"i1", "failed to get invitation by ID", func(r *InvitationRepository) (*models.Invitation, error) {
			return r.GetByID(ctx, "i1")
		}},
		{"tok-i1", "failed to get invitation by token", func(r *InvitationRepository) (*models.Invitation, error) {
			return r.GetByToken(ctx, "tok-i1")
		}},
	}
	for _, tc := range lookups {
		db, f := newFakeDB(t)
		f.on("SELECT", invitationCols, invitationRow("i1", "pending", future))
		got, err := tc.call(&InvitationRepository{db: db})
		if err != nil || got.Id != "i1" || got.Status != models.InvitationStatusPending {
			t.Fatalf("got %+v, %v", got, err)
		}
		if !invitationHasArg(f, tc.arg) {
			t.Fatalf("lookup key missing: %v", f.args)
		}

		dbMiss, _ := newFakeDB(t)
		if gotMiss, err := tc.call(&InvitationRepository{db: dbMiss}); gotMiss != nil || !errors.Is(err, ErrInvitationNotFound) {
			t.Fatalf("got %+v, %v", gotMiss, err)
		}

		dbErr, fErr := newFakeDB(t)
		fErr.onErr("SELECT", invitationBoom)
		gotErr, err := tc.call(&InvitationRepository{db: dbErr})
		if gotErr != nil || errors.Is(err, ErrInvitationNotFound) {
			t.Fatalf("got %+v, %v", gotErr, err)
		}
		assertInvitationErrPrefix(t, err, invitationBoom, tc.errMsg)
	}
}

func TestInvitationRepositoryAcceptInvitationSuccess(t *testing.T) {
	ctx := context.Background()
	future := time.Now().Add(time.Hour)
	db, f := newFakeDB(t)
	f.on("SELECT", invitationCols, invitationRow("i1", "pending", future))
	if err := (&InvitationRepository{db: db}).Accept(ctx, "tok-i1", "u9"); err != nil {
		t.Fatal(err)
	}
	if !f.sawQuery("INSERT INTO", "Collaborator") || !invitationHasArg(f, "u9") || !invitationHasArg(f, "p1") {
		t.Fatalf("collaborator not created for user and project: %v", f.queries)
	}
	assertAcceptedStatusUpdated(t, f)

	dbRole, fRole := newFakeDB(t)
	row := invitationRow("i1", "pending", future)
	row[3] = "viewer"
	fRole.on("SELECT", invitationCols, row)
	if err := (&InvitationRepository{db: dbRole}).Accept(ctx, "tok-i1", "u9"); err != nil {
		t.Fatal(err)
	}
	if !invitationHasArg(fRole, "viewer") {
		t.Fatalf("role not propagated: %v", fRole.args)
	}
}

func assertAcceptedStatusUpdated(t *testing.T, f *fakeDB) {
	t.Helper()
	if !f.sawQuery("UPDATE", "Invitation", "Status", "AcceptedAt") {
		t.Fatalf("invitation not marked accepted: %v", f.queries)
	}
	if !invitationHasArg(f, string(models.InvitationStatusAccepted)) && !invitationHasArg(f, models.InvitationStatusAccepted) {
		t.Fatalf("accepted status not bound: %v", f.args)
	}
}

func TestInvitationRepositoryAcceptInvitationRejections(t *testing.T) {
	ctx := context.Background()
	future := time.Now().Add(time.Hour)
	past := time.Now().Add(-time.Hour)

	dbExp, fExp := newFakeDB(t)
	fExp.on("SELECT", invitationCols, invitationRow("i1", "pending", past))
	assertInvitationErrPrefix(t, (&InvitationRepository{db: dbExp}).Accept(ctx, "tok-i1", "u9"), ErrInvitationExpired, "")
	if fExp.sawQuery("INSERT") || fExp.sawQuery("UPDATE ", "Invitation") {
		t.Fatalf("expired invitation must not write: %v", fExp.queries)
	}

	dbAcc, fAcc := newFakeDB(t)
	fAcc.on("SELECT", invitationCols, invitationRow("i1", "accepted", future))
	assertInvitationErrPrefix(t, (&InvitationRepository{db: dbAcc}).Accept(ctx, "tok-i1", "u9"), ErrInvitationInvalid, "")
	if fAcc.sawQuery("INSERT") {
		t.Fatalf("must not write: %v", fAcc.queries)
	}

	dbMiss, _ := newFakeDB(t)
	assertInvitationErrPrefix(t, (&InvitationRepository{db: dbMiss}).Accept(ctx, "nope", "u9"), ErrInvitationNotFound, "")
}

func TestInvitationRepositoryAcceptInvitationDBErrors(t *testing.T) {
	ctx := context.Background()
	future := time.Now().Add(time.Hour)

	dbCollab, fCollab := newFakeDB(t)
	fCollab.on("SELECT", invitationCols, invitationRow("i1", "pending", future))
	fCollab.onErr("INSERT", invitationBoom)
	assertInvitationErrPrefix(t, (&InvitationRepository{db: dbCollab}).Accept(ctx, "tok-i1", "u9"), invitationBoom, "failed to create collaborator from invitation")
	if fCollab.sawQuery("UPDATE ", "Invitation") {
		t.Fatalf("invitation must stay pending: %v", fCollab.queries)
	}

	dbStatus, fStatus := newFakeDB(t)
	fStatus.on("SELECT", invitationCols, invitationRow("i1", "pending", future))
	fStatus.onErr("UPDATE", invitationBoom)
	assertInvitationErrPrefix(t, (&InvitationRepository{db: dbStatus}).Accept(ctx, "tok-i1", "u9"), invitationBoom, "failed to mark invitation as accepted")
}

func TestInvitationRepositoryDeleteInvitation(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		affected int64
		dbErr    error
		wantErr  error
	}{
		{1, nil, nil},
		{0, nil, ErrInvitationNotFound},
		{0, invitationBoom, invitationBoom},
	} {
		db, f := newFakeDB(t)
		if tc.dbErr != nil {
			f.onErr("DELETE", tc.dbErr)
		} else {
			f.onExec("DELETE", tc.affected)
		}
		assertInvitationErrPrefix(t, (&InvitationRepository{db: db}).Delete(ctx, "i1"), tc.wantErr, "")
		if !invitationHasArg(f, "i1") {
			t.Fatalf("id not bound: %v", f.args)
		}
	}
}

func TestInvitationRepositoryUpdateAndCancel(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		affected int64
		dbErr    error
		wantErr  error
		prefix   string
	}{
		{1, nil, nil, ""},
		{0, nil, ErrInvitationNotFound, ""},
		{0, invitationBoom, invitationBoom, "failed to update invitation status"},
	} {
		db, f := newFakeDB(t)
		if tc.dbErr != nil {
			f.onErr("UPDATE", tc.dbErr)
		} else {
			f.onExec("UPDATE", tc.affected)
		}
		err := (&InvitationRepository{db: db}).UpdateStatus(ctx, "i1", models.InvitationStatusAccepted)
		assertInvitationErrPrefix(t, err, tc.wantErr, tc.prefix)
	}

	dbCancel, fCancel := newFakeDB(t)
	fCancel.onExec("UPDATE", 1)
	if err := (&InvitationRepository{db: dbCancel}).Cancel(ctx, "i1"); err != nil {
		t.Fatal(err)
	}
	if !invitationHasArg(fCancel, string(models.InvitationStatusCancelled)) && !invitationHasArg(fCancel, models.InvitationStatusCancelled) {
		t.Fatalf("cancelled status not bound: %v", fCancel.args)
	}
}
