package testutil

import (
	"context"

	"my-go-app/internal/models"
)

type MockCollaboratorRepository struct {
	CreateCollaboratorFn        func(ctx context.Context, collaborator *models.Collaborator) (*models.Collaborator, error)
	GetCollaboratorsByProjectFn func(ctx context.Context, projectID string) ([]models.Collaborator, error)
	GetCollaboratorByIDFn       func(ctx context.Context, id string) (*models.Collaborator, error)
	UpdateCollaboratorRoleFn    func(ctx context.Context, id string, role models.CollaboratorRole) error
	DeleteCollaboratorFn        func(ctx context.Context, id string) error
	IsCollaboratorFn            func(ctx context.Context, projectID, userID string) (bool, error)
}

func (m *MockCollaboratorRepository) CreateCollaborator(ctx context.Context, collaborator *models.Collaborator) (*models.Collaborator, error) {
	if m.CreateCollaboratorFn != nil {
		return m.CreateCollaboratorFn(ctx, collaborator)
	}
	return collaborator, nil
}

func (m *MockCollaboratorRepository) GetCollaboratorsByProject(ctx context.Context, projectID string) ([]models.Collaborator, error) {
	if m.GetCollaboratorsByProjectFn != nil {
		return m.GetCollaboratorsByProjectFn(ctx, projectID)
	}
	return []models.Collaborator{}, nil
}

func (m *MockCollaboratorRepository) GetCollaboratorByID(ctx context.Context, id string) (*models.Collaborator, error) {
	if m.GetCollaboratorByIDFn != nil {
		return m.GetCollaboratorByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *MockCollaboratorRepository) UpdateCollaboratorRole(ctx context.Context, id string, role models.CollaboratorRole) error {
	if m.UpdateCollaboratorRoleFn != nil {
		return m.UpdateCollaboratorRoleFn(ctx, id, role)
	}
	return nil
}

func (m *MockCollaboratorRepository) DeleteCollaborator(ctx context.Context, id string) error {
	if m.DeleteCollaboratorFn != nil {
		return m.DeleteCollaboratorFn(ctx, id)
	}
	return nil
}

func (m *MockCollaboratorRepository) IsCollaborator(ctx context.Context, projectID, userID string) (bool, error) {
	if m.IsCollaboratorFn != nil {
		return m.IsCollaboratorFn(ctx, projectID, userID)
	}
	return false, nil
}