package routes

import (
	"my-go-app/internal/handlers"

	"github.com/gofiber/fiber/v2"
)

func PrivateRoutes(app *fiber.App) {
	// Define your private routes here
	// Example:
	// app.Get("/private", func(c *fiber.Ctx) error {
	// 	return c.SendString("This is a private route")
	// })

	elementHandler := handlers.NewHandler()

	group := app.Group("/api/v1")
	group.Get("/elements", elementHandler.GetElements)
}
