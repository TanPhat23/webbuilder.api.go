package routes

import (
	"my-go-app/internal/handlers"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"

	"github.com/gofiber/fiber/v2"
)

func PublicRoutes(app *fiber.App, repos *repositories.RepositoriesInterface) {
	elementService := services.NewElementWrapperService(repos.ElementRepository)
	projectService := services.NewProjectService(repos.ProjectRepository, repos.CollaboratorRepository, repos.UserRepository)
	pageService := services.NewPageService(repos.PageRepository, repos.ProjectRepository)
	contentItemService := services.NewContentItemService(repos.ContentItemRepository)

	elementHandler := handlers.NewElementHandler(elementService)
	projectHandler := handlers.NewProjectHandler(projectService)
	pageHandler := handlers.NewPageHandler(pageService)
	contentItemHandler := handlers.NewContentItemHandler(contentItemService)

	group := app.Group("/api/v1")

	group.Get("/elements/public/:projectid", elementHandler.GetElements)
	group.Get("/elements/public/by-pages", elementHandler.GetElementsByPageIds)

	group.Get("/projects/public/:projectid", projectHandler.GetPublicProjectByID)

	group.Get("/pages/public/:projectid", pageHandler.GetPagesByProjectID)
	group.Get("/pages/public/:projectid/:pageid", pageHandler.GetPageByID)

	group.Get("/public/content", contentItemHandler.GetPublicContentItems)
	group.Get("/public/content/:contentTypeId/:slug", contentItemHandler.GetPublicContentItemBySlug)
}