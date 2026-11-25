package handlers

import (
	"my-go-app/internal/dto"
	"my-go-app/internal/models"
	"my-go-app/internal/services"
	"my-go-app/pkg/utils"

	"github.com/gofiber/fiber/v3"
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

func (h *ContentFieldHandler) ListByContentType(c fiber.Ctx) error {
	userID, ids, err := utils.RequireUserAndParams(c, "contentTypeId")
	if err != nil {
		return err
	}
	contentTypeID := ids[0]

	contentFields, err := h.contentFieldService.ListByContentTypeForUser(c.RequestCtx(), contentTypeID, userID)
	if err != nil {
		return utils.RespondError(c, err, "Content type not found", "Failed to retrieve content fields")
	}

	return utils.SendJSON(c, fiber.StatusOK, contentFields)
}

func (h *ContentFieldHandler) GetByID(c fiber.Ctx) error {
	userID, ids, err := utils.RequireUserAndParams(c, "fieldId")
	if err != nil {
		return err
	}
	fieldID := ids[0]

	contentField, err := h.contentFieldService.GetByIDForUser(c.RequestCtx(), fieldID, userID)
	if err != nil {
		return utils.RespondError(c, err, "Content field not found", "Failed to retrieve content field")
	}

	return utils.SendJSON(c, fiber.StatusOK, contentField)
}

func (h *ContentFieldHandler) Create(c fiber.Ctx) error {
	userID, ids, err := utils.RequireUserAndParams(c, "contentTypeId")
	if err != nil {
		return err
	}
	contentTypeID := ids[0]

	req := new(dto.CreateContentFieldRequest)
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	contentField := &models.ContentField{
		Name:          req.Name,
		Type:          req.Type,
		Required:      req.Required,
		ContentTypeId: contentTypeID,
	}

	created, err := h.contentFieldService.CreateForUser(c.RequestCtx(), contentField, userID)
	if err != nil {
		return utils.RespondError(c, err, "Content type not found", "Failed to create content field")
	}

	return utils.SendJSON(c, fiber.StatusCreated, created)
}

func (h *ContentFieldHandler) Update(c fiber.Ctx) error {
	userID, ids, err := utils.RequireUserAndParams(c, "fieldId")
	if err != nil {
		return err
	}
	fieldID := ids[0]

	req := new(dto.UpdateContentFieldRequest)
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	rawBody := map[string]any{}
	if req.Name != nil {
		rawBody["name"] = *req.Name
	}
	if req.Type != nil {
		rawBody["type"] = *req.Type
	}
	if req.Required != nil {
		rawBody["required"] = *req.Required
	}

	updates, err := utils.BuildColumnUpdates(rawBody, contentFieldAllowedCols)
	if err != nil {
		return err
	}
	if err := utils.RequireUpdates(updates); err != nil {
		return err
	}

	updated, err := h.contentFieldService.UpdateForUser(c.RequestCtx(), fieldID, userID, updates)
	if err != nil {
		return utils.RespondError(c, err, "Content field not found", "Failed to update content field")
	}

	return utils.SendJSON(c, fiber.StatusOK, updated)
}

func (h *ContentFieldHandler) Delete(c fiber.Ctx) error {
	userID, ids, err := utils.RequireUserAndParams(c, "fieldId")
	if err != nil {
		return err
	}
	fieldID := ids[0]

	if err := h.contentFieldService.DeleteForUser(c.RequestCtx(), fieldID, userID); err != nil {
		return utils.RespondError(c, err, "Content field not found", "Failed to delete content field")
	}

	return utils.SendNoContent(c)
}

// fiber:context-methods migrated
