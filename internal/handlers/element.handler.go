package handlers

import (
	"log"
	"my-go-app/internal/repositories"
	"my-go-app/pkg/utils"

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
