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

func (s *ElementEventWorkflowService) CreateElementEventWorkflow(ctx context.Context, eew *models.ElementEventWorkflow) (*models.ElementEventWorkflow, error) {
	return s.elementEventWorkflowRepo.CreateElementEventWorkflow(ctx, eew)
}

func (s *ElementEventWorkflowService) GetElementEventWorkflowByID(ctx context.Context, id string) (*models.ElementEventWorkflow, error) {
	return s.elementEventWorkflowRepo.GetElementEventWorkflowByID(ctx, id)
}

func (s *ElementEventWorkflowService) GetAllElementEventWorkflows(ctx context.Context) ([]models.ElementEventWorkflow, error) {
	return s.elementEventWorkflowRepo.GetAllElementEventWorkflows(ctx)
}

func (s *ElementEventWorkflowService) GetElementEventWorkflowsByElementID(ctx context.Context, elementID string) ([]models.ElementEventWorkflow, error) {
	return s.elementEventWorkflowRepo.GetElementEventWorkflowsByElementID(ctx, elementID)
}

func (s *ElementEventWorkflowService) GetElementEventWorkflowsByWorkflowID(ctx context.Context, workflowID string) ([]models.ElementEventWorkflow, error) {
	return s.elementEventWorkflowRepo.GetElementEventWorkflowsByWorkflowID(ctx, workflowID)
}

func (s *ElementEventWorkflowService) GetElementEventWorkflowsByEventName(ctx context.Context, eventName string) ([]models.ElementEventWorkflow, error) {
	return s.elementEventWorkflowRepo.GetElementEventWorkflowsByEventName(ctx, eventName)
}

func (s *ElementEventWorkflowService) GetElementEventWorkflowsByFilters(ctx context.Context, elementID, workflowID, eventName string) ([]models.ElementEventWorkflow, error) {
	return s.elementEventWorkflowRepo.GetElementEventWorkflowsByFilters(ctx, elementID, workflowID, eventName)
}

func (s *ElementEventWorkflowService) UpdateElementEventWorkflow(ctx context.Context, id string, eew *models.ElementEventWorkflow) (*models.ElementEventWorkflow, error) {
	return s.elementEventWorkflowRepo.UpdateElementEventWorkflow(ctx, id, eew)
}

func (s *ElementEventWorkflowService) DeleteElementEventWorkflow(ctx context.Context, id string) error {
	return s.elementEventWorkflowRepo.DeleteElementEventWorkflow(ctx, id)
}

func (s *ElementEventWorkflowService) DeleteElementEventWorkflowsByElementID(ctx context.Context, elementID string) error {
	return s.elementEventWorkflowRepo.DeleteElementEventWorkflowsByElementID(ctx, elementID)
}

func (s *ElementEventWorkflowService) DeleteElementEventWorkflowsByWorkflowID(ctx context.Context, workflowID string) error {
	return s.elementEventWorkflowRepo.DeleteElementEventWorkflowsByWorkflowID(ctx, workflowID)
}

func (s *ElementEventWorkflowService) GetElementEventWorkflowsByPageID(ctx context.Context, pageID string) ([]models.ElementEventWorkflow, error) {
	return s.elementEventWorkflowRepo.GetElementEventWorkflowsByPageID(ctx, pageID)
}

func (s *ElementEventWorkflowService) CheckIfWorkflowLinkedToElement(ctx context.Context, elementID, workflowID, eventName string) (bool, error) {
	return s.elementEventWorkflowRepo.CheckIfWorkflowLinkedToElement(ctx, elementID, workflowID, eventName)
}
