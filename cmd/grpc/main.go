package main

import (
	"log"
	"my-go-app/internal/database"
	"my-go-app/internal/services"
	"my-go-app/proto"
	"net"
	"os"

	"github.com/clerk/clerk-sdk-go/v2"
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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal("Failed to listen on gRPC port:", err)
	}
	s := grpc.NewServer()
	proto.RegisterElementServiceServer(s, elementService)
	reflection.Register(s)
	log.Println("gRPC server listening on :" + port)
	if err := s.Serve(lis); err != nil {
		log.Fatal("Failed to serve gRPC:", err)
	}
}
