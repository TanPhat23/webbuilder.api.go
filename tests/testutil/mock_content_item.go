package testutil

import (
	"context"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
)

type MockContentItemRepository struct {
	GetContentItemsByContentTypeFn func(ctx context.Context, contentTypeID string) ([]models.ContentItem, error)
	GetContentItemByIDFn           func(ctx context.Context, id string) (*models.ContentItem, error)
	GetContentItemBySlugFn         func(ctx context.Context, contentTypeID, slug string) (*models.ContentItem, error)
	GetPublicContentItemsFn        func(ctx context.Context, contentTypeID string, limit int, sortBy, sortOrder string) ([]models.ContentItem, error)
	CreateContentItemFn            func(ctx context.Context, ci *models.ContentItem, fieldValues []models.ContentFieldValue) (*models.ContentItem, error)
	UpdateContentItemFn            func(ctx context.Context, id string, updates map[string]any, fieldValues []models.ContentFieldValue) (*models.ContentItem, error)
	DeleteContentItemFn            func(ctx context.Context, id string) error
}

func (m *MockContentItemRepository) GetContentItemsByContentType(ctx context.Context, contentTypeID string) ([]models.ContentItem, error) {
	if m.GetContentItemsByContentTypeFn != nil {
		return m.GetContentItemsByContentTypeFn(ctx, contentTypeID)
	}
	return []models.ContentItem{}, nil
}

func (m *MockContentItemRepository) GetContentItemByID(ctx context.Context, id string) (*models.ContentItem, error) {
	if m.GetContentItemByIDFn != nil {
		return m.GetContentItemByIDFn(ctx, id)
	}
	return nil, repositories.ErrContentItemNotFound
}

func (m *MockContentItemRepository) GetContentItemBySlug(ctx context.Context, contentTypeID, slug string) (*models.ContentItem, error) {
	if m.GetContentItemBySlugFn != nil {
		return m.GetContentItemBySlugFn(ctx, contentTypeID, slug)
	}
	return nil, repositories.ErrContentItemNotFound
}

func (m *MockContentItemRepository) GetPublicContentItems(ctx context.Context, contentTypeID string, limit int, sortBy, sortOrder string) ([]models.ContentItem, error) {
	if m.GetPublicContentItemsFn != nil {
		return m.GetPublicContentItemsFn(ctx, contentTypeID, limit, sortBy, sortOrder)
	}
	return []models.ContentItem{}, nil
}

func (m *MockContentItemRepository) CreateContentItem(ctx context.Context, ci *models.ContentItem, fieldValues []models.ContentFieldValue) (*models.ContentItem, error) {
	if m.CreateContentItemFn != nil {
		return m.CreateContentItemFn(ctx, ci, fieldValues)
	}
	return ci, nil
}

func (m *MockContentItemRepository) UpdateContentItem(ctx context.Context, id string, updates map[string]any, fieldValues []models.ContentFieldValue) (*models.ContentItem, error) {
	if m.UpdateContentItemFn != nil {
		return m.UpdateContentItemFn(ctx, id, updates, fieldValues)
	}
	return nil, repositories.ErrContentItemNotFound
}

func (m *MockContentItemRepository) DeleteContentItem(ctx context.Context, id string) error {
	if m.DeleteContentItemFn != nil {
		return m.DeleteContentItemFn(ctx, id)
	}
	return nil
}