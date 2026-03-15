package handlers

import (
	"my-go-app/internal/dto"
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
		return utils.HandleRepoError(c, err, "", "Failed to retrieve custom element types")
	}

	return utils.SendJSON(c, fiber.StatusOK, customElementTypes)
}

func (h *CustomElementTypeHandler) GetCustomElementTypeByID(c *fiber.Ctx) error {
	ids, err := utils.MustParams(c, "id")
	if err != nil {
		return err
	}
	id := ids[0]

	customElementType, err := h.customElementTypeRepo.GetCustomElementTypeByID(c.Context(), id)
	if err != nil {
		if err == repositories.ErrCustomElementTypeNotFound {
			return fiber.NewError(fiber.StatusNotFound, "Custom element type not found")
		}
		return utils.HandleRepoError(c, err, "Custom element type not found", "Failed to retrieve custom element type")
	}

	return utils.SendJSON(c, fiber.StatusOK, customElementType)
}

func (h *CustomElementTypeHandler) CreateCustomElementType(c *fiber.Ctx) error {
	var req dto.CreateCustomElementTypeRequest
	if err := utils.ValidateAndParseBody(c, &req); err != nil {
		return err
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
			return fiber.NewError(fiber.StatusConflict, "Custom element type with this name already exists")
		}
		return utils.HandleRepoError(c, err, "", "Failed to create custom element type")
	}

	return utils.SendJSON(c, fiber.StatusCreated, created)
}

func (h *CustomElementTypeHandler) UpdateCustomElementType(c *fiber.Ctx) error {
	ids, err := utils.MustParams(c, "id")
	if err != nil {
		return err
	}
	id := ids[0]

	var req dto.UpdateCustomElementTypeRequest
	if err := utils.ValidateAndParseBody(c, &req); err != nil {
		return err
	}

	updates := map[string]any{}
	if req.Name != nil        { updates["name"] = *req.Name }
	if req.Description != nil { updates["description"] = *req.Description }
	if req.Category != nil    { updates["category"] = *req.Category }
	if req.Icon != nil        { updates["icon"] = *req.Icon }

	if err := utils.RequireUpdates(updates); err != nil {
		return err
	}

	updated, err := h.customElementTypeRepo.UpdateCustomElementType(c.Context(), id, updates)
	if err != nil {
		if err == repositories.ErrCustomElementTypeNotFound {
			return fiber.NewError(fiber.StatusNotFound, "Custom element type not found")
		}
		return utils.HandleRepoError(c, err, "Custom element type not found", "Failed to update custom element type")
	}

	return utils.SendJSON(c, fiber.StatusOK, updated)
}

func (h *CustomElementTypeHandler) DeleteCustomElementType(c *fiber.Ctx) error {
	ids, err := utils.MustParams(c, "id")
	if err != nil {
		return err
	}
	id := ids[0]

	if err := h.customElementTypeRepo.DeleteCustomElementType(c.Context(), id); err != nil {
		if err == repositories.ErrCustomElementTypeNotFound {
			return fiber.NewError(fiber.StatusNotFound, "Custom element type not found")
		}
		return utils.HandleRepoError(c, err, "Custom element type not found", "Failed to delete custom element type")
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Custom element type deleted successfully")
}