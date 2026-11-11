package handlers

import (
	"errors"
	"strconv"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

type NotificationHandler struct {
	repo repositories.NotificationRepositoryInterface
}

func NewNotificationHandler(repo repositories.NotificationRepositoryInterface) *NotificationHandler {
	return &NotificationHandler{repo: repo}
}

// GetNotifications lists the caller's latest notifications with stats.
// Query: filter (all|unread|<type>), search, limit (max 50).
func (h *NotificationHandler) GetNotifications(c *fiber.Ctx) error {
	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	list, err := h.repo.ListForUser(c.Context(), userID, models.NotificationFilter{
		Filter: c.Query("filter", "all"),
		Search: c.Query("search"),
		Limit:  limit,
	})
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to fetch notifications", err, userID)
	}
	return utils.SendJSON(c, fiber.StatusOK, list)
}

type setNotificationReadRequest struct {
	Read *bool `json:"read"`
}

// SetNotificationRead marks one of the caller's notifications read (default)
// or unread.
func (h *NotificationHandler) SetNotificationRead(c *fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "id")
	if err != nil {
		return err
	}
	var req setNotificationReadRequest
	if len(c.Body()) > 0 {
		if err := utils.ValidateJSONBody(c, &req); err != nil {
			return err
		}
	}
	read := true
	if req.Read != nil {
		read = *req.Read
	}
	n, err := h.repo.SetRead(c.Context(), userID, ids[0], read)
	if errors.Is(err, repositories.ErrNotificationNotFound) {
		return utils.SendError(c, fiber.StatusNotFound, "Notification not found", err, userID)
	}
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to update notification", err, userID)
	}
	return utils.SendJSON(c, fiber.StatusOK, n)
}

// MarkAllRead marks every unread notification of the caller as read.
func (h *NotificationHandler) MarkAllRead(c *fiber.Ctx) error {
	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}
	if err := h.repo.MarkAllRead(c.Context(), userID); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to mark notifications as read", err, userID)
	}
	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{"success": true})
}

type profileUpdateRequest struct {
	Description string `json:"description" validate:"max=2000"`
}

// CreateProfileUpdate records a "Profile Updated" notification for the caller.
func (h *NotificationHandler) CreateProfileUpdate(c *fiber.Ctx) error {
	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}
	var req profileUpdateRequest
	if len(c.Body()) > 0 {
		if err := utils.ValidateAndParseBody(c, &req); err != nil {
			return err
		}
	}
	if req.Description == "" {
		req.Description = "You updated your profile information."
	}
	n := &models.Notification{
		UserId:      userID,
		Type:        "settings",
		Title:       "Profile Updated",
		Description: req.Description,
	}
	if err := h.repo.Create(c.Context(), n); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to create notification", err, userID)
	}
	return utils.SendJSON(c, fiber.StatusCreated, fiber.Map{"success": true})
}
