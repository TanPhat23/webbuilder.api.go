package usecase

import (
	"context"
	domainEntity "my-go-app/internal/realtime/domain/entity"
	domainRepo "my-go-app/internal/realtime/domain/repository"
	domainUseCase "my-go-app/internal/realtime/domain/usecase"
)

type WorkflowUseCase struct {
	repo domainRepo.WorkflowRepository
}

func NewWorkflowUseCase(repo domainRepo.WorkflowRepository) domainUseCase.WorkflowUseCase {
	return &WorkflowUseCase{
		repo: repo,
	}
}

func (uc *WorkflowUseCase) CreateWorkflow(ctx context.Context, projectID string, workflow *domainEntity.EventWorkflow) error {
	return uc.repo.CreateWorkflow(ctx, projectID, workflow)
}

func (uc *WorkflowUseCase) UpdateWorkflow(ctx context.Context, projectID string, workflow *domainEntity.EventWorkflow) error {
	return uc.repo.UpdateWorkflow(ctx, projectID, workflow)
}

func (uc *WorkflowUseCase) RemoveWorkflow(ctx context.Context, projectID string, workflowID string) error {
	return uc.repo.DeleteWorkflow(ctx, projectID, workflowID)
}

func (uc *WorkflowUseCase) GetWorkflowByID(ctx context.Context, projectID string, workflowID string) (*domainEntity.EventWorkflow, error) {
	return uc.repo.GetWorkflowByID(ctx, projectID, workflowID)
}

func (uc *WorkflowUseCase) GetWorkflowsByProject(ctx context.Context, projectID string) ([]domainEntity.EventWorkflow, error) {
	return uc.repo.GetWorkflowsByProjectID(ctx, projectID)
}
