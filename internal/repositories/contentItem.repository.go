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
	if err := validateNewContentItem(ci); err != nil {
		return nil, err
	}
	if err := ensureUniqueContentItemSlug(ctx, r.db, ci.ContentTypeId, ci.Slug); err != nil {
		return nil, err
	}
	prepareNewContentItem(ci)

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := gorm.G[models.ContentItem](tx).Create(ctx, ci); err != nil {
			return fmt.Errorf("failed to create content item: %w", err)
		}
		return insertContentFieldValues(ctx, tx, ci.Id, fieldValues)
	})
	if err != nil {
		return nil, err
	}
	return r.GetContentItemByID(ctx, ci.Id)
}

func validateNewContentItem(ci *models.ContentItem) error {
	if ci == nil {
		return errors.New("content item cannot be nil")
	}
	if ci.Title == "" {
		return errors.New("content item title is required")
	}
	if ci.Slug == "" {
		return errors.New("content item slug is required")
	}
	if ci.ContentTypeId == "" {
		return errors.New("content type ID is required")
	}
	return nil
}

func ensureUniqueContentItemSlug(ctx context.Context, db *gorm.DB, contentTypeID, slug string) error {
	count, err := gorm.G[models.ContentItem](db).
		Where("\"ContentTypeId\" = ? AND \"Slug\" = ?", contentTypeID, slug).
		Count(ctx, "*")
	if err != nil {
		return fmt.Errorf("failed to check content item existence: %w", err)
	}
	if count > 0 {
		return ErrContentItemDuplicate
	}
	return nil
}

func prepareNewContentItem(ci *models.ContentItem) {
	if ci.Id == "" {
		ci.Id = cuid.New()
	}
	now := time.Now()
	ci.CreatedAt = now
	ci.UpdatedAt = now
}

func insertContentFieldValues(ctx context.Context, tx *gorm.DB, contentItemID string, fieldValues []models.ContentFieldValue) error {
	for i := range fieldValues {
		if fieldValues[i].Id == "" {
			fieldValues[i].Id = cuid.New()
		}
		fieldValues[i].ContentItemId = contentItemID
		if err := gorm.G[models.ContentFieldValue](tx).Create(ctx, &fieldValues[i]); err != nil {
			return fmt.Errorf("failed to create field value: %w", err)
		}
	}
	return nil
}

func applyContentItemUpdates(ctx context.Context, tx *gorm.DB, existing *models.ContentItem, updates map[string]any) error {
	normalizeUpdateColumns(updates)
	if newSlug, ok := updates["Slug"].(string); ok {
		count, err := gorm.G[models.ContentItem](tx).
			Where("\"ContentTypeId\" = ? AND \"Slug\" = ? AND \"Id\" != ?", existing.ContentTypeId, newSlug, existing.Id).
			Count(ctx, "*")
		if err != nil {
			return fmt.Errorf("failed to check slug uniqueness: %w", err)
		}
		if count > 0 {
			return ErrContentItemDuplicate
		}
	}

	updates["UpdatedAt"] = time.Now()
	if err := tx.Model(&models.ContentItem{}).Where("\"Id\" = ?", existing.Id).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update content item: %w", err)
	}
	return nil
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
		if err := replaceContentFieldValues(ctx, tx, id, fieldValues); err != nil {
			return err
		}
		if len(updates) > 0 {
			return applyContentItemUpdates(ctx, tx, existing, updates)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return r.GetContentItemByID(ctx, id)
}

func replaceContentFieldValues(ctx context.Context, tx *gorm.DB, id string, fieldValues []models.ContentFieldValue) error {
	if len(fieldValues) == 0 {
		return nil
	}
	if _, err := gorm.G[models.ContentFieldValue](tx).
		Where("\"ContentItemId\" = ?", id).
		Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete existing field values: %w", err)
	}
	return insertContentFieldValues(ctx, tx, id, fieldValues)
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

var publicContentItemOrders = map[string]map[string]string{
	"CreatedAt": {
		"asc":  `public."ContentItem"."CreatedAt" asc`,
		"desc": `public."ContentItem"."CreatedAt" desc`,
	},
	"UpdatedAt": {
		"asc":  `public."ContentItem"."UpdatedAt" asc`,
		"desc": `public."ContentItem"."UpdatedAt" desc`,
	},
	"Title": {
		"asc":  `public."ContentItem"."Title" asc`,
		"desc": `public."ContentItem"."Title" desc`,
	},
	"Slug": {
		"asc":  `public."ContentItem"."Slug" asc`,
		"desc": `public."ContentItem"."Slug" desc`,
	},
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

	// Validate sort by field using static order expressions
	orderDirs, ok := publicContentItemOrders[sortBy]
	if !ok {
		orderDirs = publicContentItemOrders["CreatedAt"]
	}
	orderExpr := orderDirs[sortOrder]

	var contentItems []models.ContentItem
	err := r.db.WithContext(ctx).
		Model(&models.ContentItem{}).
		Joins(`JOIN public."ContentType" ON public."ContentType"."Id" = public."ContentItem"."ContentTypeId" JOIN public."Project" ON public."Project"."Id" = public."ContentType"."ProjectId"`).
		Where(`public."ContentItem"."ContentTypeId" = ? AND public."ContentItem"."Published" = ? AND public."ContentType"."PublicRead" = true AND public."Project"."Published" = true AND public."Project"."DeletedAt" IS NULL`, contentTypeID, true).
		Preload("FieldValues.Field").
		Preload("ContentType").
		Order(orderExpr).
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
