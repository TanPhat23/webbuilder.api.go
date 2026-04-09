package routes

import (
	"my-go-app/internal/handlers"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"

	"github.com/gofiber/fiber/v3"
)

func ElementRoutes(publicGroup, privateGroup fiber.Router, repos *repositories.RepositoriesInterface) {
	elementService := services.NewElementWrapperService(repos.ElementRepository)
	elementHandler := handlers.NewElementHandler(elementService)

	publicGroup.Get("/elements/public/:projectid", elementHandler.GetElements)
	publicGroup.Get("/elements/public/by-pages", elementHandler.GetElementsByPageIds)

	privateGroup.Get("/elements/:projectid", elementHandler.GetElements)
	privateGroup.Get("/elements/by-pages", elementHandler.GetElementsByPageIds)
}
