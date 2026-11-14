package services

import (
	"context"
	"errors"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
)

type PageService struct {
	pageRepo    repositories.PageRepositoryInterface
	projectRepo repositories.ProjectRepositoryInterface
}

func NewPageService(
	pageRepo repositories.PageRepositoryInterface,
	projectRepo repositories.ProjectRepositoryInterface,
) *PageService {
	return &PageService{
		pageRepo:    pageRepo,
		projectRepo: projectRepo,
	}
}

// GetPagesByProjectID lists a project's pages. userID is the signed-in caller,
// or "" on the public route, where only published projects are readable.
func (s *PageService) GetPagesByProjectID(ctx context.Context, projectID, userID string) ([]models.Page, error) {
	if projectID == "" {
		return nil, errors.New("projectId is required")
	}

	if err := ensureProjectVisible(ctx, s.projectRepo, projectID, userID); err != nil {
		return nil, err
	}

	return s.pageRepo.GetPagesByProjectID(ctx, projectID)
}

func (s *PageService) GetPageByID(ctx context.Context, pageID string) (*models.Page, error) {
	if pageID == "" {
		return nil, errors.New("pageId is required")
	}

	return s.pageRepo.GetPageByID(ctx, pageID, "")
}

// GetPublicPageByID returns a page only if it belongs to projectID and that
// project is published, gated like GetPagesByProjectID.
func (s *PageService) GetPublicPageByID(ctx context.Context, projectID, pageID string) (*models.Page, error) {
	if projectID == "" || pageID == "" {
		return nil, errors.New("projectId and pageId are required")
	}
	project, err := s.projectRepo.GetPublicProjectByID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, repositories.ErrProjectNotFound
	}
	return s.pageRepo.GetPageByID(ctx, pageID, projectID)
}

func (s *PageService) CreatePage(ctx context.Context, page *models.Page) (*models.Page, error) {
	if page == nil {
		return nil, errors.New("page cannot be nil")
	}
	if page.ProjectId == "" {
		return nil, errors.New("projectId is required")
	}
	if page.Name == "" {
		return nil, errors.New("page name is required")
	}
	if page.Type == "" {
		return nil, errors.New("page type is required")
	}

	project, err := s.projectRepo.GetPublicProjectByID(ctx, page.ProjectId)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, errors.New("project does not exist")
	}

	return s.insertPage(ctx, page)
}

// insertPage stores a page whose project the caller has already verified.
func (s *PageService) insertPage(ctx context.Context, page *models.Page) (*models.Page, error) {
	if err := s.pageRepo.CreatePage(ctx, page); err != nil {
		return nil, err
	}
	return page, nil
}

func (s *PageService) UpdatePage(ctx context.Context, pageID string, page *models.Page) (*models.Page, error) {
	if pageID == "" {
		return nil, errors.New("pageId is required")
	}
	if page == nil {
		return nil, errors.New("page cannot be nil")
	}

	page.Id = pageID
	if err := s.pageRepo.UpdatePage(ctx, page); err != nil {
		return nil, err
	}
	return page, nil
}

func (s *PageService) UpdatePageFields(ctx context.Context, pageID string, updates map[string]interface{}) error {
	if pageID == "" {
		return errors.New("pageId is required")
	}
	if len(updates) == 0 {
		return errors.New("no updates provided")
	}

	return s.pageRepo.UpdatePageFields(ctx, pageID, updates)
}

func (s *PageService) DeletePage(ctx context.Context, pageID string) error {
	if pageID == "" {
		return errors.New("pageId is required")
	}

	return s.pageRepo.DeletePage(ctx, pageID)
}

func (s *PageService) DeletePageByProjectID(ctx context.Context, pageID, projectID, userID string) error {
	if pageID == "" {
		return errors.New("pageId is required")
	}
	if projectID == "" {
		return errors.New("projectId is required")
	}
	if userID == "" {
		return errors.New("userId is required")
	}

	_, err := s.projectRepo.GetProjectWithAccess(ctx, projectID, userID)
	if err != nil {
		return err
	}

	return s.pageRepo.DeletePageByProjectID(ctx, pageID, projectID, userID)
}

func (s *PageService) DeletePageByProjectIDWithoutVerification(ctx context.Context, pageID, projectID string) error {
	if pageID == "" {
		return errors.New("pageId is required")
	}
	if projectID == "" {
		return errors.New("projectId is required")
	}

	return s.pageRepo.DeletePageByProjectID(ctx, pageID, projectID, "")
}

// CreatePageAsUser creates a page after checking the user may edit the
// project (owner or editor). Viewers and strangers get ErrProjectUnauthorized.
func (s *PageService) CreatePageAsUser(ctx context.Context, page *models.Page, userID string) (*models.Page, error) {
	if page == nil {
		return nil, errors.New("page cannot be nil")
	}
	access, err := s.projectRepo.GetProjectAccess(ctx, page.ProjectId, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrProjectNotFound) {
			return nil, repositories.ErrProjectUnauthorized
		}
		return nil, err
	}
	if access.Role != models.RoleOwner && access.Role != models.RoleEditor {
		return nil, repositories.ErrProjectUnauthorized
	}
	if page.Name == "" {
		return nil, errors.New("page name is required")
	}
	if page.Type == "" {
		return nil, errors.New("page type is required")
	}
	// Access was verified above, so an unpublished project is fine.
	return s.insertPage(ctx, page)
}
