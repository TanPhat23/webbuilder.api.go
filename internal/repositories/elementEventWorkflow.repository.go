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

// CreateElementEventWorkflow creates a new element event workflow association
func (r *ElementEventWorkflowRepository) CreateElementEventWorkflow(ctx context.Context, eew *models.ElementEventWorkflow) (*models.ElementEventWorkflow, error) {
	if eew.Id == "" {
		eew.Id = cuid.New()
	}
	if err := r.db.WithContext(ctx).Create(eew).Error; err != nil {
		return nil, err
	}
	return eew, nil
}

// GetElementEventWorkflowByID retrieves an element event workflow by ID
func (r *ElementEventWorkflowRepository) GetElementEventWorkflowByID(ctx context.Context, id string) (*models.ElementEventWorkflow, error) {
	var eew models.ElementEventWorkflow
	err := r.db.WithContext(ctx).
		Preload("Element").
		Preload("Workflow").
		Where(`"Id" = ?`, id).
		First(&eew).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &eew, nil
}

// GetAllElementEventWorkflows retrieves all element event workflows
func (r *ElementEventWorkflowRepository) GetAllElementEventWorkflows(ctx context.Context) ([]models.ElementEventWorkflow, error) {
	var eews []models.ElementEventWorkflow
	err := r.db.WithContext(ctx).
		Preload("Element").
		Preload("Workflow").
		Find(&eews).Error
	if err != nil {
		return nil, err
	}
	return eews, nil
}

// GetElementEventWorkflowsByElementID retrieves all event workflows for a specific element
func (r *ElementEventWorkflowRepository) GetElementEventWorkflowsByElementID(ctx context.Context, elementID string) ([]models.ElementEventWorkflow, error) {
	var eews []models.ElementEventWorkflow
	err := r.db.WithContext(ctx).
		Preload("Workflow").
		Where(`"ElementId" = ?`, elementID).
		Find(&eews).Error
	if err != nil {
		return nil, err
	}
	return eews, nil
}

// GetElementEventWorkflowsByWorkflowID retrieves all elements linked to a specific workflow
func (r *ElementEventWorkflowRepository) GetElementEventWorkflowsByWorkflowID(ctx context.Context, workflowID string) ([]models.ElementEventWorkflow, error) {
	var eews []models.ElementEventWorkflow
	err := r.db.WithContext(ctx).
		Preload("Element").
		Where(`"WorkflowId" = ?`, workflowID).
		Find(&eews).Error
	if err != nil {
		return nil, err
	}
	return eews, nil
}

// GetElementEventWorkflowsByEventName retrieves all workflows for a specific event type
func (r *ElementEventWorkflowRepository) GetElementEventWorkflowsByEventName(ctx context.Context, eventName string) ([]models.ElementEventWorkflow, error) {
	var eews []models.ElementEventWorkflow
	err := r.db.WithContext(ctx).
		Preload("Element").
		Preload("Workflow").
		Where(`"EventName" = ?`, eventName).
		Find(&eews).Error
	if err != nil {
		return nil, err
	}
	return eews, nil
}

// GetElementEventWorkflowsByFilters retrieves element event workflows with optional filters
func (r *ElementEventWorkflowRepository) GetElementEventWorkflowsByFilters(ctx context.Context, elementID, workflowID, eventName string) ([]models.ElementEventWorkflow, error) {
	var eews []models.ElementEventWorkflow
	query := r.db.WithContext(ctx).Model(&models.ElementEventWorkflow{})

	if elementID != "" {
		query = query.Where(`"ElementId" = ?`, elementID)
	}
	if workflowID != "" {
		query = query.Where(`"WorkflowId" = ?`, workflowID)
	}
	if eventName != "" {
		query = query.Where(`"EventName" = ?`, eventName)
	}

	err := query.
		Preload("Element").
		Preload("Workflow").
		Find(&eews).Error
	if err != nil {
		return nil, err
	}
	return eews, nil
}

// UpdateElementEventWorkflow updates an element event workflow
func (r *ElementEventWorkflowRepository) UpdateElementEventWorkflow(ctx context.Context, id string, eew *models.ElementEventWorkflow) (*models.ElementEventWorkflow, error) {
	err := r.db.WithContext(ctx).
		Model(&models.ElementEventWorkflow{}).
		Where(`"Id" = ?`, id).
		Updates(eew).Error
	if err != nil {
		return nil, err
	}

	// Fetch the updated record
	return r.GetElementEventWorkflowByID(ctx, id)
}

// DeleteElementEventWorkflow deletes an element event workflow
func (r *ElementEventWorkflowRepository) DeleteElementEventWorkflow(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Where(`"Id" = ?`, id).
		Delete(&models.ElementEventWorkflow{}).Error
}

// DeleteElementEventWorkflowsByElementID deletes all event workflows for a specific element
func (r *ElementEventWorkflowRepository) DeleteElementEventWorkflowsByElementID(ctx context.Context, elementID string) error {
	return r.db.WithContext(ctx).
		Where(`"ElementId" = ?`, elementID).
		Delete(&models.ElementEventWorkflow{}).Error
}

// DeleteElementEventWorkflowsByWorkflowID deletes all element associations for a specific workflow
func (r *ElementEventWorkflowRepository) DeleteElementEventWorkflowsByWorkflowID(ctx context.Context, workflowID string) error {
	return r.db.WithContext(ctx).
		Where(`"WorkflowId" = ?`, workflowID).
		Delete(&models.ElementEventWorkflow{}).Error
}

// GetElementEventWorkflowsByPageID retrieves all element event workflows for elements belonging to a specific page
func (r *ElementEventWorkflowRepository) GetElementEventWorkflowsByPageID(ctx context.Context, pageID string) ([]models.ElementEventWorkflow, error) {
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

// CheckIfWorkflowLinkedToElement checks if a workflow is already linked to an element with a specific event
func (r *ElementEventWorkflowRepository) CheckIfWorkflowLinkedToElement(ctx context.Context, elementID, workflowID, eventName string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.ElementEventWorkflow{}).
		Where(`"ElementId" = ? AND "WorkflowId" = ? AND "EventName" = ?`, elementID, workflowID, eventName).
		Count(&count).Error
	return count > 0, err
}
