package repositories

import (
	"context"
	"my-go-app/internal/models"

	"gorm.io/gorm"
)

// ElementRepositoryInterface defines methods for element operations
type ElementRepositoryInterface interface {
	// GetElements retrieves elements for a project, optionally filtered by pageID
	GetElements(ctx context.Context, projectID string, pageID ...string) ([]models.EditorElement, error)
	// ReplaceElements replaces all elements for a project
	ReplaceElements(ctx context.Context, projectID string, elements []models.EditorElement) error
	// GetElementByID retrieves a single element by ID with all relations
	GetElementByID(ctx context.Context, elementID string) (*models.Element, error)
	// GetElementsByPageID retrieves all elements for a specific page
	GetElementsByPageID(ctx context.Context, pageID string) ([]models.Element, error)
	// GetElementsByPageIds retrieves all elements for multiple pages with tree structure
	GetElementsByPageIds(ctx context.Context, pageIDs []string) ([]models.EditorElement, error)
	// GetChildElements retrieves child elements of a parent element
	GetChildElements(ctx context.Context, parentID string) ([]models.Element, error)
	// GetRootElements retrieves elements without a parent (root level)
	GetRootElements(ctx context.Context, projectID string) ([]models.Element, error)
	// CreateElement creates a single element
	CreateElement(ctx context.Context, element *models.Element) error
	// UpdateElement updates a single element
	UpdateElement(ctx context.Context, element *models.Element) error
	// UpdateEventWorkflows updates the event workflows for an element
	UpdateEventWorkflows(ctx context.Context, elementID string, workflows []byte) error
	// DeleteElementByID deletes a single element by ID
	DeleteElementByID(ctx context.Context, elementID string) error
	// DeleteElementsByPageID deletes all elements in a page
	DeleteElementsByPageID(ctx context.Context, pageID string) error
	// DeleteElementsByProjectID deletes all elements in a project
	DeleteElementsByProjectID(ctx context.Context, projectID string) error
	// CountElementsByProjectID counts elements in a project
	CountElementsByProjectID(ctx context.Context, projectID string) (int64, error)
	// GetElementWithRelations retrieves element with all relations loaded
	GetElementWithRelations(ctx context.Context, elementID string) (*models.Element, error)
	// GetElementsByIDs retrieves multiple elements by IDs
	GetElementsByIDs(ctx context.Context, elementIDs []string) ([]models.Element, error)
}

// ElementCommentRepositoryInterface defines methods for element comment operations
type ElementCommentRepositoryInterface interface {
	// CreateElementComment creates a new element comment
	CreateElementComment(ctx context.Context, comment *models.ElementComment) (*models.ElementComment, error)
	// GetElementCommentByID retrieves a single element comment by ID
	GetElementCommentByID(ctx context.Context, id string) (*models.ElementComment, error)
	// GetElementComments retrieves comments for an element with filtering and pagination
	GetElementComments(ctx context.Context, elementID string, filter *models.ElementCommentFilter) ([]models.ElementComment, error)
	// UpdateElementComment updates an existing element comment
	UpdateElementComment(ctx context.Context, id string, updates map[string]any) (*models.ElementComment, error)
	// DeleteElementComment soft deletes an element comment
	DeleteElementComment(ctx context.Context, id string) error
	// GetElementCommentsByAuthorID retrieves all comments by a specific author
	GetElementCommentsByAuthorID(ctx context.Context, authorID string, limit int, offset int) ([]models.ElementComment, error)
	// CountElementComments counts comments for an element
	CountElementComments(ctx context.Context, elementID string) (int64, error)
	// ToggleResolvedStatus toggles the resolved status of a comment
	ToggleResolvedStatus(ctx context.Context, id string) (*models.ElementComment, error)
	// DeleteElementCommentsByElementID deletes all comments for an element (cascade delete)
	DeleteElementCommentsByElementID(ctx context.Context, elementID string) error
	// GetElementCommentsByProjectID retrieves all comments for elements in a project
	GetElementCommentsByProjectID(ctx context.Context, projectID string, limit int, offset int) ([]models.ElementComment, error)
	// CountElementCommentsByProjectID counts all comments for elements in a project
	CountElementCommentsByProjectID(ctx context.Context, projectID string) (int64, error)
}

// ProjectRepositoryInterface defines methods for project operations
type UserRepositoryInterface interface {
	// GetUserByID retrieves a user by ID
	GetUserByID(ctx context.Context, userID string) (*models.User, error)
	// GetUserByEmail retrieves a user by email
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	// GetUserByUsername retrieves a user by username
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	// SearchUsers searches users by email or username
	SearchUsers(ctx context.Context, query string) ([]models.User, error)
}

type ProjectRepositoryInterface interface {
	// GetPublicProjectByID retrieves a public project by ID
	GetPublicProjectByID(ctx context.Context, projectID string) (*models.Project, error)
	// GetProjectByID retrieves a project by ID with user ownership verification
	GetProjectByID(ctx context.Context, projectID, userID string) (*models.Project, error)
	// GetProjectWithAccess retrieves a project by ID with access verification (owner or collaborator)
	GetProjectWithAccess(ctx context.Context, projectID, userID string) (*models.Project, error)
	// GetProjectsByUserID retrieves all projects for a specific user
	GetProjectsByUserID(ctx context.Context, userID string) ([]models.Project, error)
	// GetCollaboratorProjects retrieves all projects where the user is a collaborator
	GetCollaboratorProjects(ctx context.Context, userID string) ([]models.Project, error)
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
	// UpdatePageFields updates specific fields of a page
	UpdatePageFields(ctx context.Context, pageID string, updates map[string]any) error
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
	// UpdateContentItem updates a content item and its field values
	UpdateContentItem(ctx context.Context, id string, updates map[string]any, fieldValues []models.ContentFieldValue) (*models.ContentItem, error)
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

// CommentRepositoryInterface defines methods for comment operations
type CommentRepositoryInterface interface {
	// CreateComment creates a new comment
	CreateComment(comment models.Comment) (*models.Comment, error)
	// GetCommentByID retrieves a comment by ID with author and reactions
	GetCommentByID(id string) (*models.Comment, error)
	// GetComments retrieves comments with filtering and pagination
	GetComments(filter models.CommentFilter) ([]models.Comment, int64, error)
	// UpdateComment updates a comment with user verification
	UpdateComment(id string, userId string, updates map[string]any) (*models.Comment, error)
	// DeleteComment soft deletes a comment with user verification
	DeleteComment(id string, userId string) error
	// CreateReaction creates or updates a reaction
	CreateReaction(reaction models.CommentReaction) (*models.CommentReaction, error)
	// DeleteReaction deletes a reaction
	DeleteReaction(commentId string, userId string, reactionType string) error
	// GetReactionsByCommentID retrieves all reactions for a comment
	GetReactionsByCommentID(commentId string) ([]models.CommentReaction, error)
	// GetReactionSummary retrieves reaction counts grouped by type
	GetReactionSummary(commentId string) ([]models.ReactionSummary, error)
	// GetCommentCountByItemID returns the number of comments for a marketplace item
	GetCommentCountByItemID(itemId string) (int64, error)
	// ModerateComment updates the status of a comment (for admin/moderation)
	ModerateComment(id string, status string) error
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

type InvitationRepositoryInterface interface {
	// CreateInvitation creates a new invitation
	CreateInvitation(ctx context.Context, invitation *models.Invitation) (*models.Invitation, error)
	// GetInvitationsByProject retrieves invitations by project ID
	GetInvitationsByProject(ctx context.Context, projectID string) ([]models.Invitation, error)
	// GetInvitationByID retrieves an invitation by ID
	GetInvitationByID(ctx context.Context, id string) (*models.Invitation, error)
	// GetInvitationByToken retrieves an invitation by token
	GetInvitationByToken(ctx context.Context, token string) (*models.Invitation, error)
	// AcceptInvitation accepts an invitation and creates a collaborator
	AcceptInvitation(ctx context.Context, token string, userID string) error
	// DeleteInvitation deletes an invitation
	DeleteInvitation(ctx context.Context, id string) error
	// UpdateInvitationStatus updates the status of an invitation
	UpdateInvitationStatus(ctx context.Context, id string, status models.InvitationStatus) error
	// CancelInvitation cancels an invitation
	CancelInvitation(ctx context.Context, id string) error
	// GetPendingInvitationsByProject gets all pending invitations for a project
	GetPendingInvitationsByProject(ctx context.Context, projectID string) ([]models.Invitation, error)
}

type CollaboratorRepositoryInterface interface {
	// CreateCollaborator creates a new collaborator
	CreateCollaborator(ctx context.Context, collaborator *models.Collaborator) (*models.Collaborator, error)
	// GetCollaboratorsByProject retrieves collaborators by project ID
	GetCollaboratorsByProject(ctx context.Context, projectID string) ([]models.Collaborator, error)
	// GetCollaboratorByID retrieves a collaborator by ID
	GetCollaboratorByID(ctx context.Context, id string) (*models.Collaborator, error)
	// UpdateCollaboratorRole updates a collaborator's role
	UpdateCollaboratorRole(ctx context.Context, id string, role models.CollaboratorRole) error
	// DeleteCollaborator deletes a collaborator
	DeleteCollaborator(ctx context.Context, id string) error
	// IsCollaborator checks if a user is a collaborator on a project
	IsCollaborator(ctx context.Context, projectID, userID string) (bool, error)
}


type EventWorkflowRepositoryInterface interface {
	// CreateEventWorkflow creates a new event workflow
	CreateEventWorkflow(ctx context.Context, workflow *models.EventWorkflow) (*models.EventWorkflow, error)
	// GetEventWorkflowByID retrieves an event workflow by ID
	GetEventWorkflowByID(ctx context.Context, id string) (*models.EventWorkflow, error)
	// GetEventWorkflowsByProjectID retrieves all event workflows for a project
	GetEventWorkflowsByProjectID(ctx context.Context, projectID string) ([]models.EventWorkflow, error)
	// GetEventWorkflowsByProjectIDWithElements retrieves all event workflows for a project with element details
	GetEventWorkflowsByProjectIDWithElements(ctx context.Context, projectID string) ([]models.EventWorkflow, error)
	// GetEnabledEventWorkflowsByProjectID retrieves all enabled event workflows for a project
	GetEnabledEventWorkflowsByProjectID(ctx context.Context, projectID string) ([]models.EventWorkflow, error)
	// GetEventWorkflowsByName retrieves event workflows by name in a project
	GetEventWorkflowsByName(ctx context.Context, projectID, name string) ([]models.EventWorkflow, error)
	// UpdateEventWorkflow updates an event workflow
	UpdateEventWorkflow(ctx context.Context, id string, workflow *models.EventWorkflow) (*models.EventWorkflow, error)
	// UpdateEventWorkflowEnabled updates the enabled status of an event workflow
	UpdateEventWorkflowEnabled(ctx context.Context, id string, enabled bool) error
	// DeleteEventWorkflow deletes an event workflow
	DeleteEventWorkflow(ctx context.Context, id string) error
	// DeleteEventWorkflowsByProjectID deletes all event workflows for a project
	DeleteEventWorkflowsByProjectID(ctx context.Context, projectID string) error
	// CountEventWorkflowsByProjectID counts event workflows in a project
	CountEventWorkflowsByProjectID(ctx context.Context, projectID string) (int64, error)
	// CheckIfWorkflowNameExists checks if a workflow name already exists in a project
	CheckIfWorkflowNameExists(ctx context.Context, projectID, name string, excludeID string) (bool, error)
	// GetEventWorkflowsWithFilters retrieves event workflows with optional filters
	GetEventWorkflowsWithFilters(ctx context.Context, projectID string, enabled *bool, searchName string) ([]models.EventWorkflow, error)
}

type ElementEventWorkflowRepositoryInterface interface {
	// CreateElementEventWorkflow creates a new element event workflow association
	CreateElementEventWorkflow(ctx context.Context, eew *models.ElementEventWorkflow) (*models.ElementEventWorkflow, error)
	// GetElementEventWorkflowByID retrieves an element event workflow by ID
	GetElementEventWorkflowByID(ctx context.Context, id string) (*models.ElementEventWorkflow, error)
	// GetElementEventWorkflowsByElementID retrieves all event workflows for a specific element
	GetElementEventWorkflowsByElementID(ctx context.Context, elementID string) ([]models.ElementEventWorkflow, error)
	// GetElementEventWorkflowsByWorkflowID retrieves all elements linked to a specific workflow
	GetElementEventWorkflowsByWorkflowID(ctx context.Context, workflowID string) ([]models.ElementEventWorkflow, error)
	// GetElementEventWorkflowsByEventName retrieves all workflows for a specific event type
	GetElementEventWorkflowsByEventName(ctx context.Context, eventName string) ([]models.ElementEventWorkflow, error)
	// GetElementEventWorkflowsByFilters retrieves element event workflows with optional filters
	GetElementEventWorkflowsByFilters(ctx context.Context, elementID, workflowID, eventName string) ([]models.ElementEventWorkflow, error)
	// UpdateElementEventWorkflow updates an element event workflow
	UpdateElementEventWorkflow(ctx context.Context, id string, eew *models.ElementEventWorkflow) (*models.ElementEventWorkflow, error)
	// DeleteElementEventWorkflow deletes an element event workflow
	DeleteElementEventWorkflow(ctx context.Context, id string) error
	// DeleteElementEventWorkflowsByElementID deletes all event workflows for a specific element
	DeleteElementEventWorkflowsByElementID(ctx context.Context, elementID string) error
	// DeleteElementEventWorkflowsByWorkflowID deletes all element associations for a specific workflow
	DeleteElementEventWorkflowsByWorkflowID(ctx context.Context, workflowID string) error
	// CheckIfWorkflowLinkedToElement checks if a workflow is already linked to an element with a specific event
	CheckIfWorkflowLinkedToElement(ctx context.Context, elementID, workflowID, eventName string) (bool, error)
}

type RepositoriesInterface struct {
	ElementRepository               ElementRepositoryInterface
	ElementCommentRepository        ElementCommentRepositoryInterface
	UserRepository                  UserRepositoryInterface
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
	InvitationRepository            InvitationRepositoryInterface
	CollaboratorRepository          CollaboratorRepositoryInterface
	CommentRepository               CommentRepositoryInterface
	EventWorkflowRepository             *EventWorkflowRepository
	ElementEventWorkflowRepository   *ElementEventWorkflowRepository
}

func NewRepositories(db *gorm.DB) *RepositoriesInterface {
	settingRepo := NewSettingRepository(db)

	return &RepositoriesInterface{
		ElementRepository:           NewElementRepository(db, settingRepo),
		ElementCommentRepository:    NewElementCommentRepository(db),
		UserRepository:              NewUserRepository(db),
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
		InvitationRepository:        NewInvitationRepository(db),
		CollaboratorRepository:      NewCollaboratorRepository(db),
		EventWorkflowRepository:             NewEventWorkflowRepository(db),
		ElementEventWorkflowRepository:   NewElementEventWorkflowRepository(db),
		CommentRepository:           NewCommentRepository(db),
	}
}
