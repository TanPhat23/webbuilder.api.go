package repositories

import (
	"my-go-app/internal/models"

	"gorm.io/gorm"
)

type ProjectRepository struct {
	DB *gorm.DB
}

func (r *ProjectRepository) GetProjects() ([]models.Project, error) {
	var projects []models.Project
	if err := r.DB.Table((models.Project{}).GetTable()).Find(&projects).Error; err != nil {
		return nil, err
	}
	return projects, nil
}

func (r *ProjectRepository) GetProjectByID(projectID string, userId string) (*models.Project, error) {
	var project models.Project
	if err := r.DB.Table((models.Project{}).GetTable()).Where(`"Id" = ? AND "OwnerId" = ?`, projectID, userId).First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &project, nil
}

func (r *ProjectRepository) GetProjectsByUserID(userID string) ([]models.Project, error) {
	var projects []models.Project
	if err := r.DB.Table((models.Project{}).GetTable()).Where(`"OwnerId" = ? AND "DeletedAt" IS NULL`, userID).Find(&projects).Error; err != nil {
		return nil, err
	}
	return projects, nil
}

func (r *ProjectRepository) GetProjectPages(projectID string, userId string) ([]models.Page, error) {
	var pages []models.Page

	err := r.DB.Table(`public."Page" AS p`).
		Joins(`LEFT JOIN public."Project" AS pr ON p."ProjectId" = pr."Id"`).
		Where(`p."ProjectId" = ? AND pr."OwnerId" = ? `, projectID, userId).
		Find(&pages).Error

	if err != nil {
		return nil, err
	}
	return pages, nil
}
