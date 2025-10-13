package repositories

import (
	"my-go-app/internal/models"
	"strings"

	"github.com/lucsky/cuid"
	"gorm.io/gorm"
)

type ContentItemRepository struct {
	db *gorm.DB
}

func NewContentItemRepository(db *gorm.DB) ContentItemRepositoryInterface {
	return &ContentItemRepository{db: db}
}

func (r *ContentItemRepository) GetContentItemsByContentType(contentTypeId string) ([]models.ContentItem, error) {
	var contentItems []models.ContentItem
	err := r.db.Table(TableContentItem.String()).Where("\"ContentTypeId\" = ?", contentTypeId).Preload("FieldValues").Find(&contentItems).Error
	return contentItems, err
}

func (r *ContentItemRepository) GetContentItemByID(id string) (*models.ContentItem, error) {
	var contentItem models.ContentItem
	err := r.db.Table(TableContentItem.String()).Preload("FieldValues").First(&contentItem, "\"Id\" = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &contentItem, nil
}

func (r *ContentItemRepository) CreateContentItem(ci models.ContentItem, fieldValues []models.ContentFieldValue) (*models.ContentItem, error) {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		ci.Id = cuid.New()
		if err := tx.Table(TableContentItem.String()).Create(&ci).Error; err != nil {
			return err
		}
		for _, fv := range fieldValues {
			fv.Id = cuid.New()
			fv.ContentItemId = ci.Id
			if err := tx.Table(TableContentFieldValue.String()).Create(&fv).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	err = r.db.Table(TableContentItem.String()).Preload("FieldValues").First(&ci, "\"Id\" = ?", ci.Id).Error
	if err != nil {
		return nil, err
	}
	return &ci, nil
}

func (r *ContentItemRepository) UpdateContentItem(id string, updates map[string]any) (*models.ContentItem, error) {
	var contentItem models.ContentItem
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if fvSlice, ok := updates["fieldValues"].([]interface{}); ok {
			if err := tx.Table(TableContentFieldValue.String()).Where("\"ContentItemId\" = ?", id).Delete(&models.ContentFieldValue{}).Error; err != nil {
				return err
			}
			for _, fv := range fvSlice {
				if fvMap, ok := fv.(map[string]interface{}); ok {
					fieldId, fidOk := fvMap["fieldId"].(string)
					value, valOk := fvMap["value"].(string)
					if fidOk && valOk {
						cfv := models.ContentFieldValue{
							Id:            cuid.New(),
							ContentItemId: id,
							FieldId:       fieldId,
							Value:         &value,
						}
						if err := tx.Table(TableContentFieldValue.String()).Create(&cfv).Error; err != nil {
							return err
						}
					}
				}
			}
			delete(updates, "fieldValues")
		}

		if err := tx.Table(TableContentItem.String()).Model(&contentItem).Where("\"Id\" = ?", id).Updates(updates).Error; err != nil {
			return err
		}

		return tx.Table(TableContentItem.String()).Preload("FieldValues").First(&contentItem, "\"Id\" = ?", id).Error
	})
	if err != nil {
		return nil, err
	}
	return &contentItem, nil
}

func (r *ContentItemRepository) DeleteContentItem(id string) error {
	return r.db.Table(TableContentItem.String()).Delete(&models.ContentItem{}, "\"Id\" = ?", id).Error
}

func (r *ContentItemRepository) GetPublicContentItems(contentTypeId string, limit int, sortBy string, sortOrder string) ([]models.ContentItem, error) {
	if sortOrder != "asc" {
		sortOrder = "desc"
	}

	validSortBy := map[string]bool{
		"CreatedAt": true,
		"UpdatedAt": true,
		"Title":     true,
	}
	if !validSortBy[sortBy] {
		sortBy = "CreatedAt"
	}

	sql := `SELECT * FROM ` + TableContentItem.String()
	args := []interface{}{}
	conditions := []string{}

	if contentTypeId != "" {
		conditions = append(conditions, `"ContentTypeId" = ?`)
		args = append(args, contentTypeId)
	}

	// Temporarily remove Published filter for debugging
	// conditions = append(conditions, `"Published" = true`)

	if len(conditions) > 0 {
		sql += ` WHERE ` + strings.Join(conditions, " AND ")
	}

	sql += ` ORDER BY "` + sortBy + `" ` + sortOrder + ` LIMIT ?`
	args = append(args, limit)

	var contentItems []models.ContentItem
	err := r.db.Raw(sql, args...).Scan(&contentItems).Error
	if err != nil {
		return nil, err
	}

	// Load FieldValues separately
	for i := range contentItems {
		err = r.db.Model(&models.ContentFieldValue{}).Where("\"ContentItemId\" = ?", contentItems[i].Id).Preload("Field").Find(&contentItems[i].FieldValues).Error
		if err != nil {
			return nil, err
		}
	}

	return contentItems, nil
}

func (r *ContentItemRepository) GetContentItemBySlug(contentTypeId string, slug string) (*models.ContentItem, error) {
	var contentItem models.ContentItem
	err := r.db.Table(TableContentItem.String()).Where("\"ContentTypeId\" = ? AND \"Slug\" = ?", contentTypeId, slug).Preload("FieldValues").First(&contentItem).Error
	if err != nil {
		return nil, err
	}
	return &contentItem, nil
}
