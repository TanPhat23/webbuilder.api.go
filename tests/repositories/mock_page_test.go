package repositories_test

import (
	"context"
	"errors"
	"testing"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/tests/testutil"
)

func TestMockPageRepository_DefaultsReturnSentinel(t *testing.T) {
	repo := &testutil.MockPageRepository{}
	_, err := repo.GetPageByID(context.Background(), "pg-1", "p1")
	if !errors.Is(err, repositories.ErrPageNotFound) {
		t.Errorf("GetPageByID default: want ErrPageNotFound, got %v", err)
	}
}

func TestMockPageRepository_GetPagesByProjectIDDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockPageRepository{}
	pages, err := repo.GetPagesByProjectID(context.Background(), "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pages) != 0 {
		t.Errorf("expected empty slice, got %d", len(pages))
	}
}

func TestMockPageRepository_GetPagesByProjectIDFnFilters(t *testing.T) {
	all := []models.Page{
		{Id: "pg-1", ProjectId: "p1"},
		{Id: "pg-2", ProjectId: "p2"},
		{Id: "pg-3", ProjectId: "p1"},
	}
	repo := &testutil.MockPageRepository{
		GetPagesByProjectIDFn: func(_ context.Context, projectID string) ([]models.Page, error) {
			var out []models.Page
			for _, p := range all {
				if p.ProjectId == projectID {
					out = append(out, p)
				}
			}
			return out, nil
		},
	}

	got, err := repo.GetPagesByProjectID(context.Background(), "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 pages for p1, got %d", len(got))
	}
	for _, p := range got {
		if p.ProjectId != "p1" {
			t.Errorf("unexpected ProjectId %q in result", p.ProjectId)
		}
	}
}

func TestMockPageRepository_GetPageByIDFnReturnsPage(t *testing.T) {
	want := &models.Page{Id: "pg-1", Name: "Home", ProjectId: "p1"}
	repo := &testutil.MockPageRepository{
		GetPageByIDFn: func(_ context.Context, pageID, projectID string) (*models.Page, error) {
			if pageID == "pg-1" && projectID == "p1" {
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
		t.Errorf("missing page: want ErrPageNotFound, got %v", err)
	}
}

func TestMockPageRepository_CreatePageDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockPageRepository{}
	if err := repo.CreatePage(context.Background(), &models.Page{Name: "Home", ProjectId: "p1"}); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestMockPageRepository_CreatePageFnIsInvoked(t *testing.T) {
	called := false
	repo := &testutil.MockPageRepository{
		CreatePageFn: func(_ context.Context, page *models.Page) error {
			called = true
			page.Id = "pg-created"
			return nil
		},
	}

	page := &models.Page{Name: "Home", ProjectId: "p1", Type: "page"}
	if err := repo.CreatePage(context.Background(), page); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("CreatePageFn was not called")
	}
	if page.Id != "pg-created" {
		t.Errorf("Id: got %q, want %q", page.Id, "pg-created")
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

	if err := repo.UpdatePage(context.Background(), &models.Page{Id: "pg-42", Name: "Updated"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "pg-42" {
		t.Errorf("page Id: got %q, want %q", capturedID, "pg-42")
	}
}

func TestMockPageRepository_UpdatePageFieldsDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockPageRepository{}
	if err := repo.UpdatePageFields(context.Background(), "pg-1", map[string]any{"Name": "New"}); err != nil {
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

	updates := map[string]any{"Name": "Renamed", "Order": 2}
	if err := repo.UpdatePageFields(context.Background(), "pg-1", updates); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedUpdates["Name"] != "Renamed" {
		t.Errorf("Name: got %v, want %q", capturedUpdates["Name"], "Renamed")
	}
	if capturedUpdates["Order"] != 2 {
		t.Errorf("Order: got %v, want 2", capturedUpdates["Order"])
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
		t.Errorf("pageID: got %q, want %q", capturedID, "pg-99")
	}
}

func TestMockPageRepository_DeletePageByProjectIDDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockPageRepository{}
	if err := repo.DeletePageByProjectID(context.Background(), "pg-1", "p1", "u1"); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestMockPageRepository_DeletePageByProjectIDFnCalled(t *testing.T) {
	var capturedPage, capturedProject, capturedUser string
	repo := &testutil.MockPageRepository{
		DeletePageByProjectIDFn: func(_ context.Context, pageID, projectID, userID string) error {
			capturedPage = pageID
			capturedProject = projectID
			capturedUser = userID
			return nil
		},
	}

	if err := repo.DeletePageByProjectID(context.Background(), "pg-1", "p1", "u1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedPage != "pg-1" {
		t.Errorf("pageID: got %q, want %q", capturedPage, "pg-1")
	}
	if capturedProject != "p1" {
		t.Errorf("projectID: got %q, want %q", capturedProject, "p1")
	}
	if capturedUser != "u1" {
		t.Errorf("userID: got %q, want %q", capturedUser, "u1")
	}
}