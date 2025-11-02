package handlers

import (
	"log"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/pkg/utils"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/lucsky/cuid"
)

type CommentHandler struct {
	commentRepository     repositories.CommentRepositoryInterface
	marketplaceRepository repositories.MarketplaceRepositoryInterface
}

func NewCommentHandler(commentRepo repositories.CommentRepositoryInterface, marketplaceRepo repositories.MarketplaceRepositoryInterface) *CommentHandler {
	return &CommentHandler{
		commentRepository:     commentRepo,
		marketplaceRepository: marketplaceRepo,
	}
}

// CreateComment creates a new comment on a marketplace item
func (h *CommentHandler) CreateComment(c *fiber.Ctx) error {
	userID, ok := c.Locals("userId").(string)
	if !ok || userID == "" {
		return utils.SendError(c, fiber.StatusUnauthorized, "Unauthorized", fiber.NewError(fiber.StatusUnauthorized, "You must be logged in to create comments"))
	}

	var req models.CreateCommentRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	// Validate required fields
	if req.Content == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "Validation failed", fiber.NewError(fiber.StatusBadRequest, "Content is required"))
	}

	if req.ItemId == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "Validation failed", fiber.NewError(fiber.StatusBadRequest, "ItemId is required"))
	}

	// Verify marketplace item exists
	item, err := h.marketplaceRepository.GetMarketplaceItemByID(req.ItemId)
	if err != nil {
		log.Println("Error validating marketplace item:", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to validate marketplace item", err)
	}
	if item == nil {
		return utils.SendError(c, fiber.StatusNotFound, "Marketplace item not found", fiber.NewError(fiber.StatusNotFound, "The marketplace item does not exist"))
	}

	// If ParentId is provided, verify parent comment exists
	if req.ParentId != nil && *req.ParentId != "" {
		parentComment, err := h.commentRepository.GetCommentByID(*req.ParentId)
		if err != nil {
			log.Println("Error validating parent comment:", err)
			return utils.SendError(c, fiber.StatusInternalServerError, "Failed to validate parent comment", err)
		}
		if parentComment == nil {
			return utils.SendError(c, fiber.StatusNotFound, "Parent comment not found", fiber.NewError(fiber.StatusNotFound, "The parent comment does not exist"))
		}
		// Verify parent comment belongs to the same item
		if parentComment.ItemId != req.ItemId {
			return utils.SendError(c, fiber.StatusBadRequest, "Validation failed", fiber.NewError(fiber.StatusBadRequest, "Parent comment does not belong to the specified item"))
		}
	}

	now := time.Now()
	comment := models.Comment{
		Id:        cuid.New(),
		Content:   req.Content,
		AuthorId:  userID,
		ItemId:    req.ItemId,
		ParentId:  req.ParentId,
		Status:    "published",
		Edited:    false,
		CreatedAt: now,
		UpdatedAt: now,
	}

	createdComment, err := h.commentRepository.CreateComment(comment)
	if err != nil {
		log.Println("Error creating comment:", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to create comment", err)
	}

	return utils.SendJSON(c, fiber.StatusCreated, createdComment)
}

// GetComments retrieves comments with filtering
func (h *CommentHandler) GetComments(c *fiber.Ctx) error {
	// Parse query parameters
	filter := models.CommentFilter{
		ItemId:    c.Query("itemId"),
		AuthorId:  c.Query("authorId"),
		Status:    c.Query("status", "published"),
		SortBy:    c.Query("sortBy", "createdAt"),
		SortOrder: c.Query("sortOrder", "desc"),
	}

	// Parse parentId filter
	if parentIdStr := c.Query("parentId"); parentIdStr != "" {
		filter.ParentId = &parentIdStr
	} else if c.Query("topLevel") == "true" {
		emptyStr := ""
		filter.ParentId = &emptyStr
	}

	// Parse pagination
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	filter.Limit = limit
	filter.Offset = offset

	comments, total, err := h.commentRepository.GetComments(filter)
	if err != nil {
		log.Println("Error retrieving comments:", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve comments", err)
	}

	response := fiber.Map{
		"data":   comments,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	}
	return utils.SendJSON(c, fiber.StatusOK, response)
}

// GetCommentByID retrieves a single comment
func (h *CommentHandler) GetCommentByID(c *fiber.Ctx) error {
	commentID, err := utils.ValidateRequiredParam(c, "commentid")
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Comment ID is required", err)
	}

	comment, err := h.commentRepository.GetCommentByID(commentID)
	if err != nil {
		log.Println("Error retrieving comment:", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve comment", err)
	}
	if comment == nil {
		return utils.SendError(c, fiber.StatusNotFound, "Comment not found", fiber.NewError(fiber.StatusNotFound, "Comment not found"))
	}

	return utils.SendJSON(c, fiber.StatusOK, comment)
}

// GetCommentsByItemID retrieves all comments for a marketplace item
func (h *CommentHandler) GetCommentsByItemID(c *fiber.Ctx) error {
	itemID, err := utils.ValidateRequiredParam(c, "itemid")
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Item ID is required", err)
	}

	// Parse query parameters
	filter := models.CommentFilter{
		ItemId:    itemID,
		Status:    c.Query("status", "published"),
		SortBy:    c.Query("sortBy", "createdAt"),
		SortOrder: c.Query("sortOrder", "desc"),
	}

	// Only get top-level comments by default
	if c.Query("includeReplies") != "false" {
		emptyStr := ""
		filter.ParentId = &emptyStr
	}

	// Parse pagination
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	filter.Limit = limit
	filter.Offset = offset

	comments, total, err := h.commentRepository.GetComments(filter)
	if err != nil {
		log.Println("Error retrieving comments:", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve comments", err)
	}

	response := fiber.Map{
		"data":   comments,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	}
	return utils.SendJSON(c, fiber.StatusOK, response)
}

// UpdateComment updates a comment
func (h *CommentHandler) UpdateComment(c *fiber.Ctx) error {
	commentID, err := utils.ValidateRequiredParam(c, "commentid")
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Comment ID is required", err)
	}

	userID, ok := c.Locals("userId").(string)
	if !ok || userID == "" {
		return utils.SendError(c, fiber.StatusUnauthorized, "Unauthorized", fiber.NewError(fiber.StatusUnauthorized, "You must be logged in to update comments"))
	}

	var req models.UpdateCommentRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	// Build updates map
	updates := make(map[string]any)
	if req.Content != nil {
		if *req.Content == "" {
			return utils.SendError(c, fiber.StatusBadRequest, "Validation failed", fiber.NewError(fiber.StatusBadRequest, "Content cannot be empty"))
		}
		updates["Content"] = *req.Content
	}
	if req.Status != nil {
		updates["Status"] = *req.Status
	}

	if len(updates) == 0 {
		return utils.SendError(c, fiber.StatusBadRequest, "Validation failed", fiber.NewError(fiber.StatusBadRequest, "No fields to update"))
	}

	updatedComment, err := h.commentRepository.UpdateComment(commentID, userID, updates)
	if err != nil {
		log.Println("Error updating comment:", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to update comment", err)
	}
	if updatedComment == nil {
		return utils.SendError(c, fiber.StatusNotFound, "Comment not found or you don't have permission to update it", fiber.NewError(fiber.StatusNotFound, "Comment not found or you don't have permission to update it"))
	}

	return utils.SendJSON(c, fiber.StatusOK, updatedComment)
}

// DeleteComment deletes a comment
func (h *CommentHandler) DeleteComment(c *fiber.Ctx) error {
	commentID, err := utils.ValidateRequiredParam(c, "commentid")
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Comment ID is required", err)
	}

	userID, ok := c.Locals("userId").(string)
	if !ok || userID == "" {
		return utils.SendError(c, fiber.StatusUnauthorized, "Unauthorized", fiber.NewError(fiber.StatusUnauthorized, "You must be logged in to delete comments"))
	}

	err = h.commentRepository.DeleteComment(commentID, userID)
	if err != nil {
		log.Println("Error deleting comment:", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to delete comment", err)
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// CreateReaction creates a reaction on a comment
func (h *CommentHandler) CreateReaction(c *fiber.Ctx) error {
	commentID, err := utils.ValidateRequiredParam(c, "commentid")
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Comment ID is required", err)
	}

	userID, ok := c.Locals("userId").(string)
	if !ok || userID == "" {
		return utils.SendError(c, fiber.StatusUnauthorized, "Unauthorized", fiber.NewError(fiber.StatusUnauthorized, "You must be logged in to react to comments"))
	}

	var req models.CreateReactionRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	if req.Type == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "Validation failed", fiber.NewError(fiber.StatusBadRequest, "Reaction type is required"))
	}

	// Verify comment exists
	comment, err := h.commentRepository.GetCommentByID(commentID)
	if err != nil {
		log.Println("Error validating comment:", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to validate comment", err)
	}
	if comment == nil {
		return utils.SendError(c, fiber.StatusNotFound, "Comment not found", fiber.NewError(fiber.StatusNotFound, "The comment does not exist"))
	}

	now := time.Now()
	reaction := models.CommentReaction{
		Id:        cuid.New(),
		CommentId: commentID,
		UserId:    userID,
		Type:      req.Type,
		CreatedAt: now,
	}

	createdReaction, err := h.commentRepository.CreateReaction(reaction)
	if err != nil {
		log.Println("Error creating reaction:", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to create reaction", err)
	}

	return utils.SendJSON(c, fiber.StatusCreated, createdReaction)
}

// DeleteReaction deletes a reaction from a comment
func (h *CommentHandler) DeleteReaction(c *fiber.Ctx) error {
	commentID, err := utils.ValidateRequiredParam(c, "commentid")
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Comment ID is required", err)
	}

	userID, ok := c.Locals("userId").(string)
	if !ok || userID == "" {
		return utils.SendError(c, fiber.StatusUnauthorized, "Unauthorized", fiber.NewError(fiber.StatusUnauthorized, "You must be logged in to remove reactions"))
	}

	reactionType := c.Query("type")
	if reactionType == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "Validation failed", fiber.NewError(fiber.StatusBadRequest, "Reaction type is required"))
	}

	err = h.commentRepository.DeleteReaction(commentID, userID, reactionType)
	if err != nil {
		log.Println("Error deleting reaction:", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to delete reaction", err)
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// GetReactionsByCommentID retrieves all reactions for a comment
func (h *CommentHandler) GetReactionsByCommentID(c *fiber.Ctx) error {
	commentID, err := utils.ValidateRequiredParam(c, "commentid")
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Comment ID is required", err)
	}

	reactions, err := h.commentRepository.GetReactionsByCommentID(commentID)
	if err != nil {
		log.Println("Error retrieving reactions:", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve reactions", err)
	}

	return utils.SendJSON(c, fiber.StatusOK, reactions)
}

// GetReactionSummary retrieves reaction summary for a comment
func (h *CommentHandler) GetReactionSummary(c *fiber.Ctx) error {
	commentID, err := utils.ValidateRequiredParam(c, "commentid")
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Comment ID is required", err)
	}

	summary, err := h.commentRepository.GetReactionSummary(commentID)
	if err != nil {
		log.Println("Error retrieving reaction summary:", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve reaction summary", err)
	}

	return utils.SendJSON(c, fiber.StatusOK, summary)
}

// GetCommentCount retrieves the total number of comments for a marketplace item
func (h *CommentHandler) GetCommentCount(c *fiber.Ctx) error {
	itemID, err := utils.ValidateRequiredParam(c, "itemid")
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Item ID is required", err)
	}

	count, err := h.commentRepository.GetCommentCountByItemID(itemID)
	if err != nil {
		log.Println("Error retrieving comment count:", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve comment count", err)
	}

	response := fiber.Map{
		"itemId": itemID,
		"count":  count,
	}
	return utils.SendJSON(c, fiber.StatusOK, response)
}

// ModerateComment updates the status of a comment (admin/moderator only)
func (h *CommentHandler) ModerateComment(c *fiber.Ctx) error {
	commentID, err := utils.ValidateRequiredParam(c, "commentid")
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Comment ID is required", err)
	}

	userID, ok := c.Locals("userId").(string)
	if !ok || userID == "" {
		return utils.SendError(c, fiber.StatusUnauthorized, "Unauthorized", fiber.NewError(fiber.StatusUnauthorized, "You must be logged in to moderate comments"))
	}

	// TODO: Add admin/moderator role check here
	// For now, we'll allow any authenticated user to moderate
	// In production, you should check user roles

	type ModerateRequest struct {
		Status string `json:"status" validate:"required"`
	}

	var req ModerateRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	if req.Status == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "Validation failed", fiber.NewError(fiber.StatusBadRequest, "Status is required"))
	}

	// Validate status values
	validStatuses := map[string]bool{
		"published": true,
		"pending":   true,
		"flagged":   true,
		"deleted":   true,
	}

	if !validStatuses[req.Status] {
		return utils.SendError(c, fiber.StatusBadRequest, "Validation failed", fiber.NewError(fiber.StatusBadRequest, "Invalid status value"))
	}

	err = h.commentRepository.ModerateComment(commentID, req.Status)
	if err != nil {
		log.Println("Error moderating comment:", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to moderate comment", err)
	}

	response := fiber.Map{
		"message": "Comment moderated successfully",
		"status":  req.Status,
	}
	return utils.SendJSON(c, fiber.StatusOK, response)
}
