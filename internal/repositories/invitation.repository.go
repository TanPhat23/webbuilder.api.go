package repositories

import (
	"context"
	"my-go-app/internal/models"
	"time"

	"github.com/lucsky/cuid"
	"gorm.io/gorm"
)

type InvitationRepository struct {
	DB *gorm.DB
}

func NewInvitationRepository(db *gorm.DB) *InvitationRepository {
	return &InvitationRepository{
		DB: db,
	}
}

func (r *InvitationRepository) CreateInvitation(ctx context.Context, invitation *models.Invitation) (*models.Invitation, error) {
	if invitation.Id == "" {
		invitation.Id = cuid.New()
	}
	// Set default status if not provided
	if invitation.Status == "" {
		invitation.Status = models.InvitationStatusPending
	}
	if err := r.DB.WithContext(ctx).Create(invitation).Error; err != nil {
		return nil, err
	}
	return invitation, nil
}

func (r *InvitationRepository) AcceptInvitation(ctx context.Context, token string, userID string) error {
	var invitation models.Invitation
	err := r.DB.WithContext(ctx).Where(&models.Invitation{Token: token}).First(&invitation).Error
	if err != nil {
		return err
	}

	// Check if expired
	if time.Now().After(invitation.ExpiresAt) {
		return gorm.ErrRecordNotFound // or custom error
	}

	// Check if already accepted or cancelled
	if invitation.Status != models.InvitationStatusPending {
		return gorm.ErrRecordNotFound // or custom error for invalid status
	}

	// Create collaborator
	collaborator := models.Collaborator{
		Id:        cuid.New(),
		UserId:    userID,
		ProjectId: invitation.ProjectId,
		Role:      invitation.Role,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := r.DB.WithContext(ctx).Create(&collaborator).Error; err != nil {
		return err
	}

	// Update invitation as accepted
	now := time.Now()
	return r.DB.WithContext(ctx).Model(&invitation).Updates(map[string]interface{}{
		"Status":     models.InvitationStatusAccepted,
		"AcceptedAt": &now,
	}).Error
}

func (r *InvitationRepository) GetInvitationsByProject(ctx context.Context, projectID string) ([]models.Invitation, error) {
	var invitations []models.Invitation
	err := r.DB.WithContext(ctx).Where(&models.Invitation{ProjectId: projectID}).Find(&invitations).Error
	return invitations, err
}

func (r *InvitationRepository) GetInvitationByID(ctx context.Context, id string) (*models.Invitation, error) {
	var invitation models.Invitation
	err := r.DB.WithContext(ctx).Where(&models.Invitation{Id: id}).First(&invitation).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &invitation, nil
}

func (r *InvitationRepository) GetInvitationByToken(ctx context.Context, token string) (*models.Invitation, error) {
	var invitation models.Invitation
	err := r.DB.WithContext(ctx).Where(&models.Invitation{Token: token}).First(&invitation).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &invitation, nil
}

func (r *InvitationRepository) DeleteInvitation(ctx context.Context, id string) error {
	return r.DB.WithContext(ctx).Delete(&models.Invitation{}, id).Error
}

// UpdateInvitationStatus updates the status of an invitation
func (r *InvitationRepository) UpdateInvitationStatus(ctx context.Context, id string, status models.InvitationStatus) error {
	return r.DB.WithContext(ctx).Model(&models.Invitation{}).Where("Id = ?", id).Update("Status", status).Error
}

// CancelInvitation cancels an invitation
func (r *InvitationRepository) CancelInvitation(ctx context.Context, id string) error {
	return r.UpdateInvitationStatus(ctx, id, models.InvitationStatusCancelled)
}

// ExpireInvitation marks an invitation as expired
func (r *InvitationRepository) ExpireInvitation(ctx context.Context, id string) error {
	return r.UpdateInvitationStatus(ctx, id, models.InvitationStatusExpired)
}

// GetPendingInvitationsByProject gets all pending invitations for a project
func (r *InvitationRepository) GetPendingInvitationsByProject(ctx context.Context, projectID string) ([]models.Invitation, error) {
	var invitations []models.Invitation
	err := r.DB.WithContext(ctx).Where(&models.Invitation{
		ProjectId: projectID,
		Status:    models.InvitationStatusPending,
	}).Find(&invitations).Error
	return invitations, err
}
