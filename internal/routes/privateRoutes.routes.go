package routes

import (
	"my-go-app/internal/handlers"
	"my-go-app/pkg/middleware"

	// "my-go-app/pkg/middleware"

	"github.com/gofiber/fiber/v2"
)

func PrivateRoutes(app *fiber.App) {
	elementHandler := handlers.NewElementHandler()
	projectHandler := handlers.NewProjectHandler()

	group := app.Group("/api/v1")

	group.Get("/elements/:projectid", middleware.AuthenticateMiddleware, elementHandler.GetElements)
	group.Get("/projects/user", middleware.AuthenticateMiddleware, projectHandler.GetProjectByUserID)
	group.Get("/projects/:projectid", middleware.AuthenticateMiddleware, projectHandler.GetProjectByID)
	group.Get("/projects/:projectid/pages", middleware.AuthenticateMiddleware, projectHandler.GetProjectPages)
}
