package handlers

import (
	"log"
	"my-go-app/internal/services"
	"my-go-app/pkg/utils"
	"strings"

	"github.com/gofiber/fiber/v3"
)

type UserHandler struct {
	userService services.UserServiceInterface
}

func NewUserHandler(userService services.UserServiceInterface) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) SearchUsers(c fiber.Ctx) error {
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Query parameter 'q' is required")
	}

	users, err := h.userService.SearchUsers(c.RequestCtx(), query)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to search users")
	}

	log.Printf("Found %d users matching query '%s'\n", len(users), query)
	return utils.SendJSON(c, fiber.StatusOK, users)
}

func (h *UserHandler) GetUserByEmail(c fiber.Ctx) error {
	email, err := utils.ValidateRequiredParam(c, "email")
	if err != nil {
		return err
	}

	user, err := h.userService.GetUserByEmail(c.RequestCtx(), email)
	if err != nil {
		return utils.HandleRepoError(c, err, "User not found", "Failed to retrieve user")
	}

	return utils.SendJSON(c, fiber.StatusOK, user)
}

func (h *UserHandler) GetUserByUsername(c fiber.Ctx) error {
	username, err := utils.ValidateRequiredParam(c, "username")
	if err != nil {
		return err
	}

	user, err := h.userService.GetUserByUsername(c.RequestCtx(), username)
	if err != nil {
		return utils.HandleRepoError(c, err, "User not found", "Failed to retrieve user")
	}

	return utils.SendJSON(c, fiber.StatusOK, user)
}

// fiber:context-methods migrated
