package services

import (
	"context"
	"my-go-app/internal/models"
)

type EmailServiceInterface interface {
	SendInvitationEmail(to, projectName, inviteLink string) error
}

type InvitationServiceInterface interface {
	CreateInvitation(ctx context.Context, projectID, email string, role models.CollaboratorRole, invitedBy string) (*models.Invitation, error)
	AcceptInvitation(ctx context.Context, token, userID string) error
	GetInvitationsByProject(ctx context.Context, projectID string) ([]models.Invitation, error)
	DeleteInvitation(ctx context.Context, id string) error
}
