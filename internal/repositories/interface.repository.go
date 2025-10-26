package repositories

import (
	"context"
	"my-go-app/internal/models"

	"gorm.io/gorm"
)

// ElementRepositoryInterface defines methods for element operations
type ElementRepositoryInterface interface {
	// GetElements retrieves elements for a project
	GetElements(ctx context.Context, projectID string) ([]models.EditorElement, error)
	// ReplaceElements replaces all elements for a project
	ReplaceElements(ctx context.Context, projectID string, elements []models.EditorElement) error
}

// ProjectRepositoryInterface defines methods for project operations
type ProjectRepositoryInterface interface {
	// GetProjects retrieves all projects
	GetProjects(ctx context.Context) ([]models.Project, error)
	// GetPublicProjectByID retrieves a public project by ID
	GetPublicProjectByID(ctx context.Context, projectID string) (*models.Project, error)
	// GetProjectByID retrieves a project by ID with user ownership verification
	GetProjectByID(ctx context.Context, projectID, userID string) (*models.Project, error)
	// GetProjectsByUserID retrieves all projects for a specific user
	GetProjectsByUserID(ctx context.Context, userID string) ([]models.Project, error)
	// GetProjectPages retrieves all pages for a project with ownership verification
	GetProjectPages(ctx context.Context, projectID, userID string) ([]models.Page, error)
	// CreateProject creates a new project
	CreateProject(ctx context.Context, project *models.Project) error
	// UpdateProject updates a project with ownership verification
	UpdateProject(ctx context.Context, projectID, userID string, updates map[string]any) (*models.Project, error)
	// DeleteProject soft deletes a project with ownership verification
	DeleteProject(ctx context.Context, projectID, userID string) error
	// HardDeleteProject permanently deletes a project
	HardDeleteProject(ctx context.Context, projectID, userID string) error
	// RestoreProject restores a soft-deleted project
	RestoreProject(ctx context.Context, projectID, userID string) error
	// ExistsForUser checks if a project exists for a specific user
	ExistsForUser(ctx context.Context, projectID, userID string) (bool, error)
	// GetProjectWithLock retrieves a project with a pessimistic lock for updates
	GetProjectWithLock(ctx context.Context, projectID, userID string) (*models.Project, error)
}

// SnapshotRepositoryInterface defines methods for snapshot operations
type SnapshotRepositoryInterface interface {
	// SaveSnapshot saves a snapshot for a project
	SaveSnapshot(ctx context.Context, projectID string, snapshot *models.Snapshot) error
	// GetSnapshotsByProjectID retrieves snapshots for a project
	GetSnapshotsByProjectID(ctx context.Context, projectID string) ([]models.Snapshot, error)
	// GetSnapshotByID retrieves a snapshot by ID
	GetSnapshotByID(ctx context.Context, snapshotID string) (*models.Snapshot, error)
	// DeleteSnapshot deletes a snapshot
	DeleteSnapshot(ctx context.Context, snapshotID string) error
}

// SettingRepositoryInterface defines methods for setting operations
type SettingRepositoryInterface interface {
	// GetSettingByElementID retrieves a setting by element ID
	GetSettingByElementID(ctx context.Context, db *gorm.DB, elementID string) (*models.Setting, error)
	// GetSettingsByElementIDs retrieves settings by element IDs
	GetSettingsByElementIDs(ctx context.Context, db *gorm.DB, elementIDs []string) ([]models.Setting, error)
	// CreateSetting creates a new setting
	CreateSetting(ctx context.Context, db *gorm.DB, setting *models.Setting) error
	// CreateSettings creates multiple settings
	CreateSettings(ctx context.Context, db *gorm.DB, settings []models.Setting) error
	// UpdateSetting updates a setting
	UpdateSetting(ctx context.Context, db *gorm.DB, setting *models.Setting) error
	// UpdateSettings updates multiple settings
	UpdateSettings(ctx context.Context, db *gorm.DB, settings []models.Setting) error
	// DeleteSetting deletes a setting by element ID
	DeleteSetting(ctx context.Context, db *gorm.DB, elementID string) error
	// DeleteSettings deletes settings by element IDs
	DeleteSettings(ctx context.Context, db *gorm.DB, elementIDs []string) error
}

// PageRepositoryInterface defines methods for page operations
type PageRepositoryInterface interface {
	// GetPagesByProjectID retrieves pages by project ID
	GetPagesByProjectID(ctx context.Context, projectID string) ([]models.Page, error)
	// GetPageByID retrieves a page by ID and project ID
	GetPageByID(ctx context.Context, pageID, projectID string) (*models.Page, error)
	// CreatePage creates a new page
	CreatePage(ctx context.Context, page *models.Page) error
	// UpdatePage updates a page
	UpdatePage(ctx context.Context, page *models.Page) error
	// DeletePage deletes a page by ID
	DeletePage(ctx context.Context, pageID string) error
	// DeletePageByProjectID deletes a page by ID with project and user verification
	DeletePageByProjectID(ctx context.Context, pageID, projectID, userID string) error
}

// ContentTypeRepositoryInterface defines methods for content type operations
type ContentTypeRepositoryInterface interface {
	// GetContentTypes retrieves all content types
	GetContentTypes(ctx context.Context) ([]models.ContentType, error)
	// GetContentTypeByID retrieves a content type by ID
	GetContentTypeByID(ctx context.Context, id string) (*models.ContentType, error)
	// CreateContentType creates a new content type
	CreateContentType(ctx context.Context, ct *models.ContentType) (*models.ContentType, error)
	// UpdateContentType updates a content type
	UpdateContentType(ctx context.Context, id string, updates map[string]any) (*models.ContentType, error)
	// DeleteContentType deletes a content type
	DeleteContentType(ctx context.Context, id string) error
}

// ContentFieldRepositoryInterface defines methods for content field operations
type ContentFieldRepositoryInterface interface {
	// GetContentFieldsByContentType retrieves content fields by content type ID
	GetContentFieldsByContentType(ctx context.Context, contentTypeID string) ([]models.ContentField, error)
	// GetContentFieldByID retrieves a content field by ID
	GetContentFieldByID(ctx context.Context, id string) (*models.ContentField, error)
	// CreateContentField creates a new content field
	CreateContentField(ctx context.Context, cf *models.ContentField) (*models.ContentField, error)
	// UpdateContentField updates a content field
	UpdateContentField(ctx context.Context, id string, updates map[string]any) (*models.ContentField, error)
	// DeleteContentField deletes a content field
	DeleteContentField(ctx context.Context, id string) error
}

// ContentItemRepositoryInterface defines methods for content item operations
type ContentItemRepositoryInterface interface {
	// GetContentItemsByContentType retrieves content items by content type ID
	GetContentItemsByContentType(ctx context.Context, contentTypeID string) ([]models.ContentItem, error)
	// GetContentItemByID retrieves a content item by ID
	GetContentItemByID(ctx context.Context, id string) (*models.ContentItem, error)
	// GetContentItemBySlug retrieves a content item by slug and content type ID
	GetContentItemBySlug(ctx context.Context, contentTypeID, slug string) (*models.ContentItem, error)
	// GetPublicContentItems retrieves public content items with pagination and sorting
	GetPublicContentItems(ctx context.Context, contentTypeID string, limit int, sortBy, sortOrder string) ([]models.ContentItem, error)
	// CreateContentItem creates a new content item with field values
	CreateContentItem(ctx context.Context, ci *models.ContentItem, fieldValues []models.ContentFieldValue) (*models.ContentItem, error)
	// UpdateContentItem updates a content item
	UpdateContentItem(ctx context.Context, id string, updates map[string]any) (*models.ContentItem, error)
	// DeleteContentItem deletes a content item
	DeleteContentItem(ctx context.Context, id string) error
}

// ContentFieldValueRepositoryInterface defines methods for content field value operations
type ContentFieldValueRepositoryInterface interface {
	// GetContentFieldValuesByContentItem retrieves content field values by content item ID
	GetContentFieldValuesByContentItem(ctx context.Context, contentItemID string) ([]models.ContentFieldValue, error)
	// GetContentFieldValueByID retrieves a content field value by ID
	GetContentFieldValueByID(ctx context.Context, id string) (*models.ContentFieldValue, error)
	// CreateContentFieldValue creates a new content field value
	CreateContentFieldValue(ctx context.Context, cfv *models.ContentFieldValue) (*models.ContentFieldValue, error)
	// UpdateContentFieldValue updates a content field value
	UpdateContentFieldValue(ctx context.Context, id string, value *string) (*models.ContentFieldValue, error)
	// DeleteContentFieldValue deletes a content field value
	DeleteContentFieldValue(ctx context.Context, id string) error
}

// MarketplaceFilter defines filters for marketplace item queries
type MarketplaceFilter struct {
	TemplateType string
	CategoryId   string
	TagId        string
	AuthorId     string
	Search       string
	SortBy       string
	SortOrder    string
	Featured     *bool
	Limit        int
	Offset       int
}

// MarketplaceRepositoryInterface defines methods for marketplace operations
type MarketplaceRepositoryInterface interface {
	// CreateMarketplaceItem creates a new marketplace item with tags and categories
	CreateMarketplaceItem(item models.MarketplaceItem, tagIds []string, categoryIds []string) (*models.MarketplaceItem, error)
	// GetMarketplaceItems retrieves marketplace items with filters
	GetMarketplaceItems(filter MarketplaceFilter) ([]models.MarketplaceItem, int64, error)
	// GetMarketplaceItemByID retrieves a marketplace item by ID
	GetMarketplaceItemByID(id string) (*models.MarketplaceItem, error)
	// UpdateMarketplaceItem updates a marketplace item with user verification
	UpdateMarketplaceItem(id string, userId string, updates map[string]any) (*models.MarketplaceItem, error)
	// DeleteMarketplaceItem deletes a marketplace item with user verification
	DeleteMarketplaceItem(id string, userId string) error
	// DownloadMarketplaceItem downloads a marketplace item as a project
	DownloadMarketplaceItem(itemId string, userId string) (*models.Project, error)
	// IncrementDownloads increments the download count for a marketplace item
	IncrementDownloads(id string) error
	// IncrementLikes increments the like count for a marketplace item
	IncrementLikes(id string) error
	// CreateCategory creates a new category
	CreateCategory(category models.Category) (*models.Category, error)
	// GetCategories retrieves all categories
	GetCategories() ([]models.Category, error)
	// GetCategoryByID retrieves a category by ID
	GetCategoryByID(id string) (*models.Category, error)
	// GetCategoryByName retrieves a category by name
	GetCategoryByName(name string) (*models.Category, error)
	// DeleteCategory deletes a category
	DeleteCategory(id string) error
	// CreateTag creates a new tag
	CreateTag(tag models.Tag) (*models.Tag, error)
	// GetTags retrieves all tags
	GetTags() ([]models.Tag, error)
	// GetTagByID retrieves a tag by ID
	GetTagByID(id string) (*models.Tag, error)
	// GetTagByName retrieves a tag by name
	GetTagByName(name string) (*models.Tag, error)
	// DeleteTag deletes a tag
	DeleteTag(id string) error
}

// ImageRepositoryInterface defines methods for image operations
type ImageRepositoryInterface interface {
	// CreateImage creates a new image
	CreateImage(image models.Image) (*models.Image, error)
	// GetImagesByUserID retrieves images by user ID
	GetImagesByUserID(userID string) ([]models.Image, error)
	// GetImageByID retrieves an image by ID with user verification
	GetImageByID(imageID string, userID string) (*models.Image, error)
	// DeleteImage deletes an image with user verification
	DeleteImage(imageID string, userID string) error
	// SoftDeleteImage soft deletes an image with user verification
	SoftDeleteImage(imageID string, userID string) error
	// GetAllImages retrieves all images with pagination
	GetAllImages(limit int, offset int) ([]models.Image, error)
}

type CustomElementRepositoryInterface interface {
	GetCustomElements(ctx context.Context, userID string, isPublic *bool) ([]models.CustomElement, error)
	GetCustomElementByID(ctx context.Context, id string, userID string) (*models.CustomElement, error)
	CreateCustomElement(ctx context.Context, customElement *models.CustomElement) (*models.CustomElement, error)
	UpdateCustomElement(ctx context.Context, id string, userID string, updates map[string]any) (*models.CustomElement, error)
	DeleteCustomElement(ctx context.Context, id string, userID string) error
	GetPublicCustomElements(ctx context.Context, category *string, limit int, offset int) ([]models.CustomElement, error)
	DuplicateCustomElement(ctx context.Context, id string, userID string, newName string) (*models.CustomElement, error)
}

type CustomElementTypeRepositoryInterface interface {
	GetCustomElementTypes(ctx context.Context) ([]models.CustomElementType, error)
	GetCustomElementTypeByID(ctx context.Context, id string) (*models.CustomElementType, error)
	GetCustomElementTypeByName(ctx context.Context, name string) (*models.CustomElementType, error)
	CreateCustomElementType(ctx context.Context, customElementType *models.CustomElementType) (*models.CustomElementType, error)
	UpdateCustomElementType(ctx context.Context, id string, updates map[string]any) (*models.CustomElementType, error)
	DeleteCustomElementType(ctx context.Context, id string) error
}

type RepositoriesInterface struct {
	ElementRepository               ElementRepositoryInterface
	ProjectRepository               ProjectRepositoryInterface
	SnapshotRepository              SnapshotRepositoryInterface
	SettingRepository               SettingRepositoryInterface
	PageRepository                  PageRepositoryInterface
	ContentTypeRepository           ContentTypeRepositoryInterface
	ContentFieldRepository          ContentFieldRepositoryInterface
	ContentItemRepository           ContentItemRepositoryInterface
	ContentFieldValueRepository     ContentFieldValueRepositoryInterface
	MarketplaceRepository           MarketplaceRepositoryInterface
	ImageRepository                 ImageRepositoryInterface
	CustomElementRepository         CustomElementRepositoryInterface
	CustomElementTypeRepository     CustomElementTypeRepositoryInterface
}

func NewRepositories(db *gorm.DB) *RepositoriesInterface {
	settingRepo := NewSettingRepository(db)

	return &RepositoriesInterface{
		ElementRepository:           NewElementRepository(db, settingRepo),
		ProjectRepository:           NewProjectRepository(db),
		SnapshotRepository:          NewSnapshotRepository(db),
		SettingRepository:           settingRepo,
		PageRepository:              NewPageRepository(db),
		ContentTypeRepository:       NewContentTypeRepository(db),
		ContentFieldRepository:      NewContentFieldRepository(db),
		ContentItemRepository:       NewContentItemRepository(db),
		ContentFieldValueRepository: NewContentFieldValueRepository(db),
		MarketplaceRepository:       NewMarketplaceRepository(db),
		ImageRepository:             NewImageRepository(db),
		CustomElementRepository:     NewCustomElementRepository(db),
		CustomElementTypeRepository: NewCustomElementTypeRepository(db),
	}
}
