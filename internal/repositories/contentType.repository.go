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
	// ErrContentTypeNotFound is returned when a content type is not found
	ErrContentTypeNotFound = errors.New("content type not found")
	// ErrContentTypeDuplicate is returned when a content type with the same name already exists
	ErrContentTypeDuplicate = errors.New("content type with this name already exists")
)

type ContentTypeRepository struct {
	db *gorm.DB
}

func NewContentTypeRepository(db *gorm.DB) ContentTypeRepositoryInterface {
	return &ContentTypeRepository{db: db}
}

func (r *ContentTypeRepository) GetContentTypes(ctx context.Context) ([]models.ContentType, error) {
	contentTypes, err := gorm.G[models.ContentType](r.db).
		Order("\"CreatedAt\" DESC").
		Find(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get content types: %w", err)
	}

	return contentTypes, nil
}

// GetContentTypesByProject retrieves all content types belonging to projectID.
func (r *ContentTypeRepository) GetContentTypesByProject(ctx context.Context, projectID string) ([]models.ContentType, error) {
	if projectID == "" {
		return nil, errors.New("projectID is required")
	}

	contentTypes, err := gorm.G[models.ContentType](r.db).
		Where("\"ProjectId\" = ?", projectID).
		Order("\"CreatedAt\" DESC").
		Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get content types: %w", err)
	}
	return contentTypes, nil
}

// GetContentTypesByUser retrieves content types for all non-deleted projects userID can view.
func (r *ContentTypeRepository) GetContentTypesByUser(ctx context.Context, userID string) ([]models.ContentType, error) {
	if userID == "" {
		return nil, errors.New("userID is required")
	}

	var contentTypes []models.ContentType
	err := r.db.WithContext(ctx).
		Model(&models.ContentType{}).
		Joins(`JOIN "Project" ON "Project"."Id" = "ContentType"."ProjectId"`).
		Where(`"Project"."DeletedAt" IS NULL AND ("Project"."OwnerId" = ? OR EXISTS (SELECT 1 FROM "Collaborator" WHERE "Collaborator"."ProjectId" = "Project"."Id" AND "Collaborator"."UserId" = ?))`, userID, userID).
		Order(`"ContentType"."CreatedAt" DESC`).
		Find(&contentTypes).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get user content types: %w", err)
	}
	return contentTypes, nil
}

func (r *ContentTypeRepository) GetContentTypeByID(ctx context.Context, id string) (*models.ContentType, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}

	contentType, err := gorm.G[models.ContentType](r.db).
		Preload("Fields", nil).
		Preload("Items", nil).
		Where("\"Id\" = ?", id).
		First(ctx)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrContentTypeNotFound
		}
		return nil, fmt.Errorf("failed to get content type: %w", err)
	}

	return &contentType, nil
}

func (r *ContentTypeRepository) GetContentTypeByName(ctx context.Context, name string) (*models.ContentType, error) {
	if name == "" {
		return nil, errors.New("name is required")
	}

	contentType, err := gorm.G[models.ContentType](r.db).
		Where("\"Name\" = ?", name).
		First(ctx)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrContentTypeNotFound
		}
		return nil, fmt.Errorf("failed to get content type by name: %w", err)
	}

	return &contentType, nil
}

func (r *ContentTypeRepository) CreateContentType(ctx context.Context, ct *models.ContentType) (*models.ContentType, error) {
	if ct == nil {
		return nil, errors.New("content type cannot be nil")
	}

	if ct.Name == "" {
		return nil, errors.New("content type name is required")
	}

	// Check for duplicate name scoped by project when ProjectId is present.
	var exists bool
	var err error
	if ct.ProjectId != nil && *ct.ProjectId != "" {
		var count int64
		count, err = gorm.G[models.ContentType](r.db).
			Where(`"ProjectId" = ? AND "Name" = ?`, *ct.ProjectId, ct.Name).
			Count(ctx, "*")
		exists = count > 0
	} else {
		exists, err = r.ExistsByName(ctx, ct.Name)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to check content type existence: %w", err)
	}

	if exists {
		return nil, ErrContentTypeDuplicate
	}

	// Generate ID if not provided
	if ct.Id == "" {
		ct.Id = cuid.New()
	}

	// Set timestamps
	now := time.Now()
	ct.CreatedAt = now
	ct.UpdatedAt = now

	err = gorm.G[models.ContentType](r.db).Create(ctx, ct)

	if err != nil {
		return nil, fmt.Errorf("failed to create content type: %w", err)
	}

	return ct, nil
}

func (r *ContentTypeRepository) UpdateContentType(ctx context.Context, id string, updates map[string]any) (*models.ContentType, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}

	if len(updates) == 0 {
		return nil, errors.New("no updates provided")
	}

	// Check if content type exists
	count, err := gorm.G[models.ContentType](r.db).
		Where("\"Id\" = ?", id).
		Count(ctx, "*")

	if err != nil {
		return nil, fmt.Errorf("failed to check content type existence: %w", err)
	}

	if count == 0 {
		return nil, ErrContentTypeNotFound
	}

	normalizeUpdateColumns(updates)

	// If updating name, check for duplicates
	if newName, ok := updates["Name"].(string); ok {
		existingCount, err := gorm.G[models.ContentType](r.db).
			Where("\"Name\" = ? AND \"Id\" != ?", newName, id).
			Count(ctx, "*")

		if err != nil {
			return nil, fmt.Errorf("failed to check name uniqueness: %w", err)
		}

		if existingCount > 0 {
			return nil, ErrContentTypeDuplicate
		}
	}

	// Always update the UpdatedAt timestamp
	updates["UpdatedAt"] = time.Now()

	// Perform update
	err = r.db.WithContext(ctx).
		Model(&models.ContentType{}).
		Where("\"Id\" = ?", id).
		Updates(updates).Error

	if err != nil {
		return nil, fmt.Errorf("failed to update content type: %w", err)
	}

	// Fetch and return updated content type
	contentType, err := gorm.G[models.ContentType](r.db).
		Where("\"Id\" = ?", id).
		First(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get updated content type: %w", err)
	}

	return &contentType, nil
}

func (r *ContentTypeRepository) DeleteContentType(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id is required")
	}

	rowsAffected, err := gorm.G[models.ContentType](r.db).
		Where("\"Id\" = ?", id).
		Delete(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete content type: %w", err)
	}

	if rowsAffected == 0 {
		return ErrContentTypeNotFound
	}

	return nil
}

func (r *ContentTypeRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	if name == "" {
		return false, errors.New("name is required")
	}

	count, err := gorm.G[models.ContentType](r.db).
		Where("\"Name\" = ?", name).
		Count(ctx, "*")

	if err != nil {
		return false, fmt.Errorf("failed to check content type existence: %w", err)
	}

	return count > 0, nil
}

func (r *ContentTypeRepository) ExistsByID(ctx context.Context, id string) (bool, error) {
	if id == "" {
		return false, errors.New("id is required")
	}

	count, err := gorm.G[models.ContentType](r.db).
		Where("\"Id\" = ?", id).
		Count(ctx, "*")

	if err != nil {
		return false, fmt.Errorf("failed to check content type existence: %w", err)
	}

	return count > 0, nil
}

func (r *ContentTypeRepository) GetContentTypesWithFieldCount(ctx context.Context) ([]map[string]any, error) {
	var results []map[string]any

	err := r.db.WithContext(ctx).
		Model(&models.ContentType{}).
		Select(`public."ContentType".*, COUNT(public."ContentField"."Id") as field_count`).
		Joins(`LEFT JOIN public."ContentField" ON public."ContentField"."ContentTypeId" = public."ContentType"."Id"`).
		Group(`public."ContentType"."Id"`).
		Order(`public."ContentType"."CreatedAt" DESC`).
		Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get content types with field count: %w", err)
	}

	return results, nil
}

func (r *ContentTypeRepository) GetContentTypesWithItemCount(ctx context.Context) ([]map[string]any, error) {
	var results []map[string]any

	err := r.db.WithContext(ctx).
		Model(&models.ContentType{}).
		Select(`public."ContentType".*, COUNT(public."ContentItem"."Id") as item_count`).
		Joins(`LEFT JOIN public."ContentItem" ON public."ContentItem"."ContentTypeId" = public."ContentType"."Id"`).
		Group(`public."ContentType"."Id"`).
		Order(`public."ContentType"."CreatedAt" DESC`).
		Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get content types with item count: %w", err)
	}

	return results, nil
}
