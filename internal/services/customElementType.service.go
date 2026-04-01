package services

import (
	"context"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
)

type CustomElementTypeService struct {
	customElementTypeRepo repositories.CustomElementTypeRepositoryInterface
}

func NewCustomElementTypeService(customElementTypeRepo repositories.CustomElementTypeRepositoryInterface) *CustomElementTypeService {
	return &CustomElementTypeService{
		customElementTypeRepo: customElementTypeRepo,
	}
}

func (s *CustomElementTypeService) GetCustomElementTypes(ctx context.Context) ([]models.CustomElementType, error) {
	return s.customElementTypeRepo.GetCustomElementTypes(ctx)
}

func (s *CustomElementTypeService) GetCustomElementTypeByID(ctx context.Context, id string) (*models.CustomElementType, error) {
	return s.customElementTypeRepo.GetCustomElementTypeByID(ctx, id)
}

func (s *CustomElementTypeService) GetCustomElementTypeByName(ctx context.Context, name string) (*models.CustomElementType, error) {
	return s.customElementTypeRepo.GetCustomElementTypeByName(ctx, name)
}

func (s *CustomElementTypeService) CreateCustomElementType(ctx context.Context, ceType *models.CustomElementType) (*models.CustomElementType, error) {
	return s.customElementTypeRepo.CreateCustomElementType(ctx, ceType)
}

func (s *CustomElementTypeService) UpdateCustomElementType(ctx context.Context, id string, updates map[string]any) (*models.CustomElementType, error) {
	return s.customElementTypeRepo.UpdateCustomElementType(ctx, id, updates)
}

func (s *CustomElementTypeService) DeleteCustomElementType(ctx context.Context, id string) error {
	return s.customElementTypeRepo.DeleteCustomElementType(ctx, id)
}