package routes

import (
	"my-go-app/internal/handlers"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"
	"my-go-app/pkg/middleware"

	"github.com/gofiber/fiber/v2"
)

func PrivateRoutes(app *fiber.App, repos *repositories.RepositoriesInterface, cloudinaryService *services.CloudinaryService) {
	elementHandler := handlers.NewElementHandler(repos.ElementRepository)
	projectHandler := handlers.NewProjectHandler(repos.ProjectRepository)
	pageHandler := handlers.NewPageHandler(repos.PageRepository)
	snapshotHandler := handlers.NewSnapshotHandler(repos.SnapshotRepository, repos.ElementRepository)
	contentTypeHandler := handlers.NewContentTypeHandler(repos.ContentTypeRepository)
	contentFieldHandler := handlers.NewContentFieldHandler(repos.ContentFieldRepository)
	contentItemHandler := handlers.NewContentItemHandler(repos.ContentItemRepository)
	imageHandler := handlers.NewImageHandler(repos.ImageRepository, cloudinaryService)
	marketplaceHandler := handlers.NewMarketplaceHandler(repos.MarketplaceRepository)
	customElementHandler := handlers.NewCustomElementHandler(repos.CustomElementRepository)
	customElementTypeHandler := handlers.NewCustomElementTypeHandler(repos.CustomElementTypeRepository)

	group := app.Group("/api/v1", middleware.AuthenticateMiddleware)

	group.Get("/elements/:projectid", elementHandler.GetElements)

	group.Get("/projects/user", projectHandler.GetProjectByUserID)
	group.Get("/projects/:projectid", projectHandler.GetProjectByID)
	group.Get("/projects/:projectid/pages", projectHandler.GetProjectPages)
	group.Delete("/projects/:projectid", projectHandler.DeleteProject)
	group.Patch("/projects/:projectid", projectHandler.UpdateProject)
	group.Delete("/projects/:projectid/pages/:pageid", pageHandler.DeletePage)

	group.Post("/snapshots/:projectid/save", snapshotHandler.SaveSnapshot)
	group.Get("/snapshots/:projectid", snapshotHandler.GetSnapshots)
	group.Get("/snapshots/:projectid/:snapshotid", snapshotHandler.GetSnapshotByID)

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
}
