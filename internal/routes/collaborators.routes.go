package routes

import (
	"my-go-app/internal/authz"
	"my-go-app/internal/handlers"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"

	"github.com/gofiber/fiber/v3"
)

func RegisterCollaborators(group fiber.Router, repos *repositories.Repositories, onAccessChanged ...func(projectID, userID string)) {
	collaboratorService := services.NewCollaboratorService(repos.CollaboratorRepository, repos.ProjectRepository)
	if len(onAccessChanged) > 0 && onAccessChanged[0] != nil {
		collaboratorService.SetAccessChangedHook(onAccessChanged[0])
	}
	collaboratorHandler := handlers.NewCollaboratorHandler(collaboratorService)
	authorizer := services.NewProjectAuthorizer(repos.ProjectRepository)

	group.Post("/collaborators", collaboratorHandler.Create)
	group.Get("/collaborators/project/:projectid", authorizer.RequireProjectPermission("projectid", authz.PermCollaboratorView), collaboratorHandler.ListByProject)
	group.Post("/collaborators/project/:projectid/leave", authorizer.RequireProjectPermission("projectid", authz.PermProjectView), collaboratorHandler.LeaveProject)
	group.Get("/collaborators/:collaboratorid", collaboratorHandler.GetByID)
	group.Patch("/collaborators/:collaboratorid", collaboratorHandler.UpdateRole)
	group.Patch("/collaborators/:collaboratorid/role", collaboratorHandler.UpdateRole)
	group.Delete("/collaborators/:collaboratorid", collaboratorHandler.Delete)
}
