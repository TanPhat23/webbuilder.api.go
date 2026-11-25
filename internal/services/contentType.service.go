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

func (s *ContentTypeService) List(ctx context.Context) ([]models.ContentType, error) {
	return s.contentTypeRepo.List(ctx)
}

func (s *ContentTypeService) ListForUser(ctx context.Context, userID, projectID string) ([]models.ContentType, error) {
	if userID == "" {
		return nil, authz.ErrUnauthorized
	}
	if projectID != "" {
		if s.projectRepo != nil {
			if _, err := authorizeProjectPerm(ctx, s.projectRepo, projectID, userID, authz.PermCMSView); err != nil {
				return nil, err
			}
		}
		return s.contentTypeRepo.ListByProject(ctx, projectID)
	}
	return s.contentTypeRepo.ListByUser(ctx, userID)
}

func (s *ContentTypeService) GetByID(ctx context.Context, id string) (*models.ContentType, error) {
	return s.contentTypeRepo.GetByID(ctx, id)
}

func (s *ContentTypeService) GetByIDForUser(ctx context.Context, id, userID string) (*models.ContentType, error) {
	if userID == "" {
		return nil, authz.ErrUnauthorized
	}
	ct, err := s.contentTypeRepo.GetByID(ctx, id)
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

func (s *ContentTypeService) Create(ctx context.Context, contentType *models.ContentType) (*models.ContentType, error) {
	return s.contentTypeRepo.Create(ctx, contentType)
}

func (s *ContentTypeService) CreateForUser(ctx context.Context, contentType *models.ContentType, userID string) (*models.ContentType, error) {
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
	return s.contentTypeRepo.Create(ctx, contentType)
}

func (s *ContentTypeService) Update(ctx context.Context, id string, contentType *models.ContentType) (*models.ContentType, error) {
	updates := map[string]any{
		"name":        contentType.Name,
		"description": contentType.Description,
	}
	return s.contentTypeRepo.Update(ctx, id, updates)
}

func (s *ContentTypeService) UpdateForUser(ctx context.Context, id, userID string, contentType *models.ContentType) (*models.ContentType, error) {
	updates := map[string]any{
		"name":        contentType.Name,
		"description": contentType.Description,
	}
	return s.PatchForUser(ctx, id, userID, updates)
}

func (s *ContentTypeService) PatchForUser(ctx context.Context, id, userID string, updates map[string]any) (*models.ContentType, error) {
	if userID == "" {
		return nil, authz.ErrUnauthorized
	}
	existing, err := s.contentTypeRepo.GetByID(ctx, id)
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
	return s.contentTypeRepo.Update(ctx, id, updates)
}

func (s *ContentTypeService) Delete(ctx context.Context, id string) error {
	return s.contentTypeRepo.Delete(ctx, id)
}

func (s *ContentTypeService) DeleteForUser(ctx context.Context, id, userID string) error {
	if userID == "" {
		return authz.ErrUnauthorized
	}
	existing, err := s.contentTypeRepo.GetByID(ctx, id)
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
	return s.contentTypeRepo.Delete(ctx, id)
}
