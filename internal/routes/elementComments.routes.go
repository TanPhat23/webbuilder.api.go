package routes

import (
	"my-go-app/internal/authz"
	"my-go-app/internal/handlers"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"

	"github.com/gofiber/fiber/v3"
)

func RegisterElementComments(group fiber.Router, repos *repositories.Repositories) {
	elementCommentService := services.NewElementCommentService(repos.ElementCommentRepository, repos.ElementRepository, repos.ProjectRepository)
	elementCommentHandler := handlers.NewElementCommentHandler(elementCommentService)
	authorizer := services.NewProjectAuthorizer(repos.ProjectRepository)

	group.Post("/element-comments", elementCommentHandler.Create)
	group.Get("/element-comments/:id", elementCommentHandler.GetByID)
	group.Get("/element-comments", elementCommentHandler.List)
	group.Get("/elements/:elementId/comments", elementCommentHandler.List)
	group.Get("/element-comments/author/:authorId", elementCommentHandler.ListByAuthor)
	group.Get("/projects/:projectId/comments", authorizer.RequireProjectPermission("projectId", authz.PermCommentView), elementCommentHandler.ListByProject)
	group.Patch("/element-comments/:id", elementCommentHandler.Update)
	group.Patch("/element-comments/:id/toggle-resolved", elementCommentHandler.ToggleResolved)
	group.Delete("/element-comments/:id", elementCommentHandler.Delete)
}
