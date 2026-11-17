package main

import (
	"log"
	"my-go-app/internal/bootstrap"
	"my-go-app/internal/services"
	"my-go-app/proto"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	repos := bootstrap.InitRepositories()

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
