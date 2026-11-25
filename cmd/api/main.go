package main

import (
	"context"
	"log"
	"my-go-app/internal/database"
	"my-go-app/internal/realtime"
	"my-go-app/internal/routes"
	"my-go-app/internal/services"
	"my-go-app/pkg/configs"
	"my-go-app/pkg/middleware"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Warning: Could not load .env file, using system environment variables")
	}
	if err := middleware.ValidateAuthEnv(); err != nil {
		log.Fatal(err)
	}
	if _, err := configs.AllowedOrigins(); err != nil {
		log.Fatal(err)
	}
	if err := services.ValidateWebhookEnv(); err != nil {
		log.Fatal(err)
	}
	//Set Clerk API key from environment variable
	clerk.SetKey(os.Getenv("CLERK_SECRET_KEY"))

	// Initialize database connection
	db, err := database.Connect()
	if err != nil {
		log.Fatal("Failed to initialize database connection pool:", err)
	}

	repos := database.NewRepositories(db)

	// Initialize Email service
	emailService := services.NewEmailService()

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:3000" // Default for development
	}
	invitationService := services.NewInvitationService(repos.InvitationRepository, repos.CollaboratorRepository, repos.ProjectRepository, repos.UserRepository, emailService, baseURL)

	// Initialize object storage (Railway Bucket / MinIO S3)
	storageService, err := services.NewS3StorageService()
	if err != nil {
		log.Println("Warning: Object storage service not initialized:", err)
		log.Println("Image upload functionality will not be available")
	}

	configs.InitLogging()

	app := fiber.New(configs.Fiber())
	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(helmet.New())
	app.Use(cors.New(configs.CORS()))
	app.Use(limiter.New(limiter.Config{
		Max:        300,
		Expiration: time.Minute,
		Next: func(c fiber.Ctx) bool {
			return strings.HasPrefix(c.Path(), "/ws")
		},
	}))
	app.Use(logger.New(configs.Logger()))
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed, // Fast compression for better performance
	}))

	app.Get("/health", func(c fiber.Ctx) error { return c.SendString("OK") })

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Collaborative editing (/ws/:project) runs in this same process. If it
	// cannot start (e.g. Redis is down) the REST API still comes up.
	redisClient, shutdownRealtime, err := realtime.Mount(ctx, app)
	if err != nil {
		log.Println("ERROR: realtime WebSocket disabled:", err)
	} else {
		defer shutdownRealtime()
	}

	routes.Register(app, repos, storageService, invitationService, redisClient)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	go func() {
		<-ctx.Done()
		_ = app.ShutdownWithContext(context.Background())
	}()
	if err := app.Listen(":"+port, fiber.ListenConfig{EnablePrefork: false}); err != nil {
		log.Fatal("Server failed:", err)
	}
}
