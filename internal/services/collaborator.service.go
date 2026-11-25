package services

import (
	"context"
	"errors"
	"fmt"

	"my-go-app/internal/authz"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
)

type CollaboratorService struct {
	collaboratorRepo repositories.CollaboratorRepositoryInterface
	projectRepo      repositories.ProjectRepositoryInterface
	onAccessChanged  func(projectID, userID string)
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

// SetAccessChangedHook registers a callback invoked after role updates or removals
// so active realtime WebSocket clients immediately refresh permissions.
func (s *CollaboratorService) SetAccessChangedHook(fn func(projectID, userID string)) {
	s.onAccessChanged = fn
}

func (s *CollaboratorService) Create(ctx context.Context, callerUserID string, collaborator *models.Collaborator) (*models.Collaborator, error) {
	if callerUserID == "" {
		return nil, authz.ErrUnauthorized
	}
	if collaborator == nil {
		return nil, errors.New("collaborator cannot be nil")
	}
	if collaborator.ProjectId == "" {
		return nil, errors.New("projectId is required")
	}
	if collaborator.UserId == "" {
		return nil, errors.New("userId is required")
	}
	if collaborator.Role == "" {
		collaborator.Role = models.RoleEditor
	}
	if !authz.IsAssignableCollaboratorRole(authz.Role(collaborator.Role)) {
		return nil, fmt.Errorf("%w: role %q cannot be assigned to a collaborator", authz.ErrForbidden, collaborator.Role)
	}

	if _, err := authorizeProjectPerm(ctx, s.projectRepo, collaborator.ProjectId, callerUserID, authz.PermCollaboratorInvite); err != nil {
		return nil, err
	}
	if collaborator.UserId == callerUserID {
		return nil, errors.New("project owner is already a member of this project")
	}
	existing, err := s.collaboratorRepo.GetByProjectAndUser(ctx, collaborator.ProjectId, collaborator.UserId)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing collaborator: %w", err)
	}
	if existing != nil {
		return nil, errors.New("user is already a collaborator on this project")
	}

	created, err := s.collaboratorRepo.Create(ctx, collaborator)
	if err != nil {
		return nil, err
	}
	if s.onAccessChanged != nil {
		s.onAccessChanged(collaborator.ProjectId, collaborator.UserId)
	}
	return created, nil
}

func (s *CollaboratorService) ListByProject(ctx context.Context, projectID, userID string) ([]models.Collaborator, error) {
	if projectID == "" {
		return nil, errors.New("projectId is required")
	}
	if userID == "" {
		return nil, authz.ErrUnauthorized
	}

	if _, err := authorizeProjectPerm(ctx, s.projectRepo, projectID, userID, authz.PermCollaboratorView); err != nil {
		return nil, err
	}

	return s.collaboratorRepo.ListByProject(ctx, projectID)
}

func (s *CollaboratorService) GetByID(ctx context.Context, id, callerUserID string) (*models.Collaborator, error) {
	if id == "" {
		return nil, errors.New("collaborator id is required")
	}
	if callerUserID == "" {
		return nil, authz.ErrUnauthorized
	}

	collaborator, err := s.collaboratorRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if collaborator == nil {
		return nil, fmt.Errorf("%w: collaborator not found", authz.ErrNotFound)
	}
	if _, err := authorizeProjectPerm(ctx, s.projectRepo, collaborator.ProjectId, callerUserID, authz.PermCollaboratorView); err != nil {
		return nil, err
	}
	return collaborator, nil
}

func (s *CollaboratorService) UpdateRole(ctx context.Context, id, callerUserID string, role models.CollaboratorRole) error {
	if id == "" {
		return errors.New("collaborator id is required")
	}
	if callerUserID == "" {
		return authz.ErrUnauthorized
	}
	if !authz.IsAssignableCollaboratorRole(authz.Role(role)) {
		return fmt.Errorf("%w: role %q cannot be assigned to a collaborator", authz.ErrForbidden, role)
	}

	collaborator, err := s.collaboratorRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to retrieve collaborator: %w", err)
	}
	if collaborator == nil {
		return fmt.Errorf("%w: collaborator does not exist", authz.ErrNotFound)
	}
	if collaborator.ProjectId == "" {
		return errors.New("collaborator projectId is required")
	}
	if collaborator.UserId == callerUserID {
		return fmt.Errorf("%w: cannot change your own role", authz.ErrForbidden)
	}

	if _, err := authorizeProjectPerm(ctx, s.projectRepo, collaborator.ProjectId, callerUserID, authz.PermCollaboratorEdit); err != nil {
		return err
	}

	if err := s.collaboratorRepo.UpdateRole(ctx, id, role); err != nil {
		return err
	}
	if s.onAccessChanged != nil {
		s.onAccessChanged(collaborator.ProjectId, collaborator.UserId)
	}
	return nil
}

func (s *CollaboratorService) Delete(ctx context.Context, id, callerUserID string) error {
	if id == "" {
		return errors.New("collaborator id is required")
	}
	if callerUserID == "" {
		return authz.ErrUnauthorized
	}

	collaborator, err := s.collaboratorRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to retrieve collaborator: %w", err)
	}
	if collaborator == nil {
		return fmt.Errorf("%w: collaborator does not exist", authz.ErrNotFound)
	}

	if collaborator.UserId != callerUserID {
		if _, err := authorizeProjectPerm(ctx, s.projectRepo, collaborator.ProjectId, callerUserID, authz.PermCollaboratorRemove); err != nil {
			return err
		}
	}

	if err := s.collaboratorRepo.Delete(ctx, id); err != nil {
		return err
	}
	if s.onAccessChanged != nil {
		s.onAccessChanged(collaborator.ProjectId, collaborator.UserId)
	}
	return nil
}

func (s *CollaboratorService) LeaveProject(ctx context.Context, projectID, userID string) error {
	if projectID == "" {
		return errors.New("projectId is required")
	}
	if userID == "" {
		return authz.ErrUnauthorized
	}

	access, err := authorizeProjectPerm(ctx, s.projectRepo, projectID, userID, authz.PermProjectView)
	if err != nil {
		return err
	}
	if access.IsOwner {
		return fmt.Errorf("%w: project owner must transfer ownership before leaving", authz.ErrForbidden)
	}

	removed, err := s.collaboratorRepo.DeleteByProjectAndUser(ctx, projectID, userID)
	if err != nil {
		return err
	}
	if !removed {
		return fmt.Errorf("%w: collaborator membership not found", authz.ErrNotFound)
	}
	if s.onAccessChanged != nil {
		s.onAccessChanged(projectID, userID)
	}
	return nil
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

func (s *CollaboratorService) RequireOwnership(ctx context.Context, projectID, userID string) error {
	if projectID == "" {
		return errors.New("projectId is required")
	}
	if userID == "" {
		return errors.New("userId is required")
	}

	_, err := authorizeProjectPerm(ctx, s.projectRepo, projectID, userID, authz.PermProjectDelete)
	return err
}
