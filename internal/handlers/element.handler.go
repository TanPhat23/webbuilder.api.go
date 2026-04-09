package handlers

import (
	"my-go-app/internal/services"
	"my-go-app/pkg/utils"
	"strings"

	"github.com/gofiber/fiber/v3"
)

type ElementHandler struct {
	elementService services.ElementServiceInterface
}

func NewElementHandler(elementService services.ElementServiceInterface) *ElementHandler {
	return &ElementHandler{
		elementService: elementService,
	}
}

func (h *ElementHandler) GetElements(c fiber.Ctx) error {
	projectID, err := utils.ValidateRequiredParam(c, "projectid")
	if err != nil {
		return err
	}

	elements, err := h.elementService.GetElements(c.RequestCtx(), projectID)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to retrieve elements")
	}

	return utils.SendJSON(c, fiber.StatusOK, elements)
}

func (h *ElementHandler) GetElementsByPageIds(c fiber.Ctx) error {
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

	elements, err := h.elementService.GetElementsByPageIds(c.RequestCtx(), pageIDs)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to retrieve elements")
	}

	return utils.SendJSON(c, fiber.StatusOK, elements)
}

// fiber:context-methods migrated
