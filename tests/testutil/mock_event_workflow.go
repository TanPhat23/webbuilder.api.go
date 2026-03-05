package testutil

import (
	"context"

	"my-go-app/internal/models"
)

type MockEventWorkflowRepository struct {
	CreateEventWorkflowFn                      func(ctx context.Context, workflow *models.EventWorkflow) (*models.EventWorkflow, error)
	GetEventWorkflowByIDFn                     func(ctx context.Context, id string) (*models.EventWorkflow, error)
	GetEventWorkflowsByProjectIDFn             func(ctx context.Context, projectID string) ([]models.EventWorkflow, error)
	GetEventWorkflowsByProjectIDWithElementsFn func(ctx context.Context, projectID string) ([]models.EventWorkflow, error)
	GetEnabledEventWorkflowsByProjectIDFn      func(ctx context.Context, projectID string) ([]models.EventWorkflow, error)
	GetEventWorkflowsByNameFn                  func(ctx context.Context, projectID, name string) ([]models.EventWorkflow, error)
	UpdateEventWorkflowFn                      func(ctx context.Context, id string, workflow *models.EventWorkflow) (*models.EventWorkflow, error)
	UpdateEventWorkflowEnabledFn               func(ctx context.Context, id string, enabled bool) error
	DeleteEventWorkflowFn                      func(ctx context.Context, id string) error
	DeleteEventWorkflowsByProjectIDFn          func(ctx context.Context, projectID string) error
	CountEventWorkflowsByProjectIDFn           func(ctx context.Context, projectID string) (int64, error)
	CheckIfWorkflowNameExistsFn                func(ctx context.Context, projectID, name string, excludeID string) (bool, error)
	GetEventWorkflowsWithFiltersFn             func(ctx context.Context, projectID string, enabled *bool, searchName string) ([]models.EventWorkflow, error)
}

func (m *MockEventWorkflowRepository) CreateEventWorkflow(ctx context.Context, workflow *models.EventWorkflow) (*models.EventWorkflow, error) {
	if m.CreateEventWorkflowFn != nil {
		return m.CreateEventWorkflowFn(ctx, workflow)
	}
	return workflow, nil
}

func (m *MockEventWorkflowRepository) GetEventWorkflowByID(ctx context.Context, id string) (*models.EventWorkflow, error) {
	if m.GetEventWorkflowByIDFn != nil {
		return m.GetEventWorkflowByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *MockEventWorkflowRepository) GetEventWorkflowsByProjectID(ctx context.Context, projectID string) ([]models.EventWorkflow, error) {
	if m.GetEventWorkflowsByProjectIDFn != nil {
		return m.GetEventWorkflowsByProjectIDFn(ctx, projectID)
	}
	return []models.EventWorkflow{}, nil
}

func (m *MockEventWorkflowRepository) GetEventWorkflowsByProjectIDWithElements(ctx context.Context, projectID string) ([]models.EventWorkflow, error) {
	if m.GetEventWorkflowsByProjectIDWithElementsFn != nil {
		return m.GetEventWorkflowsByProjectIDWithElementsFn(ctx, projectID)
	}
	return []models.EventWorkflow{}, nil
}

func (m *MockEventWorkflowRepository) GetEnabledEventWorkflowsByProjectID(ctx context.Context, projectID string) ([]models.EventWorkflow, error) {
	if m.GetEnabledEventWorkflowsByProjectIDFn != nil {
		return m.GetEnabledEventWorkflowsByProjectIDFn(ctx, projectID)
	}
	return []models.EventWorkflow{}, nil
}

func (m *MockEventWorkflowRepository) GetEventWorkflowsByName(ctx context.Context, projectID, name string) ([]models.EventWorkflow, error) {
	if m.GetEventWorkflowsByNameFn != nil {
		return m.GetEventWorkflowsByNameFn(ctx, projectID, name)
	}
	return []models.EventWorkflow{}, nil
}

func (m *MockEventWorkflowRepository) UpdateEventWorkflow(ctx context.Context, id string, workflow *models.EventWorkflow) (*models.EventWorkflow, error) {
	if m.UpdateEventWorkflowFn != nil {
		return m.UpdateEventWorkflowFn(ctx, id, workflow)
	}
	return workflow, nil
}

func (m *MockEventWorkflowRepository) UpdateEventWorkflowEnabled(ctx context.Context, id string, enabled bool) error {
	if m.UpdateEventWorkflowEnabledFn != nil {
		return m.UpdateEventWorkflowEnabledFn(ctx, id, enabled)
	}
	return nil
}

func (m *MockEventWorkflowRepository) DeleteEventWorkflow(ctx context.Context, id string) error {
	if m.DeleteEventWorkflowFn != nil {
		return m.DeleteEventWorkflowFn(ctx, id)
	}
	return nil
}

func (m *MockEventWorkflowRepository) DeleteEventWorkflowsByProjectID(ctx context.Context, projectID string) error {
	if m.DeleteEventWorkflowsByProjectIDFn != nil {
		return m.DeleteEventWorkflowsByProjectIDFn(ctx, projectID)
	}
	return nil
}

func (m *MockEventWorkflowRepository) CountEventWorkflowsByProjectID(ctx context.Context, projectID string) (int64, error) {
	if m.CountEventWorkflowsByProjectIDFn != nil {
		return m.CountEventWorkflowsByProjectIDFn(ctx, projectID)
	}
	return 0, nil
}

func (m *MockEventWorkflowRepository) CheckIfWorkflowNameExists(ctx context.Context, projectID, name string, excludeID string) (bool, error) {
	if m.CheckIfWorkflowNameExistsFn != nil {
		return m.CheckIfWorkflowNameExistsFn(ctx, projectID, name, excludeID)
	}
	return false, nil
}

func (m *MockEventWorkflowRepository) GetEventWorkflowsWithFilters(ctx context.Context, projectID string, enabled *bool, searchName string) ([]models.EventWorkflow, error) {
	if m.GetEventWorkflowsWithFiltersFn != nil {
		return m.GetEventWorkflowsWithFiltersFn(ctx, projectID, enabled, searchName)
	}
	return []models.EventWorkflow{}, nil
}