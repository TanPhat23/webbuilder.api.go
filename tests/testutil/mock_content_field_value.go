package testutil

import (
	"context"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
)

type MockContentFieldValueRepository struct {
	GetContentFieldValuesByContentItemFn func(ctx context.Context, contentItemID string) ([]models.ContentFieldValue, error)
	GetContentFieldValueByIDFn           func(ctx context.Context, id string) (*models.ContentFieldValue, error)
	CreateContentFieldValueFn            func(ctx context.Context, cfv *models.ContentFieldValue) (*models.ContentFieldValue, error)
	UpdateContentFieldValueFn            func(ctx context.Context, id string, value *string) (*models.ContentFieldValue, error)
	DeleteContentFieldValueFn            func(ctx context.Context, id string) error
}

func (m *MockContentFieldValueRepository) GetContentFieldValuesByContentItem(ctx context.Context, contentItemID string) ([]models.ContentFieldValue, error) {
	if m.GetContentFieldValuesByContentItemFn != nil {
		return m.GetContentFieldValuesByContentItemFn(ctx, contentItemID)
	}
	return []models.ContentFieldValue{}, nil
}

func (m *MockContentFieldValueRepository) GetContentFieldValueByID(ctx context.Context, id string) (*models.ContentFieldValue, error) {
	if m.GetContentFieldValueByIDFn != nil {
		return m.GetContentFieldValueByIDFn(ctx, id)
	}
	return nil, repositories.ErrContentFieldValueNotFound
}

func (m *MockContentFieldValueRepository) CreateContentFieldValue(ctx context.Context, cfv *models.ContentFieldValue) (*models.ContentFieldValue, error) {
	if m.CreateContentFieldValueFn != nil {
		return m.CreateContentFieldValueFn(ctx, cfv)
	}
	return cfv, nil
}

func (m *MockContentFieldValueRepository) UpdateContentFieldValue(ctx context.Context, id string, value *string) (*models.ContentFieldValue, error) {
	if m.UpdateContentFieldValueFn != nil {
		return m.UpdateContentFieldValueFn(ctx, id, value)
	}
	return nil, repositories.ErrContentFieldValueNotFound
}

func (m *MockContentFieldValueRepository) DeleteContentFieldValue(ctx context.Context, id string) error {
	if m.DeleteContentFieldValueFn != nil {
		return m.DeleteContentFieldValueFn(ctx, id)
	}
	return nil
}