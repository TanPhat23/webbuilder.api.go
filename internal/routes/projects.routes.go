package routes

import (
	"my-go-app/internal/authz"
	"my-go-app/internal/handlers"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
)

func ProjectRoutes(publicGroup, privateGroup fiber.Router, repos *repositories.RepositoriesInterface, onAccessChanged ...func(projectID, userID string)) {
	projectService := services.NewProjectService(repos.ProjectRepository, repos.CollaboratorRepository, repos.UserRepository)
	if len(onAccessChanged) > 0 && onAccessChanged[0] != nil {
		projectService.SetAccessChangedHook(onAccessChanged[0])
	}
	projectHandler := handlers.NewProjectHandler(projectService)
	authorizer := services.NewProjectAuthorizer(repos.ProjectRepository)

	publicGroup.Get("/projects/public/by-subdomain/:subdomain", projectHandler.GetPublicProjectBySubdomain)
	publicGroup.Get("/projects/public/:projectid", projectHandler.GetPublicProjectByID)

	privateGroup.Post("/projects", projectHandler.CreateProject)
	privateGroup.Get("/projects/:projectid/access", projectHandler.GetProjectAccess)
	privateGroup.Get("/projects/user", projectHandler.GetProjectsByUser)
	privateGroup.Get("/projects/:projectid", authorizer.RequireProjectPermission("projectid", authz.PermProjectView), projectHandler.GetProjectByID)
	privateGroup.Get("/projects/:projectid/pages", authorizer.RequireProjectPermission("projectid", authz.PermPageView), projectHandler.GetProjectPages)
	privateGroup.Post("/projects/:projectid/transfer-ownership", authorizer.RequireProjectPermission("projectid", authz.PermProjectTransfer), projectHandler.TransferOwnership)
	privateGroup.Delete("/projects/:projectid", authorizer.RequireProjectPermission("projectid", authz.PermProjectDelete), projectHandler.DeleteProject)
	privateGroup.Patch("/projects/:projectid", authorizer.RequireProjectPermission("projectid", authz.PermProjectView), projectHandler.UpdateProject)
}

func ProjectWSTicketRoutes(privateGroup fiber.Router, repos *repositories.RepositoriesInterface, redisClient *redis.Client) {
	authorizer := services.NewProjectAuthorizer(repos.ProjectRepository)
	wsTicketHandler := handlers.NewWSTicketHandler(redisClient)
	privateGroup.Post("/projects/:projectid/ws-ticket", authorizer.RequireProjectPermission("projectid", authz.PermProjectView), wsTicketHandler.MintWSTicket)
}
