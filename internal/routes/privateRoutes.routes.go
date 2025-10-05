package routes

import (
	"my-go-app/internal/handlers"
	"my-go-app/internal/repositories"
	"my-go-app/pkg/middleware"

	// "my-go-app/pkg/middleware"

	"github.com/gofiber/fiber/v2"
)

func PrivateRoutes(app *fiber.App, repos *repositories.RepositoriesInterface) {
	elementHandler := handlers.NewElementHandler(repos.ElementRepository)
	projectHandler := handlers.NewProjectHandler(repos.ProjectRepository)
	pageHandler := handlers.NewPageHandler(repos.PageRepository)
	snapshotHandler := handlers.NewSnapshotHandler(repos.SnapshotRepository, repos.ElementRepository)
	contentTypeHandler := handlers.NewContentTypeHandler(repos.ContentTypeRepository)
	contentFieldHandler := handlers.NewContentFieldHandler(repos.ContentFieldRepository)
	contentItemHandler := handlers.NewContentItemHandler(repos.ContentItemRepository)

	group := app.Group("/api/v1", middleware.AuthenticateMiddleware)

	group.Get("/elements/:projectid", elementHandler.GetElements)

	group.Get("/projects/user", projectHandler.GetProjectByUserID)
	group.Get("/projects/:projectid", projectHandler.GetProjectByID)
	group.Get("/projects/:projectid/pages", projectHandler.GetProjectPages)
	group.Delete("/projects/:projectid", projectHandler.DeleteProject)
	group.Patch("/projects/:projectid", projectHandler.UpdateProject)
	group.Delete("/projects/:projectid/pages/:pageid", pageHandler.DeletePage)

	group.Post("/snapshots/:projectid/save", snapshotHandler.SaveSnapshot)

	// CMS routes
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

}
