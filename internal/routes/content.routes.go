package routes

import (
	"my-go-app/internal/handlers"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"

	"github.com/gofiber/fiber/v3"
)

func ContentRoutes(publicGroup, privateGroup fiber.Router, repos *repositories.RepositoriesInterface) {
	contentTypeService := services.NewContentTypeService(repos.ContentTypeRepository)
	contentFieldService := services.NewContentFieldService(repos.ContentFieldRepository)
	contentItemService := services.NewContentItemService(repos.ContentItemRepository)

	contentTypeHandler := handlers.NewContentTypeHandler(contentTypeService)
	contentFieldHandler := handlers.NewContentFieldHandler(contentFieldService)
	contentItemHandler := handlers.NewContentItemHandler(contentItemService)

	publicGroup.Get("/public/content", contentItemHandler.GetPublicContentItems)
	publicGroup.Get("/public/content/:contentTypeId/:slug", contentItemHandler.GetPublicContentItemBySlug)

	privateGroup.Post("/content-types", contentTypeHandler.CreateContentType)
	privateGroup.Get("/content-types", contentTypeHandler.GetContentTypes)
	privateGroup.Get("/content-types/:id", contentTypeHandler.GetContentTypeByID)
	privateGroup.Patch("/content-types/:id", contentTypeHandler.UpdateContentType)
	privateGroup.Delete("/content-types/:id", contentTypeHandler.DeleteContentType)

	privateGroup.Post("/content-fields", contentFieldHandler.CreateContentField)
	privateGroup.Get("/content-fields/:contentTypeId", contentFieldHandler.GetContentFieldsByContentType)
	privateGroup.Get("/content-fields/by-id/:id", contentFieldHandler.GetContentFieldByID)
	privateGroup.Patch("/content-fields/:id", contentFieldHandler.UpdateContentField)
	privateGroup.Delete("/content-fields/:id", contentFieldHandler.DeleteContentField)

	privateGroup.Post("/content-items", contentItemHandler.CreateContentItem)
	privateGroup.Get("/content-items/:contentTypeId", contentItemHandler.GetContentItemsByContentType)
	privateGroup.Get("/content-items/by-id/:id", contentItemHandler.GetContentItemByID)
	privateGroup.Patch("/content-items/:id", contentItemHandler.UpdateContentItem)
	privateGroup.Delete("/content-items/:id", contentItemHandler.DeleteContentItem)
}
