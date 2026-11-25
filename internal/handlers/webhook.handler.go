package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v3"

	"my-go-app/internal/dto"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"
	"my-go-app/pkg/utils"
)

type WebhookHandler struct {
	service *services.WebhookService
}

func NewWebhookHandler(service *services.WebhookService) *WebhookHandler {
	return &WebhookHandler{service: service}
}

func (h *WebhookHandler) List(c fiber.Ctx) error {
	userID, ids, err := utils.RequireUserAndParams(c, "projectid")
	if err != nil {
		return err
	}
	webhooks, err := h.service.ListByProjectForUser(c.RequestCtx(), ids[0], userID)
	if err != nil {
		return utils.RespondError(c, err, "Project not found", "Failed to list webhooks")
	}
	return utils.SendJSON(c, fiber.StatusOK, webhooks)
}

func (h *WebhookHandler) Create(c fiber.Ctx) error {
	userID, ids, err := utils.RequireUserAndParams(c, "projectid")
	if err != nil {
		return err
	}
	req := new(dto.CreateWebhookRequest)
	if err := c.Bind().Body(req); err != nil {
		return err
	}
	created, err := h.service.CreateForUser(c.RequestCtx(), ids[0], userID, req)
	if err != nil {
		if errors.Is(err, services.ErrBlockedWebhookURL) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return utils.RespondError(c, err, "Project or collection not found", "Failed to create webhook")
	}
	return utils.SendJSON(c, fiber.StatusCreated, created)
}

func (h *WebhookHandler) Update(c fiber.Ctx) error {
	userID, ids, err := utils.RequireUserAndParams(c, "projectid", "webhookId")
	if err != nil {
		return err
	}
	req := new(dto.UpdateWebhookRequest)
	if err := c.Bind().Body(req); err != nil {
		return err
	}
	updated, err := h.service.UpdateForUser(c.RequestCtx(), ids[0], ids[1], userID, req)
	if err != nil {
		if errors.Is(err, services.ErrBlockedWebhookURL) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		if errors.Is(err, repositories.ErrWebhookNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Webhook not found"})
		}
		return utils.RespondError(c, err, "Webhook not found", "Failed to update webhook")
	}
	return utils.SendJSON(c, fiber.StatusOK, updated)
}

func (h *WebhookHandler) Delete(c fiber.Ctx) error {
	userID, ids, err := utils.RequireUserAndParams(c, "projectid", "webhookId")
	if err != nil {
		return err
	}
	if err := h.service.DeleteForUser(c.RequestCtx(), ids[0], ids[1], userID); err != nil {
		if errors.Is(err, repositories.ErrWebhookNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Webhook not found"})
		}
		return utils.RespondError(c, err, "Webhook not found", "Failed to delete webhook")
	}
	return utils.SendNoContent(c)
}

func (h *WebhookHandler) ListDeliveries(c fiber.Ctx) error {
	userID, ids, err := utils.RequireUserAndParams(c, "projectid", "webhookId")
	if err != nil {
		return err
	}
	limit := utils.ParsePositiveIntQuery(c, "limit", 25)
	deliveries, err := h.service.ListDeliveriesForUser(c.RequestCtx(), ids[0], ids[1], userID, limit)
	if err != nil {
		if errors.Is(err, repositories.ErrWebhookNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Webhook not found"})
		}
		return utils.RespondError(c, err, "Webhook not found", "Failed to list webhook deliveries")
	}
	return utils.SendJSON(c, fiber.StatusOK, deliveries)
}

func (h *WebhookHandler) SendTest(c fiber.Ctx) error {
	userID, ids, err := utils.RequireUserAndParams(c, "projectid", "webhookId")
	if err != nil {
		return err
	}
	delivery, err := h.service.SendTestWebhookForUser(c.RequestCtx(), ids[0], ids[1], userID)
	if err != nil {
		if errors.Is(err, repositories.ErrWebhookNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Webhook not found"})
		}
		return utils.RespondError(c, err, "Webhook not found", "Failed to send test webhook")
	}
	return utils.SendJSON(c, fiber.StatusOK, delivery)
}
