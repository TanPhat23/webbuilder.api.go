package repositories

import (
	"context"
	"errors"
	"my-go-app/internal/models"

	"github.com/lucsky/cuid"
	"gorm.io/gorm"
)

type ElementEventWorkflowRepository struct {
	db *gorm.DB
}

func NewElementEventWorkflowRepository(db *gorm.DB) ElementEventWorkflowRepositoryInterface {
	return &ElementEventWorkflowRepository{db: db}
}

// Create creates a new element event workflow association
func (r *ElementEventWorkflowRepository) Create(ctx context.Context, eew *models.ElementEventWorkflow) (*models.ElementEventWorkflow, error) {
	if eew.Id == "" {
		eew.Id = cuid.New()
	}
	if err := gorm.G[models.ElementEventWorkflow](r.db).Create(ctx, eew); err != nil {
		return nil, err
	}
	return eew, nil
}

// GetByID retrieves an element event workflow by ID
func (r *ElementEventWorkflowRepository) GetByID(ctx context.Context, id string) (*models.ElementEventWorkflow, error) {

	eew, err := gorm.G[models.ElementEventWorkflow](r.db).
		Preload("Element", nil).
		Preload("Workflow", nil).
		Where(`"Id" = ?`, id).
		First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &eew, nil
}

// ListAll retrieves all element event workflows
func (r *ElementEventWorkflowRepository) ListAll(ctx context.Context) ([]models.ElementEventWorkflow, error) {

	eews, err := gorm.G[models.ElementEventWorkflow](r.db).
		Preload("Element", nil).
		Preload("Workflow", nil).
		Find(ctx)
	if err != nil {
		return nil, err
	}
	return eews, nil
}

// ListByElement retrieves all event workflows for a specific element
func (r *ElementEventWorkflowRepository) ListByElement(ctx context.Context, elementID string) ([]models.ElementEventWorkflow, error) {

	eews, err := gorm.G[models.ElementEventWorkflow](r.db).
		Preload("Workflow", nil).
		Where(`"ElementId" = ?`, elementID).
		Find(ctx)
	if err != nil {
		return nil, err
	}
	return eews, nil
}

// ListByWorkflow retrieves all elements linked to a specific workflow
func (r *ElementEventWorkflowRepository) ListByWorkflow(ctx context.Context, workflowID string) ([]models.ElementEventWorkflow, error) {

	eews, err := gorm.G[models.ElementEventWorkflow](r.db).
		Preload("Element", nil).
		Where(`"WorkflowId" = ?`, workflowID).
		Find(ctx)
	if err != nil {
		return nil, err
	}
	return eews, nil
}

// ListByEvent retrieves all workflows for a specific event type
func (r *ElementEventWorkflowRepository) ListByEvent(ctx context.Context, eventName string) ([]models.ElementEventWorkflow, error) {

	eews, err := gorm.G[models.ElementEventWorkflow](r.db).
		Preload("Element", nil).
		Preload("Workflow", nil).
		Where(`"EventName" = ?`, eventName).
		Find(ctx)
	if err != nil {
		return nil, err
	}
	return eews, nil
}

// ListByFilters retrieves element event workflows with optional filters
func (r *ElementEventWorkflowRepository) ListByFilters(ctx context.Context, elementID, workflowID, eventName string) ([]models.ElementEventWorkflow, error) {

	query := gorm.G[models.ElementEventWorkflow](r.db).Where("1 = 1")

	if elementID != "" {
		query = query.Where(`"ElementId" = ?`, elementID)
	}
	if workflowID != "" {
		query = query.Where(`"WorkflowId" = ?`, workflowID)
	}
	if eventName != "" {
		query = query.Where(`"EventName" = ?`, eventName)
	}

	eews, err := query.
		Preload("Element", nil).
		Preload("Workflow", nil).
		Find(ctx)
	if err != nil {
		return nil, err
	}
	return eews, nil
}

// Update updates an element event workflow
func (r *ElementEventWorkflowRepository) Update(ctx context.Context, id string, eew *models.ElementEventWorkflow) (*models.ElementEventWorkflow, error) {
	err := r.db.WithContext(ctx).
		Model(&models.ElementEventWorkflow{}).
		Where(`"Id" = ?`, id).
		Updates(eew).Error
	if err != nil {
		return nil, err
	}

	// Fetch the updated record
	return r.GetByID(ctx, id)
}

// Delete deletes an element event workflow
func (r *ElementEventWorkflowRepository) Delete(ctx context.Context, id string) error {
	_, err := gorm.G[models.ElementEventWorkflow](r.db).Where(`"Id" = ?`, id).
		Delete(ctx)
	return err
}

// DeleteByElement deletes all event workflows for a specific element
func (r *ElementEventWorkflowRepository) DeleteByElement(ctx context.Context, elementID string) error {
	_, err := gorm.G[models.ElementEventWorkflow](r.db).Where(`"ElementId" = ?`, elementID).
		Delete(ctx)
	return err
}

// DeleteByWorkflow deletes all element associations for a specific workflow
func (r *ElementEventWorkflowRepository) DeleteByWorkflow(ctx context.Context, workflowID string) error {
	_, err := gorm.G[models.ElementEventWorkflow](r.db).Where(`"WorkflowId" = ?`, workflowID).
		Delete(ctx)
	return err
}

// ListByPage retrieves all element event workflows for elements belonging to a specific page
func (r *ElementEventWorkflowRepository) ListByPage(ctx context.Context, pageID string) ([]models.ElementEventWorkflow, error) {
	var eews []models.ElementEventWorkflow
	err := r.db.WithContext(ctx).
		Preload("Workflow").
		Joins(`JOIN "Element" ON "Element"."Id" = "ElementEventWorkflow"."ElementId"`).
		Where(`"Element"."PageId" = ?`, pageID).
		Find(&eews).Error
	if err != nil {
		return nil, err
	}
	return eews, nil
}

// IsLinked checks if a workflow is already linked to an element with a specific event
func (r *ElementEventWorkflowRepository) IsLinked(ctx context.Context, elementID, workflowID, eventName string) (bool, error) {

	count, err := gorm.G[models.ElementEventWorkflow](r.db).
		Where(`"ElementId" = ? AND "WorkflowId" = ? AND "EventName" = ?`, elementID, workflowID, eventName).
		Count(ctx, "*")
	return count > 0, err
}
