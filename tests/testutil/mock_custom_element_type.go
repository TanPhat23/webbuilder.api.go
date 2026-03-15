package testutil

import (
	"context"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
)

type MockCustomElementTypeRepository struct {
	GetCustomElementTypesFn      func(ctx context.Context) ([]models.CustomElementType, error)
	GetCustomElementTypeByIDFn   func(ctx context.Context, id string) (*models.CustomElementType, error)
	GetCustomElementTypeByNameFn func(ctx context.Context, name string) (*models.CustomElementType, error)
	CreateCustomElementTypeFn    func(ctx context.Context, customElementType *models.CustomElementType) (*models.CustomElementType, error)
	UpdateCustomElementTypeFn    func(ctx context.Context, id string, updates map[string]any) (*models.CustomElementType, error)
	DeleteCustomElementTypeFn    func(ctx context.Context, id string) error
}

func (m *MockCustomElementTypeRepository) GetCustomElementTypes(ctx context.Context) ([]models.CustomElementType, error) {
	if m.GetCustomElementTypesFn != nil {
		return m.GetCustomElementTypesFn(ctx)
	}
	return []models.CustomElementType{}, nil
}

func (m *MockCustomElementTypeRepository) GetCustomElementTypeByID(ctx context.Context, id string) (*models.CustomElementType, error) {
	if m.GetCustomElementTypeByIDFn != nil {
		return m.GetCustomElementTypeByIDFn(ctx, id)
	}
	return nil, repositories.ErrCustomElementTypeNotFound
}

func (m *MockCustomElementTypeRepository) GetCustomElementTypeByName(ctx context.Context, name string) (*models.CustomElementType, error) {
	if m.GetCustomElementTypeByNameFn != nil {
		return m.GetCustomElementTypeByNameFn(ctx, name)
	}
	return nil, repositories.ErrCustomElementTypeNotFound
}

func (m *MockCustomElementTypeRepository) CreateCustomElementType(ctx context.Context, customElementType *models.CustomElementType) (*models.CustomElementType, error) {
	if m.CreateCustomElementTypeFn != nil {
		return m.CreateCustomElementTypeFn(ctx, customElementType)
	}
	return customElementType, nil
}

func (m *MockCustomElementTypeRepository) UpdateCustomElementType(ctx context.Context, id string, updates map[string]any) (*models.CustomElementType, error) {
	if m.UpdateCustomElementTypeFn != nil {
		return m.UpdateCustomElementTypeFn(ctx, id, updates)
	}
	return nil, repositories.ErrCustomElementTypeNotFound
}

func (m *MockCustomElementTypeRepository) DeleteCustomElementType(ctx context.Context, id string) error {
	if m.DeleteCustomElementTypeFn != nil {
		return m.DeleteCustomElementTypeFn(ctx, id)
	}
	return nil
}