package repositories_test

import (
	"context"
	"errors"
	"testing"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/tests/testutil"
)

func TestMockInvitationRepository_DefaultGetByIDReturnsSentinel(t *testing.T) {
	repo := &testutil.MockInvitationRepository{}
	_, err := repo.GetInvitationByID(context.Background(), "inv-1")
	if !errors.Is(err, repositories.ErrInvitationNotFound) {
		t.Errorf("want ErrInvitationNotFound, got %v", err)
	}
}

func TestMockInvitationRepository_DefaultGetByTokenReturnsSentinel(t *testing.T) {
	repo := &testutil.MockInvitationRepository{}
	_, err := repo.GetInvitationByToken(context.Background(), "tok-abc")
	if !errors.Is(err, repositories.ErrInvitationNotFound) {
		t.Errorf("want ErrInvitationNotFound, got %v", err)
	}
}

func TestMockInvitationRepository_CreateDefaultReturnsInput(t *testing.T) {
	repo := &testutil.MockInvitationRepository{}
	inv := &models.Invitation{Email: "a@b.com", ProjectId: "p1"}
	got, err := repo.CreateInvitation(context.Background(), inv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != inv {
		t.Error("expected the same pointer to be returned by default")
	}
}

func TestMockInvitationRepository_CreateFnAssignsID(t *testing.T) {
	repo := &testutil.MockInvitationRepository{
		CreateInvitationFn: func(_ context.Context, inv *models.Invitation) (*models.Invitation, error) {
			inv.Id = "inv-generated"
			return inv, nil
		},
	}
	inv := &models.Invitation{Email: "a@b.com", ProjectId: "p1"}
	got, err := repo.CreateInvitation(context.Background(), inv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Id != "inv-generated" {
		t.Errorf("Id: got %q, want %q", got.Id, "inv-generated")
	}
}

func TestMockInvitationRepository_GetInvitationsByProjectDefaultsToEmpty(t *testing.T) {
	repo := &testutil.MockInvitationRepository{}
	invs, err := repo.GetInvitationsByProject(context.Background(), "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(invs) != 0 {
		t.Errorf("expected empty slice, got %d", len(invs))
	}
}

func TestMockInvitationRepository_GetInvitationsByProjectFnFilters(t *testing.T) {
	all := []models.Invitation{
		{Id: "inv-1", ProjectId: "p1"},
		{Id: "inv-2", ProjectId: "p2"},
		{Id: "inv-3", ProjectId: "p1"},
	}
	repo := &testutil.MockInvitationRepository{
		GetInvitationsByProjectFn: func(_ context.Context, projectID string) ([]models.Invitation, error) {
			var out []models.Invitation
			for _, inv := range all {
				if inv.ProjectId == projectID {
					out = append(out, inv)
				}
			}
			return out, nil
		},
	}
	got, err := repo.GetInvitationsByProject(context.Background(), "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 invitations for p1, got %d", len(got))
	}
}

func TestMockInvitationRepository_GetInvitationByIDFnReturnsInvitation(t *testing.T) {
	want := &models.Invitation{Id: "inv-1", Email: "a@b.com"}
	repo := &testutil.MockInvitationRepository{
		GetInvitationByIDFn: func(_ context.Context, id string) (*models.Invitation, error) {
			if id == "inv-1" {
				return want, nil
			}
			return nil, repositories.ErrInvitationNotFound
		},
	}
	got, err := repo.GetInvitationByID(context.Background(), "inv-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Id != want.Id {
		t.Errorf("Id: got %q, want %q", got.Id, want.Id)
	}
	_, err = repo.GetInvitationByID(context.Background(), "missing")
	if !errors.Is(err, repositories.ErrInvitationNotFound) {
		t.Errorf("missing id: want ErrInvitationNotFound, got %v", err)
	}
}

func TestMockInvitationRepository_GetInvitationByTokenFnReturnsInvitation(t *testing.T) {
	want := &models.Invitation{Id: "inv-1", Token: "tok-abc"}
	repo := &testutil.MockInvitationRepository{
		GetInvitationByTokenFn: func(_ context.Context, token string) (*models.Invitation, error) {
			if token == "tok-abc" {
				return want, nil
			}
			return nil, repositories.ErrInvitationNotFound
		},
	}
	got, err := repo.GetInvitationByToken(context.Background(), "tok-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Token != want.Token {
		t.Errorf("Token: got %q, want %q", got.Token, want.Token)
	}
}

func TestMockInvitationRepository_AcceptInvitationDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockInvitationRepository{}
	if err := repo.AcceptInvitation(context.Background(), "tok-x", "u1"); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestMockInvitationRepository_AcceptInvitationFnCalled(t *testing.T) {
	var capturedToken, capturedUser string
	repo := &testutil.MockInvitationRepository{
		AcceptInvitationFn: func(_ context.Context, token, userID string) error {
			capturedToken = token
			capturedUser = userID
			return nil
		},
	}
	if err := repo.AcceptInvitation(context.Background(), "tok-x", "u1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedToken != "tok-x" {
		t.Errorf("token: got %q, want %q", capturedToken, "tok-x")
	}
	if capturedUser != "u1" {
		t.Errorf("userID: got %q, want %q", capturedUser, "u1")
	}
}

func TestMockInvitationRepository_DeleteInvitationDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockInvitationRepository{}
	if err := repo.DeleteInvitation(context.Background(), "inv-1"); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestMockInvitationRepository_DeleteInvitationFnCalled(t *testing.T) {
	var capturedID string
	repo := &testutil.MockInvitationRepository{
		DeleteInvitationFn: func(_ context.Context, id string) error {
			capturedID = id
			return nil
		},
	}
	if err := repo.DeleteInvitation(context.Background(), "inv-42"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "inv-42" {
		t.Errorf("id: got %q, want %q", capturedID, "inv-42")
	}
}

func TestMockInvitationRepository_UpdateInvitationStatusDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockInvitationRepository{}
	if err := repo.UpdateInvitationStatus(context.Background(), "inv-1", models.InvitationStatusAccepted); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestMockInvitationRepository_UpdateInvitationStatusFnCalled(t *testing.T) {
	var capturedStatus models.InvitationStatus
	repo := &testutil.MockInvitationRepository{
		UpdateInvitationStatusFn: func(_ context.Context, _ string, status models.InvitationStatus) error {
			capturedStatus = status
			return nil
		},
	}
	if err := repo.UpdateInvitationStatus(context.Background(), "inv-1", models.InvitationStatusExpired); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedStatus != models.InvitationStatusExpired {
		t.Errorf("status: got %q, want %q", capturedStatus, models.InvitationStatusExpired)
	}
}

func TestMockInvitationRepository_CancelInvitationDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockInvitationRepository{}
	if err := repo.CancelInvitation(context.Background(), "inv-1"); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestMockInvitationRepository_CancelInvitationFnCalled(t *testing.T) {
	called := false
	repo := &testutil.MockInvitationRepository{
		CancelInvitationFn: func(_ context.Context, _ string) error {
			called = true
			return nil
		},
	}
	if err := repo.CancelInvitation(context.Background(), "inv-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("CancelInvitationFn was not called")
	}
}

func TestMockInvitationRepository_GetPendingDefaultsToEmpty(t *testing.T) {
	repo := &testutil.MockInvitationRepository{}
	invs, err := repo.GetPendingInvitationsByProject(context.Background(), "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(invs) != 0 {
		t.Errorf("expected empty slice, got %d", len(invs))
	}
}

func TestMockInvitationRepository_GetPendingFnFilters(t *testing.T) {
	all := []models.Invitation{
		{Id: "inv-1", ProjectId: "p1", Status: models.InvitationStatusPending},
		{Id: "inv-2", ProjectId: "p1", Status: models.InvitationStatusAccepted},
		{Id: "inv-3", ProjectId: "p1", Status: models.InvitationStatusPending},
	}
	repo := &testutil.MockInvitationRepository{
		GetPendingInvitationsByProjectFn: func(_ context.Context, projectID string) ([]models.Invitation, error) {
			var out []models.Invitation
			for _, inv := range all {
				if inv.ProjectId == projectID && inv.Status == models.InvitationStatusPending {
					out = append(out, inv)
				}
			}
			return out, nil
		},
	}
	got, err := repo.GetPendingInvitationsByProject(context.Background(), "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 pending invitations, got %d", len(got))
	}
}