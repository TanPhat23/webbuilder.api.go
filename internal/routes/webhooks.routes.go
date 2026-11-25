package routes

import (
	"my-go-app/internal/authz"
	"my-go-app/internal/handlers"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"

	"github.com/gofiber/fiber/v3"
)

func RegisterWebhooks(privateGroup fiber.Router, repos *repositories.Repositories, webhookService *services.WebhookService) {
	authorizer := services.NewProjectAuthorizer(repos.ProjectRepository)
	handler := handlers.NewWebhookHandler(webhookService)

	privateGroup.Get("/projects/:projectid/webhooks", authorizer.RequireProjectPermission("projectid", authz.PermCMSView), handler.List)
	privateGroup.Post("/projects/:projectid/webhooks", authorizer.RequireProjectPermission("projectid", authz.PermCMSEdit), handler.Create)
	privateGroup.Patch("/projects/:projectid/webhooks/:webhookId", authorizer.RequireProjectPermission("projectid", authz.PermCMSEdit), handler.Update)
	privateGroup.Delete("/projects/:projectid/webhooks/:webhookId", authorizer.RequireProjectPermission("projectid", authz.PermCMSEdit), handler.Delete)
	privateGroup.Get("/projects/:projectid/webhooks/:webhookId/deliveries", authorizer.RequireProjectPermission("projectid", authz.PermCMSView), handler.ListDeliveries)
	privateGroup.Post("/projects/:projectid/webhooks/:webhookId/test", authorizer.RequireProjectPermission("projectid", authz.PermCMSEdit), handler.SendTest)
}
