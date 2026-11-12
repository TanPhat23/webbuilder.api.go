package routes

import (
	"my-go-app/internal/handlers"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"

	"github.com/gofiber/fiber/v3"
)

func CollaboratorRoutes(group fiber.Router, repos *repositories.RepositoriesInterface) {
	collaboratorService := services.NewCollaboratorService(repos.CollaboratorRepository, repos.ProjectRepository)
	collaboratorHandler := handlers.NewCollaboratorHandler(collaboratorService)

	group.Post("/collaborators", collaboratorHandler.CreateCollaborator)
	group.Get("/collaborators/:projectid", collaboratorHandler.GetCollaboratorsByProject)
	group.Get("/collaborators/project/:projectid", collaboratorHandler.GetCollaboratorsByProject)
	group.Get("/collaborators/:collaboratorid", collaboratorHandler.GetCollaboratorByID)
	group.Patch("/collaborators/:id", collaboratorHandler.UpdateCollaboratorRole)
	group.Patch("/collaborators/:collaboratorid/role", collaboratorHandler.UpdateCollaboratorRole)
	group.Delete("/collaborators/:id", collaboratorHandler.DeleteCollaborator)
	group.Delete("/collaborators/:collaboratorid", collaboratorHandler.DeleteCollaborator)
}
