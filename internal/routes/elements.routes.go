package routes

import (
	"my-go-app/internal/handlers"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
)

func RegisterElements(publicGroup, privateGroup fiber.Router, repos *repositories.Repositories, redisClient *redis.Client) {
	elementService := services.NewElementQueryService(repos.ElementRepository)
	elementHandler := handlers.NewElementHandler(elementService, repos.ProjectRepository)

	// Literal paths first: otherwise :projectid matches "by-pages" and the handler never runs.
	publicGroup.Get("/elements/public/by-pages", elementHandler.ListPublicByPageIDs)
	publicGroup.Get("/elements/public/:projectid", cachePublicPreview(redisClient, repos.ProjectRepository, "elements"), elementHandler.ListPublic)

	privateGroup.Get("/elements/by-pages", elementHandler.ListByPageIDs)
	privateGroup.Get("/elements/:projectid", elementHandler.List)
}
