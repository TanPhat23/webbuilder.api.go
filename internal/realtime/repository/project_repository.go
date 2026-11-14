package repository

import (
	"gorm.io/gorm"

	domainEntity "my-go-app/internal/realtime/domain/entity"
	domainRepo "my-go-app/internal/realtime/domain/repository"
)

type projectRepository struct {
	db *gorm.DB
}

// NewProjectRepository creates a new instance of ProjectRepository
func NewProjectRepository(db *gorm.DB) domainRepo.ProjectRepository {
	return &projectRepository{
		db: db,
	}
}

func (r *projectRepository) CreateProject(project *domainEntity.Project) error {
	if err := r.db.Create(project).Error; err != nil {
		return err
	}
	return nil
}

func (r *projectRepository) GetProjectByID(projectID string) (*domainEntity.Project, error) {
	var project domainEntity.Project
	if err := r.db.Model(&domainEntity.Project{}).First(&project, "id = ?", projectID).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *projectRepository) UpdateProject(project *domainEntity.Project) error {
	if err := r.db.Save(project).Error; err != nil {
		return err
	}
	return nil
}

func (r *projectRepository) DeleteProject(projectID string) error {
	if err := r.db.Delete(&domainEntity.Project{}, "id = ?", projectID).Error; err != nil {
		return err
	}
	return nil
}
