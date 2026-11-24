package services

import (
	"context"
	"errors"
	"fmt"
	"my-go-app/internal/authz"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
)

type ContentTypeService struct {
	contentTypeRepo repositories.ContentTypeRepositoryInterface
	projectRepo     repositories.ProjectRepositoryInterface
}

func NewContentTypeService(contentTypeRepo repositories.ContentTypeRepositoryInterface) *ContentTypeService {
	return &ContentTypeService{
		contentTypeRepo: contentTypeRepo,
	}
}

// WithProjectRepo configures project-scoped RBAC checks for CMS content types.
func (s *ContentTypeService) WithProjectRepo(projectRepo repositories.ProjectRepositoryInterface) *ContentTypeService {
	s.projectRepo = projectRepo
	return s
}

func (s *ContentTypeService) GetContentTypes(ctx context.Context) ([]models.ContentType, error) {
	return s.contentTypeRepo.GetContentTypes(ctx)
}

func (s *ContentTypeService) GetContentTypesForUser(ctx context.Context, userID, projectID string) ([]models.ContentType, error) {
	if userID == "" {
		return nil, authz.ErrUnauthorized
	}
	if projectID != "" {
		if s.projectRepo != nil {
			if _, err := authorizeProjectPerm(ctx, s.projectRepo, projectID, userID, authz.PermCMSView); err != nil {
				return nil, err
			}
		}
		return s.contentTypeRepo.GetContentTypesByProject(ctx, projectID)
	}
	return s.contentTypeRepo.GetContentTypesByUser(ctx, userID)
}

func (s *ContentTypeService) GetContentTypeByID(ctx context.Context, id string) (*models.ContentType, error) {
	return s.contentTypeRepo.GetContentTypeByID(ctx, id)
}

func (s *ContentTypeService) GetContentTypeByIDForUser(ctx context.Context, id, userID string) (*models.ContentType, error) {
	if userID == "" {
		return nil, authz.ErrUnauthorized
	}
	ct, err := s.contentTypeRepo.GetContentTypeByID(ctx, id)
	if err != nil || ct == nil {
		return ct, err
	}
	if s.projectRepo != nil {
		if ct.ProjectId == nil || *ct.ProjectId == "" {
			return nil, fmt.Errorf("%w: content type is not scoped to a project", authz.ErrForbidden)
		}
		if _, err := authorizeProjectPerm(ctx, s.projectRepo, *ct.ProjectId, userID, authz.PermCMSView); err != nil {
			return nil, err
		}
	}
	return ct, nil
}

func (s *ContentTypeService) CreateContentType(ctx context.Context, contentType *models.ContentType) (*models.ContentType, error) {
	return s.contentTypeRepo.CreateContentType(ctx, contentType)
}

func (s *ContentTypeService) CreateContentTypeForUser(ctx context.Context, contentType *models.ContentType, userID string) (*models.ContentType, error) {
	if userID == "" {
		return nil, authz.ErrUnauthorized
	}
	if contentType == nil {
		return nil, errors.New("content type cannot be nil")
	}
	if contentType.ProjectId == nil || *contentType.ProjectId == "" {
		return nil, errors.New("projectId is required")
	}
	if s.projectRepo != nil {
		if _, err := authorizeProjectPerm(ctx, s.projectRepo, *contentType.ProjectId, userID, authz.PermCMSEdit); err != nil {
			return nil, err
		}
	}
	return s.contentTypeRepo.CreateContentType(ctx, contentType)
}

func (s *ContentTypeService) UpdateContentType(ctx context.Context, id string, contentType *models.ContentType) (*models.ContentType, error) {
	updates := map[string]any{
		"name":        contentType.Name,
		"description": contentType.Description,
	}
	return s.contentTypeRepo.UpdateContentType(ctx, id, updates)
}

func (s *ContentTypeService) UpdateContentTypeForUser(ctx context.Context, id, userID string, contentType *models.ContentType) (*models.ContentType, error) {
	if userID == "" {
		return nil, authz.ErrUnauthorized
	}
	existing, err := s.contentTypeRepo.GetContentTypeByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, repositories.ErrContentTypeNotFound
	}
	if s.projectRepo != nil {
		if existing.ProjectId == nil || *existing.ProjectId == "" {
			return nil, fmt.Errorf("%w: content type is not scoped to a project", authz.ErrForbidden)
		}
		if _, err := authorizeProjectPerm(ctx, s.projectRepo, *existing.ProjectId, userID, authz.PermCMSEdit); err != nil {
			return nil, err
		}
	}
	return s.UpdateContentType(ctx, id, contentType)
}

func (s *ContentTypeService) DeleteContentType(ctx context.Context, id string) error {
	return s.contentTypeRepo.DeleteContentType(ctx, id)
}

func (s *ContentTypeService) DeleteContentTypeForUser(ctx context.Context, id, userID string) error {
	if userID == "" {
		return authz.ErrUnauthorized
	}
	existing, err := s.contentTypeRepo.GetContentTypeByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return repositories.ErrContentTypeNotFound
	}
	if s.projectRepo != nil {
		if existing.ProjectId == nil || *existing.ProjectId == "" {
			return fmt.Errorf("%w: content type is not scoped to a project", authz.ErrForbidden)
		}
		if _, err := authorizeProjectPerm(ctx, s.projectRepo, *existing.ProjectId, userID, authz.PermCMSEdit); err != nil {
			return err
		}
	}
	return s.contentTypeRepo.DeleteContentType(ctx, id)
}
