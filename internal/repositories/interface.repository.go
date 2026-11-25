package repositories

import (
	"context"
	"my-go-app/internal/models"
	"time"

	"gorm.io/gorm"
)

// ElementRepositoryInterface defines methods for element operations
type ElementRepositoryInterface interface {
	// List retrieves elements for a project, optionally filtered by pageID
	List(ctx context.Context, projectID string, pageID ...string) ([]models.EditorElement, error)
	// Replace replaces all elements for a project
	Replace(ctx context.Context, projectID string, elements []models.EditorElement) error
	// GetByID retrieves a single element by ID with all relations
	GetByID(ctx context.Context, elementID string) (*models.Element, error)
	// ListByPage retrieves all elements for a specific page
	ListByPage(ctx context.Context, pageID string) ([]models.Element, error)
	// ListByPageIDs retrieves the elements of the given pages as a tree.
	// Only pages that belong to projectID are read, so callers cannot reach another project's pages by ID.
	ListByPageIDs(ctx context.Context, projectID string, pageIDs []string) ([]models.EditorElement, error)
	// ListChildren retrieves child elements of a parent element
	ListChildren(ctx context.Context, parentID string) ([]models.Element, error)
	// ListRoots retrieves elements without a parent (root level)
	ListRoots(ctx context.Context, projectID string) ([]models.Element, error)
	// Create creates a single element
	Create(ctx context.Context, element *models.Element) error
	// Update updates a single element
	Update(ctx context.Context, element *models.Element) error
	// UpdateEventWorkflows updates the event workflows for an element
	UpdateEventWorkflows(ctx context.Context, elementID string, workflows []byte) error
	// Delete deletes a single element by ID
	Delete(ctx context.Context, elementID string) error
	// DeleteByPage deletes all elements in a page
	DeleteByPage(ctx context.Context, pageID string) error
	// DeleteByProject deletes all elements in a project
	DeleteByProject(ctx context.Context, projectID string) error
	// CountByProject counts elements in a project
	CountByProject(ctx context.Context, projectID string) (int64, error)
	// GetWithRelations retrieves element with all relations loaded
	GetWithRelations(ctx context.Context, elementID string) (*models.Element, error)
	// ListByIDs retrieves multiple elements by IDs
	ListByIDs(ctx context.Context, elementIDs []string) ([]models.Element, error)
}

// ElementCommentRepositoryInterface defines methods for element comment operations
type ElementCommentRepositoryInterface interface {
	// Create creates a new element comment
	Create(ctx context.Context, comment *models.ElementComment) (*models.ElementComment, error)
	// GetByID retrieves a single element comment by ID
	GetByID(ctx context.Context, id string) (*models.ElementComment, error)
	// List retrieves comments for an element with filtering and pagination
	List(ctx context.Context, elementID string, filter *models.ElementCommentFilter) ([]models.ElementComment, error)
	// Update updates an existing element comment
	Update(ctx context.Context, id string, updates map[string]any) (*models.ElementComment, error)
	// Delete soft deletes an element comment
	Delete(ctx context.Context, id string) error
	// ListByAuthor retrieves all comments by a specific author
	ListByAuthor(ctx context.Context, authorID string, limit int, offset int) ([]models.ElementComment, error)
	// Count counts comments for an element
	Count(ctx context.Context, elementID string) (int64, error)
	// ToggleResolved toggles the resolved status of a comment
	ToggleResolved(ctx context.Context, id string) (*models.ElementComment, error)
	// DeleteByElement deletes all comments for an element (cascade delete)
	DeleteByElement(ctx context.Context, elementID string) error
	// ListByProject retrieves all comments for elements in a project
	ListByProject(ctx context.Context, projectID string, limit int, offset int) ([]models.ElementComment, error)
	// CountByProject counts all comments for elements in a project
	CountByProject(ctx context.Context, projectID string) (int64, error)
}

// ProjectRepositoryInterface defines methods for project operations
type UserRepositoryInterface interface {
	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, userID string) (*models.User, error)
	// GetByEmail retrieves a user by email
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	// GetByUsername retrieves a user by username
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	// Search searches users by email or username
	Search(ctx context.Context, query string) ([]models.User, error)
}

type ProjectRepositoryInterface interface {
	// GetPublicByID retrieves a public project by ID
	GetPublicByID(ctx context.Context, projectID string) (*models.Project, error)
	// GetPublicBySubdomain retrieves a published project by its subdomain
	GetPublicBySubdomain(ctx context.Context, subdomain string) (*models.Project, error)
	// GetByID retrieves a project by ID with user ownership verification
	GetByID(ctx context.Context, projectID, userID string) (*models.Project, error)
	// GetWithAccess retrieves a project by ID with access verification (owner or collaborator)
	GetWithAccess(ctx context.Context, projectID, userID string) (*models.Project, error)
	// ListByUser retrieves all projects for a specific user
	ListByUser(ctx context.Context, userID string) ([]models.Project, error)
	// ListByCollaborator retrieves all projects where the user is a collaborator
	ListByCollaborator(ctx context.Context, userID string) ([]models.Project, error)
	// ListPages retrieves all pages for a project with ownership verification
	ListPages(ctx context.Context, projectID, userID string) ([]models.Page, error)
	// Create creates a new project
	Create(ctx context.Context, project *models.Project) error
	// CreateWithPage creates a project and its first page in one transaction
	CreateWithPage(ctx context.Context, project *models.Project, page *models.Page) error
	// GetAccess returns the user's role on a non-deleted project, or ErrProjectNotFound
	GetAccess(ctx context.Context, projectID, userID string) (*models.ProjectAccess, error)
	// Update updates a project with ownership verification
	Update(ctx context.Context, projectID, userID string, updates map[string]any) (*models.Project, error)
	// UpdateFields updates pre-authorized project fields for owners or editors
	UpdateFields(ctx context.Context, projectID, userID string, updates map[string]any) (*models.Project, error)
	// TransferOwnership moves project ownership from currentOwnerID to newOwnerID
	TransferOwnership(ctx context.Context, projectID, currentOwnerID, newOwnerID string) error
	// Delete soft deletes a project with ownership verification
	Delete(ctx context.Context, projectID, userID string) error
	// HardDelete permanently deletes a project
	HardDelete(ctx context.Context, projectID, userID string) error
	// Restore restores a soft-deleted project
	Restore(ctx context.Context, projectID, userID string) error
	// ExistsForUser checks if a project exists for a specific user
	ExistsForUser(ctx context.Context, projectID, userID string) (bool, error)
	// GetWithLock retrieves a project with a pessimistic lock for updates
	GetWithLock(ctx context.Context, projectID, userID string) (*models.Project, error)
}

// SnapshotRepositoryInterface defines methods for snapshot operations
type SnapshotRepositoryInterface interface {
	// Save saves a snapshot for a project
	Save(ctx context.Context, projectID string, snapshot *models.Snapshot) error
	// ListByProject retrieves snapshots for a project
	ListByProject(ctx context.Context, projectID string) ([]models.Snapshot, error)
	// GetByID retrieves a snapshot by ID
	GetByID(ctx context.Context, snapshotID string) (*models.Snapshot, error)
	// Delete deletes a snapshot
	Delete(ctx context.Context, snapshotID string) error
}

// PageRepositoryInterface defines methods for page operations
type PageRepositoryInterface interface {
	// ListByProject retrieves pages by project ID
	ListByProject(ctx context.Context, projectID string) ([]models.Page, error)
	// GetByID retrieves a page by ID and project ID
	GetByID(ctx context.Context, pageID, projectID string) (*models.Page, error)
	// Create creates a new page
	Create(ctx context.Context, page *models.Page) error
	// Update updates a page
	Update(ctx context.Context, page *models.Page) error
	// UpdateFields updates specific fields of a page
	UpdateFields(ctx context.Context, pageID string, updates map[string]any) error
	// Delete deletes a page by ID
	Delete(ctx context.Context, pageID string) error
	// DeleteForUser deletes a page by ID with project and user verification
	DeleteForUser(ctx context.Context, pageID, projectID, userID string) error
}

// ContentTypeRepositoryInterface defines methods for content type operations
type ContentTypeRepositoryInterface interface {
	// List retrieves all content types
	List(ctx context.Context) ([]models.ContentType, error)
	// ListByProject retrieves content types scoped to a project
	ListByProject(ctx context.Context, projectID string) ([]models.ContentType, error)
	// ListByUser retrieves content types for projects accessible by userID
	ListByUser(ctx context.Context, userID string) ([]models.ContentType, error)
	// GetByID retrieves a content type by ID
	GetByID(ctx context.Context, id string) (*models.ContentType, error)
	// GetByProjectAndSlugOrID retrieves a content type by project ID and slug or ID
	GetByProjectAndSlugOrID(ctx context.Context, projectID, slugOrID string) (*models.ContentType, error)
	// Create creates a new content type
	Create(ctx context.Context, ct *models.ContentType) (*models.ContentType, error)
	// Update updates a content type
	Update(ctx context.Context, id string, updates map[string]any) (*models.ContentType, error)
	// Delete deletes a content type
	Delete(ctx context.Context, id string) error
}

// ContentFieldRepositoryInterface defines methods for content field operations
type ContentFieldRepositoryInterface interface {
	// ListByContentType retrieves content fields by content type ID
	ListByContentType(ctx context.Context, contentTypeID string) ([]models.ContentField, error)
	// GetByID retrieves a content field by ID
	GetByID(ctx context.Context, id string) (*models.ContentField, error)
	// Create creates a new content field
	Create(ctx context.Context, cf *models.ContentField) (*models.ContentField, error)
	// Update updates a content field
	Update(ctx context.Context, id string, updates map[string]any) (*models.ContentField, error)
	// Delete deletes a content field
	Delete(ctx context.Context, id string) error
}

// ContentItemRepositoryInterface defines methods for content item operations
type ContentItemRepositoryInterface interface {
	// ListByContentType retrieves content items by content type ID
	ListByContentType(ctx context.Context, contentTypeID string) ([]models.ContentItem, error)
	// GetByID retrieves a content item by ID
	GetByID(ctx context.Context, id string) (*models.ContentItem, error)
	// GetBySlug retrieves a content item by slug and content type ID
	GetBySlug(ctx context.Context, contentTypeID, slug string) (*models.ContentItem, error)
	// ListPublic retrieves public content items with pagination and sorting
	ListPublic(ctx context.Context, contentTypeID string, limit int, sortBy, sortOrder string) ([]models.ContentItem, error)
	// Create creates a new content item with field values
	Create(ctx context.Context, ci *models.ContentItem, fieldValues []models.ContentFieldValue) (*models.ContentItem, error)
	// Update updates a content item and its field values
	Update(ctx context.Context, id string, updates map[string]any, fieldValues []models.ContentFieldValue) (*models.ContentItem, error)
	// Delete deletes a content item
	Delete(ctx context.Context, id string) error
}

// ContentFieldValueRepositoryInterface defines methods for content field value operations
type ContentFieldValueRepositoryInterface interface {
	// ListByContentItem retrieves content field values by content item ID
	ListByContentItem(ctx context.Context, contentItemID string) ([]models.ContentFieldValue, error)
	// GetByID retrieves a content field value by ID
	GetByID(ctx context.Context, id string) (*models.ContentFieldValue, error)
	// Create creates a new content field value
	Create(ctx context.Context, cfv *models.ContentFieldValue) (*models.ContentFieldValue, error)
	// Update updates a content field value
	Update(ctx context.Context, id string, value *string) (*models.ContentFieldValue, error)
	// Delete deletes a content field value
	Delete(ctx context.Context, id string) error
}

// ImageRepositoryInterface defines methods for image operations
type ImageRepositoryInterface interface {
	// Create creates a new image
	Create(ctx context.Context, image models.Image) (*models.Image, error)
	// ListByUser retrieves images by user ID
	ListByUser(ctx context.Context, userID string) ([]models.Image, error)
	// GetByID retrieves an image by ID with user verification
	GetByID(ctx context.Context, imageID string, userID string) (*models.Image, error)
	// GetPublicByID retrieves a non-deleted image by ID for public delivery
	GetPublicByID(ctx context.Context, imageID string) (*models.Image, error)
	// Delete deletes an image with user verification
	Delete(ctx context.Context, imageID string, userID string) error
	// SoftDelete soft deletes an image with user verification
	SoftDelete(ctx context.Context, imageID string, userID string) error
	// ListAll retrieves all images with pagination
	ListAll(ctx context.Context, limit int, offset int) ([]models.Image, error)
}

type CustomElementRepositoryInterface interface {
	List(ctx context.Context, userID string, isPublic *bool) ([]models.CustomElement, error)
	GetByID(ctx context.Context, id string, userID string) (*models.CustomElement, error)
	Create(ctx context.Context, customElement *models.CustomElement) (*models.CustomElement, error)
	Update(ctx context.Context, id string, userID string, updates map[string]any) (*models.CustomElement, error)
	Delete(ctx context.Context, id string, userID string) error
	ListPublic(ctx context.Context, category *string, limit int, offset int) ([]models.CustomElement, error)
	Duplicate(ctx context.Context, id string, userID string, newName string) (*models.CustomElement, error)
}

type CustomElementTypeRepositoryInterface interface {
	List(ctx context.Context) ([]models.CustomElementType, error)
	GetByID(ctx context.Context, id string) (*models.CustomElementType, error)
	GetByName(ctx context.Context, name string) (*models.CustomElementType, error)
	Create(ctx context.Context, customElementType *models.CustomElementType) (*models.CustomElementType, error)
	Update(ctx context.Context, id string, updates map[string]any) (*models.CustomElementType, error)
	Delete(ctx context.Context, id string) error
}

type InvitationRepositoryInterface interface {
	// Create creates a new invitation
	Create(ctx context.Context, invitation *models.Invitation) (*models.Invitation, error)
	// ListByProject retrieves invitations by project ID
	ListByProject(ctx context.Context, projectID string) ([]models.Invitation, error)
	// GetByID retrieves an invitation by ID
	GetByID(ctx context.Context, id string) (*models.Invitation, error)
	// GetByToken retrieves an invitation by token
	GetByToken(ctx context.Context, token string) (*models.Invitation, error)
	// Accept accepts an invitation and creates a collaborator
	Accept(ctx context.Context, token string, userID string) error
	// Delete deletes an invitation
	Delete(ctx context.Context, id string) error
	// UpdateStatus updates the status of an invitation
	UpdateStatus(ctx context.Context, id string, status models.InvitationStatus) error
	// Cancel cancels an invitation
	Cancel(ctx context.Context, id string) error
	// ListPendingByProject gets all pending invitations for a project
	ListPendingByProject(ctx context.Context, projectID string) ([]models.Invitation, error)
}

type CollaboratorRepositoryInterface interface {
	// Create creates a new collaborator
	Create(ctx context.Context, collaborator *models.Collaborator) (*models.Collaborator, error)
	// ListByProject retrieves collaborators by project ID
	ListByProject(ctx context.Context, projectID string) ([]models.Collaborator, error)
	// GetByID retrieves a collaborator by ID
	GetByID(ctx context.Context, id string) (*models.Collaborator, error)
	// UpdateRole updates a collaborator's role
	UpdateRole(ctx context.Context, id string, role models.CollaboratorRole) error
	// Delete deletes a collaborator
	Delete(ctx context.Context, id string) error
	// IsCollaborator checks if a user is a collaborator on a project
	IsCollaborator(ctx context.Context, projectID, userID string) (bool, error)
	// GetByProjectAndUser returns the collaborator record for (projectID, userID) if any
	GetByProjectAndUser(ctx context.Context, projectID, userID string) (*models.Collaborator, error)
	// DeleteByProjectAndUser removes a user's collaborator membership on a project
	DeleteByProjectAndUser(ctx context.Context, projectID, userID string) (bool, error)
}

type EventWorkflowRepositoryInterface interface {
	// Create creates a new event workflow
	Create(ctx context.Context, workflow *models.EventWorkflow) (*models.EventWorkflow, error)
	// GetByID retrieves an event workflow by ID
	GetByID(ctx context.Context, id string) (*models.EventWorkflow, error)
	// ListByProject retrieves all event workflows for a project
	ListByProject(ctx context.Context, projectID string) ([]models.EventWorkflow, error)
	// ListByProjectWithElements retrieves all event workflows for a project with element details
	ListByProjectWithElements(ctx context.Context, projectID string) ([]models.EventWorkflow, error)
	// ListEnabledByProject retrieves all enabled event workflows for a project
	ListEnabledByProject(ctx context.Context, projectID string) ([]models.EventWorkflow, error)
	// ListByName retrieves event workflows by name in a project
	ListByName(ctx context.Context, projectID, name string) ([]models.EventWorkflow, error)
	// Update updates an event workflow
	Update(ctx context.Context, id string, workflow *models.EventWorkflow) (*models.EventWorkflow, error)
	// SetEnabled updates the enabled status of an event workflow
	SetEnabled(ctx context.Context, id string, enabled bool) error
	// Delete deletes an event workflow
	Delete(ctx context.Context, id string) error
	// DeleteByProject deletes all event workflows for a project
	DeleteByProject(ctx context.Context, projectID string) error
	// CountByProject counts event workflows in a project
	CountByProject(ctx context.Context, projectID string) (int64, error)
	// ExistsByName checks if a workflow name already exists in a project
	ExistsByName(ctx context.Context, projectID, name string, excludeID string) (bool, error)
	// ListWithFilters retrieves event workflows with optional filters
	ListWithFilters(ctx context.Context, projectID string, enabled *bool, searchName string) ([]models.EventWorkflow, error)
}

type ElementEventWorkflowRepositoryInterface interface {
	// Create creates a new element event workflow association
	Create(ctx context.Context, eew *models.ElementEventWorkflow) (*models.ElementEventWorkflow, error)
	// GetByID retrieves an element event workflow by ID
	GetByID(ctx context.Context, id string) (*models.ElementEventWorkflow, error)
	// ListAll retrieves all element event workflows
	ListAll(ctx context.Context) ([]models.ElementEventWorkflow, error)
	// ListByElement retrieves all event workflows for a specific element
	ListByElement(ctx context.Context, elementID string) ([]models.ElementEventWorkflow, error)
	// ListByWorkflow retrieves all elements linked to a specific workflow
	ListByWorkflow(ctx context.Context, workflowID string) ([]models.ElementEventWorkflow, error)
	// ListByEvent retrieves all workflows for a specific event type
	ListByEvent(ctx context.Context, eventName string) ([]models.ElementEventWorkflow, error)
	// ListByFilters retrieves element event workflows with optional filters
	ListByFilters(ctx context.Context, elementID, workflowID, eventName string) ([]models.ElementEventWorkflow, error)
	// Update updates an element event workflow
	Update(ctx context.Context, id string, eew *models.ElementEventWorkflow) (*models.ElementEventWorkflow, error)
	// Delete deletes an element event workflow
	Delete(ctx context.Context, id string) error
	// DeleteByElement deletes all event workflows for a specific element
	DeleteByElement(ctx context.Context, elementID string) error
	// DeleteByWorkflow deletes all element associations for a specific workflow
	DeleteByWorkflow(ctx context.Context, workflowID string) error
	// ListByPage retrieves all element event workflows for elements belonging to a specific page
	ListByPage(ctx context.Context, pageID string) ([]models.ElementEventWorkflow, error)
	// IsLinked checks if a workflow is already linked to an element with a specific event
	IsLinked(ctx context.Context, elementID, workflowID, eventName string) (bool, error)
}

// NotificationRepositoryInterface defines methods for a user's notifications.
type NotificationRepositoryInterface interface {
	ListForUser(ctx context.Context, userID string, filter models.NotificationFilter) (*models.NotificationList, error)
	SetRead(ctx context.Context, userID, notificationID string, read bool) (*models.Notification, error)
	MarkAllRead(ctx context.Context, userID string) error
	Create(ctx context.Context, n *models.Notification) error
}

// AgentTokenRepositoryInterface stores MCP personal access tokens.
type AgentTokenRepositoryInterface interface {
	Create(ctx context.Context, t *models.AgentToken) error
	ListActiveByUser(ctx context.Context, userID string) ([]models.AgentToken, error)
	Revoke(ctx context.Context, userID, id string) (bool, error)
	GetActiveByHash(ctx context.Context, hash string) (*models.AgentToken, error)
	Touch(ctx context.Context, id string, at time.Time) error
}

// WebhookRepositoryInterface stores project webhooks and delivery logs.
type WebhookRepositoryInterface interface {
	Create(ctx context.Context, webhook *models.ProjectWebhook) (*models.ProjectWebhook, error)
	ListByProject(ctx context.Context, projectID string) ([]models.ProjectWebhook, error)
	GetByID(ctx context.Context, id string) (*models.ProjectWebhook, error)
	Update(ctx context.Context, id string, updates map[string]any) (*models.ProjectWebhook, error)
	Delete(ctx context.Context, id string) error
	ListMatchingForEvent(ctx context.Context, projectID, contentTypeID, event string) ([]models.ProjectWebhook, error)
	CreateDelivery(ctx context.Context, delivery *models.WebhookDelivery) (*models.WebhookDelivery, error)
	GetDeliveryByID(ctx context.Context, id string) (*models.WebhookDelivery, error)
	UpdateDelivery(ctx context.Context, id string, updates map[string]any) error
	ListDeliveriesByWebhook(ctx context.Context, webhookID string, limit int) ([]models.WebhookDelivery, error)
	ListPendingRetries(ctx context.Context, now time.Time, limit int) ([]models.WebhookDelivery, error)
}

type Repositories struct {
	AgentTokenRepository           AgentTokenRepositoryInterface
	WebhookRepository              WebhookRepositoryInterface
	NotificationRepository         NotificationRepositoryInterface
	ElementRepository              ElementRepositoryInterface
	ElementCommentRepository       ElementCommentRepositoryInterface
	UserRepository                 UserRepositoryInterface
	ProjectRepository              ProjectRepositoryInterface
	SnapshotRepository             SnapshotRepositoryInterface
	PageRepository                 PageRepositoryInterface
	ContentTypeRepository          ContentTypeRepositoryInterface
	ContentFieldRepository         ContentFieldRepositoryInterface
	ContentItemRepository          ContentItemRepositoryInterface
	ContentFieldValueRepository    ContentFieldValueRepositoryInterface
	ImageRepository                ImageRepositoryInterface
	CustomElementRepository        CustomElementRepositoryInterface
	CustomElementTypeRepository    CustomElementTypeRepositoryInterface
	InvitationRepository           InvitationRepositoryInterface
	CollaboratorRepository         CollaboratorRepositoryInterface
	EventWorkflowRepository        EventWorkflowRepositoryInterface
	ElementEventWorkflowRepository ElementEventWorkflowRepositoryInterface
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		AgentTokenRepository:           NewAgentTokenRepository(db),
		WebhookRepository:              NewWebhookRepository(db),
		ElementRepository:              NewElementRepository(db),
		ElementCommentRepository:       NewElementCommentRepository(db),
		UserRepository:                 NewUserRepository(db),
		ProjectRepository:              NewProjectRepository(db),
		SnapshotRepository:             NewSnapshotRepository(db),
		PageRepository:                 NewPageRepository(db),
		ContentTypeRepository:          NewContentTypeRepository(db),
		ContentFieldRepository:         NewContentFieldRepository(db),
		ContentItemRepository:          NewContentItemRepository(db),
		ContentFieldValueRepository:    NewContentFieldValueRepository(db),
		ImageRepository:                NewImageRepository(db),
		CustomElementRepository:        NewCustomElementRepository(db),
		CustomElementTypeRepository:    NewCustomElementTypeRepository(db),
		InvitationRepository:           NewInvitationRepository(db),
		CollaboratorRepository:         NewCollaboratorRepository(db),
		EventWorkflowRepository:        NewEventWorkflowRepository(db),
		ElementEventWorkflowRepository: NewElementEventWorkflowRepository(db),
		NotificationRepository:         NewNotificationRepository(db),
	}
}
