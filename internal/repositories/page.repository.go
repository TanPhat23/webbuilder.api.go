package repositories

import (
	"context"
	"errors"
	"fmt"
	"my-go-app/internal/models"
	"time"

	"gorm.io/gorm"
)

var (
	// ErrPageNotFound is returned when a page is not found
	ErrPageNotFound = errors.New("page not found")
	// ErrPageUnauthorized is returned when user doesn't have access to page
	ErrPageUnauthorized = errors.New("unauthorized access to page")
)

type PageRepository struct {
	db *gorm.DB
}

func NewPageRepository(db *gorm.DB) PageRepositoryInterface {
	return &PageRepository{db: db}
}

func (r *PageRepository) GetPagesByProjectID(ctx context.Context, projectID string) ([]models.Page, error) {
	if projectID == "" {
		return nil, errors.New("projectID is required")
	}

	var pages []models.Page

	err := r.db.WithContext(ctx).
		Where(&models.Page{ProjectId: projectID, DeletedAt: nil}).
		Order("\"CreatedAt\" ASC").
		Find(&pages).Error

	if err != nil {
		return nil, fmt.
Errorf("failed to get pages by project ID: %w", err)
	}

	return pages, nil
}

func (r *PageRepository) GetPageByID(ctx context.Context, pageID, projectID string) (*models.Page, error) {
	if pageID == "" || projectID == "" {
		return nil, errors.New("pageID and projectID are required")
	}

	var page models.Page

	err := r.db.WithContext(ctx).
		Where(&models.Page{Id: pageID, ProjectId: projectID, DeletedAt: nil}).
		First(&page).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPageNotFound
		}
		return nil, fmt.Errorf("failed to get page: %w", err)
	}

	return &page, nil
}

func (r *PageRepository) CreatePage(ctx context.Context, page *models.Page) error {
	if page == nil {
		return errors.New("page cannot be nil")
	}

	if page.Name == "" {
		return errors.New("page name is required")
	}

	if page.ProjectId == "" {
		return errors.New("project ID is required")
	}

	if page.Type == "" {
		return errors.New("page type is required")
	}

	now := time.Now()
	page.CreatedAt = now
	page.UpdatedAt = now

	err := r.db.WithContext(ctx).
		Model(&models.Page{}).
		Create(page).Error

	if err != nil {
		return fmt.Errorf("failed to create page: %w", err)
	}

	return nil
}

func (r *PageRepository) UpdatePage(ctx context.Context, page *models.Page) error {
	if page == nil {
		return errors.New("page cannot be nil")
	}

	if page.Id == "" {
		return errors.New("page ID is required")
	}

	// Always update the UpdatedAt timestamp
	page.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).
		Where(&models.Page{Id: page.Id, DeletedAt: nil}).
		Updates(page)

	if result.Error != nil {
		return fmt.Errorf("failed to update page: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrPageNotFound
	}

	return nil
}

func (r *PageRepository) UpdatePageFields(ctx context.Context, pageID string, updates map[string]any) error {
	if pageID == "" {
		return errors.New("pageID is required")
	}

	if len(updates) == 0 {
		return errors.New("no updates provided")
	}

	// Always update the UpdatedAt timestamp
	updates["UpdatedAt"] = time.Now()

	result := r.db.WithContext(ctx).
		Where(&models.Page{Id: pageID, DeletedAt: nil}).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to update page fields: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrPageNotFound
	}

	return nil
}

func (r *PageRepository) DeletePage(ctx context.Context, pageID string) error {
	if pageID == "" {
		return errors.New("pageID is required")
	}

	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&models.Page{}).
		Where(&models.Page{Id: pageID, DeletedAt: nil}).
		Update("\"DeletedAt\"", now)

	if result.Error != nil {
		return fmt.Errorf("failed to delete page: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrPageNotFound
	}

	return nil
}

func (r *PageRepository) DeletePageByProjectID(ctx context.Context, pageID, projectID, userID string) error {
	if pageID == "" || projectID == "" || userID == "" {
		return errors.New("pageID, projectID, and userID are required")
	}

	// First verify project ownership
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where(&models.Project{ID: projectID, OwnerId: userID, DeletedAt: nil}).
		Count(&count).Error

	if err != nil {
		return fmt.Errorf("failed to verify project ownership: %w", err)
	}

	if count == 0 {
		return ErrPageUnauthorized
	}

	// Now delete the page
	now := time.Now()
	result := r.db.WithContext(ctx).
		Where(&models.Page{Id: pageID, ProjectId: projectID, DeletedAt: nil}).
		Update("DeletedAt", now)

	if result.Error != nil {
		return fmt.Errorf("failed to delete page: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrPageNotFound
	}

	return nil
}

func (r *PageRepository) HardDeletePage(ctx context.Context, pageID string) error {
	if pageID == "" {
		return errors.New("pageID is required")
	}

	result := r.db.WithContext(ctx).
		Unscoped().
		Where(&models.Page{Id: pageID}).
		Delete(&models.Page{})

	if result.Error != nil {
		return fmt.Errorf("failed to hard delete page: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrPageNotFound
	}

	return nil
}

func (r *PageRepository) RestorePage(ctx context.Context, pageID string) error {
	if pageID == "" {
		return errors.New("pageID is required")
	}

	result := r.db.WithContext(ctx).
		Model(&models.Page{}).
		Where(&models.Page{Id: pageID}).
		Where("DeletedAt IS NOT NULL").
		Update("\"DeletedAt\"", nil)

	if result.Error != nil {
		return fmt.Errorf("failed to restore page: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrPageNotFound
	}

	return nil
}

func (r *PageRepository) ExistsInProject(ctx context.Context, pageID, projectID string) (bool, error) {
	if pageID == "" || projectID == "" {
		return false, errors.New("pageID and projectID are required")
	}

	var count int64
	err := r.db.WithContext(ctx).
		Where(&models.Page{Id: pageID, ProjectId: projectID, DeletedAt: nil}).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("failed to check page existence: %w", err)
	}

	return count > 0, nil
}

func (r *PageRepository) CountPagesByProjectID(ctx context.Context, projectID string) (int64, error) {
	if projectID == "" {
		return 0, errors.New("projectID is required")
	}

	var count int64
	err := r.db.WithContext(ctx).
		Where(&models.Page{ProjectId: projectID, DeletedAt: nil}).
		Count(&count).Error

	if err != nil {
		return 0, fmt.Errorf("failed to count pages: %w", err)
	}

	return count, nil
}
