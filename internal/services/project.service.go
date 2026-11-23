package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"

	"my-go-app/pkg/ids"
)

type ProjectService struct {
	projectRepo      repositories.ProjectRepositoryInterface
	collaboratorRepo repositories.CollaboratorRepositoryInterface
	userRepo         repositories.UserRepositoryInterface
}

func NewProjectService(
	projectRepo repositories.ProjectRepositoryInterface,
	collaboratorRepo repositories.CollaboratorRepositoryInterface,
	userRepo repositories.UserRepositoryInterface,
) *ProjectService {
	return &ProjectService{
		projectRepo:      projectRepo,
		collaboratorRepo: collaboratorRepo,
		userRepo:         userRepo,
	}
}

func (s *ProjectService) GetPublicProjectByID(ctx context.Context, projectID string) (*models.Project, error) {
	if projectID == "" {
		return nil, errors.New("projectId is required")
	}

	return s.projectRepo.GetPublicProjectByID(ctx, projectID)
}

func (s *ProjectService) GetProjectByID(ctx context.Context, projectID, userID string) (*models.Project, error) {
	if projectID == "" {
		return nil, errors.New("projectId is required")
	}
	if userID == "" {
		return nil, errors.New("userId is required")
	}

	project, err := s.projectRepo.GetProjectByID(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, errors.New("project does not exist")
	}

	return project, nil
}

func (s *ProjectService) GetProjectWithAccess(ctx context.Context, projectID, userID string) (*models.Project, error) {
	if projectID == "" {
		return nil, errors.New("projectId is required")
	}
	if userID == "" {
		return nil, errors.New("userId is required")
	}

	project, err := s.projectRepo.GetProjectWithAccess(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, errors.New("project does not exist")
	}

	return project, nil
}

func (s *ProjectService) GetProjectsByUserID(ctx context.Context, userID string) ([]models.Project, error) {
	if userID == "" {
		return nil, errors.New("userId is required")
	}

	return s.projectRepo.GetProjectsByUserID(ctx, userID)
}

func (s *ProjectService) GetCollaboratorProjects(ctx context.Context, userID string) ([]models.Project, error) {
	if userID == "" {
		return nil, errors.New("userId is required")
	}

	return s.projectRepo.GetCollaboratorProjects(ctx, userID)
}

func (s *ProjectService) GetProjectPages(ctx context.Context, projectID, userID string) ([]models.Page, error) {
	if projectID == "" {
		return nil, errors.New("projectId is required")
	}

	if err := ensureProjectVisible(ctx, s.projectRepo, projectID, userID); err != nil {
		return nil, fmt.Errorf("failed to verify project: %w", err)
	}

	return s.projectRepo.GetProjectPages(ctx, projectID, userID)
}

func (s *ProjectService) CreateProject(ctx context.Context, project *models.Project) (*models.Project, error) {
	if project == nil {
		return nil, errors.New("project cannot be nil")
	}
	if project.Name == "" {
		return nil, errors.New("project name is required")
	}
	if project.OwnerId == "" {
		return nil, errors.New("ownerId is required")
	}

	if err := s.projectRepo.CreateProject(ctx, project); err != nil {
		return nil, err
	}
	return project, nil
}

func (s *ProjectService) requireProjectAccess(ctx context.Context, projectID, userID string) (*models.Project, error) {
	if projectID == "" {
		return nil, errors.New("projectId is required")
	}
	if userID == "" {
		return nil, errors.New("userId is required")
	}
	existing, err := s.GetProjectWithAccess(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("project does not exist")
	}
	return existing, nil
}

func mergeProjectUpdateFields(existing, project *models.Project) map[string]any {
	if project.Name == "" {
		project.Name = existing.Name
	}
	if project.Description == nil {
		project.Description = existing.Description
	}
	if project.Subdomain == nil {
		project.Subdomain = existing.Subdomain
	}
	if project.Styles == nil {
		project.Styles = existing.Styles
	}
	if project.Header == nil {
		project.Header = existing.Header
	}
	return map[string]any{
		"Name":        project.Name,
		"Description": project.Description,
		"Published":   project.Published,
		"Subdomain":   project.Subdomain,
		"Styles":      project.Styles,
		"Header":      project.Header,
	}
}

func (s *ProjectService) UpdateProject(ctx context.Context, projectID, userID string, project *models.Project) (*models.Project, error) {
	if project == nil {
		return nil, errors.New("project cannot be nil")
	}
	existing, err := s.requireProjectAccess(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}
	return s.projectRepo.UpdateProject(ctx, projectID, userID, mergeProjectUpdateFields(existing, project))
}

func (s *ProjectService) DeleteProject(ctx context.Context, projectID, userID string) error {
	if _, err := s.requireProjectAccess(ctx, projectID, userID); err != nil {
		return err
	}
	return s.projectRepo.DeleteProject(ctx, projectID, userID)
}

func (s *ProjectService) HardDeleteProject(ctx context.Context, projectID, userID string) error {
	if _, err := s.requireProjectAccess(ctx, projectID, userID); err != nil {
		return err
	}
	return s.projectRepo.HardDeleteProject(ctx, projectID, userID)
}

func (s *ProjectService) RestoreProject(ctx context.Context, projectID, userID string) error {
	if _, err := s.requireProjectAccess(ctx, projectID, userID); err != nil {
		return err
	}
	return s.projectRepo.RestoreProject(ctx, projectID, userID)
}

func (s *ProjectService) ExistsForUser(ctx context.Context, projectID, userID string) (bool, error) {
	if projectID == "" {
		return false, errors.New("projectId is required")
	}
	if userID == "" {
		return false, errors.New("userId is required")
	}

	return s.projectRepo.ExistsForUser(ctx, projectID, userID)
}

func (s *ProjectService) GetProjectWithLock(ctx context.Context, projectID, userID string) (*models.Project, error) {
	if projectID == "" {
		return nil, errors.New("projectId is required")
	}
	if userID == "" {
		return nil, errors.New("userId is required")
	}

	return s.projectRepo.GetProjectWithLock(ctx, projectID, userID)
}

// CreateProjectWithPage creates a project and its mandatory first page in
// one transaction, so a project never exists without a page.
func (s *ProjectService) CreateProjectWithPage(ctx context.Context, project *models.Project) (*models.Project, error) {
	if project == nil || project.Name == "" || project.OwnerId == "" {
		return nil, errors.New("project name and ownerId are required")
	}
	if project.ID == "" {
		id, err := ids.NewV7()
		if err != nil {
			return nil, err
		}
		project.ID = id
	}
	pageID, err := ids.NewV7()
	if err != nil {
		return nil, err
	}
	page := &models.Page{
		Id:     pageID,
		Name:   "",
		Type:   "page",
		Styles: json.RawMessage(`{}`),
	}
	if err := s.projectRepo.CreateProjectWithPage(ctx, project, page); err != nil {
		return nil, err
	}
	return project, nil
}

func (s *ProjectService) GetProjectAccess(ctx context.Context, projectID, userID string) (*models.ProjectAccess, error) {
	return s.projectRepo.GetProjectAccess(ctx, projectID, userID)
}
