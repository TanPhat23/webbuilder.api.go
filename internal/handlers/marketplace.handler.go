package handlers

import (
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
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

// CreateMarketplaceItem creates a new marketplace item
func (h *MarketplaceHandler) CreateMarketplaceItem(c *fiber.Ctx) error {
	userID, ok := c.Locals("userId").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":        "Unauthorized",
			"errorMessage": "You must be logged in to create marketplace items",
		})
	}

	var req models.CreateMarketplaceItemRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Invalid request body",
			"errorMessage": err.Error(),
		})
	}

	// Validate required fields
	if req.Title == "" || req.Description == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Validation failed",
			"errorMessage": "Title and description are required",
		})
	}

	// Get author name from context or use a default
	authorName, ok := c.Locals("userName").(string)
	if !ok || authorName == "" {
		authorName = "Anonymous"
	}

	// Set default template type if not provided
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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to create marketplace item",
			"errorMessage": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(createdItem)
}

// GetMarketplaceItems retrieves marketplace items with filtering
func (h *MarketplaceHandler) GetMarketplaceItems(c *fiber.Ctx) error {
	// Parse query parameters
	filter := repositories.MarketplaceFilter{
		TemplateType: c.Query("templateType"),
		CategoryId:   c.Query("categoryId"),
		TagId:        c.Query("tagId"),
		AuthorId:     c.Query("authorId"),
		Search:       c.Query("search"),
		SortBy:       c.Query("sortBy", "createdAt"),
		SortOrder:    c.Query("sortOrder", "desc"),
	}

	// Parse featured filter
	if featuredStr := c.Query("featured"); featuredStr != "" {
		featured := featuredStr == "true"
		filter.Featured = &featured
	}

	// Parse pagination
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	filter.Limit = limit
	filter.Offset = offset

	items, total, err := h.marketplaceRepository.GetMarketplaceItems(filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to retrieve marketplace items",
			"errorMessage": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data":   items,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetMarketplaceItemByID retrieves a single marketplace item
func (h *MarketplaceHandler) GetMarketplaceItemByID(c *fiber.Ctx) error {
	itemID := c.Params("itemid")
	if itemID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Item ID is required",
			"errorMessage": "Missing itemid parameter in URL",
		})
	}

	item, err := h.marketplaceRepository.GetMarketplaceItemByID(itemID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to retrieve marketplace item",
			"errorMessage": err.Error(),
		})
	}
	if item == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Marketplace item not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(item)
}

// UpdateMarketplaceItem updates a marketplace item
func (h *MarketplaceHandler) UpdateMarketplaceItem(c *fiber.Ctx) error {
	itemID := c.Params("itemid")
	if itemID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Item ID is required",
			"errorMessage": "Missing itemid parameter in URL",
		})
	}

	userID, ok := c.Locals("userId").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":        "Unauthorized",
			"errorMessage": "You must be logged in to update marketplace items",
		})
	}

	var req models.UpdateMarketplaceItemRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Invalid request body",
			"errorMessage": err.Error(),
		})
	}

	// Build updates map
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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to update marketplace item",
			"errorMessage": err.Error(),
		})
	}
	if updatedItem == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Marketplace item not found or you don't have permission to update it",
		})
	}

	return c.Status(fiber.StatusOK).JSON(updatedItem)
}

// DeleteMarketplaceItem deletes a marketplace item
func (h *MarketplaceHandler) DeleteMarketplaceItem(c *fiber.Ctx) error {
	itemID := c.Params("itemid")
	if itemID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Item ID is required",
			"errorMessage": "Missing itemid parameter in URL",
		})
	}

	userID, ok := c.Locals("userId").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":        "Unauthorized",
			"errorMessage": "You must be logged in to delete marketplace items",
		})
	}

	err := h.marketplaceRepository.DeleteMarketplaceItem(itemID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to delete marketplace item",
			"errorMessage": err.Error(),
		})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// DownloadMarketplaceItem downloads a marketplace item by cloning its project
func (h *MarketplaceHandler) DownloadMarketplaceItem(c *fiber.Ctx) error {
	userID, ok := c.Locals("userId").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":        "Unauthorized",
			"errorMessage": "You must be logged in to download marketplace items",
		})
	}

	itemID := c.Params("itemid")
	if itemID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Item ID is required",
			"errorMessage": "Missing itemid parameter in URL",
		})
	}

	project, err := h.marketplaceRepository.DownloadMarketplaceItem(itemID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to download marketplace item",
			"errorMessage": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Marketplace item downloaded successfully",
		"project": project,
	})
}

// IncrementDownloads increments the download count for an item
func (h *MarketplaceHandler) IncrementDownloads(c *fiber.Ctx) error {
	itemID := c.Params("itemid")
	if itemID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Item ID is required",
			"errorMessage": "Missing itemid parameter in URL",
		})
	}

	err := h.marketplaceRepository.IncrementDownloads(itemID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to increment downloads",
			"errorMessage": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Download count incremented",
	})
}

// IncrementLikes increments the like count for an item
func (h *MarketplaceHandler) IncrementLikes(c *fiber.Ctx) error {
	itemID := c.Params("itemid")
	if itemID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Item ID is required",
			"errorMessage": "Missing itemid parameter in URL",
		})
	}

	err := h.marketplaceRepository.IncrementLikes(itemID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to increment likes",
			"errorMessage": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Like count incremented",
	})
}

// CreateCategory creates a new category
func (h *MarketplaceHandler) CreateCategory(c *fiber.Ctx) error {
	userID, ok := c.Locals("userId").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":        "Unauthorized",
			"errorMessage": "You must be logged in to create categories",
		})
	}

	var req models.CreateCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Invalid request body",
			"errorMessage": err.Error(),
		})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Validation failed",
			"errorMessage": "Category name is required",
		})
	}

	category := models.Category{
		Id:   cuid.New(),
		Name: req.Name,
	}

	createdCategory, err := h.marketplaceRepository.CreateCategory(category)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to create category",
			"errorMessage": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(createdCategory)
}

// GetCategories retrieves all categories
func (h *MarketplaceHandler) GetCategories(c *fiber.Ctx) error {
	categories, err := h.marketplaceRepository.GetCategories()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to retrieve categories",
			"errorMessage": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(categories)
}

// DeleteCategory deletes a category
func (h *MarketplaceHandler) DeleteCategory(c *fiber.Ctx) error {
	categoryID := c.Params("categoryid")
	if categoryID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Category ID is required",
			"errorMessage": "Missing categoryid parameter in URL",
		})
	}

	userID, ok := c.Locals("userId").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":        "Unauthorized",
			"errorMessage": "You must be logged in to delete categories",
		})
	}

	err := h.marketplaceRepository.DeleteCategory(categoryID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to delete category",
			"errorMessage": err.Error(),
		})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// CreateTag creates a new tag
func (h *MarketplaceHandler) CreateTag(c *fiber.Ctx) error {
	userID, ok := c.Locals("userId").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":        "Unauthorized",
			"errorMessage": "You must be logged in to create tags",
		})
	}

	var req models.CreateTagRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Invalid request body",
			"errorMessage": err.Error(),
		})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Validation failed",
			"errorMessage": "Tag name is required",
		})
	}

	tag := models.Tag{
		Id:   cuid.New(),
		Name: req.Name,
	}

	createdTag, err := h.marketplaceRepository.CreateTag(tag)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to create tag",
			"errorMessage": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(createdTag)
}

// GetTags retrieves all tags
func (h *MarketplaceHandler) GetTags(c *fiber.Ctx) error {
	tags, err := h.marketplaceRepository.GetTags()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to retrieve tags",
			"errorMessage": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(tags)
}

// DeleteTag deletes a tag
func (h *MarketplaceHandler) DeleteTag(c *fiber.Ctx) error {
	tagID := c.Params("tagid")
	if tagID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Tag ID is required",
			"errorMessage": "Missing tagid parameter in URL",
		})
	}

	userID, ok := c.Locals("userId").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":        "Unauthorized",
			"errorMessage": "You must be logged in to delete tags",
		})
	}

	err := h.marketplaceRepository.DeleteTag(tagID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to delete tag",
			"errorMessage": err.Error(),
		})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}
