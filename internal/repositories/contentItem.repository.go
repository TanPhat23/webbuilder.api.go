package repositories

import (
	"context"
	"errors"
	"fmt"
	"my-go-app/internal/models"
	"time"

	"github.com/lucsky/cuid"
	"gorm.io/gorm"
)

var (
	ErrContentItemNotFound  = errors.New("content item not found")
	ErrContentItemDuplicate = errors.New("content item with this slug already exists")
)

type ContentItemRepository struct {
	db *gorm.DB
}

func NewContentItemRepository(db *gorm.DB) ContentItemRepositoryInterface {
	return &ContentItemRepository{db: db}
}

func (r *ContentItemRepository) GetContentItemsByContentType(ctx context.Context, contentTypeID string) ([]models.ContentItem, error) {
	if contentTypeID == "" {
		return nil, errors.New("contentTypeID is required")
	}

	contentItems, err := gorm.G[models.ContentItem](r.db).
		Where("\"ContentTypeId\" = ?", contentTypeID).
		Preload("FieldValues.Field", nil).
		Preload("ContentType", nil).
		Order("\"CreatedAt\" DESC").
		Find(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get content items by content type: %w", err)
	}

	return contentItems, nil
}

func (r *ContentItemRepository) GetContentItemByID(ctx context.Context, id string) (*models.ContentItem, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}

	contentItem, err := gorm.G[models.ContentItem](r.db).
		Where("\"Id\" = ?", id).
		Preload("FieldValues.Field", nil).
		Preload("ContentType", nil).
		First(ctx)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrContentItemNotFound
		}
		return nil, fmt.Errorf("failed to get content item: %w", err)
	}

	return &contentItem, nil
}

func (r *ContentItemRepository) CreateContentItem(ctx context.Context, ci *models.ContentItem, fieldValues []models.ContentFieldValue) (*models.ContentItem, error) {
	if ci == nil {
		return nil, errors.New("content item cannot be nil")
	}

	if ci.Title == "" {
		return nil, errors.New("content item title is required")
	}

	if ci.Slug == "" {
		return nil, errors.New("content item slug is required")
	}

	if ci.ContentTypeId == "" {
		return nil, errors.New("content type ID is required")
	}

	// Check for duplicate slug
	count, err := gorm.G[models.ContentItem](r.db).
		Where("\"ContentTypeId\" = ? AND \"Slug\" = ?", ci.ContentTypeId, ci.Slug).
		Count(ctx, "*")

	if err != nil {
		return nil, fmt.Errorf("failed to check content item existence: %w", err)
	}

	if count > 0 {
		return nil, ErrContentItemDuplicate
	}

	// Generate ID if not provided
	if ci.Id == "" {
		ci.Id = cuid.New()
	}

	// Set timestamps
	now := time.Now()
	ci.CreatedAt = now
	ci.UpdatedAt = now

	// Create content item and field values in a transaction
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create content item
		if err := gorm.G[models.ContentItem](tx).Create(ctx, ci); err != nil {
			return fmt.Errorf("failed to create content item: %w", err)
		}

		// Create field values
		for i := range fieldValues {
			if fieldValues[i].Id == "" {
				fieldValues[i].Id = cuid.New()
			}
			fieldValues[i].ContentItemId = ci.Id

			if err := gorm.G[models.ContentFieldValue](tx).Create(ctx, &fieldValues[i]); err != nil {
				return fmt.Errorf("failed to create field value: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return r.GetContentItemByID(ctx, ci.Id)
}

func (r *ContentItemRepository) UpdateContentItem(ctx context.Context, id string, updates map[string]any, fieldValues []models.ContentFieldValue) (*models.ContentItem, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}

	if len(updates) == 0 && len(fieldValues) == 0 {
		return nil, errors.New("no updates provided")
	}

	existing, err := r.GetContentItemByID(ctx, id)
	if err != nil {
		return nil, err
	}

	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if len(fieldValues) > 0 {
			if _, err := gorm.G[models.ContentFieldValue](tx).
				Where("\"ContentItemId\" = ?", id).
				Delete(ctx); err != nil {
				return fmt.Errorf("failed to delete existing field values: %w", err)
			}

			for i := range fieldValues {
				if fieldValues[i].Id == "" {
					fieldValues[i].Id = cuid.New()
				}
				fieldValues[i].ContentItemId = id

				if err := gorm.G[models.ContentFieldValue](tx).Create(ctx, &fieldValues[i]); err != nil {
					return fmt.Errorf("failed to create field value: %w", err)
				}
			}
		}

		if len(updates) > 0 {
			if newSlug, ok := updates["slug"].(string); ok {
				count, err := gorm.G[models.ContentItem](tx).
					Where("\"ContentTypeId\" = ? AND \"Slug\" = ? AND \"Id\" != ?", existing.ContentTypeId, newSlug, id).
					Count(ctx, "*")

				if err != nil {
					return fmt.Errorf("failed to check slug uniqueness: %w", err)
				}

				if count > 0 {
					return ErrContentItemDuplicate
				}
			}

			updates["UpdatedAt"] = time.Now()

			if err := tx.Model(&models.ContentItem{}).
				Where("\"Id\" = ?", id).
				Updates(updates).Error; err != nil {
				return fmt.Errorf("failed to update content item: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Fetch and return updated content item
	return r.GetContentItemByID(ctx, id)
}

func (r *ContentItemRepository) DeleteContentItem(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id is required")
	}

	rowsAffected, err := gorm.G[models.ContentItem](r.db).
		Where("\"Id\" = ?", id).
		Delete(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete content item: %w", err)
	}

	if rowsAffected == 0 {
		return ErrContentItemNotFound
	}

	return nil
}

func (r *ContentItemRepository) GetPublicContentItems(ctx context.Context, contentTypeID string, limit int, sortBy, sortOrder string) ([]models.ContentItem, error) {
	if contentTypeID == "" {
		return nil, errors.New("contentTypeID is required")
	}

	// Validate and set defaults
	if limit <= 0 || limit > MaxPageSize {
		limit = DefaultPageSize
	}

	// Validate sort order
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	// Validate sort by field
	validSortFields := map[string]bool{
		"CreatedAt": true,
		"UpdatedAt": true,
		"Title":     true,
		"Slug":      true,
	}

	if !validSortFields[sortBy] {
		sortBy = "CreatedAt"
	}

	contentItems, err := gorm.G[models.ContentItem](r.db).
		Where("\"ContentTypeId\" = ? AND \"Published\" = ?", contentTypeID, true).
		Preload("FieldValues.Field", nil).
		Preload("ContentType", nil).
		Order(fmt.Sprintf("\"%s\" %s", sortBy, sortOrder)).
		Limit(limit).
		Find(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get public content items: %w", err)
	}

	return contentItems, nil
}

func (r *ContentItemRepository) GetContentItemBySlug(ctx context.Context, contentTypeID, slug string) (*models.ContentItem, error) {
	if contentTypeID == "" || slug == "" {
		return nil, errors.New("contentTypeID and slug are required")
	}

	contentItem, err := gorm.G[models.ContentItem](r.db).
		Where("\"ContentTypeId\" = ? AND \"Slug\" = ?", contentTypeID, slug).
		Preload("FieldValues.Field", nil).
		Preload("ContentType", nil).
		First(ctx)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrContentItemNotFound
		}
		return nil, fmt.Errorf("failed to get content item by slug: %w", err)
	}

	return &contentItem, nil
}
