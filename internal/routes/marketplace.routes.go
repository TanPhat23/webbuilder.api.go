package routes

import (
	"my-go-app/internal/handlers"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"

	"github.com/gofiber/fiber/v3"
)

func MarketplaceRoutes(group fiber.Router, repos *repositories.RepositoriesInterface) {
	marketplaceService := services.NewMarketplaceService(repos.MarketplaceRepository)
	marketplaceHandler := handlers.NewMarketplaceHandler(marketplaceService)

	group.Post("/marketplace", marketplaceHandler.CreateMarketplaceItem)
	group.Get("/marketplace", marketplaceHandler.GetMarketplaceItems)
	group.Get("/marketplace/:id", marketplaceHandler.GetMarketplaceItemByID)
	group.Patch("/marketplace/:id", marketplaceHandler.UpdateMarketplaceItem)
	group.Delete("/marketplace/:id", marketplaceHandler.DeleteMarketplaceItem)
}
