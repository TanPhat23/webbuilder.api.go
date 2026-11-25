package routes

import (
	"my-go-app/internal/handlers"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"

	"github.com/gofiber/fiber/v3"
)

func RegisterImages(publicGroup, group fiber.Router, repos *repositories.Repositories, storage services.ObjectStorage) {
	imageService := services.NewImageService(repos.ImageRepository, storage)
	imageHandler := handlers.NewImageHandler(imageService)

	publicGroup.Get("/public/images/:imageid", imageHandler.ServePublic)

	group.Post("/images", imageHandler.Upload)
	group.Get("/images", imageHandler.ListByUser)
	group.Get("/images/:imageid", imageHandler.GetByID)
	group.Delete("/images/:imageid", imageHandler.Delete)
}
