package routes

import (
	"my-go-app/internal/handlers"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"

	"github.com/gofiber/fiber/v3"
)

func RegisterUsers(group fiber.Router, repos *repositories.Repositories) {
	userService := services.NewUserService(repos.UserRepository)
	userHandler := handlers.NewUserHandler(userService)

	group.Get("/users/search", userHandler.Search)
	group.Get("/users/email/:email", userHandler.GetByEmail)
	group.Get("/users/username/:username", userHandler.GetByUsername)
	group.Get("/users/:userid", userHandler.Search)
}
