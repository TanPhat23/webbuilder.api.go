package handlers

import (
	"log"
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
		log.Println("Error retrieving elements:", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve elements", err)
	}
	return utils.SendJSON(c, fiber.StatusOK, elements)
}

func (h *ElementHandler) GetElementsByPageIds(c *fiber.Ctx) error {
	// Get pageIds from query parameter (comma-separated)
	pageIdsParam := c.Query("pageIds")
	if pageIdsParam == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "pageIds query parameter is required", nil)
	}

	// Split the comma-separated page IDs
	pageIDs := []string{}
	for _, id := range strings.Split(pageIdsParam, ",") {
		trimmedID := strings.TrimSpace(id)
		if trimmedID != "" {
			pageIDs = append(pageIDs, trimmedID)
		}
	}

	if len(pageIDs) == 0 {
		return utils.SendError(c, fiber.StatusBadRequest, "At least one valid pageId is required", nil)
	}

	elements, err := h.elementRepo.GetElementsByPageIds(c.Context(), pageIDs)
	if err != nil {
		log.Println("Error retrieving elements by page IDs:", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve elements", err)
	}
	return utils.SendJSON(c, fiber.StatusOK, elements)
}
