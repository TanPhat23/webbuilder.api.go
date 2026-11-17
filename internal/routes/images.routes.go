package routes

import (
	"my-go-app/internal/handlers"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"

	"github.com/gofiber/fiber/v3"
)

func ImageRoutes(group fiber.Router, repos *repositories.RepositoriesInterface, cloudinaryService *services.CloudinaryService) {
	imageService := services.NewImageService(repos.ImageRepository, cloudinaryService)
	imageHandler := handlers.NewImageHandler(imageService, cloudinaryService)

	group.Post("/images", imageHandler.UploadImage)
	group.Get("/images", imageHandler.GetUserImages)
	group.Get("/images/:imageid", imageHandler.GetImageByID)
	group.Delete("/images/:imageid", imageHandler.DeleteImage)
}
