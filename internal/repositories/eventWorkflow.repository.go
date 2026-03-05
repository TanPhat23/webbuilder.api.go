package repositories

import (
	"context"
	"errors"
	"my-go-app/internal/models"

	"github.com/lucsky/cuid"
	"gorm.io/gorm"
)

type EventWorkflowRepository struct {
	db *gorm.DB
}

func NewEventWorkflowRepository(db *gorm.DB) EventWorkflowRepositoryInterface {
	return &EventWorkflowRepository{db: db}
}

// CreateEventWorkflow creates a new event workflow
func (r *EventWorkflowRepository) CreateEventWorkflow(ctx context.Context, workflow *models.EventWorkflow) (*models.EventWorkflow, error) {
	if workflow.Id == "" {
		workflow.Id = cuid.New()
	}
	if err := r.db.WithContext(ctx).Create(workflow).Error; err != nil {
		return nil, err
	}
	return workflow, nil
}

// GetEventWorkflowByID retrieves an event workflow by ID
func (r *EventWorkflowRepository) GetEventWorkflowByID(ctx context.Context, id string) (*models.EventWorkflow, error) {
	var workflow models.EventWorkflow
	err := r.db.WithContext(ctx).
		Preload("Project").
		Preload("ElementEventWorkflows").
		Where(`"Id" = ?`, id).
		First(&workflow).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &workflow, nil
}

// GetEventWorkflowsByProjectID retrieves all event workflows for a project
func (r *EventWorkflowRepository) GetEventWorkflowsByProjectID(ctx context.Context, projectID string) ([]models.EventWorkflow, error) {
	var workflows []models.EventWorkflow
	err := r.db.WithContext(ctx).
		Preload("ElementEventWorkflows").
		Where(`"ProjectId" = ?`, projectID).
		Find(&workflows).Error
	if err != nil {
		return nil, err
	}
	return workflows, nil
}

// GetEventWorkflowsByProjectIDWithElements retrieves all event workflows for a project with element details
func (r *EventWorkflowRepository) GetEventWorkflowsByProjectIDWithElements(ctx context.Context, projectID string) ([]models.EventWorkflow, error) {
	var workflows []models.EventWorkflow
	err := r.db.WithContext(ctx).
		Preload("ElementEventWorkflows").
		Preload("ElementEventWorkflows.Element").
		Where(`"ProjectId" = ?`, projectID).
		Find(&workflows).Error
	if err != nil {
		return nil, err
	}
	return workflows, nil
}

// GetEnabledEventWorkflowsByProjectID retrieves all enabled event workflows for a project
func (r *EventWorkflowRepository) GetEnabledEventWorkflowsByProjectID(ctx context.Context, projectID string) ([]models.EventWorkflow, error) {
	var workflows []models.EventWorkflow
	err := r.db.WithContext(ctx).
		Preload("ElementEventWorkflows").
		Where(`"ProjectId" = ? AND "Enabled" = ?`, projectID, true).
		Find(&workflows).Error
	if err != nil {
		return nil, err
	}
	return workflows, nil
}

// GetEventWorkflowsByName retrieves event workflows by name in a project
func (r *EventWorkflowRepository) GetEventWorkflowsByName(ctx context.Context, projectID, name string) ([]models.EventWorkflow, error) {
	var workflows []models.EventWorkflow
	err := r.db.WithContext(ctx).
		Preload("ElementEventWorkflows").
		Where(`"ProjectId" = ? AND "Name" ILIKE ?`, projectID, "%"+name+"%").
		Find(&workflows).Error
	if err != nil {
		return nil, err
	}
	return workflows, nil
}

// UpdateEventWorkflow updates an event workflow
func (r *EventWorkflowRepository) UpdateEventWorkflow(ctx context.Context, id string, workflow *models.EventWorkflow) (*models.EventWorkflow, error) {
	updates := map[string]interface{}{
		"Name":        workflow.Name,
		"Description": workflow.Description,
		"Enabled":     workflow.Enabled,
	}

	if len(workflow.CanvasData) > 0 {
		updates["CanvasData"] = workflow.CanvasData
	}
	if len(workflow.Handlers) > 0 {
		updates["Handlers"] = workflow.Handlers
	}

	if err := r.db.WithContext(ctx).
		Model(&models.EventWorkflow{Id: id}).
		Updates(updates).Error; err != nil {
		return nil, err
	}

	// Fetch the updated record
	return r.GetEventWorkflowByID(ctx, id)
}

// UpdateEventWorkflowEnabled updates the enabled status of an event workflow
func (r *EventWorkflowRepository) UpdateEventWorkflowEnabled(ctx context.Context, id string, enabled bool) error {
	return r.db.WithContext(ctx).
		Model(&models.EventWorkflow{}).
		Where(`"Id" = ?`, id).
		Update("Enabled", enabled).Error
}

// DeleteEventWorkflow deletes an event workflow
func (r *EventWorkflowRepository) DeleteEventWorkflow(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Where(`"Id" = ?`, id).
		Delete(&models.EventWorkflow{}).Error
}

// DeleteEventWorkflowsByProjectID deletes all event workflows for a project
func (r *EventWorkflowRepository) DeleteEventWorkflowsByProjectID(ctx context.Context, projectID string) error {
	return r.db.WithContext(ctx).
		Where(`"ProjectId" = ?`, projectID).
		Delete(&models.EventWorkflow{}).Error
}

// CountEventWorkflowsByProjectID counts event workflows in a project
func (r *EventWorkflowRepository) CountEventWorkflowsByProjectID(ctx context.Context, projectID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.EventWorkflow{}).
		Where(`"ProjectId" = ?`, projectID).
		Count(&count).Error
	return count, err
}

// CheckIfWorkflowNameExists checks if a workflow name already exists in a project
func (r *EventWorkflowRepository) CheckIfWorkflowNameExists(ctx context.Context, projectID, name string, excludeID string) (bool, error) {
	var count int64
	query := r.db.WithContext(ctx).
		Model(&models.EventWorkflow{}).
		Where(`"ProjectId" = ? AND "Name" = ?`, projectID, name)

	if excludeID != "" {
		query = query.Where(`"Id" != ?`, excludeID)
	}

	err := query.Count(&count).Error
	return count > 0, err
}

// GetEventWorkflowsWithFilters retrieves event workflows with optional filters
func (r *EventWorkflowRepository) GetEventWorkflowsWithFilters(ctx context.Context, projectID string, enabled *bool, searchName string) ([]models.EventWorkflow, error) {
	var workflows []models.EventWorkflow
	query := r.db.WithContext(ctx).Model(&models.EventWorkflow{})

	if projectID != "" {
		query = query.Where(`"ProjectId" = ?`, projectID)
	}

	if enabled != nil {
		query = query.Where(`"Enabled" = ?`, *enabled)
	}

	if searchName != "" {
		query = query.Where(`"Name" ILIKE ?`, "%"+searchName+"%")
	}

	err := query.
		Preload("ElementEventWorkflows").
		Find(&workflows).Error
	if err != nil {
		return nil, err
	}
	return workflows, nil
}
