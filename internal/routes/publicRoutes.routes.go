package routes

import (
	"my-go-app/internal/handlers"
	"my-go-app/internal/repositories"

	"github.com/gofiber/fiber/v2"
)

func PublicRoutes(app *fiber.App, repos *repositories.RepositoriesInterface) {
	// Define your public routes here
	// Example:
	// app.Get("/public", func(c *fiber.Ctx) error {
	// 	return c.SendString("This is a public route")
	// })

	elementHandler := handlers.NewElementHandler(repos.ElementRepository)
	projectHandler := handlers.NewProjectHandler(repos.ProjectRepository)
	pageHandler := handlers.NewPageHandler(repos.PageRepository)
	contentItemHandler := handlers.NewContentItemHandler(repos.ContentItemRepository)
	// marketplaceHandler := handlers.NewMarketplaceHandler(repos.MarketplaceRepository)
	group := app.Group("/api/v1")

	// Public element routes
	group.Get("/elements/public/:projectid", elementHandler.GetElements)
	group.Get("/elements/public/by-pages", elementHandler.GetElementsByPageIds)

	// Public project routes
	group.Get("/projects/public/:projectid", projectHandler.GetPublicProjectByID)

	// Public page routes
	group.Get("/pages/public/:projectid", pageHandler.GetPagesByProjectID)
	group.Get("/pages/public/:projectid/:pageid", pageHandler.GetPageByID)

	// Public content routes
	group.Get("/public/content", contentItemHandler.GetPublicContentItems)
	group.Get("/public/content/:contentTypeId/:slug", contentItemHandler.GetPublicContentItemBySlug)

	// Public marketplace routes
	// group.Get("/marketplace/categories", marketplaceHandler.GetCategories)
	// group.Get("/marketplace/tags", marketplaceHandler.GetTags)
	// group.Get("/marketplace/items", marketplaceHandler.GetMarketplaceItems)
	// group.Get("/marketplace/items/:itemid", marketplaceHandler.GetMarketplaceItemByID)
}
