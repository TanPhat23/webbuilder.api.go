package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"my-go-app/internal/realtime/domain/entity"
	domainRepo "my-go-app/internal/realtime/domain/repository"
)

type workflowRepository struct {
	db *gorm.DB
}

// NewWorkflowRepository creates a new instance of WorkflowRepository
func NewWorkflowRepository(db *gorm.DB) domainRepo.WorkflowRepository {
	return &workflowRepository{
		db: db,
	}
}

// CreateWorkflow persists a new workflow to the database.
func (r *workflowRepository) CreateWorkflow(ctx context.Context, projectID string, workflow *entity.EventWorkflow) error {
	workflow.ProjectID = projectID
	if err := r.db.WithContext(ctx).Create(workflow).Error; err != nil {
		return fmt.Errorf("failed to create workflow: %w", err)
	}
	return nil
}

// UpdateWorkflow updates an existing workflow, verified by projectID.
func (r *workflowRepository) UpdateWorkflow(ctx context.Context, projectID string, workflow *entity.EventWorkflow) error {
	result := r.db.WithContext(ctx).
		Model(&entity.EventWorkflow{}).
		Where(`"Id" = ? AND "ProjectId" = ?`, workflow.ID, projectID).
		Updates(workflow)

	if result.Error != nil {
		return fmt.Errorf("failed to update workflow: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("workflow not found or unauthorized")
	}

	return nil
}

// DeleteWorkflow removes a workflow by ID, verified by projectID.
func (r *workflowRepository) DeleteWorkflow(ctx context.Context, projectID string, workflowID string) error {
	result := r.db.WithContext(ctx).
		Where(`"Id" = ? AND "ProjectId" = ?`, workflowID, projectID).
		Delete(&entity.EventWorkflow{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete workflow: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("workflow not found or unauthorized")
	}

	return nil
}

// GetWorkflowByID retrieves a single workflow by ID and projectID.
func (r *workflowRepository) GetWorkflowByID(ctx context.Context, projectID string, workflowID string) (*entity.EventWorkflow, error) {
	var workflow entity.EventWorkflow
	err := r.db.WithContext(ctx).
		Where(`"Id" = ? AND "ProjectId" = ?`, workflowID, projectID).
		First(&workflow).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get workflow: %w", err)
	}
	return &workflow, nil
}

// GetWorkflowsByProjectID retrieves all workflows for a specific project.
func (r *workflowRepository) GetWorkflowsByProjectID(ctx context.Context, projectID string) ([]entity.EventWorkflow, error) {
	var workflows []entity.EventWorkflow
	err := r.db.WithContext(ctx).
		Where(`"ProjectId" = ?`, projectID).
		Order(`"CreatedAt" DESC`).
		Find(&workflows).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get workflows by project id: %w", err)
	}
	return workflows, nil
}
