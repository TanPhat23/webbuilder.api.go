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

func (s *CustomElementTypeService) List(ctx context.Context) ([]models.CustomElementType, error) {
	return s.customElementTypeRepo.List(ctx)
}

func (s *CustomElementTypeService) GetByID(ctx context.Context, id string) (*models.CustomElementType, error) {
	return s.customElementTypeRepo.GetByID(ctx, id)
}

func (s *CustomElementTypeService) GetByName(ctx context.Context, name string) (*models.CustomElementType, error) {
	return s.customElementTypeRepo.GetByName(ctx, name)
}

func (s *CustomElementTypeService) Create(ctx context.Context, ceType *models.CustomElementType) (*models.CustomElementType, error) {
	return s.customElementTypeRepo.Create(ctx, ceType)
}

func (s *CustomElementTypeService) Update(ctx context.Context, id string, updates map[string]any) (*models.CustomElementType, error) {
	return s.customElementTypeRepo.Update(ctx, id, updates)
}

func (s *CustomElementTypeService) Delete(ctx context.Context, id string) error {
	return s.customElementTypeRepo.Delete(ctx, id)
}
