package repositories

import (
	"my-go-app/internal/models"

	"github.com/lucsky/cuid"
	"gorm.io/gorm"
)

type ContentFieldValueRepository struct {
	db *gorm.DB
}

func NewContentFieldValueRepository(db *gorm.DB) ContentFieldValueRepositoryInterface {
	return &ContentFieldValueRepository{db: db}
}

func (r *ContentFieldValueRepository) GetContentFieldValuesByContentItem(contentItemId string) ([]models.ContentFieldValue, error) {
	var contentFieldValues []models.ContentFieldValue
	err := r.db.Table(TableContentFieldValue.String()).Where("\"ContentItemId\" = ?", contentItemId).Find(&contentFieldValues).Error
	return contentFieldValues, err
}

func (r *ContentFieldValueRepository) GetContentFieldValueByID(id string) (*models.ContentFieldValue, error) {
	var contentFieldValue models.ContentFieldValue
	err := r.db.Table(TableContentFieldValue.String()).First(&contentFieldValue, "\"Id\" = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &contentFieldValue, nil
}

func (r *ContentFieldValueRepository) CreateContentFieldValue(cfv models.ContentFieldValue) (*models.ContentFieldValue, error) {
	cfv.Id = cuid.New()
	err := r.db.Table(TableContentFieldValue.String()).Create(&cfv).Error
	if err != nil {
		return nil, err
	}
	return &cfv, nil
}

func (r *ContentFieldValueRepository) UpdateContentFieldValue(id string, value *string) (*models.ContentFieldValue, error) {
	var contentFieldValue models.ContentFieldValue
	err := r.db.Table(TableContentFieldValue.String()).Model(&contentFieldValue).Where("\"Id\" = ?", id).Update("value", value).Error
	if err != nil {
		return nil, err
	}
	err = r.db.Table(TableContentFieldValue.String()).First(&contentFieldValue, "\"Id\" = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &contentFieldValue, nil
}

func (r *ContentFieldValueRepository) DeleteContentFieldValue(id string) error {
	return r.db.Table(TableContentFieldValue.String()).Delete(&models.ContentFieldValue{}, "\"Id\" = ?", id).Error
}
