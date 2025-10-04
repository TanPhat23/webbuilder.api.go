package handlers

import (
	"encoding/json"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"

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
	contentTypeId := c.Params("contentTypeId")
	if contentTypeId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Content type ID is required",
			"errorMessage": "Missing contentTypeId parameter in URL",
		})
	}

	contentFields, err := h.contentFieldRepository.GetContentFieldsByContentType(contentTypeId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to retrieve content fields",
			"errorMessage": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(contentFields)
}

func (h *ContentFieldHandler) GetContentFieldByID(c *fiber.Ctx) error {
	id := c.Params("fieldId")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Content field ID is required",
			"errorMessage": "Missing fieldId parameter in URL",
		})
	}

	contentField, err := h.contentFieldRepository.GetContentFieldByID(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to retrieve content field",
			"errorMessage": err.Error(),
		})
	}
	if contentField == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Content field not found",
		})
	}
	return c.Status(fiber.StatusOK).JSON(contentField)
}

func (h *ContentFieldHandler) CreateContentField(c *fiber.Ctx) error {
	contentTypeId := c.Params("contentTypeId")
	if contentTypeId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Content type ID is required",
			"errorMessage": "Missing contentTypeId parameter in URL",
		})
	}

	var contentField models.ContentField
	if err := json.Unmarshal(c.Body(), &contentField); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Invalid JSON body",
			"errorMessage": err.Error(),
		})
	}
	contentField.ContentTypeId = contentTypeId

	createdContentField, err := h.contentFieldRepository.CreateContentField(contentField)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to create content field",
			"errorMessage": err.Error(),
		})
	}
	return c.Status(fiber.StatusCreated).JSON(createdContentField)
}

func (h *ContentFieldHandler) UpdateContentField(c *fiber.Ctx) error {
	id := c.Params("fieldId")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Content field ID is required",
			"errorMessage": "Missing fieldId parameter in URL",
		})
	}

	var updates map[string]any
	if err := json.Unmarshal(c.Body(), &updates); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Invalid JSON body",
			"errorMessage": err.Error(),
		})
	}

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

	updatedContentField, err := h.contentFieldRepository.UpdateContentField(id, columnUpdates)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to update content field",
			"errorMessage": err.Error(),
		})
	}
	if updatedContentField == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Content field not found",
		})
	}
	return c.Status(fiber.StatusOK).JSON(updatedContentField)
}

func (h *ContentFieldHandler) DeleteContentField(c *fiber.Ctx) error {
	id := c.Params("fieldId")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Content field ID is required",
			"errorMessage": "Missing fieldId parameter in URL",
		})
	}

	err := h.contentFieldRepository.DeleteContentField(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to delete content field",
			"errorMessage": err.Error(),
		})
	}
	return c.Status(fiber.StatusNoContent).Send(nil)
}
