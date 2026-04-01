package services

import (
	"context"
	"mime/multipart"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
)

type EmailServiceInterface interface {
	SendInvitationEmail(to, projectName, inviteLink string) error
}

type InvitationServiceInterface interface {
	CreateInvitation(ctx context.Context, projectID, email string, role models.CollaboratorRole, invitedBy string) (*models.Invitation, error)
	AcceptInvitation(ctx context.Context, token, userID string) error
	GetInvitationsByProject(ctx context.Context, projectID string) ([]models.Invitation, error)
	GetInvitationByID(ctx context.Context, id string) (*models.Invitation, error)
	DeleteInvitation(ctx context.Context, id string) error
	CheckProjectOwnership(ctx context.Context, projectID, userID string) error
	CancelInvitation(ctx context.Context, id string) error
	UpdateInvitationStatus(ctx context.Context, id string, status models.InvitationStatus) error
	GetPendingInvitationsByProject(ctx context.Context, projectID string) ([]models.Invitation, error)
}

type UserServiceInterface interface {
	GetUserByID(ctx context.Context, userID string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	SearchUsers(ctx context.Context, query string) ([]models.User, error)
}

type ProjectServiceInterface interface {
	GetPublicProjectByID(ctx context.Context, projectID string) (*models.Project, error)
	GetProjectByID(ctx context.Context, projectID, userID string) (*models.Project, error)
	GetProjectWithAccess(ctx context.Context, projectID, userID string) (*models.Project, error)
	GetProjectsByUserID(ctx context.Context, userID string) ([]models.Project, error)
	GetCollaboratorProjects(ctx context.Context, userID string) ([]models.Project, error)
	GetProjectPages(ctx context.Context, projectID string) ([]models.Page, error)
	CreateProject(ctx context.Context, project *models.Project) (*models.Project, error)
	UpdateProject(ctx context.Context, projectID, userID string, project *models.Project) (*models.Project, error)
	DeleteProject(ctx context.Context, projectID, userID string) error
	HardDeleteProject(ctx context.Context, projectID, userID string) error
	RestoreProject(ctx context.Context, projectID, userID string) error
	ExistsForUser(ctx context.Context, projectID, userID string) (bool, error)
	GetProjectWithLock(ctx context.Context, projectID, userID string) (*models.Project, error)
}

type PageServiceInterface interface {
	GetPagesByProjectID(ctx context.Context, projectID string) ([]models.Page, error)
	GetPageByID(ctx context.Context, pageID string) (*models.Page, error)
	CreatePage(ctx context.Context, page *models.Page) (*models.Page, error)
	UpdatePage(ctx context.Context, pageID string, page *models.Page) (*models.Page, error)
	UpdatePageFields(ctx context.Context, pageID string, updates map[string]interface{}) error
	DeletePage(ctx context.Context, pageID string) error
	DeletePageByProjectID(ctx context.Context, pageID, projectID, userID string) error
}

type ElementServiceInterface interface {
	GetElements(ctx context.Context, projectID string, pageID ...string) ([]models.EditorElement, error)
	GetElementsByPageID(ctx context.Context, pageID string) ([]models.Element, error)
	GetElementsByPageIds(ctx context.Context, pageIDs []string) ([]models.EditorElement, error)
	GetElementsByIDs(ctx context.Context, elementIDs []string) ([]models.Element, error)
}

type ImageServiceInterface interface {
	CreateImage(ctx context.Context, image models.Image) (*models.Image, error)
	CreateUploadedImage(ctx context.Context, userID string, fileName string, imageName *string, file multipart.File) (*models.ImageUploadResponse, error)
	CreateBase64UploadedImage(ctx context.Context, userID string, imageData string, imageName *string) (*models.ImageUploadResponse, error)
	GetImagesByUserID(ctx context.Context, userID string) ([]models.Image, error)
	GetImageByID(ctx context.Context, imageID, userID string) (*models.Image, error)
	DeleteImage(ctx context.Context, imageID, userID string) error
	SoftDeleteImage(ctx context.Context, imageID, userID string) error
	GetAllImages(ctx context.Context, limit, offset int) ([]models.Image, error)
}

type CollaboratorServiceInterface interface {
	CreateCollaborator(ctx context.Context, collaborator *models.Collaborator) (*models.Collaborator, error)
	GetCollaboratorsByProject(ctx context.Context, projectID string) ([]models.Collaborator, error)
	GetCollaboratorByID(ctx context.Context, id string) (*models.Collaborator, error)
	UpdateCollaboratorRole(ctx context.Context, id string, role models.CollaboratorRole) error
	DeleteCollaborator(ctx context.Context, id string) error
	IsCollaborator(ctx context.Context, userID, projectID string) (bool, error)
}

type SnapshotServiceInterface interface {
	SaveSnapshot(ctx context.Context, projectID string, snapshot *models.Snapshot) error
	GetSnapshotsByProjectID(ctx context.Context, projectID string) ([]models.Snapshot, error)
	GetSnapshotByID(ctx context.Context, id string) (*models.Snapshot, error)
	DeleteSnapshot(ctx context.Context, id string) error
	DeleteSnapshotWithAccess(ctx context.Context, snapshotID, projectID, userID string) error
}

type ContentTypeServiceInterface interface {
	GetContentTypes(ctx context.Context) ([]models.ContentType, error)
	GetContentTypeByID(ctx context.Context, id string) (*models.ContentType, error)
	CreateContentType(ctx context.Context, contentType *models.ContentType) (*models.ContentType, error)
	UpdateContentType(ctx context.Context, id string, contentType *models.ContentType) (*models.ContentType, error)
	DeleteContentType(ctx context.Context, id string) error
}

type ContentFieldServiceInterface interface {
	GetContentFieldsByContentType(ctx context.Context, contentTypeID string) ([]models.ContentField, error)
	GetContentFieldByID(ctx context.Context, id string) (*models.ContentField, error)
	CreateContentField(ctx context.Context, field *models.ContentField) (*models.ContentField, error)
	UpdateContentField(ctx context.Context, id string, updates map[string]any) (*models.ContentField, error)
	DeleteContentField(ctx context.Context, id string) error
}

type ContentItemServiceInterface interface {
	GetContentItemsByContentType(ctx context.Context, contentTypeID string) ([]models.ContentItem, error)
	GetContentItemByID(ctx context.Context, id string) (*models.ContentItem, error)
	GetContentItemBySlug(ctx context.Context, contentTypeID, slug string) (*models.ContentItem, error)
	GetPublicContentItems(ctx context.Context, contentTypeID string, limit int, sortBy, sortOrder string) ([]models.ContentItem, error)
	CreateContentItem(ctx context.Context, item *models.ContentItem, fieldValues []models.ContentFieldValue) (*models.ContentItem, error)
	UpdateContentItem(ctx context.Context, id string, updates map[string]any, fieldValues []models.ContentFieldValue) (*models.ContentItem, error)
	DeleteContentItem(ctx context.Context, id string) error
}

type CustomElementServiceInterface interface {
	GetCustomElements(ctx context.Context, userID string, isPublic *bool) ([]models.CustomElement, error)
	GetCustomElementByID(ctx context.Context, id, userID string) (*models.CustomElement, error)
	CreateCustomElement(ctx context.Context, element *models.CustomElement) (*models.CustomElement, error)
	UpdateCustomElement(ctx context.Context, id, userID string, updates map[string]any) (*models.CustomElement, error)
	DeleteCustomElement(ctx context.Context, id, userID string) error
	GetPublicCustomElements(ctx context.Context, category *string, limit, offset int) ([]models.CustomElement, error)
	DuplicateCustomElement(ctx context.Context, id, userID, newName string) (*models.CustomElement, error)
}

type CustomElementTypeServiceInterface interface {
	GetCustomElementTypes(ctx context.Context) ([]models.CustomElementType, error)
	GetCustomElementTypeByID(ctx context.Context, id string) (*models.CustomElementType, error)
	GetCustomElementTypeByName(ctx context.Context, name string) (*models.CustomElementType, error)
	CreateCustomElementType(ctx context.Context, ceType *models.CustomElementType) (*models.CustomElementType, error)
	UpdateCustomElementType(ctx context.Context, id string, updates map[string]any) (*models.CustomElementType, error)
	DeleteCustomElementType(ctx context.Context, id string) error
}

type EventWorkflowServiceInterface interface {
	CreateEventWorkflow(ctx context.Context, workflow *models.EventWorkflow) (*models.EventWorkflow, error)
	GetEventWorkflowByID(ctx context.Context, id string) (*models.EventWorkflow, error)
	GetEventWorkflowsByProjectID(ctx context.Context, projectID string) ([]models.EventWorkflow, error)
	GetEventWorkflowsByProjectIDWithElements(ctx context.Context, projectID string) ([]models.EventWorkflow, error)
	GetEnabledEventWorkflowsByProjectID(ctx context.Context, projectID string) ([]models.EventWorkflow, error)
	GetEventWorkflowsByName(ctx context.Context, projectID, name string) ([]models.EventWorkflow, error)
	UpdateEventWorkflow(ctx context.Context, id string, workflow *models.EventWorkflow) (*models.EventWorkflow, error)
	UpdateEventWorkflowEnabled(ctx context.Context, id string, enabled bool) error
	DeleteEventWorkflow(ctx context.Context, id string) error
	DeleteEventWorkflowsByProjectID(ctx context.Context, projectID string) error
	CountEventWorkflowsByProjectID(ctx context.Context, projectID string) (int64, error)
	CheckIfWorkflowNameExists(ctx context.Context, projectID, name, excludeID string) (bool, error)
	GetEventWorkflowsWithFilters(ctx context.Context, projectID string, enabled *bool, searchName string) ([]models.EventWorkflow, error)
}

type ElementEventWorkflowServiceInterface interface {
	CreateElementEventWorkflow(ctx context.Context, eew *models.ElementEventWorkflow) (*models.ElementEventWorkflow, error)
	GetElementEventWorkflowByID(ctx context.Context, id string) (*models.ElementEventWorkflow, error)
	GetAllElementEventWorkflows(ctx context.Context) ([]models.ElementEventWorkflow, error)
	GetElementEventWorkflowsByElementID(ctx context.Context, elementID string) ([]models.ElementEventWorkflow, error)
	GetElementEventWorkflowsByWorkflowID(ctx context.Context, workflowID string) ([]models.ElementEventWorkflow, error)
	GetElementEventWorkflowsByEventName(ctx context.Context, eventName string) ([]models.ElementEventWorkflow, error)
	GetElementEventWorkflowsByFilters(ctx context.Context, elementID, workflowID, eventName string) ([]models.ElementEventWorkflow, error)
	UpdateElementEventWorkflow(ctx context.Context, id string, eew *models.ElementEventWorkflow) (*models.ElementEventWorkflow, error)
	DeleteElementEventWorkflow(ctx context.Context, id string) error
	DeleteElementEventWorkflowsByElementID(ctx context.Context, elementID string) error
	DeleteElementEventWorkflowsByWorkflowID(ctx context.Context, workflowID string) error
	GetElementEventWorkflowsByPageID(ctx context.Context, pageID string) ([]models.ElementEventWorkflow, error)
	CheckIfWorkflowLinkedToElement(ctx context.Context, elementID, workflowID, eventName string) (bool, error)
}

type CommentServiceInterface interface {
	CreateComment(ctx context.Context, comment *models.Comment) (*models.Comment, error)
	GetCommentByID(ctx context.Context, id string) (*models.Comment, error)
	GetComments(ctx context.Context, itemID string) ([]models.Comment, error)
	GetCommentsByItemID(ctx context.Context, itemID string, filter models.CommentFilter) ([]models.Comment, int64, error)
	UpdateComment(ctx context.Context, id string, userID string, updates map[string]any) (*models.Comment, error)
	DeleteComment(ctx context.Context, id string, userID string) error
	CreateReaction(ctx context.Context, reaction *models.CommentReaction) (*models.CommentReaction, error)
	DeleteReaction(ctx context.Context, commentID string, userID string, reactionType string) error
	GetReactionsByCommentID(ctx context.Context, commentID string) ([]models.CommentReaction, error)
	GetReactionSummary(ctx context.Context, commentID string) (map[string]int, error)
	GetCommentCountByItemID(ctx context.Context, itemID string) (int64, error)
	ModerateComment(ctx context.Context, id string, status string) error
}

type ElementCommentServiceInterface interface {
	CreateElementComment(ctx context.Context, comment *models.ElementComment) (*models.ElementComment, error)
	GetElementCommentByID(ctx context.Context, id string) (*models.ElementComment, error)
	GetElementComments(ctx context.Context, elementID string, filter models.ElementCommentFilter) ([]models.ElementComment, error)
	UpdateElementComment(ctx context.Context, id string, userID string, updates map[string]any) (*models.ElementComment, error)
	DeleteElementComment(ctx context.Context, id string, userID string) error
	GetElementCommentsByAuthorID(ctx context.Context, authorID string, limit int, offset int) ([]models.ElementComment, error)
	CountElementComments(ctx context.Context, elementID string) (int64, error)
	ToggleResolvedStatus(ctx context.Context, id string) error
	DeleteElementCommentsByElementID(ctx context.Context, elementID string) error
	GetElementCommentsByProjectID(ctx context.Context, projectID string, limit int, offset int) ([]models.ElementComment, error)
	CountElementCommentsByProjectID(ctx context.Context, projectID string) (int64, error)
}

type MarketplaceServiceInterface interface {
	CreateMarketplaceItem(ctx context.Context, item models.MarketplaceItem, tagIds []string, categoryIds []string) (*models.MarketplaceItem, error)
	GetMarketplaceItems(ctx context.Context, filter repositories.MarketplaceFilter) ([]models.MarketplaceItem, int64, error)
	GetMarketplaceItemByID(ctx context.Context, id string) (*models.MarketplaceItem, error)
	UpdateMarketplaceItem(ctx context.Context, id string, item *models.MarketplaceItem, userId string) (*models.MarketplaceItem, error)
	DeleteMarketplaceItem(ctx context.Context, id string, userId string) error
	DownloadMarketplaceItem(ctx context.Context, itemID, userID string) error
	IncrementDownloads(ctx context.Context, itemID string) error
	IncrementLikes(ctx context.Context, itemID string) error
	CreateCategory(ctx context.Context, category *models.Category) (*models.Category, error)
	GetCategories(ctx context.Context) ([]models.Category, error)
	GetCategoryByID(ctx context.Context, id string) (*models.Category, error)
	GetCategoryByName(ctx context.Context, name string) (*models.Category, error)
	DeleteCategory(ctx context.Context, id string) error
	CreateTag(ctx context.Context, tag *models.Tag) (*models.Tag, error)
	GetTags(ctx context.Context) ([]models.Tag, error)
	GetTagByID(ctx context.Context, id string) (*models.Tag, error)
	GetTagByName(ctx context.Context, name string) (*models.Tag, error)
	DeleteTag(ctx context.Context, id string) error
}
