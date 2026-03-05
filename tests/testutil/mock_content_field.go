package testutil

import (
	"context"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
)

type MockContentFieldRepository struct {
	GetContentFieldsByContentTypeFn func(ctx context.Context, contentTypeID string) ([]models.ContentField, error)
	GetContentFieldByIDFn           func(ctx context.Context, id string) (*models.ContentField, error)
	CreateContentFieldFn            func(ctx context.Context, cf *models.ContentField) (*models.ContentField, error)
	UpdateContentFieldFn            func(ctx context.Context, id string, updates map[string]any) (*models.ContentField, error)
	DeleteContentFieldFn            func(ctx context.Context, id string) error
}

func (m *MockContentFieldRepository) GetContentFieldsByContentType(ctx context.Context, contentTypeID string) ([]models.ContentField, error) {
	if m.GetContentFieldsByContentTypeFn != nil {
		return m.GetContentFieldsByContentTypeFn(ctx, contentTypeID)
	}
	return []models.ContentField{}, nil
}

func (m *MockContentFieldRepository) GetContentFieldByID(ctx context.Context, id string) (*models.ContentField, error) {
	if m.GetContentFieldByIDFn != nil {
		return m.GetContentFieldByIDFn(ctx, id)
	}
	return nil, repositories.ErrContentFieldNotFound
}

func (m *MockContentFieldRepository) CreateContentField(ctx context.Context, cf *models.ContentField) (*models.ContentField, error) {
	if m.CreateContentFieldFn != nil {
		return m.CreateContentFieldFn(ctx, cf)
	}
	return cf, nil
}

func (m *MockContentFieldRepository) UpdateContentField(ctx context.Context, id string, updates map[string]any) (*models.ContentField, error) {
	if m.UpdateContentFieldFn != nil {
		return m.UpdateContentFieldFn(ctx, id, updates)
	}
	return nil, repositories.ErrContentFieldNotFound
}

func (m *MockContentFieldRepository) DeleteContentField(ctx context.Context, id string) error {
	if m.DeleteContentFieldFn != nil {
		return m.DeleteContentFieldFn(ctx, id)
	}
	return nil
}