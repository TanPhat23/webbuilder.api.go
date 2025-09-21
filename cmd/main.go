package main

import (
	"log"
	"my-go-app/internal/database"
	"my-go-app/internal/routes"
	"my-go-app/pkg/configs"
	"os"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Warning: Could not load .env file, using system environment variables")
	}
	//Set Clerk API key from environment variable
	clerk.SetKey(os.Getenv("CLERK_SECRET_KEY"))

	// Initialize database connection
	db, err := database.InitDB()
	if err != nil {
		log.Fatal("Failed to initialize database connection pool:", err)
	}

	repos := database.NewRepositories(db)

	app := fiber.New(configs.FiberConfig())
	app.Use(cors.New(configs.CorsConfig()))
	app.Use(logger.New(configs.LoggerConfig()))
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed, // Fast compression for better performance
	}))

	routes.PrivateRoutes(app, repos)
	routes.PublicRoutes(app, repos)

	app.Listen(":8080")
}
