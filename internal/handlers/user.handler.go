package handlers

import (
	"log"
	"my-go-app/internal/repositories"
	"my-go-app/pkg/utils"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	userRepository repositories.UserRepositoryInterface
}

func NewUserHandler(userRepo repositories.UserRepositoryInterface) *UserHandler {
	return &UserHandler{
		userRepository: userRepo,
	}
}

func (h *UserHandler) SearchUsers(c *fiber.Ctx) error {
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Query parameter 'q' is required")
	}

	users, err := h.userRepository.SearchUsers(c.Context(), query)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to search users")
	}

	log.Printf("Found %d users matching query '%s'\n", len(users), query)
	return utils.SendJSON(c, fiber.StatusOK, users)
}

func (h *UserHandler) GetUserByEmail(c *fiber.Ctx) error {
	email, err := utils.ValidateRequiredParam(c, "email")
	if err != nil {
		return err
	}

	user, err := h.userRepository.GetUserByEmail(c.Context(), email)
	if err != nil {
		return utils.HandleRepoError(c, err, "User not found", "Failed to retrieve user")
	}

	return utils.SendJSON(c, fiber.StatusOK, user)
}

func (h *UserHandler) GetUserByUsername(c *fiber.Ctx) error {
	username, err := utils.ValidateRequiredParam(c, "username")
	if err != nil {
		return err
	}

	user, err := h.userRepository.GetUserByUsername(c.Context(), username)
	if err != nil {
		return utils.HandleRepoError(c, err, "User not found", "Failed to retrieve user")
	}

	return utils.SendJSON(c, fiber.StatusOK, user)
}