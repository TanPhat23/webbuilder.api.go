package repositories

import (
	"context"
	"errors"
	"fmt"
	"my-go-app/internal/authz"
	"my-go-app/internal/models"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrProjectNotFound     = errors.New("project not found")
	ErrProjectUnauthorized = errors.New("unauthorized access to project")
)

type ProjectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) ProjectRepositoryInterface {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) GetProjectByID(ctx context.Context, projectID, userID string) (*models.Project, error) {
	if projectID == "" || userID == "" {
		return nil, errors.New("projectID and userID are required")
	}

	project, err := gorm.G[models.Project](r.db).
		Where(`"Id" = ? AND "OwnerId" = ? AND "DeletedAt" IS NULL`, projectID, userID).
		First(ctx)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return &project, nil
}

// GetProjectWithAccess returns the project when userID owns it or collaborates on it.
// Every project read goes through this check, so the common case is one query.
func (r *ProjectRepository) GetProjectWithAccess(ctx context.Context, projectID, userID string) (*models.Project, error) {
	if projectID == "" || userID == "" {
		return nil, errors.New("projectID and userID are required")
	}

	project, err := gorm.G[models.Project](r.db).
		Where(`"Id" = ? AND "DeletedAt" IS NULL AND ("OwnerId" = ? OR EXISTS (
			SELECT 1 FROM "Collaborator" WHERE "Collaborator"."ProjectId" = "Project"."Id" AND "Collaborator"."UserId" = ?))`,
			projectID, userID, userID).
		First(ctx)
	if err == nil {
		return &project, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	// Only the failure path needs to tell "no such project" from "not yours": a
	// collaborator of a deleted project gets not-found, anyone else unauthorized.
	if _, err := gorm.G[models.Collaborator](r.db).
		Where(`"ProjectId" = ? AND "UserId" = ?`, projectID, userID).
		First(ctx); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectUnauthorized
		}
		return nil, fmt.Errorf("failed to check collaborator access: %w", err)
	}
	return nil, ErrProjectNotFound
}

func (r *ProjectRepository) GetPublicProjectByID(ctx context.Context, projectID string) (*models.Project, error) {
	if projectID == "" {
		return nil, errors.New("projectID is required")
	}

	project, err := gorm.G[models.Project](r.db).
		Where(`"Id" = ? AND "Published" = true AND "DeletedAt" IS NULL`, projectID).
		First(ctx)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get public project: %w", err)
	}

	return &project, nil
}

// projectListOmit are the heavy JSON columns the project lists never show.
var projectListOmit = []string{"Logic", "Styles", "Header"}

// GetProjectsByUserID returns the projects a user owns, then those they collaborate on,
// newest first within each group, in a single query. Logic, Styles and Header are left
// out: they can be large and only the editor needs them.
func (r *ProjectRepository) GetProjectsByUserID(ctx context.Context, userID string) ([]models.Project, error) {
	if userID == "" {
		return nil, errors.New("userID is required")
	}

	var projects []models.Project
	err := r.db.WithContext(ctx).
		Model(&models.Project{}).
		Omit(projectListOmit...).
		Where(`"DeletedAt" IS NULL AND ("OwnerId" = ? OR "Id" IN (SELECT "ProjectId" FROM "Collaborator" WHERE "UserId" = ?))`, userID, userID).
		// Order() cannot take a bound expression (it drops it silently), Clauses() can.
		Clauses(clause.OrderBy{Expression: clause.Expr{SQL: `("OwnerId" = ?) DESC, "CreatedAt" DESC`, Vars: []any{userID}}}).
		Find(&projects).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}

	return projects, nil
}

func (r *ProjectRepository) GetCollaboratorProjects(ctx context.Context, userID string) ([]models.Project, error) {
	if userID == "" {
		return nil, errors.New("userID is required")
	}

	var collabProjects []models.Project
	err := r.db.WithContext(ctx).
		Model(&models.Project{}).
		Omit(projectListOmit...).
		Joins(`JOIN "Collaborator" ON "Collaborator"."ProjectId" = "Project"."Id"`).
		Where(`"Collaborator"."UserId" = ? AND "Project"."DeletedAt" IS NULL`, userID).
		Order(`"Project"."CreatedAt" DESC`).
		Find(&collabProjects).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get collaborator projects: %w", err)
	}

	return collabProjects, nil
}

func (r *ProjectRepository) GetProjectPages(ctx context.Context, projectID, userID string) ([]models.Page, error) {
	if projectID == "" || userID == "" {
		return nil, errors.New("projectID and userID are required")
	}

	_, err := r.GetProjectWithAccess(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}

	pages, err := gorm.G[models.Page](r.db).
		Where(`"ProjectId" = ? AND "DeletedAt" IS NULL`, projectID).
		Order(`"CreatedAt" ASC`).
		Find(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get project pages: %w", err)
	}

	return pages, nil
}

func (r *ProjectRepository) CreateProject(ctx context.Context, project *models.Project) error {
	if project == nil {
		return errors.New("project cannot be nil")
	}

	now := time.Now()
	project.CreatedAt = now
	project.UpdatedAt = now

	err := gorm.G[models.Project](r.db).Create(ctx, project)

	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	return nil
}

func (r *ProjectRepository) ensureOwner(ctx context.Context, projectID, userID string) error {
	if projectID == "" || userID == "" {
		return errors.New("projectID and userID are required")
	}
	count, err := gorm.G[models.Project](r.db).
		Where(`"Id" = ? AND "OwnerId" = ? AND "DeletedAt" IS NULL`, projectID, userID).
		Count(ctx, "*")
	if err != nil {
		return fmt.Errorf("failed to verify project ownership: %w", err)
	}
	if count == 0 {
		return ErrProjectUnauthorized
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
	if err := r.ensureOwner(ctx, projectID, userID); err != nil {
		return nil, err
	}

	// Callers (handlers) are expected to pass a pre-built, allowlisted map.
	// We only stamp UpdatedAt here so it is always consistent.
	updates["UpdatedAt"] = time.Now()

	err := r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where(`"Id" = ? AND "OwnerId" = ? AND "DeletedAt" IS NULL`, projectID, userID).
		Updates(updates).Error

	if err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	return r.GetProjectByID(ctx, projectID, userID)
}

func (r *ProjectRepository) DeleteProject(ctx context.Context, projectID, userID string) error {
	if err := r.ensureOwner(ctx, projectID, userID); err != nil {
		return err
	}

	now := time.Now()
	_, err := gorm.G[models.Project](r.db).
		Where(&models.Project{ID: projectID, OwnerId: userID}).
		Update(ctx, `"DeletedAt"`, now)

	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	return nil
}

func (r *ProjectRepository) HardDeleteProject(ctx context.Context, projectID, userID string) error {
	if projectID == "" || userID == "" {
		return errors.New("projectID and userID are required")
	}

	rows, err := gorm.G[models.Project](r.db).
		Scopes(func(stmt *gorm.Statement) { stmt.Unscoped = true }).
		Where(&models.Project{ID: projectID, OwnerId: userID}).
		Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to hard delete project: %w", err)
	}
	if rows == 0 {
		return ErrProjectUnauthorized
	}

	return nil
}

func (r *ProjectRepository) RestoreProject(ctx context.Context, projectID, userID string) error {
	if projectID == "" || userID == "" {
		return errors.New("projectID and userID are required")
	}

	rows, err := gorm.G[models.Project](r.db).
		Where(&models.Project{ID: projectID, OwnerId: userID}).
		Where(`"DeletedAt" IS NOT NULL`).
		Update(ctx, `"DeletedAt"`, nil)
	if err != nil {
		return fmt.Errorf("failed to restore project: %w", err)
	}
	if rows == 0 {
		return ErrProjectNotFound
	}

	return nil
}

func (r *ProjectRepository) ExistsForUser(ctx context.Context, projectID, userID string) (bool, error) {
	if projectID == "" || userID == "" {
		return false, errors.New("projectID and userID are required")
	}

	count, err := gorm.G[models.Project](r.db).
		Where(`"Id" = ? AND "OwnerId" = ? AND "DeletedAt" IS NULL`, projectID, userID).
		Count(ctx, "*")

	if err != nil {
		return false, fmt.Errorf("failed to check project existence: %w", err)
	}

	return count > 0, nil
}

func (r *ProjectRepository) GetProjectWithLock(ctx context.Context, projectID, userID string) (*models.Project, error) {
	if projectID == "" || userID == "" {
		return nil, errors.New("projectID and userID are required")
	}

	project, err := gorm.G[models.Project](r.db, clause.Locking{Strength: "UPDATE"}).
		Where(`"Id" = ? AND "OwnerId" = ? AND "DeletedAt" IS NULL`, projectID, userID).
		First(ctx)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project with lock: %w", err)
	}

	return &project, nil
}

func (r *ProjectRepository) CreateProjectWithPage(ctx context.Context, project *models.Project, page *models.Page) error {
	if err := validateProjectWithPage(project, page); err != nil {
		return err
	}
	stampProjectAndPage(project, page)

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := gorm.G[models.Project](tx).Omit(clause.Associations).Create(ctx, project); err != nil {
			return fmt.Errorf("failed to create project: %w", err)
		}
		if err := gorm.G[models.Page](tx).Omit(clause.Associations).Create(ctx, page); err != nil {
			return fmt.Errorf("failed to create first page: %w", err)
		}
		return nil
	})
}

func validateProjectWithPage(project *models.Project, page *models.Page) error {
	if project == nil || page == nil {
		return errors.New("project and page are required")
	}
	if project.ID == "" || project.Name == "" || project.OwnerId == "" {
		return errors.New("project id, name and owner are required")
	}
	return nil
}

func stampProjectAndPage(project *models.Project, page *models.Page) {
	now := time.Now().UTC()
	if project.AuditFields == nil {
		project.AuditFields = &models.AuditFields{}
	}
	project.CreatedAt, project.UpdatedAt = now, now
	page.ProjectId = project.ID
	page.CreatedAt, page.UpdatedAt = now, now
}

func (r *ProjectRepository) GetProjectAccess(ctx context.Context, projectID, userID string) (*models.ProjectAccess, error) {
	if projectID == "" || userID == "" {
		return nil, errors.New("projectID and userID are required")
	}
	project, err := gorm.G[models.Project](r.db).
		Select(`"Id"`, `"OwnerId"`).
		Where(`"Id" = ? AND "DeletedAt" IS NULL`, projectID).
		First(ctx)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrProjectNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	access := &models.ProjectAccess{UserId: userID, ProjectId: projectID}
	if project.OwnerId == userID {
		access.Role, access.IsOwner = models.RoleOwner, true
		access.Permissions = authz.PermissionsForRole(authz.RoleOwner)
		return access, nil
	}

	collaborator, err := gorm.G[models.Collaborator](r.db).
		Select(`"Role"`).
		Where(`"ProjectId" = ? AND "UserId" = ?`, projectID, userID).
		First(ctx)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrProjectNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get collaborator: %w", err)
	}
	access.Role, access.IsCollaborator = collaborator.Role, true
	access.Permissions = authz.PermissionsForRole(authz.Role(collaborator.Role))
	return access, nil
}

// UpdateProjectFields updates pre-authorized project columns without forcing
// OwnerId equality in SQL (so editors can update content fields allowed by RBAC).
func (r *ProjectRepository) UpdateProjectFields(ctx context.Context, projectID, userID string, updates map[string]any) (*models.Project, error) {
	if projectID == "" || userID == "" {
		return nil, errors.New("projectID and userID are required")
	}
	if len(updates) == 0 {
		return nil, errors.New("no updates provided")
	}
	if _, err := r.GetProjectWithAccess(ctx, projectID, userID); err != nil {
		return nil, err
	}

	updates["UpdatedAt"] = time.Now()
	err := r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where(`"Id" = ? AND "DeletedAt" IS NULL`, projectID).
		Updates(updates).Error
	if err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	return r.GetProjectWithAccess(ctx, projectID, userID)
}

// TransferOwnership atomically moves Project.OwnerId from currentOwnerID to
// newOwnerID and converts the previous owner into an editor collaborator.
func (r *ProjectRepository) TransferOwnership(ctx context.Context, projectID, currentOwnerID, newOwnerID string) error {
	if projectID == "" || currentOwnerID == "" || newOwnerID == "" {
		return errors.New("projectID, currentOwnerID and newOwnerID are required")
	}
	if currentOwnerID == newOwnerID {
		return errors.New("new owner must be a different user")
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var project models.Project
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(`"Id" = ? AND "DeletedAt" IS NULL`, projectID).
			First(&project).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrProjectNotFound
		}
		if err != nil {
			return fmt.Errorf("failed to lock project: %w", err)
		}
		if project.OwnerId != currentOwnerID {
			return ErrProjectUnauthorized
		}

		var targetUser models.User
		if err := tx.Select(`"Id"`).Where(`"Id" = ?`, newOwnerID).First(&targetUser).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("target user not found")
			}
			return fmt.Errorf("failed to verify new owner: %w", err)
		}

		now := time.Now()
		if err := tx.Model(&models.Project{}).
			Where(`"Id" = ?`, projectID).
			Updates(map[string]any{"OwnerId": newOwnerID, "UpdatedAt": now}).Error; err != nil {
			return fmt.Errorf("failed to transfer ownership: %w", err)
		}

		if err := tx.Where(`"ProjectId" = ? AND "UserId" = ?`, projectID, newOwnerID).
			Delete(&models.Collaborator{}).Error; err != nil {
			return fmt.Errorf("failed to clear new owner collaborator row: %w", err)
		}

		var prevCollab models.Collaborator
		err = tx.Where(`"ProjectId" = ? AND "UserId" = ?`, projectID, currentOwnerID).First(&prevCollab).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			collab := models.Collaborator{
				Id:        "collab-" + projectID + "-" + currentOwnerID,
				UserId:    currentOwnerID,
				ProjectId: projectID,
				Role:      models.RoleEditor,
				BaseAuditFields: models.BaseAuditFields{
					CreatedAt: now,
					UpdatedAt: now,
				},
			}
			if err := tx.Create(&collab).Error; err != nil {
				return fmt.Errorf("failed to add previous owner as editor: %w", err)
			}
		} else if err != nil {
			return fmt.Errorf("failed to check previous owner collaborator row: %w", err)
		} else if err := tx.Model(&prevCollab).Updates(map[string]any{"Role": models.RoleEditor, "UpdatedAt": now}).Error; err != nil {
			return fmt.Errorf("failed to update previous owner role: %w", err)
		}

		return nil
	})
}
