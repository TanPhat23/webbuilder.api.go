package routes

import (
	"my-go-app/internal/authz"
	"my-go-app/internal/handlers"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"

	"github.com/gofiber/fiber/v3"
)

func RegisterContent(publicGroup, privateGroup fiber.Router, repos *repositories.Repositories, emitter ...services.ItemEventEmitter) {
	var itemEmitter services.ItemEventEmitter
	if len(emitter) > 0 {
		itemEmitter = emitter[0]
	}
	contentTypeService := services.NewContentTypeService(repos.ContentTypeRepository).WithProjectRepo(repos.ProjectRepository)
	contentFieldService := services.NewContentFieldService(repos.ContentFieldRepository).WithAuth(repos.ContentTypeRepository, repos.ProjectRepository)
	contentItemService := services.NewContentItemService(repos.ContentItemRepository).WithAuth(repos.ContentTypeRepository, repos.ProjectRepository).WithEventEmitter(itemEmitter)

	contentTypeHandler := handlers.NewContentTypeHandler(contentTypeService)
	contentFieldHandler := handlers.NewContentFieldHandler(contentFieldService)
	contentItemHandler := handlers.NewContentItemHandler(contentItemService)
	publicCollectionService := services.NewPublicCollectionService(repos.ContentTypeRepository, repos.ContentItemRepository).WithEventEmitter(itemEmitter)
	publicCollectionHandler := handlers.NewPublicCollectionHandler(publicCollectionService)
	authorizer := services.NewProjectAuthorizer(repos.ProjectRepository)

	publicGroup.Get("/public/content", contentItemHandler.ListPublic)
	publicGroup.Get("/public/content/:contentTypeId/*", contentItemHandler.GetPublicBySlug)
	publicGroup.Get("/public/projects/:projectId/collections/:collection/items", publicCollectionHandler.ListPublicItems)
	publicGroup.Post("/public/projects/:projectId/collections/:collection/items", publicCollectionHandler.SubmitPublicItem)

	privateGroup.Get("/projects/:projectid/content-types", authorizer.RequireProjectPermission("projectid", authz.PermCMSView), contentTypeHandler.List)
	privateGroup.Post("/projects/:projectid/content-types", authorizer.RequireProjectPermission("projectid", authz.PermCMSEdit), contentTypeHandler.Create)

	privateGroup.Post("/content-types", contentTypeHandler.Create)
	privateGroup.Get("/content-types", contentTypeHandler.List)
	privateGroup.Get("/content-types/:id", contentTypeHandler.GetByID)
	privateGroup.Patch("/content-types/:id", contentTypeHandler.Update)
	privateGroup.Delete("/content-types/:id", contentTypeHandler.Delete)

	// Fields and items nested under a content type (the shape the web UI calls).
	privateGroup.Post("/content-types/:contentTypeId/fields", contentFieldHandler.Create)
	privateGroup.Get("/content-types/:contentTypeId/fields", contentFieldHandler.ListByContentType)
	privateGroup.Get("/content-types/:contentTypeId/fields/:fieldId", contentFieldHandler.GetByID)
	privateGroup.Patch("/content-types/:contentTypeId/fields/:fieldId", contentFieldHandler.Update)
	privateGroup.Delete("/content-types/:contentTypeId/fields/:fieldId", contentFieldHandler.Delete)
	privateGroup.Post("/content-types/:contentTypeId/items", contentItemHandler.Create)
	privateGroup.Get("/content-types/:contentTypeId/items", contentItemHandler.ListByContentType)
	privateGroup.Get("/content-types/:contentTypeId/items/:itemId", contentItemHandler.GetByID)
	privateGroup.Patch("/content-types/:contentTypeId/items/:itemId", contentItemHandler.Update)
	privateGroup.Delete("/content-types/:contentTypeId/items/:itemId", contentItemHandler.Delete)

	// Flat forms. The param names must match what the handlers read.
	privateGroup.Get("/content-fields/:contentTypeId", contentFieldHandler.ListByContentType)
	privateGroup.Get("/content-fields/by-id/:fieldId", contentFieldHandler.GetByID)
	privateGroup.Patch("/content-fields/:fieldId", contentFieldHandler.Update)
	privateGroup.Delete("/content-fields/:fieldId", contentFieldHandler.Delete)

	privateGroup.Get("/content-items/:contentTypeId", contentItemHandler.ListByContentType)
	privateGroup.Get("/content-items/by-id/:itemId", contentItemHandler.GetByID)
	privateGroup.Patch("/content-items/:itemId", contentItemHandler.Update)
	privateGroup.Delete("/content-items/:itemId", contentItemHandler.Delete)
}
