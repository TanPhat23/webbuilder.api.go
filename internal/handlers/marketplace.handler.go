package handlers

import (
	"fmt"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/pkg/utils"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/lucsky/cuid"
)

type MarketplaceHandler struct {
	marketplaceRepository repositories.MarketplaceRepositoryInterface
}

func NewMarketplaceHandler(marketplaceRepo repositories.MarketplaceRepositoryInterface) *MarketplaceHandler {
	return &MarketplaceHandler{
		marketplaceRepository: marketplaceRepo,
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

	for _, tagId := range req.TagIds {
		tag, err := h.marketplaceRepository.GetTagByID(tagId)
		if err != nil {
			return utils.HandleRepoError(c, err, "", "Failed to validate tag")
		}
		if tag == nil {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Tag with ID %s does not exist", tagId))
		}
	}

	for _, categoryId := range req.CategoryIds {
		category, err := h.marketplaceRepository.GetCategoryByID(categoryId)
		if err != nil {
			return utils.HandleRepoError(c, err, "", "Failed to validate category")
		}
		if category == nil {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Category with ID %s does not exist", categoryId))
		}
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

	createdItem, err := h.marketplaceRepository.CreateMarketplaceItem(item, req.TagIds, req.CategoryIds)
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

	items, total, err := h.marketplaceRepository.GetMarketplaceItems(filter)
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

	item, err := h.marketplaceRepository.GetMarketplaceItemByID(itemID)
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

	updates := make(map[string]any)
	if req.Title != nil        { updates["Title"] = *req.Title }
	if req.Description != nil  { updates["Description"] = *req.Description }
	if req.Preview != nil      { updates["Preview"] = *req.Preview }
	if req.TemplateType != nil { updates["TemplateType"] = *req.TemplateType }
	if req.Featured != nil     { updates["Featured"] = *req.Featured }
	if req.PageCount != nil    { updates["PageCount"] = *req.PageCount }
	if req.TagIds != nil       { updates["TagIds"] = req.TagIds }
	if req.CategoryIds != nil  { updates["CategoryIds"] = req.CategoryIds }

	if err := utils.RequireUpdates(updates); err != nil {
		return err
	}

	updated, err := h.marketplaceRepository.UpdateMarketplaceItem(itemID, userID, updates)
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

	if err := h.marketplaceRepository.DeleteMarketplaceItem(itemID, userID); err != nil {
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

	project, err := h.marketplaceRepository.DownloadMarketplaceItem(itemID, userID)
	if err != nil {
		return utils.HandleRepoError(c, err, "Marketplace item not found", "Failed to download marketplace item")
	}

	return utils.SendJSON(c, fiber.StatusCreated, fiber.Map{
		"message": "Marketplace item downloaded successfully",
		"project": project,
	})
}

func (h *MarketplaceHandler) IncrementLikes(c *fiber.Ctx) error {
	ids, err := utils.MustParams(c, "itemid")
	if err != nil {
		return err
	}
	itemID := ids[0]

	if err := h.marketplaceRepository.IncrementLikes(itemID); err != nil {
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

	created, err := h.marketplaceRepository.CreateCategory(models.Category{
		Id:   cuid.New(),
		Name: req.Name,
	})
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to create category")
	}

	return utils.SendJSON(c, fiber.StatusCreated, created)
}

func (h *MarketplaceHandler) GetCategories(c *fiber.Ctx) error {
	categories, err := h.marketplaceRepository.GetCategories()
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

	if err := h.marketplaceRepository.DeleteCategory(categoryID); err != nil {
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

	created, err := h.marketplaceRepository.CreateTag(models.Tag{
		Id:   cuid.New(),
		Name: req.Name,
	})
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to create tag")
	}

	return utils.SendJSON(c, fiber.StatusCreated, created)
}

func (h *MarketplaceHandler) GetTags(c *fiber.Ctx) error {
	tags, err := h.marketplaceRepository.GetTags()
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

	if err := h.marketplaceRepository.DeleteTag(tagID); err != nil {
		return utils.HandleRepoError(c, err, "Tag not found", "Failed to delete tag")
	}

	return utils.SendNoContent(c)
}