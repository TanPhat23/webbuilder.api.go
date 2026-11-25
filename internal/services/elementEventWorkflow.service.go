package services

import (
	"context"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
)

type ElementEventWorkflowService struct {
	elementEventWorkflowRepo repositories.ElementEventWorkflowRepositoryInterface
	elementRepo              repositories.ElementRepositoryInterface
	projectRepo              repositories.ProjectRepositoryInterface
	pageRepo                 repositories.PageRepositoryInterface
}

func NewElementEventWorkflowService(
	elementEventWorkflowRepo repositories.ElementEventWorkflowRepositoryInterface,
	elementRepo repositories.ElementRepositoryInterface,
	projectRepo repositories.ProjectRepositoryInterface,
	pageRepo repositories.PageRepositoryInterface,
) *ElementEventWorkflowService {
	return &ElementEventWorkflowService{
		elementEventWorkflowRepo: elementEventWorkflowRepo,
		elementRepo:              elementRepo,
		projectRepo:              projectRepo,
		pageRepo:                 pageRepo,
	}
}

func (s *ElementEventWorkflowService) Create(ctx context.Context, eew *models.ElementEventWorkflow) (*models.ElementEventWorkflow, error) {
	return s.elementEventWorkflowRepo.Create(ctx, eew)
}

func (s *ElementEventWorkflowService) GetByID(ctx context.Context, id string) (*models.ElementEventWorkflow, error) {
	return s.elementEventWorkflowRepo.GetByID(ctx, id)
}

func (s *ElementEventWorkflowService) ListAll(ctx context.Context) ([]models.ElementEventWorkflow, error) {
	return s.elementEventWorkflowRepo.ListAll(ctx)
}

func (s *ElementEventWorkflowService) ListByElement(ctx context.Context, elementID string) ([]models.ElementEventWorkflow, error) {
	return s.elementEventWorkflowRepo.ListByElement(ctx, elementID)
}

func (s *ElementEventWorkflowService) ListByWorkflow(ctx context.Context, workflowID string) ([]models.ElementEventWorkflow, error) {
	return s.elementEventWorkflowRepo.ListByWorkflow(ctx, workflowID)
}

func (s *ElementEventWorkflowService) ListByEvent(ctx context.Context, eventName string) ([]models.ElementEventWorkflow, error) {
	return s.elementEventWorkflowRepo.ListByEvent(ctx, eventName)
}

func (s *ElementEventWorkflowService) ListByFilters(ctx context.Context, elementID, workflowID, eventName string) ([]models.ElementEventWorkflow, error) {
	return s.elementEventWorkflowRepo.ListByFilters(ctx, elementID, workflowID, eventName)
}

func (s *ElementEventWorkflowService) Update(ctx context.Context, id string, eew *models.ElementEventWorkflow) (*models.ElementEventWorkflow, error) {
	return s.elementEventWorkflowRepo.Update(ctx, id, eew)
}

func (s *ElementEventWorkflowService) Delete(ctx context.Context, id string) error {
	return s.elementEventWorkflowRepo.Delete(ctx, id)
}

func (s *ElementEventWorkflowService) DeleteByElement(ctx context.Context, elementID string) error {
	return s.elementEventWorkflowRepo.DeleteByElement(ctx, elementID)
}

func (s *ElementEventWorkflowService) DeleteByWorkflow(ctx context.Context, workflowID string) error {
	return s.elementEventWorkflowRepo.DeleteByWorkflow(ctx, workflowID)
}

func (s *ElementEventWorkflowService) ListByPage(ctx context.Context, pageID string) ([]models.ElementEventWorkflow, error) {
	return s.elementEventWorkflowRepo.ListByPage(ctx, pageID)
}

func (s *ElementEventWorkflowService) IsLinked(ctx context.Context, elementID, workflowID, eventName string) (bool, error) {
	return s.elementEventWorkflowRepo.IsLinked(ctx, elementID, workflowID, eventName)
}
