package testutil

import (
	"context"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
)

type MockCustomElementRepository struct {
	GetCustomElementsFn       func(ctx context.Context, userID string, isPublic *bool) ([]models.CustomElement, error)
	GetCustomElementByIDFn    func(ctx context.Context, id string, userID string) (*models.CustomElement, error)
	CreateCustomElementFn     func(ctx context.Context, customElement *models.CustomElement) (*models.CustomElement, error)
	UpdateCustomElementFn     func(ctx context.Context, id string, userID string, updates map[string]any) (*models.CustomElement, error)
	DeleteCustomElementFn     func(ctx context.Context, id string, userID string) error
	GetPublicCustomElementsFn func(ctx context.Context, category *string, limit int, offset int) ([]models.CustomElement, error)
	DuplicateCustomElementFn  func(ctx context.Context, id string, userID string, newName string) (*models.CustomElement, error)
}

func (m *MockCustomElementRepository) GetCustomElements(ctx context.Context, userID string, isPublic *bool) ([]models.CustomElement, error) {
	if m.GetCustomElementsFn != nil {
		return m.GetCustomElementsFn(ctx, userID, isPublic)
	}
	return []models.CustomElement{}, nil
}

func (m *MockCustomElementRepository) GetCustomElementByID(ctx context.Context, id string, userID string) (*models.CustomElement, error) {
	if m.GetCustomElementByIDFn != nil {
		return m.GetCustomElementByIDFn(ctx, id, userID)
	}
	return nil, repositories.ErrCustomElementNotFound
}

func (m *MockCustomElementRepository) CreateCustomElement(ctx context.Context, customElement *models.CustomElement) (*models.CustomElement, error) {
	if m.CreateCustomElementFn != nil {
		return m.CreateCustomElementFn(ctx, customElement)
	}
	return customElement, nil
}

func (m *MockCustomElementRepository) UpdateCustomElement(ctx context.Context, id string, userID string, updates map[string]any) (*models.CustomElement, error) {
	if m.UpdateCustomElementFn != nil {
		return m.UpdateCustomElementFn(ctx, id, userID, updates)
	}
	return nil, repositories.ErrCustomElementNotFound
}

func (m *MockCustomElementRepository) DeleteCustomElement(ctx context.Context, id string, userID string) error {
	if m.DeleteCustomElementFn != nil {
		return m.DeleteCustomElementFn(ctx, id, userID)
	}
	return nil
}

func (m *MockCustomElementRepository) GetPublicCustomElements(ctx context.Context, category *string, limit int, offset int) ([]models.CustomElement, error) {
	if m.GetPublicCustomElementsFn != nil {
		return m.GetPublicCustomElementsFn(ctx, category, limit, offset)
	}
	return []models.CustomElement{}, nil
}

func (m *MockCustomElementRepository) DuplicateCustomElement(ctx context.Context, id string, userID string, newName string) (*models.CustomElement, error) {
	if m.DuplicateCustomElementFn != nil {
		return m.DuplicateCustomElementFn(ctx, id, userID, newName)
	}
	return nil, repositories.ErrCustomElementNotFound
}