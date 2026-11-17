package routes

import (
	"my-go-app/internal/agenttokens"
	"my-go-app/internal/handlers"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"
	"my-go-app/pkg/middleware"

	"github.com/gofiber/fiber/v2"
)

// PrivateRoutes registers every authenticated /api/v1 route.
//
// Fiber matches routes in registration order, so the order of the mount calls
// below (and of the routes inside each mount method) is significant: do not
// reorder them without checking for overlapping patterns.
func PrivateRoutes(app *fiber.App, repos *repositories.RepositoriesInterface, cloudinaryService *services.CloudinaryService, invitationService *services.InvitationService) {
	h := newPrivateHandlers(repos, cloudinaryService, invitationService)

	group := app.Group("/api/v1", middleware.AuthenticateMiddleware)
	h.mountAgentTokens(group)
	h.mountElements(group)
	h.mountProjects(group)
	h.mountPages(group)
	h.mountSnapshots(group)
	h.mountContentTypes(group)
	h.mountContentFields(group)
	h.mountContentItems(group)
	h.mountImages(group)
	h.mountMarketplace(group)
	h.mountCustomElements(group)
	h.mountCustomElementTypes(group)
	h.mountInvitations(group)
	h.mountCollaborators(group)
	h.mountComments(group)
	h.mountElementComments(group)
	h.mountUsers(group)
	h.mountNotifications(group)
	h.mountEventWorkflows(group)
	h.mountElementEventWorkflows(group)
}

// privateHandlers holds the handlers behind the authenticated routes.
type privateHandlers struct {
	element              *handlers.ElementHandler
	project              *handlers.ProjectHandler
	page                 *handlers.PageHandler
	snapshot             *handlers.SnapshotHandler
	contentType          *handlers.ContentTypeHandler
	contentField         *handlers.ContentFieldHandler
	contentItem          *handlers.ContentItemHandler
	image                *handlers.ImageHandler
	marketplace          *handlers.MarketplaceHandler
	customElement        *handlers.CustomElementHandler
	customElementType    *handlers.CustomElementTypeHandler
	invitation           *handlers.InvitationHandler
	collaborator         *handlers.CollaboratorHandler
	comment              *handlers.CommentHandler
	elementComment       *handlers.ElementCommentHandler
	user                 *handlers.UserHandler
	eventWorkflow        *handlers.EventWorkflowHandler
	notification         *handlers.NotificationHandler
	agentToken           *handlers.AgentTokenHandler
	elementEventWorkflow *handlers.ElementEventWorkflowHandler
}

// newPrivateHandlers wires services and handlers on top of the repositories.
func newPrivateHandlers(repos *repositories.RepositoriesInterface, cloudinaryService *services.CloudinaryService, invitationService *services.InvitationService) *privateHandlers {
	elementService := services.NewElementWrapperService(repos.ElementRepository)
	projectService := services.NewProjectService(repos.ProjectRepository, repos.CollaboratorRepository, repos.UserRepository)
	pageService := services.NewPageService(repos.PageRepository, repos.ProjectRepository)
	snapshotService := services.NewSnapshotService(repos.SnapshotRepository, repos.ElementRepository, repos.ProjectRepository)
	contentTypeService := services.NewContentTypeService(repos.ContentTypeRepository)
	contentFieldService := services.NewContentFieldService(repos.ContentFieldRepository)
	contentItemService := services.NewContentItemService(repos.ContentItemRepository)
	imageService := services.NewImageService(repos.ImageRepository, cloudinaryService)
	marketplaceService := services.NewMarketplaceService(repos.MarketplaceRepository)
	customElementService := services.NewCustomElementService(repos.CustomElementRepository)
	customElementTypeService := services.NewCustomElementTypeService(repos.CustomElementTypeRepository)
	collaboratorService := services.NewCollaboratorService(repos.CollaboratorRepository, repos.ProjectRepository)
	commentService := services.NewCommentService(repos.CommentRepository, repos.MarketplaceRepository)
	elementCommentService := services.NewElementCommentService(repos.ElementCommentRepository, repos.ElementRepository, repos.ProjectRepository)
	userService := services.NewUserService(repos.UserRepository)
	eventWorkflowService := services.NewEventWorkflowService(repos.EventWorkflowRepository, repos.ProjectRepository)
	elementEventWorkflowService := services.NewElementEventWorkflowService(repos.ElementEventWorkflowRepository, repos.ElementRepository, repos.ProjectRepository, repos.PageRepository)

	return &privateHandlers{
		element:              handlers.NewElementHandler(elementService),
		project:              handlers.NewProjectHandler(projectService),
		page:                 handlers.NewPageHandler(pageService),
		snapshot:             handlers.NewSnapshotHandler(snapshotService, repos.ElementRepository),
		contentType:          handlers.NewContentTypeHandler(contentTypeService),
		contentField:         handlers.NewContentFieldHandler(contentFieldService),
		contentItem:          handlers.NewContentItemHandler(contentItemService),
		image:                handlers.NewImageHandler(imageService, cloudinaryService),
		marketplace:          handlers.NewMarketplaceHandler(marketplaceService),
		customElement:        handlers.NewCustomElementHandler(customElementService),
		customElementType:    handlers.NewCustomElementTypeHandler(customElementTypeService),
		invitation:           handlers.NewInvitationHandler(invitationService),
		collaborator:         handlers.NewCollaboratorHandler(collaboratorService),
		comment:              handlers.NewCommentHandler(commentService),
		elementComment:       handlers.NewElementCommentHandler(elementCommentService),
		user:                 handlers.NewUserHandler(userService),
		eventWorkflow:        handlers.NewEventWorkflowHandler(eventWorkflowService, elementEventWorkflowService),
		notification:         handlers.NewNotificationHandler(repos.NotificationRepository),
		agentToken:           handlers.NewAgentTokenHandler(agenttokens.NewService(repos.AgentTokenRepository), handlers.PublicMCPURL()),
		elementEventWorkflow: handlers.NewElementEventWorkflowHandler(elementEventWorkflowService, projectService, pageService),
	}
}

func (h *privateHandlers) mountAgentTokens(g fiber.Router) {
	g.Post("/agent-tokens", h.agentToken.Create)
	g.Get("/agent-tokens", h.agentToken.List)
	g.Delete("/agent-tokens/:id", h.agentToken.Revoke)
}

func (h *privateHandlers) mountElements(g fiber.Router) {
	g.Get("/elements/by-pages", h.element.GetElementsByPageIds)
	g.Get("/elements/:projectid", h.element.GetElements)
}

func (h *privateHandlers) mountProjects(g fiber.Router) {
	g.Post("/projects", h.project.CreateProject)
	g.Get("/projects/user", h.project.GetProjectsByUser)
	g.Get("/projects/:projectid/access", h.project.GetProjectAccess)
	g.Get("/projects/:projectid", h.project.GetProjectByID)
	g.Get("/projects/:projectid/pages", h.project.GetProjectPages)
	g.Delete("/projects/:projectid", h.project.DeleteProject)
	g.Patch("/projects/:projectid", h.project.UpdateProject)
	g.Delete("/projects/:projectid/pages/:pageid", h.page.DeletePage)
}

func (h *privateHandlers) mountPages(g fiber.Router) {
	g.Get("/pages/:projectid", h.page.GetPagesByProjectID)
	g.Get("/pages/:projectid/:pageid", h.page.GetPageByID)
	g.Post("/pages/:projectid", h.page.CreatePage)
	g.Patch("/pages/:projectid/:pageid", h.page.UpdatePage)
}

func (h *privateHandlers) mountSnapshots(g fiber.Router) {
	g.Post("/snapshots/:projectid/save", h.snapshot.SaveSnapshot)
	g.Get("/snapshots/:projectid", h.snapshot.GetSnapshots)
	g.Get("/snapshots/:projectid/:snapshotid", h.snapshot.GetSnapshotByID)
	g.Delete("/snapshots/:projectid/:snapshotid", h.snapshot.DeleteSnapshot)
}

func (h *privateHandlers) mountContentTypes(g fiber.Router) {
	g.Post("/content-types", h.contentType.CreateContentType)
	g.Get("/content-types", h.contentType.GetContentTypes)
	g.Get("/content-types/:id", h.contentType.GetContentTypeByID)
	g.Patch("/content-types/:id", h.contentType.UpdateContentType)
	g.Delete("/content-types/:id", h.contentType.DeleteContentType)

	// Fields and items nested under a content type (the shape the web UI calls).
	g.Post("/content-types/:contentTypeId/fields", h.contentField.CreateContentField)
	g.Get("/content-types/:contentTypeId/fields", h.contentField.GetContentFieldsByContentType)
	g.Get("/content-types/:contentTypeId/fields/:fieldId", h.contentField.GetContentFieldByID)
	g.Patch("/content-types/:contentTypeId/fields/:fieldId", h.contentField.UpdateContentField)
	g.Delete("/content-types/:contentTypeId/fields/:fieldId", h.contentField.DeleteContentField)
	g.Post("/content-types/:contentTypeId/items", h.contentItem.CreateContentItem)
	g.Get("/content-types/:contentTypeId/items", h.contentItem.GetContentItemsByContentType)
	g.Get("/content-types/:contentTypeId/items/:itemId", h.contentItem.GetContentItemByID)
	g.Patch("/content-types/:contentTypeId/items/:itemId", h.contentItem.UpdateContentItem)
	g.Delete("/content-types/:contentTypeId/items/:itemId", h.contentItem.DeleteContentItem)
}

func (h *privateHandlers) mountContentFields(g fiber.Router) {
	g.Get("/content-fields/:contentTypeId", h.contentField.GetContentFieldsByContentType)
	g.Get("/content-fields/by-id/:fieldId", h.contentField.GetContentFieldByID)
	g.Patch("/content-fields/:fieldId", h.contentField.UpdateContentField)
	g.Delete("/content-fields/:fieldId", h.contentField.DeleteContentField)
}

func (h *privateHandlers) mountContentItems(g fiber.Router) {
	g.Get("/content-items/:contentTypeId", h.contentItem.GetContentItemsByContentType)
	g.Get("/content-items/by-id/:itemId", h.contentItem.GetContentItemByID)
	g.Patch("/content-items/:itemId", h.contentItem.UpdateContentItem)
	g.Delete("/content-items/:itemId", h.contentItem.DeleteContentItem)
}

func (h *privateHandlers) mountImages(g fiber.Router) {
	g.Post("/images", h.image.UploadImage)
	g.Get("/images", h.image.GetUserImages)
	g.Get("/images/:imageid", h.image.GetImageByID)
	g.Delete("/images/:imageid", h.image.DeleteImage)
}

func (h *privateHandlers) mountMarketplace(g fiber.Router) {
	g.Get("/analytics/marketplace", h.marketplace.GetMyItemsForAnalytics)
	g.Post("/marketplace", h.marketplace.CreateMarketplaceItem)
	g.Get("/marketplace", h.marketplace.GetMarketplaceItems)
	g.Get("/marketplace/:itemid", h.marketplace.GetMarketplaceItemByID)
	g.Patch("/marketplace/:itemid", h.marketplace.UpdateMarketplaceItem)
	g.Delete("/marketplace/:itemid", h.marketplace.DeleteMarketplaceItem)
}

func (h *privateHandlers) mountCustomElements(g fiber.Router) {
	g.Post("/custom-elements", h.customElement.CreateCustomElement)
	g.Get("/custom-elements", h.customElement.GetCustomElements)
	g.Get("/custom-elements/:id", h.customElement.GetCustomElementByID)
	g.Patch("/custom-elements/:id", h.customElement.UpdateCustomElement)
	g.Delete("/custom-elements/:id", h.customElement.DeleteCustomElement)
}

func (h *privateHandlers) mountCustomElementTypes(g fiber.Router) {
	g.Post("/custom-element-types", h.customElementType.CreateCustomElementType)
	g.Get("/custom-element-types", h.customElementType.GetCustomElementTypes)
	g.Get("/custom-element-types/:id", h.customElementType.GetCustomElementTypeByID)
	g.Patch("/custom-element-types/:id", h.customElementType.UpdateCustomElementType)
	g.Delete("/custom-element-types/:id", h.customElementType.DeleteCustomElementType)
}

func (h *privateHandlers) mountInvitations(g fiber.Router) {
	g.Post("/invitations", h.invitation.CreateInvitation)
	g.Post("/invitations/accept", h.invitation.AcceptInvitation)
	g.Post("/invitations/:token/accept", h.invitation.AcceptInvitation)
	g.Get("/invitations", h.invitation.GetInvitationsByProject)
	g.Get("/invitations/project/:projectid", h.invitation.GetInvitationsByProject)
	g.Get("/invitations/project/:projectid/pending", h.invitation.GetPendingInvitationsByProject)
	g.Patch("/invitations/:invitationid/cancel", h.invitation.CancelInvitation)
	g.Patch("/invitations/:invitationid/status", h.invitation.UpdateInvitationStatus)
	g.Delete("/invitations/:invitationid", h.invitation.DeleteInvitation)
}

func (h *privateHandlers) mountCollaborators(g fiber.Router) {
	g.Post("/collaborators", h.collaborator.CreateCollaborator)
	g.Get("/collaborators/:projectid", h.collaborator.GetCollaboratorsByProject)
	g.Get("/collaborators/project/:projectid", h.collaborator.GetCollaboratorsByProject)
	g.Get("/collaborators/by-id/:collaboratorid", h.collaborator.GetCollaboratorByID)
	g.Patch("/collaborators/:collaboratorid", h.collaborator.UpdateCollaboratorRole)
	g.Patch("/collaborators/:collaboratorid/role", h.collaborator.UpdateCollaboratorRole)
	g.Delete("/collaborators/:id", h.collaborator.DeleteCollaborator)
}

func (h *privateHandlers) mountComments(g fiber.Router) {
	g.Post("/comments", h.comment.CreateComment)
	g.Get("/comments/:commentid", h.comment.GetCommentByID)
	g.Get("/comments", h.comment.GetComments)
	g.Patch("/comments/:commentid", h.comment.UpdateComment)
	g.Delete("/comments/:commentid", h.comment.DeleteComment)
}

func (h *privateHandlers) mountElementComments(g fiber.Router) {
	g.Post("/element-comments", h.elementComment.CreateElementComment)
	g.Get("/element-comments/:id", h.elementComment.GetElementCommentByID)
	g.Get("/element-comments", h.elementComment.GetElementComments)
	g.Get("/elements/:elementId/comments", h.elementComment.GetElementComments)
	g.Get("/element-comments/author/:authorId", h.elementComment.GetCommentsByAuthorID)
	g.Get("/projects/:projectId/comments", h.elementComment.GetCommentsByProjectID)
	g.Patch("/element-comments/:id", h.elementComment.UpdateElementComment)
	g.Patch("/element-comments/:id/toggle-resolved", h.elementComment.ToggleResolvedStatus)
	g.Delete("/element-comments/:id", h.elementComment.DeleteElementComment)
}

func (h *privateHandlers) mountUsers(g fiber.Router) {
	g.Get("/users/search", h.user.SearchUsers)
	g.Get("/users/email/:email", h.user.GetUserByEmail)
	g.Get("/users/username/:username", h.user.GetUserByUsername)
	g.Get("/users/:userid", h.user.SearchUsers)
}

func (h *privateHandlers) mountNotifications(g fiber.Router) {
	g.Get("/notifications", h.notification.GetNotifications)
	g.Post("/notifications/read-all", h.notification.MarkAllRead)
	g.Post("/notifications/profile-update", h.notification.CreateProfileUpdate)
	g.Patch("/notifications/:id", h.notification.SetNotificationRead)
}

func (h *privateHandlers) mountEventWorkflows(g fiber.Router) {
	g.Get("/event-workflows/:projectid", h.eventWorkflow.GetEventWorkflowsByProject)
	g.Get("/event-workflows/by-id/:id", h.eventWorkflow.GetEventWorkflowByID)
	g.Get("/event-workflows/:id/elements", h.eventWorkflow.GetEventWorkflowElements)
}

func (h *privateHandlers) mountElementEventWorkflows(g fiber.Router) {
	g.Get("/element-event-workflows/:id", h.elementEventWorkflow.GetElementEventWorkflowByID)
	g.Get("/element-event-workflows", h.elementEventWorkflow.GetElementEventWorkflows)
}
