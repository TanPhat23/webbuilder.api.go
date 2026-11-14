package usecase

import (
	domainRepo "my-go-app/internal/realtime/domain/repository"
	domainUsecase "my-go-app/internal/realtime/domain/usecase"
	"my-go-app/internal/realtime/utils/logger"
)

// NewUseCase initializes all use cases and returns a central container.
func NewUseCase(repo *domainRepo.Repository, lg logger.Logger) *domainUsecase.UseCase {
	docUC := NewDocumentUseCase(repo)
	persistUC := NewPersistenceUsecase(repo, lg)
	elemUC := NewElementUseCase(repo, docUC, persistUC)
	workflowUC := NewWorkflowUseCase(repo.WorkflowRepo)
	elementWorkflowUC := NewElementWorkflowsUseCase(repo)

	return &domainUsecase.UseCase{
		Element:          elemUC,
		Document:         docUC,
		Persist:          persistUC,
		Workflow:         workflowUC,
		ElementWorkflows: elementWorkflowUC,
		Logic:            NewLogicUseCase(repo.LogicRepo),
	}
}
