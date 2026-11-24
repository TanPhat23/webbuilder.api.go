package routes

import (
	"my-go-app/internal/authz"
	"my-go-app/internal/handlers"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"

	"github.com/gofiber/fiber/v3"
)

func ContentRoutes(publicGroup, privateGroup fiber.Router, repos *repositories.RepositoriesInterface, emitter ...services.ItemEventEmitter) {
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

	publicGroup.Get("/public/content", contentItemHandler.GetPublicContentItems)
	publicGroup.Get("/public/content/:contentTypeId/:slug", contentItemHandler.GetPublicContentItemBySlug)
	publicGroup.Get("/public/projects/:projectId/collections/:collection/items", publicCollectionHandler.ListPublicItems)
	publicGroup.Post("/public/projects/:projectId/collections/:collection/items", publicCollectionHandler.SubmitPublicItem)

	privateGroup.Get("/projects/:projectid/content-types", authorizer.RequireProjectPermission("projectid", authz.PermCMSView), contentTypeHandler.GetContentTypes)
	privateGroup.Post("/projects/:projectid/content-types", authorizer.RequireProjectPermission("projectid", authz.PermCMSEdit), contentTypeHandler.CreateContentType)

	privateGroup.Post("/content-types", contentTypeHandler.CreateContentType)
	privateGroup.Get("/content-types", contentTypeHandler.GetContentTypes)
	privateGroup.Get("/content-types/:id", contentTypeHandler.GetContentTypeByID)
	privateGroup.Patch("/content-types/:id", contentTypeHandler.UpdateContentType)
	privateGroup.Delete("/content-types/:id", contentTypeHandler.DeleteContentType)

	// Fields and items nested under a content type (the shape the web UI calls).
	privateGroup.Post("/content-types/:contentTypeId/fields", contentFieldHandler.CreateContentField)
	privateGroup.Get("/content-types/:contentTypeId/fields", contentFieldHandler.GetContentFieldsByContentType)
	privateGroup.Get("/content-types/:contentTypeId/fields/:fieldId", contentFieldHandler.GetContentFieldByID)
	privateGroup.Patch("/content-types/:contentTypeId/fields/:fieldId", contentFieldHandler.UpdateContentField)
	privateGroup.Delete("/content-types/:contentTypeId/fields/:fieldId", contentFieldHandler.DeleteContentField)
	privateGroup.Post("/content-types/:contentTypeId/items", contentItemHandler.CreateContentItem)
	privateGroup.Get("/content-types/:contentTypeId/items", contentItemHandler.GetContentItemsByContentType)
	privateGroup.Get("/content-types/:contentTypeId/items/:itemId", contentItemHandler.GetContentItemByID)
	privateGroup.Patch("/content-types/:contentTypeId/items/:itemId", contentItemHandler.UpdateContentItem)
	privateGroup.Delete("/content-types/:contentTypeId/items/:itemId", contentItemHandler.DeleteContentItem)

	// Flat forms. The param names must match what the handlers read.
	privateGroup.Get("/content-fields/:contentTypeId", contentFieldHandler.GetContentFieldsByContentType)
	privateGroup.Get("/content-fields/by-id/:fieldId", contentFieldHandler.GetContentFieldByID)
	privateGroup.Patch("/content-fields/:fieldId", contentFieldHandler.UpdateContentField)
	privateGroup.Delete("/content-fields/:fieldId", contentFieldHandler.DeleteContentField)

	privateGroup.Get("/content-items/:contentTypeId", contentItemHandler.GetContentItemsByContentType)
	privateGroup.Get("/content-items/by-id/:itemId", contentItemHandler.GetContentItemByID)
	privateGroup.Patch("/content-items/:itemId", contentItemHandler.UpdateContentItem)
	privateGroup.Delete("/content-items/:itemId", contentItemHandler.DeleteContentItem)
}
