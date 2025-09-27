package repositories

import (
	"my-go-app/internal/models"

	"gorm.io/gorm"
)

type PageRepository struct {
	DB *gorm.DB
}

func (r *PageRepository) GetPagesByProjectID(projectID string) ([]models.Page, error) {
	var pages []models.Page
	if err := r.DB.Table(TablePage.String()).Where(`"ProjectId" = ? AND "DeletedAt" IS NULL`, projectID).Find(&pages).Error; err != nil {
		return nil, err
	}
	return pages, nil
}

func (r *PageRepository) GetPageByID(pageID string, projectID string) (*models.Page, error) {
	var page models.Page
	if err := r.DB.Table(TablePage.String()).Where(`"Id" = ? AND "ProjectId" = ? AND "DeletedAt" IS NULL`, pageID, projectID).First(&page).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &page, nil
}

func (r *PageRepository) CreatePage(page models.Page) error {
	return r.DB.Table(TablePage.String()).Create(&page).Error
}

func (r *PageRepository) UpdatePage(page models.Page) error {
	return r.DB.Table(TablePage.String()).Where(`"Id" = ?`, page.Id).Updates(&page).Error
}

func (r *PageRepository) DeletePage(pageID string) error {
	return r.DB.Table(TablePage.String()).Where(`"Id" = ?`, pageID).Update("DeletedAt", "NOW()").Error
}

func (r *PageRepository) DeletePageByProjectID(pageID string, projectID string, userID string) error {
	return r.DB.Table(TablePage.String()).Where(`"Id" = ?`, pageID).Update("DeletedAt", "NOW()").Error
}
