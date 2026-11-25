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

func SetupRoutes(app *fiber.App, repos *repositories.RepositoriesInterface, storage services.ObjectStorage, invitationService *services.InvitationService, redisClient *redis.Client) {
	publicGroup := app.Group("/api/v1")
	privateGroup := app.Group("/api/v1", middleware.NewAuthenticateMiddleware(repos.AgentTokenRepository))

	onAccessChanged := publishAccessChange(redisClient)

	var webhookService *services.WebhookService
	var itemEmitter services.ItemEventEmitter
	if repos.WebhookRepository != nil {
		webhookService = services.NewWebhookService(repos.WebhookRepository, repos.ProjectRepository, repos.ContentTypeRepository)
		webhookService.StartWorkers(context.Background(), 4)
		itemEmitter = webhookService
	}

	ElementRoutes(publicGroup, privateGroup, repos, redisClient)
	PageRoutes(publicGroup, privateGroup, repos, redisClient)
	ProjectRoutes(publicGroup, privateGroup, repos, onAccessChanged)
	ProjectWSTicketRoutes(privateGroup, repos, redisClient)
	ContentRoutes(publicGroup, privateGroup, repos, itemEmitter)
	if webhookService != nil {
		WebhookRoutes(privateGroup, repos, webhookService)
	}

	AgentTokenRoutes(privateGroup, repos)
	SnapshotRoutes(privateGroup, repos)
	ImageRoutes(publicGroup, privateGroup, repos, storage)
	CustomElementRoutes(privateGroup, repos)
	InvitationRoutes(privateGroup, invitationService)
	CollaboratorRoutes(privateGroup, repos, onAccessChanged)
	ElementCommentRoutes(privateGroup, repos)
	UserRoutes(privateGroup, repos)
	WorkflowRoutes(privateGroup, repos)
	NotificationRoutes(privateGroup, repos)
}
