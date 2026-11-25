package routes

import (
	"my-go-app/internal/handlers"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"
	"my-go-app/pkg/middleware"

	"github.com/gofiber/fiber/v3"
)

func RegisterCustomElements(group fiber.Router, repos *repositories.Repositories) {
	customElementService := services.NewCustomElementService(repos.CustomElementRepository)
	customElementTypeService := services.NewCustomElementTypeService(repos.CustomElementTypeRepository)
	customElementHandler := handlers.NewCustomElementHandler(customElementService)
	customElementTypeHandler := handlers.NewCustomElementTypeHandler(customElementTypeService)
	adminOnly := middleware.RequireAdmin()

	group.Post("/custom-elements", customElementHandler.Create)
	group.Get("/custom-elements", customElementHandler.List)
	group.Get("/custom-elements/:id", customElementHandler.GetByID)
	group.Patch("/custom-elements/:id", customElementHandler.Update)
	group.Delete("/custom-elements/:id", customElementHandler.Delete)

	group.Post("/custom-element-types", adminOnly, customElementTypeHandler.Create)
	group.Get("/custom-element-types", customElementTypeHandler.List)
	group.Get("/custom-element-types/:id", customElementTypeHandler.GetByID)
	group.Patch("/custom-element-types/:id", adminOnly, customElementTypeHandler.Update)
	group.Delete("/custom-element-types/:id", adminOnly, customElementTypeHandler.Delete)
}
