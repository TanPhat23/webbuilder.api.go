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
	if err := r.DB.WithContext(ctx).Create(invitation).Error; err != nil {
		return nil, err
	}
	return invitation, nil
}

func (r *InvitationRepository) GetInvitationsByProject(ctx context.Context, projectID string) ([]models.Invitation, error) {
	var invitations []models.Invitation
	err := r.DB.WithContext(ctx).Where(&models.Invitation{ProjectId: projectID}).Find(&invitations).Error
	return invitations, err
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

	// Update invitation accepted
	now := time.Now()
	return r.DB.WithContext(ctx).Model(&invitation).Updates(map[string]interface{}{
		"AcceptedAt": &now,
	}).Error
}

func (r *InvitationRepository) DeleteInvitation(ctx context.Context, id string) error {
	return r.DB.WithContext(ctx).Delete(&models.Invitation{}, id).Error
}
