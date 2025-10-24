package repositories

import (
	"context"
	"errors"
	"fmt"
	"my-go-app/internal/models"

	"gorm.io/gorm"
)

var (
	ErrCustomElementTypeNotFound      = errors.New("custom element type not found")
	ErrCustomElementTypeAlreadyExists = errors.New("custom element type with this name already exists")
)

type CustomElementTypeRepository struct {
	db *gorm.DB
}

func NewCustomElementTypeRepository(db *gorm.DB) CustomElementTypeRepositoryInterface {
	return &CustomElementTypeRepository{
		db: db,
	}
}

func (r *CustomElementTypeRepository) GetCustomElementTypes(ctx context.Context) ([]models.CustomElementType, error) {
	var customElementTypes []models.CustomElementType

	err := r.db.WithContext(ctx).
		Model(&models.CustomElementType{}).
		Order("\"Name\" ASC").
		Find(&customElementTypes).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get custom element types: %w", err)
	}

	return customElementTypes, nil
}

func (r *CustomElementTypeRepository) GetCustomElementTypeByID(ctx context.Context, id string) (*models.CustomElementType, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}

	var customElementType models.CustomElementType

	err := r.db.WithContext(ctx).
		Where("\"Id\" = ?", id).
		First(&customElementType).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomElementTypeNotFound
		}
		return nil, fmt.Errorf("failed to get custom element type: %w", err)
	}

	return &customElementType, nil
}

func (r *CustomElementTypeRepository) GetCustomElementTypeByName(ctx context.Context, name string) (*models.CustomElementType, error) {
	if name == "" {
		return nil, errors.New("name is required")
	}

	var customElementType models.CustomElementType

	err := r.db.WithContext(ctx).
		Where("\"Name\" = ?", name).
		First(&customElementType).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomElementTypeNotFound
		}
		return nil, fmt.Errorf("failed to get custom element type: %w", err)
	}

	return &customElementType, nil
}

func (r *CustomElementTypeRepository) CreateCustomElementType(ctx context.Context, customElementType *models.CustomElementType) (*models.CustomElementType, error) {
	if customElementType == nil {
		return nil, errors.New("customElementType is required")
	}
	if customElementType.Name == "" {
		return nil, errors.New("name is required")
	}

	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.CustomElementType{}).
		Where("\"Name\" = ?", customElementType.Name).
		Count(&count).Error

	if err != nil {
		return nil, fmt.Errorf("failed to check for existing custom element type: %w", err)
	}

	if count > 0 {
		return nil, ErrCustomElementTypeAlreadyExists
	}

	err = r.db.WithContext(ctx).Create(customElementType).Error
	if err != nil {
		return nil, fmt.Errorf("failed to create custom element type: %w", err)
	}

	return customElementType, nil
}

func (r *CustomElementTypeRepository) UpdateCustomElementType(ctx context.Context, id string, updates map[string]any) (*models.CustomElementType, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}

	var customElementType models.CustomElementType
	err := r.db.WithContext(ctx).
		Where("\"Id\" = ?", id).
		First(&customElementType).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomElementTypeNotFound
		}
		return nil, fmt.Errorf("failed to find custom element type: %w", err)
	}

	err = r.db.WithContext(ctx).
		Model(&customElementType).
		Updates(updates).Error

	if err != nil {
		return nil, fmt.Errorf("failed to update custom element type: %w", err)
	}

	err = r.db.WithContext(ctx).
		Where("\"Id\" = ?", id).
		First(&customElementType).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch updated custom element type: %w", err)
	}

	return &customElementType, nil
}

func (r *CustomElementTypeRepository) DeleteCustomElementType(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id is required")
	}

	result := r.db.WithContext(ctx).
		Where("\"Id\" = ?", id).
		Delete(&models.CustomElementType{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete custom element type: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrCustomElementTypeNotFound
	}

	return nil
}
