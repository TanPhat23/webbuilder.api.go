package routes

import (
	"my-go-app/internal/authz"
	"my-go-app/internal/handlers"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"

	"github.com/gofiber/fiber/v3"
)

func SnapshotRoutes(group fiber.Router, repos *repositories.RepositoriesInterface) {
	snapshotService := services.NewSnapshotService(repos.SnapshotRepository, repos.ElementRepository, repos.ProjectRepository)
	snapshotHandler := handlers.NewSnapshotHandler(snapshotService, repos.ElementRepository)
	authorizer := services.NewProjectAuthorizer(repos.ProjectRepository)

	group.Post("/snapshots/:projectid/save", authorizer.RequireProjectPermission("projectid", authz.PermSnapshotCreate), snapshotHandler.SaveSnapshot)
	group.Get("/snapshots/:projectid", authorizer.RequireProjectPermission("projectid", authz.PermSnapshotView), snapshotHandler.GetSnapshots)
	// The project id is part of the path: it was shadowed by /snapshots/:projectid
	// and DeleteSnapshot reads :projectid as well as :snapshotid.
	group.Get("/snapshots/:projectid/:snapshotid", authorizer.RequireProjectPermission("projectid", authz.PermSnapshotView), snapshotHandler.GetSnapshotByID)
	group.Delete("/snapshots/:projectid/:snapshotid", authorizer.RequireProjectPermission("projectid", authz.PermSnapshotDelete), snapshotHandler.DeleteSnapshot)
}
