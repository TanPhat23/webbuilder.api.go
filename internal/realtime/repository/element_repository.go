package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	domainEntity "my-go-app/internal/realtime/domain/entity"
	domainRepo "my-go-app/internal/realtime/domain/repository"
)

type elementRepository struct {
	db *gorm.DB
}

// NewElementRepository creates a new instance of ElementRepository
func NewElementRepository(db *gorm.DB) domainRepo.ElementRepository {
	return &elementRepository{db: db}
}

// CreateElement persists a new element to the database, ensuring it belongs to the correct page
func (r *elementRepository) CreateElement(ctx context.Context, projectID string, pageID string, element *domainEntity.Element) error {
	element.PageID = pageID

	if err := r.db.WithContext(ctx).Create(element).Error; err != nil {
		return fmt.Errorf("failed to create element: %w", err)
	}
	return nil
}

// Create Elements from a flattened elements array
func (r *elementRepository) CreateElements(ctx context.Context, projectID string, pageId string, elements []domainEntity.Element) error {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}

	if err := tx.Where(`"PageId" = ?`, pageId).Delete(&domainEntity.Element{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to clear elements for page %s: %w", pageId, err)
	}

	for i := range elements {
		elements[i].PageID = pageId
	}

	if len(elements) > 0 {
		if err := tx.CreateInBatches(elements, len(elements)).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit elements replace: %w", err)
	}

	return nil
}

// UpdateElement updates an existing element in the database, scoped by pageID
func (r *elementRepository) UpdateElement(ctx context.Context, projectID string, pageID string, element *domainEntity.Element) error {
	result := r.db.WithContext(ctx).
		Model(&domainEntity.Element{}).
		Where(`"Id" = ? AND "PageId" = ?`, element.ID, pageID).
		Save(element)

	if result.Error != nil {
		return fmt.Errorf("failed to update element: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("element not found in specified page")
	}

	return nil
}

// UpdateElements updates multiple elements within a transaction, scoped by pageID
func (r *elementRepository) UpdateElements(ctx context.Context, projectID string, pageID string, elements []domainEntity.Element) error {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}

	ids := make([]string, 0, len(elements))
	for i := range elements {
		elements[i].PageID = pageID
		ids = append(ids, elements[i].ID)
	}
	if len(elements) > 0 {
		result := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "Id"}},
			DoUpdates: clause.AssignmentColumns([]string{"Name", "Type", "Content", "Href", "Src", "Styles", "Settings", "TailwindStyles", "Order", "OrderKey", "IsLocked", "ParentId"}),
			Where:     clause.Where{Exprs: []clause.Expression{clause.Eq{Column: clause.Column{Table: "Element", Name: "PageId"}, Value: pageID}}},
		}).CreateInBatches(elements, len(elements))
		if result.Error != nil || result.RowsAffected != int64(len(elements)) {
			tx.Rollback()
			if result.Error != nil {
				return fmt.Errorf("failed to save elements: %w", result.Error)
			}
			return fmt.Errorf("an element belongs to another page")
		}
	}
	removed := tx.Where(`"PageId" = ?`, pageID)
	if len(ids) > 0 {
		removed = removed.Where(`"Id" NOT IN ?`, ids)
	}
	if err := removed.Delete(&domainEntity.Element{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to remove deleted elements: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit updates: %w", err)
	}
	return nil
}

// UpsertElements creates or updates multiple elements within a transaction, scoped by pageID
func (r *elementRepository) UpsertElements(ctx context.Context, projectID string, pageID string, elements []domainEntity.Element) error {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}

	for i := range elements {
		elements[i].PageID = pageID
	}

	if len(elements) > 0 {
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "Id"}},
			DoUpdates: clause.AssignmentColumns([]string{"Name", "Type", "Content", "Href", "Src", "Styles", "Settings", "TailwindStyles", "Order", "IsLocked", "ParentId", "PageId"}),
		}).CreateInBatches(elements, len(elements)).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to upsert elements: %w", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit upsert: %w", err)
	}

	return nil
}

// DeleteElement removes an element by ID, scoped by pageID
func (r *elementRepository) DeleteElement(ctx context.Context, projectID string, pageID string, elementID string) error {
	result := r.db.WithContext(ctx).
		Where(`"Id" = ? AND "PageId" = ?`, elementID, pageID).
		Delete(&domainEntity.Element{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete element: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("element not found in specified page")
	}

	return nil
}

// GetElementByID retrieves a single element with its relations, scoped by pageID
func (r *elementRepository) GetElementByID(ctx context.Context, projectID string, pageID string, elementID string) (*domainEntity.Element, error) {
	var elem domainEntity.Element
	err := r.db.WithContext(ctx).
		Preload("Comments").
		Preload("Comments.Author").
		Preload("EventWorkflows").
		Preload("EventWorkflows.Workflow").
		Where(`"Id" = ? AND "PageId" = ?`, elementID, pageID).
		First(&elem).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get element by id: %w", err)
	}
	return &elem, nil
}

// GetElementsByPageID retrieves all elements for a specific page, ordered by Order
func (r *elementRepository) GetElementsByPageID(ctx context.Context, projectID string, pageID string) ([]domainEntity.Element, error) {
	var elems []domainEntity.Element
	err := r.db.WithContext(ctx).
		Model(&domainEntity.Element{}).
		Joins(`JOIN "Page" ON "Element"."PageId" = "Page"."Id"`).
		Where(`"Page"."ProjectId" = ? AND "Element"."PageId" = ?`, projectID, pageID).
		Order(`COALESCE("Element"."ParentId",'') ASC, "Element"."Order" ASC`).
		Preload("Comments").
		Preload("Comments.Author").
		Preload("EventWorkflows").
		Preload("EventWorkflows.Workflow").
		Find(&elems).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get elements by page id: %w", err)
	}
	return elems, nil
}

// GetElementsByProjectID retrieves all elements for a specific project
func (r *elementRepository) GetElementsByProjectID(ctx context.Context, projectID string) ([]domainEntity.Element, error) {
	var elems []domainEntity.Element

	err := r.db.WithContext(ctx).
		Model(&domainEntity.Element{}).
		Joins(`JOIN "Page" ON "Element"."PageId" = "Page"."Id"`).
		Where(`"Page"."ProjectId" = ?`, projectID).
		Order(`"Page"."Id" ASC, COALESCE("Element"."ParentId",'') ASC, "Element"."Order" ASC`).
		Preload("Comments").
		Preload("Comments.Author").
		Preload("EventWorkflows").
		Preload("EventWorkflows.Workflow").
		Find(&elems).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get elements by project id: %w", err)
	}
	return elems, nil
}
