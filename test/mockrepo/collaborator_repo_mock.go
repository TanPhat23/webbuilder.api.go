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

func (m *MockCollaboratorRepo) SetCreate(fn func(context.Context, *models.Collaborator) (*models.Collaborator, error)) *MockCollaboratorRepo {
	m.Set("Create", fn)
	return m
}

func (m *MockCollaboratorRepo) SetListByProject(fn func(context.Context, string) ([]models.Collaborator, error)) *MockCollaboratorRepo {
	m.Set("ListByProject", fn)
	return m
}

func (m *MockCollaboratorRepo) SetGetByID(fn func(context.Context, string) (*models.Collaborator, error)) *MockCollaboratorRepo {
	m.Set("GetByID", fn)
	return m
}

func (m *MockCollaboratorRepo) SetUpdateRole(fn func(context.Context, string, models.CollaboratorRole) error) *MockCollaboratorRepo {
	m.Set("UpdateRole", fn)
	return m
}

func (m *MockCollaboratorRepo) SetDelete(fn func(context.Context, string) error) *MockCollaboratorRepo {
	m.Set("Delete", fn)
	return m
}

func (m *MockCollaboratorRepo) SetIsCollaborator(fn func(context.Context, string, string) (bool, error)) *MockCollaboratorRepo {
	m.Set("IsCollaborator", fn)
	return m
}

func (m *MockCollaboratorRepo) Create(ctx context.Context, collaborator *models.Collaborator) (*models.Collaborator, error) {
	if fn := m.Get("Create"); fn != nil {
		return fn.(func(context.Context, *models.Collaborator) (*models.Collaborator, error))(ctx, collaborator)
	}
	return nil, nil
}

func (m *MockCollaboratorRepo) ListByProject(ctx context.Context, projectID string) ([]models.Collaborator, error) {
	if fn := m.Get("ListByProject"); fn != nil {
		return fn.(func(context.Context, string) ([]models.Collaborator, error))(ctx, projectID)
	}
	return nil, nil
}

func (m *MockCollaboratorRepo) GetByID(ctx context.Context, id string) (*models.Collaborator, error) {
	if fn := m.Get("GetByID"); fn != nil {
		return fn.(func(context.Context, string) (*models.Collaborator, error))(ctx, id)
	}
	return nil, nil
}

func (m *MockCollaboratorRepo) UpdateRole(ctx context.Context, id string, role models.CollaboratorRole) error {
	if fn := m.Get("UpdateRole"); fn != nil {
		return fn.(func(context.Context, string, models.CollaboratorRole) error)(ctx, id, role)
	}
	return nil
}

func (m *MockCollaboratorRepo) Delete(ctx context.Context, id string) error {
	if fn := m.Get("Delete"); fn != nil {
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

func (m *MockCollaboratorRepo) GetByProjectAndUser(ctx context.Context, projectID, userID string) (*models.Collaborator, error) {
	if fn := m.Get("GetByProjectAndUser"); fn != nil {
		return fn.(func(context.Context, string, string) (*models.Collaborator, error))(ctx, projectID, userID)
	}
	return nil, nil
}

func (m *MockCollaboratorRepo) DeleteByProjectAndUser(ctx context.Context, projectID, userID string) (bool, error) {
	if fn := m.Get("DeleteByProjectAndUser"); fn != nil {
		return fn.(func(context.Context, string, string) (bool, error))(ctx, projectID, userID)
	}
	return false, nil
}
