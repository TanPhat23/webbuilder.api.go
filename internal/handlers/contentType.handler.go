package handlers

import (
	"my-go-app/internal/dto"
	"my-go-app/internal/models"
	"my-go-app/internal/services"
	"my-go-app/pkg/utils"

	"github.com/gofiber/fiber/v3"
)

var contentTypeAllowedCols = map[string]string{
	"name":        "Name",
	"description": "Description",
}

type ContentTypeHandler struct {
	contentTypeService services.ContentTypeServiceInterface
}

func NewContentTypeHandler(contentTypeService services.ContentTypeServiceInterface) *ContentTypeHandler {
	return &ContentTypeHandler{
		contentTypeService: contentTypeService,
	}
}

func (h *ContentTypeHandler) GetContentTypes(c fiber.Ctx) error {
	contentTypes, err := h.contentTypeService.GetContentTypes(c.RequestCtx())
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to retrieve content types")
	}

	return utils.SendJSON(c, fiber.StatusOK, contentTypes)
}

func (h *ContentTypeHandler) GetContentTypeByID(c fiber.Ctx) error {
	ids, err := utils.MustParams(c, "id")
	if err != nil {
		return err
	}
	id := ids[0]

	contentType, err := h.contentTypeService.GetContentTypeByID(c.RequestCtx(), id)
	if err != nil {
		return utils.HandleRepoError(c, err, "Content type not found", "Failed to retrieve content type")
	}

	return utils.SendJSON(c, fiber.StatusOK, contentType)
}

func (h *ContentTypeHandler) CreateContentType(c fiber.Ctx) error {
	req := new(dto.CreateContentTypeRequest)
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	contentType := &models.ContentType{
		Name:        req.Name,
		Description: req.Description,
	}

	created, err := h.contentTypeService.CreateContentType(c.RequestCtx(), contentType)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to create content type")
	}

	return utils.SendJSON(c, fiber.StatusCreated, created)
}

func (h *ContentTypeHandler) UpdateContentType(c fiber.Ctx) error {
	ids, err := utils.MustParams(c, "id")
	if err != nil {
		return err
	}
	id := ids[0]

	req := new(dto.UpdateContentTypeRequest)
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	if req.Name == nil && req.Description == nil {
		return utils.SendJSON(c, fiber.StatusBadRequest, fiber.Map{"error": "At least one field must be updated"})
	}

	contentType := &models.ContentType{}
	if req.Name != nil {
		contentType.Name = *req.Name
	}
	if req.Description != nil {
		contentType.Description = req.Description
	}

	updated, err := h.contentTypeService.UpdateContentType(c.RequestCtx(), id, contentType)
	if err != nil {
		return utils.HandleRepoError(c, err, "Content type not found", "Failed to update content type")
	}

	return utils.SendJSON(c, fiber.StatusOK, updated)
}

func (h *ContentTypeHandler) DeleteContentType(c fiber.Ctx) error {
	ids, err := utils.MustParams(c, "id")
	if err != nil {
		return err
	}
	id := ids[0]

	if err := h.contentTypeService.DeleteContentType(c.RequestCtx(), id); err != nil {
		return utils.HandleRepoError(c, err, "Content type not found", "Failed to delete content type")
	}

	return utils.SendNoContent(c)
}

