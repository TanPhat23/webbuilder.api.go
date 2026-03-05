package repositories_test

import (
	"context"
	"errors"
	"testing"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/tests/testutil"
)

// ─── MockProjectRepository (extended) ────────────────────────────────────────

func TestMockProjectRepository_GetCollaboratorProjectsDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockProjectRepository{}
	got, err := repo.GetCollaboratorProjects(context.Background(), "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d", len(got))
	}
}

func TestMockProjectRepository_GetCollaboratorProjectsFnFilters(t *testing.T) {
	all := []models.Project{
		{ID: "p1", OwnerId: "u2"},
		{ID: "p2", OwnerId: "u2"},
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
	if len(got) != 2 {
		t.Errorf("expected 2, got %d", len(got))
	}
}

func TestMockProjectRepository_GetProjectPagesDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockProjectRepository{}
	got, err := repo.GetProjectPages(context.Background(), "p1", "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d", len(got))
	}
}

func TestMockProjectRepository_GetProjectPagesFnReturnsPages(t *testing.T) {
	want := []models.Page{
		{Id: "pg-1", ProjectId: "p1"},
		{Id: "pg-2", ProjectId: "p1"},
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
	if len(got) != 2 {
		t.Errorf("expected 2 pages, got %d", len(got))
	}
}

func TestMockProjectRepository_DeleteProjectDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockProjectRepository{}
	if err := repo.DeleteProject(context.Background(), "p1", "u1"); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestMockProjectRepository_DeleteProjectFnCalled(t *testing.T) {
	var capturedID string
	repo := &testutil.MockProjectRepository{
		DeleteProjectFn: func(_ context.Context, projectID, _ string) error {
			capturedID = projectID
			return nil
		},
	}
	if err := repo.DeleteProject(context.Background(), "p-del", "u1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "p-del" {
		t.Errorf("captured ID: got %q, want %q", capturedID, "p-del")
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
	var capturedID string
	repo := &testutil.MockProjectRepository{
		RestoreProjectFn: func(_ context.Context, projectID, _ string) error {
			capturedID = projectID
			return nil
		},
	}
	if err := repo.RestoreProject(context.Background(), "p-restore", "u1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "p-restore" {
		t.Errorf("captured ID: got %q, want %q", capturedID, "p-restore")
	}
}

func TestMockProjectRepository_GetProjectWithLockDefaultReturnsSentinel(t *testing.T) {
	repo := &testutil.MockProjectRepository{}
	_, err := repo.GetProjectWithLock(context.Background(), "p1", "u1")
	if !errors.Is(err, repositories.ErrProjectNotFound) {
		t.Errorf("want ErrProjectNotFound, got %v", err)
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
}

// ─── MockImageRepository (extended) ──────────────────────────────────────────

func TestMockImageRepository_GetImageByIDDefaultReturnsSentinel(t *testing.T) {
	repo := &testutil.MockImageRepository{}
	_, err := repo.GetImageByID(context.Background(), "img-1", "u1")
	if !errors.Is(err, repositories.ErrImageNotFound) {
		t.Errorf("want ErrImageNotFound, got %v", err)
	}
}

func TestMockImageRepository_GetImageByIDFnReturnsImage(t *testing.T) {
	want := &models.Image{ImageId: "img-1", UserId: "u1"}
	repo := &testutil.MockImageRepository{
		GetImageByIDFn: func(_ context.Context, imageID, userID string) (*models.Image, error) {
			if imageID == "img-1" && userID == "u1" {
				return want, nil
			}
			return nil, repositories.ErrImageNotFound
		},
	}
	got, err := repo.GetImageByID(context.Background(), "img-1", "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ImageId != want.ImageId {
		t.Errorf("ImageId: got %q, want %q", got.ImageId, want.ImageId)
	}

	_, err = repo.GetImageByID(context.Background(), "img-missing", "u1")
	if !errors.Is(err, repositories.ErrImageNotFound) {
		t.Errorf("missing: want ErrImageNotFound, got %v", err)
	}
}

func TestMockImageRepository_DeleteImageDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockImageRepository{}
	if err := repo.DeleteImage(context.Background(), "img-1", "u1"); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestMockImageRepository_DeleteImageFnCalled(t *testing.T) {
	var capturedID string
	repo := &testutil.MockImageRepository{
		DeleteImageFn: func(_ context.Context, imageID, _ string) error {
			capturedID = imageID
			return nil
		},
	}
	if err := repo.DeleteImage(context.Background(), "img-del", "u1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "img-del" {
		t.Errorf("captured ID: got %q, want %q", capturedID, "img-del")
	}
}

func TestMockImageRepository_GetAllImagesFnPaginates(t *testing.T) {
	all := []models.Image{
		{ImageId: "img-1"}, {ImageId: "img-2"}, {ImageId: "img-3"},
	}
	repo := &testutil.MockImageRepository{
		GetAllImagesFn: func(_ context.Context, limit, offset int) ([]models.Image, error) {
			if offset >= len(all) {
				return []models.Image{}, nil
			}
			end := offset + limit
			if end > len(all) {
				end = len(all)
			}
			return all[offset:end], nil
		},
	}
	got, err := repo.GetAllImages(context.Background(), 2, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2, got %d", len(got))
	}
	if got[0].ImageId != "img-2" {
		t.Errorf("first result: got %q, want %q", got[0].ImageId, "img-2")
	}
}

// ─── MockPageRepository (extended) ───────────────────────────────────────────

func TestMockPageRepository_GetPageByIDDefaultReturnsSentinel(t *testing.T) {
	repo := &testutil.MockPageRepository{}
	_, err := repo.GetPageByID(context.Background(), "pg-1", "p1")
	if !errors.Is(err, repositories.ErrPageNotFound) {
		t.Errorf("want ErrPageNotFound, got %v", err)
	}
}

func TestMockPageRepository_GetPageByIDFnReturnsPage(t *testing.T) {
	want := &models.Page{Id: "pg-1", ProjectId: "p1"}
	repo := &testutil.MockPageRepository{
		GetPageByIDFn: func(_ context.Context, pageID, _ string) (*models.Page, error) {
			if pageID == "pg-1" {
				return want, nil
			}
			return nil, repositories.ErrPageNotFound
		},
	}
	got, err := repo.GetPageByID(context.Background(), "pg-1", "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Id != want.Id {
		t.Errorf("Id: got %q, want %q", got.Id, want.Id)
	}

	_, err = repo.GetPageByID(context.Background(), "pg-missing", "p1")
	if !errors.Is(err, repositories.ErrPageNotFound) {
		t.Errorf("missing: want ErrPageNotFound, got %v", err)
	}
}

func TestMockPageRepository_UpdatePageDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockPageRepository{}
	if err := repo.UpdatePage(context.Background(), &models.Page{Id: "pg-1"}); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestMockPageRepository_UpdatePageFnCalled(t *testing.T) {
	var capturedID string
	repo := &testutil.MockPageRepository{
		UpdatePageFn: func(_ context.Context, page *models.Page) error {
			capturedID = page.Id
			return nil
		},
	}
	if err := repo.UpdatePage(context.Background(), &models.Page{Id: "pg-42"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "pg-42" {
		t.Errorf("captured ID: got %q, want %q", capturedID, "pg-42")
	}
}

func TestMockPageRepository_UpdatePageFieldsDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockPageRepository{}
	if err := repo.UpdatePageFields(context.Background(), "pg-1", map[string]any{"Name": "Home"}); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestMockPageRepository_UpdatePageFieldsFnReceivesUpdates(t *testing.T) {
	var capturedUpdates map[string]any
	repo := &testutil.MockPageRepository{
		UpdatePageFieldsFn: func(_ context.Context, _ string, updates map[string]any) error {
			capturedUpdates = updates
			return nil
		},
	}
	updates := map[string]any{"Name": "About", "Slug": "about"}
	if err := repo.UpdatePageFields(context.Background(), "pg-1", updates); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedUpdates["Name"] != "About" {
		t.Errorf("Name: got %v, want %q", capturedUpdates["Name"], "About")
	}
}

func TestMockPageRepository_DeletePageDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockPageRepository{}
	if err := repo.DeletePage(context.Background(), "pg-1"); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestMockPageRepository_DeletePageFnCalled(t *testing.T) {
	var capturedID string
	repo := &testutil.MockPageRepository{
		DeletePageFn: func(_ context.Context, pageID string) error {
			capturedID = pageID
			return nil
		},
	}
	if err := repo.DeletePage(context.Background(), "pg-99"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "pg-99" {
		t.Errorf("captured ID: got %q, want %q", capturedID, "pg-99")
	}
}

func TestMockPageRepository_DeletePageByProjectIDDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockPageRepository{}
	if err := repo.DeletePageByProjectID(context.Background(), "pg-1", "p1", "u1"); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestMockPageRepository_DeletePageByProjectIDFnCalled(t *testing.T) {
	var capturedPageID, capturedProjectID string
	repo := &testutil.MockPageRepository{
		DeletePageByProjectIDFn: func(_ context.Context, pageID, projectID, _ string) error {
			capturedPageID = pageID
			capturedProjectID = projectID
			return nil
		},
	}
	if err := repo.DeletePageByProjectID(context.Background(), "pg-5", "p-5", "u1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedPageID != "pg-5" {
		t.Errorf("pageID: got %q, want %q", capturedPageID, "pg-5")
	}
	if capturedProjectID != "p-5" {
		t.Errorf("projectID: got %q, want %q", capturedProjectID, "p-5")
	}
}

// ─── MockSnapshotRepository (extended) ───────────────────────────────────────

func TestMockSnapshotRepository_GetSnapshotByIDDefaultReturnsSentinel(t *testing.T) {
	repo := &testutil.MockSnapshotRepository{}
	_, err := repo.GetSnapshotByID(context.Background(), "snap-1")
	if !errors.Is(err, repositories.ErrSnapshotNotFound) {
		t.Errorf("want ErrSnapshotNotFound, got %v", err)
	}
}

func TestMockSnapshotRepository_GetSnapshotByIDFnReturnsSnapshot(t *testing.T) {
	want := &models.Snapshot{Id: "snap-1", ProjectId: "p1", Name: "v1"}
	repo := &testutil.MockSnapshotRepository{
		GetSnapshotByIDFn: func(_ context.Context, id string) (*models.Snapshot, error) {
			if id == "snap-1" {
				return want, nil
			}
			return nil, repositories.ErrSnapshotNotFound
		},
	}
	got, err := repo.GetSnapshotByID(context.Background(), "snap-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Id != want.Id {
		t.Errorf("Id: got %q, want %q", got.Id, want.Id)
	}

	_, err = repo.GetSnapshotByID(context.Background(), "snap-missing")
	if !errors.Is(err, repositories.ErrSnapshotNotFound) {
		t.Errorf("missing: want ErrSnapshotNotFound, got %v", err)
	}
}

func TestMockSnapshotRepository_DeleteSnapshotDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockSnapshotRepository{}
	if err := repo.DeleteSnapshot(context.Background(), "snap-1"); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestMockSnapshotRepository_DeleteSnapshotFnCalled(t *testing.T) {
	var capturedID string
	repo := &testutil.MockSnapshotRepository{
		DeleteSnapshotFn: func(_ context.Context, id string) error {
			capturedID = id
			return nil
		},
	}
	if err := repo.DeleteSnapshot(context.Background(), "snap-del"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "snap-del" {
		t.Errorf("captured ID: got %q, want %q", capturedID, "snap-del")
	}
}

func TestMockSnapshotRepository_GetSnapshotsByProjectIDFnFilters(t *testing.T) {
	all := []models.Snapshot{
		{Id: "s1", ProjectId: "p1"},
		{Id: "s2", ProjectId: "p2"},
		{Id: "s3", ProjectId: "p1"},
	}
	repo := &testutil.MockSnapshotRepository{
		GetSnapshotsByProjectIDFn: func(_ context.Context, projectID string) ([]models.Snapshot, error) {
			var out []models.Snapshot
			for _, s := range all {
				if s.ProjectId == projectID {
					out = append(out, s)
				}
			}
			return out, nil
		},
	}
	got, err := repo.GetSnapshotsByProjectID(context.Background(), "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 snapshots for p1, got %d", len(got))
	}
}

// ─── MockCollaboratorRepository (extended) ───────────────────────────────────

func TestMockCollaboratorRepository_CreateCollaboratorDefaultReturnsInput(t *testing.T) {
	repo := &testutil.MockCollaboratorRepository{}
	input := &models.Collaborator{Id: "col-1", UserId: "u1", ProjectId: "p1"}
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
			c.Id = "col-generated"
			return c, nil
		},
	}
	col := &models.Collaborator{UserId: "u1", ProjectId: "p1", Role: models.RoleEditor}
	got, err := repo.CreateCollaborator(context.Background(), col)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Id != "col-generated" {
		t.Errorf("Id: got %q, want %q", got.Id, "col-generated")
	}
}

func TestMockCollaboratorRepository_GetCollaboratorsByProjectDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockCollaboratorRepository{}
	got, err := repo.GetCollaboratorsByProject(context.Background(), "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d", len(got))
	}
}

func TestMockCollaboratorRepository_GetCollaboratorsByProjectFnFilters(t *testing.T) {
	all := []models.Collaborator{
		{Id: "col-1", ProjectId: "p1", UserId: "u1"},
		{Id: "col-2", ProjectId: "p2", UserId: "u2"},
		{Id: "col-3", ProjectId: "p1", UserId: "u3"},
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
	got, err := repo.GetCollaboratorByID(context.Background(), "col-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

func TestMockCollaboratorRepository_GetCollaboratorByIDFnReturnsCollaborator(t *testing.T) {
	want := &models.Collaborator{Id: "col-1", UserId: "u1", ProjectId: "p1"}
	repo := &testutil.MockCollaboratorRepository{
		GetCollaboratorByIDFn: func(_ context.Context, id string) (*models.Collaborator, error) {
			if id == "col-1" {
				return want, nil
			}
			return nil, nil
		},
	}
	got, err := repo.GetCollaboratorByID(context.Background(), "col-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Id != want.Id {
		t.Errorf("Id: got %q, want %q", got.Id, want.Id)
	}
}

func TestMockCollaboratorRepository_DeleteCollaboratorDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockCollaboratorRepository{}
	if err := repo.DeleteCollaborator(context.Background(), "col-1"); err != nil {
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
	if err := repo.DeleteCollaborator(context.Background(), "col-del"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "col-del" {
		t.Errorf("captured ID: got %q, want %q", capturedID, "col-del")
	}
}

// ─── MockEventWorkflowRepository (extended) ──────────────────────────────────

func TestMockEventWorkflowRepository_GetEventWorkflowByIDDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockEventWorkflowRepository{}
	got, err := repo.GetEventWorkflowByID(context.Background(), "wf-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

func TestMockEventWorkflowRepository_GetEventWorkflowByIDFnReturnsWorkflow(t *testing.T) {
	want := &models.EventWorkflow{Id: "wf-1", Name: "On Click", ProjectId: "p1"}
	repo := &testutil.MockEventWorkflowRepository{
		GetEventWorkflowByIDFn: func(_ context.Context, id string) (*models.EventWorkflow, error) {
			if id == "wf-1" {
				return want, nil
			}
			return nil, nil
		},
	}
	got, err := repo.GetEventWorkflowByID(context.Background(), "wf-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Id != want.Id {
		t.Errorf("Id: got %q, want %q", got.Id, want.Id)
	}
}

func TestMockEventWorkflowRepository_GetEventWorkflowsByProjectIDDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockEventWorkflowRepository{}
	got, err := repo.GetEventWorkflowsByProjectID(context.Background(), "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d", len(got))
	}
}

func TestMockEventWorkflowRepository_GetEventWorkflowsByProjectIDFnFilters(t *testing.T) {
	all := []models.EventWorkflow{
		{Id: "wf-1", ProjectId: "p1"},
		{Id: "wf-2", ProjectId: "p2"},
		{Id: "wf-3", ProjectId: "p1"},
	}
	repo := &testutil.MockEventWorkflowRepository{
		GetEventWorkflowsByProjectIDFn: func(_ context.Context, projectID string) ([]models.EventWorkflow, error) {
			var out []models.EventWorkflow
			for _, wf := range all {
				if wf.ProjectId == projectID {
					out = append(out, wf)
				}
			}
			return out, nil
		},
	}
	got, err := repo.GetEventWorkflowsByProjectID(context.Background(), "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 workflows for p1, got %d", len(got))
	}
}

func TestMockEventWorkflowRepository_GetEventWorkflowsByProjectIDWithElementsDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockEventWorkflowRepository{}
	got, err := repo.GetEventWorkflowsByProjectIDWithElements(context.Background(), "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d", len(got))
	}
}

func TestMockEventWorkflowRepository_GetEventWorkflowsByNameDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockEventWorkflowRepository{}
	got, err := repo.GetEventWorkflowsByName(context.Background(), "p1", "On Click")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d", len(got))
	}
}

func TestMockEventWorkflowRepository_GetEventWorkflowsByNameFnFilters(t *testing.T) {
	all := []models.EventWorkflow{
		{Id: "wf-1", ProjectId: "p1", Name: "On Click"},
		{Id: "wf-2", ProjectId: "p1", Name: "On Submit"},
		{Id: "wf-3", ProjectId: "p1", Name: "On Click"},
	}
	repo := &testutil.MockEventWorkflowRepository{
		GetEventWorkflowsByNameFn: func(_ context.Context, projectID, name string) ([]models.EventWorkflow, error) {
			var out []models.EventWorkflow
			for _, wf := range all {
				if wf.ProjectId == projectID && wf.Name == name {
					out = append(out, wf)
				}
			}
			return out, nil
		},
	}
	got, err := repo.GetEventWorkflowsByName(context.Background(), "p1", "On Click")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 workflows named 'On Click', got %d", len(got))
	}
}

func TestMockEventWorkflowRepository_UpdateEventWorkflowDefaultReturnsInput(t *testing.T) {
	repo := &testutil.MockEventWorkflowRepository{}
	input := &models.EventWorkflow{Id: "wf-1", Name: "Updated"}
	got, err := repo.UpdateEventWorkflow(context.Background(), "wf-1", input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Id != input.Id {
		t.Errorf("Id: got %q, want %q", got.Id, input.Id)
	}
}

func TestMockEventWorkflowRepository_UpdateEventWorkflowFnReturnsUpdated(t *testing.T) {
	want := &models.EventWorkflow{Id: "wf-1", Name: "Renamed"}
	repo := &testutil.MockEventWorkflowRepository{
		UpdateEventWorkflowFn: func(_ context.Context, id string, _ *models.EventWorkflow) (*models.EventWorkflow, error) {
			if id == "wf-1" {
				return want, nil
			}
			return nil, errors.New("not found")
		},
	}
	got, err := repo.UpdateEventWorkflow(context.Background(), "wf-1", &models.EventWorkflow{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != want.Name {
		t.Errorf("Name: got %q, want %q", got.Name, want.Name)
	}
}

func TestMockEventWorkflowRepository_UpdateEventWorkflowEnabledDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockEventWorkflowRepository{}
	if err := repo.UpdateEventWorkflowEnabled(context.Background(), "wf-1", true); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestMockEventWorkflowRepository_UpdateEventWorkflowEnabledFnCalled(t *testing.T) {
	var capturedID string
	var capturedEnabled bool
	repo := &testutil.MockEventWorkflowRepository{
		UpdateEventWorkflowEnabledFn: func(_ context.Context, id string, enabled bool) error {
			capturedID = id
			capturedEnabled = enabled
			return nil
		},
	}
	if err := repo.UpdateEventWorkflowEnabled(context.Background(), "wf-42", false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "wf-42" {
		t.Errorf("id: got %q, want %q", capturedID, "wf-42")
	}
	if capturedEnabled {
		t.Error("expected enabled=false, got true")
	}
}

func TestMockEventWorkflowRepository_DeleteEventWorkflowDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockEventWorkflowRepository{}
	if err := repo.DeleteEventWorkflow(context.Background(), "wf-1"); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestMockEventWorkflowRepository_DeleteEventWorkflowFnCalled(t *testing.T) {
	var capturedID string
	repo := &testutil.MockEventWorkflowRepository{
		DeleteEventWorkflowFn: func(_ context.Context, id string) error {
			capturedID = id
			return nil
		},
	}
	if err := repo.DeleteEventWorkflow(context.Background(), "wf-del"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "wf-del" {
		t.Errorf("captured ID: got %q, want %q", capturedID, "wf-del")
	}
}

func TestMockEventWorkflowRepository_DeleteEventWorkflowsByProjectIDDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockEventWorkflowRepository{}
	if err := repo.DeleteEventWorkflowsByProjectID(context.Background(), "p1"); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestMockEventWorkflowRepository_DeleteEventWorkflowsByProjectIDFnCalled(t *testing.T) {
	var capturedProjectID string
	repo := &testutil.MockEventWorkflowRepository{
		DeleteEventWorkflowsByProjectIDFn: func(_ context.Context, projectID string) error {
			capturedProjectID = projectID
			return nil
		},
	}
	if err := repo.DeleteEventWorkflowsByProjectID(context.Background(), "p-clean"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedProjectID != "p-clean" {
		t.Errorf("captured projectID: got %q, want %q", capturedProjectID, "p-clean")
	}
}

func TestMockEventWorkflowRepository_GetEventWorkflowsWithFiltersDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockEventWorkflowRepository{}
	got, err := repo.GetEventWorkflowsWithFilters(context.Background(), "p1", nil, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d", len(got))
	}
}

func TestMockEventWorkflowRepository_GetEventWorkflowsWithFiltersFnFiltersEnabledAndName(t *testing.T) {
	enabled := true
	all := []models.EventWorkflow{
		{Id: "wf-1", ProjectId: "p1", Name: "On Click", Enabled: true},
		{Id: "wf-2", ProjectId: "p1", Name: "On Submit", Enabled: false},
		{Id: "wf-3", ProjectId: "p1", Name: "On Click Hover", Enabled: true},
	}
	repo := &testutil.MockEventWorkflowRepository{
		GetEventWorkflowsWithFiltersFn: func(_ context.Context, projectID string, en *bool, name string) ([]models.EventWorkflow, error) {
			var out []models.EventWorkflow
			for _, wf := range all {
				if wf.ProjectId != projectID {
					continue
				}
				if en != nil && wf.Enabled != *en {
					continue
				}
				if name != "" {
					if len(wf.Name) < len(name) || wf.Name[:len(name)] != name {
						continue
					}
				}
				out = append(out, wf)
			}
			return out, nil
		},
	}
	got, err := repo.GetEventWorkflowsWithFilters(context.Background(), "p1", &enabled, "On Click")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 matching workflows, got %d", len(got))
	}
	for _, wf := range got {
		if !wf.Enabled {
			t.Errorf("workflow %q should be enabled", wf.Id)
		}
	}
}

// ─── MockElementEventWorkflowRepository (extended) ───────────────────────────

func TestMockElementEventWorkflowRepository_CreateDefaultReturnsInput(t *testing.T) {
	repo := &testutil.MockElementEventWorkflowRepository{}
	input := &models.ElementEventWorkflow{Id: "eew-1", ElementId: "el-1", WorkflowId: "wf-1", EventName: "onClick"}
	got, err := repo.CreateElementEventWorkflow(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Id != input.Id {
		t.Errorf("Id: got %q, want %q", got.Id, input.Id)
	}
}

func TestMockElementEventWorkflowRepository_CreateFnAssignsID(t *testing.T) {
	repo := &testutil.MockElementEventWorkflowRepository{
		CreateElementEventWorkflowFn: func(_ context.Context, eew *models.ElementEventWorkflow) (*models.ElementEventWorkflow, error) {
			eew.Id = "eew-generated"
			return eew, nil
		},
	}
	eew := &models.ElementEventWorkflow{ElementId: "el-1", WorkflowId: "wf-1", EventName: "onClick"}
	got, err := repo.CreateElementEventWorkflow(context.Background(), eew)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Id != "eew-generated" {
		t.Errorf("Id: got %q, want %q", got.Id, "eew-generated")
	}
}

func TestMockElementEventWorkflowRepository_GetByIDDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockElementEventWorkflowRepository{}
	got, err := repo.GetElementEventWorkflowByID(context.Background(), "eew-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

func TestMockElementEventWorkflowRepository_GetByIDFnReturnsItem(t *testing.T) {
	want := &models.ElementEventWorkflow{Id: "eew-1", ElementId: "el-1"}
	repo := &testutil.MockElementEventWorkflowRepository{
		GetElementEventWorkflowByIDFn: func(_ context.Context, id string) (*models.ElementEventWorkflow, error) {
			if id == "eew-1" {
				return want, nil
			}
			return nil, nil
		},
	}
	got, err := repo.GetElementEventWorkflowByID(context.Background(), "eew-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Id != want.Id {
		t.Errorf("Id: got %q, want %q", got.Id, want.Id)
	}
}

func TestMockElementEventWorkflowRepository_GetByElementIDDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockElementEventWorkflowRepository{}
	got, err := repo.GetElementEventWorkflowsByElementID(context.Background(), "el-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d", len(got))
	}
}

func TestMockElementEventWorkflowRepository_GetByElementIDFnFilters(t *testing.T) {
	all := []models.ElementEventWorkflow{
		{Id: "eew-1", ElementId: "el-1"},
		{Id: "eew-2", ElementId: "el-2"},
		{Id: "eew-3", ElementId: "el-1"},
	}
	repo := &testutil.MockElementEventWorkflowRepository{
		GetElementEventWorkflowsByElementIDFn: func(_ context.Context, elementID string) ([]models.ElementEventWorkflow, error) {
			var out []models.ElementEventWorkflow
			for _, e := range all {
				if e.ElementId == elementID {
					out = append(out, e)
				}
			}
			return out, nil
		},
	}
	got, err := repo.GetElementEventWorkflowsByElementID(context.Background(), "el-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 for el-1, got %d", len(got))
	}
}

func TestMockElementEventWorkflowRepository_GetByWorkflowIDDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockElementEventWorkflowRepository{}
	got, err := repo.GetElementEventWorkflowsByWorkflowID(context.Background(), "wf-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d", len(got))
	}
}

func TestMockElementEventWorkflowRepository_GetByWorkflowIDFnFilters(t *testing.T) {
	all := []models.ElementEventWorkflow{
		{Id: "eew-1", WorkflowId: "wf-1"},
		{Id: "eew-2", WorkflowId: "wf-2"},
		{Id: "eew-3", WorkflowId: "wf-1"},
	}
	repo := &testutil.MockElementEventWorkflowRepository{
		GetElementEventWorkflowsByWorkflowIDFn: func(_ context.Context, workflowID string) ([]models.ElementEventWorkflow, error) {
			var out []models.ElementEventWorkflow
			for _, e := range all {
				if e.WorkflowId == workflowID {
					out = append(out, e)
				}
			}
			return out, nil
		},
	}
	got, err := repo.GetElementEventWorkflowsByWorkflowID(context.Background(), "wf-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 for wf-1, got %d", len(got))
	}
}

func TestMockElementEventWorkflowRepository_GetByEventNameDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockElementEventWorkflowRepository{}
	got, err := repo.GetElementEventWorkflowsByEventName(context.Background(), "onClick")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d", len(got))
	}
}

func TestMockElementEventWorkflowRepository_GetByEventNameFnFilters(t *testing.T) {
	all := []models.ElementEventWorkflow{
		{Id: "eew-1", EventName: "onClick"},
		{Id: "eew-2", EventName: "onHover"},
		{Id: "eew-3", EventName: "onClick"},
	}
	repo := &testutil.MockElementEventWorkflowRepository{
		GetElementEventWorkflowsByEventNameFn: func(_ context.Context, eventName string) ([]models.ElementEventWorkflow, error) {
			var out []models.ElementEventWorkflow
			for _, e := range all {
				if e.EventName == eventName {
					out = append(out, e)
				}
			}
			return out, nil
		},
	}
	got, err := repo.GetElementEventWorkflowsByEventName(context.Background(), "onClick")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 onClick workflows, got %d", len(got))
	}
}

func TestMockElementEventWorkflowRepository_GetByFiltersDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockElementEventWorkflowRepository{}
	got, err := repo.GetElementEventWorkflowsByFilters(context.Background(), "el-1", "wf-1", "onClick")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d", len(got))
	}
}

func TestMockElementEventWorkflowRepository_GetByFiltersFnMatchesAll(t *testing.T) {
	all := []models.ElementEventWorkflow{
		{Id: "eew-1", ElementId: "el-1", WorkflowId: "wf-1", EventName: "onClick"},
		{Id: "eew-2", ElementId: "el-1", WorkflowId: "wf-2", EventName: "onClick"},
		{Id: "eew-3", ElementId: "el-2", WorkflowId: "wf-1", EventName: "onHover"},
	}
	repo := &testutil.MockElementEventWorkflowRepository{
		GetElementEventWorkflowsByFiltersFn: func(_ context.Context, elementID, workflowID, eventName string) ([]models.ElementEventWorkflow, error) {
			var out []models.ElementEventWorkflow
			for _, e := range all {
				if elementID != "" && e.ElementId != elementID {
					continue
				}
				if workflowID != "" && e.WorkflowId != workflowID {
					continue
				}
				if eventName != "" && e.EventName != eventName {
					continue
				}
				out = append(out, e)
			}
			return out, nil
		},
	}
	got, err := repo.GetElementEventWorkflowsByFilters(context.Background(), "el-1", "", "onClick")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 matching, got %d", len(got))
	}
}

func TestMockElementEventWorkflowRepository_UpdateDefaultReturnsInput(t *testing.T) {
	repo := &testutil.MockElementEventWorkflowRepository{}
	input := &models.ElementEventWorkflow{Id: "eew-1", EventName: "onHover"}
	got, err := repo.UpdateElementEventWorkflow(context.Background(), "eew-1", input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Id != input.Id {
		t.Errorf("Id: got %q, want %q", got.Id, input.Id)
	}
}

func TestMockElementEventWorkflowRepository_UpdateFnReturnsUpdated(t *testing.T) {
	want := &models.ElementEventWorkflow{Id: "eew-1", EventName: "onSubmit"}
	repo := &testutil.MockElementEventWorkflowRepository{
		UpdateElementEventWorkflowFn: func(_ context.Context, id string, _ *models.ElementEventWorkflow) (*models.ElementEventWorkflow, error) {
			if id == "eew-1" {
				return want, nil
			}
			return nil, errors.New("not found")
		},
	}
	got, err := repo.UpdateElementEventWorkflow(context.Background(), "eew-1", &models.ElementEventWorkflow{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.EventName != want.EventName {
		t.Errorf("EventName: got %q, want %q", got.EventName, want.EventName)
	}
}

func TestMockElementEventWorkflowRepository_DeleteDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockElementEventWorkflowRepository{}
	if err := repo.DeleteElementEventWorkflow(context.Background(), "eew-1"); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestMockElementEventWorkflowRepository_DeleteFnCalled(t *testing.T) {
	var capturedID string
	repo := &testutil.MockElementEventWorkflowRepository{
		DeleteElementEventWorkflowFn: func(_ context.Context, id string) error {
			capturedID = id
			return nil
		},
	}
	if err := repo.DeleteElementEventWorkflow(context.Background(), "eew-del"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "eew-del" {
		t.Errorf("captured ID: got %q, want %q", capturedID, "eew-del")
	}
}

func TestMockElementEventWorkflowRepository_DeleteByWorkflowIDDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockElementEventWorkflowRepository{}
	if err := repo.DeleteElementEventWorkflowsByWorkflowID(context.Background(), "wf-1"); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestMockElementEventWorkflowRepository_DeleteByWorkflowIDFnCalled(t *testing.T) {
	var capturedID string
	repo := &testutil.MockElementEventWorkflowRepository{
		DeleteElementEventWorkflowsByWorkflowIDFn: func(_ context.Context, workflowID string) error {
			capturedID = workflowID
			return nil
		},
	}
	if err := repo.DeleteElementEventWorkflowsByWorkflowID(context.Background(), "wf-purge"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "wf-purge" {
		t.Errorf("captured workflowID: got %q, want %q", capturedID, "wf-purge")
	}
}