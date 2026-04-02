package test

import (
	"context"
	"my-go-app/internal/models"
)

// MockCollaboratorRepo implements CollaboratorRepositoryInterface for testing.
type MockCollaboratorRepo struct {
	*GenericMock
}

// NewMockCollaboratorRepo creates a new MockCollaboratorRepo instance.
func NewMockCollaboratorRepo() *MockCollaboratorRepo {
	return &MockCollaboratorRepo{GenericMock: &GenericMock{funcs: make(map[string]any)}}
}

func (m *MockCollaboratorRepo) SetCreateCollaborator(fn func(context.Context, *models.Collaborator) (*models.Collaborator, error)) *MockCollaboratorRepo {
	m.Set("CreateCollaborator", fn)
	return m
}

func (m *MockCollaboratorRepo) SetGetCollaboratorsByProject(fn func(context.Context, string) ([]models.Collaborator, error)) *MockCollaboratorRepo {
	m.Set("GetCollaboratorsByProject", fn)
	return m
}

func (m *MockCollaboratorRepo) SetGetCollaboratorByID(fn func(context.Context, string) (*models.Collaborator, error)) *MockCollaboratorRepo {
	m.Set("GetCollaboratorByID", fn)
	return m
}

func (m *MockCollaboratorRepo) SetUpdateCollaboratorRole(fn func(context.Context, string, models.CollaboratorRole) error) *MockCollaboratorRepo {
	m.Set("UpdateCollaboratorRole", fn)
	return m
}

func (m *MockCollaboratorRepo) SetDeleteCollaborator(fn func(context.Context, string) error) *MockCollaboratorRepo {
	m.Set("DeleteCollaborator", fn)
	return m
}

func (m *MockCollaboratorRepo) SetIsCollaborator(fn func(context.Context, string, string) (bool, error)) *MockCollaboratorRepo {
	m.Set("IsCollaborator", fn)
	return m
}

func (m *MockCollaboratorRepo) CreateCollaborator(ctx context.Context, collaborator *models.Collaborator) (*models.Collaborator, error) {
	if fn := m.Get("CreateCollaborator"); fn != nil {
		return fn.(func(context.Context, *models.Collaborator) (*models.Collaborator, error))(ctx, collaborator)
	}
	return nil, nil
}

func (m *MockCollaboratorRepo) GetCollaboratorsByProject(ctx context.Context, projectID string) ([]models.Collaborator, error) {
	if fn := m.Get("GetCollaboratorsByProject"); fn != nil {
		return fn.(func(context.Context, string) ([]models.Collaborator, error))(ctx, projectID)
	}
	return nil, nil
}

func (m *MockCollaboratorRepo) GetCollaboratorByID(ctx context.Context, id string) (*models.Collaborator, error) {
	if fn := m.Get("GetCollaboratorByID"); fn != nil {
		return fn.(func(context.Context, string) (*models.Collaborator, error))(ctx, id)
	}
	return nil, nil
}

func (m *MockCollaboratorRepo) UpdateCollaboratorRole(ctx context.Context, id string, role models.CollaboratorRole) error {
	if fn := m.Get("UpdateCollaboratorRole"); fn != nil {
		return fn.(func(context.Context, string, models.CollaboratorRole) error)(ctx, id, role)
	}
	return nil
}

func (m *MockCollaboratorRepo) DeleteCollaborator(ctx context.Context, id string) error {
	if fn := m.Get("DeleteCollaborator"); fn != nil {
		return fn.(func(context.Context, string) error)(ctx, id)
	}
	return nil
}

func (m *MockCollaboratorRepo) IsCollaborator(ctx context.Context, projectID, userID string) (bool, error) {
	if fn := m.Get("IsCollaborator"); fn != nil {
		return fn.(func(context.Context, string, string) (bool, error))(ctx, projectID, userID)
	}
	return false, nil
}