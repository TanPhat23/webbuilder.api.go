package services

import (
	"context"
	"mime/multipart"
	"my-go-app/internal/models"
)

type EmailServiceInterface interface {
	SendInvitationEmail(to, projectName, inviteLink string) error
}

type InvitationServiceInterface interface {
	Create(ctx context.Context, projectID, email string, role models.CollaboratorRole, invitedBy string) (*models.Invitation, error)
	Accept(ctx context.Context, token, userID string) error
	ListByProject(ctx context.Context, projectID, userID string) ([]models.Invitation, error)
	GetByID(ctx context.Context, id string) (*models.Invitation, error)
	Delete(ctx context.Context, id string) error
	RequireProjectOwnership(ctx context.Context, projectID, userID string) error
	Cancel(ctx context.Context, id string) error
	UpdateStatus(ctx context.Context, id string, status models.InvitationStatus) error
	ListPendingByProject(ctx context.Context, projectID, userID string) ([]models.Invitation, error)
}

type UserServiceInterface interface {
	GetByID(ctx context.Context, userID string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	Search(ctx context.Context, query string) ([]models.User, error)
}

type ProjectServiceInterface interface {
	GetPublicByID(ctx context.Context, projectID string) (*models.Project, error)
	GetPublicBySubdomain(ctx context.Context, subdomain string) (*models.Project, error)
	GetByID(ctx context.Context, projectID, userID string) (*models.Project, error)
	GetWithAccess(ctx context.Context, projectID, userID string) (*models.Project, error)
	ListByUser(ctx context.Context, userID string) ([]models.Project, error)
	ListByCollaborator(ctx context.Context, userID string) ([]models.Project, error)
	ListPages(ctx context.Context, projectID, userID string) ([]models.Page, error)
	Create(ctx context.Context, project *models.Project) (*models.Project, error)
	CreateWithPage(ctx context.Context, project *models.Project) (*models.Project, error)
	GetAccess(ctx context.Context, projectID, userID string) (*models.ProjectAccess, error)
	Update(ctx context.Context, projectID, userID string, project *models.Project) (*models.Project, error)
	UpdateFieldsForUser(ctx context.Context, projectID, userID string, updates map[string]any) (*models.Project, error)
	TransferOwnership(ctx context.Context, projectID, currentOwnerID, newOwnerID string) error
	Delete(ctx context.Context, projectID, userID string) error
	HardDelete(ctx context.Context, projectID, userID string) error
	Restore(ctx context.Context, projectID, userID string) error
	ExistsForUser(ctx context.Context, projectID, userID string) (bool, error)
	GetWithLock(ctx context.Context, projectID, userID string) (*models.Project, error)
}

type PageServiceInterface interface {
	ListByProject(ctx context.Context, projectID, userID string) ([]models.Page, error)
	GetByID(ctx context.Context, pageID string) (*models.Page, error)
	GetByIDForUser(ctx context.Context, projectID, pageID, userID string) (*models.Page, error)
	GetPublicByID(ctx context.Context, projectID, pageID string) (*models.Page, error)
	Create(ctx context.Context, page *models.Page) (*models.Page, error)
	CreateForUser(ctx context.Context, page *models.Page, userID string) (*models.Page, error)
	Update(ctx context.Context, pageID string, page *models.Page) (*models.Page, error)
	UpdateFields(ctx context.Context, pageID string, updates map[string]interface{}) error
	UpdateFieldsForUser(ctx context.Context, projectID, pageID, userID string, updates map[string]interface{}) error
	Delete(ctx context.Context, pageID string) error
	DeleteForUser(ctx context.Context, pageID, projectID, userID string) error
}

type ElementServiceInterface interface {
	List(ctx context.Context, projectID string, pageID ...string) ([]models.EditorElement, error)
	ListByPage(ctx context.Context, pageID string) ([]models.Element, error)
	ListByPageIDs(ctx context.Context, projectID string, pageIDs []string) ([]models.EditorElement, error)
	ListByIDs(ctx context.Context, elementIDs []string) ([]models.Element, error)
}

type ImageServiceInterface interface {
	Create(ctx context.Context, image models.Image) (*models.Image, error)
	Upload(ctx context.Context, userID string, imageName *string, file multipart.File) (*models.ImageUploadResponse, error)
	UploadBase64(ctx context.Context, userID string, imageData string, imageName *string) (*models.ImageUploadResponse, error)
	ListByUser(ctx context.Context, userID string) ([]models.Image, error)
	GetByID(ctx context.Context, imageID, userID string) (*models.Image, error)
	PublicURL(ctx context.Context, imageID string) (string, error)
	Delete(ctx context.Context, imageID, userID string) error
	SoftDelete(ctx context.Context, imageID, userID string) error
	ListAll(ctx context.Context, limit, offset int) ([]models.Image, error)
}

type CollaboratorServiceInterface interface {
	Create(ctx context.Context, callerUserID string, collaborator *models.Collaborator) (*models.Collaborator, error)
	ListByProject(ctx context.Context, projectID, userID string) ([]models.Collaborator, error)
	GetByID(ctx context.Context, id, callerUserID string) (*models.Collaborator, error)
	UpdateRole(ctx context.Context, id, callerUserID string, role models.CollaboratorRole) error
	Delete(ctx context.Context, id, callerUserID string) error
	LeaveProject(ctx context.Context, projectID, userID string) error
	IsCollaborator(ctx context.Context, userID, projectID string) (bool, error)
}

type SnapshotServiceInterface interface {
	Save(ctx context.Context, projectID string, snapshot *models.Snapshot) error
	SaveForUser(ctx context.Context, projectID, userID string, snapshot *models.Snapshot) error
	ListByProject(ctx context.Context, projectID, userID string) ([]models.Snapshot, error)
	GetByID(ctx context.Context, id string) (*models.Snapshot, error)
	GetByIDForUser(ctx context.Context, projectID, snapshotID, userID string) (*models.Snapshot, error)
	Delete(ctx context.Context, id string) error
	DeleteForUser(ctx context.Context, snapshotID, projectID, userID string) error
}

type ContentTypeServiceInterface interface {
	List(ctx context.Context) ([]models.ContentType, error)
	ListForUser(ctx context.Context, userID, projectID string) ([]models.ContentType, error)
	GetByID(ctx context.Context, id string) (*models.ContentType, error)
	GetByIDForUser(ctx context.Context, id, userID string) (*models.ContentType, error)
	Create(ctx context.Context, contentType *models.ContentType) (*models.ContentType, error)
	CreateForUser(ctx context.Context, contentType *models.ContentType, userID string) (*models.ContentType, error)
	Update(ctx context.Context, id string, contentType *models.ContentType) (*models.ContentType, error)
	UpdateForUser(ctx context.Context, id, userID string, contentType *models.ContentType) (*models.ContentType, error)
	PatchForUser(ctx context.Context, id, userID string, updates map[string]any) (*models.ContentType, error)
	Delete(ctx context.Context, id string) error
	DeleteForUser(ctx context.Context, id, userID string) error
}

type ContentFieldServiceInterface interface {
	ListByContentType(ctx context.Context, contentTypeID string) ([]models.ContentField, error)
	ListByContentTypeForUser(ctx context.Context, contentTypeID, userID string) ([]models.ContentField, error)
	GetByID(ctx context.Context, id string) (*models.ContentField, error)
	GetByIDForUser(ctx context.Context, id, userID string) (*models.ContentField, error)
	Create(ctx context.Context, field *models.ContentField) (*models.ContentField, error)
	CreateForUser(ctx context.Context, field *models.ContentField, userID string) (*models.ContentField, error)
	Update(ctx context.Context, id string, updates map[string]any) (*models.ContentField, error)
	UpdateForUser(ctx context.Context, id, userID string, updates map[string]any) (*models.ContentField, error)
	Delete(ctx context.Context, id string) error
	DeleteForUser(ctx context.Context, id, userID string) error
}

type ContentItemServiceInterface interface {
	ListByContentType(ctx context.Context, contentTypeID string) ([]models.ContentItem, error)
	ListByContentTypeForUser(ctx context.Context, contentTypeID, userID string) ([]models.ContentItem, error)
	GetByID(ctx context.Context, id string) (*models.ContentItem, error)
	GetByIDForUser(ctx context.Context, id, userID string) (*models.ContentItem, error)
	GetBySlug(ctx context.Context, contentTypeID, slug string) (*models.ContentItem, error)
	ListPublic(ctx context.Context, contentTypeID string, limit int, sortBy, sortOrder string) ([]models.ContentItem, error)
	Create(ctx context.Context, item *models.ContentItem, fieldValues []models.ContentFieldValue) (*models.ContentItem, error)
	CreateForUser(ctx context.Context, item *models.ContentItem, fieldValues []models.ContentFieldValue, userID string) (*models.ContentItem, error)
	Update(ctx context.Context, id string, updates map[string]any, fieldValues []models.ContentFieldValue) (*models.ContentItem, error)
	UpdateForUser(ctx context.Context, id, userID string, updates map[string]any, fieldValues []models.ContentFieldValue) (*models.ContentItem, error)
	Delete(ctx context.Context, id string) error
	DeleteForUser(ctx context.Context, id, userID string) error
}

type CustomElementServiceInterface interface {
	List(ctx context.Context, userID string, isPublic *bool) ([]models.CustomElement, error)
	GetByID(ctx context.Context, id, userID string) (*models.CustomElement, error)
	Create(ctx context.Context, element *models.CustomElement) (*models.CustomElement, error)
	Update(ctx context.Context, id, userID string, updates map[string]any) (*models.CustomElement, error)
	Delete(ctx context.Context, id, userID string) error
	ListPublic(ctx context.Context, category *string, limit, offset int) ([]models.CustomElement, error)
	Duplicate(ctx context.Context, id, userID, newName string) (*models.CustomElement, error)
}

type CustomElementTypeServiceInterface interface {
	List(ctx context.Context) ([]models.CustomElementType, error)
	GetByID(ctx context.Context, id string) (*models.CustomElementType, error)
	GetByName(ctx context.Context, name string) (*models.CustomElementType, error)
	Create(ctx context.Context, ceType *models.CustomElementType) (*models.CustomElementType, error)
	Update(ctx context.Context, id string, updates map[string]any) (*models.CustomElementType, error)
	Delete(ctx context.Context, id string) error
}

type EventWorkflowServiceInterface interface {
	Create(ctx context.Context, workflow *models.EventWorkflow, userID string) (*models.EventWorkflow, error)
	GetByID(ctx context.Context, id string) (*models.EventWorkflow, error)
	ListByProject(ctx context.Context, projectID string) ([]models.EventWorkflow, error)
	ListByProjectWithElements(ctx context.Context, projectID string) ([]models.EventWorkflow, error)
	ListEnabledByProject(ctx context.Context, projectID string) ([]models.EventWorkflow, error)
	ListByName(ctx context.Context, projectID, name string) ([]models.EventWorkflow, error)
	Update(ctx context.Context, id string, workflow *models.EventWorkflow) (*models.EventWorkflow, error)
	SetEnabled(ctx context.Context, id string, enabled bool) error
	Delete(ctx context.Context, id string) error
	DeleteByProject(ctx context.Context, projectID string) error
	CountByProject(ctx context.Context, projectID string) (int64, error)
	ExistsByName(ctx context.Context, projectID, name, excludeID string) (bool, error)
	ListWithFilters(ctx context.Context, projectID, userID string, enabled *bool, searchName string) ([]models.EventWorkflow, error)
}

type ElementEventWorkflowServiceInterface interface {
	Create(ctx context.Context, eew *models.ElementEventWorkflow) (*models.ElementEventWorkflow, error)
	GetByID(ctx context.Context, id string) (*models.ElementEventWorkflow, error)
	ListAll(ctx context.Context) ([]models.ElementEventWorkflow, error)
	ListByElement(ctx context.Context, elementID string) ([]models.ElementEventWorkflow, error)
	ListByWorkflow(ctx context.Context, workflowID string) ([]models.ElementEventWorkflow, error)
	ListByEvent(ctx context.Context, eventName string) ([]models.ElementEventWorkflow, error)
	ListByFilters(ctx context.Context, elementID, workflowID, eventName string) ([]models.ElementEventWorkflow, error)
	Update(ctx context.Context, id string, eew *models.ElementEventWorkflow) (*models.ElementEventWorkflow, error)
	Delete(ctx context.Context, id string) error
	DeleteByElement(ctx context.Context, elementID string) error
	DeleteByWorkflow(ctx context.Context, workflowID string) error
	ListByPage(ctx context.Context, pageID string) ([]models.ElementEventWorkflow, error)
	IsLinked(ctx context.Context, elementID, workflowID, eventName string) (bool, error)
}

type ElementCommentServiceInterface interface {
	Create(ctx context.Context, comment *models.ElementComment) (*models.ElementComment, error)
	GetByID(ctx context.Context, id string) (*models.ElementComment, error)
	GetByIDForUser(ctx context.Context, id, userID string) (*models.ElementComment, error)
	List(ctx context.Context, elementID string, filter models.ElementCommentFilter) ([]models.ElementComment, error)
	ListForUser(ctx context.Context, elementID, userID string, filter models.ElementCommentFilter) ([]models.ElementComment, error)
	Update(ctx context.Context, id string, userID string, updates map[string]any) (*models.ElementComment, error)
	Delete(ctx context.Context, id string, userID string) error
	ListByAuthor(ctx context.Context, authorID string, limit int, offset int) ([]models.ElementComment, error)
	Count(ctx context.Context, elementID string) (int64, error)
	ToggleResolved(ctx context.Context, id string) error
	ToggleResolvedForUser(ctx context.Context, id, userID string) error
	DeleteByElement(ctx context.Context, elementID string) error
	ListByProject(ctx context.Context, projectID, userID string, limit int, offset int) ([]models.ElementComment, error)
	CountByProject(ctx context.Context, projectID, userID string) (int64, error)
}
