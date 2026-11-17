package routes

import (
	"my-go-app/internal/handlers"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"

	"github.com/gofiber/fiber/v3"
)

func ElementCommentRoutes(group fiber.Router, repos *repositories.RepositoriesInterface) {
	elementCommentService := services.NewElementCommentService(repos.ElementCommentRepository, repos.ElementRepository, repos.ProjectRepository)
	elementCommentHandler := handlers.NewElementCommentHandler(elementCommentService)

	group.Post("/element-comments", elementCommentHandler.CreateElementComment)
	group.Get("/element-comments/:id", elementCommentHandler.GetElementCommentByID)
	group.Get("/element-comments", elementCommentHandler.GetElementComments)
	group.Get("/elements/:elementId/comments", elementCommentHandler.GetElementComments)
	group.Get("/element-comments/author/:authorId", elementCommentHandler.GetCommentsByAuthorID)
	group.Get("/projects/:projectId/comments", elementCommentHandler.GetCommentsByProjectID)
	group.Patch("/element-comments/:id", elementCommentHandler.UpdateElementComment)
	group.Patch("/element-comments/:id/toggle-resolved", elementCommentHandler.ToggleResolvedStatus)
	group.Delete("/element-comments/:id", elementCommentHandler.DeleteElementComment)
}
