package services

import (
	"context"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
)

type ElementQueryService struct {
	elementRepo repositories.ElementRepositoryInterface
}

func NewElementQueryService(elementRepo repositories.ElementRepositoryInterface) ElementServiceInterface {
	return &ElementQueryService{
		elementRepo: elementRepo,
	}
}

func (s *ElementQueryService) List(ctx context.Context, projectID string, pageID ...string) ([]models.EditorElement, error) {
	return s.elementRepo.List(ctx, projectID, pageID...)
}

func (s *ElementQueryService) ListByPage(ctx context.Context, pageID string) ([]models.Element, error) {
	return s.elementRepo.ListByPage(ctx, pageID)
}

func (s *ElementQueryService) ListByPageIDs(ctx context.Context, projectID string, pageIDs []string) ([]models.EditorElement, error) {
	return s.elementRepo.ListByPageIDs(ctx, projectID, pageIDs)
}

func (s *ElementQueryService) ListByIDs(ctx context.Context, elementIDs []string) ([]models.Element, error) {
	return s.elementRepo.ListByIDs(ctx, elementIDs)
}
