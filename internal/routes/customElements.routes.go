package routes

import (
	"my-go-app/internal/handlers"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"

	"github.com/gofiber/fiber/v3"
)

func CustomElementRoutes(group fiber.Router, repos *repositories.RepositoriesInterface) {
	customElementService := services.NewCustomElementService(repos.CustomElementRepository)
	customElementTypeService := services.NewCustomElementTypeService(repos.CustomElementTypeRepository)
	customElementHandler := handlers.NewCustomElementHandler(customElementService)
	customElementTypeHandler := handlers.NewCustomElementTypeHandler(customElementTypeService)

	group.Post("/custom-elements", customElementHandler.CreateCustomElement)
	group.Get("/custom-elements", customElementHandler.GetCustomElements)
	group.Get("/custom-elements/:id", customElementHandler.GetCustomElementByID)
	group.Patch("/custom-elements/:id", customElementHandler.UpdateCustomElement)
	group.Delete("/custom-elements/:id", customElementHandler.DeleteCustomElement)

	group.Post("/custom-element-types", customElementTypeHandler.CreateCustomElementType)
	group.Get("/custom-element-types", customElementTypeHandler.GetCustomElementTypes)
	group.Get("/custom-element-types/:id", customElementTypeHandler.GetCustomElementTypeByID)
	group.Patch("/custom-element-types/:id", customElementTypeHandler.UpdateCustomElementType)
	group.Delete("/custom-element-types/:id", customElementTypeHandler.DeleteCustomElementType)
}
