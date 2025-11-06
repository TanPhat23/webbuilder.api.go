package handlers

import (
	"log"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/pkg/utils"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ElementCommentHandler struct {
	elementCommentRepo repositories.ElementCommentRepositoryInterface
}

func NewElementCommentHandler(elementCommentRepo repositories.ElementCommentRepositoryInterface) *ElementCommentHandler {
	return &ElementCommentHandler{
		elementCommentRepo: elementCommentRepo,
	}
}

// CreateElementComment creates a new element comment
// POST /element-comments
func (h *ElementCommentHandler) CreateElementComment(c *fiber.Ctx) error {
	var req models.CreateElementCommentRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	if req.Content == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "Validation failed", fiber.NewError(fiber.StatusBadRequest, "Content is required"))
	}

	if req.ElementId == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "Validation failed", fiber.NewError(fiber.StatusBadRequest, "ElementId is required"))
	}

	// Get user ID from context (assuming it's set by middleware)
	userID, ok := c.Locals("userId").(string)
	if !ok || userID == "" {
		return utils.SendError(c, fiber.StatusUnauthorized, "User ID not found", nil)
	}

	comment := &models.ElementComment{
		Id:        uuid.NewString(),
		Content:   req.Content,
		ElementId: req.ElementId,
		AuthorId:  userID,
	}

	createdComment, err := h.elementCommentRepo.CreateElementComment(c.Context(), comment)
	if err != nil {
		log.Printf("Error creating element comment: %v\n", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to create comment", err)
	}

	return utils.SendJSON(c, fiber.StatusCreated, createdComment)
}

// GetElementCommentByID retrieves a single element comment by ID
// GET /element-comments/:id
func (h *ElementCommentHandler) GetElementCommentByID(c *fiber.Ctx) error {
	commentID := c.Params("id")
	if commentID == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "Comment ID is required", nil)
	}

	comment, err := h.elementCommentRepo.GetElementCommentByID(c.Context(), commentID)
	if err != nil {
		log.Printf("Error retrieving element comment: %v\n", err)
		return utils.SendError(c, fiber.StatusNotFound, "Comment not found", err)
	}

	return utils.SendJSON(c, fiber.StatusOK, comment)
}

// GetElementComments retrieves comments for an element with filtering and pagination
// GET /elements/:elementId/comments
func (h *ElementCommentHandler) GetElementComments(c *fiber.Ctx) error {
	elementID, err := utils.ValidateRequiredParam(c, "elementId")
	if err != nil {
		return err
	}

	// Parse query parameters for filtering and pagination
	filter := &models.ElementCommentFilter{
		Limit:     20,
		Offset:    0,
		SortBy:    "CreatedAt",
		SortOrder: "DESC",
	}

	// Parse limit
	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			filter.Limit = l
		}
	}

	// Parse offset
	if offset := c.Query("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil && o >= 0 {
			filter.Offset = o
		}
	}

	// Parse author filter
	if authorID := c.Query("authorId"); authorID != "" {
		filter.AuthorId = authorID
	}

	// Parse resolved filter
	if resolved := c.Query("resolved"); resolved != "" {
		if r, err := strconv.ParseBool(resolved); err == nil {
			filter.Resolved = &r
		}
	}

	// Parse sort options
	if sortBy := c.Query("sortBy"); sortBy != "" {
		filter.SortBy = sortBy
	}

	if sortOrder := c.Query("sortOrder"); sortOrder == "ASC" || sortOrder == "DESC" {
		filter.SortOrder = sortOrder
	}

	comments, err := h.elementCommentRepo.GetElementComments(c.Context(), elementID, filter)
	if err != nil {
		log.Printf("Error retrieving element comments: %v\n", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve comments", err)
	}

	// Get count for pagination
	count, err := h.elementCommentRepo.CountElementComments(c.Context(), elementID)
	if err != nil {
		log.Printf("Error counting comments: %v\n", err)
		count = 0
	}

	return c.JSON(fiber.Map{
		"data":   comments,
		"total":  count,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// UpdateElementComment updates an existing element comment
// PATCH /element-comments/:id
func (h *ElementCommentHandler) UpdateElementComment(c *fiber.Ctx) error {
	commentID := c.Params("id")
	if commentID == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "Comment ID is required", nil)
	}

	var req models.UpdateElementCommentRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	// Get user ID from context (for authorization check)
	userID, ok := c.Locals("userId").(string)
	if !ok || userID == "" {
		return utils.SendError(c, fiber.StatusUnauthorized, "User ID not found", nil)
	}

	// Retrieve current comment to verify ownership
	comment, err := h.elementCommentRepo.GetElementCommentByID(c.Context(), commentID)
	if err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Comment not found", err)
	}

	if comment.AuthorId != userID {
		return utils.SendError(c, fiber.StatusForbidden, "You can only update your own comments", nil)
	}

	// Build updates map
	updates := make(map[string]any)
	if req.Content != nil {
		updates["Content"] = *req.Content
	}
	if req.Resolved != nil {
		updates["Resolved"] = *req.Resolved
	}

	if len(updates) == 0 {
		return utils.SendError(c, fiber.StatusBadRequest, "No fields to update", nil)
	}

	updatedComment, err := h.elementCommentRepo.UpdateElementComment(c.Context(), commentID, updates)
	if err != nil {
		log.Printf("Error updating element comment: %v\n", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to update comment", err)
	}

	return utils.SendJSON(c, fiber.StatusOK, updatedComment)
}

// DeleteElementComment deletes an element comment
// DELETE /element-comments/:id
func (h *ElementCommentHandler) DeleteElementComment(c *fiber.Ctx) error {
	commentID := c.Params("id")
	if commentID == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "Comment ID is required", nil)
	}

	// Get user ID from context (for authorization check)
	userID, ok := c.Locals("userId").(string)
	if !ok || userID == "" {
		return utils.SendError(c, fiber.StatusUnauthorized, "User ID not found", nil)
	}

	// Retrieve current comment to verify ownership
	comment, err := h.elementCommentRepo.GetElementCommentByID(c.Context(), commentID)
	if err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Comment not found", err)
	}

	if comment.AuthorId != userID {
		return utils.SendError(c, fiber.StatusForbidden, "You can only delete your own comments", nil)
	}

	err = h.elementCommentRepo.DeleteElementComment(c.Context(), commentID)
	if err != nil {
		log.Printf("Error deleting element comment: %v\n", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to delete comment", err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ToggleResolvedStatus toggles the resolved status of a comment
// PATCH /element-comments/:id/toggle-resolved
func (h *ElementCommentHandler) ToggleResolvedStatus(c *fiber.Ctx) error {
	commentID := c.Params("id")
	if commentID == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "Comment ID is required", nil)
	}

	comment, err := h.elementCommentRepo.ToggleResolvedStatus(c.Context(), commentID)
	if err != nil {
		log.Printf("Error toggling comment resolved status: %v\n", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to toggle resolved status", err)
	}

	return utils.SendJSON(c, fiber.StatusOK, comment)
}

// GetCommentsByAuthorID retrieves all comments by a specific author
// GET /element-comments/author/:authorId
func (h *ElementCommentHandler) GetCommentsByAuthorID(c *fiber.Ctx) error {
	authorID, err := utils.ValidateRequiredParam(c, "authorId")
	if err != nil {
		return err
	}

	// Parse pagination parameters
	limit := 20
	offset := 0

	if l := c.Query("limit"); l != "" {
		if parsedL, err := strconv.Atoi(l); err == nil && parsedL > 0 {
			limit = parsedL
		}
	}

	if o := c.Query("offset"); o != "" {
		if parsedO, err := strconv.Atoi(o); err == nil && parsedO >= 0 {
			offset = parsedO
		}
	}

	comments, err := h.elementCommentRepo.GetElementCommentsByAuthorID(c.Context(), authorID, limit, offset)
	if err != nil {
		log.Printf("Error retrieving comments by author: %v\n", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve comments", err)
	}

	return utils.SendJSON(c, fiber.StatusOK, comments)
}
