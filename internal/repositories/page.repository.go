package repositories

import (
	"context"
	"errors"
	"fmt"
	"my-go-app/internal/models"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"gorm.io/gorm"
)

var (
	// ErrPageNotFound is returned when a page is not found
	ErrPageNotFound = errors.New("page not found")
	// ErrPageUnauthorized is returned when user doesn't have access to page
	ErrPageUnauthorized = errors.New("unauthorized access to page")
	// ErrPageNameTaken is returned when the project already has a page with this name
	ErrPageNameTaken = errors.New("a page with this name already exists in the project")
	// ErrPageSlugTaken is returned when the project already has a route page with this slug
	ErrPageSlugTaken = errors.New("a page with this URL path already exists in the project")
)

func mapPageConstraintError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		switch pgErr.ConstraintName {
		case "Page_ProjectId_Name_key":
			return ErrPageNameTaken
		case "Page_ProjectId_Slug_key":
			return ErrPageSlugTaken
		}
	}
	return nil
}

type PageRepository struct {
	db *gorm.DB
}

func NewPageRepository(db *gorm.DB) PageRepositoryInterface {
	return &PageRepository{db: db}
}

func (r *PageRepository) ListByProject(ctx context.Context, projectID string) ([]models.Page, error) {
	if projectID == "" {
		return nil, errors.New("projectID is required")
	}

	pages, err := gorm.G[models.Page](r.db).
		Where(`"ProjectId" = ? AND "DeletedAt" IS NULL`, projectID).
		Order("\"CreatedAt\" ASC").
		Find(ctx)

	if err != nil {
		return nil, fmt.
			Errorf("failed to get pages by project ID: %w", err)
	}

	return pages, nil
}

func (r *PageRepository) GetByID(ctx context.Context, pageID, projectID string) (*models.Page, error) {
	if pageID == "" {
		return nil, errors.New("pageID is required")
	}

	query := gorm.G[models.Page](r.db).Where("\"Id\" = ? AND \"DeletedAt\" IS NULL", pageID)

	if projectID != "" {
		query = query.Where("\"ProjectId\" = ?", projectID)
	}

	page, err := query.First(ctx)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPageNotFound
		}
		return nil, fmt.Errorf("failed to get page: %w", err)
	}

	return &page, nil
}

func (r *PageRepository) Create(ctx context.Context, page *models.Page) error {
	if err := validateNewPage(page); err != nil {
		return err
	}
	if len(page.Styles) == 0 {
		page.Styles = []byte("{}")
	}

	now := time.Now()
	page.CreatedAt = now
	page.UpdatedAt = now

	if err := gorm.G[models.Page](r.db).Create(ctx, page); err != nil {
		if mapped := mapPageConstraintError(err); mapped != nil {
			return mapped
		}
		return fmt.Errorf("failed to create page: %w", err)
	}
	return nil
}

func validateNewPage(page *models.Page) error {
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
	return nil
}

func (r *PageRepository) Update(ctx context.Context, page *models.Page) error {
	if page == nil {
		return errors.New("page cannot be nil")
	}

	if page.Id == "" {
		return errors.New("page ID is required")
	}

	// Always update the UpdatedAt timestamp
	page.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).
		Where(&models.Page{Id: page.Id}).
		Where(`"DeletedAt" IS NULL`).
		Updates(page)

	if result.Error != nil {
		if mapped := mapPageConstraintError(result.Error); mapped != nil {
			return mapped
		}
		return fmt.Errorf("failed to update page: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrPageNotFound
	}

	return nil
}

func (r *PageRepository) UpdateFields(ctx context.Context, pageID string, updates map[string]any) error {
	if pageID == "" {
		return errors.New("pageID is required")
	}

	if len(updates) == 0 {
		return errors.New("no updates provided")
	}

	// Always update the UpdatedAt timestamp
	updates["UpdatedAt"] = time.Now()

	result := r.db.WithContext(ctx).
		Model(&models.Page{}).
		Where(`"Id" = ? AND "DeletedAt" IS NULL`, pageID).
		Updates(updates)

	if result.Error != nil {
		if mapped := mapPageConstraintError(result.Error); mapped != nil {
			return mapped
		}
		return fmt.Errorf("failed to update page fields: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrPageNotFound
	}

	return nil
}

func (r *PageRepository) Delete(ctx context.Context, pageID string) error {
	if pageID == "" {
		return errors.New("pageID is required")
	}

	now := time.Now()
	rows, err := gorm.G[models.Page](r.db).
		Where(&models.Page{Id: pageID}).
		Where(`"DeletedAt" IS NULL`).
		Update(ctx, "\"DeletedAt\"", now)
	if err != nil {
		return fmt.Errorf("failed to delete page: %w", err)
	}
	if rows == 0 {
		return ErrPageNotFound
	}

	return nil
}

func (r *PageRepository) DeleteForUser(ctx context.Context, pageID, projectID, userID string) error {
	if pageID == "" || projectID == "" || userID == "" {
		return errors.New("pageID, projectID, and userID are required")
	}

	// First verify project write access (owner or editor collaborator)
	count, err := gorm.G[models.Project](r.db).
		Where(`"Id" = ? AND "DeletedAt" IS NULL AND ("OwnerId" = ? OR EXISTS (SELECT 1 FROM "Collaborator" WHERE "ProjectId" = "Project"."Id" AND "UserId" = ? AND "Role" IN ('owner', 'editor')))`, projectID, userID, userID).
		Count(ctx, "*")

	if err != nil {
		return fmt.Errorf("failed to verify project ownership: %w", err)
	}

	if count == 0 {
		return ErrPageUnauthorized
	}

	// Now delete the page
	now := time.Now()
	rows, err := gorm.G[models.Page](r.db).
		Where(&models.Page{Id: pageID, ProjectId: projectID}).
		Where(`"DeletedAt" IS NULL`).
		Update(ctx, "DeletedAt", now)
	if err != nil {
		return fmt.Errorf("failed to delete page: %w", err)
	}
	if rows == 0 {
		return ErrPageNotFound
	}

	return nil
}

func (r *PageRepository) HardDelete(ctx context.Context, pageID string) error {
	if pageID == "" {
		return errors.New("pageID is required")
	}

	rows, err := gorm.G[models.Page](r.db).
		Scopes(func(stmt *gorm.Statement) { stmt.Unscoped = true }).
		Where(&models.Page{Id: pageID}).
		Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to hard delete page: %w", err)
	}
	if rows == 0 {
		return ErrPageNotFound
	}

	return nil
}

func (r *PageRepository) Restore(ctx context.Context, pageID string) error {
	if pageID == "" {
		return errors.New("pageID is required")
	}

	rows, err := gorm.G[models.Page](r.db).
		Where(&models.Page{Id: pageID}).
		Where(`"DeletedAt" IS NOT NULL`).
		Update(ctx, "\"DeletedAt\"", nil)
	if err != nil {
		return fmt.Errorf("failed to restore page: %w", err)
	}
	if rows == 0 {
		return ErrPageNotFound
	}

	return nil
}

func (r *PageRepository) ExistsInProject(ctx context.Context, pageID, projectID string) (bool, error) {
	if pageID == "" || projectID == "" {
		return false, errors.New("pageID and projectID are required")
	}

	count, err := gorm.G[models.Page](r.db).
		Where(&models.Page{Id: pageID, ProjectId: projectID}).
		Where(`"DeletedAt" IS NULL`).
		Count(ctx, "*")

	if err != nil {
		return false, fmt.Errorf("failed to check page existence: %w", err)
	}

	return count > 0, nil
}

func (r *PageRepository) CountByProject(ctx context.Context, projectID string) (int64, error) {
	if projectID == "" {
		return 0, errors.New("projectID is required")
	}

	count, err := gorm.G[models.Page](r.db).
		Where(&models.Page{ProjectId: projectID}).
		Where(`"DeletedAt" IS NULL`).
		Count(ctx, "*")

	if err != nil {
		return 0, fmt.Errorf("failed to count pages: %w", err)
	}

	return count, nil
}
