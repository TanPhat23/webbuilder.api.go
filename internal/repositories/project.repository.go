package repositories

import (
	"my-go-app/internal/models"

	"gorm.io/gorm"
)

type ProjectRepository struct {
	DB *gorm.DB
}

func (r *ProjectRepository) CreateProject(project models.Project) (*models.Project, error) {
	if err := r.DB.Table(TableProject.String()).Create(&project).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *ProjectRepository) GetProjects() ([]models.Project, error) {
	var projects []models.Project
	if err := r.DB.Table(TableProject.String()).Find(&projects).Error; err != nil {
		return nil, err
	}
	return projects, nil
}

func (r *ProjectRepository) GetProjectByID(projectID string, userId string) (*models.Project, error) {
	var project models.Project
	if err := r.DB.Table(TableProject.String()).Where(`"Id" = ? AND "OwnerId" = ?`, projectID, userId).First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &project, nil
}

func (r *ProjectRepository) GetPublicProjectByID(projectID string) (*models.Project, error) {
	var project models.Project
	if err := r.DB.Table(TableProject.String()).Where(`"Id" = ? AND "Published" = ?`, projectID, true).First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &project, nil
}

func (r *ProjectRepository) GetProjectsByUserID(userID string) ([]models.Project, error) {
	var projects []models.Project
	if err := r.DB.Table(TableProject.String()).Where(`"OwnerId" = ? AND "DeletedAt" IS NULL`, userID).Find(&projects).Error; err != nil {
		return nil, err
	}
	return projects, nil
}

func (r *ProjectRepository) GetProjectPages(projectID string, userId string) ([]models.Page, error) {
	var count int64
	if err := r.DB.Table(TableProject.String()).
		Where(`"Id" = ? AND "OwnerId" = ?`, projectID, userId).
		Count(&count).Error; err != nil {
		return nil, err
	}

	if count == 0 {
		return []models.Page{}, nil
	}

	var pages []models.Page
	if err := r.DB.Table(TablePage.String()).
		Where(`"ProjectId" = ?`, projectID).
		Find(&pages).Error; err != nil {
		return nil, err
	}

	return pages, nil
}

func (r *ProjectRepository) UpdateProject(projectID string, userID string, updates map[string]any) (*models.Project, error) {
	// First verify ownership
	var count int64
	if err := r.DB.Table(TableProject.String()).
		Where(`"Id" = ? AND "OwnerId" = ?`, projectID, userID).
		Count(&count).Error; err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, nil
	}

	// The handler already converts keys to column names, so use updates directly
	if err := r.DB.Table(TableProject.String()).
		Where(`"Id" = ? AND "OwnerId" = ?`, projectID, userID).
		Updates(updates).Error; err != nil {
		return nil, err
	}

	return r.GetProjectByID(projectID, userID)
}

func (r *ProjectRepository) DeleteProject(projectID string, userID string) error {
	// Soft delete by setting DeletedAt
	return r.DB.Table(TableProject.String()).
		Where(`"Id" = ? AND "OwnerId" = ?`, projectID, userID).
		Update("DeletedAt", "NOW()").Error
}
