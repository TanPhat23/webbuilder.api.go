package routes

import (
	"my-go-app/internal/handlers"
	"my-go-app/pkg/middleware"

	// "my-go-app/pkg/middleware"

	"github.com/gofiber/fiber/v2"
)

func PrivateRoutes(app *fiber.App) {
	// Define your private routes here
	// Example:
	// app.Get("/private", func(c *fiber.Ctx) error {
	// 	return c.SendString("This is a private route")
	// })

	elementHandler := handlers.NewElementHandler()
	projectHandler := handlers.NewProjectHandler()
	group := app.Group("/api/v1")

	group.Get("/elements/:projectid", middleware.AuthenticateMiddleware, elementHandler.GetElements)

	group.Get("/projects/user", middleware.AuthenticateMiddleware, projectHandler.GetProjectByUserID)
	group.Get("/projects/project/:projectid", middleware.AuthenticateMiddleware, projectHandler.GetProjectByID)

}
