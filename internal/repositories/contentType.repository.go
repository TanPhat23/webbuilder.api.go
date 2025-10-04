package repositories

import (
	"my-go-app/internal/models"

	"github.com/lucsky/cuid"
	"gorm.io/gorm"
)

type ContentTypeRepository struct {
	db *gorm.DB
}

func NewContentTypeRepository(db *gorm.DB) ContentTypeRepositoryInterface {
	return &ContentTypeRepository{db: db}
}

func (r *ContentTypeRepository) GetContentTypes() ([]models.ContentType, error) {
	var contentTypes []models.ContentType
	err := r.db.Table(TableContentType.String()).Find(&contentTypes).Error
	return contentTypes, err
}

func (r *ContentTypeRepository) GetContentTypeByID(id string) (*models.ContentType, error) {
	var contentType models.ContentType
	err := r.db.Table(TableContentType.String()).Preload("Fields").Preload("Items").First(&contentType, "\"Id\" = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &contentType, nil
}

func (r *ContentTypeRepository) CreateContentType(ct models.ContentType) (*models.ContentType, error) {
	ct.Id = cuid.New()
	err := r.db.Table(TableContentType.String()).Create(&ct).Error
	if err != nil {
		return nil, err
	}
	return &ct, nil
}

func (r *ContentTypeRepository) UpdateContentType(id string, updates map[string]any) (*models.ContentType, error) {
	var contentType models.ContentType
	err := r.db.Table(TableContentType.String()).Model(&contentType).Where("\"Id\" = ?", id).Updates(updates).Error
	if err != nil {
		return nil, err
	}
	err = r.db.Table(TableContentType.String()).First(&contentType, "\"Id\" = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &contentType, nil
}

func (r *ContentTypeRepository) DeleteContentType(id string) error {
	return r.db.Table(TableContentType.String()).Delete(&models.ContentType{}, "\"Id\" = ?", id).Error
}
