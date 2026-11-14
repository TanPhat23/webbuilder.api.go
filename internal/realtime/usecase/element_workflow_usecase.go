package usecase

import (
	"context"
	domainEntity "my-go-app/internal/realtime/domain/entity"
	domainRepo "my-go-app/internal/realtime/domain/repository"
	domainUseCase "my-go-app/internal/realtime/domain/usecase"
)

type ElementWorkflowsUseCase struct {
	repo *domainRepo.Repository
}

func NewElementWorkflowsUseCase(repo *domainRepo.Repository) domainUseCase.ElementWorkflowsUseCase {
	return &ElementWorkflowsUseCase{
		repo: repo,
	}
}

func (uc *ElementWorkflowsUseCase) CreateElementWorkflow(ctx context.Context, projectID string, workflow *domainEntity.ElementEventWorkflow) error {
	return uc.repo.ElementWorkflowRepo.CreateElementWorkflow(ctx, projectID, workflow)
}

func (uc *ElementWorkflowsUseCase) GetElementWorkflowByElementID(ctx context.Context, projectID string, elementID string) (*domainEntity.ElementEventWorkflow, error) {
	return uc.repo.ElementWorkflowRepo.GetElementWorkflowByElementID(ctx, projectID, elementID)
}

func (uc *ElementWorkflowsUseCase) GetWorkflowsByProject(ctx context.Context, projectID string) ([]domainEntity.ElementEventWorkflow, error) {
	return uc.repo.ElementWorkflowRepo.GetElementWorkflowsByProjectID(ctx, projectID)
}

func (uc *ElementWorkflowsUseCase) GetWorkflowsByPage(ctx context.Context, projectID string, pageID string) ([]domainEntity.ElementEventWorkflow, error) {
	return uc.repo.ElementWorkflowRepo.GetElementWorkflowsByPageID(ctx, projectID, pageID)
}

func (uc *ElementWorkflowsUseCase) UpdateElementWorkflow(ctx context.Context, projectID string, workflow *domainEntity.ElementEventWorkflow) error {
	return uc.repo.ElementWorkflowRepo.UpdateElementWorkflow(ctx, projectID, workflow)
}

func (uc *ElementWorkflowsUseCase) RemoveElementWorkflow(ctx context.Context, projectID string, elementID string) error {
	return uc.repo.ElementWorkflowRepo.DeleteElementWorkflow(ctx, projectID, elementID)
}
