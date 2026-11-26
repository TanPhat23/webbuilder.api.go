package handlers

import (
	"my-go-app/internal/dto"
	"my-go-app/internal/models"
	"my-go-app/internal/services"
	"my-go-app/pkg/utils"
	"net/url"
	"strings"

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

func (h *ContentItemHandler) ListByContentType(c fiber.Ctx) error {
	userID, ids, err := utils.RequireUserAndParams(c, "contentTypeId")
	if err != nil {
		return err
	}
	contentTypeID := ids[0]

	contentItems, err := h.contentItemService.ListByContentTypeForUser(c.RequestCtx(), contentTypeID, userID)
	if err != nil {
		return utils.RespondError(c, err, "Content type not found", "Failed to retrieve content items")
	}

	return utils.SendJSON(c, fiber.StatusOK, contentItems)
}

func (h *ContentItemHandler) GetByID(c fiber.Ctx) error {
	userID, ids, err := utils.RequireUserAndParams(c, "itemId")
	if err != nil {
		return err
	}
	itemID := ids[0]

	contentItem, err := h.contentItemService.GetByIDForUser(c.RequestCtx(), itemID, userID)
	if err != nil {
		return utils.RespondError(c, err, "Content item not found", "Failed to retrieve content item")
	}

	return utils.SendJSON(c, fiber.StatusOK, contentItem)
}

func (h *ContentItemHandler) Create(c fiber.Ctx) error {
	userID, ids, err := utils.RequireUserAndParams(c, "contentTypeId")
	if err != nil {
		return err
	}
	contentTypeID := ids[0]

	req := new(dto.CreateContentItemRequest)
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	published := false
	if req.Published != nil {
		published = *req.Published
	}

	contentItem := &models.ContentItem{
		ContentTypeId: contentTypeID,
		Slug:          req.Slug,
		Title:         req.Title,
		Published:     published,
	}

	created, err := h.contentItemService.CreateForUser(c.RequestCtx(), contentItem, req.FieldValues, userID)
	if err != nil {
		return utils.RespondError(c, err, "Content type not found", "Failed to create content item")
	}

	return utils.SendJSON(c, fiber.StatusCreated, created)
}

func (h *ContentItemHandler) Update(c fiber.Ctx) error {
	userID, ids, err := utils.RequireUserAndParams(c, "itemId")
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

	updated, err := h.contentItemService.UpdateForUser(c.RequestCtx(), itemID, userID, columnUpdates, req.FieldValues)
	if err != nil {
		return utils.RespondError(c, err, "Content item not found", "Failed to update content item")
	}

	return utils.SendJSON(c, fiber.StatusOK, updated)
}

func (h *ContentItemHandler) Delete(c fiber.Ctx) error {
	userID, ids, err := utils.RequireUserAndParams(c, "itemId")
	if err != nil {
		return err
	}
	itemID := ids[0]

	if err := h.contentItemService.DeleteForUser(c.RequestCtx(), itemID, userID); err != nil {
		return utils.RespondError(c, err, "Content item not found", "Failed to delete content item")
	}

	return utils.SendNoContent(c)
}

func (h *ContentItemHandler) ListPublic(c fiber.Ctx) error {
	contentTypeId := c.Query("contentTypeId")
	sortOrder := c.Query("sortOrder", "desc")
	sortBy := resolveContentItemSortBy(c.Query("sortBy", "createdAt"))
	limit := utils.ParsePositiveIntQuery(c, "limit", 10)

	contentItems, err := h.contentItemService.ListPublic(c.RequestCtx(), contentTypeId, limit, sortBy, sortOrder)
	if err != nil {
		return utils.RespondError(c, err, "", "Failed to retrieve content items")
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

func (h *ContentItemHandler) GetPublicBySlug(c fiber.Ctx) error {
	ids, err := utils.RequireParams(c, "contentTypeId")
	if err != nil {
		return err
	}
	contentTypeID := ids[0]
	rawSlug := c.Params("slug")
	if rawSlug == "" {
		rawSlug = c.Params("*")
	}
	if decoded, err := url.PathUnescape(rawSlug); err == nil {
		rawSlug = decoded
	}
	slug := strings.Trim(strings.TrimSpace(rawSlug), "/")
	if slug == "" {
		return fiber.NewError(fiber.StatusBadRequest, "slug is required")
	}

	contentItem, err := h.contentItemService.GetBySlug(c.RequestCtx(), contentTypeID, slug)
	if err != nil {
		return utils.RespondError(c, err, "Content item not found", "Failed to retrieve content item")
	}

	if !contentItem.Published {
		return fiber.NewError(fiber.StatusNotFound, "Content item not found")
	}
	if contentItem.ContentType != nil && !contentItem.ContentType.PublicRead {
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
