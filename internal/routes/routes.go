package routes

import (
	"context"
	"encoding/json"

	"my-go-app/internal/repositories"
	"my-go-app/internal/services"
	"my-go-app/pkg/middleware"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
)

const rbacSyncChannel = "rbac:access:changed"

func publishAccessChange(rdb *redis.Client) func(projectID, userID string) {
	if rdb == nil {
		return nil
	}
	return func(projectID, userID string) {
		payload, err := json.Marshal(struct {
			ProjectID string `json:"projectId"`
			UserID    string `json:"userId"`
		}{ProjectID: projectID, UserID: userID})
		if err != nil {
			return
		}
		_ = rdb.Publish(context.Background(), rbacSyncChannel, payload).Err()
	}
}

func Register(app *fiber.App, repos *repositories.Repositories, storage services.ObjectStorage, invitationService *services.InvitationService, redisClient *redis.Client) {
	publicGroup := app.Group("/api/v1")
	privateGroup := app.Group("/api/v1", middleware.NewAuthenticate(repos.AgentTokenRepository))

	onAccessChanged := publishAccessChange(redisClient)

	var webhookService *services.WebhookService
	var itemEmitter services.ItemEventEmitter
	if repos.WebhookRepository != nil {
		webhookService = services.NewWebhookService(repos.WebhookRepository, repos.ProjectRepository, repos.ContentTypeRepository)
		webhookService.StartWorkers(context.Background(), 4)
		itemEmitter = webhookService
	}

	RegisterElements(publicGroup, privateGroup, repos, redisClient)
	RegisterPages(publicGroup, privateGroup, repos, redisClient)
	RegisterProjects(publicGroup, privateGroup, repos, onAccessChanged)
	RegisterProjectWSTickets(privateGroup, repos, redisClient)
	RegisterContent(publicGroup, privateGroup, repos, itemEmitter)
	if webhookService != nil {
		RegisterWebhooks(privateGroup, repos, webhookService)
	}

	RegisterAgentTokens(privateGroup, repos)
	RegisterSnapshots(privateGroup, repos)
	RegisterImages(publicGroup, privateGroup, repos, storage)
	RegisterCustomElements(privateGroup, repos)
	RegisterInvitations(privateGroup, invitationService)
	RegisterCollaborators(privateGroup, repos, onAccessChanged)
	RegisterElementComments(privateGroup, repos)
	RegisterUsers(privateGroup, repos)
	RegisterWorkflows(privateGroup, repos)
	RegisterNotifications(privateGroup, repos)
}
