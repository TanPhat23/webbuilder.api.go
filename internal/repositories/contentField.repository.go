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
	ErrContentFieldNotFound  = errors.New("content field not found")
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

	contentFields, err := gorm.G[models.ContentField](r.db).
		Where("\"ContentTypeId\" = ?", contentTypeID).
		Order("\"Name\" ASC").
		Find(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get content fields by content type: %w", err)
	}

	return contentFields, nil
}

func (r *ContentFieldRepository) GetContentFieldByID(ctx context.Context, id string) (*models.ContentField, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}

	contentField, err := gorm.G[models.ContentField](r.db).
		Where("\"Id\" = ?", id).
		First(ctx)

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

	// Check for duplicate name in the same content type
	count, err := gorm.G[models.ContentField](r.db).
		Where("\"ContentTypeId\" = ? AND \"Name\" = ?", cf.ContentTypeId, cf.Name).
		Count(ctx, "*")

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

	err = gorm.G[models.ContentField](r.db).Create(ctx, cf)

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
		count, err := gorm.G[models.ContentField](r.db).
			Where("\"ContentTypeId\" = ? AND \"Name\" = ? AND \"Id\" != ?", existing.ContentTypeId, newName, id).
			Count(ctx, "*")

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
	contentField, err := gorm.G[models.ContentField](r.db).
		Where("\"Id\" = ?", id).
		First(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get updated content field: %w", err)
	}

	return &contentField, nil
}

func (r *ContentFieldRepository) DeleteContentField(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id is required")
	}

	rowsAffected, err := gorm.G[models.ContentField](r.db).
		Where("\"Id\" = ?", id).
		Delete(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete content field: %w", err)
	}

	if rowsAffected == 0 {
		return ErrContentFieldNotFound
	}

	return nil
}
