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

func (s *PageService) GetPagesByProjectID(ctx context.Context, projectID string) ([]models.Page, error) {
	if projectID == "" {
		return nil, errors.New("projectId is required")
	}

	_, err := s.projectRepo.GetPublicProjectByID(ctx, projectID)
	if err != nil {
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