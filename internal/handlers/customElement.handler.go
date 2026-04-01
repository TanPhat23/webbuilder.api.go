package handlers

import (
	"log"
	"my-go-app/internal/dto"
	"my-go-app/internal/models"
	"my-go-app/internal/services"
	"my-go-app/pkg/utils"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
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

func (h *CustomElementHandler) GetCustomElements(c *fiber.Ctx) error {
	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	var isPublicPtr *bool
	if isPublicStr := c.Query("isPublic"); isPublicStr != "" {
		if isPublic, err := strconv.ParseBool(isPublicStr); err == nil {
			isPublicPtr = &isPublic
		}
	}

	customElements, err := h.customElementService.GetCustomElements(c.Context(), userID, isPublicPtr)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to retrieve custom elements")
	}

	return utils.SendJSON(c, fiber.StatusOK, customElements)
}

func (h *CustomElementHandler) GetPublicCustomElements(c *fiber.Ctx) error {
	limit, err := strconv.Atoi(c.Query("limit", "20"))
	if err != nil {
		limit = 20
	}

	offset, err := strconv.Atoi(c.Query("offset", "0"))
	if err != nil {
		offset = 0
	}

	var categoryPtr *string
	if category := c.Query("category"); category != "" {
		categoryPtr = &category
	}

	customElements, err := h.customElementService.GetPublicCustomElements(c.Context(), categoryPtr, limit, offset)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to retrieve public custom elements")
	}

	return utils.SendJSON(c, fiber.StatusOK, customElements)
}

func (h *CustomElementHandler) GetCustomElementByID(c *fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "id")
	if err != nil {
		return err
	}
	id := ids[0]

	customElement, err := h.customElementService.GetCustomElementByID(c.Context(), id, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return fiber.NewError(fiber.StatusNotFound, "Custom element not found")
		}
		return utils.HandleRepoError(c, err, "Custom element not found", "Failed to retrieve custom element")
	}

	return utils.SendJSON(c, fiber.StatusOK, customElement)
}

func (h *CustomElementHandler) CreateCustomElement(c *fiber.Ctx) error {
	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	var req dto.CreateCustomElementRequest
	if err := utils.ValidateAndParseBody(c, &req); err != nil {
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

	created, err := h.customElementService.CreateCustomElement(c.Context(), customElement)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return fiber.NewError(fiber.StatusConflict, "Custom element with this name already exists")
		}
		log.Println("Error creating custom element:", err)
		return utils.HandleRepoError(c, err, "", "Failed to create custom element")
	}

	return utils.SendJSON(c, fiber.StatusCreated, created)
}

func (h *CustomElementHandler) UpdateCustomElement(c *fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "id")
	if err != nil {
		return err
	}
	id := ids[0]

	var req dto.UpdateCustomElementRequest
	if err := utils.ValidateAndParseBody(c, &req); err != nil {
		return err
	}

	// Build a clean map from the typed DTO, stripping immutable fields.
	updates := map[string]any{}
	if req.Name != nil        { updates["name"] = *req.Name }
	if req.TypeId != nil      { updates["typeId"] = *req.TypeId }
	if req.Description != nil { updates["description"] = *req.Description }
	if req.Category != nil    { updates["category"] = *req.Category }
	if req.Icon != nil        { updates["icon"] = *req.Icon }
	if req.Thumbnail != nil   { updates["thumbnail"] = *req.Thumbnail }
	if req.Tags != nil        { updates["tags"] = *req.Tags }
	if req.IsPublic != nil    { updates["isPublic"] = *req.IsPublic }
	if req.Version != nil     { updates["version"] = *req.Version }
	if req.Structure != nil   { updates["structure"] = req.Structure }
	if req.DefaultProps != nil { updates["defaultProps"] = req.DefaultProps }

	if err := utils.RequireUpdates(updates); err != nil {
		return err
	}

	updated, err := h.customElementService.UpdateCustomElement(c.Context(), id, userID, updates)
	if err != nil {
		if strings.Contains(err.Error(), "unauthorized") {
			return fiber.NewError(fiber.StatusForbidden, "Unauthorized to update this custom element")
		}
		return utils.HandleRepoError(c, err, "Custom element not found", "Failed to update custom element")
	}

	return utils.SendJSON(c, fiber.StatusOK, updated)
}

func (h *CustomElementHandler) DeleteCustomElement(c *fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "id")
	if err != nil {
		return err
	}
	id := ids[0]

	if err := h.customElementService.DeleteCustomElement(c.Context(), id, userID); err != nil {
		if strings.Contains(err.Error(), "unauthorized") {
			return fiber.NewError(fiber.StatusForbidden, "Unauthorized to delete this custom element")
		}
		return utils.HandleRepoError(c, err, "Custom element not found", "Failed to delete custom element")
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Custom element deleted successfully")
}

func (h *CustomElementHandler) DuplicateCustomElement(c *fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "id")
	if err != nil {
		return err
	}
	id := ids[0]

	var req dto.DuplicateCustomElementRequest
	if err := utils.ValidateAndParseBody(c, &req); err != nil {
		return err
	}

	duplicate, err := h.customElementService.DuplicateCustomElement(c.Context(), id, userID, req.NewName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return fiber.NewError(fiber.StatusNotFound, "Custom element not found")
		}
		return utils.HandleRepoError(c, err, "", "Failed to duplicate custom element")
	}

	return utils.SendJSON(c, fiber.StatusCreated, duplicate)
}