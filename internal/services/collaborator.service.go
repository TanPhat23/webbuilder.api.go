package services

import (
	"context"
	"errors"
	"fmt"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
)

type CollaboratorService struct {
	collaboratorRepo repositories.CollaboratorRepositoryInterface
	projectRepo      repositories.ProjectRepositoryInterface
}

func NewCollaboratorService(
	collaboratorRepo repositories.CollaboratorRepositoryInterface,
	projectRepo repositories.ProjectRepositoryInterface,
) *CollaboratorService {
	return &CollaboratorService{
		collaboratorRepo: collaboratorRepo,
		projectRepo:      projectRepo,
	}
}

func (s *CollaboratorService) CreateCollaborator(ctx context.Context, collaborator *models.Collaborator) (*models.Collaborator, error) {
	if collaborator == nil {
		return nil, errors.New("collaborator cannot be nil")
	}
	if collaborator.ProjectId == "" {
		return nil, errors.New("projectId is required")
	}
	if collaborator.UserId == "" {
		return nil, errors.New("userId is required")
	}

	project, err := s.projectRepo.GetPublicProjectByID(ctx, collaborator.ProjectId)
	if err != nil {
		return nil, fmt.Errorf("failed to verify project: %w", err)
	}
	if project == nil {
		return nil, errors.New("project does not exist")
	}

	return s.collaboratorRepo.CreateCollaborator(ctx, collaborator)
}

func (s *CollaboratorService) GetCollaboratorsByProject(ctx context.Context, projectID string) ([]models.Collaborator, error) {
	if projectID == "" {
		return nil, errors.New("projectId is required")
	}

	project, err := s.projectRepo.GetPublicProjectByID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify project: %w", err)
	}
	if project == nil {
		return nil, errors.New("project does not exist")
	}

	return s.collaboratorRepo.GetCollaboratorsByProject(ctx, projectID)
}

func (s *CollaboratorService) GetCollaboratorByID(ctx context.Context, id string) (*models.Collaborator, error) {
	if id == "" {
		return nil, errors.New("collaborator id is required")
	}

	return s.collaboratorRepo.GetCollaboratorByID(ctx, id)
}

func (s *CollaboratorService) UpdateCollaboratorRole(ctx context.Context, id string, role models.CollaboratorRole) error {
	if id == "" {
		return errors.New("collaborator id is required")
	}

	collaborator, err := s.GetCollaboratorByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to retrieve collaborator: %w", err)
	}
	if collaborator == nil {
		return errors.New("collaborator does not exist")
	}

	if collaborator.ProjectId == "" {
		return errors.New("collaborator projectId is required")
	}

	project, err := s.projectRepo.GetProjectByID(ctx, collaborator.ProjectId, collaborator.UserId)
	if err != nil {
		return fmt.Errorf("failed to verify project ownership: %w", err)
	}
	if project == nil {
		return errors.New("project does not exist")
	}

	return s.collaboratorRepo.UpdateCollaboratorRole(ctx, id, role)
}

func (s *CollaboratorService) DeleteCollaborator(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("collaborator id is required")
	}

	collaborator, err := s.GetCollaboratorByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to retrieve collaborator: %w", err)
	}
	if collaborator == nil {
		return errors.New("collaborator does not exist")
	}

	return s.collaboratorRepo.DeleteCollaborator(ctx, id)
}

func (s *CollaboratorService) IsCollaborator(ctx context.Context, userID, projectID string) (bool, error) {
	if userID == "" {
		return false, errors.New("userId is required")
	}
	if projectID == "" {
		return false, errors.New("projectId is required")
	}

	return s.collaboratorRepo.IsCollaborator(ctx, projectID, userID)
}

func (s *CollaboratorService) CheckOwnership(ctx context.Context, projectID, userID string) error {
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