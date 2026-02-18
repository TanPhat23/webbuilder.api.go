package handlers

import (
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

type ContentTypeHandler struct {
	contentTypeRepository repositories.ContentTypeRepositoryInterface
}

func NewContentTypeHandler(contentTypeRepo repositories.ContentTypeRepositoryInterface) *ContentTypeHandler {
	return &ContentTypeHandler{
		contentTypeRepository: contentTypeRepo,
	}
}

func (h *ContentTypeHandler) GetContentTypes(c *fiber.Ctx) error {
	contentTypes, err := h.contentTypeRepository.GetContentTypes(c.Context())
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve content types", err)
	}
	return utils.SendJSON(c, fiber.StatusOK, contentTypes)
}

func (h *ContentTypeHandler) GetContentTypeByID(c *fiber.Ctx) error {
	id, err := utils.ValidateRequiredParam(c, "id")
	if err != nil {
		return err
	}

	contentType, err := h.contentTypeRepository.GetContentTypeByID(c.Context(), id)
	if err != nil {
		if err.Error() == "content type not found" {
			return utils.SendError(c, fiber.StatusNotFound, "Content type not found", err)
		}
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve content type", err)
	}
	return utils.SendJSON(c, fiber.StatusOK, contentType)
}

func (h *ContentTypeHandler) CreateContentType(c *fiber.Ctx) error {
	var contentType models.ContentType
	if err := utils.ValidateAndParseBody(c, &contentType); err != nil {
		return err
	}

	createdContentType, err := h.contentTypeRepository.CreateContentType(c.Context(), &contentType)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to create content type", err)
	}
	return utils.SendJSON(c, fiber.StatusCreated, createdContentType)
}

func (h *ContentTypeHandler) UpdateContentType(c *fiber.Ctx) error {
	id, err := utils.ValidateRequiredParam(c, "id")
	if err != nil {
		return err
	}

	var updates map[string]any
	if err := utils.ValidateJSONBody(c, &updates); err != nil {
		return err
	}

	columnUpdates := h.buildColumnUpdates(updates)

	updatedContentType, err := h.contentTypeRepository.UpdateContentType(c.Context(), id, columnUpdates)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to update content type", err)
	}
	return utils.SendJSON(c, fiber.StatusOK, updatedContentType)
}

func (h *ContentTypeHandler) DeleteContentType(c *fiber.Ctx) error {
	id, err := utils.ValidateRequiredParam(c, "id")
	if err != nil {
		return err
	}

	if err := h.contentTypeRepository.DeleteContentType(c.Context(), id); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to delete content type", err)
	}
	return utils.SendNoContent(c)
}

func (h *ContentTypeHandler) buildColumnUpdates(updates map[string]any) map[string]any {
	columnUpdates := make(map[string]any)
	for k, v := range updates {
		switch k {
		case "description":
			columnUpdates["Description"] = v
		case "name":
			columnUpdates["Name"] = v
		case "updatedAt":
			columnUpdates["UpdatedAt"] = v
		default:
			columnUpdates[k] = v
		}
	}
	return columnUpdates
}
