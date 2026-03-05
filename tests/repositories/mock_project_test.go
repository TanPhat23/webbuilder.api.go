package repositories_test

import (
	"context"
	"errors"
	"testing"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/tests/testutil"
)

func TestMockProjectRepository_DefaultsReturnSentinel(t *testing.T) {
	repo := &testutil.MockProjectRepository{}
	ctx := context.Background()

	_, err := repo.GetProjectByID(ctx, "p1", "u1")
	if !errors.Is(err, repositories.ErrProjectNotFound) {
		t.Errorf("GetProjectByID default: want ErrProjectNotFound, got %v", err)
	}

	_, err = repo.GetProjectWithAccess(ctx, "p1", "u1")
	if !errors.Is(err, repositories.ErrProjectNotFound) {
		t.Errorf("GetProjectWithAccess default: want ErrProjectNotFound, got %v", err)
	}

	_, err = repo.GetPublicProjectByID(ctx, "p1")
	if !errors.Is(err, repositories.ErrProjectNotFound) {
		t.Errorf("GetPublicProjectByID default: want ErrProjectNotFound, got %v", err)
	}

	_, err = repo.GetProjectWithLock(ctx, "p1", "u1")
	if !errors.Is(err, repositories.ErrProjectNotFound) {
		t.Errorf("GetProjectWithLock default: want ErrProjectNotFound, got %v", err)
	}

	_, err = repo.UpdateProject(ctx, "p1", "u1", map[string]any{})
	if !errors.Is(err, repositories.ErrProjectNotFound) {
		t.Errorf("UpdateProject default: want ErrProjectNotFound, got %v", err)
	}
}

func TestMockProjectRepository_CreateProjectFnIsInvoked(t *testing.T) {
	called := false
	repo := &testutil.MockProjectRepository{
		CreateProjectFn: func(_ context.Context, p *models.Project) error {
			called = true
			p.ID = "generated-id"
			return nil
		},
	}

	proj := &models.Project{Name: "Test", OwnerId: "u1"}
	if err := repo.CreateProject(context.Background(), proj); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("CreateProjectFn was not called")
	}
	if proj.ID != "generated-id" {
		t.Errorf("ID: got %q, want %q", proj.ID, "generated-id")
	}
}

func TestMockProjectRepository_CreateProjectDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockProjectRepository{}
	proj := &models.Project{Name: "Test", OwnerId: "u1"}
	if err := repo.CreateProject(context.Background(), proj); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestMockProjectRepository_UpdateProjectFnReturnsProject(t *testing.T) {
	want := &models.Project{ID: "p1", Name: "Updated"}
	repo := &testutil.MockProjectRepository{
		UpdateProjectFn: func(_ context.Context, projectID, _ string, _ map[string]any) (*models.Project, error) {
			if projectID == "p1" {
				return want, nil
			}
			return nil, repositories.ErrProjectNotFound
		},
	}

	got, err := repo.UpdateProject(context.Background(), "p1", "u1", map[string]any{"Name": "Updated"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != want.Name {
		t.Errorf("Name: got %q, want %q", got.Name, want.Name)
	}

	_, err = repo.UpdateProject(context.Background(), "p-missing", "u1", map[string]any{})
	if !errors.Is(err, repositories.ErrProjectNotFound) {
		t.Errorf("missing project: want ErrProjectNotFound, got %v", err)
	}
}

func TestMockProjectRepository_ExistsForUserReturnsFalseByDefault(t *testing.T) {
	repo := &testutil.MockProjectRepository{}
	exists, err := repo.ExistsForUser(context.Background(), "p1", "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected false by default, got true")
	}
}

func TestMockProjectRepository_ExistsForUserFnReturnsTrue(t *testing.T) {
	repo := &testutil.MockProjectRepository{
		ExistsForUserFn: func(_ context.Context, projectID, userID string) (bool, error) {
			return projectID == "p1" && userID == "u1", nil
		},
	}

	exists, err := repo.ExistsForUser(context.Background(), "p1", "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected true for matching project+user")
	}

	exists, err = repo.ExistsForUser(context.Background(), "p1", "u2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected false for non-matching user")
	}
}

func TestMockProjectRepository_GetProjectsByUserIDDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockProjectRepository{}
	projects, err := repo.GetProjectsByUserID(context.Background(), "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(projects) != 0 {
		t.Errorf("expected empty slice, got %d", len(projects))
	}
}

func TestMockProjectRepository_GetProjectsByUserIDFnFilters(t *testing.T) {
	all := []models.Project{
		{ID: "p1", OwnerId: "u1"},
		{ID: "p2", OwnerId: "u2"},
		{ID: "p3", OwnerId: "u1"},
	}
	repo := &testutil.MockProjectRepository{
		GetProjectsByUserIDFn: func(_ context.Context, userID string) ([]models.Project, error) {
			var out []models.Project
			for _, p := range all {
				if p.OwnerId == userID {
					out = append(out, p)
				}
			}
			return out, nil
		},
	}

	got, err := repo.GetProjectsByUserID(context.Background(), "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 projects for u1, got %d", len(got))
	}
}

func TestMockProjectRepository_GetCollaboratorProjectsDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockProjectRepository{}
	projects, err := repo.GetCollaboratorProjects(context.Background(), "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(projects) != 0 {
		t.Errorf("expected empty slice, got %d", len(projects))
	}
}

func TestMockProjectRepository_GetCollaboratorProjectsFnFilters(t *testing.T) {
	all := []models.Project{
		{ID: "p1", Name: "Alpha"},
		{ID: "p2", Name: "Beta"},
	}
	repo := &testutil.MockProjectRepository{
		GetCollaboratorProjectsFn: func(_ context.Context, _ string) ([]models.Project, error) {
			return all, nil
		},
	}

	got, err := repo.GetCollaboratorProjects(context.Background(), "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != len(all) {
		t.Errorf("expected %d projects, got %d", len(all), len(got))
	}
}

func TestMockProjectRepository_GetProjectPagesDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockProjectRepository{}
	pages, err := repo.GetProjectPages(context.Background(), "p1", "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pages) != 0 {
		t.Errorf("expected empty slice, got %d", len(pages))
	}
}

func TestMockProjectRepository_GetProjectPagesFnReturnsPages(t *testing.T) {
	want := []models.Page{
		{Id: "pg-1", Name: "Home", ProjectId: "p1"},
		{Id: "pg-2", Name: "About", ProjectId: "p1"},
	}
	repo := &testutil.MockProjectRepository{
		GetProjectPagesFn: func(_ context.Context, projectID, _ string) ([]models.Page, error) {
			if projectID == "p1" {
				return want, nil
			}
			return []models.Page{}, nil
		},
	}

	got, err := repo.GetProjectPages(context.Background(), "p1", "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != len(want) {
		t.Errorf("expected %d pages, got %d", len(want), len(got))
	}
}

func TestMockProjectRepository_DeleteProjectDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockProjectRepository{}
	if err := repo.DeleteProject(context.Background(), "p1", "u1"); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestMockProjectRepository_DeleteProjectFnCalled(t *testing.T) {
	var capturedProject, capturedUser string
	repo := &testutil.MockProjectRepository{
		DeleteProjectFn: func(_ context.Context, projectID, userID string) error {
			capturedProject = projectID
			capturedUser = userID
			return nil
		},
	}

	if err := repo.DeleteProject(context.Background(), "p-42", "u1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedProject != "p-42" {
		t.Errorf("projectID: got %q, want %q", capturedProject, "p-42")
	}
	if capturedUser != "u1" {
		t.Errorf("userID: got %q, want %q", capturedUser, "u1")
	}
}

func TestMockProjectRepository_HardDeleteProjectDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockProjectRepository{}
	if err := repo.HardDeleteProject(context.Background(), "p1", "u1"); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestMockProjectRepository_HardDeleteProjectFnCalled(t *testing.T) {
	called := false
	repo := &testutil.MockProjectRepository{
		HardDeleteProjectFn: func(_ context.Context, _, _ string) error {
			called = true
			return nil
		},
	}
	if err := repo.HardDeleteProject(context.Background(), "p1", "u1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("HardDeleteProjectFn was not called")
	}
}

func TestMockProjectRepository_RestoreProjectDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockProjectRepository{}
	if err := repo.RestoreProject(context.Background(), "p1", "u1"); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestMockProjectRepository_RestoreProjectFnCalled(t *testing.T) {
	called := false
	repo := &testutil.MockProjectRepository{
		RestoreProjectFn: func(_ context.Context, _, _ string) error {
			called = true
			return nil
		},
	}
	if err := repo.RestoreProject(context.Background(), "p1", "u1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("RestoreProjectFn was not called")
	}
}

func TestMockProjectRepository_GetProjectWithLockFnReturnsProject(t *testing.T) {
	want := &models.Project{ID: "p1", Name: "Locked"}
	repo := &testutil.MockProjectRepository{
		GetProjectWithLockFn: func(_ context.Context, projectID, _ string) (*models.Project, error) {
			if projectID == "p1" {
				return want, nil
			}
			return nil, repositories.ErrProjectNotFound
		},
	}

	got, err := repo.GetProjectWithLock(context.Background(), "p1", "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != want.ID {
		t.Errorf("ID: got %q, want %q", got.ID, want.ID)
	}

	_, err = repo.GetProjectWithLock(context.Background(), "p-missing", "u1")
	if !errors.Is(err, repositories.ErrProjectNotFound) {
		t.Errorf("missing project: want ErrProjectNotFound, got %v", err)
	}
}