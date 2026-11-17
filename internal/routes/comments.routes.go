package routes

import (
	"my-go-app/internal/handlers"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"

	"github.com/gofiber/fiber/v3"
)

func CommentRoutes(group fiber.Router, repos *repositories.RepositoriesInterface) {
	commentService := services.NewCommentService(repos.CommentRepository, repos.MarketplaceRepository)
	commentHandler := handlers.NewCommentHandler(commentService)

	group.Post("/comments", commentHandler.CreateComment)
	group.Get("/comments/:commentid", commentHandler.GetCommentByID)
	group.Get("/comments", commentHandler.GetComments)
	group.Patch("/comments/:commentid", commentHandler.UpdateComment)
	group.Delete("/comments/:commentid", commentHandler.DeleteComment)
}
