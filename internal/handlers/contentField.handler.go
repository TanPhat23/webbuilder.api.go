package handlers

import (
	"my-go-app/internal/dto"
	"my-go-app/internal/models"
	"my-go-app/internal/services"
	"my-go-app/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

var contentFieldAllowedCols = map[string]string{
	"name":     "Name",
	"type":     "Type",
	"required": "Required",
}

type ContentFieldHandler struct {
	contentFieldService services.ContentFieldServiceInterface
}

func NewContentFieldHandler(contentFieldService services.ContentFieldServiceInterface) *ContentFieldHandler {
	return &ContentFieldHandler{
		contentFieldService: contentFieldService,
	}
}

func (h *ContentFieldHandler) GetContentFieldsByContentType(c *fiber.Ctx) error {
	ids, err := utils.MustParams(c, "contentTypeId")
	if err != nil {
		return err
	}
	contentTypeID := ids[0]

	contentFields, err := h.contentFieldService.GetContentFieldsByContentType(c.Context(), contentTypeID)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to retrieve content fields")
	}

	return utils.SendJSON(c, fiber.StatusOK, contentFields)
}

func (h *ContentFieldHandler) GetContentFieldByID(c *fiber.Ctx) error {
	ids, err := utils.MustParams(c, "fieldId")
	if err != nil {
		return err
	}
	fieldID := ids[0]

	contentField, err := h.contentFieldService.GetContentFieldByID(c.Context(), fieldID)
	if err != nil {
		return utils.HandleRepoError(c, err, "Content field not found", "Failed to retrieve content field")
	}

	return utils.SendJSON(c, fiber.StatusOK, contentField)
}

func (h *ContentFieldHandler) CreateContentField(c *fiber.Ctx) error {
	ids, err := utils.MustParams(c, "contentTypeId")
	if err != nil {
		return err
	}
	contentTypeID := ids[0]

	var req dto.CreateContentFieldRequest
	if err := utils.ValidateAndParseBody(c, &req); err != nil {
		return err
	}

	contentField := &models.ContentField{
		Name:          req.Name,
		Type:          req.Type,
		Required:      req.Required,
		ContentTypeId: contentTypeID,
	}

	created, err := h.contentFieldService.CreateContentField(c.Context(), contentField)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to create content field")
	}

	return utils.SendJSON(c, fiber.StatusCreated, created)
}

func (h *ContentFieldHandler) UpdateContentField(c *fiber.Ctx) error {
	ids, err := utils.MustParams(c, "fieldId")
	if err != nil {
		return err
	}
	fieldID := ids[0]

	var req dto.UpdateContentFieldRequest
	if err := utils.ValidateAndParseBody(c, &req); err != nil {
		return err
	}

	rawBody := map[string]any{}
	if req.Name != nil     { rawBody["name"] = *req.Name }
	if req.Type != nil     { rawBody["type"] = *req.Type }
	if req.Required != nil { rawBody["required"] = *req.Required }

	updates, err := utils.BuildColumnUpdates(rawBody, contentFieldAllowedCols)
	if err != nil {
		return err
	}
	if err := utils.RequireUpdates(updates); err != nil {
		return err
	}

	updated, err := h.contentFieldService.UpdateContentField(c.Context(), fieldID, updates)
	if err != nil {
		return utils.HandleRepoError(c, err, "Content field not found", "Failed to update content field")
	}

	return utils.SendJSON(c, fiber.StatusOK, updated)
}

func (h *ContentFieldHandler) DeleteContentField(c *fiber.Ctx) error {
	ids, err := utils.MustParams(c, "fieldId")
	if err != nil {
		return err
	}
	fieldID := ids[0]

	if err := h.contentFieldService.DeleteContentField(c.Context(), fieldID); err != nil {
		return utils.HandleRepoError(c, err, "Content field not found", "Failed to delete content field")
	}

	return utils.SendNoContent(c)
}