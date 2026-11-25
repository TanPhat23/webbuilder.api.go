package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"my-go-app/internal/authz"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"

	"my-go-app/pkg/ids"
)

type ProjectService struct {
	projectRepo      repositories.ProjectRepositoryInterface
	collaboratorRepo repositories.CollaboratorRepositoryInterface
	userRepo         repositories.UserRepositoryInterface
	onAccessChanged  func(projectID, userID string)
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

// SetAccessChangedHook registers a callback invoked after ownership transfer
// so active realtime connections can refresh permissions.
func (s *ProjectService) SetAccessChangedHook(fn func(projectID, userID string)) {
	s.onAccessChanged = fn
}

var ErrInvalidSubdomain = errors.New("invalid subdomain")

func normalizeProjectSubdomainPtr(raw *string) (*string, error) {
	if raw == nil {
		return nil, nil
	}
	normalized, ok, err := models.NormalizeSubdomain(*raw)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidSubdomain, err)
	}
	if !ok {
		return nil, nil
	}
	return &normalized, nil
}

func (s *ProjectService) GetPublicByID(ctx context.Context, projectID string) (*models.Project, error) {
	if projectID == "" {
		return nil, errors.New("projectId is required")
	}

	return s.projectRepo.GetPublicByID(ctx, projectID)
}

func (s *ProjectService) GetPublicBySubdomain(ctx context.Context, subdomain string) (*models.Project, error) {
	if subdomain == "" {
		return nil, errors.New("subdomain is required")
	}
	return s.projectRepo.GetPublicBySubdomain(ctx, subdomain)
}

func (s *ProjectService) GetByID(ctx context.Context, projectID, userID string) (*models.Project, error) {
	if projectID == "" {
		return nil, errors.New("projectId is required")
	}
	if userID == "" {
		return nil, errors.New("userId is required")
	}

	project, err := s.projectRepo.GetByID(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, errors.New("project does not exist")
	}

	return project, nil
}

func (s *ProjectService) GetWithAccess(ctx context.Context, projectID, userID string) (*models.Project, error) {
	if projectID == "" {
		return nil, errors.New("projectId is required")
	}
	if userID == "" {
		return nil, errors.New("userId is required")
	}

	project, err := s.projectRepo.GetWithAccess(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, errors.New("project does not exist")
	}

	return project, nil
}

func (s *ProjectService) ListByUser(ctx context.Context, userID string) ([]models.Project, error) {
	if userID == "" {
		return nil, errors.New("userId is required")
	}

	return s.projectRepo.ListByUser(ctx, userID)
}

func (s *ProjectService) ListByCollaborator(ctx context.Context, userID string) ([]models.Project, error) {
	if userID == "" {
		return nil, errors.New("userId is required")
	}

	return s.projectRepo.ListByCollaborator(ctx, userID)
}

func (s *ProjectService) ListPages(ctx context.Context, projectID, userID string) ([]models.Page, error) {
	if projectID == "" {
		return nil, errors.New("projectId is required")
	}

	if err := ensureProjectVisible(ctx, s.projectRepo, projectID, userID); err != nil {
		return nil, fmt.Errorf("failed to verify project: %w", err)
	}

	return s.projectRepo.ListPages(ctx, projectID, userID)
}

func (s *ProjectService) Create(ctx context.Context, project *models.Project) (*models.Project, error) {
	if project == nil {
		return nil, errors.New("project cannot be nil")
	}
	if project.Name == "" {
		return nil, errors.New("project name is required")
	}
	if project.OwnerId == "" {
		return nil, errors.New("ownerId is required")
	}
	sub, err := normalizeProjectSubdomainPtr(project.Subdomain)
	if err != nil {
		return nil, err
	}
	project.Subdomain = sub

	if err := s.projectRepo.Create(ctx, project); err != nil {
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
	existing, err := s.GetWithAccess(ctx, projectID, userID)
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

func (s *ProjectService) Update(ctx context.Context, projectID, userID string, project *models.Project) (*models.Project, error) {
	if project == nil {
		return nil, errors.New("project cannot be nil")
	}
	existing, err := s.requireProjectAccess(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}
	return s.projectRepo.Update(ctx, projectID, userID, mergeProjectUpdateFields(existing, project))
}

// UpdateFieldsForUser enforces per-field RBAC permissions before updating:
// Published requires project:publish (owner), Subdomain requires project:settings (owner),
// and Name/Description/Styles/Header require project:edit (owner or editor).
func (s *ProjectService) UpdateFieldsForUser(ctx context.Context, projectID, userID string, updates map[string]any) (*models.Project, error) {
	if projectID == "" {
		return nil, errors.New("projectId is required")
	}
	if userID == "" {
		return nil, errors.New("userId is required")
	}
	if len(updates) == 0 {
		return nil, errors.New("no updates provided")
	}

	authorizer := NewProjectAuthorizer(s.projectRepo)
	if _, hasPublished := updates["Published"]; hasPublished {
		if _, err := authorizer.Authorize(ctx, projectID, userID, authz.PermProjectPublish); err != nil {
			return nil, err
		}
	}
	if _, hasSubdomain := updates["Subdomain"]; hasSubdomain {
		if _, err := authorizer.Authorize(ctx, projectID, userID, authz.PermProjectSettings); err != nil {
			return nil, err
		}
		switch v := updates["Subdomain"].(type) {
		case nil:
			updates["Subdomain"] = nil
		case string:
			sub, err := normalizeProjectSubdomainPtr(&v)
			if err != nil {
				return nil, err
			}
			if sub == nil {
				updates["Subdomain"] = nil
			} else {
				updates["Subdomain"] = *sub
			}
		}
	}
	for _, field := range []string{"Name", "Description", "Styles", "Header"} {
		if _, ok := updates[field]; ok {
			if _, err := authorizer.Authorize(ctx, projectID, userID, authz.PermProjectEdit); err != nil {
				return nil, err
			}
			break
		}
	}

	return s.projectRepo.UpdateFields(ctx, projectID, userID, updates)
}

// TransferOwnership moves Project.OwnerId to newOwnerID when authorized with project:transfer.
func (s *ProjectService) TransferOwnership(ctx context.Context, projectID, currentOwnerID, newOwnerID string) error {
	if projectID == "" {
		return errors.New("projectId is required")
	}
	if currentOwnerID == "" {
		return errors.New("userId is required")
	}
	if newOwnerID == "" {
		return errors.New("newOwnerId is required")
	}
	if _, err := authorizeProjectPerm(ctx, s.projectRepo, projectID, currentOwnerID, authz.PermProjectTransfer); err != nil {
		return err
	}
	if err := s.projectRepo.TransferOwnership(ctx, projectID, currentOwnerID, newOwnerID); err != nil {
		return err
	}
	if s.onAccessChanged != nil {
		s.onAccessChanged(projectID, currentOwnerID)
		s.onAccessChanged(projectID, newOwnerID)
	}
	return nil
}

func (s *ProjectService) Delete(ctx context.Context, projectID, userID string) error {
	if _, err := s.requireProjectAccess(ctx, projectID, userID); err != nil {
		return err
	}
	return s.projectRepo.Delete(ctx, projectID, userID)
}

func (s *ProjectService) HardDelete(ctx context.Context, projectID, userID string) error {
	if _, err := s.requireProjectAccess(ctx, projectID, userID); err != nil {
		return err
	}
	return s.projectRepo.HardDelete(ctx, projectID, userID)
}

func (s *ProjectService) Restore(ctx context.Context, projectID, userID string) error {
	if _, err := s.requireProjectAccess(ctx, projectID, userID); err != nil {
		return err
	}
	return s.projectRepo.Restore(ctx, projectID, userID)
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

func (s *ProjectService) GetWithLock(ctx context.Context, projectID, userID string) (*models.Project, error) {
	if projectID == "" {
		return nil, errors.New("projectId is required")
	}
	if userID == "" {
		return nil, errors.New("userId is required")
	}

	return s.projectRepo.GetWithLock(ctx, projectID, userID)
}

// CreateWithPage creates a project and its mandatory first page in
// one transaction, so a project never exists without a page.
func (s *ProjectService) CreateWithPage(ctx context.Context, project *models.Project) (*models.Project, error) {
	if project == nil || project.Name == "" || project.OwnerId == "" {
		return nil, errors.New("project name and ownerId are required")
	}
	sub, err := normalizeProjectSubdomainPtr(project.Subdomain)
	if err != nil {
		return nil, err
	}
	project.Subdomain = sub
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
		Slug:   "",
		Type:   "page",
		Styles: json.RawMessage(`{}`),
	}
	if err := s.projectRepo.CreateWithPage(ctx, project, page); err != nil {
		return nil, err
	}
	return project, nil
}

func (s *ProjectService) GetAccess(ctx context.Context, projectID, userID string) (*models.ProjectAccess, error) {
	return s.projectRepo.GetAccess(ctx, projectID, userID)
}
