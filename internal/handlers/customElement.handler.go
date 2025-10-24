package handlers

import (
	"encoding/json"
	"log"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/pkg/utils"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type CustomElementHandler struct {
	customElementRepo repositories.CustomElementRepositoryInterface
}

func NewCustomElementHandler(customElementRepo repositories.CustomElementRepositoryInterface) *CustomElementHandler {
	return &CustomElementHandler{
		customElementRepo: customElementRepo,
	}
}

func (h *CustomElementHandler) GetCustomElements(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)
	if userID == "" {
		return utils.SendError(c, fiber.StatusUnauthorized, "User not authenticated", nil)
	}

	isPublicStr := c.Query("isPublic")

	var isPublicPtr *bool
	if isPublicStr != "" {
		isPublic, err := strconv.ParseBool(isPublicStr)
		if err == nil {
			isPublicPtr = &isPublic
		}
	}

	customElements, err := h.customElementRepo.GetCustomElements(c.Context(), userID, isPublicPtr)
	if err != nil {
		log.Println("Error retrieving custom elements:", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve custom elements", err)
	}

	return utils.SendJSON(c, fiber.StatusOK, customElements)
}

func (h *CustomElementHandler) GetCustomElementByID(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)
	if userID == "" {
		return utils.SendError(c, fiber.StatusUnauthorized, "User not authenticated", nil)
	}

	id, err := utils.ValidateRequiredParam(c, "id")
	if err != nil {
		return err
	}

	customElement, err := h.customElementRepo.GetCustomElementByID(c.Context(), id, userID)
	if err != nil {
		if err == repositories.ErrCustomElementNotFound {
			return utils.SendError(c, fiber.StatusNotFound, "Custom element not found", err)
		}
		log.Println("Error retrieving custom element:", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve custom element", err)
	}

	return utils.SendJSON(c, fiber.StatusOK, customElement)
}

func (h *CustomElementHandler) CreateCustomElement(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)
	if userID == "" {
		return utils.SendError(c, fiber.StatusUnauthorized, "User not authenticated", nil)
	}

	var req struct {
		Name         string          `json:"name"`
		TypeId       *string         `json:"typeId"`
		Description  *string         `json:"description"`
		Category     *string         `json:"category"`
		Icon         *string         `json:"icon"`
		Thumbnail    *string         `json:"thumbnail"`
		Structure    json.RawMessage `json:"structure"`
		DefaultProps json.RawMessage `json:"defaultProps"`
		Tags         *string         `json:"tags"`
		IsPublic     bool            `json:"isPublic"`
		Version      string          `json:"version"`
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	if req.Name == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "Name is required", nil)
	}

	if req.Structure == nil || len(req.Structure) == 0 {
		return utils.SendError(c, fiber.StatusBadRequest, "Structure is required", nil)
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

	created, err := h.customElementRepo.CreateCustomElement(c.Context(), customElement)
	if err != nil {
		if err == repositories.ErrCustomElementAlreadyExists {
			return utils.SendError(c, fiber.StatusConflict, "Custom element with this name already exists", err)
		}
		log.Println("Error creating custom element:", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to create custom element", err)
	}

	return utils.SendJSON(c, fiber.StatusCreated, created)
}

func (h *CustomElementHandler) UpdateCustomElement(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)
	if userID == "" {
		return utils.SendError(c, fiber.StatusUnauthorized, "User not authenticated", nil)
	}

	id, err := utils.ValidateRequiredParam(c, "id")
	if err != nil {
		return err
	}

	var req map[string]any
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	delete(req, "id")
	delete(req, "userId")
	delete(req, "createdAt")

	updated, err := h.customElementRepo.UpdateCustomElement(c.Context(), id, userID, req)
	if err != nil {
		if err == repositories.ErrCustomElementUnauthorized {
			return utils.SendError(c, fiber.StatusForbidden, "Unauthorized to update this custom element", err)
		}
		log.Println("Error updating custom element:", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to update custom element", err)
	}

	return utils.SendJSON(c, fiber.StatusOK, updated)
}

func (h *CustomElementHandler) DeleteCustomElement(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)
	if userID == "" {
		return utils.SendError(c, fiber.StatusUnauthorized, "User not authenticated", nil)
	}

	id, err := utils.ValidateRequiredParam(c, "id")
	if err != nil {
		return err
	}

	err = h.customElementRepo.DeleteCustomElement(c.Context(), id, userID)
	if err != nil {
		if err == repositories.ErrCustomElementUnauthorized {
			return utils.SendError(c, fiber.StatusForbidden, "Unauthorized to delete this custom element", err)
		}
		log.Println("Error deleting custom element:", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to delete custom element", err)
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{
		"message": "Custom element deleted successfully",
	})
}

func (h *CustomElementHandler) GetPublicCustomElements(c *fiber.Ctx) error {
	category := c.Query("category")
	limitStr := c.Query("limit", "20")
	offsetStr := c.Query("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	var categoryPtr *string
	if category != "" {
		categoryPtr = &category
	}

	customElements, err := h.customElementRepo.GetPublicCustomElements(c.Context(), categoryPtr, limit, offset)
	if err != nil {
		log.Println("Error retrieving public custom elements:", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve public custom elements", err)
	}

	return utils.SendJSON(c, fiber.StatusOK, customElements)
}

func (h *CustomElementHandler) DuplicateCustomElement(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)
	if userID == "" {
		return utils.SendError(c, fiber.StatusUnauthorized, "User not authenticated", nil)
	}

	id, err := utils.ValidateRequiredParam(c, "id")
	if err != nil {
		return err
	}

	var req struct {
		NewName string `json:"newName"`
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	if req.NewName == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "New name is required", nil)
	}

	duplicate, err := h.customElementRepo.DuplicateCustomElement(c.Context(), id, userID, req.NewName)
	if err != nil {
		if err == repositories.ErrCustomElementNotFound {
			return utils.SendError(c, fiber.StatusNotFound, "Custom element not found", err)
		}
		log.Println("Error duplicating custom element:", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to duplicate custom element", err)
	}

	return utils.SendJSON(c, fiber.StatusCreated, duplicate)
}
