package repositories

import (
	"context"
	"my-go-app/internal/models"

	"github.com/lucsky/cuid"
	"gorm.io/gorm"
)

type CollaboratorRepository struct {
	db *gorm.DB
}

func NewCollaboratorRepository(db *gorm.DB) CollaboratorRepositoryInterface {
	return &CollaboratorRepository{db: db}
}

func (r *CollaboratorRepository) CreateCollaborator(ctx context.Context, collaborator *models.Collaborator) (*models.Collaborator, error) {
	if collaborator.Id == "" {
		collaborator.Id = cuid.New()
	}
	if err := r.db.WithContext(ctx).Create(collaborator).Error; err != nil {
		return nil, err
	}
	return collaborator, nil
}

func (r *CollaboratorRepository) GetCollaboratorsByProject(ctx context.Context, projectID string) ([]models.Collaborator, error) {
	var collaborators []models.Collaborator
	err := r.db.WithContext(ctx).
		Where(&models.Collaborator{ProjectId: projectID}).
		Preload("User").
		Find(&collaborators).Error
	if err != nil {
		return nil, err
	}

	// Normalise missing name fields for regular collaborators.
	for i := range collaborators {
		if collaborators[i].User.FirstName == nil {
			collaborators[i].User.FirstName = &collaborators[i].User.Email
		}
		if collaborators[i].User.LastName == nil {
			empty := ""
			collaborators[i].User.LastName = &empty
		}
	}

	// Fetch the project to get OwnerId.
	var project models.Project
	if err = r.db.WithContext(ctx).Where(&models.Project{ID: projectID}).First(&project).Error; err != nil {
		return nil, err
	}

	// Fetch the owner and append them as a synthetic "owner" collaborator.
	var owner models.User
	if err = r.db.WithContext(ctx).Where(&models.User{Id: project.OwnerId}).First(&owner).Error; err == nil {
		if owner.FirstName == nil {
			owner.FirstName = &owner.Email
		}
		if owner.LastName == nil {
			empty := ""
			owner.LastName = &empty
		}

		ownerCollaborator := models.Collaborator{
			Id:        "owner-" + projectID,
			UserId:    project.OwnerId,
			ProjectId: projectID,
			Role:      models.RoleOwner,
			CreatedAt: project.CreatedAt,
			UpdatedAt: project.UpdatedAt,
			User:      owner,
		}
		collaborators = append(collaborators, ownerCollaborator)
	}

	return collaborators, nil
}

func (r *CollaboratorRepository) GetCollaboratorByID(ctx context.Context, id string) (*models.Collaborator, error) {
	var collaborator models.Collaborator
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Project").
		Where(`"Id" = ?`, id).
		First(&collaborator).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &collaborator, nil
}

func (r *CollaboratorRepository) UpdateCollaboratorRole(ctx context.Context, id string, role models.CollaboratorRole) error {
	return r.db.WithContext(ctx).
		Model(&models.Collaborator{}).
		Where(`"Id" = ?`, id).
		Update("Role", role).Error
}

func (r *CollaboratorRepository) DeleteCollaborator(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Where(`"Id" = ?`, id).
		Delete(&models.Collaborator{}).Error
}

func (r *CollaboratorRepository) IsCollaborator(ctx context.Context, projectID, userID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Collaborator{}).
		Where(&models.Collaborator{ProjectId: projectID, UserId: userID}).
		Count(&count).Error
	return count > 0, err
}
