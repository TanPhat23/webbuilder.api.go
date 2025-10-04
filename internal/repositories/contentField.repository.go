package repositories

import (
	"my-go-app/internal/models"

	"github.com/lucsky/cuid"
	"gorm.io/gorm"
)

type ContentFieldRepository struct {
	db *gorm.DB
}

func NewContentFieldRepository(db *gorm.DB) ContentFieldRepositoryInterface {
	return &ContentFieldRepository{db: db}
}

func (r *ContentFieldRepository) GetContentFieldsByContentType(contentTypeId string) ([]models.ContentField, error) {
	var contentFields []models.ContentField
	err := r.db.Table(TableContentField.String()).Where("\"ContentTypeId\" = ?", contentTypeId).Find(&contentFields).Error
	return contentFields, err
}

func (r *ContentFieldRepository) GetContentFieldByID(id string) (*models.ContentField, error) {
	var contentField models.ContentField
	err := r.db.Table(TableContentField.String()).First(&contentField, "\"Id\" = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &contentField, nil
}

func (r *ContentFieldRepository) CreateContentField(cf models.ContentField) (*models.ContentField, error) {
	cf.Id = cuid.New()
	err := r.db.Table(TableContentField.String()).Create(&cf).Error
	if err != nil {
		return nil, err
	}
	return &cf, nil
}

func (r *ContentFieldRepository) UpdateContentField(id string, updates map[string]any) (*models.ContentField, error) {
	var contentField models.ContentField
	err := r.db.Table(TableContentField.String()).Model(&contentField).Where("\"Id\" = ?", id).Updates(updates).Error
	if err != nil {
		return nil, err
	}
	err = r.db.Table(TableContentField.String()).First(&contentField, "\"Id\" = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &contentField, nil
}

func (r *ContentFieldRepository) DeleteContentField(id string) error {
	return r.db.Table(TableContentField.String()).Delete(&models.ContentField{}, "\"Id\" = ?", id).Error
}
