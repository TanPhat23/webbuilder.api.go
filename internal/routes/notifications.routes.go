package routes

import (
	"my-go-app/internal/handlers"
	"my-go-app/internal/repositories"

	"github.com/gofiber/fiber/v3"
)

func RegisterNotifications(group fiber.Router, repos *repositories.Repositories) {
	notificationHandler := handlers.NewNotificationHandler(repos.NotificationRepository)

	group.Get("/notifications", notificationHandler.List)
	group.Post("/notifications/read-all", notificationHandler.MarkAllRead)
	group.Post("/notifications/profile-update", notificationHandler.CreateProfileUpdate)
	group.Patch("/notifications/:id", notificationHandler.SetRead)
}
