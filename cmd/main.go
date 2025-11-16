package main

import (
	"context"
	"log"
	"my-go-app/internal/database"
	"my-go-app/internal/routes"
	"my-go-app/internal/services"
	"my-go-app/pkg/configs"
	"my-go-app/proto"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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
		lis, err := net.Listen("tcp", ":8081")
		if err != nil {
			log.Fatal("Failed to listen on gRPC port:", err)
		}
		s := grpc.NewServer()
		proto.RegisterElementServiceServer(s, elementService)
		reflection.Register(s)
		log.Println("gRPC server listening on :8081")

		// Goroutine to handle shutdown
		go func() {
			<-ctx.Done()
			log.Println("Shutting down gRPC server...")
			s.GracefulStop()
		}()

		if err := s.Serve(lis); err != nil {
			log.Fatal("Failed to serve gRPC:", err)
		}
	}()

	// Initialize Cloudinary service
	cloudinaryService, err := services.NewCloudinaryService()
	if err != nil {
		log.Println("Warning: Cloudinary service not initialized:", err)
		log.Println("Image upload functionality will not be available")
	}

	app := fiber.New(configs.FiberConfig())
	app.Use(cors.New(configs.CorsConfig()))
	app.Use(logger.New(configs.LoggerConfig()))
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed, // Fast compression for better performance
	}))

	routes.PublicRoutes(app, repos)
	routes.PrivateRoutes(app, repos, cloudinaryService, invitationService)

	// Handle graceful shutdown
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		log.Println("Received signal, shutting down...")
		cancel()
		app.Shutdown()
	}()

	app.Listen(":8080")
}
