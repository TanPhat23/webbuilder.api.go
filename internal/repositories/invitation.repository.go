package repositories

import (
	"context"
	"errors"
	"fmt"
	"my-go-app/internal/models"
	"time"

	"github.com/lucsky/cuid"
	"gorm.io/gorm"
)

var (
	ErrInvitationNotFound = errors.New("invitation not found")
	ErrInvitationExpired  = errors.New("invitation has expired")
	ErrInvitationInvalid  = errors.New("invitation is no longer pending")
)

type InvitationRepository struct {
	db *gorm.DB
}

func NewInvitationRepository(db *gorm.DB) InvitationRepositoryInterface {
	return &InvitationRepository{db: db}
}

func (r *InvitationRepository) CreateInvitation(ctx context.Context, invitation *models.Invitation) (*models.Invitation, error) {
	if invitation == nil {
		return nil, errors.New("invitation cannot be nil")
	}
	if invitation.Id == "" {
		invitation.Id = cuid.New()
	}
	if invitation.Status == "" {
		invitation.Status = models.InvitationStatusPending
	}
	if err := r.db.WithContext(ctx).Create(invitation).Error; err != nil {
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}
	return invitation, nil
}

func (r *InvitationRepository) GetInvitationsByProject(ctx context.Context, projectID string) ([]models.Invitation, error) {
	if projectID == "" {
		return nil, errors.New("projectID is required")
	}
	var invitations []models.Invitation
	if err := r.db.WithContext(ctx).
		Where(&models.Invitation{ProjectId: projectID}).
		Find(&invitations).Error; err != nil {
		return nil, fmt.Errorf("failed to get invitations by project: %w", err)
	}
	return invitations, nil
}

func (r *InvitationRepository) GetInvitationByID(ctx context.Context, id string) (*models.Invitation, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}
	var invitation models.Invitation
	err := r.db.WithContext(ctx).
		Where(&models.Invitation{Id: id}).
		First(&invitation).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvitationNotFound
		}
		return nil, fmt.Errorf("failed to get invitation by ID: %w", err)
	}
	return &invitation, nil
}

func (r *InvitationRepository) GetInvitationByToken(ctx context.Context, token string) (*models.Invitation, error) {
	if token == "" {
		return nil, errors.New("token is required")
	}
	var invitation models.Invitation
	err := r.db.WithContext(ctx).
		Where(&models.Invitation{Token: token}).
		First(&invitation).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvitationNotFound
		}
		return nil, fmt.Errorf("failed to get invitation by token: %w", err)
	}
	return &invitation, nil
}

func (r *InvitationRepository) AcceptInvitation(ctx context.Context, token string, userID string) error {
	if token == "" || userID == "" {
		return errors.New("token and userID are required")
	}

	invitation, err := r.GetInvitationByToken(ctx, token)
	if err != nil {
		return err
	}

	if time.Now().After(invitation.ExpiresAt) {
		return ErrInvitationExpired
	}

	if invitation.Status != models.InvitationStatusPending {
		return ErrInvitationInvalid
	}

	collaborator := models.Collaborator{
		Id:        cuid.New(),
		UserId:    userID,
		ProjectId: invitation.ProjectId,
		Role:      invitation.Role,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := r.db.WithContext(ctx).Create(&collaborator).Error; err != nil {
		return fmt.Errorf("failed to create collaborator from invitation: %w", err)
	}

	now := time.Now()
	if err := r.db.WithContext(ctx).
		Model(invitation).
		Updates(map[string]any{
			"Status":     models.InvitationStatusAccepted,
			"AcceptedAt": &now,
		}).Error; err != nil {
		return fmt.Errorf("failed to mark invitation as accepted: %w", err)
	}

	return nil
}

func (r *InvitationRepository) DeleteInvitation(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id is required")
	}
	result := r.db.WithContext(ctx).Delete(&models.Invitation{}, `"Id" = ?`, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete invitation: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrInvitationNotFound
	}
	return nil
}

func (r *InvitationRepository) UpdateInvitationStatus(ctx context.Context, id string, status models.InvitationStatus) error {
	if id == "" {
		return errors.New("id is required")
	}
	result := r.db.WithContext(ctx).
		Model(&models.Invitation{}).
		Where(`"Id" = ?`, id).
		Update("Status", status)
	if result.Error != nil {
		return fmt.Errorf("failed to update invitation status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrInvitationNotFound
	}
	return nil
}

func (r *InvitationRepository) CancelInvitation(ctx context.Context, id string) error {
	return r.UpdateInvitationStatus(ctx, id, models.InvitationStatusCancelled)
}

func (r *InvitationRepository) GetPendingInvitationsByProject(ctx context.Context, projectID string) ([]models.Invitation, error) {
	if projectID == "" {
		return nil, errors.New("projectID is required")
	}
	var invitations []models.Invitation
	if err := r.db.WithContext(ctx).
		Where(&models.Invitation{
			ProjectId: projectID,
			Status:    models.InvitationStatusPending,
		}).
		Find(&invitations).Error; err != nil {
		return nil, fmt.Errorf("failed to get pending invitations: %w", err)
	}
	return invitations, nil
}