package routes

import (
	"my-go-app/internal/handlers"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"

	"github.com/gofiber/fiber/v3"
)

func SnapshotRoutes(group fiber.Router, repos *repositories.RepositoriesInterface) {
	snapshotService := services.NewSnapshotService(repos.SnapshotRepository, repos.ElementRepository, repos.ProjectRepository)
	snapshotHandler := handlers.NewSnapshotHandler(snapshotService, repos.ElementRepository)

	group.Post("/snapshots/:projectid/save", snapshotHandler.SaveSnapshot)
	group.Get("/snapshots/:projectid", snapshotHandler.GetSnapshots)
	group.Get("/snapshots/:snapshotid", snapshotHandler.GetSnapshotByID)
	group.Delete("/snapshots/:snapshotid", snapshotHandler.DeleteSnapshot)
}
