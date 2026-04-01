package handlers

import (
	"errors"
	"log"
	"my-go-app/internal/models"
	"my-go-app/internal/services"
	"my-go-app/pkg/utils"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/lucsky/cuid"
)

type CommentHandler struct {
	commentService services.CommentServiceInterface
}

func NewCommentHandler(commentService services.CommentServiceInterface) *CommentHandler {
	return &CommentHandler{
		commentService: commentService,
	}
}

func (h *CommentHandler) CreateComment(c *fiber.Ctx) error {
	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	var req models.CreateCommentRequest
	if err := utils.ValidateAndParseBody(c, &req); err != nil {
		return err
	}

	var parentCommentID *string
	if req.ParentId != nil && *req.ParentId != "" {
		parentCommentID = req.ParentId
	}

	comment := &models.Comment{
		Id:        cuid.New(),
		Content:   req.Content,
		AuthorId:  userID,
		ItemId:    req.ItemId,
		ParentId:  parentCommentID,
		Status:    "published",
		Edited:    false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	created, err := h.commentService.CreateComment(c.Context(), comment)
	if err != nil {
		log.Println("Error creating comment:", err)
		return utils.HandleRepoError(c, err, "", "Failed to create comment")
	}

	return utils.SendJSON(c, fiber.StatusCreated, created)
}

func (h *CommentHandler) GetComments(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	filter := models.CommentFilter{
		ItemId:    c.Query("itemId"),
		AuthorId:  c.Query("authorId"),
		Status:    c.Query("status", "published"),
		SortBy:    c.Query("sortBy", "createdAt"),
		SortOrder: c.Query("sortOrder", "desc"),
		Limit:     limit,
		Offset:    offset,
	}

	if parentIdStr := c.Query("parentId"); parentIdStr != "" {
		filter.ParentId = &parentIdStr
	} else if c.Query("topLevel") == "true" {
		emptyStr := ""
		filter.ParentId = &emptyStr
	}

	comments, total, err := h.commentService.GetCommentsByItemID(c.Context(), filter.ItemId, filter)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to retrieve comments")
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{
		"data":   comments,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *CommentHandler) GetCommentByID(c *fiber.Ctx) error {
	ids, err := utils.MustParams(c, "commentid")
	if err != nil {
		return err
	}
	commentID := ids[0]

	comment, err := h.commentService.GetCommentByID(c.Context(), commentID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Comment not found")
	}

	return utils.SendJSON(c, fiber.StatusOK, comment)
}

func (h *CommentHandler) GetCommentsByItemID(c *fiber.Ctx) error {
	ids, err := utils.MustParams(c, "itemid")
	if err != nil {
		return err
	}
	itemID := ids[0]

	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	filter := models.CommentFilter{
		ItemId:    itemID,
		Status:    c.Query("status", "published"),
		SortBy:    c.Query("sortBy", "createdAt"),
		SortOrder: c.Query("sortOrder", "desc"),
		Limit:     limit,
		Offset:    offset,
	}

	if c.Query("includeReplies") != "false" {
		emptyStr := ""
		filter.ParentId = &emptyStr
	}

	comments, total, err := h.commentService.GetCommentsByItemID(c.Context(), itemID, filter)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to retrieve comments")
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{
		"data":   comments,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *CommentHandler) UpdateComment(c *fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "commentid")
	if err != nil {
		return err
	}
	commentID := ids[0]

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

	if err := utils.RequireUpdates(updates); err != nil {
		return err
	}

	updated, err := h.commentService.UpdateComment(c.Context(), commentID, userID, updates)
	if err != nil {
		if errors.Is(err, errors.New("unauthorized: user is not the comment author")) {
			return fiber.NewError(fiber.StatusForbidden, "You do not have permission to update this comment")
		}
		return utils.HandleRepoError(c, err, "Comment not found", "Failed to update comment")
	}

	return utils.SendJSON(c, fiber.StatusOK, updated)
}

func (h *CommentHandler) DeleteComment(c *fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "commentid")
	if err != nil {
		return err
	}
	commentID := ids[0]

	if err := h.commentService.DeleteComment(c.Context(), commentID, userID); err != nil {
		if errors.Is(err, errors.New("unauthorized: user is not the comment author")) {
			return fiber.NewError(fiber.StatusForbidden, "You do not have permission to delete this comment")
		}
		return utils.HandleRepoError(c, err, "Comment not found", "Failed to delete comment")
	}

	return utils.SendNoContent(c)
}

func (h *CommentHandler) CreateReaction(c *fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "commentid")
	if err != nil {
		return err
	}
	commentID := ids[0]

	var req models.CreateReactionRequest
	if err := utils.ValidateAndParseBody(c, &req); err != nil {
		return err
	}

	reaction := models.CommentReaction{
		Id:        cuid.New(),
		CommentId: commentID,
		UserId:    userID,
		Type:      req.Type,
		CreatedAt: time.Now(),
	}

	created, err := h.commentService.CreateReaction(c.Context(), &reaction)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to create reaction")
	}

	return utils.SendJSON(c, fiber.StatusCreated, created)
}

func (h *CommentHandler) DeleteReaction(c *fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "commentid")
	if err != nil {
		return err
	}
	commentID := ids[0]

	reactionType := c.Query("type")
	if reactionType == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Reaction type query parameter is required")
	}

	if err := h.commentService.DeleteReaction(c.Context(), commentID, userID, reactionType); err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Reaction not found")
	}

	return utils.SendNoContent(c)
}

func (h *CommentHandler) GetReactionsByCommentID(c *fiber.Ctx) error {
	ids, err := utils.MustParams(c, "commentid")
	if err != nil {
		return err
	}
	commentID := ids[0]

	reactions, err := h.commentService.GetReactionsByCommentID(c.Context(), commentID)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to retrieve reactions")
	}

	return utils.SendJSON(c, fiber.StatusOK, reactions)
}

func (h *CommentHandler) GetReactionSummary(c *fiber.Ctx) error {
	ids, err := utils.MustParams(c, "commentid")
	if err != nil {
		return err
	}
	commentID := ids[0]

	summary, err := h.commentService.GetReactionSummary(c.Context(), commentID)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to retrieve reaction summary")
	}

	return utils.SendJSON(c, fiber.StatusOK, summary)
}

func (h *CommentHandler) GetCommentCount(c *fiber.Ctx) error {
	ids, err := utils.MustParams(c, "itemid")
	if err != nil {
		return err
	}
	itemID := ids[0]

	count, err := h.commentService.GetCommentCountByItemID(c.Context(), itemID)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to retrieve comment count")
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{
		"itemId": itemID,
		"count":  count,
	})
}

func (h *CommentHandler) ModerateComment(c *fiber.Ctx) error {
	_, ids, err := utils.MustUserAndParams(c, "commentid")
	if err != nil {
		return err
	}
	commentID := ids[0]

	var req struct {
		Status string `json:"status" validate:"required,oneof=published pending flagged deleted"`
	}
	if err := utils.ValidateAndParseBody(c, &req); err != nil {
		return err
	}

	if err := h.commentService.ModerateComment(c.Context(), commentID, req.Status); err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Comment not found")
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{
		"message": "Comment moderated successfully",
		"status":  req.Status,
	})
}