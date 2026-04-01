package services

import (
	"context"
	"errors"
	"fmt"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
)

type ContentItemService struct {
	contentItemRepo repositories.ContentItemRepositoryInterface
}

func NewContentItemService(contentItemRepo repositories.ContentItemRepositoryInterface) *ContentItemService {
	return &ContentItemService{
		contentItemRepo: contentItemRepo,
	}
}

func (s *ContentItemService) GetContentItemsByContentType(ctx context.Context, contentTypeID string) ([]models.ContentItem, error) {
	if contentTypeID == "" {
		return nil, errors.New("contentTypeId is required")
	}

	return s.contentItemRepo.GetContentItemsByContentType(ctx, contentTypeID)
}

func (s *ContentItemService) GetContentItemByID(ctx context.Context, id string) (*models.ContentItem, error) {
	if id == "" {
		return nil, errors.New("content item id is required")
	}

	return s.contentItemRepo.GetContentItemByID(ctx, id)
}

func (s *ContentItemService) GetContentItemBySlug(ctx context.Context, contentTypeID, slug string) (*models.ContentItem, error) {
	if contentTypeID == "" {
		return nil, errors.New("contentTypeId is required")
	}
	if slug == "" {
		return nil, errors.New("slug is required")
	}

	return s.contentItemRepo.GetContentItemBySlug(ctx, contentTypeID, slug)
}

func (s *ContentItemService) GetPublicContentItems(ctx context.Context, contentTypeID string, limit int, sortBy, sortOrder string) ([]models.ContentItem, error) {
	if contentTypeID == "" {
		return nil, errors.New("contentTypeId is required")
	}
	if limit <= 0 {
		return nil, errors.New("limit must be greater than zero")
	}
	if sortBy == "" {
		return nil, errors.New("sortBy is required")
	}
	if sortOrder == "" {
		return nil, errors.New("sortOrder is required")
	}

	return s.contentItemRepo.GetPublicContentItems(ctx, contentTypeID, limit, sortBy, sortOrder)
}

func (s *ContentItemService) CreateContentItem(ctx context.Context, item *models.ContentItem, fieldValues []models.ContentFieldValue) (*models.ContentItem, error) {
	if item == nil {
		return nil, errors.New("content item cannot be nil")
	}
	if item.ContentTypeId == "" {
		return nil, errors.New("contentTypeId is required")
	}
	if item.Slug == "" {
		return nil, errors.New("slug is required")
	}
	if item.Title == "" {
		return nil, errors.New("title is required")
	}

	if fieldValues == nil {
		fieldValues = []models.ContentFieldValue{}
	}
	item.FieldValues = nil

	created, err := s.contentItemRepo.CreateContentItem(ctx, item, fieldValues)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *ContentItemService) UpdateContentItem(ctx context.Context, id string, updates map[string]any, fieldValues []models.ContentFieldValue) (*models.ContentItem, error) {
	if id == "" {
		return nil, errors.New("content item id is required")
	}
	if len(updates) == 0 && len(fieldValues) == 0 {
		return nil, errors.New("no updates provided")
	}

	if updates == nil {
		updates = map[string]any{}
	}
	if fieldValues == nil {
		fieldValues = []models.ContentFieldValue{}
	}

	if slug, ok := updates["Slug"].(string); ok && slug == "" {
		return nil, errors.New("slug cannot be empty")
	}
	if title, ok := updates["Title"].(string); ok && title == "" {
		return nil, errors.New("title cannot be empty")
	}

	updated, err := s.contentItemRepo.UpdateContentItem(ctx, id, updates, fieldValues)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *ContentItemService) DeleteContentItem(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("content item id is required")
	}

	return s.contentItemRepo.DeleteContentItem(ctx, id)
}

func (s *ContentItemService) ValidatePublicSort(sortBy, sortOrder string) (string, string, error) {
	if sortBy == "" {
		return "", "", errors.New("sortBy is required")
	}
	if sortOrder == "" {
		return "", "", errors.New("sortOrder is required")
	}

	validSortOrders := map[string]bool{
		"asc":  true,
		"desc": true,
	}
	if !validSortOrders[sortOrder] {
		return "", "", fmt.Errorf("invalid sortOrder: %s", sortOrder)
	}

	return sortBy, sortOrder, nil
}