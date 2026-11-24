package handlers

import (
	"encoding/json"
	"errors"

	"github.com/gofiber/fiber/v3"

	"my-go-app/internal/repositories"
	"my-go-app/internal/services"
	"my-go-app/pkg/utils"
)

type PublicCollectionHandler struct {
	service *services.PublicCollectionService
}

func NewPublicCollectionHandler(service *services.PublicCollectionService) *PublicCollectionHandler {
	return &PublicCollectionHandler{service: service}
}

func (h *PublicCollectionHandler) ListPublicItems(c fiber.Ctx) error {
	projectID := c.Params("projectId")
	collection := c.Params("collection")
	if projectID == "" || collection == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "projectId and collection are required"})
	}

	limit := utils.ParsePositiveIntQuery(c, "limit", 20)
	offset := utils.ParsePositiveIntQuery(c, "offset", 0)
	if c.Query("offset") == "" {
		offset = 0
	}
	sortBy := resolveContentItemSortBy(c.Query("sortBy", "createdAt"))
	sortOrder := c.Query("sortOrder", "desc")

	page, err := h.service.ListPublicItems(c.RequestCtx(), projectID, collection, limit, offset, sortBy, sortOrder)
	if err != nil {
		if errors.Is(err, repositories.ErrContentTypeNotFound) || errors.Is(err, services.ErrCollectionPublicReadDisabled) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Collection not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve collection items"})
	}
	return utils.SendJSON(c, fiber.StatusOK, page)
}

func (h *PublicCollectionHandler) SubmitPublicItem(c fiber.Ctx) error {
	projectID := c.Params("projectId")
	collection := c.Params("collection")
	if projectID == "" || collection == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "projectId and collection are required"})
	}

	var payload map[string]any
	if err := json.Unmarshal(c.Body(), &payload); err != nil || payload == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Request body must be a JSON object"})
	}

	result, err := h.service.SubmitPublicItem(c.RequestCtx(), projectID, collection, c.IP(), payload)
	if err != nil {
		if errors.Is(err, repositories.ErrContentTypeNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Collection not found"})
		}
		if errors.Is(err, services.ErrCollectionPublicCreateDisabled) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Public submissions are disabled for this collection"})
		}
		if errors.Is(err, services.ErrCollectionRateLimited) {
			c.Set("Retry-After", "60")
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"error": "Too many submissions, please try again later"})
		}
		var valErr *services.ValidationError
		if errors.As(err, &valErr) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":  "validation_failed",
				"fields": valErr.Fields,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to submit form"})
	}

	return utils.SendJSON(c, fiber.StatusCreated, result)
}
