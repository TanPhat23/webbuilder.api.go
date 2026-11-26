package services

import (
	"context"
	"errors"
	"my-go-app/internal/authz"
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

// ListByProject lists a project's pages. userID is the signed-in caller,
// or "" on the public route, where only published projects are readable.
func (s *PageService) ListByProject(ctx context.Context, projectID, userID string) ([]models.Page, error) {
	if projectID == "" {
		return nil, errors.New("projectId is required")
	}

	if err := ensureProjectVisible(ctx, s.projectRepo, projectID, userID); err != nil {
		return nil, err
	}

	pages, err := s.pageRepo.ListByProject(ctx, projectID)
	if err != nil || userID != "" {
		return pages, err
	}

	routes := make([]models.Page, 0, len(pages))
	for _, page := range pages {
		if page.Type != models.PageTypeComponent {
			routes = append(routes, page)
		}
	}
	return routes, nil
}

func (s *PageService) GetByID(ctx context.Context, pageID string) (*models.Page, error) {
	if pageID == "" {
		return nil, errors.New("pageId is required")
	}

	return s.pageRepo.GetByID(ctx, pageID, "")
}

// GetByIDForUser verifies project read access (page:view) and ensures the
// page belongs to projectID.
func (s *PageService) GetByIDForUser(ctx context.Context, projectID, pageID, userID string) (*models.Page, error) {
	if projectID == "" {
		return nil, errors.New("projectId is required")
	}
	if pageID == "" {
		return nil, errors.New("pageId is required")
	}
	if _, err := authorizeProjectPerm(ctx, s.projectRepo, projectID, userID, authz.PermPageView); err != nil {
		return nil, err
	}
	return s.pageRepo.GetByID(ctx, pageID, projectID)
}

// GetPublicByID returns routable pages of a published project.
func (s *PageService) GetPublicByID(ctx context.Context, projectID, pageID string) (*models.Page, error) {
	if projectID == "" || pageID == "" {
		return nil, errors.New("projectId and pageId are required")
	}
	project, err := s.projectRepo.GetPublicByID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, repositories.ErrProjectNotFound
	}
	page, err := s.pageRepo.GetByID(ctx, pageID, projectID)
	if err != nil {
		return nil, err
	}
	if page == nil || page.Type == models.PageTypeComponent {
		return nil, repositories.ErrPageNotFound
	}
	return page, nil
}

func (s *PageService) Create(ctx context.Context, page *models.Page) (*models.Page, error) {
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

	project, err := s.projectRepo.GetPublicByID(ctx, page.ProjectId)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, errors.New("project does not exist")
	}

	return s.insertPage(ctx, page)
}

var ErrInvalidPageSlug = errors.New("dynamic :param or * segments are only allowed on dynamic pages")

func (s *PageService) checkRouteCollision(ctx context.Context, projectID, excludePageID, pageType, slug string) error {
	if pageType == models.PageTypeComponent {
		return nil
	}
	if pageType != models.PageTypeDynamic && models.HasDynamicSegments(slug) {
		return ErrInvalidPageSlug
	}
	if !models.HasDynamicSegments(slug) {
		return nil
	}
	canonical := models.CanonicalRoutePattern(slug)
	pages, err := s.pageRepo.ListByProject(ctx, projectID)
	if err != nil {
		return err
	}
	for _, existing := range pages {
		if existing.Id == excludePageID || existing.Type == models.PageTypeComponent {
			continue
		}
		if models.CanonicalRoutePattern(existing.Slug) == canonical {
			return repositories.ErrPageSlugTaken
		}
	}
	return nil
}

// insertPage stores a page whose project the caller has already verified.
func (s *PageService) insertPage(ctx context.Context, page *models.Page) (*models.Page, error) {
	if err := s.checkRouteCollision(ctx, page.ProjectId, "", page.Type, page.Slug); err != nil {
		return nil, err
	}
	if err := s.pageRepo.Create(ctx, page); err != nil {
		return nil, err
	}
	return page, nil
}

func (s *PageService) Update(ctx context.Context, pageID string, page *models.Page) (*models.Page, error) {
	if pageID == "" {
		return nil, errors.New("pageId is required")
	}
	if page == nil {
		return nil, errors.New("page cannot be nil")
	}

	page.Id = pageID
	if err := s.pageRepo.Update(ctx, page); err != nil {
		return nil, err
	}
	return page, nil
}

func (s *PageService) UpdateFields(ctx context.Context, pageID string, updates map[string]interface{}) error {
	if pageID == "" {
		return errors.New("pageId is required")
	}
	if len(updates) == 0 {
		return errors.New("no updates provided")
	}

	return s.pageRepo.UpdateFields(ctx, pageID, updates)
}

// UpdateFieldsForUser verifies page:edit on projectID and confirms pageID
// belongs to projectID before applying updates.
func (s *PageService) UpdateFieldsForUser(ctx context.Context, projectID, pageID, userID string, updates map[string]interface{}) error {
	if projectID == "" {
		return errors.New("projectId is required")
	}
	if pageID == "" {
		return errors.New("pageId is required")
	}
	if len(updates) == 0 {
		return errors.New("no updates provided")
	}
	if _, err := authorizeProjectPerm(ctx, s.projectRepo, projectID, userID, authz.PermPageEdit); err != nil {
		return err
	}
	existing, err := s.pageRepo.GetByID(ctx, pageID, projectID)
	if err != nil {
		return err
	}
	effectiveType := existing.Type
	if nextType, ok := updates["Type"].(string); ok && nextType != "" {
		effectiveType = nextType
	}
	effectiveSlug := existing.Slug
	if nextSlug, ok := updates["Slug"].(string); ok {
		effectiveSlug = nextSlug
	}
	if err := s.checkRouteCollision(ctx, projectID, pageID, effectiveType, effectiveSlug); err != nil {
		return err
	}
	if effectiveType != models.PageTypeDynamic && existing.ContentTypeId != nil {
		updates["ContentTypeId"] = nil
	}
	return s.pageRepo.UpdateFields(ctx, pageID, updates)
}

func (s *PageService) Delete(ctx context.Context, pageID string) error {
	if pageID == "" {
		return errors.New("pageId is required")
	}

	return s.pageRepo.Delete(ctx, pageID)
}

func (s *PageService) DeleteForUser(ctx context.Context, pageID, projectID, userID string) error {
	if pageID == "" {
		return errors.New("pageId is required")
	}
	if projectID == "" {
		return errors.New("projectId is required")
	}
	if userID == "" {
		return errors.New("userId is required")
	}

	if _, err := authorizeProjectPerm(ctx, s.projectRepo, projectID, userID, authz.PermPageDelete); err != nil {
		return err
	}

	return s.pageRepo.DeleteForUser(ctx, pageID, projectID, userID)
}

func (s *PageService) DeleteInProject(ctx context.Context, pageID, projectID string) error {
	if pageID == "" {
		return errors.New("pageId is required")
	}
	if projectID == "" {
		return errors.New("projectId is required")
	}

	return s.pageRepo.DeleteForUser(ctx, pageID, projectID, "")
}

// CreateForUser creates a page after checking the user holds page:create
// (owner or editor). Viewers and strangers get ErrForbidden.
func (s *PageService) CreateForUser(ctx context.Context, page *models.Page, userID string) (*models.Page, error) {
	if page == nil {
		return nil, errors.New("page cannot be nil")
	}
	if _, err := authorizeProjectPerm(ctx, s.projectRepo, page.ProjectId, userID, authz.PermPageCreate); err != nil {
		return nil, err
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
