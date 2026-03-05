package repositories_test

import (
	"context"
	"testing"

	"my-go-app/internal/models"
	"my-go-app/tests/testutil"
)

func TestMockElementEventWorkflowRepository_GetAllDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockElementEventWorkflowRepository{}
	eews, err := repo.GetAllElementEventWorkflows(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(eews) != 0 {
		t.Errorf("expected empty slice, got %d", len(eews))
	}
}

func TestMockElementEventWorkflowRepository_GetAllFnReturnsAll(t *testing.T) {
	all := []models.ElementEventWorkflow{
		{Id: "eew-1", ElementId: "el-1", WorkflowId: "wf-1"},
		{Id: "eew-2", ElementId: "el-2", WorkflowId: "wf-1"},
	}
	repo := &testutil.MockElementEventWorkflowRepository{
		GetAllElementEventWorkflowsFn: func(_ context.Context) ([]models.ElementEventWorkflow, error) {
			return all, nil
		},
	}
	got, err := repo.GetAllElementEventWorkflows(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != len(all) {
		t.Errorf("expected %d, got %d", len(all), len(got))
	}
}

func TestMockElementEventWorkflowRepository_CreateDefaultReturnsInput(t *testing.T) {
	repo := &testutil.MockElementEventWorkflowRepository{}
	input := &models.ElementEventWorkflow{ElementId: "el-1", WorkflowId: "wf-1", EventName: "onClick"}
	got, err := repo.CreateElementEventWorkflow(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ElementId != input.ElementId {
		t.Errorf("ElementId: got %q, want %q", got.ElementId, input.ElementId)
	}
}

func TestMockElementEventWorkflowRepository_CreateFnAssignsID(t *testing.T) {
	repo := &testutil.MockElementEventWorkflowRepository{
		CreateElementEventWorkflowFn: func(_ context.Context, eew *models.ElementEventWorkflow) (*models.ElementEventWorkflow, error) {
			eew.Id = "eew-generated"
			return eew, nil
		},
	}
	input := &models.ElementEventWorkflow{ElementId: "el-1", WorkflowId: "wf-1", EventName: "onClick"}
	got, err := repo.CreateElementEventWorkflow(context.Background(), input)
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
		t.Errorf("expected nil by default, got %+v", got)
	}
}

func TestMockElementEventWorkflowRepository_GetByIDFnReturnsItem(t *testing.T) {
	want := &models.ElementEventWorkflow{Id: "eew-1", ElementId: "el-1", EventName: "onHover"}
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
	none, err := repo.GetElementEventWorkflowByID(context.Background(), "eew-missing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if none != nil {
		t.Errorf("expected nil for missing ID, got %+v", none)
	}
}

func TestMockElementEventWorkflowRepository_GetByElementIDDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockElementEventWorkflowRepository{}
	eews, err := repo.GetElementEventWorkflowsByElementID(context.Background(), "el-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(eews) != 0 {
		t.Errorf("expected empty slice, got %d", len(eews))
	}
}

func TestMockElementEventWorkflowRepository_GetByElementIDFnFilters(t *testing.T) {
	all := []models.ElementEventWorkflow{
		{Id: "eew-1", ElementId: "el-1", WorkflowId: "wf-1"},
		{Id: "eew-2", ElementId: "el-2", WorkflowId: "wf-1"},
		{Id: "eew-3", ElementId: "el-1", WorkflowId: "wf-2"},
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
	for _, e := range got {
		if e.ElementId != "el-1" {
			t.Errorf("unexpected ElementId %q", e.ElementId)
		}
	}
}

func TestMockElementEventWorkflowRepository_GetByWorkflowIDDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockElementEventWorkflowRepository{}
	eews, err := repo.GetElementEventWorkflowsByWorkflowID(context.Background(), "wf-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(eews) != 0 {
		t.Errorf("expected empty slice, got %d", len(eews))
	}
}

func TestMockElementEventWorkflowRepository_GetByWorkflowIDFnFilters(t *testing.T) {
	all := []models.ElementEventWorkflow{
		{Id: "eew-1", ElementId: "el-1", WorkflowId: "wf-1"},
		{Id: "eew-2", ElementId: "el-2", WorkflowId: "wf-2"},
		{Id: "eew-3", ElementId: "el-3", WorkflowId: "wf-1"},
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
	eews, err := repo.GetElementEventWorkflowsByEventName(context.Background(), "onClick")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(eews) != 0 {
		t.Errorf("expected empty slice, got %d", len(eews))
	}
}

func TestMockElementEventWorkflowRepository_GetByEventNameFnFilters(t *testing.T) {
	all := []models.ElementEventWorkflow{
		{Id: "eew-1", ElementId: "el-1", EventName: "onClick"},
		{Id: "eew-2", ElementId: "el-2", EventName: "onHover"},
		{Id: "eew-3", ElementId: "el-3", EventName: "onClick"},
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
		t.Errorf("expected 2 for onClick, got %d", len(got))
	}
}

func TestMockElementEventWorkflowRepository_GetByFiltersDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockElementEventWorkflowRepository{}
	eews, err := repo.GetElementEventWorkflowsByFilters(context.Background(), "el-1", "wf-1", "onClick")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(eews) != 0 {
		t.Errorf("expected empty slice, got %d", len(eews))
	}
}

func TestMockElementEventWorkflowRepository_GetByFiltersFnMatchesAll(t *testing.T) {
	all := []models.ElementEventWorkflow{
		{Id: "eew-1", ElementId: "el-1", WorkflowId: "wf-1", EventName: "onClick"},
		{Id: "eew-2", ElementId: "el-1", WorkflowId: "wf-2", EventName: "onClick"},
		{Id: "eew-3", ElementId: "el-2", WorkflowId: "wf-1", EventName: "onClick"},
		{Id: "eew-4", ElementId: "el-1", WorkflowId: "wf-1", EventName: "onHover"},
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
	got, err := repo.GetElementEventWorkflowsByFilters(context.Background(), "el-1", "wf-1", "onClick")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("expected 1 exact match, got %d", len(got))
	}
	if got[0].Id != "eew-1" {
		t.Errorf("Id: got %q, want %q", got[0].Id, "eew-1")
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
		UpdateElementEventWorkflowFn: func(_ context.Context, id string, eew *models.ElementEventWorkflow) (*models.ElementEventWorkflow, error) {
			if id == "eew-1" {
				return want, nil
			}
			return eew, nil
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
		t.Errorf("expected nil, got %v", err)
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
	if err := repo.DeleteElementEventWorkflow(context.Background(), "eew-99"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "eew-99" {
		t.Errorf("id: got %q, want %q", capturedID, "eew-99")
	}
}

func TestMockElementEventWorkflowRepository_DeleteByElementIDDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockElementEventWorkflowRepository{}
	if err := repo.DeleteElementEventWorkflowsByElementID(context.Background(), "el-1"); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestMockElementEventWorkflowRepository_DeleteByElementIDFnCalled(t *testing.T) {
	var capturedElementID string
	repo := &testutil.MockElementEventWorkflowRepository{
		DeleteElementEventWorkflowsByElementIDFn: func(_ context.Context, elementID string) error {
			capturedElementID = elementID
			return nil
		},
	}
	if err := repo.DeleteElementEventWorkflowsByElementID(context.Background(), "el-42"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedElementID != "el-42" {
		t.Errorf("elementID: got %q, want %q", capturedElementID, "el-42")
	}
}

func TestMockElementEventWorkflowRepository_DeleteByWorkflowIDDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockElementEventWorkflowRepository{}
	if err := repo.DeleteElementEventWorkflowsByWorkflowID(context.Background(), "wf-1"); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestMockElementEventWorkflowRepository_DeleteByWorkflowIDFnCalled(t *testing.T) {
	var capturedWorkflowID string
	repo := &testutil.MockElementEventWorkflowRepository{
		DeleteElementEventWorkflowsByWorkflowIDFn: func(_ context.Context, workflowID string) error {
			capturedWorkflowID = workflowID
			return nil
		},
	}
	if err := repo.DeleteElementEventWorkflowsByWorkflowID(context.Background(), "wf-99"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedWorkflowID != "wf-99" {
		t.Errorf("workflowID: got %q, want %q", capturedWorkflowID, "wf-99")
	}
}

func TestMockElementEventWorkflowRepository_GetByPageIDDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockElementEventWorkflowRepository{}
	eews, err := repo.GetElementEventWorkflowsByPageID(context.Background(), "pg-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(eews) != 0 {
		t.Errorf("expected empty slice, got %d", len(eews))
	}
}

func TestMockElementEventWorkflowRepository_GetByPageIDFnFilters(t *testing.T) {
	all := []models.ElementEventWorkflow{
		{Id: "eew-1", ElementId: "el-a"},
		{Id: "eew-2", ElementId: "el-b"},
	}
	repo := &testutil.MockElementEventWorkflowRepository{
		GetElementEventWorkflowsByPageIDFn: func(_ context.Context, _ string) ([]models.ElementEventWorkflow, error) {
			return all, nil
		},
	}
	got, err := repo.GetElementEventWorkflowsByPageID(context.Background(), "pg-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != len(all) {
		t.Errorf("expected %d, got %d", len(all), len(got))
	}
}

func TestMockElementEventWorkflowRepository_CheckLinkedReturnsFalseByDefault(t *testing.T) {
	repo := &testutil.MockElementEventWorkflowRepository{}
	linked, err := repo.CheckIfWorkflowLinkedToElement(context.Background(), "el-1", "wf-1", "onClick")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if linked {
		t.Error("expected false by default")
	}
}

func TestMockElementEventWorkflowRepository_CheckLinkedFnDetectsLink(t *testing.T) {
	linked := map[string]bool{"el-1:wf-1:onClick": true}
	repo := &testutil.MockElementEventWorkflowRepository{
		CheckIfWorkflowLinkedToElementFn: func(_ context.Context, elementID, workflowID, eventName string) (bool, error) {
			return linked[elementID+":"+workflowID+":"+eventName], nil
		},
	}

	got, err := repo.CheckIfWorkflowLinkedToElement(context.Background(), "el-1", "wf-1", "onClick")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got {
		t.Error("expected true for existing link")
	}

	got, err = repo.CheckIfWorkflowLinkedToElement(context.Background(), "el-1", "wf-1", "onHover")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got {
		t.Error("expected false for non-existing event link")
	}
}