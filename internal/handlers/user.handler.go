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

func (h *UserHandler) Search(c fiber.Ctx) error {
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Query parameter 'q' is required")
	}
	if len([]rune(query)) < 2 {
		return fiber.NewError(fiber.StatusBadRequest, "Query parameter 'q' must be at least 2 characters")
	}

	users, err := h.userService.Search(c.RequestCtx(), query)
	if err != nil {
		return utils.RespondError(c, err, "", "Failed to search users")
	}

	log.Printf("Found %d users matching query '%s'\n", len(users), query)
	return utils.SendJSON(c, fiber.StatusOK, users)
}

func (h *UserHandler) GetByEmail(c fiber.Ctx) error {
	email, err := utils.RequireParam(c, "email")
	if err != nil {
		return err
	}

	user, err := h.userService.GetByEmail(c.RequestCtx(), email)
	if err != nil {
		return utils.RespondError(c, err, "User not found", "Failed to retrieve user")
	}

	return utils.SendJSON(c, fiber.StatusOK, user)
}

func (h *UserHandler) GetByUsername(c fiber.Ctx) error {
	username, err := utils.RequireParam(c, "username")
	if err != nil {
		return err
	}

	user, err := h.userService.GetByUsername(c.RequestCtx(), username)
	if err != nil {
		return utils.RespondError(c, err, "User not found", "Failed to retrieve user")
	}

	return utils.SendJSON(c, fiber.StatusOK, user)
}

// fiber:context-methods migrated
