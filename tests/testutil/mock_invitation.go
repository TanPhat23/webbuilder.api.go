package testutil

import (
	"context"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
)

type MockInvitationRepository struct {
	CreateInvitationFn               func(ctx context.Context, inv *models.Invitation) (*models.Invitation, error)
	GetInvitationsByProjectFn        func(ctx context.Context, projectID string) ([]models.Invitation, error)
	GetInvitationByIDFn              func(ctx context.Context, id string) (*models.Invitation, error)
	GetInvitationByTokenFn           func(ctx context.Context, token string) (*models.Invitation, error)
	AcceptInvitationFn               func(ctx context.Context, token, userID string) error
	DeleteInvitationFn               func(ctx context.Context, id string) error
	UpdateInvitationStatusFn         func(ctx context.Context, id string, status models.InvitationStatus) error
	CancelInvitationFn               func(ctx context.Context, id string) error
	GetPendingInvitationsByProjectFn func(ctx context.Context, projectID string) ([]models.Invitation, error)
}

func (m *MockInvitationRepository) CreateInvitation(ctx context.Context, inv *models.Invitation) (*models.Invitation, error) {
	if m.CreateInvitationFn != nil {
		return m.CreateInvitationFn(ctx, inv)
	}
	return inv, nil
}

func (m *MockInvitationRepository) GetInvitationsByProject(ctx context.Context, projectID string) ([]models.Invitation, error) {
	if m.GetInvitationsByProjectFn != nil {
		return m.GetInvitationsByProjectFn(ctx, projectID)
	}
	return []models.Invitation{}, nil
}

func (m *MockInvitationRepository) GetInvitationByID(ctx context.Context, id string) (*models.Invitation, error) {
	if m.GetInvitationByIDFn != nil {
		return m.GetInvitationByIDFn(ctx, id)
	}
	return nil, repositories.ErrInvitationNotFound
}

func (m *MockInvitationRepository) GetInvitationByToken(ctx context.Context, token string) (*models.Invitation, error) {
	if m.GetInvitationByTokenFn != nil {
		return m.GetInvitationByTokenFn(ctx, token)
	}
	return nil, repositories.ErrInvitationNotFound
}

func (m *MockInvitationRepository) AcceptInvitation(ctx context.Context, token, userID string) error {
	if m.AcceptInvitationFn != nil {
		return m.AcceptInvitationFn(ctx, token, userID)
	}
	return nil
}

func (m *MockInvitationRepository) DeleteInvitation(ctx context.Context, id string) error {
	if m.DeleteInvitationFn != nil {
		return m.DeleteInvitationFn(ctx, id)
	}
	return nil
}

func (m *MockInvitationRepository) UpdateInvitationStatus(ctx context.Context, id string, status models.InvitationStatus) error {
	if m.UpdateInvitationStatusFn != nil {
		return m.UpdateInvitationStatusFn(ctx, id, status)
	}
	return nil
}

func (m *MockInvitationRepository) CancelInvitation(ctx context.Context, id string) error {
	if m.CancelInvitationFn != nil {
		return m.CancelInvitationFn(ctx, id)
	}
	return nil
}

func (m *MockInvitationRepository) GetPendingInvitationsByProject(ctx context.Context, projectID string) ([]models.Invitation, error) {
	if m.GetPendingInvitationsByProjectFn != nil {
		return m.GetPendingInvitationsByProjectFn(ctx, projectID)
	}
	return []models.Invitation{}, nil
}