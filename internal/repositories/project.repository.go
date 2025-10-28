package repositories

import (
	"context"
	"errors"
	"fmt"
	"my-go-app/internal/models"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrProjectNotFound = errors.New("project not found")
	ErrProjectUnauthorized = errors.New("unauthorized access to project")
)

type ProjectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) ProjectRepositoryInterface {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) GetProjects(ctx context.Context) ([]models.Project, error) {
	var projects []models.Project

	err := r.db.WithContext(ctx).
		Where(&models.Project{DeletedAt: nil}).
		Order("\"CreatedAt\" DESC").
		Find(&projects).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}

	return projects, nil
}

func (r *ProjectRepository) GetProjectByID(ctx context.Context, projectID, userID string) (*models.Project, error) {
	if projectID == "" || userID == "" {
		return nil, errors.New("projectID and userID are required")
	}

	var project models.Project

	err := r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("\"Id\" = ? AND \"OwnerId\" = ? AND \"DeletedAt\" IS NULL", projectID, userID).
		First(&project).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return &project, nil
}

func (r *ProjectRepository) GetProjectWithAccess(ctx context.Context, projectID, userID string) (*models.Project, error) {
	if projectID == "" || userID == "" {
		return nil, errors.New("projectID and userID are required")
	}

	var project models.Project

	// Check if user is owner
	err := r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("\"Id\" = ? AND \"OwnerId\" = ? AND \"DeletedAt\" IS NULL", projectID, userID).
		First(&project).Error

	if err == nil {
		return &project, nil
	}

	// If not owner, check if collaborator
	var collaborator models.Collaborator
	err = r.db.WithContext(ctx).
		Model(&models.Collaborator{}).
		Where("\"ProjectId\" = ? AND \"UserId\" = ?", projectID, userID).
		First(&collaborator).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectUnauthorized
		}
		return nil, fmt.Errorf("failed to check collaborator access: %w", err)
	}

	// Get the project
	err = r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("\"Id\" = ? AND \"DeletedAt\" IS NULL", projectID).
		First(&project).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return &project, nil
}

func (r *ProjectRepository) GetPublicProjectByID(ctx context.Context, projectID string) (*models.Project, error) {
	if projectID == "" {
		return nil, errors.New("projectID is required")
	}

	var project models.Project

	err := r.db.WithContext(ctx).
		Where("\"Id\" = ? AND \"Published\" = true AND \"DeletedAt\" IS NULL", projectID).
		First(&project).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get public project: %w", err)
	}

	return &project, nil
}

func (r *ProjectRepository) GetProjectsByUserID(ctx context.Context, userID string) ([]models.Project, error) {
	if userID == "" {
		return nil, errors.New("userID is required")
	}

	var projects []models.Project

	err := r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("\"OwnerId\" = ? AND \"DeletedAt\" IS NULL", userID).
		Order("\"CreatedAt\" DESC").
		Find(&projects).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get projects by user ID: %w", err)
	}

	return projects, nil
}

func (r *ProjectRepository) GetProjectPages(ctx context.Context, projectID, userID string) ([]models.Page, error) {
	if projectID == "" || userID == "" {
		return nil, errors.New("projectID and userID are required")
	}

	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where(&models.Project{ID: projectID, OwnerId: userID, DeletedAt: nil}).
		Count(&count).Error

	if err != nil {
		return nil, fmt.Errorf("failed to verify project ownership: %w", err)
	}

	if count == 0 {
		return nil, ErrProjectUnauthorized
	}

	var pages []models.Page
	err = r.db.WithContext(ctx).
		Where(&models.Page{ProjectId: projectID, DeletedAt: nil}).
		Order("\"CreatedAt\" ASC").
		Find(&pages).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get project pages: %w", err)
	}

	return pages, nil
}

func (r *ProjectRepository) CreateProject(ctx context.Context, project *models.Project) error {
	if project == nil {
		return errors.New("project cannot be nil")
	}

	if project.Name == "" {
		return errors.New("project name is required")
	}

	if project.OwnerId == "" {
		return errors.New("owner ID is required")
	}

	now := time.Now()
	project.CreatedAt = now
	project.UpdatedAt = now

	err := r.db.WithContext(ctx).
		Model(&models.Project{}).
		Create(project).Error

	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	return nil
}

func (r *ProjectRepository) UpdateProject(ctx context.Context, projectID, userID string, updates map[string]any) (*models.Project, error) {
	if projectID == "" || userID == "" {
		return nil, errors.New("projectID and userID are required")
	}

	if len(updates) == 0 {
		return nil, errors.New("no updates provided")
	}

	// Verify ownership first
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("\"Id\" = ? AND \"OwnerId\" = ? AND \"DeletedAt\" IS NULL", projectID, userID).
		Count(&count).Error

	if err != nil {
		return nil, fmt.Errorf("failed to verify project ownership: %w", err)
	}

	if count == 0 {
		return nil, ErrProjectUnauthorized
	}

	// Map JSON field names to database column names
	updateMap := make(map[string]any)
	for k, v := range updates {
		switch k {
		case "Name":
			updateMap["Name"] = v
		case "Description":
			updateMap["Description"] = v
		case "Styles":
			updateMap["Styles"] = v
		case "Header":
			updateMap["Header"] = v
		case "Published":
			updateMap["Published"] = v
		case "Subdomain":
			updateMap["Subdomain"] = v
		}
	}

	// Always update the UpdatedAt timestamp
	updateMap["UpdatedAt"] = time.Now()

	// Perform update
	err = r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("\"Id\" = ? AND \"OwnerId\" = ? AND \"DeletedAt\" IS NULL", projectID, userID).
		Updates(updateMap).Error

	if err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	// Fetch and return updated project
	return r.GetProjectByID(ctx, projectID, userID)
}

func (r *ProjectRepository) DeleteProject(ctx context.Context, projectID, userID string) error {
	if projectID == "" || userID == "" {
		return errors.New("projectID and userID are required")
	}

	// Verify ownership
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("\"Id\" = ? AND \"OwnerId\" = ? AND \"DeletedAt\" IS NULL", projectID, userID).
		Count(&count).Error

	if err != nil {
		return fmt.Errorf("failed to verify project ownership: %w", err)
	}

	if count == 0 {
		return ErrProjectUnauthorized
	}

	// Soft delete
	now := time.Now()
	err = r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where(&models.Project{ID: projectID, OwnerId: userID}).
		Update("\"DeletedAt\"", now).Error

	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	return nil
}

func (r *ProjectRepository) HardDeleteProject(ctx context.Context, projectID, userID string) error {
	if projectID == "" || userID == "" {
		return errors.New("projectID and userID are required")
	}

	result := r.db.WithContext(ctx).
		Unscoped().
		Where(&models.Project{ID: projectID, OwnerId: userID}).
		Delete(&models.Project{})

	if result.Error != nil {
		return fmt.Errorf("failed to hard delete project: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrProjectUnauthorized
	}

	return nil
}

func (r *ProjectRepository) RestoreProject(ctx context.Context, projectID, userID string) error {
	if projectID == "" || userID == "" {
		return errors.New("projectID and userID are required")
	}

	result := r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where(&models.Project{ID: projectID, OwnerId: userID}).
		Where("DeletedAt IS NOT NULL").
		Update("\"DeletedAt\"", nil)

	if result.Error != nil {
		return fmt.Errorf("failed to restore project: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrProjectNotFound
	}

	return nil
}

func (r *ProjectRepository) ExistsForUser(ctx context.Context, projectID, userID string) (bool, error) {
	if projectID == "" || userID == "" {
		return false, errors.New("projectID and userID are required")
	}

	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("\"Id\" = ? AND \"OwnerId\" = ? AND \"DeletedAt\" IS NULL", projectID, userID).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("failed to check project existence: %w", err)
	}

	return count > 0, nil
}

func (r *ProjectRepository) GetProjectWithLock(ctx context.Context, projectID, userID string) (*models.Project, error) {
	if projectID == "" || userID == "" {
		return nil, errors.New("projectID and userID are required")
	}

	var project models.Project

	err := r.db.WithContext(ctx).
		Model(&models.Project{}).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("\"Id\" = ? AND \"OwnerId\" = ? AND \"DeletedAt\" IS NULL", projectID, userID).
		First(&project).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project with lock: %w", err)
	}

	return &project, nil
}
