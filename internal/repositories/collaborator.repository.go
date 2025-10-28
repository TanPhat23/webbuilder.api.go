package repositories

import (
	"context"
	"my-go-app/internal/models"

	"github.com/lucsky/cuid"
	"gorm.io/gorm"
)

type CollaboratorRepository struct {
	DB *gorm.DB
}

func NewCollaboratorRepository(db *gorm.DB) *CollaboratorRepository {
	return &CollaboratorRepository{
		DB: db,
	}
}

func (r *CollaboratorRepository) CreateCollaborator(ctx context.Context, collaborator *models.Collaborator) (*models.Collaborator, error) {
	if collaborator.Id == "" {
		collaborator.Id = cuid.New()
	}
	if err := r.DB.WithContext(ctx).Create(collaborator).Error; err != nil {
		return nil, err
	}
	return collaborator, nil
}

func (r *CollaboratorRepository) GetCollaboratorsByProject(ctx context.Context, projectID string) ([]models.Collaborator, error) {
	var collaborators []models.Collaborator
	err := r.DB.WithContext(ctx).Where(&models.Collaborator{ProjectId: projectID}).Preload("User").Find(&collaborators).Error
	return collaborators, err
}

func (r *CollaboratorRepository) GetCollaboratorByID(ctx context.Context, id string) (*models.Collaborator, error) {
	var collaborator models.Collaborator
	err := r.DB.WithContext(ctx).Preload("User").Preload("Project").First(&collaborator, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &collaborator, nil
}

func (r *CollaboratorRepository) UpdateCollaboratorRole(ctx context.Context, id string, role models.CollaboratorRole) error {
	return r.DB.WithContext(ctx).Model(&models.Collaborator{}).Where("id = ?", id).Update("role", role).Error
}

func (r *CollaboratorRepository) DeleteCollaborator(ctx context.Context, id string) error {
	return r.DB.WithContext(ctx).Delete(&models.Collaborator{}, id).Error
}

func (r *CollaboratorRepository) IsCollaborator(ctx context.Context, projectID, userID string) (bool, error) {
	var count int64
	err := r.DB.WithContext(ctx).Model(&models.Collaborator{}).Where(&models.Collaborator{ProjectId: projectID, UserId: userID}).Count(&count).Error
	return count > 0, err
}
