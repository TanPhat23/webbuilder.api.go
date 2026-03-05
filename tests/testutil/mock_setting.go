package testutil

import (
	"context"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"gorm.io/gorm"
)

type MockSettingRepository struct {
	GetSettingByElementIDFn   func(ctx context.Context, db *gorm.DB, elementID string) (*models.Setting, error)
	GetSettingsByElementIDsFn func(ctx context.Context, db *gorm.DB, elementIDs []string) ([]models.Setting, error)
	CreateSettingFn           func(ctx context.Context, db *gorm.DB, setting *models.Setting) error
	CreateSettingsFn          func(ctx context.Context, db *gorm.DB, settings []models.Setting) error
	UpdateSettingFn           func(ctx context.Context, db *gorm.DB, setting *models.Setting) error
	UpdateSettingsFn          func(ctx context.Context, db *gorm.DB, settings []models.Setting) error
	DeleteSettingFn           func(ctx context.Context, db *gorm.DB, elementID string) error
	DeleteSettingsFn          func(ctx context.Context, db *gorm.DB, elementIDs []string) error
}

func (m *MockSettingRepository) GetSettingByElementID(ctx context.Context, db *gorm.DB, elementID string) (*models.Setting, error) {
	if m.GetSettingByElementIDFn != nil {
		return m.GetSettingByElementIDFn(ctx, db, elementID)
	}
	return nil, repositories.ErrSettingNotFound
}

func (m *MockSettingRepository) GetSettingsByElementIDs(ctx context.Context, db *gorm.DB, elementIDs []string) ([]models.Setting, error) {
	if m.GetSettingsByElementIDsFn != nil {
		return m.GetSettingsByElementIDsFn(ctx, db, elementIDs)
	}
	return []models.Setting{}, nil
}

func (m *MockSettingRepository) CreateSetting(ctx context.Context, db *gorm.DB, setting *models.Setting) error {
	if m.CreateSettingFn != nil {
		return m.CreateSettingFn(ctx, db, setting)
	}
	return nil
}

func (m *MockSettingRepository) CreateSettings(ctx context.Context, db *gorm.DB, settings []models.Setting) error {
	if m.CreateSettingsFn != nil {
		return m.CreateSettingsFn(ctx, db, settings)
	}
	return nil
}

func (m *MockSettingRepository) UpdateSetting(ctx context.Context, db *gorm.DB, setting *models.Setting) error {
	if m.UpdateSettingFn != nil {
		return m.UpdateSettingFn(ctx, db, setting)
	}
	return nil
}

func (m *MockSettingRepository) UpdateSettings(ctx context.Context, db *gorm.DB, settings []models.Setting) error {
	if m.UpdateSettingsFn != nil {
		return m.UpdateSettingsFn(ctx, db, settings)
	}
	return nil
}

func (m *MockSettingRepository) DeleteSetting(ctx context.Context, db *gorm.DB, elementID string) error {
	if m.DeleteSettingFn != nil {
		return m.DeleteSettingFn(ctx, db, elementID)
	}
	return nil
}

func (m *MockSettingRepository) DeleteSettings(ctx context.Context, db *gorm.DB, elementIDs []string) error {
	if m.DeleteSettingsFn != nil {
		return m.DeleteSettingsFn(ctx, db, elementIDs)
	}
	return nil
}