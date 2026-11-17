package main

import (
	"context"
	"log"
	"my-go-app/internal/bootstrap"
	"my-go-app/internal/realtime"
	"my-go-app/internal/routes"
	"my-go-app/internal/services"
	"my-go-app/pkg/configs"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

func main() {
	repos := bootstrap.InitRepositories()

	// Initialize Email service
	emailService := services.NewEmailService()

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:3000" // Default for development
	}
	invitationService := services.NewInvitationService(repos.InvitationRepository, repos.CollaboratorRepository, repos.ProjectRepository, repos.UserRepository, emailService, baseURL)

	// Initialize Cloudinary service
	cloudinaryService, err := services.NewCloudinaryService()
	if err != nil {
		log.Println("Warning: Cloudinary service not initialized:", err)
		log.Println("Image upload functionality will not be available")
	}

	configs.SetupLogging()

	app := fiber.New(configs.FiberConfig())
	app.Use(requestid.New())
	app.Use(cors.New(configs.CorsConfig()))
	app.Use(logger.New(configs.LoggerConfig()))
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed, // Fast compression for better performance
	}))

	app.Get("/health", func(c *fiber.Ctx) error { return c.SendString("OK") })

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Collaborative editing (/ws/:project) runs in this same process. If it
	// cannot start (e.g. Redis is down) the REST API still comes up.
	if shutdownRealtime, err := realtime.Mount(ctx, app); err != nil {
		log.Println("ERROR: realtime WebSocket disabled:", err)
	} else {
		defer shutdownRealtime()
	}

	routes.PublicRoutes(app, repos)
	routes.PrivateRoutes(app, repos, cloudinaryService, invitationService)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	go func() {
		<-ctx.Done()
		_ = app.ShutdownWithContext(context.Background())
	}()
	if err := app.Listen(":" + port); err != nil {
		log.Fatal("Server failed:", err)
	}
}
