package routes

import (
	"my-go-app/internal/handlers"
	"my-go-app/internal/repositories"

	"github.com/gofiber/fiber/v3"
)

func NotificationRoutes(group fiber.Router, repos *repositories.RepositoriesInterface) {
	notificationHandler := handlers.NewNotificationHandler(repos.NotificationRepository)

	group.Get("/notifications", notificationHandler.GetNotifications)
	group.Post("/notifications/read-all", notificationHandler.MarkAllRead)
	group.Post("/notifications/profile-update", notificationHandler.CreateProfileUpdate)
	group.Patch("/notifications/:id", notificationHandler.SetNotificationRead)
}
