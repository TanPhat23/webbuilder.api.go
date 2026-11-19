package routes

import (
	"my-go-app/internal/handlers"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
)

func PageRoutes(publicGroup, privateGroup fiber.Router, repos *repositories.RepositoriesInterface, redisClient *redis.Client) {
	pageService := services.NewPageService(repos.PageRepository, repos.ProjectRepository)
	pageHandler := handlers.NewPageHandler(pageService)

	publicGroup.Get("/pages/public/:projectid", cachePublicPreview(redisClient, repos.ProjectRepository, "route-pages"), pageHandler.GetPublicPagesByProjectID)
	publicGroup.Get("/pages/public/:projectid/:pageid", pageHandler.GetPublicPageByID)

	privateGroup.Get("/pages/:projectid", pageHandler.GetPagesByProjectID)
	privateGroup.Get("/pages/:projectid/:pageid", pageHandler.GetPageByID)
	privateGroup.Post("/pages/:projectid", pageHandler.CreatePage)
	privateGroup.Patch("/pages/:projectid/:pageid", pageHandler.UpdatePage)
	privateGroup.Delete("/pages/:projectid/:pageid", pageHandler.DeletePage)
}
