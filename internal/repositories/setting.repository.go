package repositories

import (
	"my-go-app/internal/models"

	"gorm.io/gorm"
)

type SettingRepository struct {
	DB *gorm.DB
}

func (r *SettingRepository) GetSettingByElementID(db *gorm.DB, elementID string) (*models.Setting, error) {
	if db == nil {
		db = r.DB
	}
	var setting models.Setting
	if err := db.Table(TableSetting.String()).Where(`"ElementId" = ?`, elementID).First(&setting).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &setting, nil
}

func (r *SettingRepository) CreateSetting(db *gorm.DB, setting models.Setting) error {
	if db == nil {
		db = r.DB
	}
	return db.Table(TableSetting.String()).Create(&setting).Error
}

func (r *SettingRepository) UpdateSetting(db *gorm.DB, setting models.Setting) error {
	if db == nil {
		db = r.DB
	}
	return db.Table(TableSetting.String()).Where(`"ElementId" = ?`, setting.ElementId).Updates(&setting).Error
}

func (r *SettingRepository) DeleteSetting(db *gorm.DB, elementID string) error {
	if db == nil {
		db = r.DB
	}
	return db.Table(TableSetting.String()).Where(`"ElementId" = ?`, elementID).Delete(&models.Setting{}).Error
}

func (r *SettingRepository) GetSettingsByElementIDs(db *gorm.DB, elementIDs []string) ([]models.Setting, error) {
	if db == nil {
		db = r.DB
	}
	var settings []models.Setting
	if err := db.Table(TableSetting.String()).Where(`"ElementId" IN (?)`, elementIDs).Find(&settings).Error; err != nil {
		return nil, err
	}
	return settings, nil
}

func (r *SettingRepository) CreateSettings(db *gorm.DB, settings []models.Setting) error {
	if db == nil {
		db = r.DB
	}
	if len(settings) == 0 {
		return nil
	}
	return db.Table(TableSetting.String()).CreateInBatches(settings, 500).Error
}

func (r *SettingRepository) UpdateSettings(db *gorm.DB, settings []models.Setting) error {
	if db == nil {
		db = r.DB
	}
	for _, setting := range settings {
		if err := db.Table(TableSetting.String()).Where(`"ElementId" = ?`, setting.ElementId).Updates(&setting).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *SettingRepository) DeleteSettings(db *gorm.DB, elementIDs []string) error {
	if db == nil {
		db = r.DB
	}
	return db.Table(TableSetting.String()).Where(`"ElementId" IN (?)`, elementIDs).Delete(&models.Setting{}).Error
}
