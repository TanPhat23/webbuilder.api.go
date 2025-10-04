package routes

import (
	"my-go-app/internal/handlers"
	"my-go-app/internal/repositories"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
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
	group := app.Group("/api/v1")
	group.Use(limiter.New())
	group.Get("/elements/public/:projectid", elementHandler.GetElements)
	group.Post("/elements/public/:projectid", elementHandler.CreateElements)
	group.Get("/projects/public", projectHandler.GetProject)
	group.Get("/public/content", contentItemHandler.GetPublicContentItems)
	group.Get("/public/content/:contentTypeId/:slug", contentItemHandler.GetPublicContentItemBySlug)

}
