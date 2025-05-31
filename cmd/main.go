package main

import (
	"log"
	"my-go-app/pkg/configs"
	"my-go-app/pkg/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	app := fiber.New(configs.FiberConfig())
	app.Use(cors.New(configs.CorsConfig()))

	routes.PrivateRoutes(app)

	app.Listen(":8080")
}
