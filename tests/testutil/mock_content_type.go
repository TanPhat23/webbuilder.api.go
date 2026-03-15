package testutil

import (
	"context"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
)

type MockContentTypeRepository struct {
	GetContentTypesFn    func(ctx context.Context) ([]models.ContentType, error)
	GetContentTypeByIDFn func(ctx context.Context, id string) (*models.ContentType, error)
	CreateContentTypeFn  func(ctx context.Context, ct *models.ContentType) (*models.ContentType, error)
	UpdateContentTypeFn  func(ctx context.Context, id string, updates map[string]any) (*models.ContentType, error)
	DeleteContentTypeFn  func(ctx context.Context, id string) error
}

func (m *MockContentTypeRepository) GetContentTypes(ctx context.Context) ([]models.ContentType, error) {
	if m.GetContentTypesFn != nil {
		return m.GetContentTypesFn(ctx)
	}
	return []models.ContentType{}, nil
}

func (m *MockContentTypeRepository) GetContentTypeByID(ctx context.Context, id string) (*models.ContentType, error) {
	if m.GetContentTypeByIDFn != nil {
		return m.GetContentTypeByIDFn(ctx, id)
	}
	return nil, repositories.ErrContentTypeNotFound
}

func (m *MockContentTypeRepository) CreateContentType(ctx context.Context, ct *models.ContentType) (*models.ContentType, error) {
	if m.CreateContentTypeFn != nil {
		return m.CreateContentTypeFn(ctx, ct)
	}
	return ct, nil
}

func (m *MockContentTypeRepository) UpdateContentType(ctx context.Context, id string, updates map[string]any) (*models.ContentType, error) {
	if m.UpdateContentTypeFn != nil {
		return m.UpdateContentTypeFn(ctx, id, updates)
	}
	return nil, repositories.ErrContentTypeNotFound
}

func (m *MockContentTypeRepository) DeleteContentType(ctx context.Context, id string) error {
	if m.DeleteContentTypeFn != nil {
		return m.DeleteContentTypeFn(ctx, id)
	}
	return nil
}