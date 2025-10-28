package services

import (
	"context"
	"fmt"
	"time"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"

	"github.com/lucsky/cuid"
)

type InvitationService struct {
	invitationRepo   repositories.InvitationRepositoryInterface
	collaboratorRepo repositories.CollaboratorRepositoryInterface
	projectRepo      repositories.ProjectRepositoryInterface
	userRepo         repositories.UserRepositoryInterface
	emailService     EmailServiceInterface
	baseURL          string
}

func NewInvitationService(
	invitationRepo repositories.InvitationRepositoryInterface,
	collaboratorRepo repositories.CollaboratorRepositoryInterface,
	projectRepo repositories.ProjectRepositoryInterface,
	userRepo repositories.UserRepositoryInterface,
	emailService EmailServiceInterface,
	baseURL string,
) *InvitationService {
	return &InvitationService{
		invitationRepo:   invitationRepo,
		collaboratorRepo: collaboratorRepo,
		projectRepo:      projectRepo,
		userRepo:         userRepo,
		emailService:     emailService,
		baseURL:          baseURL,
	}
}

func (s *InvitationService) CreateInvitation(ctx context.Context, projectID, email string, role models.CollaboratorRole, invitedBy string) (*models.Invitation, error) {
	// Check if user is owner of the project
	project, err := s.projectRepo.GetProjectByID(ctx, projectID, invitedBy)
	if err != nil {
		return nil, err
	}

	// Get the inviter's email to prevent self-invitation
	inviter, err := s.userRepo.GetUserByID(ctx, invitedBy)
	if err != nil {
		return nil, fmt.Errorf("failed to get inviter details: %w", err)
	}

	// Prevent self-invitation
	if inviter.Email == email {
		return nil, fmt.Errorf("cannot invite yourself")
	}

	// Check if invitation already exists
	existing, err := s.invitationRepo.GetInvitationsByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	for _, inv := range existing {
		if inv.Email == email && inv.AcceptedAt == nil {
			return nil, fmt.Errorf("invitation already sent to this email")
		}
	}

	// Create invitation
	invitation := &models.Invitation{
		Id:        cuid.New(),
		Email:     email,
		ProjectId: projectID,
		Role:      role,
		Token:     cuid.New(),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		CreatedAt: time.Now(),
	}

	created, err := s.invitationRepo.CreateInvitation(ctx, invitation)
	if err != nil {
		return nil, err
	}

	// Send email
	inviteLink := fmt.Sprintf("%s/accept-invitation?token=%s", s.baseURL, invitation.Token)
	err = s.emailService.SendInvitationEmail(email, project.Name, inviteLink)
	if err != nil {
		fmt.Printf("Failed to send invitation email: %v\n", err)
	}

	return created, nil
}

func (s *InvitationService) CheckProjectOwnership(ctx context.Context, projectID, userID string) error {
	_, err := s.projectRepo.GetProjectByID(ctx, projectID, userID)
	return err
}

func (s *InvitationService) GetInvitationByID(ctx context.Context, id string) (*models.Invitation, error) {
	return s.invitationRepo.GetInvitationByID(ctx, id)
}

func (s *InvitationService) AcceptInvitation(ctx context.Context, token, userID string) error {
	return s.invitationRepo.AcceptInvitation(ctx, token, userID)
}

func (s *InvitationService) GetInvitationsByProject(ctx context.Context, projectID string) ([]models.Invitation, error) {
	return s.invitationRepo.GetInvitationsByProject(ctx, projectID)
}

func (s *InvitationService) DeleteInvitation(ctx context.Context, id string) error {
	return s.invitationRepo.DeleteInvitation(ctx, id)
}
