package routes

import (
	"my-go-app/internal/handlers"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"

	"github.com/gofiber/fiber/v3"
)

func ProjectRoutes(publicGroup, privateGroup fiber.Router, repos *repositories.RepositoriesInterface) {
	projectService := services.NewProjectService(repos.ProjectRepository, repos.CollaboratorRepository, repos.UserRepository)
	projectHandler := handlers.NewProjectHandler(projectService)

	publicGroup.Get("/projects/public/:projectid", projectHandler.GetPublicProjectByID)

	privateGroup.Get("/projects/user", projectHandler.GetProjectsByUser)
	privateGroup.Get("/projects/:projectid", projectHandler.GetProjectByID)
	privateGroup.Get("/projects/:projectid/pages", projectHandler.GetProjectPages)
	privateGroup.Delete("/projects/:projectid", projectHandler.DeleteProject)
	privateGroup.Patch("/projects/:projectid", projectHandler.UpdateProject)
}
