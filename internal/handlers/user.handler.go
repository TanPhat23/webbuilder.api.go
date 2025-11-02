package handlers

import (
	"my-go-app/internal/repositories"
	"my-go-app/pkg/utils"

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
	query := c.Query("q")
	if query == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "Query parameter 'q' is required", nil, "")
	}

	users, err := h.userRepository.SearchUsers(c.Context(), query)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to search users", err, "")
	}

	return utils.SendJSON(c, fiber.StatusOK, users)
}

func (h *UserHandler) GetUserByEmail(c *fiber.Ctx) error {
	email := c.Params("email")
	if email == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "Email parameter is required", nil, "")
	}

	user, err := h.userRepository.GetUserByEmail(c.Context(), email)
	if err != nil {
		if err.Error() == "user not found" {
			return utils.SendError(c, fiber.StatusNotFound, "User not found", err, "")
		}
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve user", err, "")
	}

	return utils.SendJSON(c, fiber.StatusOK, user)
}

func (h *UserHandler) GetUserByUsername(c *fiber.Ctx) error {
	username := c.Params("username")
	if username == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "Username parameter is required", nil, "")
	}

	user, err := h.userRepository.GetUserByUsername(c.Context(), username)
	if err != nil {
		if err.Error() == "user not found" {
			return utils.SendError(c, fiber.StatusNotFound, "User not found", err, "")
		}
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve user", err, "")
	}

	return utils.SendJSON(c, fiber.StatusOK, user)
}
