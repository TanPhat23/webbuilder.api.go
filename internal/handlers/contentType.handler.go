package handlers

import (
	"encoding/json"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"

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
	contentTypes, err := h.contentTypeRepository.GetContentTypes()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to retrieve content types",
			"errorMessage": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(contentTypes)
}

func (h *ContentTypeHandler) GetContentTypeByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Content type ID is required",
			"errorMessage": "Missing id parameter in URL",
		})
	}

	contentType, err := h.contentTypeRepository.GetContentTypeByID(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to retrieve content type",
			"errorMessage": err.Error(),
		})
	}
	if contentType == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Content type not found",
		})
	}
	return c.Status(fiber.StatusOK).JSON(contentType)
}

func (h *ContentTypeHandler) CreateContentType(c *fiber.Ctx) error {
	var contentType models.ContentType
	if err := json.Unmarshal(c.Body(), &contentType); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Invalid JSON body",
			"errorMessage": err.Error(),
		})
	}

	createdContentType, err := h.contentTypeRepository.CreateContentType(contentType)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to create content type",
			"errorMessage": err.Error(),
		})
	}
	return c.Status(fiber.StatusCreated).JSON(createdContentType)
}

func (h *ContentTypeHandler) UpdateContentType(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Content type ID is required",
			"errorMessage": "Missing id parameter in URL",
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

	updatedContentType, err := h.contentTypeRepository.UpdateContentType(id, columnUpdates)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to update content type",
			"errorMessage": err.Error(),
		})
	}
	if updatedContentType == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Content type not found",
		})
	}
	return c.Status(fiber.StatusOK).JSON(updatedContentType)
}

func (h *ContentTypeHandler) DeleteContentType(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Content type ID is required",
			"errorMessage": "Missing id parameter in URL",
		})
	}

	err := h.contentTypeRepository.DeleteContentType(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to delete content type",
			"errorMessage": err.Error(),
		})
	}
	return c.Status(fiber.StatusNoContent).Send(nil)
}
