package repositories_test

import (
	"context"
	"testing"

	"my-go-app/internal/models"
	"my-go-app/tests/testutil"
)

func TestMockElementRepository_GetElementsDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockElementRepository{}
	elements, err := repo.GetElements(context.Background(), "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(elements) != 0 {
		t.Errorf("expected empty slice, got %d", len(elements))
	}
}

func TestMockElementRepository_GetElementsFnFilters(t *testing.T) {
	all := []models.EditorElement{
		&models.Element{Id: "el-1", PageId: strPtr("pg-1")},
		&models.Element{Id: "el-2", PageId: strPtr("pg-2")},
		&models.Element{Id: "el-3", PageId: strPtr("pg-1")},
	}
	repo := &testutil.MockElementRepository{
		GetElementsFn: func(_ context.Context, projectID string, pageID ...string) ([]models.EditorElement, error) {
			if len(pageID) == 0 {
				return all, nil
			}
			var out []models.EditorElement
			for _, el := range all {
				base := el.GetElement()
				if base.PageId != nil && *base.PageId == pageID[0] {
					out = append(out, el)
				}
			}
			return out, nil
		},
	}

	got, err := repo.GetElements(context.Background(), "p1", "pg-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 elements for pg-1, got %d", len(got))
	}
}

func TestMockElementRepository_GetElementByIDDefaultsToNil(t *testing.T) {
	repo := &testutil.MockElementRepository{}
	got, err := repo.GetElementByID(context.Background(), "el-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil by default, got %+v", got)
	}
}

func TestMockElementRepository_GetElementByIDFnReturnsElement(t *testing.T) {
	want := &models.Element{Id: "el-1", Type: "Text"}
	repo := &testutil.MockElementRepository{
		GetElementByIDFn: func(_ context.Context, elementID string) (*models.Element, error) {
			if elementID == "el-1" {
				return want, nil
			}
			return nil, nil
		},
	}

	got, err := repo.GetElementByID(context.Background(), "el-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Id != want.Id {
		t.Errorf("Id: got %q, want %q", got.Id, want.Id)
	}

	none, err := repo.GetElementByID(context.Background(), "el-missing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if none != nil {
		t.Errorf("expected nil for missing ID, got %+v", none)
	}
}

func TestMockElementRepository_GetElementsByPageIDDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockElementRepository{}
	elements, err := repo.GetElementsByPageID(context.Background(), "pg-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(elements) != 0 {
		t.Errorf("expected empty slice, got %d", len(elements))
	}
}

func TestMockElementRepository_GetElementsByPageIDFnFilters(t *testing.T) {
	all := []models.Element{
		{Id: "el-1", Type: "Text", PageId: strPtr("pg-1")},
		{Id: "el-2", Type: "Button", PageId: strPtr("pg-2")},
		{Id: "el-3", Type: "Image", PageId: strPtr("pg-1")},
	}
	repo := &testutil.MockElementRepository{
		GetElementsByPageIDFn: func(_ context.Context, pageID string) ([]models.Element, error) {
			var out []models.Element
			for _, el := range all {
				if el.PageId != nil && *el.PageId == pageID {
					out = append(out, el)
				}
			}
			return out, nil
		},
	}

	got, err := repo.GetElementsByPageID(context.Background(), "pg-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 elements for pg-1, got %d", len(got))
	}
	for _, el := range got {
		if el.PageId == nil || *el.PageId != "pg-1" {
			t.Errorf("unexpected PageId in result")
		}
	}
}

func TestMockElementRepository_GetElementsByPageIdsDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockElementRepository{}
	elements, err := repo.GetElementsByPageIds(context.Background(), []string{"pg-1", "pg-2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(elements) != 0 {
		t.Errorf("expected empty slice, got %d", len(elements))
	}
}

func TestMockElementRepository_GetChildElementsDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockElementRepository{}
	children, err := repo.GetChildElements(context.Background(), "el-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(children) != 0 {
		t.Errorf("expected empty slice, got %d", len(children))
	}
}

func TestMockElementRepository_GetChildElementsFnFilters(t *testing.T) {
	parent := strPtr("el-parent")
	all := []models.Element{
		{Id: "el-1", Type: "Text", ParentId: parent},
		{Id: "el-2", Type: "Button", ParentId: strPtr("el-other")},
		{Id: "el-3", Type: "Image", ParentId: parent},
	}
	repo := &testutil.MockElementRepository{
		GetChildElementsFn: func(_ context.Context, parentID string) ([]models.Element, error) {
			var out []models.Element
			for _, el := range all {
				if el.ParentId != nil && *el.ParentId == parentID {
					out = append(out, el)
				}
			}
			return out, nil
		},
	}

	got, err := repo.GetChildElements(context.Background(), "el-parent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 children, got %d", len(got))
	}
}

func TestMockElementRepository_GetRootElementsDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockElementRepository{}
	roots, err := repo.GetRootElements(context.Background(), "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(roots) != 0 {
		t.Errorf("expected empty slice, got %d", len(roots))
	}
}

func TestMockElementRepository_CreateElementDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockElementRepository{}
	if err := repo.CreateElement(context.Background(), &models.Element{Type: "Text"}); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestMockElementRepository_CreateElementFnIsInvoked(t *testing.T) {
	called := false
	repo := &testutil.MockElementRepository{
		CreateElementFn: func(_ context.Context, el *models.Element) error {
			called = true
			el.Id = "el-created"
			return nil
		},
	}

	el := &models.Element{Type: "Button"}
	if err := repo.CreateElement(context.Background(), el); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("CreateElementFn was not called")
	}
	if el.Id != "el-created" {
		t.Errorf("Id: got %q, want %q", el.Id, "el-created")
	}
}

func TestMockElementRepository_UpdateElementDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockElementRepository{}
	if err := repo.UpdateElement(context.Background(), &models.Element{Id: "el-1"}); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestMockElementRepository_UpdateElementFnCalled(t *testing.T) {
	var capturedID string
	repo := &testutil.MockElementRepository{
		UpdateElementFn: func(_ context.Context, el *models.Element) error {
			capturedID = el.Id
			return nil
		},
	}

	if err := repo.UpdateElement(context.Background(), &models.Element{Id: "el-42", Type: "Text"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "el-42" {
		t.Errorf("id: got %q, want %q", capturedID, "el-42")
	}
}

func TestMockElementRepository_DeleteElementByIDDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockElementRepository{}
	if err := repo.DeleteElementByID(context.Background(), "el-1"); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestMockElementRepository_DeleteElementByIDFnCalled(t *testing.T) {
	var capturedID string
	repo := &testutil.MockElementRepository{
		DeleteElementByIDFn: func(_ context.Context, elementID string) error {
			capturedID = elementID
			return nil
		},
	}

	if err := repo.DeleteElementByID(context.Background(), "el-99"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "el-99" {
		t.Errorf("elementID: got %q, want %q", capturedID, "el-99")
	}
}

func TestMockElementRepository_DeleteElementsByPageIDDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockElementRepository{}
	if err := repo.DeleteElementsByPageID(context.Background(), "pg-1"); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestMockElementRepository_DeleteElementsByProjectIDDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockElementRepository{}
	if err := repo.DeleteElementsByProjectID(context.Background(), "p1"); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestMockElementRepository_CountElementsByProjectIDDefaultsToZero(t *testing.T) {
	repo := &testutil.MockElementRepository{}
	n, err := repo.CountElementsByProjectID(context.Background(), "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0, got %d", n)
	}
}

func TestMockElementRepository_CountElementsByProjectIDFnReturnsCount(t *testing.T) {
	repo := &testutil.MockElementRepository{
		CountElementsByProjectIDFn: func(_ context.Context, projectID string) (int64, error) {
			if projectID == "p1" {
				return 5, nil
			}
			return 0, nil
		},
	}

	n, err := repo.CountElementsByProjectID(context.Background(), "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 5 {
		t.Errorf("expected 5, got %d", n)
	}

	n, err = repo.CountElementsByProjectID(context.Background(), "p-other")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 for unknown project, got %d", n)
	}
}

func TestMockElementRepository_GetElementWithRelationsDefaultsToNil(t *testing.T) {
	repo := &testutil.MockElementRepository{}
	got, err := repo.GetElementWithRelations(context.Background(), "el-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil by default, got %+v", got)
	}
}

func TestMockElementRepository_GetElementWithRelationsFnReturnsElement(t *testing.T) {
	want := &models.Element{Id: "el-1", Type: "Section"}
	repo := &testutil.MockElementRepository{
		GetElementWithRelationsFn: func(_ context.Context, elementID string) (*models.Element, error) {
			if elementID == "el-1" {
				return want, nil
			}
			return nil, nil
		},
	}

	got, err := repo.GetElementWithRelations(context.Background(), "el-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Id != want.Id {
		t.Errorf("Id: got %q, want %q", got.Id, want.Id)
	}
}

func TestMockElementRepository_GetElementsByIDsDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockElementRepository{}
	elements, err := repo.GetElementsByIDs(context.Background(), []string{"el-1", "el-2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(elements) != 0 {
		t.Errorf("expected empty slice, got %d", len(elements))
	}
}

func TestMockElementRepository_GetElementsByIDsFnFilters(t *testing.T) {
	store := map[string]models.Element{
		"el-1": {Id: "el-1", Type: "Text"},
		"el-2": {Id: "el-2", Type: "Button"},
		"el-3": {Id: "el-3", Type: "Image"},
	}
	repo := &testutil.MockElementRepository{
		GetElementsByIDsFn: func(_ context.Context, ids []string) ([]models.Element, error) {
			var out []models.Element
			for _, id := range ids {
				if el, ok := store[id]; ok {
					out = append(out, el)
				}
			}
			return out, nil
		},
	}

	got, err := repo.GetElementsByIDs(context.Background(), []string{"el-1", "el-3"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2, got %d", len(got))
	}
}

func TestMockElementRepository_UpdateEventWorkflowsDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockElementRepository{}
	if err := repo.UpdateEventWorkflows(context.Background(), "el-1", []byte(`[]`)); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestMockElementRepository_UpdateEventWorkflowsFnCalled(t *testing.T) {
	var capturedID string
	var capturedPayload []byte
	repo := &testutil.MockElementRepository{
		UpdateEventWorkflowsFn: func(_ context.Context, elementID string, workflows []byte) error {
			capturedID = elementID
			capturedPayload = workflows
			return nil
		},
	}

	payload := []byte(`[{"id":"wf-1"}]`)
	if err := repo.UpdateEventWorkflows(context.Background(), "el-42", payload); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "el-42" {
		t.Errorf("elementID: got %q, want %q", capturedID, "el-42")
	}
	if string(capturedPayload) != string(payload) {
		t.Errorf("payload: got %q, want %q", capturedPayload, payload)
	}
}

func TestMockElementRepository_ReplaceElementsDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockElementRepository{}
	if err := repo.ReplaceElements(context.Background(), "p1", []models.EditorElement{}); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestMockElementRepository_ReplaceElementsFnCalled(t *testing.T) {
	var capturedProjectID string
	var capturedCount int
	repo := &testutil.MockElementRepository{
		ReplaceElementsFn: func(_ context.Context, projectID string, elements []models.EditorElement) error {
			capturedProjectID = projectID
			capturedCount = len(elements)
			return nil
		},
	}

	elements := []models.EditorElement{
		&models.Element{Id: "el-1", Type: "Text"},
		&models.Element{Id: "el-2", Type: "Button"},
	}
	if err := repo.ReplaceElements(context.Background(), "p1", elements); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedProjectID != "p1" {
		t.Errorf("projectID: got %q, want %q", capturedProjectID, "p1")
	}
	if capturedCount != 2 {
		t.Errorf("element count: got %d, want 2", capturedCount)
	}
}

func strPtr(s string) *string {
	return &s
}