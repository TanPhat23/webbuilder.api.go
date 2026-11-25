package main

import (
	"context"
	"log"
	"my-go-app/internal/database"
	"my-go-app/internal/grpcserver"
	"my-go-app/internal/routes"
	"my-go-app/internal/services"
	"my-go-app/pkg/configs"
	"my-go-app/pkg/middleware"
	"my-go-app/proto"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
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
	grpcCfg, err := grpcserver.LoadConfigFromEnv("8081")
	if err != nil {
		log.Fatal("Invalid gRPC configuration:", err)
	}
	grpcSrv, err := grpcserver.NewServer(grpcCfg)
	if err != nil {
		log.Fatal("Failed to initialize gRPC server:", err)
	}

	//Set Clerk API key from environment variable
	clerk.SetKey(os.Getenv("CLERK_SECRET_KEY"))

	// Initialize database connection
	db, err := database.Connect()
	if err != nil {
		log.Fatal("Failed to initialize database connection pool:", err)
	}

	repos := database.NewRepositories(db)

	elementService := services.NewElementService(repos.SnapshotRepository, repos.ElementRepository, repos.EventWorkflowRepository, repos.ElementEventWorkflowRepository)

	// Initialize Email service
	emailService := services.NewEmailService()

	// Initialize Invitation service
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:3000" // Default for development
	}
	invitationService := services.NewInvitationService(repos.InvitationRepository, repos.CollaboratorRepository, repos.ProjectRepository, repos.UserRepository, emailService, baseURL)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())

	// Start gRPC server
	go func() {
		addr := grpcCfg.Address()
		lis, err := net.Listen("tcp", addr)
		if err != nil {
			log.Fatal("Failed to listen on gRPC port:", err)
		}
		proto.RegisterElementServiceServer(grpcSrv, elementService)
		log.Println("gRPC server listening on " + addr)

		// Goroutine to handle shutdown
		go func() {
			<-ctx.Done()
			log.Println("Shutting down gRPC server...")
			grpcSrv.GracefulStop()
		}()

		if err := grpcSrv.Serve(lis); err != nil {
			log.Fatal("Failed to serve gRPC:", err)
		}
	}()

	// Initialize object storage (Railway Bucket / MinIO S3)
	storageService, err := services.NewS3StorageService()
	if err != nil {
		log.Println("Warning: Object storage service not initialized:", err)
		log.Println("Image upload functionality will not be available")
	}

	app := fiber.New(configs.Fiber())
	app.Use(cors.New(configs.CORS()))
	app.Use(logger.New(configs.Logger()))
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed, // Fast compression for better performance
	}))

	routes.Register(app, repos, storageService, invitationService, nil)

	// Handle graceful shutdown
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		log.Println("Received signal, shutting down...")
		cancel()
		app.Shutdown()
	}()

	app.Listen(":8080", fiber.ListenConfig{EnablePrefork: false})
}
