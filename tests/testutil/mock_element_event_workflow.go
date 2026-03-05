package testutil

import (
	"context"

	"my-go-app/internal/models"
)

type MockElementEventWorkflowRepository struct {
	CreateElementEventWorkflowFn              func(ctx context.Context, eew *models.ElementEventWorkflow) (*models.ElementEventWorkflow, error)
	GetElementEventWorkflowByIDFn             func(ctx context.Context, id string) (*models.ElementEventWorkflow, error)
	GetAllElementEventWorkflowsFn             func(ctx context.Context) ([]models.ElementEventWorkflow, error)
	GetElementEventWorkflowsByElementIDFn     func(ctx context.Context, elementID string) ([]models.ElementEventWorkflow, error)
	GetElementEventWorkflowsByWorkflowIDFn    func(ctx context.Context, workflowID string) ([]models.ElementEventWorkflow, error)
	GetElementEventWorkflowsByEventNameFn     func(ctx context.Context, eventName string) ([]models.ElementEventWorkflow, error)
	GetElementEventWorkflowsByFiltersFn       func(ctx context.Context, elementID, workflowID, eventName string) ([]models.ElementEventWorkflow, error)
	UpdateElementEventWorkflowFn              func(ctx context.Context, id string, eew *models.ElementEventWorkflow) (*models.ElementEventWorkflow, error)
	DeleteElementEventWorkflowFn              func(ctx context.Context, id string) error
	DeleteElementEventWorkflowsByElementIDFn  func(ctx context.Context, elementID string) error
	DeleteElementEventWorkflowsByWorkflowIDFn func(ctx context.Context, workflowID string) error
	GetElementEventWorkflowsByPageIDFn        func(ctx context.Context, pageID string) ([]models.ElementEventWorkflow, error)
	CheckIfWorkflowLinkedToElementFn          func(ctx context.Context, elementID, workflowID, eventName string) (bool, error)
}

func (m *MockElementEventWorkflowRepository) CreateElementEventWorkflow(ctx context.Context, eew *models.ElementEventWorkflow) (*models.ElementEventWorkflow, error) {
	if m.CreateElementEventWorkflowFn != nil {
		return m.CreateElementEventWorkflowFn(ctx, eew)
	}
	return eew, nil
}

func (m *MockElementEventWorkflowRepository) GetElementEventWorkflowByID(ctx context.Context, id string) (*models.ElementEventWorkflow, error) {
	if m.GetElementEventWorkflowByIDFn != nil {
		return m.GetElementEventWorkflowByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *MockElementEventWorkflowRepository) GetAllElementEventWorkflows(ctx context.Context) ([]models.ElementEventWorkflow, error) {
	if m.GetAllElementEventWorkflowsFn != nil {
		return m.GetAllElementEventWorkflowsFn(ctx)
	}
	return []models.ElementEventWorkflow{}, nil
}

func (m *MockElementEventWorkflowRepository) GetElementEventWorkflowsByElementID(ctx context.Context, elementID string) ([]models.ElementEventWorkflow, error) {
	if m.GetElementEventWorkflowsByElementIDFn != nil {
		return m.GetElementEventWorkflowsByElementIDFn(ctx, elementID)
	}
	return []models.ElementEventWorkflow{}, nil
}

func (m *MockElementEventWorkflowRepository) GetElementEventWorkflowsByWorkflowID(ctx context.Context, workflowID string) ([]models.ElementEventWorkflow, error) {
	if m.GetElementEventWorkflowsByWorkflowIDFn != nil {
		return m.GetElementEventWorkflowsByWorkflowIDFn(ctx, workflowID)
	}
	return []models.ElementEventWorkflow{}, nil
}

func (m *MockElementEventWorkflowRepository) GetElementEventWorkflowsByEventName(ctx context.Context, eventName string) ([]models.ElementEventWorkflow, error) {
	if m.GetElementEventWorkflowsByEventNameFn != nil {
		return m.GetElementEventWorkflowsByEventNameFn(ctx, eventName)
	}
	return []models.ElementEventWorkflow{}, nil
}

func (m *MockElementEventWorkflowRepository) GetElementEventWorkflowsByFilters(ctx context.Context, elementID, workflowID, eventName string) ([]models.ElementEventWorkflow, error) {
	if m.GetElementEventWorkflowsByFiltersFn != nil {
		return m.GetElementEventWorkflowsByFiltersFn(ctx, elementID, workflowID, eventName)
	}
	return []models.ElementEventWorkflow{}, nil
}

func (m *MockElementEventWorkflowRepository) UpdateElementEventWorkflow(ctx context.Context, id string, eew *models.ElementEventWorkflow) (*models.ElementEventWorkflow, error) {
	if m.UpdateElementEventWorkflowFn != nil {
		return m.UpdateElementEventWorkflowFn(ctx, id, eew)
	}
	return eew, nil
}

func (m *MockElementEventWorkflowRepository) DeleteElementEventWorkflow(ctx context.Context, id string) error {
	if m.DeleteElementEventWorkflowFn != nil {
		return m.DeleteElementEventWorkflowFn(ctx, id)
	}
	return nil
}

func (m *MockElementEventWorkflowRepository) DeleteElementEventWorkflowsByElementID(ctx context.Context, elementID string) error {
	if m.DeleteElementEventWorkflowsByElementIDFn != nil {
		return m.DeleteElementEventWorkflowsByElementIDFn(ctx, elementID)
	}
	return nil
}

func (m *MockElementEventWorkflowRepository) DeleteElementEventWorkflowsByWorkflowID(ctx context.Context, workflowID string) error {
	if m.DeleteElementEventWorkflowsByWorkflowIDFn != nil {
		return m.DeleteElementEventWorkflowsByWorkflowIDFn(ctx, workflowID)
	}
	return nil
}

func (m *MockElementEventWorkflowRepository) GetElementEventWorkflowsByPageID(ctx context.Context, pageID string) ([]models.ElementEventWorkflow, error) {
	if m.GetElementEventWorkflowsByPageIDFn != nil {
		return m.GetElementEventWorkflowsByPageIDFn(ctx, pageID)
	}
	return []models.ElementEventWorkflow{}, nil
}

func (m *MockElementEventWorkflowRepository) CheckIfWorkflowLinkedToElement(ctx context.Context, elementID, workflowID, eventName string) (bool, error) {
	if m.CheckIfWorkflowLinkedToElementFn != nil {
		return m.CheckIfWorkflowLinkedToElementFn(ctx, elementID, workflowID, eventName)
	}
	return false, nil
}