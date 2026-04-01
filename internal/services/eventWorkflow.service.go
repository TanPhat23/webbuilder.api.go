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

func (s *EventWorkflowService) CreateEventWorkflow(ctx context.Context, workflow *models.EventWorkflow) (*models.EventWorkflow, error) {
	if workflow == nil {
		return nil, errors.New("workflow cannot be nil")
	}
	if workflow.ProjectId == "" {
		return nil, errors.New("projectId is required")
	}
	if workflow.Name == "" {
		return nil, errors.New("workflow name is required")
	}

	project, err := s.projectRepo.GetPublicProjectByID(ctx, workflow.ProjectId)
	if err != nil {
		return nil, fmt.Errorf("failed to verify project: %w", err)
	}
	if project == nil {
		return nil, errors.New("project does not exist")
	}

	if workflow.CanvasData == nil {
		workflow.CanvasData = []byte("{}")
	}
	if workflow.Handlers == nil {
		workflow.Handlers = []byte("[]")
	}

	return s.eventWorkflowRepo.CreateEventWorkflow(ctx, workflow)
}

func (s *EventWorkflowService) GetEventWorkflowByID(ctx context.Context, id string) (*models.EventWorkflow, error) {
	if id == "" {
		return nil, errors.New("workflow id is required")
	}
	return s.eventWorkflowRepo.GetEventWorkflowByID(ctx, id)
}

func (s *EventWorkflowService) GetEventWorkflowsByProjectID(ctx context.Context, projectID string) ([]models.EventWorkflow, error) {
	if projectID == "" {
		return nil, errors.New("projectId is required")
	}

	project, err := s.projectRepo.GetPublicProjectByID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify project: %w", err)
	}
	if project == nil {
		return nil, errors.New("project does not exist")
	}

	return s.eventWorkflowRepo.GetEventWorkflowsByProjectID(ctx, projectID)
}

func (s *EventWorkflowService) GetEventWorkflowsByProjectIDWithElements(ctx context.Context, projectID string) ([]models.EventWorkflow, error) {
	if projectID == "" {
		return nil, errors.New("projectId is required")
	}

	project, err := s.projectRepo.GetPublicProjectByID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify project: %w", err)
	}
	if project == nil {
		return nil, errors.New("project does not exist")
	}

	return s.eventWorkflowRepo.GetEventWorkflowsByProjectIDWithElements(ctx, projectID)
}

func (s *EventWorkflowService) GetEnabledEventWorkflowsByProjectID(ctx context.Context, projectID string) ([]models.EventWorkflow, error) {
	if projectID == "" {
		return nil, errors.New("projectId is required")
	}

	project, err := s.projectRepo.GetPublicProjectByID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify project: %w", err)
	}
	if project == nil {
		return nil, errors.New("project does not exist")
	}

	return s.eventWorkflowRepo.GetEnabledEventWorkflowsByProjectID(ctx, projectID)
}

func (s *EventWorkflowService) GetEventWorkflowsByName(ctx context.Context, projectID, name string) ([]models.EventWorkflow, error) {
	if projectID == "" {
		return nil, errors.New("projectId is required")
	}
	if name == "" {
		return nil, errors.New("workflow name is required")
	}

	project, err := s.projectRepo.GetPublicProjectByID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify project: %w", err)
	}
	if project == nil {
		return nil, errors.New("project does not exist")
	}

	return s.eventWorkflowRepo.GetEventWorkflowsByName(ctx, projectID, name)
}

func (s *EventWorkflowService) UpdateEventWorkflow(ctx context.Context, id string, workflow *models.EventWorkflow) (*models.EventWorkflow, error) {
	if id == "" {
		return nil, errors.New("workflow id is required")
	}
	if workflow == nil {
		return nil, errors.New("workflow cannot be nil")
	}
	if workflow.Name == "" {
		return nil, errors.New("workflow name is required")
	}

	existing, err := s.GetEventWorkflowByID(ctx, id)
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

	return s.eventWorkflowRepo.UpdateEventWorkflow(ctx, id, workflow)
}

func (s *EventWorkflowService) UpdateEventWorkflowEnabled(ctx context.Context, id string, enabled bool) error {
	if id == "" {
		return errors.New("workflow id is required")
	}

	existing, err := s.GetEventWorkflowByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("workflow does not exist")
	}

	return s.eventWorkflowRepo.UpdateEventWorkflowEnabled(ctx, id, enabled)
}

func (s *EventWorkflowService) DeleteEventWorkflow(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("workflow id is required")
	}

	existing, err := s.GetEventWorkflowByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("workflow does not exist")
	}

	return s.eventWorkflowRepo.DeleteEventWorkflow(ctx, id)
}

func (s *EventWorkflowService) DeleteEventWorkflowsByProjectID(ctx context.Context, projectID string) error {
	if projectID == "" {
		return errors.New("projectId is required")
	}

	project, err := s.projectRepo.GetPublicProjectByID(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to verify project: %w", err)
	}
	if project == nil {
		return errors.New("project does not exist")
	}

	return s.eventWorkflowRepo.DeleteEventWorkflowsByProjectID(ctx, projectID)
}

func (s *EventWorkflowService) CountEventWorkflowsByProjectID(ctx context.Context, projectID string) (int64, error) {
	if projectID == "" {
		return 0, errors.New("projectId is required")
	}

	project, err := s.projectRepo.GetPublicProjectByID(ctx, projectID)
	if err != nil {
		return 0, fmt.Errorf("failed to verify project: %w", err)
	}
	if project == nil {
		return 0, errors.New("project does not exist")
	}

	return s.eventWorkflowRepo.CountEventWorkflowsByProjectID(ctx, projectID)
}

func (s *EventWorkflowService) CheckIfWorkflowNameExists(ctx context.Context, projectID, name, excludeID string) (bool, error) {
	if projectID == "" {
		return false, errors.New("projectId is required")
	}
	if name == "" {
		return false, errors.New("workflow name is required")
	}

	project, err := s.projectRepo.GetPublicProjectByID(ctx, projectID)
	if err != nil {
		return false, fmt.Errorf("failed to verify project: %w", err)
	}
	if project == nil {
		return false, errors.New("project does not exist")
	}

	return s.eventWorkflowRepo.CheckIfWorkflowNameExists(ctx, projectID, name, excludeID)
}

func (s *EventWorkflowService) GetEventWorkflowsWithFilters(ctx context.Context, projectID string, enabled *bool, searchName string) ([]models.EventWorkflow, error) {
	if projectID == "" {
		return nil, errors.New("projectId is required")
	}

	project, err := s.projectRepo.GetPublicProjectByID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify project: %w", err)
	}
	if project == nil {
		return nil, errors.New("project does not exist")
	}

	return s.eventWorkflowRepo.GetEventWorkflowsWithFilters(ctx, projectID, enabled, searchName)
}