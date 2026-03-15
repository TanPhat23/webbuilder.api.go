package repositories_test

import (
	"context"
	"testing"

	"my-go-app/internal/models"
	"my-go-app/tests/testutil"
)

func TestMockCollaboratorRepository_IsCollaboratorReturnsFalseByDefault(t *testing.T) {
	repo := &testutil.MockCollaboratorRepository{}
	ok, err := repo.IsCollaborator(context.Background(), "p1", "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected false by default")
	}
}

func TestMockCollaboratorRepository_IsCollaboratorFnOverride(t *testing.T) {
	collabs := map[string]bool{"p1:u1": true, "p1:u2": true}
	repo := &testutil.MockCollaboratorRepository{
		IsCollaboratorFn: func(_ context.Context, projectID, userID string) (bool, error) {
			return collabs[projectID+":"+userID], nil
		},
	}

	for _, tc := range []struct {
		project, user string
		want          bool
	}{
		{"p1", "u1", true},
		{"p1", "u2", true},
		{"p1", "u3", false},
		{"p2", "u1", false},
	} {
		got, err := repo.IsCollaborator(context.Background(), tc.project, tc.user)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != tc.want {
			t.Errorf("IsCollaborator(%q, %q): got %v, want %v", tc.project, tc.user, got, tc.want)
		}
	}
}

func TestMockCollaboratorRepository_UpdateCollaboratorRoleFnCalled(t *testing.T) {
	var capturedRole models.CollaboratorRole
	repo := &testutil.MockCollaboratorRepository{
		UpdateCollaboratorRoleFn: func(_ context.Context, id string, role models.CollaboratorRole) error {
			capturedRole = role
			return nil
		},
	}

	if err := repo.UpdateCollaboratorRole(context.Background(), "collab-1", models.RoleViewer); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedRole != models.RoleViewer {
		t.Errorf("role: got %q, want %q", capturedRole, models.RoleViewer)
	}
}

func TestMockCollaboratorRepository_CreateCollaboratorDefaultReturnsInput(t *testing.T) {
	repo := &testutil.MockCollaboratorRepository{}
	input := &models.Collaborator{Id: "collab-1", ProjectId: "p1", UserId: "u1"}
	got, err := repo.CreateCollaborator(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Id != input.Id {
		t.Errorf("Id: got %q, want %q", got.Id, input.Id)
	}
}

func TestMockCollaboratorRepository_CreateCollaboratorFnAssignsID(t *testing.T) {
	repo := &testutil.MockCollaboratorRepository{
		CreateCollaboratorFn: func(_ context.Context, c *models.Collaborator) (*models.Collaborator, error) {
			c.Id = "collab-generated"
			return c, nil
		},
	}
	collab := &models.Collaborator{ProjectId: "p1", UserId: "u1"}
	got, err := repo.CreateCollaborator(context.Background(), collab)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Id != "collab-generated" {
		t.Errorf("Id: got %q, want %q", got.Id, "collab-generated")
	}
}

func TestMockCollaboratorRepository_GetCollaboratorsByProjectDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockCollaboratorRepository{}
	collabs, err := repo.GetCollaboratorsByProject(context.Background(), "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(collabs) != 0 {
		t.Errorf("expected empty slice, got %d", len(collabs))
	}
}

func TestMockCollaboratorRepository_GetCollaboratorsByProjectFnFilters(t *testing.T) {
	all := []models.Collaborator{
		{Id: "c1", ProjectId: "p1"},
		{Id: "c2", ProjectId: "p2"},
		{Id: "c3", ProjectId: "p1"},
	}
	repo := &testutil.MockCollaboratorRepository{
		GetCollaboratorsByProjectFn: func(_ context.Context, projectID string) ([]models.Collaborator, error) {
			var out []models.Collaborator
			for _, c := range all {
				if c.ProjectId == projectID {
					out = append(out, c)
				}
			}
			return out, nil
		},
	}

	got, err := repo.GetCollaboratorsByProject(context.Background(), "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 collaborators for p1, got %d", len(got))
	}
}

func TestMockCollaboratorRepository_GetCollaboratorByIDDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockCollaboratorRepository{}
	got, err := repo.GetCollaboratorByID(context.Background(), "collab-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil by default, got %+v", got)
	}
}

func TestMockCollaboratorRepository_GetCollaboratorByIDFnReturnsCollaborator(t *testing.T) {
	want := &models.Collaborator{Id: "collab-1", ProjectId: "p1", UserId: "u1"}
	repo := &testutil.MockCollaboratorRepository{
		GetCollaboratorByIDFn: func(_ context.Context, id string) (*models.Collaborator, error) {
			if id == "collab-1" {
				return want, nil
			}
			return nil, nil
		},
	}

	got, err := repo.GetCollaboratorByID(context.Background(), "collab-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Id != want.Id {
		t.Errorf("Id: got %q, want %q", got.Id, want.Id)
	}

	none, err := repo.GetCollaboratorByID(context.Background(), "collab-missing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if none != nil {
		t.Errorf("expected nil for missing ID, got %+v", none)
	}
}

func TestMockCollaboratorRepository_DeleteCollaboratorDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockCollaboratorRepository{}
	if err := repo.DeleteCollaborator(context.Background(), "collab-1"); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestMockCollaboratorRepository_DeleteCollaboratorFnCalled(t *testing.T) {
	var capturedID string
	repo := &testutil.MockCollaboratorRepository{
		DeleteCollaboratorFn: func(_ context.Context, id string) error {
			capturedID = id
			return nil
		},
	}

	if err := repo.DeleteCollaborator(context.Background(), "collab-99"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "collab-99" {
		t.Errorf("id: got %q, want %q", capturedID, "collab-99")
	}
}