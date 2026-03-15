package testutil

import (
	"context"

	"my-go-app/internal/models"
)

type MockElementRepository struct {
	GetElementsFn               func(ctx context.Context, projectID string, pageID ...string) ([]models.EditorElement, error)
	ReplaceElementsFn           func(ctx context.Context, projectID string, elements []models.EditorElement) error
	GetElementByIDFn            func(ctx context.Context, elementID string) (*models.Element, error)
	GetElementsByPageIDFn       func(ctx context.Context, pageID string) ([]models.Element, error)
	GetElementsByPageIdsFn      func(ctx context.Context, pageIDs []string) ([]models.EditorElement, error)
	GetChildElementsFn          func(ctx context.Context, parentID string) ([]models.Element, error)
	GetRootElementsFn           func(ctx context.Context, projectID string) ([]models.Element, error)
	CreateElementFn             func(ctx context.Context, element *models.Element) error
	UpdateElementFn             func(ctx context.Context, element *models.Element) error
	UpdateEventWorkflowsFn      func(ctx context.Context, elementID string, workflows []byte) error
	DeleteElementByIDFn         func(ctx context.Context, elementID string) error
	DeleteElementsByPageIDFn    func(ctx context.Context, pageID string) error
	DeleteElementsByProjectIDFn func(ctx context.Context, projectID string) error
	CountElementsByProjectIDFn  func(ctx context.Context, projectID string) (int64, error)
	GetElementWithRelationsFn   func(ctx context.Context, elementID string) (*models.Element, error)
	GetElementsByIDsFn          func(ctx context.Context, elementIDs []string) ([]models.Element, error)
}

func (m *MockElementRepository) GetElements(ctx context.Context, projectID string, pageID ...string) ([]models.EditorElement, error) {
	if m.GetElementsFn != nil {
		return m.GetElementsFn(ctx, projectID, pageID...)
	}
	return []models.EditorElement{}, nil
}

func (m *MockElementRepository) ReplaceElements(ctx context.Context, projectID string, elements []models.EditorElement) error {
	if m.ReplaceElementsFn != nil {
		return m.ReplaceElementsFn(ctx, projectID, elements)
	}
	return nil
}

func (m *MockElementRepository) GetElementByID(ctx context.Context, elementID string) (*models.Element, error) {
	if m.GetElementByIDFn != nil {
		return m.GetElementByIDFn(ctx, elementID)
	}
	return nil, nil
}

func (m *MockElementRepository) GetElementsByPageID(ctx context.Context, pageID string) ([]models.Element, error) {
	if m.GetElementsByPageIDFn != nil {
		return m.GetElementsByPageIDFn(ctx, pageID)
	}
	return []models.Element{}, nil
}

func (m *MockElementRepository) GetElementsByPageIds(ctx context.Context, pageIDs []string) ([]models.EditorElement, error) {
	if m.GetElementsByPageIdsFn != nil {
		return m.GetElementsByPageIdsFn(ctx, pageIDs)
	}
	return []models.EditorElement{}, nil
}

func (m *MockElementRepository) GetChildElements(ctx context.Context, parentID string) ([]models.Element, error) {
	if m.GetChildElementsFn != nil {
		return m.GetChildElementsFn(ctx, parentID)
	}
	return []models.Element{}, nil
}

func (m *MockElementRepository) GetRootElements(ctx context.Context, projectID string) ([]models.Element, error) {
	if m.GetRootElementsFn != nil {
		return m.GetRootElementsFn(ctx, projectID)
	}
	return []models.Element{}, nil
}

func (m *MockElementRepository) CreateElement(ctx context.Context, element *models.Element) error {
	if m.CreateElementFn != nil {
		return m.CreateElementFn(ctx, element)
	}
	return nil
}

func (m *MockElementRepository) UpdateElement(ctx context.Context, element *models.Element) error {
	if m.UpdateElementFn != nil {
		return m.UpdateElementFn(ctx, element)
	}
	return nil
}

func (m *MockElementRepository) UpdateEventWorkflows(ctx context.Context, elementID string, workflows []byte) error {
	if m.UpdateEventWorkflowsFn != nil {
		return m.UpdateEventWorkflowsFn(ctx, elementID, workflows)
	}
	return nil
}

func (m *MockElementRepository) DeleteElementByID(ctx context.Context, elementID string) error {
	if m.DeleteElementByIDFn != nil {
		return m.DeleteElementByIDFn(ctx, elementID)
	}
	return nil
}

func (m *MockElementRepository) DeleteElementsByPageID(ctx context.Context, pageID string) error {
	if m.DeleteElementsByPageIDFn != nil {
		return m.DeleteElementsByPageIDFn(ctx, pageID)
	}
	return nil
}

func (m *MockElementRepository) DeleteElementsByProjectID(ctx context.Context, projectID string) error {
	if m.DeleteElementsByProjectIDFn != nil {
		return m.DeleteElementsByProjectIDFn(ctx, projectID)
	}
	return nil
}

func (m *MockElementRepository) CountElementsByProjectID(ctx context.Context, projectID string) (int64, error) {
	if m.CountElementsByProjectIDFn != nil {
		return m.CountElementsByProjectIDFn(ctx, projectID)
	}
	return 0, nil
}

func (m *MockElementRepository) GetElementWithRelations(ctx context.Context, elementID string) (*models.Element, error) {
	if m.GetElementWithRelationsFn != nil {
		return m.GetElementWithRelationsFn(ctx, elementID)
	}
	return nil, nil
}

func (m *MockElementRepository) GetElementsByIDs(ctx context.Context, elementIDs []string) ([]models.Element, error) {
	if m.GetElementsByIDsFn != nil {
		return m.GetElementsByIDsFn(ctx, elementIDs)
	}
	return []models.Element{}, nil
}