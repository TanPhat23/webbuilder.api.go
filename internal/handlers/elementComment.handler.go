package handlers

import (
	"my-go-app/internal/models"
	"my-go-app/internal/services"
	"my-go-app/pkg/utils"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ElementCommentHandler struct {
	elementCommentService services.ElementCommentServiceInterface
}

func NewElementCommentHandler(elementCommentService services.ElementCommentServiceInterface) *ElementCommentHandler {
	return &ElementCommentHandler{
		elementCommentService: elementCommentService,
	}
}

// POST /element-comments
func (h *ElementCommentHandler) CreateElementComment(c *fiber.Ctx) error {
	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	var req models.CreateElementCommentRequest
	if err := utils.ValidateAndParseBody(c, &req); err != nil {
		return err
	}

	comment := &models.ElementComment{
		Id:        uuid.NewString(),
		Content:   req.Content,
		ElementId: req.ElementId,
		AuthorId:  userID,
	}

	created, err := h.elementCommentService.CreateElementComment(c.Context(), comment)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to create comment")
	}

	return utils.SendJSON(c, fiber.StatusCreated, created)
}

// GET /element-comments/:id
func (h *ElementCommentHandler) GetElementCommentByID(c *fiber.Ctx) error {
	ids, err := utils.MustParams(c, "id")
	if err != nil {
		return err
	}
	commentID := ids[0]

	comment, err := h.elementCommentService.GetElementCommentByID(c.Context(), commentID)
	if err != nil {
		return utils.HandleRepoError(c, err, "Comment not found", "Failed to retrieve comment")
	}

	return utils.SendJSON(c, fiber.StatusOK, comment)
}

// GET /elements/:elementId/comments
func (h *ElementCommentHandler) GetElementComments(c *fiber.Ctx) error {
	ids, err := utils.MustParams(c, "elementId")
	if err != nil {
		return err
	}
	elementID := ids[0]

	filter := &models.ElementCommentFilter{
		Limit:     20,
		Offset:    0,
		SortBy:    "CreatedAt",
		SortOrder: "DESC",
	}

	if l, err := strconv.Atoi(c.Query("limit", "20")); err == nil && l > 0 {
		filter.Limit = l
	}
	if o, err := strconv.Atoi(c.Query("offset", "0")); err == nil && o >= 0 {
		filter.Offset = o
	}
	if authorID := c.Query("authorId"); authorID != "" {
		filter.AuthorId = authorID
	}
	if resolved := c.Query("resolved"); resolved != "" {
		if r, err := strconv.ParseBool(resolved); err == nil {
			filter.Resolved = &r
		}
	}
	if sortBy := c.Query("sortBy"); sortBy != "" {
		filter.SortBy = sortBy
	}
	if sortOrder := c.Query("sortOrder"); sortOrder == "ASC" || sortOrder == "DESC" {
		filter.SortOrder = sortOrder
	}

	comments, err := h.elementCommentService.GetElementComments(c.Context(), elementID, *filter)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to retrieve comments")
	}

	count, err := h.elementCommentService.CountElementComments(c.Context(), elementID)
	if err != nil {
		count = 0
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{
		"data":   comments,
		"total":  count,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// PATCH /element-comments/:id
func (h *ElementCommentHandler) UpdateElementComment(c *fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "id")
	if err != nil {
		return err
	}
	commentID := ids[0]

	var req models.UpdateElementCommentRequest
	if err := utils.ValidateAndParseBody(c, &req); err != nil {
		return err
	}

	updates := map[string]any{}
	if req.Content != nil  { updates["Content"] = *req.Content }
	if req.Resolved != nil { updates["Resolved"] = *req.Resolved }

	if err := utils.RequireUpdates(updates); err != nil {
		return err
	}

	updated, err := h.elementCommentService.UpdateElementComment(c.Context(), commentID, userID, updates)
	if err != nil {
		return utils.HandleRepoError(c, err, "Comment not found", "Failed to update comment")
	}

	return utils.SendJSON(c, fiber.StatusOK, updated)
}

// DELETE /element-comments/:id
func (h *ElementCommentHandler) DeleteElementComment(c *fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "id")
	if err != nil {
		return err
	}
	commentID := ids[0]

	if err := h.elementCommentService.DeleteElementComment(c.Context(), commentID, userID); err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to delete comment")
	}

	return utils.SendNoContent(c)
}

// PATCH /element-comments/:id/toggle-resolved
func (h *ElementCommentHandler) ToggleResolvedStatus(c *fiber.Ctx) error {
	ids, err := utils.MustParams(c, "id")
	if err != nil {
		return err
	}
	commentID := ids[0]

	err = h.elementCommentService.ToggleResolvedStatus(c.Context(), commentID)
	if err != nil {
		return utils.HandleRepoError(c, err, "Comment not found", "Failed to toggle resolved status")
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{"message": "Resolved status toggled successfully"})
}

// GET /element-comments/author/:authorId
func (h *ElementCommentHandler) GetCommentsByAuthorID(c *fiber.Ctx) error {
	ids, err := utils.MustParams(c, "authorId")
	if err != nil {
		return err
	}
	authorID := ids[0]

	limit := 20
	offset := 0
	if l, err := strconv.Atoi(c.Query("limit", "20")); err == nil && l > 0 {
		limit = l
	}
	if o, err := strconv.Atoi(c.Query("offset", "0")); err == nil && o >= 0 {
		offset = o
	}

	comments, err := h.elementCommentService.GetElementCommentsByAuthorID(c.Context(), authorID, limit, offset)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to retrieve comments")
	}

	return utils.SendJSON(c, fiber.StatusOK, comments)
}

// GET /projects/:projectId/comments
func (h *ElementCommentHandler) GetCommentsByProjectID(c *fiber.Ctx) error {
	ids, err := utils.MustParams(c, "projectId")
	if err != nil {
		return err
	}
	projectID := ids[0]

	limit := 20
	offset := 0
	if l, err := strconv.Atoi(c.Query("limit", "20")); err == nil && l > 0 {
		limit = l
	}
	if o, err := strconv.Atoi(c.Query("offset", "0")); err == nil && o >= 0 {
		offset = o
	}

	comments, err := h.elementCommentService.GetElementCommentsByProjectID(c.Context(), projectID, limit, offset)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to retrieve comments")
	}

	count, err := h.elementCommentService.CountElementCommentsByProjectID(c.Context(), projectID)
	if err != nil {
		count = 0
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{
		"data":   comments,
		"total":  count,
		"limit":  limit,
		"offset": offset,
	})
}