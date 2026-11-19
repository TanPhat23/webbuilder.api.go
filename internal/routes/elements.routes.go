package routes

import (
	"my-go-app/internal/handlers"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
)

func ElementRoutes(publicGroup, privateGroup fiber.Router, repos *repositories.RepositoriesInterface, redisClient *redis.Client) {
	elementService := services.NewElementWrapperService(repos.ElementRepository)
	elementHandler := handlers.NewElementHandler(elementService)

	// Literal paths first: otherwise :projectid matches "by-pages" and the handler never runs.
	publicGroup.Get("/elements/public/by-pages", elementHandler.GetElementsByPageIds)
	publicGroup.Get("/elements/public/:projectid", cachePublicPreview(redisClient, repos.ProjectRepository, "elements"), elementHandler.GetElements)

	privateGroup.Get("/elements/by-pages", elementHandler.GetElementsByPageIds)
	privateGroup.Get("/elements/:projectid", elementHandler.GetElements)
}
