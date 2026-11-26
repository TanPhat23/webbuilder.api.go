package routes

import (
	"my-go-app/internal/authz"
	"my-go-app/internal/handlers"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
)

func RegisterPreviewTickets(publicGroup, privateGroup fiber.Router, repos *repositories.Repositories, redisClient *redis.Client) {
	projectService := services.NewProjectService(repos.ProjectRepository, repos.CollaboratorRepository, repos.UserRepository)
	elementService := services.NewElementQueryService(repos.ElementRepository)
	authorizer := services.NewProjectAuthorizer(repos.ProjectRepository)

	previewHandler := handlers.NewPreviewTicketHandler(
		redisClient,
		projectService,
		elementService,
		repos.ContentTypeRepository,
		repos.ContentItemRepository,
	)

	privateGroup.Post(
		"/projects/:projectid/preview-ticket",
		authorizer.RequireProjectPermission("projectid", authz.PermProjectView),
		previewHandler.Mint,
	)

	publicGroup.Post("/preview/exchange", previewHandler.Exchange)

	previewRead := publicGroup.Group("/preview/:projectid", previewHandler.RequirePreviewToken)
	previewRead.Get("/project", previewHandler.GetProject)
	previewRead.Get("/pages", previewHandler.ListPages)
	previewRead.Get("/elements", previewHandler.ListElements)
	previewRead.Get("/cms/:contentTypeId", previewHandler.ListCMSItems)
	previewRead.Get("/cms/:contentTypeId/item/*", previewHandler.GetCMSItemBySlug)
}
