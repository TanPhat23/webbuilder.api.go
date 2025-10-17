package repositories

import (
	"my-go-app/internal/models"

	"gorm.io/gorm"
)

type ElementRepositoryInterface interface {
	GetElements(projectID string) ([]models.EditorElement, error)
	ReplaceElements(projectID string, elements []models.EditorElement) error
}
type ProjectRepositoryInterface interface {
	CreateProject(project models.Project) (*models.Project, error)
	GetProjects() ([]models.Project, error)
	GetProjectByID(projectID string, userId string) (*models.Project, error)
	GetPublicProjectByID(projectID string) (*models.Project, error)
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

type ContentTypeRepositoryInterface interface {
	GetContentTypes() ([]models.ContentType, error)
	GetContentTypeByID(id string) (*models.ContentType, error)
	CreateContentType(ct models.ContentType) (*models.ContentType, error)
	UpdateContentType(id string, updates map[string]any) (*models.ContentType, error)
	DeleteContentType(id string) error
}

type ContentFieldRepositoryInterface interface {
	GetContentFieldsByContentType(contentTypeId string) ([]models.ContentField, error)
	GetContentFieldByID(id string) (*models.ContentField, error)
	CreateContentField(cf models.ContentField) (*models.ContentField, error)
	UpdateContentField(id string, updates map[string]any) (*models.ContentField, error)
	DeleteContentField(id string) error
}

type ContentItemRepositoryInterface interface {
	GetContentItemsByContentType(contentTypeId string) ([]models.ContentItem, error)
	GetContentItemByID(id string) (*models.ContentItem, error)
	GetContentItemBySlug(contentTypeId string, slug string) (*models.ContentItem, error)
	GetPublicContentItems(contentTypeId string, limit int, sortBy string, sortOrder string) ([]models.ContentItem, error)
	CreateContentItem(ci models.ContentItem, fieldValues []models.ContentFieldValue) (*models.ContentItem, error)
	UpdateContentItem(id string, updates map[string]any) (*models.ContentItem, error)
	DeleteContentItem(id string) error
}

type ContentFieldValueRepositoryInterface interface {
	GetContentFieldValuesByContentItem(contentItemId string) ([]models.ContentFieldValue, error)
	GetContentFieldValueByID(id string) (*models.ContentFieldValue, error)
	CreateContentFieldValue(cfv models.ContentFieldValue) (*models.ContentFieldValue, error)
	UpdateContentFieldValue(id string, value *string) (*models.ContentFieldValue, error)
	DeleteContentFieldValue(id string) error
}

type ImageRepositoryInterface interface {
	CreateImage(image models.Image) (*models.Image, error)
	GetImagesByUserID(userID string) ([]models.Image, error)
	GetImageByID(imageID string, userID string) (*models.Image, error)
	DeleteImage(imageID string, userID string) error
	SoftDeleteImage(imageID string, userID string) error
	GetAllImages(limit int, offset int) ([]models.Image, error)
}

type MarketplaceRepositoryInterface interface {
	CreateMarketplaceItem(item models.MarketplaceItem, tagIds []string, categoryIds []string) (*models.MarketplaceItem, error)
	GetMarketplaceItems(filter MarketplaceFilter) ([]models.MarketplaceItem, int64, error)
	GetMarketplaceItemByID(id string) (*models.MarketplaceItem, error)
	UpdateMarketplaceItem(id string, userId string, updates map[string]any) (*models.MarketplaceItem, error)
	DeleteMarketplaceItem(id string, userId string) error
	DownloadMarketplaceItem(itemId string, userId string) (*models.Project, error)
	IncrementDownloads(id string) error
	IncrementLikes(id string) error
	CreateCategory(category models.Category) (*models.Category, error)
	GetCategories() ([]models.Category, error)
	GetCategoryByID(id string) (*models.Category, error)
	GetCategoryByName(name string) (*models.Category, error)
	DeleteCategory(id string) error
	CreateTag(tag models.Tag) (*models.Tag, error)
	GetTags() ([]models.Tag, error)
	GetTagByID(id string) (*models.Tag, error)
	GetTagByName(name string) (*models.Tag, error)
	DeleteTag(id string) error
}

type MarketplaceFilter struct {
	TemplateType string
	Featured     *bool
	CategoryId   string
	TagId        string
	AuthorId     string
	Search       string
	SortBy       string
	SortOrder    string
	Limit        int
	Offset       int
}

type RepositoriesInterface struct {
	ElementRepository           ElementRepositoryInterface
	ProjectRepository           ProjectRepositoryInterface
	SnapshotRepository          SnapshotRepositoryInterface
	SettingRepository           SettingRepositoryInterface
	PageRepository              PageRepositoryInterface
	ContentTypeRepository       ContentTypeRepositoryInterface
	ContentFieldRepository      ContentFieldRepositoryInterface
	ContentItemRepository       ContentItemRepositoryInterface
	ContentFieldValueRepository ContentFieldValueRepositoryInterface
	ImageRepository             ImageRepositoryInterface
	MarketplaceRepository       MarketplaceRepositoryInterface
}
