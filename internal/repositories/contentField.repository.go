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
	ErrContentFieldNotFound = errors.New("content field not found")
	ErrContentFieldDuplicate = errors.New("content field with this name already exists for this content type")
)

type ContentFieldRepository struct {
	db *gorm.DB
}

func NewContentFieldRepository(db *gorm.DB) ContentFieldRepositoryInterface {
	return &ContentFieldRepository{db: db}
}

func (r *ContentFieldRepository) GetContentFieldsByContentType(ctx context.Context, contentTypeID string) ([]models.ContentField, error) {
	if contentTypeID == "" {
		return nil, errors.New("contentTypeID is required")
	}

	var contentFields []models.ContentField

	err := r.db.WithContext(ctx).
		Model(&models.ContentField{}).
		Where("\"ContentTypeId\" = ?", contentTypeID).
		Order("\"Name\" ASC").
		Find(&contentFields).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get content fields by content type: %w", err)
	}

	return contentFields, nil
}

func (r *ContentFieldRepository) GetContentFieldByID(ctx context.Context, id string) (*models.ContentField, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}

	var contentField models.ContentField

	err := r.db.WithContext(ctx).
		Model(&models.ContentField{}).
		Where("\"Id\" = ?", id).
		First(&contentField).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrContentFieldNotFound
		}
		return nil, fmt.Errorf("failed to get content field: %w", err)
	}

	return &contentField, nil
}

func (r *ContentFieldRepository) CreateContentField(ctx context.Context, cf *models.ContentField) (*models.ContentField, error) {
	var err error
	if cf == nil {
		return nil, errors.New("content field cannot be nil")
	}

	if cf.Name == "" {
		return nil, errors.New("content field name is required")
	}

	if cf.ContentTypeId == "" {
		return nil, errors.New("content type ID is required")
	}

	if cf.Type == "" {
		return nil, errors.New("content field type is required")
	}

	// Check for duplicate name in the same content type
	var count int64
	err = r.db.WithContext(ctx).
		Model(&models.ContentField{}).
		Where("\"ContentTypeId\" = ? AND \"Name\" = ?", cf.ContentTypeId, cf.Name).
		Count(&count).Error

	if err != nil {
		return nil, fmt.Errorf("failed to check content field existence: %w", err)
	}

	if count > 0 {
		return nil, ErrContentFieldDuplicate
	}

	// Generate ID if not provided
	if cf.Id == "" {
		cf.Id = cuid.New()
	}

	err = r.db.WithContext(ctx).
		Model(&models.ContentField{}).
		Create(cf).Error

	if err != nil {
		return nil, fmt.Errorf("failed to create content field: %w", err)
	}

	return cf, nil
}

func (r *ContentFieldRepository) UpdateContentField(ctx context.Context, id string, updates map[string]any) (*models.ContentField, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}

	if len(updates) == 0 {
		return nil, errors.New("no updates provided")
	}

	// Check if content field exists
	existing, err := r.GetContentFieldByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// If updating name, check for duplicates
	if newName, ok := updates["name"].(string); ok {
		var count int64
		err = r.db.WithContext(ctx).
			Model(&models.ContentField{}).
			Where("\"ContentTypeId\" = ? AND \"Name\" = ? AND \"Id\" != ?", existing.ContentTypeId, newName, id).
			Count(&count).Error

		if err != nil {
			return nil, fmt.Errorf("failed to check name uniqueness: %w", err)
		}

		if count > 0 {
			return nil, ErrContentFieldDuplicate
		}
	}

	// Perform update
	err = r.db.WithContext(ctx).
		Model(&models.ContentField{}).
		Where("\"Id\" = ?", id).
		Updates(updates).Error

	if err != nil {
		return nil, fmt.Errorf("failed to update content field: %w", err)
	}

	// Fetch and return updated content field
	var contentField models.ContentField
	err = r.db.WithContext(ctx).
		Model(&models.ContentField{}).
		Where("\"Id\" = ?", id).
		First(&contentField).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get updated content field: %w", err)
	}

	return &contentField, nil
}

func (r *ContentFieldRepository) DeleteContentField(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id is required")
	}

	result := r.db.WithContext(ctx).
		Where("\"Id\" = ?", id).
		Delete(&models.ContentField{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete content field: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrContentFieldNotFound
	}

	return nil
}
