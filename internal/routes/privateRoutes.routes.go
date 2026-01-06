package routes

import (
	"my-go-app/internal/handlers"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"
	"my-go-app/pkg/middleware"

	"github.com/gofiber/fiber/v2"
)

func PrivateRoutes(app *fiber.App, repos *repositories.RepositoriesInterface, cloudinaryService *services.CloudinaryService, invitationService *services.InvitationService) {
	elementHandler := handlers.NewElementHandler(repos.ElementRepository)
	projectHandler := handlers.NewProjectHandler(repos.ProjectRepository)
	pageHandler := handlers.NewPageHandler(repos.PageRepository)
	snapshotHandler := handlers.NewSnapshotHandler(repos.SnapshotRepository, repos.ElementRepository, repos.ProjectRepository)
	contentTypeHandler := handlers.NewContentTypeHandler(repos.ContentTypeRepository)
	contentFieldHandler := handlers.NewContentFieldHandler(repos.ContentFieldRepository)
	contentItemHandler := handlers.NewContentItemHandler(repos.ContentItemRepository)
	imageHandler := handlers.NewImageHandler(repos.ImageRepository, cloudinaryService)
	marketplaceHandler := handlers.NewMarketplaceHandler(repos.MarketplaceRepository)
	customElementHandler := handlers.NewCustomElementHandler(repos.CustomElementRepository)
	customElementTypeHandler := handlers.NewCustomElementTypeHandler(repos.CustomElementTypeRepository)
	invitationHandler := handlers.NewInvitationHandler(invitationService)
	collaboratorHandler := handlers.NewCollaboratorHandler(repos.CollaboratorRepository, repos.ProjectRepository)
	commentHandler := handlers.NewCommentHandler(repos.CommentRepository, repos.MarketplaceRepository)
	elementCommentHandler := handlers.NewElementCommentHandler(repos.ElementCommentRepository)
	userHandler := handlers.NewUserHandler(repos.UserRepository)
	eventWorkflowHandler := handlers.NewEventWorkflowHandler(repos.EventWorkflowRepository, repos.ProjectRepository, repos.ElementRepository, repos.ElementEventWorkflowRepository)
	elementEventWorkflowHandler := handlers.NewElementEventWorkflowHandler(repos.ElementEventWorkflowRepository, repos.ElementRepository, repos.EventWorkflowRepository, repos.ProjectRepository)

	group := app.Group("/api/v1", middleware.AuthenticateMiddleware)

	group.Get("/elements/:projectid", elementHandler.GetElements)
	group.Get("/elements/by-pages", elementHandler.GetElementsByPageIds)

	group.Get("/projects/user", projectHandler.GetProjectsByUser)
	group.Get("/projects/:projectid", projectHandler.GetProjectByID)
	group.Get("/projects/:projectid/pages", projectHandler.GetProjectPages)
	group.Delete("/projects/:projectid", projectHandler.DeleteProject)
	group.Patch("/projects/:projectid", projectHandler.UpdateProject)
	group.Delete("/projects/:projectid/pages/:pageid", pageHandler.DeletePage)

	// Dedicated page routes
	group.Get("/pages/:projectid", pageHandler.GetPagesByProjectID)
	group.Get("/pages/:projectid/:pageid", pageHandler.GetPageByID)
	group.Post("/pages/:projectid", pageHandler.CreatePage)
	group.Patch("/pages/:projectid/:pageid", pageHandler.UpdatePage)

	group.Post("/snapshots/:projectid/save", snapshotHandler.SaveSnapshot)
	group.Get("/snapshots/:projectid", snapshotHandler.GetSnapshots)
	group.Get("/snapshots/:projectid/:snapshotid", snapshotHandler.GetSnapshotByID)
	group.Delete("/snapshots/:projectid/:snapshotid", snapshotHandler.DeleteSnapshot)

	group.Get("/content-types", contentTypeHandler.GetContentTypes)
	group.Post("/content-types", contentTypeHandler.CreateContentType)
	group.Get("/content-types/:id", contentTypeHandler.GetContentTypeByID)
	group.Patch("/content-types/:id", contentTypeHandler.UpdateContentType)
	group.Delete("/content-types/:id", contentTypeHandler.DeleteContentType)

	group.Get("/content-types/:contentTypeId/fields", contentFieldHandler.GetContentFieldsByContentType)
	group.Post("/content-types/:contentTypeId/fields", contentFieldHandler.CreateContentField)
	group.Get("/content-types/:contentTypeId/fields/:fieldId", contentFieldHandler.GetContentFieldByID)
	group.Patch("/content-types/:contentTypeId/fields/:fieldId", contentFieldHandler.UpdateContentField)
	group.Delete("/content-types/:contentTypeId/fields/:fieldId", contentFieldHandler.DeleteContentField)

	group.Get("/content-types/:contentTypeId/items", contentItemHandler.GetContentItemsByContentType)
	group.Post("/content-types/:contentTypeId/items", contentItemHandler.CreateContentItem)
	group.Get("/content-types/:contentTypeId/items/:itemId", contentItemHandler.GetContentItemByID)
	group.Patch("/content-types/:contentTypeId/items/:itemId", contentItemHandler.UpdateContentItem)
	group.Delete("/content-types/:contentTypeId/items/:itemId", contentItemHandler.DeleteContentItem)

	// Image upload routes
	group.Post("/images", imageHandler.UploadImage)
	group.Post("/images/base64", imageHandler.UploadBase64Image)
	group.Get("/images", imageHandler.GetUserImages)
	group.Get("/images/:imageid", imageHandler.GetImageByID)
	group.Delete("/images/:imageid", imageHandler.DeleteImage)

	// Marketplace routes
	group.Post("/marketplace/items", marketplaceHandler.CreateMarketplaceItem)
	group.Get("/marketplace/items", marketplaceHandler.GetMarketplaceItems)
	group.Get("/marketplace/items/:itemid", marketplaceHandler.GetMarketplaceItemByID)
	group.Patch("/marketplace/items/:itemid", marketplaceHandler.UpdateMarketplaceItem)
	group.Delete("/marketplace/items/:itemid", marketplaceHandler.DeleteMarketplaceItem)
	group.Post("/marketplace/items/:itemid/download", marketplaceHandler.DownloadMarketplaceItem)
	group.Post("/marketplace/items/:itemid/like", marketplaceHandler.IncrementLikes)

	// Category routes
	group.Post("/marketplace/categories", marketplaceHandler.CreateCategory)
	group.Get("/marketplace/categories", marketplaceHandler.GetCategories)
	group.Delete("/marketplace/categories/:categoryid", marketplaceHandler.DeleteCategory)

	// Tag routes
	group.Post("/marketplace/tags", marketplaceHandler.CreateTag)
	group.Get("/marketplace/tags", marketplaceHandler.GetTags)
	group.Delete("/marketplace/tags/:tagid", marketplaceHandler.DeleteTag)

	// Comment routes
	group.Post("/comments", commentHandler.CreateComment)
	group.Get("/comments", commentHandler.GetComments)
	group.Get("/comments/:commentid", commentHandler.GetCommentByID)
	group.Patch("/comments/:commentid", commentHandler.UpdateComment)
	group.Delete("/comments/:commentid", commentHandler.DeleteComment)

	// Comment reactions
	group.Post("/comments/:commentid/reactions", commentHandler.CreateReaction)
	group.Delete("/comments/:commentid/reactions", commentHandler.DeleteReaction)
	group.Get("/comments/:commentid/reactions", commentHandler.GetReactionsByCommentID)
	group.Get("/comments/:commentid/reactions/summary", commentHandler.GetReactionSummary)

	// Marketplace item comments
	group.Get("/marketplace/items/:itemid/comments", commentHandler.GetCommentsByItemID)
	group.Get("/marketplace/items/:itemid/comments/count", commentHandler.GetCommentCount)

	// Comment moderation
	group.Patch("/comments/:commentid/moderate", commentHandler.ModerateComment)

	// Custom element routes
	group.Get("/customelements", customElementHandler.GetCustomElements)
	group.Get("/customelements/public", customElementHandler.GetPublicCustomElements)
	group.Get("/customelements/:id", customElementHandler.GetCustomElementByID)
	group.Post("/customelements", customElementHandler.CreateCustomElement)
	group.Patch("/customelements/:id", customElementHandler.UpdateCustomElement)
	group.Delete("/customelements/:id", customElementHandler.DeleteCustomElement)
	group.Post("/customelements/:id/duplicate", customElementHandler.DuplicateCustomElement)

	// Custom element type routes
	group.Get("/customelementtypes", customElementTypeHandler.GetCustomElementTypes)
	group.Get("/customelementtypes/:id", customElementTypeHandler.GetCustomElementTypeByID)
	group.Post("/customelementtypes", customElementTypeHandler.CreateCustomElementType)
	group.Patch("/customelementtypes/:id", customElementTypeHandler.UpdateCustomElementType)
	group.Delete("/customelementtypes/:id", customElementTypeHandler.DeleteCustomElementType)

	// Invitation routes
	group.Post("/invitations", invitationHandler.CreateInvitation)
	group.Get("/invitations/project/:projectid", invitationHandler.GetInvitationsByProject)
	group.Get("/invitations/project/:projectid/pending", invitationHandler.GetPendingInvitationsByProject)
	group.Post("/invitations/accept", invitationHandler.AcceptInvitation)
	group.Patch("/invitations/:invitationid/cancel", invitationHandler.CancelInvitation)
	group.Patch("/invitations/:invitationid/status", invitationHandler.UpdateInvitationStatus)
	group.Delete("/invitations/:invitationid", invitationHandler.DeleteInvitation)

	// Collaborator routes
	group.Get("/collaborators/project/:projectid", collaboratorHandler.GetCollaboratorsByProject)
	group.Get("/collaborators/:collaboratorid", collaboratorHandler.GetCollaboratorByID)
	group.Patch("/collaborators/:collaboratorid/role", collaboratorHandler.UpdateCollaboratorRole)
	group.Delete("/collaborators/:collaboratorid", collaboratorHandler.DeleteCollaborator)

	// Element Comment routes
	group.Post("/element-comments", elementCommentHandler.CreateElementComment)
	group.Get("/element-comments/:id", elementCommentHandler.GetElementCommentByID)
	group.Get("/elements/:elementId/comments", elementCommentHandler.GetElementComments)
	group.Patch("/element-comments/:id", elementCommentHandler.UpdateElementComment)
	group.Delete("/element-comments/:id", elementCommentHandler.DeleteElementComment)
	group.Patch("/element-comments/:id/toggle-resolved", elementCommentHandler.ToggleResolvedStatus)
	group.Get("/element-comments/author/:authorId", elementCommentHandler.GetCommentsByAuthorID)
	group.Get("/projects/:projectId/comments", elementCommentHandler.GetCommentsByProjectID)

	// User routes
	group.Get("/users/search", userHandler.SearchUsers)
	group.Get("/users/email/:email", userHandler.GetUserByEmail)
	group.Get("/users/username/:username", userHandler.GetUserByUsername)

	// Element Event Workflow routes
	group.Post("/element-event-workflows", elementEventWorkflowHandler.CreateElementEventWorkflow)
	group.Get("/element-event-workflows", elementEventWorkflowHandler.GetElementEventWorkflows)
	group.Get("/element-event-workflows/:id", elementEventWorkflowHandler.GetElementEventWorkflowByID)
	group.Patch("/element-event-workflows/:id", elementEventWorkflowHandler.UpdateElementEventWorkflow)
	group.Delete("/element-event-workflows/:id", elementEventWorkflowHandler.DeleteElementEventWorkflow)
	group.Delete("/element-event-workflows/element/:elementId", elementEventWorkflowHandler.DeleteElementEventWorkflowsByElement)
	group.Delete("/element-event-workflows/workflow/:workflowId", elementEventWorkflowHandler.DeleteElementEventWorkflowsByWorkflow)

	// Event Workflow routes
	group.Post("/event-workflows", eventWorkflowHandler.CreateEventWorkflow)
	group.Get("/projects/:projectid/event-workflows", eventWorkflowHandler.GetEventWorkflowsByProject)
	group.Get("/event-workflows/:id", eventWorkflowHandler.GetEventWorkflowByID)
	group.Patch("/event-workflows/:id", eventWorkflowHandler.UpdateEventWorkflow)
	group.Patch("/event-workflows/:id/enabled", eventWorkflowHandler.UpdateEventWorkflowEnabled)
	group.Delete("/event-workflows/:id", eventWorkflowHandler.DeleteEventWorkflow)
	group.Get("/event-workflows/:id/elements", eventWorkflowHandler.GetEventWorkflowElements)
}
