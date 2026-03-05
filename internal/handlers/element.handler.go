package handlers

import (
	"my-go-app/internal/repositories"
	"my-go-app/pkg/utils"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type ElementHandler struct {
	elementRepo repositories.ElementRepositoryInterface
}

func NewElementHandler(elementRepo repositories.ElementRepositoryInterface) *ElementHandler {
	return &ElementHandler{
		elementRepo: elementRepo,
	}
}

func (h *ElementHandler) GetElements(c *fiber.Ctx) error {
	projectID, err := utils.ValidateRequiredParam(c, "projectid")
	if err != nil {
		return err
	}

	elements, err := h.elementRepo.GetElements(c.Context(), projectID)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to retrieve elements")
	}

	return utils.SendJSON(c, fiber.StatusOK, elements)
}

func (h *ElementHandler) GetElementsByPageIds(c *fiber.Ctx) error {
	pageIdsParam := c.Query("pageIds")
	if pageIdsParam == "" {
		return fiber.NewError(fiber.StatusBadRequest, "pageIds query parameter is required")
	}

	var pageIDs []string
	for _, id := range strings.Split(pageIdsParam, ",") {
		if trimmed := strings.TrimSpace(id); trimmed != "" {
			pageIDs = append(pageIDs, trimmed)
		}
	}

	if len(pageIDs) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "At least one valid pageId is required")
	}

	elements, err := h.elementRepo.GetElementsByPageIds(c.Context(), pageIDs)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to retrieve elements")
	}

	return utils.SendJSON(c, fiber.StatusOK, elements)
}