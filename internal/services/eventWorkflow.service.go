package services

import (
	"context"
	"errors"
	"fmt"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
)

type EventWorkflowService struct {
	eventWorkflowRepo repositories.EventWorkflowRepositoryInterface
	projectRepo       repositories.ProjectRepositoryInterface
}

func NewEventWorkflowService(
	eventWorkflowRepo repositories.EventWorkflowRepositoryInterface,
	projectRepo repositories.ProjectRepositoryInterface,
) *EventWorkflowService {
	return &EventWorkflowService{
		eventWorkflowRepo: eventWorkflowRepo,
		projectRepo:       projectRepo,
	}
}

func (s *EventWorkflowService) Create(ctx context.Context, workflow *models.EventWorkflow, userID string) (*models.EventWorkflow, error) {
	if workflow == nil {
		return nil, errors.New("workflow cannot be nil")
	}
	if workflow.ProjectId == "" {
		return nil, errors.New("projectId is required")
	}
	if workflow.Name == "" {
		return nil, errors.New("workflow name is required")
	}

	if err := ensureProjectVisible(ctx, s.projectRepo, workflow.ProjectId, userID); err != nil {
		return nil, fmt.Errorf("failed to verify project: %w", err)
	}

	if workflow.CanvasData == nil {
		workflow.CanvasData = []byte("{}")
	}
	if workflow.Handlers == nil {
		workflow.Handlers = []byte("[]")
	}

	return s.eventWorkflowRepo.Create(ctx, workflow)
}

func (s *EventWorkflowService) GetByID(ctx context.Context, id string) (*models.EventWorkflow, error) {
	if id == "" {
		return nil, errors.New("workflow id is required")
	}
	return s.eventWorkflowRepo.GetByID(ctx, id)
}

func (s *EventWorkflowService) ListByProject(ctx context.Context, projectID string) ([]models.EventWorkflow, error) {
	if projectID == "" {
		return nil, errors.New("projectId is required")
	}

	project, err := s.projectRepo.GetPublicByID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify project: %w", err)
	}
	if project == nil {
		return nil, errors.New("project does not exist")
	}

	return s.eventWorkflowRepo.ListByProject(ctx, projectID)
}

func (s *EventWorkflowService) ListByProjectWithElements(ctx context.Context, projectID string) ([]models.EventWorkflow, error) {
	if projectID == "" {
		return nil, errors.New("projectId is required")
	}

	project, err := s.projectRepo.GetPublicByID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify project: %w", err)
	}
	if project == nil {
		return nil, errors.New("project does not exist")
	}

	return s.eventWorkflowRepo.ListByProjectWithElements(ctx, projectID)
}

func (s *EventWorkflowService) ListEnabledByProject(ctx context.Context, projectID string) ([]models.EventWorkflow, error) {
	if projectID == "" {
		return nil, errors.New("projectId is required")
	}

	project, err := s.projectRepo.GetPublicByID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify project: %w", err)
	}
	if project == nil {
		return nil, errors.New("project does not exist")
	}

	return s.eventWorkflowRepo.ListEnabledByProject(ctx, projectID)
}

func (s *EventWorkflowService) ListByName(ctx context.Context, projectID, name string) ([]models.EventWorkflow, error) {
	if projectID == "" {
		return nil, errors.New("projectId is required")
	}
	if name == "" {
		return nil, errors.New("workflow name is required")
	}

	project, err := s.projectRepo.GetPublicByID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify project: %w", err)
	}
	if project == nil {
		return nil, errors.New("project does not exist")
	}

	return s.eventWorkflowRepo.ListByName(ctx, projectID, name)
}

func (s *EventWorkflowService) Update(ctx context.Context, id string, workflow *models.EventWorkflow) (*models.EventWorkflow, error) {
	if id == "" {
		return nil, errors.New("workflow id is required")
	}
	if workflow == nil {
		return nil, errors.New("workflow cannot be nil")
	}
	if workflow.Name == "" {
		return nil, errors.New("workflow name is required")
	}

	existing, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("workflow does not exist")
	}

	if workflow.ProjectId != "" && workflow.ProjectId != existing.ProjectId {
		return nil, errors.New("projectId cannot be changed")
	}

	if workflow.Description == nil {
		workflow.Description = existing.Description
	}
	if len(workflow.CanvasData) == 0 {
		workflow.CanvasData = existing.CanvasData
	}
	if len(workflow.Handlers) == 0 {
		workflow.Handlers = existing.Handlers
	}
	if workflow.ProjectId == "" {
		workflow.ProjectId = existing.ProjectId
	}
	if workflow.CreatedAt.IsZero() {
		workflow.CreatedAt = existing.CreatedAt
	}

	return s.eventWorkflowRepo.Update(ctx, id, workflow)
}

func (s *EventWorkflowService) SetEnabled(ctx context.Context, id string, enabled bool) error {
	if id == "" {
		return errors.New("workflow id is required")
	}

	existing, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("workflow does not exist")
	}

	return s.eventWorkflowRepo.SetEnabled(ctx, id, enabled)
}

func (s *EventWorkflowService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("workflow id is required")
	}

	existing, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("workflow does not exist")
	}

	return s.eventWorkflowRepo.Delete(ctx, id)
}

func (s *EventWorkflowService) DeleteByProject(ctx context.Context, projectID string) error {
	if projectID == "" {
		return errors.New("projectId is required")
	}

	project, err := s.projectRepo.GetPublicByID(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to verify project: %w", err)
	}
	if project == nil {
		return errors.New("project does not exist")
	}

	return s.eventWorkflowRepo.DeleteByProject(ctx, projectID)
}

func (s *EventWorkflowService) CountByProject(ctx context.Context, projectID string) (int64, error) {
	if projectID == "" {
		return 0, errors.New("projectId is required")
	}

	project, err := s.projectRepo.GetPublicByID(ctx, projectID)
	if err != nil {
		return 0, fmt.Errorf("failed to verify project: %w", err)
	}
	if project == nil {
		return 0, errors.New("project does not exist")
	}

	return s.eventWorkflowRepo.CountByProject(ctx, projectID)
}

func (s *EventWorkflowService) ExistsByName(ctx context.Context, projectID, name, excludeID string) (bool, error) {
	if projectID == "" {
		return false, errors.New("projectId is required")
	}
	if name == "" {
		return false, errors.New("workflow name is required")
	}

	project, err := s.projectRepo.GetPublicByID(ctx, projectID)
	if err != nil {
		return false, fmt.Errorf("failed to verify project: %w", err)
	}
	if project == nil {
		return false, errors.New("project does not exist")
	}

	return s.eventWorkflowRepo.ExistsByName(ctx, projectID, name, excludeID)
}

func (s *EventWorkflowService) ListWithFilters(ctx context.Context, projectID, userID string, enabled *bool, searchName string) ([]models.EventWorkflow, error) {
	if projectID == "" {
		return nil, errors.New("projectId is required")
	}

	if err := ensureProjectVisible(ctx, s.projectRepo, projectID, userID); err != nil {
		return nil, fmt.Errorf("failed to verify project: %w", err)
	}

	return s.eventWorkflowRepo.ListWithFilters(ctx, projectID, enabled, searchName)
}
