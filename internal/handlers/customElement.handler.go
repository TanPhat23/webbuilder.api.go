package handlers

import (
	"errors"
	"my-go-app/internal/dto"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"
	"my-go-app/pkg/utils"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type CustomElementHandler struct {
	customElementService services.CustomElementServiceInterface
}

func NewCustomElementHandler(customElementService services.CustomElementServiceInterface) *CustomElementHandler {
	return &CustomElementHandler{
		customElementService: customElementService,
	}
}

func (h *CustomElementHandler) List(c fiber.Ctx) error {
	userID, err := utils.RequireUserID(c)
	if err != nil {
		return err
	}

	var isPublicPtr *bool
	if isPublicStr := c.Query("isPublic"); isPublicStr != "" {
		if isPublic, err := strconv.ParseBool(isPublicStr); err == nil {
			isPublicPtr = &isPublic
		}
	}

	customElements, err := h.customElementService.List(c.RequestCtx(), userID, isPublicPtr)
	if err != nil {
		return utils.RespondError(c, err, "", "Failed to retrieve custom elements")
	}

	return utils.SendJSON(c, fiber.StatusOK, customElements)
}

func (h *CustomElementHandler) ListPublic(c fiber.Ctx) error {
	limit, offset := utils.ParsePagination(c, 20)

	var categoryPtr *string
	if category := c.Query("category"); category != "" {
		categoryPtr = &category
	}

	customElements, err := h.customElementService.ListPublic(c.RequestCtx(), categoryPtr, limit, offset)
	if err != nil {
		return utils.RespondError(c, err, "", "Failed to retrieve public custom elements")
	}

	return utils.SendJSON(c, fiber.StatusOK, customElements)
}

func (h *CustomElementHandler) GetByID(c fiber.Ctx) error {
	userID, ids, err := utils.RequireUserAndParams(c, "id")
	if err != nil {
		return err
	}
	id := ids[0]

	customElement, err := h.customElementService.GetByID(c.RequestCtx(), id, userID)
	if err != nil {
		return respondCustomElementError(c, err, "Failed to retrieve custom element", "")
	}

	return utils.SendJSON(c, fiber.StatusOK, customElement)
}

func (h *CustomElementHandler) Create(c fiber.Ctx) error {
	userID, err := utils.RequireUserID(c)
	if err != nil {
		return err
	}

	req := new(dto.CreateCustomElementRequest)
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	if req.Version == "" {
		req.Version = "1.0.0"
	}

	customElement := &models.CustomElement{
		Id:           uuid.NewString(),
		Name:         req.Name,
		TypeId:       req.TypeId,
		Description:  req.Description,
		Category:     req.Category,
		Icon:         req.Icon,
		Thumbnail:    req.Thumbnail,
		Structure:    req.Structure,
		DefaultProps: req.DefaultProps,
		Tags:         req.Tags,
		UserId:       userID,
		IsPublic:     req.IsPublic,
		Version:      req.Version,
	}

	created, err := h.customElementService.Create(c.RequestCtx(), customElement)
	if err != nil {
		return respondCustomElementError(c, err, "Failed to create custom element", "")
	}

	return utils.SendJSON(c, fiber.StatusCreated, created)
}

func (h *CustomElementHandler) Update(c fiber.Ctx) error {
	userID, ids, err := utils.RequireUserAndParams(c, "id")
	if err != nil {
		return err
	}
	id := ids[0]

	req := new(dto.UpdateCustomElementRequest)
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	updates := buildCustomElementUpdates(req)
	if err := utils.RequireUpdates(updates); err != nil {
		return err
	}

	updated, err := h.customElementService.Update(c.RequestCtx(), id, userID, updates)
	if err != nil {
		return respondCustomElementError(c, err, "Failed to update custom element", "Unauthorized to update this custom element")
	}

	return utils.SendJSON(c, fiber.StatusOK, updated)
}

func buildCustomElementUpdates(req *dto.UpdateCustomElementRequest) map[string]any {
	updates := map[string]any{}
	utils.SetIfNotNil(updates, "name", req.Name)
	utils.SetIfNotNil(updates, "typeId", req.TypeId)
	utils.SetIfNotNil(updates, "description", req.Description)
	utils.SetIfNotNil(updates, "category", req.Category)
	utils.SetIfNotNil(updates, "icon", req.Icon)
	utils.SetIfNotNil(updates, "thumbnail", req.Thumbnail)
	utils.SetIfNotNil(updates, "tags", req.Tags)
	utils.SetIfNotNil(updates, "isPublic", req.IsPublic)
	utils.SetIfNotNil(updates, "version", req.Version)
	if req.Structure != nil {
		updates["structure"] = req.Structure
	}
	if req.DefaultProps != nil {
		updates["defaultProps"] = req.DefaultProps
	}
	return updates
}

func respondCustomElementError(c fiber.Ctx, err error, internalMsg, forbiddenMsg string) error {
	if errors.Is(err, repositories.ErrCustomElementAlreadyExists) {
		return fiber.NewError(fiber.StatusConflict, "Custom element with this name already exists")
	}
	if forbiddenMsg != "" && strings.Contains(err.Error(), "unauthorized") {
		return fiber.NewError(fiber.StatusForbidden, forbiddenMsg)
	}
	return utils.RespondError(c, err, "Custom element not found", internalMsg)
}

func (h *CustomElementHandler) Delete(c fiber.Ctx) error {
	userID, ids, err := utils.RequireUserAndParams(c, "id")
	if err != nil {
		return err
	}
	id := ids[0]

	if err := h.customElementService.Delete(c.RequestCtx(), id, userID); err != nil {
		return respondCustomElementError(c, err, "Failed to delete custom element", "Unauthorized to delete this custom element")
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Custom element deleted successfully")
}

func (h *CustomElementHandler) Duplicate(c fiber.Ctx) error {
	userID, ids, err := utils.RequireUserAndParams(c, "id")
	if err != nil {
		return err
	}
	id := ids[0]

	req := new(dto.DuplicateCustomElementRequest)
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	duplicate, err := h.customElementService.Duplicate(c.RequestCtx(), id, userID, req.NewName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return fiber.NewError(fiber.StatusNotFound, "Custom element not found")
		}
		return utils.RespondError(c, err, "", "Failed to duplicate custom element")
	}

	return utils.SendJSON(c, fiber.StatusCreated, duplicate)
}

// fiber:context-methods migrated
