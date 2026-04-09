package routes

import (
	"my-go-app/internal/handlers"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"

	"github.com/gofiber/fiber/v3"
)

func UserRoutes(group fiber.Router, repos *repositories.RepositoriesInterface) {
	userService := services.NewUserService(repos.UserRepository)
	userHandler := handlers.NewUserHandler(userService)

	group.Get("/users/:userid", userHandler.SearchUsers)
	group.Get("/users/email/:email", userHandler.GetUserByEmail)
	group.Get("/users/username/:username", userHandler.GetUserByUsername)
	group.Get("/users/search", userHandler.SearchUsers)
}
