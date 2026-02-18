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

// CreateComment creates a new comment on a marketplace item.
func (h *CommentHandler) CreateComment(c *fiber.Ctx) error {
	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	var req models.CreateCommentRequest
	if err := utils.ValidateAndParseBody(c, &req); err != nil {
		return err
	}

	item, err := h.marketplaceRepository.GetMarketplaceItemByID(req.ItemId)
	if err != nil {
		log.Println("Error validating marketplace item:", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to validate marketplace item", err)
	}
	if item == nil {
		return utils.SendError(c, fiber.StatusNotFound, "Marketplace item not found", nil)
	}

	if req.ParentId != nil && *req.ParentId != "" {
		parentComment, err := h.commentRepository.GetCommentByID(*req.ParentId)
		if err != nil {
			log.Println("Error validating parent comment:", err)
			return utils.SendError(c, fiber.StatusInternalServerError, "Failed to validate parent comment", err)
		}
		if parentComment == nil {
			return utils.SendError(c, fiber.StatusNotFound, "Parent comment not found", nil)
		}
		if parentComment.ItemId != req.ItemId {
			return utils.SendError(c, fiber.StatusBadRequest, "Parent comment does not belong to the specified item", nil)
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

// GetComments retrieves comments with filtering.
func (h *CommentHandler) GetComments(c *fiber.Ctx) error {
	filter := models.CommentFilter{
		ItemId:    c.Query("itemId"),
		AuthorId:  c.Query("authorId"),
		Status:    c.Query("status", "published"),
		SortBy:    c.Query("sortBy", "createdAt"),
		SortOrder: c.Query("sortOrder", "desc"),
	}

	if parentIdStr := c.Query("parentId"); parentIdStr != "" {
		filter.ParentId = &parentIdStr
	} else if c.Query("topLevel") == "true" {
		emptyStr := ""
		filter.ParentId = &emptyStr
	}

	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	filter.Limit = limit
	filter.Offset = offset

	comments, total, err := h.commentRepository.GetComments(filter)
	if err != nil {
		log.Println("Error retrieving comments:", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve comments", err)
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{
		"data":   comments,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetCommentByID retrieves a single comment.
func (h *CommentHandler) GetCommentByID(c *fiber.Ctx) error {
	commentID, err := utils.ValidateRequiredParam(c, "commentid")
	if err != nil {
		return err
	}

	comment, err := h.commentRepository.GetCommentByID(commentID)
	if err != nil {
		log.Println("Error retrieving comment:", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve comment", err)
	}
	if comment == nil {
		return utils.SendError(c, fiber.StatusNotFound, "Comment not found", nil)
	}

	return utils.SendJSON(c, fiber.StatusOK, comment)
}

// GetCommentsByItemID retrieves all comments for a marketplace item.
func (h *CommentHandler) GetCommentsByItemID(c *fiber.Ctx) error {
	itemID, err := utils.ValidateRequiredParam(c, "itemid")
	if err != nil {
		return err
	}

	filter := models.CommentFilter{
		ItemId:    itemID,
		Status:    c.Query("status", "published"),
		SortBy:    c.Query("sortBy", "createdAt"),
		SortOrder: c.Query("sortOrder", "desc"),
	}

	if c.Query("includeReplies") != "false" {
		emptyStr := ""
		filter.ParentId = &emptyStr
	}

	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	filter.Limit = limit
	filter.Offset = offset

	comments, total, err := h.commentRepository.GetComments(filter)
	if err != nil {
		log.Println("Error retrieving comments:", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve comments", err)
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{
		"data":   comments,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// UpdateComment updates a comment's content or status.
func (h *CommentHandler) UpdateComment(c *fiber.Ctx) error {
	commentID, err := utils.ValidateRequiredParam(c, "commentid")
	if err != nil {
		return err
	}

	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	var req models.UpdateCommentRequest
	if err := utils.ValidateAndParseBody(c, &req); err != nil {
		return err
	}

	updates := make(map[string]any)
	if req.Content != nil {
		if *req.Content == "" {
			return fiber.NewError(fiber.StatusBadRequest, "Content cannot be empty")
		}
		updates["Content"] = *req.Content
	}
	if req.Status != nil {
		updates["Status"] = *req.Status
	}

	if len(updates) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "No fields to update")
	}

	updatedComment, err := h.commentRepository.UpdateComment(commentID, userID, updates)
	if err != nil {
		log.Println("Error updating comment:", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to update comment", err)
	}
	if updatedComment == nil {
		return utils.SendError(c, fiber.StatusNotFound, "Comment not found or you don't have permission to update it", nil)
	}

	return utils.SendJSON(c, fiber.StatusOK, updatedComment)
}

// DeleteComment deletes a comment.
func (h *CommentHandler) DeleteComment(c *fiber.Ctx) error {
	commentID, err := utils.ValidateRequiredParam(c, "commentid")
	if err != nil {
		return err
	}

	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	if err := h.commentRepository.DeleteComment(commentID, userID); err != nil {
		log.Println("Error deleting comment:", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to delete comment", err)
	}

	return utils.SendNoContent(c)
}

// CreateReaction creates a reaction on a comment.
func (h *CommentHandler) CreateReaction(c *fiber.Ctx) error {
	commentID, err := utils.ValidateRequiredParam(c, "commentid")
	if err != nil {
		return err
	}

	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	var req models.CreateReactionRequest
	if err := utils.ValidateAndParseBody(c, &req); err != nil {
		return err
	}

	comment, err := h.commentRepository.GetCommentByID(commentID)
	if err != nil {
		log.Println("Error validating comment:", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to validate comment", err)
	}
	if comment == nil {
		return utils.SendError(c, fiber.StatusNotFound, "Comment not found", nil)
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

// DeleteReaction removes a reaction from a comment.
func (h *CommentHandler) DeleteReaction(c *fiber.Ctx) error {
	commentID, err := utils.ValidateRequiredParam(c, "commentid")
	if err != nil {
		return err
	}

	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	reactionType := c.Query("type")
	if reactionType == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Reaction type query parameter is required")
	}

	if err := h.commentRepository.DeleteReaction(commentID, userID, reactionType); err != nil {
		log.Println("Error deleting reaction:", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to delete reaction", err)
	}

	return utils.SendNoContent(c)
}

// GetReactionsByCommentID retrieves all reactions for a comment.
func (h *CommentHandler) GetReactionsByCommentID(c *fiber.Ctx) error {
	commentID, err := utils.ValidateRequiredParam(c, "commentid")
	if err != nil {
		return err
	}

	reactions, err := h.commentRepository.GetReactionsByCommentID(commentID)
	if err != nil {
		log.Println("Error retrieving reactions:", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve reactions", err)
	}

	return utils.SendJSON(c, fiber.StatusOK, reactions)
}

// GetReactionSummary retrieves the aggregated reaction counts for a comment.
func (h *CommentHandler) GetReactionSummary(c *fiber.Ctx) error {
	commentID, err := utils.ValidateRequiredParam(c, "commentid")
	if err != nil {
		return err
	}

	summary, err := h.commentRepository.GetReactionSummary(commentID)
	if err != nil {
		log.Println("Error retrieving reaction summary:", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve reaction summary", err)
	}

	return utils.SendJSON(c, fiber.StatusOK, summary)
}

// GetCommentCount retrieves the total number of comments for a marketplace item.
func (h *CommentHandler) GetCommentCount(c *fiber.Ctx) error {
	itemID, err := utils.ValidateRequiredParam(c, "itemid")
	if err != nil {
		return err
	}

	count, err := h.commentRepository.GetCommentCountByItemID(itemID)
	if err != nil {
		log.Println("Error retrieving comment count:", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve comment count", err)
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{
		"itemId": itemID,
		"count":  count,
	})
}

// ModerateComment updates the status of a comment (admin/moderator only).
func (h *CommentHandler) ModerateComment(c *fiber.Ctx) error {
	commentID, err := utils.ValidateRequiredParam(c, "commentid")
	if err != nil {
		return err
	}

	_, err = utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	var req struct {
		Status string `json:"status" validate:"required,oneof=published pending flagged deleted"`
	}
	if err := utils.ValidateAndParseBody(c, &req); err != nil {
		return err
	}

	if err := h.commentRepository.ModerateComment(commentID, req.Status); err != nil {
		log.Println("Error moderating comment:", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to moderate comment", err)
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{
		"message": "Comment moderated successfully",
		"status":  req.Status,
	})
}
