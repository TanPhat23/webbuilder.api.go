package handlers

import (
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

type ContentFieldHandler struct {
	contentFieldRepository repositories.ContentFieldRepositoryInterface
}

func NewContentFieldHandler(contentFieldRepo repositories.ContentFieldRepositoryInterface) *ContentFieldHandler {
	return &ContentFieldHandler{
		contentFieldRepository: contentFieldRepo,
	}
}

func (h *ContentFieldHandler) GetContentFieldsByContentType(c *fiber.Ctx) error {
	contentTypeId, err := utils.ValidateRequiredParam(c, "contentTypeId")
	if err != nil {
		return err
	}

	contentFields, err := h.contentFieldRepository.GetContentFieldsByContentType(c.Context(), contentTypeId)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve content fields", err)
	}
	return utils.SendJSON(c, fiber.StatusOK, contentFields)
}

func (h *ContentFieldHandler) GetContentFieldByID(c *fiber.Ctx) error {
	id, err := utils.ValidateRequiredParam(c, "fieldId")
	if err != nil {
		return err
	}

	contentField, err := h.contentFieldRepository.GetContentFieldByID(c.Context(), id)
	if err != nil {
		if err.Error() == "content field not found" {
			return utils.SendError(c, fiber.StatusNotFound, "Content field not found", err)
		}
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve content field", err)
	}
	return utils.SendJSON(c, fiber.StatusOK, contentField)
}

func (h *ContentFieldHandler) CreateContentField(c *fiber.Ctx) error {
	contentTypeId, err := utils.ValidateRequiredParam(c, "contentTypeId")
	if err != nil {
		return err
	}

	var contentField models.ContentField
	if err := utils.ValidateJSONBody(c, &contentField); err != nil {
		return err
	}
	contentField.ContentTypeId = contentTypeId

	createdContentField, err := h.contentFieldRepository.CreateContentField(c.Context(), &contentField)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to create content field", err)
	}
	return utils.SendJSON(c, fiber.StatusCreated, createdContentField)
}

func (h *ContentFieldHandler) UpdateContentField(c *fiber.Ctx) error {
	id, err := utils.ValidateRequiredParam(c, "fieldId")
	if err != nil {
		return err
	}

	var updates map[string]any
	if err := utils.ValidateJSONBody(c, &updates); err != nil {
		return err
	}

	columnUpdates := h.buildColumnUpdates(updates)

	updatedContentField, err := h.contentFieldRepository.UpdateContentField(c.Context(), id, columnUpdates)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to update content field", err)
	}
	return utils.SendJSON(c, fiber.StatusOK, updatedContentField)
}

func (h *ContentFieldHandler) DeleteContentField(c *fiber.Ctx) error {
	id, err := utils.ValidateRequiredParam(c, "fieldId")
	if err != nil {
		return err
	}

	err = h.contentFieldRepository.DeleteContentField(c.Context(), id)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to delete content field", err)
	}
	return utils.SendNoContent(c)
}

func (h *ContentFieldHandler) buildColumnUpdates(updates map[string]any) map[string]any {
	columnUpdates := make(map[string]any)
	for k, v := range updates {
		switch k {
		case "name":
			columnUpdates["Name"] = v
		case "required":
			columnUpdates["Required"] = v
		case "type":
			columnUpdates["Type"] = v
		default:
			columnUpdates[k] = v
		}
	}
	return columnUpdates
}
