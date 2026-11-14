package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"my-go-app/internal/realtime/domain/entity"
	domainRepo "my-go-app/internal/realtime/domain/repository"
)

type elementWorkflowRepository struct {
	db *gorm.DB
}

// NewElementWorkflowRepository creates a new instance of ElementWorkflowRepository
func NewElementWorkflowRepository(db *gorm.DB) domainRepo.ElementWorkflowRepository {
	return &elementWorkflowRepository{db: db}
}

// projectScope leverages GORM's relation-based joins to filter records by ProjectId.
func (r *elementWorkflowRepository) projectScope(projectID string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Joins("Element").
			Joins("Element.Page").
			Where("\"Element__Page\".\"ProjectId\" = ?", projectID)
	}
}

// pageScope leverages GORM's relation-based joins to filter records by ProjectId and PageId.
func (r *elementWorkflowRepository) pageScope(projectID string, pageID string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Joins("Element").
			Joins("Element.Page").
			Where("\"Element__Page\".\"ProjectId\" = ? AND \"Element\".\"PageId\" = ?", projectID, pageID)
	}
}

// elementsInProject is a helper subquery to find all element IDs belonging to a specific project.
func (r *elementWorkflowRepository) elementsInProject(projectID string) *gorm.DB {
	return r.db.Model(&entity.Element{}).
		Select("\"Element\".\"Id\"").
		Joins("Page").
		Where("\"Page\".\"ProjectId\" = ?", projectID)
}

func (r *elementWorkflowRepository) CreateElementWorkflow(ctx context.Context, projectID string, workflow *entity.ElementEventWorkflow) error {
	// Verify Element and Workflow belong to the project before linking
	var count int64

	err := r.db.WithContext(ctx).Model(&entity.Element{}).
		Joins("Page").
		Where("\"Element\".\"Id\" = ? AND \"Page\".\"ProjectId\" = ?", workflow.ElementID, projectID).
		Count(&count).Error
	if err != nil || count == 0 {
		return fmt.Errorf("invalid element for project: %w", err)
	}

	err = r.db.WithContext(ctx).Model(&entity.EventWorkflow{}).
		Where("\"Id\" = ? AND \"ProjectId\" = ?", workflow.WorkflowID, projectID).
		Count(&count).Error
	if err != nil || count == 0 {
		return fmt.Errorf("invalid workflow for project: %w", err)
	}

	if err := r.db.WithContext(ctx).Create(workflow).Error; err != nil {
		return fmt.Errorf("failed to create element workflow: %w", err)
	}
	return nil
}

func (r *elementWorkflowRepository) GetElementWorkflowByElementID(ctx context.Context, projectID string, elementID string) (*entity.ElementEventWorkflow, error) {
	var workflow entity.ElementEventWorkflow
	err := r.db.WithContext(ctx).
		Preload("Element").
		Preload("Workflow").
		Scopes(r.projectScope(projectID)).
		Where("\"ElementEventWorkflow\".\"ElementId\" = ?", elementID).
		First(&workflow).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get element workflow: %w", err)
	}
	return &workflow, nil
}

func (r *elementWorkflowRepository) GetElementWorkflowsByProjectID(ctx context.Context, projectID string) ([]entity.ElementEventWorkflow, error) {
	var workflows []entity.ElementEventWorkflow
	err := r.db.WithContext(ctx).
		Preload("Element").
		Preload("Workflow").
		Scopes(r.projectScope(projectID)).
		Find(&workflows).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get element workflows: %w", err)
	}
	return workflows, nil
}

func (r *elementWorkflowRepository) GetElementWorkflowsByPageID(ctx context.Context, projectID string, pageID string) ([]entity.ElementEventWorkflow, error) {
	var workflows []entity.ElementEventWorkflow
	err := r.db.WithContext(ctx).
		Preload("Element").
		Preload("Workflow").
		Scopes(r.pageScope(projectID, pageID)).
		Find(&workflows).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get element workflows by page: %w", err)
	}
	return workflows, nil
}

func (r *elementWorkflowRepository) UpdateElementWorkflow(ctx context.Context, projectID string, workflow *entity.ElementEventWorkflow) error {
	result := r.db.WithContext(ctx).
		Model(workflow).
		Where("\"Id\" = ? AND \"ElementId\" IN (?)", workflow.ID, r.elementsInProject(projectID)).
		Updates(workflow)

	if result.Error != nil {
		return fmt.Errorf("failed to update element workflow: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("element workflow not found or unauthorized")
	}
	return nil
}

func (r *elementWorkflowRepository) DeleteElementWorkflow(ctx context.Context, projectID string, elementID string) error {
	result := r.db.WithContext(ctx).
		Where("\"ElementId\" = ? AND \"ElementId\" IN (?)", elementID, r.elementsInProject(projectID)).
		Delete(&entity.ElementEventWorkflow{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete element workflow: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("element workflow not found or unauthorized")
	}
	return nil
}
