package handlers

import (
	"encoding/json"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"

	"github.com/gofiber/fiber/v2"
)

type ContentItemHandler struct {
	contentItemRepository repositories.ContentItemRepositoryInterface
}

func NewContentItemHandler(contentItemRepo repositories.ContentItemRepositoryInterface) *ContentItemHandler {
	return &ContentItemHandler{
		contentItemRepository: contentItemRepo,
	}
}

func (h *ContentItemHandler) GetContentItemsByContentType(c *fiber.Ctx) error {
	contentTypeId := c.Params("contentTypeId")
	if contentTypeId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Content type ID is required",
			"errorMessage": "Missing contentTypeId parameter in URL",
		})
	}

	contentItems, err := h.contentItemRepository.GetContentItemsByContentType(contentTypeId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to retrieve content items",
			"errorMessage": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(contentItems)
}

func (h *ContentItemHandler) GetContentItemByID(c *fiber.Ctx) error {
	id := c.Params("itemId")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Content item ID is required",
			"errorMessage": "Missing itemId parameter in URL",
		})
	}

	contentItem, err := h.contentItemRepository.GetContentItemByID(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to retrieve content item",
			"errorMessage": err.Error(),
		})
	}
	if contentItem == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Content item not found",
		})
	}
	return c.Status(fiber.StatusOK).JSON(contentItem)
}

func (h *ContentItemHandler) CreateContentItem(c *fiber.Ctx) error {
	contentTypeId := c.Params("contentTypeId")
	if contentTypeId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Content type ID is required",
			"errorMessage": "Missing contentTypeId parameter in URL",
		})
	}

	var contentItem models.ContentItem
	if err := json.Unmarshal(c.Body(), &contentItem); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Invalid JSON body",
			"errorMessage": err.Error(),
		})
	}
	fieldValues := contentItem.FieldValues
	contentItem.FieldValues = nil
	contentItem.ContentTypeId = contentTypeId

	createdContentItem, err := h.contentItemRepository.CreateContentItem(contentItem, fieldValues)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to create content item",
			"errorMessage": err.Error(),
		})
	}
	return c.Status(fiber.StatusCreated).JSON(createdContentItem)
}

func (h *ContentItemHandler) UpdateContentItem(c *fiber.Ctx) error {
	id := c.Params("itemId")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Content item ID is required",
			"errorMessage": "Missing itemId parameter in URL",
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
		case "published":
			columnUpdates["Published"] = v
		case "slug":
			columnUpdates["Slug"] = v
		case "title":
			columnUpdates["Title"] = v
		case "updatedAt":
			columnUpdates["UpdatedAt"] = v
		default:
			columnUpdates[k] = v
		}
	}

	updatedContentItem, err := h.contentItemRepository.UpdateContentItem(id, columnUpdates)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to update content item",
			"errorMessage": err.Error(),
		})
	}
	if updatedContentItem == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Content item not found",
		})
	}
	return c.Status(fiber.StatusOK).JSON(updatedContentItem)
}

func (h *ContentItemHandler) DeleteContentItem(c *fiber.Ctx) error {
	id := c.Params("itemId")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Content item ID is required",
			"errorMessage": "Missing itemId parameter in URL",
		})
	}

	err := h.contentItemRepository.DeleteContentItem(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to delete content item",
			"errorMessage": err.Error(),
		})
	}
	return c.Status(fiber.StatusNoContent).Send(nil)
}

func (h *ContentItemHandler) GetPublicContentItemBySlug(c *fiber.Ctx) error {
	contentTypeId := c.Params("contentTypeId")
	slug := c.Params("slug")
	if contentTypeId == "" || slug == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Content type ID and slug are required",
			"errorMessage": "Missing contentTypeId or slug parameter in URL",
		})
	}

	contentItem, err := h.contentItemRepository.GetContentItemBySlug(contentTypeId, slug)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to retrieve content item",
			"errorMessage": err.Error(),
		})
	}
	if contentItem == nil || !contentItem.Published {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Content item not found",
		})
	}
	return c.Status(fiber.StatusOK).JSON(contentItem)
}
