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

	updatesMap := make(map[string]any)
	for k, v := range updates {
		switch k {
		case "name":
			updatesMap["Name"] = v
		case "description":
			updatesMap["Description"] = v
		case "styles":
			updatesMap["Styles"] = v
		case "header":
			updatesMap["Header"] = v
		case "published":
			updatesMap["Published"] = v
		case "subdomain":
			updatesMap["Subdomain"] = v
		default:
			updatesMap[k] = v
		}
	}

	if err := r.DB.Table(TableProject.String()).
		Where(`"Id" = ? AND "OwnerId" = ?`, projectID, userID).
		Updates(updatesMap).Error; err != nil {
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
