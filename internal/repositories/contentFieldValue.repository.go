package repositories

import (
	"context"
	"errors"
	"fmt"
	"my-go-app/internal/models"

	"github.com/lucsky/cuid"
	"gorm.io/gorm"
)

var (
	ErrContentFieldValueNotFound = errors.New("content field value not found")
)

type ContentFieldValueRepository struct {
	db *gorm.DB
}

func NewContentFieldValueRepository(db *gorm.DB) ContentFieldValueRepositoryInterface {
	return &ContentFieldValueRepository{db: db}
}

func (r *ContentFieldValueRepository) GetContentFieldValuesByContentItem(ctx context.Context, contentItemID string) ([]models.ContentFieldValue, error) {
	if contentItemID == "" {
		return nil, errors.New("contentItemID is required")
	}

	var contentFieldValues []models.ContentFieldValue

	err := r.db.WithContext(ctx).
		Model(&models.ContentFieldValue{}).
		Where("\"ContentItemId\" = ?", contentItemID).
		Find(&contentFieldValues).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get content field values by content item: %w", err)
	}

	return contentFieldValues, nil
}

func (r *ContentFieldValueRepository) GetContentFieldValueByID(ctx context.Context, id string) (*models.ContentFieldValue, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}

	var contentFieldValue models.ContentFieldValue

	err := r.db.WithContext(ctx).
		Model(&models.ContentFieldValue{}).
		Where("\"Id\" = ?", id).
		First(&contentFieldValue).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrContentFieldValueNotFound
		}
		return nil, fmt.Errorf("failed to get content field value: %w", err)
	}

	return &contentFieldValue, nil
}

func (r *ContentFieldValueRepository) CreateContentFieldValue(ctx context.Context, cfv *models.ContentFieldValue) (*models.ContentFieldValue, error) {
	if cfv == nil {
		return nil, errors.New("content field value cannot be nil")
	}

	if cfv.ContentItemId == "" {
		return nil, errors.New("content item ID is required")
	}

	if cfv.FieldId == "" {
		return nil, errors.New("field ID is required")
	}

	// Generate ID if not provided
	if cfv.Id == "" {
		cfv.Id = cuid.New()
	}

	err := r.db.WithContext(ctx).
		Model(&models.ContentFieldValue{}).
		Create(cfv).Error

	if err != nil {
		return nil, fmt.Errorf("failed to create content field value: %w", err)
	}

	return cfv, nil
}

func (r *ContentFieldValueRepository) UpdateContentFieldValue(ctx context.Context, id string, value *string) (*models.ContentFieldValue, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}

	result := r.db.WithContext(ctx).
		Model(&models.ContentFieldValue{}).
		Where("\"Id\" = ?", id).
		Update("\"Value\"", value)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to update content field value: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return nil, ErrContentFieldValueNotFound
	}

	// Fetch and return updated value
	var contentFieldValue models.ContentFieldValue
	err := r.db.WithContext(ctx).
		Model(&models.ContentFieldValue{}).
		Where("\"Id\" = ?", id).
		First(&contentFieldValue).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get updated content field value: %w", err)
	}

	return &contentFieldValue, nil
}

func (r *ContentFieldValueRepository) DeleteContentFieldValue(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id is required")
	}

	result := r.db.WithContext(ctx).
		Where("\"Id\" = ?", id).
		Delete(&models.ContentFieldValue{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete content field value: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrContentFieldValueNotFound
	}

	return nil
}
