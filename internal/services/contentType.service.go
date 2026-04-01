package services

import (
	"context"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
)

type ContentTypeService struct {
	contentTypeRepo repositories.ContentTypeRepositoryInterface
}

func NewContentTypeService(contentTypeRepo repositories.ContentTypeRepositoryInterface) *ContentTypeService {
	return &ContentTypeService{
		contentTypeRepo: contentTypeRepo,
	}
}

func (s *ContentTypeService) GetContentTypes(ctx context.Context) ([]models.ContentType, error) {
	return s.contentTypeRepo.GetContentTypes(ctx)
}

func (s *ContentTypeService) GetContentTypeByID(ctx context.Context, id string) (*models.ContentType, error) {
	return s.contentTypeRepo.GetContentTypeByID(ctx, id)
}

func (s *ContentTypeService) CreateContentType(ctx context.Context, contentType *models.ContentType) (*models.ContentType, error) {
	return s.contentTypeRepo.CreateContentType(ctx, contentType)
}

func (s *ContentTypeService) UpdateContentType(ctx context.Context, id string, contentType *models.ContentType) (*models.ContentType, error) {
	updates := map[string]any{
		"name":        contentType.Name,
		"description": contentType.Description,
	}
	return s.contentTypeRepo.UpdateContentType(ctx, id, updates)
}

func (s *ContentTypeService) DeleteContentType(ctx context.Context, id string) error {
	return s.contentTypeRepo.DeleteContentType(ctx, id)
}