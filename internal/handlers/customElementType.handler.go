package handlers

import (
	"my-go-app/internal/dto"
	"my-go-app/internal/models"
	"my-go-app/internal/services"
	"my-go-app/pkg/utils"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type CustomElementTypeHandler struct {
	customElementTypeService services.CustomElementTypeServiceInterface
}

func NewCustomElementTypeHandler(customElementTypeService services.CustomElementTypeServiceInterface) *CustomElementTypeHandler {
	return &CustomElementTypeHandler{
		customElementTypeService: customElementTypeService,
	}
}

func (h *CustomElementTypeHandler) List(c fiber.Ctx) error {
	customElementTypes, err := h.customElementTypeService.List(c.RequestCtx())
	if err != nil {
		return utils.RespondError(c, err, "", "Failed to retrieve custom element types")
	}

	return utils.SendJSON(c, fiber.StatusOK, customElementTypes)
}

func (h *CustomElementTypeHandler) GetByID(c fiber.Ctx) error {
	ids, err := utils.RequireParams(c, "id")
	if err != nil {
		return err
	}
	id := ids[0]

	customElementType, err := h.customElementTypeService.GetByID(c.RequestCtx(), id)
	if err != nil {
		return utils.RespondError(c, err, "Custom element type not found", "Failed to retrieve custom element type")
	}

	return utils.SendJSON(c, fiber.StatusOK, customElementType)
}

func (h *CustomElementTypeHandler) Create(c fiber.Ctx) error {
	req := new(dto.CreateCustomElementTypeRequest)
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	customElementType := &models.CustomElementType{
		Id:          uuid.NewString(),
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Icon:        req.Icon,
	}

	created, err := h.customElementTypeService.Create(c.RequestCtx(), customElementType)
	if err != nil {
		return utils.RespondError(c, err, "", "Failed to create custom element type")
	}

	return utils.SendJSON(c, fiber.StatusCreated, created)
}

func (h *CustomElementTypeHandler) Update(c fiber.Ctx) error {
	ids, err := utils.RequireParams(c, "id")
	if err != nil {
		return err
	}
	id := ids[0]

	req := new(dto.UpdateCustomElementTypeRequest)
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	updates := map[string]any{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Category != nil {
		updates["category"] = *req.Category
	}
	if req.Icon != nil {
		updates["icon"] = *req.Icon
	}

	if err := utils.RequireUpdates(updates); err != nil {
		return err
	}

	updated, err := h.customElementTypeService.Update(c.RequestCtx(), id, updates)
	if err != nil {
		return utils.RespondError(c, err, "Custom element type not found", "Failed to update custom element type")
	}

	return utils.SendJSON(c, fiber.StatusOK, updated)
}

func (h *CustomElementTypeHandler) Delete(c fiber.Ctx) error {
	ids, err := utils.RequireParams(c, "id")
	if err != nil {
		return err
	}
	id := ids[0]

	if err := h.customElementTypeService.Delete(c.RequestCtx(), id); err != nil {
		return utils.RespondError(c, err, "Custom element type not found", "Failed to delete custom element type")
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Custom element type deleted successfully")
}

// fiber:context-methods migrated
