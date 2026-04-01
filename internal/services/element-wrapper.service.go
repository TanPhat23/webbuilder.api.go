package services

import (
	"context"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
)

type ElementWrapperService struct {
	elementRepo repositories.ElementRepositoryInterface
}

func NewElementWrapperService(elementRepo repositories.ElementRepositoryInterface) ElementServiceInterface {
	return &ElementWrapperService{
		elementRepo: elementRepo,
	}
}

func (s *ElementWrapperService) GetElements(ctx context.Context, projectID string, pageID ...string) ([]models.EditorElement, error) {
	return s.elementRepo.GetElements(ctx, projectID, pageID...)
}

func (s *ElementWrapperService) GetElementsByPageID(ctx context.Context, pageID string) ([]models.Element, error) {
	return s.elementRepo.GetElementsByPageID(ctx, pageID)
}

func (s *ElementWrapperService) GetElementsByPageIds(ctx context.Context, pageIDs []string) ([]models.EditorElement, error) {
	return s.elementRepo.GetElementsByPageIds(ctx, pageIDs)
}

func (s *ElementWrapperService) GetElementsByIDs(ctx context.Context, elementIDs []string) ([]models.Element, error) {
	return s.elementRepo.GetElementsByIDs(ctx, elementIDs)
}
