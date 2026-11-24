package routes

import (
	"my-go-app/internal/authz"
	"my-go-app/internal/handlers"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"

	"github.com/gofiber/fiber/v3"
)

func CollaboratorRoutes(group fiber.Router, repos *repositories.RepositoriesInterface, onAccessChanged ...func(projectID, userID string)) {
	collaboratorService := services.NewCollaboratorService(repos.CollaboratorRepository, repos.ProjectRepository)
	if len(onAccessChanged) > 0 && onAccessChanged[0] != nil {
		collaboratorService.SetAccessChangedHook(onAccessChanged[0])
	}
	collaboratorHandler := handlers.NewCollaboratorHandler(collaboratorService)
	authorizer := services.NewProjectAuthorizer(repos.ProjectRepository)

	group.Post("/collaborators", collaboratorHandler.CreateCollaborator)
	group.Get("/collaborators/project/:projectid", authorizer.RequireProjectPermission("projectid", authz.PermCollaboratorView), collaboratorHandler.GetCollaboratorsByProject)
	group.Post("/collaborators/project/:projectid/leave", authorizer.RequireProjectPermission("projectid", authz.PermProjectView), collaboratorHandler.LeaveProject)
	group.Get("/collaborators/:collaboratorid", collaboratorHandler.GetCollaboratorByID)
	group.Patch("/collaborators/:collaboratorid", collaboratorHandler.UpdateCollaboratorRole)
	group.Patch("/collaborators/:collaboratorid/role", collaboratorHandler.UpdateCollaboratorRole)
	group.Delete("/collaborators/:collaboratorid", collaboratorHandler.DeleteCollaborator)
}
