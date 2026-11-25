package routes

import (
	"my-go-app/internal/authz"
	"my-go-app/internal/handlers"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
)

func RegisterProjects(publicGroup, privateGroup fiber.Router, repos *repositories.Repositories, onAccessChanged ...func(projectID, userID string)) {
	projectService := services.NewProjectService(repos.ProjectRepository, repos.CollaboratorRepository, repos.UserRepository)
	if len(onAccessChanged) > 0 && onAccessChanged[0] != nil {
		projectService.SetAccessChangedHook(onAccessChanged[0])
	}
	projectHandler := handlers.NewProjectHandler(projectService)
	authorizer := services.NewProjectAuthorizer(repos.ProjectRepository)

	publicGroup.Get("/projects/public/by-subdomain/:subdomain", projectHandler.GetPublicBySubdomain)
	publicGroup.Get("/projects/public/:projectid", projectHandler.GetPublicByID)

	privateGroup.Post("/projects", projectHandler.Create)
	privateGroup.Get("/projects/:projectid/access", projectHandler.GetAccess)
	privateGroup.Get("/projects/user", projectHandler.ListByUser)
	privateGroup.Get("/projects/:projectid", authorizer.RequireProjectPermission("projectid", authz.PermProjectView), projectHandler.GetByID)
	privateGroup.Get("/projects/:projectid/pages", authorizer.RequireProjectPermission("projectid", authz.PermPageView), projectHandler.ListPages)
	privateGroup.Post("/projects/:projectid/transfer-ownership", authorizer.RequireProjectPermission("projectid", authz.PermProjectTransfer), projectHandler.TransferOwnership)
	privateGroup.Delete("/projects/:projectid", authorizer.RequireProjectPermission("projectid", authz.PermProjectDelete), projectHandler.Delete)
	privateGroup.Patch("/projects/:projectid", authorizer.RequireProjectPermission("projectid", authz.PermProjectView), projectHandler.Update)
}

func RegisterProjectWSTickets(privateGroup fiber.Router, repos *repositories.Repositories, redisClient *redis.Client) {
	authorizer := services.NewProjectAuthorizer(repos.ProjectRepository)
	wsTicketHandler := handlers.NewWSTicketHandler(redisClient)
	privateGroup.Post("/projects/:projectid/ws-ticket", authorizer.RequireProjectPermission("projectid", authz.PermProjectView), wsTicketHandler.Mint)
}
