package services

import (
	"context"
	"errors"
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
	if projectID == "" {
		return nil, errors.New("projectId is required")
	}
	if email == "" {
		return nil, errors.New("email is required")
	}
	if invitedBy == "" {
		return nil, errors.New("invitedBy is required")
	}
	if role == "" {
		role = models.RoleEditor
	}

	project, err := s.projectRepo.GetProjectByID(ctx, projectID, invitedBy)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, errors.New("project does not exist")
	}

	inviter, err := s.userRepo.GetUserByID(ctx, invitedBy)
	if err != nil {
		return nil, fmt.Errorf("failed to get inviter details: %w", err)
	}
	if inviter == nil {
		return nil, errors.New("inviter does not exist")
	}

	if inviter.Email == email {
		return nil, errors.New("cannot invite yourself")
	}

	existing, err := s.invitationRepo.GetInvitationsByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	for _, inv := range existing {
		if inv.Email == email && inv.AcceptedAt == nil {
			return nil, errors.New("invitation already sent to this email")
		}
	}

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

	if s.emailService != nil {
		inviteLink := fmt.Sprintf("%s/accept-invitation?token=%s", s.baseURL, invitation.Token)
		if err := s.emailService.SendInvitationEmail(email, project.Name, inviteLink); err != nil {
			fmt.Printf("Failed to send invitation email: %v\n", err)
		}
	}

	return created, nil
}

func (s *InvitationService) CheckProjectOwnership(ctx context.Context, projectID, userID string) error {
	if projectID == "" {
		return errors.New("projectId is required")
	}
	if userID == "" {
		return errors.New("userId is required")
	}

	project, err := s.projectRepo.GetProjectByID(ctx, projectID, userID)
	if err != nil {
		return err
	}
	if project == nil {
		return errors.New("project does not exist")
	}
	if project.OwnerId != userID {
		return errors.New("user is not the owner of this project")
	}

	return nil
}

func (s *InvitationService) GetInvitationByID(ctx context.Context, id string) (*models.Invitation, error) {
	if id == "" {
		return nil, errors.New("invitation id is required")
	}

	return s.invitationRepo.GetInvitationByID(ctx, id)
}

func (s *InvitationService) AcceptInvitation(ctx context.Context, token, userID string) error {
	if token == "" {
		return errors.New("token is required")
	}
	if userID == "" {
		return errors.New("userId is required")
	}

	return s.invitationRepo.AcceptInvitation(ctx, token, userID)
}

func (s *InvitationService) GetInvitationsByProject(ctx context.Context, projectID string) ([]models.Invitation, error) {
	if projectID == "" {
		return nil, errors.New("projectId is required")
	}

	if err := s.CheckProjectOwnership(ctx, projectID, ""); err != nil {
		return nil, err
	}

	return s.invitationRepo.GetInvitationsByProject(ctx, projectID)
}

func (s *InvitationService) DeleteInvitation(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("invitation id is required")
	}

	invitation, err := s.GetInvitationByID(ctx, id)
	if err != nil {
		return err
	}
	if invitation == nil {
		return errors.New("invitation does not exist")
	}

	return s.invitationRepo.DeleteInvitation(ctx, id)
}

func (s *InvitationService) CancelInvitation(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("invitation id is required")
	}

	invitation, err := s.GetInvitationByID(ctx, id)
	if err != nil {
		return err
	}
	if invitation == nil {
		return errors.New("invitation does not exist")
	}

	return s.invitationRepo.CancelInvitation(ctx, id)
}

func (s *InvitationService) UpdateInvitationStatus(ctx context.Context, id string, status models.InvitationStatus) error {
	if id == "" {
		return errors.New("invitation id is required")
	}
	if status == "" {
		return errors.New("status is required")
	}

	invitation, err := s.GetInvitationByID(ctx, id)
	if err != nil {
		return err
	}
	if invitation == nil {
		return errors.New("invitation does not exist")
	}

	return s.invitationRepo.UpdateInvitationStatus(ctx, id, status)
}

func (s *InvitationService) GetPendingInvitationsByProject(ctx context.Context, projectID string) ([]models.Invitation, error) {
	if projectID == "" {
		return nil, errors.New("projectId is required")
	}

	if err := s.CheckProjectOwnership(ctx, projectID, ""); err != nil {
		return nil, err
	}

	return s.invitationRepo.GetPendingInvitationsByProject(ctx, projectID)
}