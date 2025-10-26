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
	contentItemHandler := handlers.NewContentItemHandler(repos.ContentItemRepository)
	// marketplaceHandler := handlers.NewMarketplaceHandler(repos.MarketplaceRepository)
	group := app.Group("/api/v1")
	group.Get("/elements/public/:projectid", elementHandler.GetElements)
	group.Get("/projects/public", projectHandler.GetProject)
	group.Get("/projects/public/:projectid", projectHandler.GetPublicProjectByID)
	group.Get("/public/content", contentItemHandler.GetPublicContentItems)
	group.Get("/public/content/:contentTypeId/:slug", contentItemHandler.GetPublicContentItemBySlug)

	// Public marketplace routes
	// group.Get("/marketplace/categories", marketplaceHandler.GetCategories)
	// group.Get("/marketplace/tags", marketplaceHandler.GetTags)
	// group.Get("/marketplace/items", marketplaceHandler.GetMarketplaceItems)
	// group.Get("/marketplace/items/:itemid", marketplaceHandler.GetMarketplaceItemByID)
}
