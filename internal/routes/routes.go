package routes

import (
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"
	"my-go-app/pkg/middleware"

	"github.com/gofiber/fiber/v3"
)

func SetupRoutes(app *fiber.App, repos *repositories.RepositoriesInterface, cloudinaryService *services.CloudinaryService, invitationService *services.InvitationService) {
	publicGroup := app.Group("/api/v1")
	privateGroup := app.Group("/api/v1", middleware.AuthenticateMiddleware)

	ElementRoutes(publicGroup, privateGroup, repos)
	PageRoutes(publicGroup, privateGroup, repos)
	ProjectRoutes(publicGroup, privateGroup, repos)
	ContentRoutes(publicGroup, privateGroup, repos)

	SnapshotRoutes(privateGroup, repos)
	ImageRoutes(privateGroup, repos, cloudinaryService)
	MarketplaceRoutes(privateGroup, repos)
	CustomElementRoutes(privateGroup, repos)
	InvitationRoutes(privateGroup, invitationService)
	CollaboratorRoutes(privateGroup, repos)
	CommentRoutes(privateGroup, repos)
	ElementCommentRoutes(privateGroup, repos)
	UserRoutes(privateGroup, repos)
	WorkflowRoutes(privateGroup, repos)
}
