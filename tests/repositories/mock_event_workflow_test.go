package repositories_test

import (
	"context"
	"testing"

	"my-go-app/internal/models"
	"my-go-app/tests/testutil"
)

func TestMockEventWorkflowRepository_CountDefaultsToZero(t *testing.T) {
	repo := &testutil.MockEventWorkflowRepository{}
	n, err := repo.CountEventWorkflowsByProjectID(context.Background(), "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0, got %d", n)
	}
}

func TestMockEventWorkflowRepository_CountFnReturnsCount(t *testing.T) {
	repo := &testutil.MockEventWorkflowRepository{
		CountEventWorkflowsByProjectIDFn: func(_ context.Context, projectID string) (int64, error) {
			if projectID == "p1" {
				return 3, nil
			}
			return 0, nil
		},
	}
	n, err := repo.CountEventWorkflowsByProjectID(context.Background(), "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 3 {
		t.Errorf("expected 3, got %d", n)
	}
}

func TestMockEventWorkflowRepository_CheckNameExistsReturnsFalseByDefault(t *testing.T) {
	repo := &testutil.MockEventWorkflowRepository{}
	exists, err := repo.CheckIfWorkflowNameExists(context.Background(), "p1", "My Workflow", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected false by default")
	}
}

func TestMockEventWorkflowRepository_CheckNameExistsFnDetectsDuplicate(t *testing.T) {
	repo := &testutil.MockEventWorkflowRepository{
		CheckIfWorkflowNameExistsFn: func(_ context.Context, _, name, excludeID string) (bool, error) {
			return name == "Existing" && excludeID != "wf-1", nil
		},
	}

	exists, err := repo.CheckIfWorkflowNameExists(context.Background(), "p1", "Existing", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected true for duplicate name")
	}

	exists, err = repo.CheckIfWorkflowNameExists(context.Background(), "p1", "Existing", "wf-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected false when excluding own ID")
	}
}

func TestMockEventWorkflowRepository_CreateDefaultReturnsInput(t *testing.T) {
	repo := &testutil.MockEventWorkflowRepository{}
	wf := &models.EventWorkflow{Name: "On Click", ProjectId: "p1"}
	got, err := repo.CreateEventWorkflow(context.Background(), wf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != wf.Name {
		t.Errorf("Name: got %q, want %q", got.Name, wf.Name)
	}
}

func TestMockEventWorkflowRepository_CreateFnAssignsID(t *testing.T) {
	repo := &testutil.MockEventWorkflowRepository{
		CreateEventWorkflowFn: func(_ context.Context, wf *models.EventWorkflow) (*models.EventWorkflow, error) {
			wf.Id = "wf-generated"
			return wf, nil
		},
	}

	wf := &models.EventWorkflow{Name: "On Click", ProjectId: "p1"}
	got, err := repo.CreateEventWorkflow(context.Background(), wf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Id != "wf-generated" {
		t.Errorf("Id: got %q, want %q", got.Id, "wf-generated")
	}
}

func TestMockEventWorkflowRepository_GetByIDDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockEventWorkflowRepository{}
	got, err := repo.GetEventWorkflowByID(context.Background(), "wf-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil by default, got %+v", got)
	}
}

func TestMockEventWorkflowRepository_GetByIDFnReturnsWorkflow(t *testing.T) {
	want := &models.EventWorkflow{Id: "wf-1", Name: "On Hover", ProjectId: "p1"}
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

	none, err := repo.GetEventWorkflowByID(context.Background(), "wf-missing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if none != nil {
		t.Errorf("expected nil for missing ID, got %+v", none)
	}
}

func TestMockEventWorkflowRepository_GetByProjectIDDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockEventWorkflowRepository{}
	wfs, err := repo.GetEventWorkflowsByProjectID(context.Background(), "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(wfs) != 0 {
		t.Errorf("expected empty slice, got %d", len(wfs))
	}
}

func TestMockEventWorkflowRepository_GetByProjectIDFnFilters(t *testing.T) {
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

func TestMockEventWorkflowRepository_GetByProjectIDWithElementsDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockEventWorkflowRepository{}
	wfs, err := repo.GetEventWorkflowsByProjectIDWithElements(context.Background(), "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(wfs) != 0 {
		t.Errorf("expected empty slice, got %d", len(wfs))
	}
}

func TestMockEventWorkflowRepository_GetEnabledDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockEventWorkflowRepository{}
	wfs, err := repo.GetEnabledEventWorkflowsByProjectID(context.Background(), "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(wfs) != 0 {
		t.Errorf("expected empty slice, got %d", len(wfs))
	}
}

func TestMockEventWorkflowRepository_GetEnabledFnFilters(t *testing.T) {
	all := []models.EventWorkflow{
		{Id: "wf-1", ProjectId: "p1", Enabled: true},
		{Id: "wf-2", ProjectId: "p1", Enabled: false},
		{Id: "wf-3", ProjectId: "p1", Enabled: true},
	}
	repo := &testutil.MockEventWorkflowRepository{
		GetEnabledEventWorkflowsByProjectIDFn: func(_ context.Context, projectID string) ([]models.EventWorkflow, error) {
			var out []models.EventWorkflow
			for _, wf := range all {
				if wf.ProjectId == projectID && wf.Enabled {
					out = append(out, wf)
				}
			}
			return out, nil
		},
	}

	got, err := repo.GetEnabledEventWorkflowsByProjectID(context.Background(), "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 enabled workflows, got %d", len(got))
	}
	for _, wf := range got {
		if !wf.Enabled {
			t.Errorf("workflow %q should be enabled", wf.Id)
		}
	}
}

func TestMockEventWorkflowRepository_GetByNameDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockEventWorkflowRepository{}
	wfs, err := repo.GetEventWorkflowsByName(context.Background(), "p1", "On Click")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(wfs) != 0 {
		t.Errorf("expected empty slice, got %d", len(wfs))
	}
}

func TestMockEventWorkflowRepository_GetByNameFnFilters(t *testing.T) {
	all := []models.EventWorkflow{
		{Id: "wf-1", ProjectId: "p1", Name: "On Click"},
		{Id: "wf-2", ProjectId: "p1", Name: "On Hover"},
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

func TestMockEventWorkflowRepository_UpdateDefaultReturnsInput(t *testing.T) {
	repo := &testutil.MockEventWorkflowRepository{}
	wf := &models.EventWorkflow{Id: "wf-1", Name: "Updated"}
	got, err := repo.UpdateEventWorkflow(context.Background(), "wf-1", wf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != wf.Name {
		t.Errorf("Name: got %q, want %q", got.Name, wf.Name)
	}
}

func TestMockEventWorkflowRepository_UpdateFnReturnsUpdated(t *testing.T) {
	want := &models.EventWorkflow{Id: "wf-1", Name: "Renamed", Enabled: true}
	repo := &testutil.MockEventWorkflowRepository{
		UpdateEventWorkflowFn: func(_ context.Context, id string, wf *models.EventWorkflow) (*models.EventWorkflow, error) {
			if id == "wf-1" {
				return want, nil
			}
			return wf, nil
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

func TestMockEventWorkflowRepository_UpdateEnabledDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockEventWorkflowRepository{}
	if err := repo.UpdateEventWorkflowEnabled(context.Background(), "wf-1", true); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestMockEventWorkflowRepository_UpdateEnabledFnCalled(t *testing.T) {
	var capturedID string
	var capturedEnabled bool
	repo := &testutil.MockEventWorkflowRepository{
		UpdateEventWorkflowEnabledFn: func(_ context.Context, id string, enabled bool) error {
			capturedID = id
			capturedEnabled = enabled
			return nil
		},
	}

	if err := repo.UpdateEventWorkflowEnabled(context.Background(), "wf-1", false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "wf-1" {
		t.Errorf("id: got %q, want %q", capturedID, "wf-1")
	}
	if capturedEnabled {
		t.Error("expected enabled=false to be passed")
	}
}

func TestMockEventWorkflowRepository_DeleteDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockEventWorkflowRepository{}
	if err := repo.DeleteEventWorkflow(context.Background(), "wf-1"); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestMockEventWorkflowRepository_DeleteFnCalled(t *testing.T) {
	var capturedID string
	repo := &testutil.MockEventWorkflowRepository{
		DeleteEventWorkflowFn: func(_ context.Context, id string) error {
			capturedID = id
			return nil
		},
	}
	if err := repo.DeleteEventWorkflow(context.Background(), "wf-99"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "wf-99" {
		t.Errorf("id: got %q, want %q", capturedID, "wf-99")
	}
}

func TestMockEventWorkflowRepository_DeleteByProjectIDDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockEventWorkflowRepository{}
	if err := repo.DeleteEventWorkflowsByProjectID(context.Background(), "p1"); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestMockEventWorkflowRepository_DeleteByProjectIDFnCalled(t *testing.T) {
	var capturedProjectID string
	repo := &testutil.MockEventWorkflowRepository{
		DeleteEventWorkflowsByProjectIDFn: func(_ context.Context, projectID string) error {
			capturedProjectID = projectID
			return nil
		},
	}
	if err := repo.DeleteEventWorkflowsByProjectID(context.Background(), "p-42"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedProjectID != "p-42" {
		t.Errorf("projectID: got %q, want %q", capturedProjectID, "p-42")
	}
}

func TestMockEventWorkflowRepository_GetWithFiltersDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockEventWorkflowRepository{}
	wfs, err := repo.GetEventWorkflowsWithFilters(context.Background(), "p1", nil, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(wfs) != 0 {
		t.Errorf("expected empty slice, got %d", len(wfs))
	}
}

func TestMockEventWorkflowRepository_GetWithFiltersFnFiltersEnabledAndName(t *testing.T) {
	tru := true
	all := []models.EventWorkflow{
		{Id: "wf-1", ProjectId: "p1", Name: "Click Handler", Enabled: true},
		{Id: "wf-2", ProjectId: "p1", Name: "Hover Effect", Enabled: false},
		{Id: "wf-3", ProjectId: "p1", Name: "Click Logger", Enabled: true},
	}
	repo := &testutil.MockEventWorkflowRepository{
		GetEventWorkflowsWithFiltersFn: func(_ context.Context, projectID string, enabled *bool, searchName string) ([]models.EventWorkflow, error) {
			var out []models.EventWorkflow
			for _, wf := range all {
				if wf.ProjectId != projectID {
					continue
				}
				if enabled != nil && wf.Enabled != *enabled {
					continue
				}
				if searchName != "" && len(wf.Name) < len(searchName) {
					continue
				}
				if searchName != "" && wf.Name[:len(searchName)] != searchName {
					continue
				}
				out = append(out, wf)
			}
			return out, nil
		},
	}

	got, err := repo.GetEventWorkflowsWithFilters(context.Background(), "p1", &tru, "Click")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 enabled workflows starting with 'Click', got %d", len(got))
	}
	for _, wf := range got {
		if !wf.Enabled {
			t.Errorf("workflow %q should be enabled", wf.Id)
		}
	}
}