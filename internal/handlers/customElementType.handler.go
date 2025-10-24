package handlers

import (
	"log"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type CustomElementTypeHandler struct {
	customElementTypeRepo repositories.CustomElementTypeRepositoryInterface
}

func NewCustomElementTypeHandler(customElementTypeRepo repositories.CustomElementTypeRepositoryInterface) *CustomElementTypeHandler {
	return &CustomElementTypeHandler{
		customElementTypeRepo: customElementTypeRepo,
	}
}

func (h *CustomElementTypeHandler) GetCustomElementTypes(c *fiber.Ctx) error {
	customElementTypes, err := h.customElementTypeRepo.GetCustomElementTypes(c.Context())
	if err != nil {
		log.Println("Error retrieving custom element types:", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve custom element types", err)
	}

	return utils.SendJSON(c, fiber.StatusOK, customElementTypes)
}

func (h *CustomElementTypeHandler) GetCustomElementTypeByID(c *fiber.Ctx) error {
	id, err := utils.ValidateRequiredParam(c, "id")
	if err != nil {
		return err
	}

	customElementType, err := h.customElementTypeRepo.GetCustomElementTypeByID(c.Context(), id)
	if err != nil {
		if err == repositories.ErrCustomElementTypeNotFound {
			return utils.SendError(c, fiber.StatusNotFound, "Custom element type not found", err)
		}
		log.Println("Error retrieving custom element type:", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve custom element type", err)
	}

	return utils.SendJSON(c, fiber.StatusOK, customElementType)
}

func (h *CustomElementTypeHandler) CreateCustomElementType(c *fiber.Ctx) error {
	var req struct {
		Name        string  `json:"name"`
		Description *string `json:"description"`
		Category    *string `json:"category"`
		Icon        *string `json:"icon"`
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	if req.Name == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "Name is required", nil)
	}

	customElementType := &models.CustomElementType{
		Id:          uuid.NewString(),
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Icon:        req.Icon,
	}

	created, err := h.customElementTypeRepo.CreateCustomElementType(c.Context(), customElementType)
	if err != nil {
		if err == repositories.ErrCustomElementTypeAlreadyExists {
			return utils.SendError(c, fiber.StatusConflict, "Custom element type with this name already exists", err)
		}
		log.Println("Error creating custom element type:", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to create custom element type", err)
	}

	return utils.SendJSON(c, fiber.StatusCreated, created)
}

func (h *CustomElementTypeHandler) UpdateCustomElementType(c *fiber.Ctx) error {
	id, err := utils.ValidateRequiredParam(c, "id")
	if err != nil {
		return err
	}

	var req map[string]any
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	delete(req, "id")
	delete(req, "createdAt")

	updated, err := h.customElementTypeRepo.UpdateCustomElementType(c.Context(), id, req)
	if err != nil {
		if err == repositories.ErrCustomElementTypeNotFound {
			return utils.SendError(c, fiber.StatusNotFound, "Custom element type not found", err)
		}
		log.Println("Error updating custom element type:", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to update custom element type", err)
	}

	return utils.SendJSON(c, fiber.StatusOK, updated)
}

func (h *CustomElementTypeHandler) DeleteCustomElementType(c *fiber.Ctx) error {
	id, err := utils.ValidateRequiredParam(c, "id")
	if err != nil {
		return err
	}

	err = h.customElementTypeRepo.DeleteCustomElementType(c.Context(), id)
	if err != nil {
		if err == repositories.ErrCustomElementTypeNotFound {
			return utils.SendError(c, fiber.StatusNotFound, "Custom element type not found", err)
		}
		log.Println("Error deleting custom element type:", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to delete custom element type", err)
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{
		"message": "Custom element type deleted successfully",
	})
}
