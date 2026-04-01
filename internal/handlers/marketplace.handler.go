package handlers

import (
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"
	"my-go-app/pkg/utils"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/lucsky/cuid"
)

type MarketplaceHandler struct {
	marketplaceService services.MarketplaceServiceInterface
}

func NewMarketplaceHandler(marketplaceService services.MarketplaceServiceInterface) *MarketplaceHandler {
	return &MarketplaceHandler{
		marketplaceService: marketplaceService,
	}
}

func (h *MarketplaceHandler) CreateMarketplaceItem(c *fiber.Ctx) error {
	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	var req models.CreateMarketplaceItemRequest
	if err := utils.ValidateAndParseBody(c, &req); err != nil {
		return err
	}

	authorName, ok := c.Locals("userName").(string)
	if !ok || authorName == "" {
		authorName = "Anonymous"
	}

	templateType := "block"
	if req.TemplateType != "" {
		templateType = req.TemplateType
	}

	now := time.Now()
	item := models.MarketplaceItem{
		Id:           cuid.New(),
		Title:        req.Title,
		Description:  req.Description,
		Preview:      req.Preview,
		TemplateType: templateType,
		Featured:     false,
		PageCount:    req.PageCount,
		Downloads:    0,
		Likes:        0,
		AuthorId:     userID,
		AuthorName:   authorName,
		Verified:     false,
		ProjectId:    req.ProjectId,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	createdItem, err := h.marketplaceService.CreateMarketplaceItem(c.Context(), item, req.TagIds, req.CategoryIds)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to create marketplace item")
	}

	return utils.SendJSON(c, fiber.StatusCreated, createdItem)
}

func (h *MarketplaceHandler) GetMarketplaceItems(c *fiber.Ctx) error {
	filter := repositories.MarketplaceFilter{
		TemplateType: c.Query("templateType"),
		CategoryId:   c.Query("categoryId"),
		TagId:        c.Query("tagId"),
		AuthorId:     c.Query("authorId"),
		Search:       c.Query("search"),
		SortBy:       c.Query("sortBy", "createdAt"),
		SortOrder:    c.Query("sortOrder", "desc"),
	}

	if featuredStr := c.Query("featured"); featuredStr != "" {
		featured := featuredStr == "true"
		filter.Featured = &featured
	}

	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	filter.Limit = limit
	filter.Offset = offset

	items, total, err := h.marketplaceService.GetMarketplaceItems(c.Context(), filter)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to retrieve marketplace items")
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{
		"data":   items,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *MarketplaceHandler) GetMarketplaceItemByID(c *fiber.Ctx) error {
	ids, err := utils.MustParams(c, "itemid")
	if err != nil {
		return err
	}
	itemID := ids[0]

	item, err := h.marketplaceService.GetMarketplaceItemByID(c.Context(), itemID)
	if err != nil {
		return utils.HandleRepoError(c, err, "Marketplace item not found", "Failed to retrieve marketplace item")
	}
	if item == nil {
		return fiber.NewError(fiber.StatusNotFound, "Marketplace item not found")
	}

	return utils.SendJSON(c, fiber.StatusOK, item)
}

func (h *MarketplaceHandler) UpdateMarketplaceItem(c *fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "itemid")
	if err != nil {
		return err
	}
	itemID := ids[0]

	var req models.UpdateMarketplaceItemRequest
	if err := utils.ValidateAndParseBody(c, &req); err != nil {
		return err
	}

	updateItem := &models.MarketplaceItem{
		Id:          itemID,
		Title:       "",
		Description: "",
		Preview:     nil,
		TemplateType: "",
		Featured:    false,
		PageCount:   nil,
		ProjectId:   nil,
	}

	if req.Title != nil {
		updateItem.Title = *req.Title
	}
	if req.Description != nil {
		updateItem.Description = *req.Description
	}
	if req.Preview != nil {
		updateItem.Preview = req.Preview
	}
	if req.TemplateType != nil {
		updateItem.TemplateType = *req.TemplateType
	}
	if req.Featured != nil {
		updateItem.Featured = *req.Featured
	}
	if req.PageCount != nil {
		updateItem.PageCount = req.PageCount
	}

	updated, err := h.marketplaceService.UpdateMarketplaceItem(c.Context(), itemID, updateItem, userID)
	if err != nil {
		return utils.HandleRepoError(c, err, "Marketplace item not found", "Failed to update marketplace item")
	}
	if updated == nil {
		return fiber.NewError(fiber.StatusNotFound, "Marketplace item not found or you don't have permission to update it")
	}

	return utils.SendJSON(c, fiber.StatusOK, updated)
}

func (h *MarketplaceHandler) DeleteMarketplaceItem(c *fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "itemid")
	if err != nil {
		return err
	}
	itemID := ids[0]

	if err := h.marketplaceService.DeleteMarketplaceItem(c.Context(), itemID, userID); err != nil {
		return utils.HandleRepoError(c, err, "Marketplace item not found", "Failed to delete marketplace item")
	}

	return utils.SendNoContent(c)
}

func (h *MarketplaceHandler) DownloadMarketplaceItem(c *fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "itemid")
	if err != nil {
		return err
	}
	itemID := ids[0]

	if err := h.marketplaceService.DownloadMarketplaceItem(c.Context(), itemID, userID); err != nil {
		return utils.HandleRepoError(c, err, "Marketplace item not found", "Failed to download marketplace item")
	}

	return utils.SendJSON(c, fiber.StatusCreated, fiber.Map{
		"message": "Marketplace item downloaded successfully",
	})
}

func (h *MarketplaceHandler) IncrementLikes(c *fiber.Ctx) error {
	ids, err := utils.MustParams(c, "itemid")
	if err != nil {
		return err
	}
	itemID := ids[0]

	if err := h.marketplaceService.IncrementLikes(c.Context(), itemID); err != nil {
		return utils.HandleRepoError(c, err, "Marketplace item not found", "Failed to increment likes")
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Like count incremented")
}

func (h *MarketplaceHandler) CreateCategory(c *fiber.Ctx) error {
	if _, err := utils.ValidateUserID(c); err != nil {
		return err
	}

	var req models.CreateCategoryRequest
	if err := utils.ValidateAndParseBody(c, &req); err != nil {
		return err
	}

	created, err := h.marketplaceService.CreateCategory(c.Context(), &models.Category{
		Id:   cuid.New(),
		Name: req.Name,
	})
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to create category")
	}

	return utils.SendJSON(c, fiber.StatusCreated, created)
}

func (h *MarketplaceHandler) GetCategories(c *fiber.Ctx) error {
	categories, err := h.marketplaceService.GetCategories(c.Context())
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to retrieve categories")
	}

	return utils.SendJSON(c, fiber.StatusOK, categories)
}

func (h *MarketplaceHandler) DeleteCategory(c *fiber.Ctx) error {
	_, ids, err := utils.MustUserAndParams(c, "categoryid")
	if err != nil {
		return err
	}
	categoryID := ids[0]

	if err := h.marketplaceService.DeleteCategory(c.Context(), categoryID); err != nil {
		return utils.HandleRepoError(c, err, "Category not found", "Failed to delete category")
	}

	return utils.SendNoContent(c)
}

func (h *MarketplaceHandler) CreateTag(c *fiber.Ctx) error {
	if _, err := utils.ValidateUserID(c); err != nil {
		return err
	}

	var req models.CreateTagRequest
	if err := utils.ValidateAndParseBody(c, &req); err != nil {
		return err
	}

	created, err := h.marketplaceService.CreateTag(c.Context(), &models.Tag{
		Id:   cuid.New(),
		Name: req.Name,
	})
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to create tag")
	}

	return utils.SendJSON(c, fiber.StatusCreated, created)
}

func (h *MarketplaceHandler) GetTags(c *fiber.Ctx) error {
	tags, err := h.marketplaceService.GetTags(c.Context())
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to retrieve tags")
	}

	return utils.SendJSON(c, fiber.StatusOK, tags)
}

func (h *MarketplaceHandler) DeleteTag(c *fiber.Ctx) error {
	_, ids, err := utils.MustUserAndParams(c, "tagid")
	if err != nil {
		return err
	}
	tagID := ids[0]

	if err := h.marketplaceService.DeleteTag(c.Context(), tagID); err != nil {
		return utils.HandleRepoError(c, err, "Tag not found", "Failed to delete tag")
	}

	return utils.SendNoContent(c)
}