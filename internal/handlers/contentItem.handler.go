package handlers

import (
	"my-go-app/internal/dto"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/pkg/utils"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

var contentItemAllowedCols = map[string]string{
	"published": "Published",
	"slug":      "Slug",
	"title":     "Title",
	"updatedAt": "UpdatedAt",
}

type ContentItemHandler struct {
	contentItemRepository repositories.ContentItemRepositoryInterface
}

func NewContentItemHandler(contentItemRepo repositories.ContentItemRepositoryInterface) *ContentItemHandler {
	return &ContentItemHandler{
		contentItemRepository: contentItemRepo,
	}
}

func (h *ContentItemHandler) GetContentItemsByContentType(c *fiber.Ctx) error {
	ids, err := utils.MustParams(c, "contentTypeId")
	if err != nil {
		return err
	}
	contentTypeID := ids[0]

	contentItems, err := h.contentItemRepository.GetContentItemsByContentType(c.Context(), contentTypeID)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to retrieve content items")
	}

	return utils.SendJSON(c, fiber.StatusOK, contentItems)
}

func (h *ContentItemHandler) GetContentItemByID(c *fiber.Ctx) error {
	ids, err := utils.MustParams(c, "itemId")
	if err != nil {
		return err
	}
	itemID := ids[0]

	contentItem, err := h.contentItemRepository.GetContentItemByID(c.Context(), itemID)
	if err != nil {
		return utils.HandleRepoError(c, err, "Content item not found", "Failed to retrieve content item")
	}

	return utils.SendJSON(c, fiber.StatusOK, contentItem)
}

func (h *ContentItemHandler) CreateContentItem(c *fiber.Ctx) error {
	ids, err := utils.MustParams(c, "contentTypeId")
	if err != nil {
		return err
	}
	contentTypeID := ids[0]

	var contentItem models.ContentItem
	if err := utils.ValidateJSONBody(c, &contentItem); err != nil {
		return err
	}

	fieldValues := contentItem.FieldValues
	contentItem.FieldValues = nil
	contentItem.ContentTypeId = contentTypeID

	created, err := h.contentItemRepository.CreateContentItem(c.Context(), &contentItem, fieldValues)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to create content item")
	}

	return utils.SendJSON(c, fiber.StatusCreated, created)
}

func (h *ContentItemHandler) UpdateContentItem(c *fiber.Ctx) error {
	ids, err := utils.MustParams(c, "itemId")
	if err != nil {
		return err
	}
	itemID := ids[0]

	var req dto.UpdateContentItemRequest
	if err := utils.ValidateAndParseBody(c, &req); err != nil {
		return err
	}

	rawBody := map[string]any{}
	if req.Published != nil { rawBody["published"] = *req.Published }
	if req.Slug != nil      { rawBody["slug"] = *req.Slug }
	if req.Title != nil     { rawBody["title"] = *req.Title }

	columnUpdates, err := utils.BuildColumnUpdates(rawBody, contentItemAllowedCols)
	if err != nil {
		return err
	}

	// Prevent no-op PATCHes: if there are no column updates and no field-value updates,
	// mirror other handlers by requiring at least one update.
	hasFieldValueUpdates := req.FieldValues != nil && len(req.FieldValues) > 0
	if !hasFieldValueUpdates {
		if err := utils.RequireUpdates(columnUpdates); err != nil {
			return err
		}
	}

	updated, err := h.contentItemRepository.UpdateContentItem(c.Context(), itemID, columnUpdates, req.FieldValues)
	if err != nil {
		return utils.HandleRepoError(c, err, "Content item not found", "Failed to update content item")
	}

	return utils.SendJSON(c, fiber.StatusOK, updated)
}

func (h *ContentItemHandler) DeleteContentItem(c *fiber.Ctx) error {
	ids, err := utils.MustParams(c, "itemId")
	if err != nil {
		return err
	}
	itemID := ids[0]

	if err := h.contentItemRepository.DeleteContentItem(c.Context(), itemID); err != nil {
		return utils.HandleRepoError(c, err, "Content item not found", "Failed to delete content item")
	}

	return utils.SendNoContent(c)
}

func (h *ContentItemHandler) GetPublicContentItems(c *fiber.Ctx) error {
	contentTypeId := c.Query("contentTypeId")
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
	if l, err := strconv.Atoi(c.Query("limit", "10")); err == nil && l > 0 {
		limit = l
	}

	contentItems, err := h.contentItemRepository.GetPublicContentItems(c.Context(), contentTypeId, limit, sortBy, sortOrder)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to retrieve content items")
	}

	return utils.SendJSON(c, fiber.StatusOK, h.flattenContentItems(contentItems))
}

func (h *ContentItemHandler) GetPublicContentItemBySlug(c *fiber.Ctx) error {
	ids, err := utils.MustParams(c, "contentTypeId", "slug")
	if err != nil {
		return err
	}
	contentTypeID, slug := ids[0], ids[1]

	contentItem, err := h.contentItemRepository.GetContentItemBySlug(c.Context(), contentTypeID, slug)
	if err != nil {
		return utils.HandleRepoError(c, err, "Content item not found", "Failed to retrieve content item")
	}

	if !contentItem.Published {
		return fiber.NewError(fiber.StatusNotFound, "Content item not found")
	}

	return utils.SendJSON(c, fiber.StatusOK, h.flattenContentItem(contentItem))
}

func (h *ContentItemHandler) flattenContentItem(item *models.ContentItem) map[string]any {
	flattened := map[string]any{
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

func (h *ContentItemHandler) flattenContentItems(items []models.ContentItem) []map[string]any {
	flattened := make([]map[string]any, 0, len(items))
	for _, item := range items {
		flattened = append(flattened, h.flattenContentItem(&item))
	}
	return flattened
}