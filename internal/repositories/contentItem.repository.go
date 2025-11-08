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
	ErrContentItemNotFound = errors.New("content item not found")
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

	var contentItems []models.ContentItem

	err := r.db.WithContext(ctx).
		Model(&models.ContentItem{}).
		Where("\"ContentTypeId\" = ?", contentTypeID).
		Preload("FieldValues.Field").
		Preload("ContentType").
		Order("\"CreatedAt\" DESC").
		Find(&contentItems).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get content items by content type: %w", err)
	}

	return contentItems, nil
}

func (r *ContentItemRepository) GetContentItemByID(ctx context.Context, id string) (*models.ContentItem, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}

	var contentItem models.ContentItem

	err := r.db.WithContext(ctx).
		Model(&models.ContentItem{}).
		Where("\"Id\" = ?", id).
		Preload("FieldValues.Field").
		Preload("ContentType").
		First(&contentItem).Error

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
	var count int64
	var err error
	err = r.db.WithContext(ctx).
		Model(&models.ContentItem{}).
		Where("\"ContentTypeId\" = ? AND \"Slug\" = ?", ci.ContentTypeId, ci.Slug).
		Count(&count).Error

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
		if err := tx.Model(&models.ContentItem{}).Create(ci).Error; err != nil {
			return fmt.Errorf("failed to create content item: %w", err)
		}

		// Create field values
		for i := range fieldValues {
			if fieldValues[i].Id == "" {
				fieldValues[i].Id = cuid.New()
			}
			fieldValues[i].ContentItemId = ci.Id

			if err := tx.Model(&models.ContentFieldValue{}).Create(&fieldValues[i]).Error; err != nil {
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
			if err := tx.Model(&models.ContentFieldValue{}).
				Where("\"ContentItemId\" = ?", id).
				Delete(&models.ContentFieldValue{}).Error; err != nil {
				return fmt.Errorf("failed to delete existing field values: %w", err)
			}

			for i := range fieldValues {
				if fieldValues[i].Id == "" {
					fieldValues[i].Id = cuid.New()
				}
				fieldValues[i].ContentItemId = id

				if err := tx.Model(&models.ContentFieldValue{}).Create(&fieldValues[i]).Error; err != nil {
					return fmt.Errorf("failed to create field value: %w", err)
				}
			}
		}

		if len(updates) > 0 {
			if newSlug, ok := updates["slug"].(string); ok {
				var count int64
				err := tx.Model(&models.ContentItem{}).
					Where("\"ContentTypeId\" = ? AND \"Slug\" = ? AND \"Id\" != ?", existing.ContentTypeId, newSlug, id).
					Count(&count).Error

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

	result := r.db.WithContext(ctx).
		Where("\"Id\" = ?", id).
		Delete(&models.ContentItem{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete content item: %w", result.Error)
	}

	if result.RowsAffected == 0 {
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

	var contentItems []models.ContentItem

	err := r.db.WithContext(ctx).
		Model(&models.ContentItem{}).
		Where("\"ContentTypeId\" = ? AND \"Published\" = ?", contentTypeID, true).
		Preload("FieldValues.Field").
		Preload("ContentType").
		Order(fmt.Sprintf("\"%s\" %s", sortBy, sortOrder)).
		Limit(limit).
		Find(&contentItems).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get public content items: %w", err)
	}

	return contentItems, nil
}

func (r *ContentItemRepository) GetContentItemBySlug(ctx context.Context, contentTypeID, slug string) (*models.ContentItem, error) {
	if contentTypeID == "" || slug == "" {
		return nil, errors.New("contentTypeID and slug are required")
	}

	var contentItem models.ContentItem

	err := r.db.WithContext(ctx).
		Model(&models.ContentItem{}).
		Where("\"ContentTypeId\" = ? AND \"Slug\" = ?", contentTypeID, slug).
		Preload("FieldValues.Field").
		Preload("ContentType").
		First(&contentItem).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrContentItemNotFound
		}
		return nil, fmt.Errorf("failed to get content item by slug: %w", err)
	}

	return &contentItem, nil
}
