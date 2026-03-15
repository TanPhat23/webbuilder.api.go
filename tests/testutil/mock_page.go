package testutil

import (
	"context"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
)

type MockPageRepository struct {
	GetPagesByProjectIDFn   func(ctx context.Context, projectID string) ([]models.Page, error)
	GetPageByIDFn           func(ctx context.Context, pageID, projectID string) (*models.Page, error)
	CreatePageFn            func(ctx context.Context, page *models.Page) error
	UpdatePageFn            func(ctx context.Context, page *models.Page) error
	UpdatePageFieldsFn      func(ctx context.Context, pageID string, updates map[string]any) error
	DeletePageFn            func(ctx context.Context, pageID string) error
	DeletePageByProjectIDFn func(ctx context.Context, pageID, projectID, userID string) error
}

func (m *MockPageRepository) GetPagesByProjectID(ctx context.Context, projectID string) ([]models.Page, error) {
	if m.GetPagesByProjectIDFn != nil {
		return m.GetPagesByProjectIDFn(ctx, projectID)
	}
	return []models.Page{}, nil
}

func (m *MockPageRepository) GetPageByID(ctx context.Context, pageID, projectID string) (*models.Page, error) {
	if m.GetPageByIDFn != nil {
		return m.GetPageByIDFn(ctx, pageID, projectID)
	}
	return nil, repositories.ErrPageNotFound
}

func (m *MockPageRepository) CreatePage(ctx context.Context, page *models.Page) error {
	if m.CreatePageFn != nil {
		return m.CreatePageFn(ctx, page)
	}
	return nil
}

func (m *MockPageRepository) UpdatePage(ctx context.Context, page *models.Page) error {
	if m.UpdatePageFn != nil {
		return m.UpdatePageFn(ctx, page)
	}
	return nil
}

func (m *MockPageRepository) UpdatePageFields(ctx context.Context, pageID string, updates map[string]any) error {
	if m.UpdatePageFieldsFn != nil {
		return m.UpdatePageFieldsFn(ctx, pageID, updates)
	}
	return nil
}

func (m *MockPageRepository) DeletePage(ctx context.Context, pageID string) error {
	if m.DeletePageFn != nil {
		return m.DeletePageFn(ctx, pageID)
	}
	return nil
}

func (m *MockPageRepository) DeletePageByProjectID(ctx context.Context, pageID, projectID, userID string) error {
	if m.DeletePageByProjectIDFn != nil {
		return m.DeletePageByProjectIDFn(ctx, pageID, projectID, userID)
	}
	return nil
}