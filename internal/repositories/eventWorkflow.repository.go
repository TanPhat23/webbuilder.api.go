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

// Create creates a new event workflow
func (r *EventWorkflowRepository) Create(ctx context.Context, workflow *models.EventWorkflow) (*models.EventWorkflow, error) {
	if workflow.Id == "" {
		workflow.Id = cuid.New()
	}
	if err := gorm.G[models.EventWorkflow](r.db).Create(ctx, workflow); err != nil {
		return nil, err
	}
	return workflow, nil
}

// GetByID retrieves an event workflow by ID
func (r *EventWorkflowRepository) GetByID(ctx context.Context, id string) (*models.EventWorkflow, error) {

	workflow, err := gorm.G[models.EventWorkflow](r.db).
		Preload("Project", nil).
		Preload("ElementEventWorkflows", nil).
		Where(`"Id" = ?`, id).
		First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &workflow, nil
}

// ListByProject retrieves all event workflows for a project
func (r *EventWorkflowRepository) ListByProject(ctx context.Context, projectID string) ([]models.EventWorkflow, error) {

	workflows, err := gorm.G[models.EventWorkflow](r.db).
		Preload("ElementEventWorkflows", nil).
		Where(`"ProjectId" = ?`, projectID).
		Find(ctx)
	if err != nil {
		return nil, err
	}
	return workflows, nil
}

// ListByProjectWithElements retrieves all event workflows for a project with element details
func (r *EventWorkflowRepository) ListByProjectWithElements(ctx context.Context, projectID string) ([]models.EventWorkflow, error) {

	workflows, err := gorm.G[models.EventWorkflow](r.db).
		Preload("ElementEventWorkflows", nil).
		Preload("ElementEventWorkflows.Element", nil).
		Where(`"ProjectId" = ?`, projectID).
		Find(ctx)
	if err != nil {
		return nil, err
	}
	return workflows, nil
}

// ListEnabledByProject retrieves all enabled event workflows for a project
func (r *EventWorkflowRepository) ListEnabledByProject(ctx context.Context, projectID string) ([]models.EventWorkflow, error) {

	workflows, err := gorm.G[models.EventWorkflow](r.db).
		Preload("ElementEventWorkflows", nil).
		Where(`"ProjectId" = ? AND "Enabled" = ?`, projectID, true).
		Find(ctx)
	if err != nil {
		return nil, err
	}
	return workflows, nil
}

// ListByName retrieves event workflows by name in a project
func (r *EventWorkflowRepository) ListByName(ctx context.Context, projectID, name string) ([]models.EventWorkflow, error) {

	workflows, err := gorm.G[models.EventWorkflow](r.db).
		Preload("ElementEventWorkflows", nil).
		Where(`"ProjectId" = ? AND "Name" ILIKE ?`, projectID, "%"+name+"%").
		Find(ctx)
	if err != nil {
		return nil, err
	}
	return workflows, nil
}

// Update updates an event workflow
func (r *EventWorkflowRepository) Update(ctx context.Context, id string, workflow *models.EventWorkflow) (*models.EventWorkflow, error) {
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
	return r.GetByID(ctx, id)
}

// SetEnabled updates the enabled status of an event workflow
func (r *EventWorkflowRepository) SetEnabled(ctx context.Context, id string, enabled bool) error {
	_, err := gorm.G[models.EventWorkflow](r.db).
		Where(`"Id" = ?`, id).
		Update(ctx, "Enabled", enabled)
	return err
}

// Delete deletes an event workflow
func (r *EventWorkflowRepository) Delete(ctx context.Context, id string) error {
	_, err := gorm.G[models.EventWorkflow](r.db).Where(`"Id" = ?`, id).
		Delete(ctx)
	return err
}

// DeleteByProject deletes all event workflows for a project
func (r *EventWorkflowRepository) DeleteByProject(ctx context.Context, projectID string) error {
	_, err := gorm.G[models.EventWorkflow](r.db).Where(`"ProjectId" = ?`, projectID).
		Delete(ctx)
	return err
}

// CountByProject counts event workflows in a project
func (r *EventWorkflowRepository) CountByProject(ctx context.Context, projectID string) (int64, error) {

	count, err := gorm.G[models.EventWorkflow](r.db).
		Where(`"ProjectId" = ?`, projectID).
		Count(ctx, "*")
	return count, err
}

// ExistsByName checks if a workflow name already exists in a project
func (r *EventWorkflowRepository) ExistsByName(ctx context.Context, projectID, name string, excludeID string) (bool, error) {

	query := gorm.G[models.EventWorkflow](r.db).
		Where(`"ProjectId" = ? AND "Name" = ?`, projectID, name)

	if excludeID != "" {
		query = query.Where(`"Id" != ?`, excludeID)
	}

	count, err := query.Count(ctx, "*")
	return count > 0, err
}

// ListWithFilters retrieves event workflows with optional filters
func (r *EventWorkflowRepository) ListWithFilters(ctx context.Context, projectID string, enabled *bool, searchName string) ([]models.EventWorkflow, error) {

	query := gorm.G[models.EventWorkflow](r.db).Where("1 = 1")

	if projectID != "" {
		query = query.Where(`"ProjectId" = ?`, projectID)
	}

	if enabled != nil {
		query = query.Where(`"Enabled" = ?`, *enabled)
	}

	if searchName != "" {
		query = query.Where(`"Name" ILIKE ?`, "%"+searchName+"%")
	}

	workflows, err := query.
		Preload("ElementEventWorkflows", nil).
		Find(ctx)
	if err != nil {
		return nil, err
	}
	return workflows, nil
}
