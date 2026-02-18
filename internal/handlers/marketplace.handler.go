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

// CreateMarketplaceItem creates a new marketplace item.
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
			return utils.SendError(c, fiber.StatusInternalServerError, "Failed to validate tag", err)
		}
		if tag == nil {
			return utils.SendError(c, fiber.StatusBadRequest, fmt.Sprintf("Tag with ID %s does not exist", tagId), nil)
		}
	}

	for _, categoryId := range req.CategoryIds {
		category, err := h.marketplaceRepository.GetCategoryByID(categoryId)
		if err != nil {
			return utils.SendError(c, fiber.StatusInternalServerError, "Failed to validate category", err)
		}
		if category == nil {
			return utils.SendError(c, fiber.StatusBadRequest, fmt.Sprintf("Category with ID %s does not exist", categoryId), nil)
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
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to create marketplace item", err)
	}

	return utils.SendJSON(c, fiber.StatusCreated, createdItem)
}

// GetMarketplaceItems retrieves marketplace items with filtering and pagination.
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
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve marketplace items", err)
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{
		"data":   items,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetMarketplaceItemByID retrieves a single marketplace item.
func (h *MarketplaceHandler) GetMarketplaceItemByID(c *fiber.Ctx) error {
	itemID, err := utils.ValidateRequiredParam(c, "itemid")
	if err != nil {
		return err
	}

	item, err := h.marketplaceRepository.GetMarketplaceItemByID(itemID)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve marketplace item", err)
	}
	if item == nil {
		return utils.SendError(c, fiber.StatusNotFound, "Marketplace item not found", nil)
	}

	return utils.SendJSON(c, fiber.StatusOK, item)
}

// UpdateMarketplaceItem updates a marketplace item.
func (h *MarketplaceHandler) UpdateMarketplaceItem(c *fiber.Ctx) error {
	itemID, err := utils.ValidateRequiredParam(c, "itemid")
	if err != nil {
		return err
	}

	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	var req models.UpdateMarketplaceItemRequest
	if err := utils.ValidateAndParseBody(c, &req); err != nil {
		return err
	}

	updates := make(map[string]any)
	if req.Title != nil {
		updates["Title"] = *req.Title
	}
	if req.Description != nil {
		updates["Description"] = *req.Description
	}
	if req.Preview != nil {
		updates["Preview"] = *req.Preview
	}
	if req.TemplateType != nil {
		updates["TemplateType"] = *req.TemplateType
	}
	if req.Featured != nil {
		updates["Featured"] = *req.Featured
	}
	if req.PageCount != nil {
		updates["PageCount"] = *req.PageCount
	}
	if req.TagIds != nil {
		updates["TagIds"] = req.TagIds
	}
	if req.CategoryIds != nil {
		updates["CategoryIds"] = req.CategoryIds
	}

	updatedItem, err := h.marketplaceRepository.UpdateMarketplaceItem(itemID, userID, updates)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to update marketplace item", err)
	}
	if updatedItem == nil {
		return utils.SendError(c, fiber.StatusNotFound, "Marketplace item not found or you don't have permission to update it", nil)
	}

	return utils.SendJSON(c, fiber.StatusOK, updatedItem)
}

// DeleteMarketplaceItem deletes a marketplace item.
func (h *MarketplaceHandler) DeleteMarketplaceItem(c *fiber.Ctx) error {
	itemID, err := utils.ValidateRequiredParam(c, "itemid")
	if err != nil {
		return err
	}

	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	if err := h.marketplaceRepository.DeleteMarketplaceItem(itemID, userID); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to delete marketplace item", err)
	}

	return utils.SendNoContent(c)
}

// DownloadMarketplaceItem clones a marketplace item's project for the authenticated user.
func (h *MarketplaceHandler) DownloadMarketplaceItem(c *fiber.Ctx) error {
	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	itemID, err := utils.ValidateRequiredParam(c, "itemid")
	if err != nil {
		return err
	}

	project, err := h.marketplaceRepository.DownloadMarketplaceItem(itemID, userID)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to download marketplace item", err)
	}

	return utils.SendJSON(c, fiber.StatusCreated, fiber.Map{
		"message": "Marketplace item downloaded successfully",
		"project": project,
	})
}

// IncrementDownloads increments the download count for an item.
func (h *MarketplaceHandler) IncrementDownloads(c *fiber.Ctx) error {
	itemID, err := utils.ValidateRequiredParam(c, "itemid")
	if err != nil {
		return err
	}

	if err := h.marketplaceRepository.IncrementDownloads(itemID); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to increment downloads", err)
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Download count incremented")
}

// IncrementLikes increments the like count for an item.
func (h *MarketplaceHandler) IncrementLikes(c *fiber.Ctx) error {
	itemID, err := utils.ValidateRequiredParam(c, "itemid")
	if err != nil {
		return err
	}

	if err := h.marketplaceRepository.IncrementLikes(itemID); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to increment likes", err)
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Like count incremented")
}

// CreateCategory creates a new category.
func (h *MarketplaceHandler) CreateCategory(c *fiber.Ctx) error {
	if _, err := utils.ValidateUserID(c); err != nil {
		return err
	}

	var req models.CreateCategoryRequest
	if err := utils.ValidateAndParseBody(c, &req); err != nil {
		return err
	}

	createdCategory, err := h.marketplaceRepository.CreateCategory(models.Category{
		Id:   cuid.New(),
		Name: req.Name,
	})
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to create category", err)
	}

	return utils.SendJSON(c, fiber.StatusCreated, createdCategory)
}

// GetCategories retrieves all categories.
func (h *MarketplaceHandler) GetCategories(c *fiber.Ctx) error {
	categories, err := h.marketplaceRepository.GetCategories()
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve categories", err)
	}

	return utils.SendJSON(c, fiber.StatusOK, categories)
}

// DeleteCategory deletes a category.
func (h *MarketplaceHandler) DeleteCategory(c *fiber.Ctx) error {
	categoryID, err := utils.ValidateRequiredParam(c, "categoryid")
	if err != nil {
		return err
	}

	if _, err := utils.ValidateUserID(c); err != nil {
		return err
	}

	if err := h.marketplaceRepository.DeleteCategory(categoryID); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to delete category", err)
	}

	return utils.SendNoContent(c)
}

// CreateTag creates a new tag.
func (h *MarketplaceHandler) CreateTag(c *fiber.Ctx) error {
	if _, err := utils.ValidateUserID(c); err != nil {
		return err
	}

	var req models.CreateTagRequest
	if err := utils.ValidateAndParseBody(c, &req); err != nil {
		return err
	}

	createdTag, err := h.marketplaceRepository.CreateTag(models.Tag{
		Id:   cuid.New(),
		Name: req.Name,
	})
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to create tag", err)
	}

	return utils.SendJSON(c, fiber.StatusCreated, createdTag)
}

// GetTags retrieves all tags.
func (h *MarketplaceHandler) GetTags(c *fiber.Ctx) error {
	tags, err := h.marketplaceRepository.GetTags()
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve tags", err)
	}

	return utils.SendJSON(c, fiber.StatusOK, tags)
}

// DeleteTag deletes a tag.
func (h *MarketplaceHandler) DeleteTag(c *fiber.Ctx) error {
	tagID, err := utils.ValidateRequiredParam(c, "tagid")
	if err != nil {
		return err
	}

	if _, err := utils.ValidateUserID(c); err != nil {
		return err
	}

	if err := h.marketplaceRepository.DeleteTag(tagID); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to delete tag", err)
	}

	return utils.SendNoContent(c)
}
