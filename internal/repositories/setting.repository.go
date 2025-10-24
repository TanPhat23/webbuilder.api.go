package repositories

import (
	"context"
	"errors"
	"fmt"
	"my-go-app/internal/models"

	"gorm.io/gorm"
)

var (
	ErrSettingNotFound = errors.New("setting not found")
)

type SettingRepository struct {
	db *gorm.DB
}

func NewSettingRepository(db *gorm.DB) SettingRepositoryInterface {
	return &SettingRepository{db: db}
}

func (r *SettingRepository) GetSettingByElementID(ctx context.Context, db *gorm.DB, elementID string) (*models.Setting, error) {
	if elementID == "" {
		return nil, errors.New("elementID is required")
	}

	if db == nil {
		db = r.db
	}

	var setting models.Setting

	err := db.WithContext(ctx).
		Model(&models.Setting{}).
		Where("\"ElementId\" = ?", elementID).
		First(&setting).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSettingNotFound
		}
		return nil, fmt.Errorf("failed to get setting by element ID: %w", err)
	}

	return &setting, nil
}

func (r *SettingRepository) GetSettingsByElementIDs(ctx context.Context, db *gorm.DB, elementIDs []string) ([]models.Setting, error) {
	if len(elementIDs) == 0 {
		return []models.Setting{}, nil
	}

	if db == nil {
		db = r.db
	}

	var settings []models.Setting

	err := db.WithContext(ctx).
		Model(&models.Setting{}).
		Where("\"ElementId\" IN ?", elementIDs).
		Find(&settings).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get settings by element IDs: %w", err)
	}

	return settings, nil
}

func (r *SettingRepository) CreateSetting(ctx context.Context, db *gorm.DB, setting *models.Setting) error {
	if setting == nil {
		return errors.New("setting cannot be nil")
	}

	if setting.ElementId == "" {
		return errors.New("element ID is required")
	}

	if db == nil {
		db = r.db
	}

	err := db.WithContext(ctx).
		Model(&models.Setting{}).
		Create(setting).Error

	if err != nil {
		return fmt.Errorf("failed to create setting: %w", err)
	}

	return nil
}

func (r *SettingRepository) CreateSettings(ctx context.Context, db *gorm.DB, settings []models.Setting) error {
	if len(settings) == 0 {
		return nil
	}

	if db == nil {
		db = r.db
	}

	// Validate all settings before insertion
	for i, setting := range settings {
		if setting.ElementId == "" {
			return fmt.Errorf("setting at index %d has empty element ID", i)
		}
	}

	err := db.WithContext(ctx).
		Model(&models.Setting{}).
		CreateInBatches(settings, DefaultBatchSize).Error

	if err != nil {
		return fmt.Errorf("failed to create settings: %w", err)
	}

	return nil
}

func (r *SettingRepository) UpdateSetting(ctx context.Context, db *gorm.DB, setting *models.Setting) error {
	if setting == nil {
		return errors.New("setting cannot be nil")
	}

	if setting.ElementId == "" {
		return errors.New("element ID is required")
	}

	if db == nil {
		db = r.db
	}

	result := db.WithContext(ctx).
		Model(&models.Setting{}).
		Where("\"ElementId\" = ?", setting.ElementId).
		Updates(setting)

	if result.Error != nil {
		return fmt.Errorf("failed to update setting: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrSettingNotFound
	}

	return nil
}

func (r *SettingRepository) UpdateSettings(ctx context.Context, db *gorm.DB, settings []models.Setting) error {
	if len(settings) == 0 {
		return nil
	}

	if db == nil {
		db = r.db
	}

	// Update each setting individually within a transaction
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, setting := range settings {
			if setting.ElementId == "" {
				return fmt.Errorf("setting at index %d has empty element ID", i)
			}

			result := tx.Model(&models.Setting{}).
				Where("\"ElementId\" = ?", setting.ElementId).
				Updates(&setting)

			if result.Error != nil {
				return fmt.Errorf("failed to update setting at index %d: %w", i, result.Error)
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to update settings: %w", err)
	}

	return nil
}

func (r *SettingRepository) DeleteSetting(ctx context.Context, db *gorm.DB, elementID string) error {
	if elementID == "" {
		return errors.New("elementID is required")
	}

	if db == nil {
		db = r.db
	}

	result := db.WithContext(ctx).
		Where("\"ElementId\" = ?", elementID).
		Delete(&models.Setting{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete setting: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrSettingNotFound
	}

	return nil
}

func (r *SettingRepository) DeleteSettings(ctx context.Context, db *gorm.DB, elementIDs []string) error {
	if len(elementIDs) == 0 {
		return nil
	}

	if db == nil {
		db = r.db
	}

	result := db.WithContext(ctx).
		Where("\"ElementId\" IN ?", elementIDs).
		Delete(&models.Setting{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete settings: %w", result.Error)
	}

	return nil
}

func (r *SettingRepository) ExistsByElementID(ctx context.Context, elementID string) (bool, error) {
	if elementID == "" {
		return false, errors.New("elementID is required")
	}

	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Setting{}).
		Where("\"ElementId\" = ?", elementID).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("failed to check setting existence: %w", err)
	}

	return count > 0, nil
}

func (r *SettingRepository) UpsertSetting(ctx context.Context, db *gorm.DB, setting *models.Setting) error {
	if setting == nil {
		return errors.New("setting cannot be nil")
	}

	if setting.ElementId == "" {
		return errors.New("element ID is required")
	}

	if db == nil {
		db = r.db
	}

	// Check if setting exists
	existing, err := r.GetSettingByElementID(ctx, db, setting.ElementId)
	if err != nil && !errors.Is(err, ErrSettingNotFound) {
		return fmt.Errorf("failed to check existing setting: %w", err)
	}

	if existing != nil {
		// Update existing setting
		return r.UpdateSetting(ctx, db, setting)
	}

	// Create new setting
	return r.CreateSetting(ctx, db, setting)
}
