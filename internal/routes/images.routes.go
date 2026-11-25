package routes

import (
	"my-go-app/internal/handlers"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"

	"github.com/gofiber/fiber/v3"
)

func ImageRoutes(publicGroup, group fiber.Router, repos *repositories.RepositoriesInterface, storage services.ObjectStorage) {
	imageService := services.NewImageService(repos.ImageRepository, storage)
	imageHandler := handlers.NewImageHandler(imageService)

	publicGroup.Get("/public/images/:imageid", imageHandler.RedirectPublicImage)

	group.Post("/images", imageHandler.UploadImage)
	group.Get("/images", imageHandler.GetUserImages)
	group.Get("/images/:imageid", imageHandler.GetImageByID)
	group.Delete("/images/:imageid", imageHandler.DeleteImage)
}
