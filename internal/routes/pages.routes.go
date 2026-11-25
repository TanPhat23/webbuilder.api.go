package routes

import (
	"my-go-app/internal/authz"
	"my-go-app/internal/handlers"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
)

func RegisterPages(publicGroup, privateGroup fiber.Router, repos *repositories.Repositories, redisClient *redis.Client) {
	pageService := services.NewPageService(repos.PageRepository, repos.ProjectRepository)
	pageHandler := handlers.NewPageHandler(pageService)
	authorizer := services.NewProjectAuthorizer(repos.ProjectRepository)

	publicGroup.Get("/pages/public/:projectid", cachePublicPreview(redisClient, repos.ProjectRepository, "route-pages"), pageHandler.ListPublic)
	publicGroup.Get("/pages/public/:projectid/:pageid", pageHandler.GetPublicByID)

	privateGroup.Get("/pages/:projectid", pageHandler.ListByProject)
	privateGroup.Get("/pages/:projectid/:pageid", authorizer.RequireProjectPermission("projectid", authz.PermPageView), pageHandler.GetByID)
	privateGroup.Post("/pages/:projectid", authorizer.RequireProjectPermission("projectid", authz.PermPageCreate), pageHandler.Create)
	privateGroup.Patch("/pages/:projectid/:pageid", authorizer.RequireProjectPermission("projectid", authz.PermPageEdit), pageHandler.Update)
	privateGroup.Delete("/pages/:projectid/:pageid", authorizer.RequireProjectPermission("projectid", authz.PermPageDelete), pageHandler.Delete)
}
