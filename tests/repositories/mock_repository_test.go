package repositories_test

import (
	"context"
	"errors"
	"testing"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/tests/testutil"
)

// ─── Compile-time interface satisfaction ─────────────────────────────────────

var (
	_ repositories.ElementRepositoryInterface           = (*testutil.MockElementRepository)(nil)
	_ repositories.CustomElementRepositoryInterface     = (*testutil.MockCustomElementRepository)(nil)
	_ repositories.CustomElementTypeRepositoryInterface = (*testutil.MockCustomElementTypeRepository)(nil)
	_ repositories.ElementCommentRepositoryInterface    = (*testutil.MockElementCommentRepository)(nil)
)

// ─── MockElementRepository ────────────────────────────────────────────────────

func TestMockElementRepository_GetElementsDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockElementRepository{}
	got, err := repo.GetElements(context.Background(), "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d", len(got))
	}
}

func TestMockElementRepository_GetElementsFnFilters(t *testing.T) {
	el1 := &models.Element{Id: "el-1", Type: "Text"}
	el2 := &models.Element{Id: "el-2", Type: "Button"}
	repo := &testutil.MockElementRepository{
		GetElementsFn: func(_ context.Context, _ string, _ ...string) ([]models.EditorElement, error) {
			return []models.EditorElement{el1, el2}, nil
		},
	}

	got, err := repo.GetElements(context.Background(), "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 elements, got %d", len(got))
	}
}

func TestMockElementRepository_GetElementByIDDefaultsToNil(t *testing.T) {
	repo := &testutil.MockElementRepository{}
	got, err := repo.GetElementByID(context.Background(), "el-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

func TestMockElementRepository_GetElementByIDFnReturnsElement(t *testing.T) {
	want := &models.Element{Id: "el-1", Type: "Section"}
	repo := &testutil.MockElementRepository{
		GetElementByIDFn: func(_ context.Context, id string) (*models.Element, error) {
			if id == "el-1" {
				return want, nil
			}
			return nil, repositories.ErrElementNotFound
		},
	}

	got, err := repo.GetElementByID(context.Background(), "el-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Id != want.Id {
		t.Errorf("Id: got %q, want %q", got.Id, want.Id)
	}

	_, err = repo.GetElementByID(context.Background(), "el-missing")
	if !errors.Is(err, repositories.ErrElementNotFound) {
		t.Errorf("missing element: want ErrElementNotFound, got %v", err)
	}
}

func TestMockElementRepository_GetElementsByPageIDDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockElementRepository{}
	got, err := repo.GetElementsByPageID(context.Background(), "pg-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d", len(got))
	}
}

func TestMockElementRepository_GetElementsByPageIDFnFilters(t *testing.T) {
	all := []models.Element{
		{Id: "el-1", PageId: strPtr("pg-1")},
		{Id: "el-2", PageId: strPtr("pg-2")},
		{Id: "el-3", PageId: strPtr("pg-1")},
	}
	repo := &testutil.MockElementRepository{
		GetElementsByPageIDFn: func(_ context.Context, pageID string) ([]models.Element, error) {
			var out []models.Element
			for _, e := range all {
				if e.PageId != nil && *e.PageId == pageID {
					out = append(out, e)
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
}

func TestMockElementRepository_GetElementsByPageIdsDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockElementRepository{}
	got, err := repo.GetElementsByPageIds(context.Background(), []string{"pg-1", "pg-2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d", len(got))
	}
}

func TestMockElementRepository_GetChildElementsDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockElementRepository{}
	got, err := repo.GetChildElements(context.Background(), "el-parent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d", len(got))
	}
}

func TestMockElementRepository_GetChildElementsFnFilters(t *testing.T) {
	parentID := "el-parent"
	all := []models.Element{
		{Id: "el-c1", ParentId: &parentID},
		{Id: "el-c2", ParentId: &parentID},
	}
	repo := &testutil.MockElementRepository{
		GetChildElementsFn: func(_ context.Context, pid string) ([]models.Element, error) {
			if pid == parentID {
				return all, nil
			}
			return []models.Element{}, nil
		},
	}

	got, err := repo.GetChildElements(context.Background(), parentID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 children, got %d", len(got))
	}
}

func TestMockElementRepository_GetRootElementsDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockElementRepository{}
	got, err := repo.GetRootElements(context.Background(), "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d", len(got))
	}
}

func TestMockElementRepository_CreateElementFnIsInvoked(t *testing.T) {
	called := false
	repo := &testutil.MockElementRepository{
		CreateElementFn: func(_ context.Context, el *models.Element) error {
			called = true
			el.Id = "el-generated"
			return nil
		},
	}

	el := &models.Element{Type: "Text"}
	if err := repo.CreateElement(context.Background(), el); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("CreateElementFn was not called")
	}
	if el.Id != "el-generated" {
		t.Errorf("Id: got %q, want %q", el.Id, "el-generated")
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

	if err := repo.UpdateElement(context.Background(), &models.Element{Id: "el-42"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "el-42" {
		t.Errorf("captured ID: got %q, want %q", capturedID, "el-42")
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
		DeleteElementByIDFn: func(_ context.Context, id string) error {
			capturedID = id
			return nil
		},
	}

	if err := repo.DeleteElementByID(context.Background(), "el-99"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "el-99" {
		t.Errorf("captured ID: got %q, want %q", capturedID, "el-99")
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
		CountElementsByProjectIDFn: func(_ context.Context, _ string) (int64, error) {
			return 7, nil
		},
	}

	n, err := repo.CountElementsByProjectID(context.Background(), "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 7 {
		t.Errorf("expected 7, got %d", n)
	}
}

func TestMockElementRepository_GetElementWithRelationsDefaultsToNil(t *testing.T) {
	repo := &testutil.MockElementRepository{}
	got, err := repo.GetElementWithRelations(context.Background(), "el-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

func TestMockElementRepository_GetElementWithRelationsFnReturnsElement(t *testing.T) {
	want := &models.Element{Id: "el-1", Type: "Frame"}
	repo := &testutil.MockElementRepository{
		GetElementWithRelationsFn: func(_ context.Context, id string) (*models.Element, error) {
			if id == "el-1" {
				return want, nil
			}
			return nil, repositories.ErrElementNotFound
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
	got, err := repo.GetElementsByIDs(context.Background(), []string{"el-1", "el-2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d", len(got))
	}
}

func TestMockElementRepository_GetElementsByIDsFnFilters(t *testing.T) {
	all := []models.Element{
		{Id: "el-1", Type: "Text"},
		{Id: "el-2", Type: "Button"},
		{Id: "el-3", Type: "Image"},
	}
	repo := &testutil.MockElementRepository{
		GetElementsByIDsFn: func(_ context.Context, ids []string) ([]models.Element, error) {
			wanted := make(map[string]bool, len(ids))
			for _, id := range ids {
				wanted[id] = true
			}
			var out []models.Element
			for _, e := range all {
				if wanted[e.Id] {
					out = append(out, e)
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
		t.Errorf("expected 2 elements, got %d", len(got))
	}
}

func TestMockElementRepository_UpdateEventWorkflowsDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockElementRepository{}
	if err := repo.UpdateEventWorkflows(context.Background(), "el-1", []byte(`[]`)); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestMockElementRepository_ReplaceElementsDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockElementRepository{}
	if err := repo.ReplaceElements(context.Background(), "p1", []models.EditorElement{}); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

// ─── MockCustomElementRepository ─────────────────────────────────────────────

func TestMockCustomElementRepository_GetCustomElementsDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockCustomElementRepository{}
	got, err := repo.GetCustomElements(context.Background(), "u1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d", len(got))
	}
}

func TestMockCustomElementRepository_GetCustomElementsFnFilters(t *testing.T) {
	all := []models.CustomElement{
		{Id: "ce-1", UserId: "u1"},
		{Id: "ce-2", UserId: "u2"},
		{Id: "ce-3", UserId: "u1"},
	}
	repo := &testutil.MockCustomElementRepository{
		GetCustomElementsFn: func(_ context.Context, userID string, _ *bool) ([]models.CustomElement, error) {
			var out []models.CustomElement
			for _, ce := range all {
				if ce.UserId == userID {
					out = append(out, ce)
				}
			}
			return out, nil
		},
	}

	got, err := repo.GetCustomElements(context.Background(), "u1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 custom elements for u1, got %d", len(got))
	}
}

func TestMockCustomElementRepository_GetCustomElementByIDDefaultReturnsSentinel(t *testing.T) {
	repo := &testutil.MockCustomElementRepository{}
	_, err := repo.GetCustomElementByID(context.Background(), "ce-1", "u1")
	if !errors.Is(err, repositories.ErrCustomElementNotFound) {
		t.Errorf("want ErrCustomElementNotFound, got %v", err)
	}
}

func TestMockCustomElementRepository_GetCustomElementByIDFnReturnsElement(t *testing.T) {
	want := &models.CustomElement{Id: "ce-1", Name: "My Component", UserId: "u1"}
	repo := &testutil.MockCustomElementRepository{
		GetCustomElementByIDFn: func(_ context.Context, id, userID string) (*models.CustomElement, error) {
			if id == "ce-1" && userID == "u1" {
				return want, nil
			}
			return nil, repositories.ErrCustomElementNotFound
		},
	}

	got, err := repo.GetCustomElementByID(context.Background(), "ce-1", "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Id != want.Id {
		t.Errorf("Id: got %q, want %q", got.Id, want.Id)
	}

	_, err = repo.GetCustomElementByID(context.Background(), "ce-missing", "u1")
	if !errors.Is(err, repositories.ErrCustomElementNotFound) {
		t.Errorf("missing: want ErrCustomElementNotFound, got %v", err)
	}
}

func TestMockCustomElementRepository_CreateCustomElementDefaultReturnsInput(t *testing.T) {
	repo := &testutil.MockCustomElementRepository{}
	input := &models.CustomElement{Id: "ce-new", Name: "Widget", UserId: "u1"}
	got, err := repo.CreateCustomElement(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Id != input.Id {
		t.Errorf("Id: got %q, want %q", got.Id, input.Id)
	}
}

func TestMockCustomElementRepository_CreateCustomElementFnAssignsID(t *testing.T) {
	repo := &testutil.MockCustomElementRepository{
		CreateCustomElementFn: func(_ context.Context, ce *models.CustomElement) (*models.CustomElement, error) {
			ce.Id = "ce-generated"
			return ce, nil
		},
	}

	ce := &models.CustomElement{Name: "Widget", UserId: "u1"}
	got, err := repo.CreateCustomElement(context.Background(), ce)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Id != "ce-generated" {
		t.Errorf("Id: got %q, want %q", got.Id, "ce-generated")
	}
}

func TestMockCustomElementRepository_UpdateCustomElementDefaultReturnsSentinel(t *testing.T) {
	repo := &testutil.MockCustomElementRepository{}
	_, err := repo.UpdateCustomElement(context.Background(), "ce-1", "u1", map[string]any{"Name": "New"})
	if !errors.Is(err, repositories.ErrCustomElementNotFound) {
		t.Errorf("want ErrCustomElementNotFound, got %v", err)
	}
}

func TestMockCustomElementRepository_UpdateCustomElementFnReturnsUpdated(t *testing.T) {
	want := &models.CustomElement{Id: "ce-1", Name: "Updated Widget"}
	repo := &testutil.MockCustomElementRepository{
		UpdateCustomElementFn: func(_ context.Context, id, _ string, _ map[string]any) (*models.CustomElement, error) {
			if id == "ce-1" {
				return want, nil
			}
			return nil, repositories.ErrCustomElementNotFound
		},
	}

	got, err := repo.UpdateCustomElement(context.Background(), "ce-1", "u1", map[string]any{"Name": "Updated Widget"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != want.Name {
		t.Errorf("Name: got %q, want %q", got.Name, want.Name)
	}
}

func TestMockCustomElementRepository_DeleteCustomElementDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockCustomElementRepository{}
	if err := repo.DeleteCustomElement(context.Background(), "ce-1", "u1"); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestMockCustomElementRepository_DeleteCustomElementFnCalled(t *testing.T) {
	var capturedID string
	repo := &testutil.MockCustomElementRepository{
		DeleteCustomElementFn: func(_ context.Context, id, _ string) error {
			capturedID = id
			return nil
		},
	}

	if err := repo.DeleteCustomElement(context.Background(), "ce-77", "u1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "ce-77" {
		t.Errorf("captured ID: got %q, want %q", capturedID, "ce-77")
	}
}

func TestMockCustomElementRepository_GetPublicCustomElementsDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockCustomElementRepository{}
	got, err := repo.GetPublicCustomElements(context.Background(), nil, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d", len(got))
	}
}

func TestMockCustomElementRepository_GetPublicCustomElementsFnFiltersCategory(t *testing.T) {
	cat := "buttons"
	all := []models.CustomElement{
		{Id: "ce-1", Category: strPtr("buttons"), IsPublic: true},
		{Id: "ce-2", Category: strPtr("forms"), IsPublic: true},
		{Id: "ce-3", Category: strPtr("buttons"), IsPublic: true},
	}
	repo := &testutil.MockCustomElementRepository{
		GetPublicCustomElementsFn: func(_ context.Context, category *string, _, _ int) ([]models.CustomElement, error) {
			var out []models.CustomElement
			for _, ce := range all {
				if category == nil || (ce.Category != nil && *ce.Category == *category) {
					out = append(out, ce)
				}
			}
			return out, nil
		},
	}

	got, err := repo.GetPublicCustomElements(context.Background(), &cat, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 elements in 'buttons' category, got %d", len(got))
	}
}

func TestMockCustomElementRepository_DuplicateCustomElementDefaultReturnsSentinel(t *testing.T) {
	repo := &testutil.MockCustomElementRepository{}
	_, err := repo.DuplicateCustomElement(context.Background(), "ce-1", "u1", "Copy of Widget")
	if !errors.Is(err, repositories.ErrCustomElementNotFound) {
		t.Errorf("want ErrCustomElementNotFound, got %v", err)
	}
}

func TestMockCustomElementRepository_DuplicateCustomElementFnCreatesCopy(t *testing.T) {
	repo := &testutil.MockCustomElementRepository{
		DuplicateCustomElementFn: func(_ context.Context, id, userID, newName string) (*models.CustomElement, error) {
			if id == "ce-original" {
				return &models.CustomElement{Id: "ce-copy", Name: newName, UserId: userID}, nil
			}
			return nil, repositories.ErrCustomElementNotFound
		},
	}

	got, err := repo.DuplicateCustomElement(context.Background(), "ce-original", "u1", "Copy")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "Copy" {
		t.Errorf("Name: got %q, want %q", got.Name, "Copy")
	}
	if got.UserId != "u1" {
		t.Errorf("UserId: got %q, want %q", got.UserId, "u1")
	}
}

// ─── MockCustomElementTypeRepository ─────────────────────────────────────────

func TestMockCustomElementTypeRepository_GetCustomElementTypesDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockCustomElementTypeRepository{}
	got, err := repo.GetCustomElementTypes(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d", len(got))
	}
}

func TestMockCustomElementTypeRepository_GetCustomElementTypesFnReturnsAll(t *testing.T) {
	want := []models.CustomElementType{
		{Id: "cet-1", Name: "Layout"},
		{Id: "cet-2", Name: "Form"},
	}
	repo := &testutil.MockCustomElementTypeRepository{
		GetCustomElementTypesFn: func(_ context.Context) ([]models.CustomElementType, error) {
			return want, nil
		},
	}

	got, err := repo.GetCustomElementTypes(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != len(want) {
		t.Errorf("expected %d types, got %d", len(want), len(got))
	}
}

func TestMockCustomElementTypeRepository_GetCustomElementTypeByIDDefaultReturnsSentinel(t *testing.T) {
	repo := &testutil.MockCustomElementTypeRepository{}
	_, err := repo.GetCustomElementTypeByID(context.Background(), "cet-1")
	if !errors.Is(err, repositories.ErrCustomElementTypeNotFound) {
		t.Errorf("want ErrCustomElementTypeNotFound, got %v", err)
	}
}

func TestMockCustomElementTypeRepository_GetCustomElementTypeByIDFnReturnsType(t *testing.T) {
	want := &models.CustomElementType{Id: "cet-1", Name: "Layout"}
	repo := &testutil.MockCustomElementTypeRepository{
		GetCustomElementTypeByIDFn: func(_ context.Context, id string) (*models.CustomElementType, error) {
			if id == "cet-1" {
				return want, nil
			}
			return nil, repositories.ErrCustomElementTypeNotFound
		},
	}

	got, err := repo.GetCustomElementTypeByID(context.Background(), "cet-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != want.Name {
		t.Errorf("Name: got %q, want %q", got.Name, want.Name)
	}

	_, err = repo.GetCustomElementTypeByID(context.Background(), "cet-missing")
	if !errors.Is(err, repositories.ErrCustomElementTypeNotFound) {
		t.Errorf("missing: want ErrCustomElementTypeNotFound, got %v", err)
	}
}

func TestMockCustomElementTypeRepository_GetCustomElementTypeByNameDefaultReturnsSentinel(t *testing.T) {
	repo := &testutil.MockCustomElementTypeRepository{}
	_, err := repo.GetCustomElementTypeByName(context.Background(), "Layout")
	if !errors.Is(err, repositories.ErrCustomElementTypeNotFound) {
		t.Errorf("want ErrCustomElementTypeNotFound, got %v", err)
	}
}

func TestMockCustomElementTypeRepository_GetCustomElementTypeByNameFnLooksUpByName(t *testing.T) {
	store := map[string]*models.CustomElementType{
		"Layout": {Id: "cet-1", Name: "Layout"},
		"Form":   {Id: "cet-2", Name: "Form"},
	}
	repo := &testutil.MockCustomElementTypeRepository{
		GetCustomElementTypeByNameFn: func(_ context.Context, name string) (*models.CustomElementType, error) {
			if t, ok := store[name]; ok {
				return t, nil
			}
			return nil, repositories.ErrCustomElementTypeNotFound
		},
	}

	got, err := repo.GetCustomElementTypeByName(context.Background(), "Form")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Id != "cet-2" {
		t.Errorf("Id: got %q, want %q", got.Id, "cet-2")
	}
}

func TestMockCustomElementTypeRepository_CreateCustomElementTypeDefaultReturnsInput(t *testing.T) {
	repo := &testutil.MockCustomElementTypeRepository{}
	input := &models.CustomElementType{Id: "cet-new", Name: "Widget"}
	got, err := repo.CreateCustomElementType(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Id != input.Id {
		t.Errorf("Id: got %q, want %q", got.Id, input.Id)
	}
}

func TestMockCustomElementTypeRepository_CreateCustomElementTypeFnAssignsID(t *testing.T) {
	repo := &testutil.MockCustomElementTypeRepository{
		CreateCustomElementTypeFn: func(_ context.Context, cet *models.CustomElementType) (*models.CustomElementType, error) {
			cet.Id = "cet-generated"
			return cet, nil
		},
	}

	cet := &models.CustomElementType{Name: "Navigation"}
	got, err := repo.CreateCustomElementType(context.Background(), cet)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Id != "cet-generated" {
		t.Errorf("Id: got %q, want %q", got.Id, "cet-generated")
	}
}

func TestMockCustomElementTypeRepository_UpdateCustomElementTypeDefaultReturnsSentinel(t *testing.T) {
	repo := &testutil.MockCustomElementTypeRepository{}
	_, err := repo.UpdateCustomElementType(context.Background(), "cet-1", map[string]any{"Name": "Updated"})
	if !errors.Is(err, repositories.ErrCustomElementTypeNotFound) {
		t.Errorf("want ErrCustomElementTypeNotFound, got %v", err)
	}
}

func TestMockCustomElementTypeRepository_UpdateCustomElementTypeFnReturnsUpdated(t *testing.T) {
	want := &models.CustomElementType{Id: "cet-1", Name: "Updated Layout"}
	repo := &testutil.MockCustomElementTypeRepository{
		UpdateCustomElementTypeFn: func(_ context.Context, id string, _ map[string]any) (*models.CustomElementType, error) {
			if id == "cet-1" {
				return want, nil
			}
			return nil, repositories.ErrCustomElementTypeNotFound
		},
	}

	got, err := repo.UpdateCustomElementType(context.Background(), "cet-1", map[string]any{"Name": "Updated Layout"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != want.Name {
		t.Errorf("Name: got %q, want %q", got.Name, want.Name)
	}
}

func TestMockCustomElementTypeRepository_DeleteCustomElementTypeDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockCustomElementTypeRepository{}
	if err := repo.DeleteCustomElementType(context.Background(), "cet-1"); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestMockCustomElementTypeRepository_DeleteCustomElementTypeFnCalled(t *testing.T) {
	var capturedID string
	repo := &testutil.MockCustomElementTypeRepository{
		DeleteCustomElementTypeFn: func(_ context.Context, id string) error {
			capturedID = id
			return nil
		},
	}

	if err := repo.DeleteCustomElementType(context.Background(), "cet-55"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "cet-55" {
		t.Errorf("captured ID: got %q, want %q", capturedID, "cet-55")
	}
}

// ─── MockElementCommentRepository ────────────────────────────────────────────

func TestMockElementCommentRepository_CreateElementCommentDefaultReturnsInput(t *testing.T) {
	repo := &testutil.MockElementCommentRepository{}
	input := &models.ElementComment{Id: "ec-1", Content: "Nice!", AuthorId: "u1", ElementId: "el-1"}
	got, err := repo.CreateElementComment(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Id != input.Id {
		t.Errorf("Id: got %q, want %q", got.Id, input.Id)
	}
}

func TestMockElementCommentRepository_CreateElementCommentFnAssignsID(t *testing.T) {
	repo := &testutil.MockElementCommentRepository{
		CreateElementCommentFn: func(_ context.Context, c *models.ElementComment) (*models.ElementComment, error) {
			c.Id = "ec-generated"
			return c, nil
		},
	}

	c := &models.ElementComment{Content: "Hello", AuthorId: "u1", ElementId: "el-1"}
	got, err := repo.CreateElementComment(context.Background(), c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Id != "ec-generated" {
		t.Errorf("Id: got %q, want %q", got.Id, "ec-generated")
	}
}

func TestMockElementCommentRepository_GetElementCommentByIDDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockElementCommentRepository{}
	got, err := repo.GetElementCommentByID(context.Background(), "ec-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

func TestMockElementCommentRepository_GetElementCommentByIDFnReturnsComment(t *testing.T) {
	want := &models.ElementComment{Id: "ec-1", Content: "Nice work"}
	repo := &testutil.MockElementCommentRepository{
		GetElementCommentByIDFn: func(_ context.Context, id string) (*models.ElementComment, error) {
			if id == "ec-1" {
				return want, nil
			}
			return nil, nil
		},
	}

	got, err := repo.GetElementCommentByID(context.Background(), "ec-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Id != want.Id {
		t.Errorf("Id: got %q, want %q", got.Id, want.Id)
	}
}

func TestMockElementCommentRepository_GetElementCommentsDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockElementCommentRepository{}
	got, err := repo.GetElementComments(context.Background(), "el-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d", len(got))
	}
}

func TestMockElementCommentRepository_GetElementCommentsFnFilters(t *testing.T) {
	resolved := false
	all := []models.ElementComment{
		{Id: "ec-1", ElementId: "el-1", Resolved: false},
		{Id: "ec-2", ElementId: "el-1", Resolved: true},
		{Id: "ec-3", ElementId: "el-1", Resolved: false},
	}
	repo := &testutil.MockElementCommentRepository{
		GetElementCommentsFn: func(_ context.Context, _ string, filter *models.ElementCommentFilter) ([]models.ElementComment, error) {
			var out []models.ElementComment
			for _, c := range all {
				if filter == nil || filter.Resolved == nil || c.Resolved == *filter.Resolved {
					out = append(out, c)
				}
			}
			return out, nil
		},
	}

	got, err := repo.GetElementComments(context.Background(), "el-1", &models.ElementCommentFilter{Resolved: &resolved})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 unresolved comments, got %d", len(got))
	}
	for _, c := range got {
		if c.Resolved {
			t.Errorf("comment %q should not be resolved", c.Id)
		}
	}
}

func TestMockElementCommentRepository_UpdateElementCommentDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockElementCommentRepository{}
	got, err := repo.UpdateElementComment(context.Background(), "ec-1", map[string]any{"Content": "Updated"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

func TestMockElementCommentRepository_UpdateElementCommentFnReturnsUpdated(t *testing.T) {
	want := &models.ElementComment{Id: "ec-1", Content: "Updated content"}
	repo := &testutil.MockElementCommentRepository{
		UpdateElementCommentFn: func(_ context.Context, id string, _ map[string]any) (*models.ElementComment, error) {
			if id == "ec-1" {
				return want, nil
			}
			return nil, nil
		},
	}

	got, err := repo.UpdateElementComment(context.Background(), "ec-1", map[string]any{"Content": "Updated content"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Content != want.Content {
		t.Errorf("Content: got %q, want %q", got.Content, want.Content)
	}
}

func TestMockElementCommentRepository_DeleteElementCommentDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockElementCommentRepository{}
	if err := repo.DeleteElementComment(context.Background(), "ec-1"); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestMockElementCommentRepository_DeleteElementCommentFnCalled(t *testing.T) {
	var capturedID string
	repo := &testutil.MockElementCommentRepository{
		DeleteElementCommentFn: func(_ context.Context, id string) error {
			capturedID = id
			return nil
		},
	}

	if err := repo.DeleteElementComment(context.Background(), "ec-42"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "ec-42" {
		t.Errorf("captured ID: got %q, want %q", capturedID, "ec-42")
	}
}

func TestMockElementCommentRepository_GetElementCommentsByAuthorIDDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockElementCommentRepository{}
	got, err := repo.GetElementCommentsByAuthorID(context.Background(), "u1", 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d", len(got))
	}
}

func TestMockElementCommentRepository_GetElementCommentsByAuthorIDFnFilters(t *testing.T) {
	all := []models.ElementComment{
		{Id: "ec-1", AuthorId: "u1"},
		{Id: "ec-2", AuthorId: "u2"},
		{Id: "ec-3", AuthorId: "u1"},
	}
	repo := &testutil.MockElementCommentRepository{
		GetElementCommentsByAuthorIDFn: func(_ context.Context, authorID string, _, _ int) ([]models.ElementComment, error) {
			var out []models.ElementComment
			for _, c := range all {
				if c.AuthorId == authorID {
					out = append(out, c)
				}
			}
			return out, nil
		},
	}

	got, err := repo.GetElementCommentsByAuthorID(context.Background(), "u1", 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 comments by u1, got %d", len(got))
	}
}

func TestMockElementCommentRepository_CountElementCommentsDefaultsToZero(t *testing.T) {
	repo := &testutil.MockElementCommentRepository{}
	n, err := repo.CountElementComments(context.Background(), "el-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0, got %d", n)
	}
}

func TestMockElementCommentRepository_CountElementCommentsFnReturnsCount(t *testing.T) {
	repo := &testutil.MockElementCommentRepository{
		CountElementCommentsFn: func(_ context.Context, _ string) (int64, error) {
			return 5, nil
		},
	}

	n, err := repo.CountElementComments(context.Background(), "el-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 5 {
		t.Errorf("expected 5, got %d", n)
	}
}

func TestMockElementCommentRepository_ToggleResolvedStatusDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockElementCommentRepository{}
	got, err := repo.ToggleResolvedStatus(context.Background(), "ec-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

func TestMockElementCommentRepository_ToggleResolvedStatusFnToggles(t *testing.T) {
	comment := &models.ElementComment{Id: "ec-1", Resolved: false}
	repo := &testutil.MockElementCommentRepository{
		ToggleResolvedStatusFn: func(_ context.Context, id string) (*models.ElementComment, error) {
			if id == comment.Id {
				comment.Resolved = !comment.Resolved
				return comment, nil
			}
			return nil, nil
		},
	}

	got, err := repo.ToggleResolvedStatus(context.Background(), "ec-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Resolved {
		t.Error("expected comment to be resolved after toggle")
	}

	got, err = repo.ToggleResolvedStatus(context.Background(), "ec-1")
	if err != nil {
		t.Fatalf("unexpected error on second toggle: %v", err)
	}
	if got.Resolved {
		t.Error("expected comment to be unresolved after second toggle")
	}
}

func TestMockElementCommentRepository_DeleteElementCommentsByElementIDDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockElementCommentRepository{}
	if err := repo.DeleteElementCommentsByElementID(context.Background(), "el-1"); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestMockElementCommentRepository_DeleteElementCommentsByElementIDFnCalled(t *testing.T) {
	var capturedElementID string
	repo := &testutil.MockElementCommentRepository{
		DeleteElementCommentsByElementIDFn: func(_ context.Context, elementID string) error {
			capturedElementID = elementID
			return nil
		},
	}

	if err := repo.DeleteElementCommentsByElementID(context.Background(), "el-99"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedElementID != "el-99" {
		t.Errorf("captured elementID: got %q, want %q", capturedElementID, "el-99")
	}
}

func TestMockElementCommentRepository_GetElementCommentsByProjectIDDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockElementCommentRepository{}
	got, err := repo.GetElementCommentsByProjectID(context.Background(), "p1", 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d", len(got))
	}
}

func TestMockElementCommentRepository_GetElementCommentsByProjectIDFnFilters(t *testing.T) {
	all := []models.ElementComment{
		{Id: "ec-1", ElementId: "el-a"},
		{Id: "ec-2", ElementId: "el-b"},
	}
	repo := &testutil.MockElementCommentRepository{
		GetElementCommentsByProjectIDFn: func(_ context.Context, _ string, limit, _ int) ([]models.ElementComment, error) {
			if limit > 0 && limit < len(all) {
				return all[:limit], nil
			}
			return all, nil
		},
	}

	got, err := repo.GetElementCommentsByProjectID(context.Background(), "p1", 1, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("expected 1 comment with limit=1, got %d", len(got))
	}
}

func TestMockElementCommentRepository_CountElementCommentsByProjectIDDefaultsToZero(t *testing.T) {
	repo := &testutil.MockElementCommentRepository{}
	n, err := repo.CountElementCommentsByProjectID(context.Background(), "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0, got %d", n)
	}
}

func TestMockElementCommentRepository_CountElementCommentsByProjectIDFnReturnsCount(t *testing.T) {
	repo := &testutil.MockElementCommentRepository{
		CountElementCommentsByProjectIDFn: func(_ context.Context, _ string) (int64, error) {
			return 12, nil
		},
	}

	n, err := repo.CountElementCommentsByProjectID(context.Background(), "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 12 {
		t.Errorf("expected 12, got %d", n)
	}
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func strPtr(s string) *string { return &s }
