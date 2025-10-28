package repositories

import (
	"context"
	"fmt"
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
	if err != nil {
		return nil, err
	}

	fmt.Printf("Found %d collaborators from DB\n", len(collaborators))

	// Modify User fields for collaborators
	for i := range collaborators {
		if collaborators[i].User.FirstName == nil {
			collaborators[i].User.FirstName = &collaborators[i].User.Email
		}
		if collaborators[i].User.LastName == nil {
			empty := ""
			collaborators[i].User.LastName = &empty
		}
	}

	// Fetch the project to get OwnerId
	var project models.Project
	err = r.DB.WithContext(ctx).Where(&models.Project{ID: projectID}).First(&project).Error
	if err != nil {
		fmt.Printf("Error fetching project: %v\n", err)
		return nil, err
	}

	// Fetch the owner user
	var owner models.User
	err = r.DB.WithContext(ctx).Where(&models.User{Id: project.OwnerId}).First(&owner).Error
	if err != nil {
		fmt.Printf("Error fetching owner user: %v, skipping owner in collaborators\n", err)
	} else {
		// Modify owner User fields
		if owner.FirstName == nil {
			owner.FirstName = &owner.Email
		}
		if owner.LastName == nil {
			empty := ""
			owner.LastName = &empty
		}

		// Add owner as a collaborator
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

	fmt.Printf("Total collaborators after adding owner: %d\n", len(collaborators))

	return collaborators, nil
}

func (r *CollaboratorRepository) GetCollaboratorByID(ctx context.Context, id string) (*models.Collaborator, error) {
	var collaborator models.Collaborator
	err := r.DB.WithContext(ctx).Preload("User").Preload("Project").Where("\"Id\" = ?", id).First(&collaborator).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &collaborator, nil
}

func (r *CollaboratorRepository) UpdateCollaboratorRole(ctx context.Context, id string, role models.CollaboratorRole) error {
	return r.DB.WithContext(ctx).Model(&models.Collaborator{}).Where("\"Id\" = ?", id).Update("\"Role\"", role).Error
}

func (r *CollaboratorRepository) DeleteCollaborator(ctx context.Context, id string) error {
	return r.DB.WithContext(ctx).Where("\"Id\" = ?", id).Delete(&models.Collaborator{}).Error
}

func (r *CollaboratorRepository) IsCollaborator(ctx context.Context, projectID, userID string) (bool, error) {
	var count int64
	err := r.DB.WithContext(ctx).Model(&models.Collaborator{}).Where(&models.Collaborator{ProjectId: projectID, UserId: userID}).Count(&count).Error
	return count > 0, err
}
