package routes

import (
	"my-go-app/internal/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

func PublicRoutes(app *fiber.App) {
	// Define your public routes here
	// Example:
	// app.Get("/public", func(c *fiber.Ctx) error {
	// 	return c.SendString("This is a public route")
	// })

	elmentHandler := handlers.NewElementHandler()
	projectHandler := handlers.NewProjectHandler()
	group := app.Group("/api/v1")
	group.Use(limiter.New())
	group.Get("/elements/public/:projectid", elmentHandler.GetElements)

	group.Get("/projects/public", projectHandler.GetProject)
	
}
