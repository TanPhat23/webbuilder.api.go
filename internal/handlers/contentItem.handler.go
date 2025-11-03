package handlers

import (
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/pkg/utils"
	"strconv"

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
	contentTypeId, err := utils.ValidateRequiredParam(c, "contentTypeId")
	if err != nil {
		return err
	}

	contentItems, err := h.contentItemRepository.GetContentItemsByContentType(c.Context(), contentTypeId)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve content items", err)
	}
	return utils.SendJSON(c, fiber.StatusOK, h.flattenContentItems(contentItems))
}

func (h *ContentItemHandler) GetContentItemByID(c *fiber.Ctx) error {
	id, err := utils.ValidateRequiredParam(c, "itemId")
	if err != nil {
		return err
	}

	contentItem, err := h.contentItemRepository.GetContentItemByID(c.Context(), id)
	if err != nil {
		if err.Error() == "content item not found" {
			return utils.SendError(c, fiber.StatusNotFound, "Content item not found", err)
		}
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve content item", err)
	}
	return utils.SendJSON(c, fiber.StatusOK, h.flattenContentItem(contentItem))
}

func (h *ContentItemHandler) CreateContentItem(c *fiber.Ctx) error {
	contentTypeId, err := utils.ValidateRequiredParam(c, "contentTypeId")
	if err != nil {
		return err
	}

	var contentItem models.ContentItem
	if err := utils.ValidateJSONBody(c, &contentItem); err != nil {
		return err
	}
	fieldValues := contentItem.FieldValues
	contentItem.FieldValues = nil
	contentItem.ContentTypeId = contentTypeId

	createdContentItem, err := h.contentItemRepository.CreateContentItem(c.Context(), &contentItem, fieldValues)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to create content item", err)
	}
	return utils.SendJSON(c, fiber.StatusCreated, h.flattenContentItem(createdContentItem))
}

func (h *ContentItemHandler) UpdateContentItem(c *fiber.Ctx) error {
	id, err := utils.ValidateRequiredParam(c, "itemId")
	if err != nil {
		return err
	}

	var updates map[string]any
	if err := utils.ValidateJSONBody(c, &updates); err != nil {
		return err
	}

	// Extract fieldValues before processing column updates
	var fieldValues []models.ContentFieldValue
	if fvData, ok := updates["fieldValues"]; ok {
		if fvSlice, ok := fvData.([]interface{}); ok {
			for _, fv := range fvSlice {
				if fvMap, ok := fv.(map[string]interface{}); ok {
					fieldID, fidOK := fvMap["fieldId"].(string)
					value, valOK := fvMap["value"].(string)

					if fidOK && valOK {
						fieldValues = append(fieldValues, models.ContentFieldValue{
							FieldId: fieldID,
							Value:   &value,
						})
					}
				}
			}
		}
		// Remove fieldValues from updates map so it's not processed as a column
		delete(updates, "fieldValues")
	}

	// Build column updates
	columnUpdates := h.buildColumnUpdates(updates)

	// Pass fieldValues separately to the repository
	updatedContentItem, err := h.contentItemRepository.UpdateContentItem(c.Context(), id, columnUpdates, fieldValues)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to update content item", err)
	}
	return utils.SendJSON(c, fiber.StatusOK, h.flattenContentItem(updatedContentItem))
}

func (h *ContentItemHandler) DeleteContentItem(c *fiber.Ctx) error {
</parameter>
</invoke>
	id, err := utils.ValidateRequiredParam(c, "itemId")
	if err != nil {
		return err
	}

	err = h.contentItemRepository.DeleteContentItem(c.Context(), id)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to delete content item", err)
	}
	return utils.SendNoContent(c)
}

func (h *ContentItemHandler) GetPublicContentItems(c *fiber.Ctx) error {
	contentTypeId := c.Query("contentTypeId")
	limitStr := c.Query("limit")
	sortBy := c.Query("sortBy", "createdAt")
	sortOrder := c.Query("sortOrder", "desc")

	sortByMap := map[string]string{
		"createdAt": "CreatedAt",
		"updatedAt": "UpdatedAt",
		"title":     "Title",
	}
	if mapped, ok := sortByMap[sortBy]; ok {
		sortBy = mapped
	} else {
		sortBy = "CreatedAt"
	}

	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	contentItems, err := h.contentItemRepository.GetPublicContentItems(c.Context(), contentTypeId, limit, sortBy, sortOrder)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve content items", err)
	}
	return utils.SendJSON(c, fiber.StatusOK, h.flattenContentItems(contentItems))
}

func (h *ContentItemHandler) GetPublicContentItemBySlug(c *fiber.Ctx) error {
	contentTypeId, err := utils.ValidateRequiredParam(c, "contentTypeId")
	if err != nil {
		return err
	}
	slug, err := utils.ValidateRequiredParam(c, "slug")
	if err != nil {
		return err
	}

	contentItem, err := h.contentItemRepository.GetContentItemBySlug(c.Context(), contentTypeId, slug)
	if err != nil {
		if err.Error() == "content item not found" {
			return utils.SendError(c, fiber.StatusNotFound, "Content item not found", err)
		}
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve content item", err)
	}
	if !contentItem.Published {
		return utils.SendError(c, fiber.StatusNotFound, "Content item not found", nil)
	}
	return utils.SendJSON(c, fiber.StatusOK, h.flattenContentItem(contentItem))
}

func (h *ContentItemHandler) buildColumnUpdates(updates map[string]any) map[string]any {
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
	return columnUpdates
}

func (h *ContentItemHandler) flattenContentItem(item *models.ContentItem) map[string]interface{} {
	flattened := map[string]interface{}{
		"contentTypeId": item.ContentTypeId,
		"createdAt":     item.CreatedAt,
		"id":            item.Id,
		"published":     item.Published,
		"slug":          item.Slug,
		"title":         item.Title,
		"updatedAt":     item.UpdatedAt,
		"contentType":   item.ContentType,
	}
	for _, fv := range item.FieldValues {
		if fv.Field.Name != "" {
			flattened[fv.Field.Name] = fv.Value
		}
	}
	return flattened
}

func (h *ContentItemHandler) flattenContentItems(items []models.ContentItem) []map[string]interface{} {
	var flattened []map[string]interface{}
	for _, item := range items {
		flattened = append(flattened, h.flattenContentItem(&item))
	}
	return flattened
}
