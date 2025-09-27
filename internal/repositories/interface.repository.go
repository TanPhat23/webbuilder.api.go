package repositories

import (
	"my-go-app/internal/models"

	"gorm.io/gorm"
)

type ElementRepositoryInterface interface {
	GetElements(projectID string) ([]models.EditorElement, error)
	CreateElement(elements []models.EditorElement, projectID string) error
	InsertElementAfter(projectID string, previousElementID string, element models.EditorElement) error
	ReplaceElements(projectID string, elements []models.EditorElement) error
	UpdateElement(element models.EditorElement, settings *string) error
	DeleteElement(elementID string) error
}
type ProjectRepositoryInterface interface {
	GetProjects() ([]models.Project, error)
	GetProjectByID(projectID string, userId string) (*models.Project, error)
	GetProjectsByUserID(userID string) ([]models.Project, error)
	GetProjectPages(projectID string, userId string) ([]models.Page, error)
	UpdateProject(projectID string, userId string, updates map[string]any) (*models.Project, error)
	DeleteProject(projectID string, userID string) error
}

type SnapshotRepositoryInterface interface {
	SaveSnapshot(projectId string, snapshot models.Snapshot) error
}

type SettingRepositoryInterface interface {
	GetSettingByElementID(db *gorm.DB, elementID string) (*models.Setting, error)
	GetSettingsByElementIDs(db *gorm.DB, elementIDs []string) ([]models.Setting, error)
	CreateSetting(db *gorm.DB, setting models.Setting) error
	CreateSettings(db *gorm.DB, settings []models.Setting) error
	UpdateSetting(db *gorm.DB, setting models.Setting) error
	UpdateSettings(db *gorm.DB, settings []models.Setting) error
	DeleteSetting(db *gorm.DB, elementID string) error
	DeleteSettings(db *gorm.DB, elementIDs []string) error
}

type PageRepositoryInterface interface {
	GetPagesByProjectID(projectID string) ([]models.Page, error)
	GetPageByID(pageID string, projectID string) (*models.Page, error)
	CreatePage(page models.Page) error
	UpdatePage(page models.Page) error
	DeletePage(pageID string) error
	DeletePageByProjectID(pageID string, projectID string, userID string) error
}

type RepositoriesInterface struct {
	ElementRepository  ElementRepositoryInterface
	ProjectRepository  ProjectRepositoryInterface
	SnapshotRepository SnapshotRepositoryInterface
	SettingRepository  SettingRepositoryInterface
	PageRepository     PageRepositoryInterface
}
