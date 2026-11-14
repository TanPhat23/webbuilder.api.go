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
	if err := gorm.G[models.Collaborator](r.db).Create(ctx, collaborator); err != nil {
		return nil, err
	}
	return collaborator, nil
}

func (r *CollaboratorRepository) GetCollaboratorsByProject(ctx context.Context, projectID string) ([]models.Collaborator, error) {
	collaborators, err := gorm.G[models.Collaborator](r.db).
		Where(&models.Collaborator{ProjectId: projectID}).
		Preload("User", nil).
		Find(ctx)
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
	project, err := gorm.G[models.Project](r.db).Where(&models.Project{ID: projectID}).First(ctx)
	if err != nil {
		return nil, err
	}

	// Fetch the owner and append them as a synthetic "owner" collaborator.
	owner, err := gorm.G[models.User](r.db).Where(&models.User{Id: project.OwnerId}).First(ctx)
	if err == nil {
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
			BaseAuditFields: models.BaseAuditFields{
				CreatedAt: project.CreatedAt,
				UpdatedAt: project.UpdatedAt,
			},
			User: owner,
		}
		collaborators = append(collaborators, ownerCollaborator)
	}

	return collaborators, nil
}

func (r *CollaboratorRepository) GetCollaboratorByID(ctx context.Context, id string) (*models.Collaborator, error) {

	collaborator, err := gorm.G[models.Collaborator](r.db).
		Preload("User", nil).
		Preload("Project", nil).
		Where(`"Id" = ?`, id).
		First(ctx)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &collaborator, nil
}

func (r *CollaboratorRepository) UpdateCollaboratorRole(ctx context.Context, id string, role models.CollaboratorRole) error {
	_, err := gorm.G[models.Collaborator](r.db).
		Where(`"Id" = ?`, id).
		Update(ctx, "Role", role)
	return err
}

func (r *CollaboratorRepository) DeleteCollaborator(ctx context.Context, id string) error {
	_, err := gorm.G[models.Collaborator](r.db).Where(`"Id" = ?`, id).
		Delete(ctx)
	return err
}

func (r *CollaboratorRepository) IsCollaborator(ctx context.Context, projectID, userID string) (bool, error) {

	count, err := gorm.G[models.Collaborator](r.db).
		Where(&models.Collaborator{ProjectId: projectID, UserId: userID}).
		Count(ctx, "*")
	return count > 0, err
}
