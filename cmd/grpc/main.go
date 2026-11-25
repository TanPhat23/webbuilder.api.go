package main

import (
	"log"
	"my-go-app/internal/database"
	"my-go-app/internal/grpcserver"
	"my-go-app/internal/services"
	"my-go-app/proto"
	"net"
	"os"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Warning: Could not load .env file, using system environment variables")
	}

	grpcCfg, err := grpcserver.LoadConfigFromEnv("8081")
	if err != nil {
		log.Fatal("Invalid gRPC configuration:", err)
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

	s, err := grpcserver.NewServer(grpcCfg)
	if err != nil {
		log.Fatal("Failed to initialize gRPC server:", err)
	}
	proto.RegisterElementServiceServer(s, elementService)

	addr := grpcCfg.Address()
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("Failed to listen on gRPC port:", err)
	}
	log.Println("gRPC server listening on " + addr)
	if err := s.Serve(lis); err != nil {
		log.Fatal("Failed to serve gRPC:", err)
	}
}
