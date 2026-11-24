package handlers

import (
	"my-go-app/internal/dto"
	"my-go-app/internal/models"
	"my-go-app/internal/services"
	"my-go-app/pkg/utils"

	"github.com/gofiber/fiber/v3"
)

var contentItemAllowedCols = map[string]string{
	"published": "Published",
	"slug":      "Slug",
	"title":     "Title",
	"updatedAt": "UpdatedAt",
}

type ContentItemHandler struct {
	contentItemService services.ContentItemServiceInterface
}

func NewContentItemHandler(contentItemService services.ContentItemServiceInterface) *ContentItemHandler {
	return &ContentItemHandler{
		contentItemService: contentItemService,
	}
}

func (h *ContentItemHandler) GetContentItemsByContentType(c fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "contentTypeId")
	if err != nil {
		return err
	}
	contentTypeID := ids[0]

	contentItems, err := h.contentItemService.GetContentItemsByContentTypeForUser(c.RequestCtx(), contentTypeID, userID)
	if err != nil {
		return utils.HandleRepoError(c, err, "Content type not found", "Failed to retrieve content items")
	}

	return utils.SendJSON(c, fiber.StatusOK, contentItems)
}

func (h *ContentItemHandler) GetContentItemByID(c fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "itemId")
	if err != nil {
		return err
	}
	itemID := ids[0]

	contentItem, err := h.contentItemService.GetContentItemByIDForUser(c.RequestCtx(), itemID, userID)
	if err != nil {
		return utils.HandleRepoError(c, err, "Content item not found", "Failed to retrieve content item")
	}

	return utils.SendJSON(c, fiber.StatusOK, contentItem)
}

func (h *ContentItemHandler) CreateContentItem(c fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "contentTypeId")
	if err != nil {
		return err
	}
	contentTypeID := ids[0]

	contentItem := new(models.ContentItem)
	if err := c.Bind().Body(contentItem); err != nil {
		return err
	}

	fieldValues := contentItem.FieldValues
	contentItem.FieldValues = nil
	contentItem.ContentTypeId = contentTypeID

	created, err := h.contentItemService.CreateContentItemForUser(c.RequestCtx(), contentItem, fieldValues, userID)
	if err != nil {
		return utils.HandleRepoError(c, err, "Content type not found", "Failed to create content item")
	}

	return utils.SendJSON(c, fiber.StatusCreated, created)
}

func (h *ContentItemHandler) UpdateContentItem(c fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "itemId")
	if err != nil {
		return err
	}
	itemID := ids[0]

	req := new(dto.UpdateContentItemRequest)
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	rawBody := map[string]any{}
	utils.SetIfNotNil(rawBody, "published", req.Published)
	utils.SetIfNotNil(rawBody, "slug", req.Slug)
	utils.SetIfNotNil(rawBody, "title", req.Title)

	columnUpdates, err := utils.BuildColumnUpdates(rawBody, contentItemAllowedCols)
	if err != nil {
		return err
	}

	if len(req.FieldValues) == 0 {
		if err := utils.RequireUpdates(columnUpdates); err != nil {
			return err
		}
	}

	updated, err := h.contentItemService.UpdateContentItemForUser(c.RequestCtx(), itemID, userID, columnUpdates, req.FieldValues)
	if err != nil {
		return utils.HandleRepoError(c, err, "Content item not found", "Failed to update content item")
	}

	return utils.SendJSON(c, fiber.StatusOK, updated)
}

func (h *ContentItemHandler) DeleteContentItem(c fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "itemId")
	if err != nil {
		return err
	}
	itemID := ids[0]

	if err := h.contentItemService.DeleteContentItemForUser(c.RequestCtx(), itemID, userID); err != nil {
		return utils.HandleRepoError(c, err, "Content item not found", "Failed to delete content item")
	}

	return utils.SendNoContent(c)
}

func (h *ContentItemHandler) GetPublicContentItems(c fiber.Ctx) error {
	contentTypeId := c.Query("contentTypeId")
	sortOrder := c.Query("sortOrder", "desc")
	sortBy := resolveContentItemSortBy(c.Query("sortBy", "createdAt"))
	limit := utils.ParsePositiveIntQuery(c, "limit", 10)

	contentItems, err := h.contentItemService.GetPublicContentItems(c.RequestCtx(), contentTypeId, limit, sortBy, sortOrder)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to retrieve content items")
	}

	return utils.SendJSON(c, fiber.StatusOK, h.flattenContentItems(contentItems))
}

func resolveContentItemSortBy(raw string) string {
	switch raw {
	case "updatedAt":
		return "UpdatedAt"
	case "title":
		return "Title"
	default:
		return "CreatedAt"
	}
}

func (h *ContentItemHandler) GetPublicContentItemBySlug(c fiber.Ctx) error {
	ids, err := utils.MustParams(c, "contentTypeId", "slug")
	if err != nil {
		return err
	}
	contentTypeID, slug := ids[0], ids[1]

	contentItem, err := h.contentItemService.GetContentItemBySlug(c.RequestCtx(), contentTypeID, slug)
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

// fiber:context-methods migrated
