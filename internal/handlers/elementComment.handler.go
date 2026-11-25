package handlers

import (
	"my-go-app/internal/models"
	"my-go-app/internal/services"
	"my-go-app/pkg/utils"
	"strconv"

	"github.com/gofiber/fiber/v3"
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
func (h *ElementCommentHandler) Create(c fiber.Ctx) error {
	userID, err := utils.RequireUserID(c)
	if err != nil {
		return err
	}

	req := new(models.CreateElementCommentRequest)
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	comment := &models.ElementComment{
		Id:        uuid.NewString(),
		Content:   req.Content,
		ElementId: req.ElementId,
		AuthorId:  userID,
	}

	created, err := h.elementCommentService.Create(c.RequestCtx(), comment)
	if err != nil {
		return utils.RespondError(c, err, "", "Failed to create comment")
	}

	return utils.SendJSON(c, fiber.StatusCreated, created)
}

// GET /element-comments/:id
func (h *ElementCommentHandler) GetByID(c fiber.Ctx) error {
	userID, ids, err := utils.RequireUserAndParams(c, "id")
	if err != nil {
		return err
	}
	commentID := ids[0]

	comment, err := h.elementCommentService.GetByIDForUser(c.RequestCtx(), commentID, userID)
	if err != nil {
		return utils.RespondError(c, err, "Comment not found", "Failed to retrieve comment")
	}

	return utils.SendJSON(c, fiber.StatusOK, comment)
}

// GET /elements/:elementId/comments
func (h *ElementCommentHandler) List(c fiber.Ctx) error {
	userID, ids, err := utils.RequireUserAndParams(c, "elementId")
	if err != nil {
		return err
	}
	elementID := ids[0]

	limit, offset := utils.ParsePagination(c, 20)
	filter := models.ElementCommentFilter{
		Limit:     limit,
		Offset:    offset,
		SortBy:    c.Query("sortBy", "CreatedAt"),
		SortOrder: "DESC",
		AuthorId:  c.Query("authorId"),
	}
	if resolved := c.Query("resolved"); resolved != "" {
		if r, err := strconv.ParseBool(resolved); err == nil {
			filter.Resolved = &r
		}
	}
	if sortOrder := c.Query("sortOrder"); sortOrder == "ASC" || sortOrder == "DESC" {
		filter.SortOrder = sortOrder
	}

	comments, err := h.elementCommentService.ListForUser(c.RequestCtx(), elementID, userID, filter)
	if err != nil {
		return utils.RespondError(c, err, "", "Failed to retrieve comments")
	}

	count, err := h.elementCommentService.Count(c.RequestCtx(), elementID)
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
func (h *ElementCommentHandler) Update(c fiber.Ctx) error {
	userID, ids, err := utils.RequireUserAndParams(c, "id")
	if err != nil {
		return err
	}
	commentID := ids[0]

	req := new(models.UpdateElementCommentRequest)
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	updates := map[string]any{}
	utils.SetIfNotNil(updates, "content", req.Content)
	utils.SetIfNotNil(updates, "resolved", req.Resolved)

	if err := utils.RequireUpdates(updates); err != nil {
		return err
	}

	updated, err := h.elementCommentService.Update(c.RequestCtx(), commentID, userID, updates)
	if err != nil {
		return utils.RespondError(c, err, "Comment not found", "Failed to update comment")
	}

	return utils.SendJSON(c, fiber.StatusOK, updated)
}

// DELETE /element-comments/:id
func (h *ElementCommentHandler) Delete(c fiber.Ctx) error {
	userID, ids, err := utils.RequireUserAndParams(c, "id")
	if err != nil {
		return err
	}
	commentID := ids[0]

	if err := h.elementCommentService.Delete(c.RequestCtx(), commentID, userID); err != nil {
		return utils.RespondError(c, err, "", "Failed to delete comment")
	}

	return utils.SendNoContent(c)
}

// PATCH /element-comments/:id/toggle-resolved
func (h *ElementCommentHandler) ToggleResolved(c fiber.Ctx) error {
	userID, ids, err := utils.RequireUserAndParams(c, "id")
	if err != nil {
		return err
	}
	commentID := ids[0]

	err = h.elementCommentService.ToggleResolvedForUser(c.RequestCtx(), commentID, userID)
	if err != nil {
		return utils.RespondError(c, err, "Comment not found", "Failed to toggle resolved status")
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{"message": "Resolved status toggled successfully"})
}

// GET /element-comments/author/:authorId
func (h *ElementCommentHandler) ListByAuthor(c fiber.Ctx) error {
	userID, ids, err := utils.RequireUserAndParams(c, "authorId")
	if err != nil {
		return err
	}
	authorID := ids[0]
	if authorID != userID {
		return utils.SendError(c, fiber.StatusForbidden, "You can only view your own comments", nil, userID)
	}

	limit, offset := utils.ParsePagination(c, 20)

	comments, err := h.elementCommentService.ListByAuthor(c.RequestCtx(), authorID, limit, offset)
	if err != nil {
		return utils.RespondError(c, err, "", "Failed to retrieve comments")
	}

	return utils.SendJSON(c, fiber.StatusOK, comments)
}

// GET /projects/:projectId/comments
func (h *ElementCommentHandler) ListByProject(c fiber.Ctx) error {
	userID, ids, err := utils.RequireUserAndParams(c, "projectId")
	if err != nil {
		return err
	}
	projectID := ids[0]

	limit, offset := utils.ParsePagination(c, 20)

	comments, err := h.elementCommentService.ListByProject(c.RequestCtx(), projectID, userID, limit, offset)
	if err != nil {
		return utils.RespondError(c, err, "Project not found", "Failed to retrieve comments")
	}

	count, err := h.elementCommentService.CountByProject(c.RequestCtx(), projectID, userID)
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

// fiber:context-methods migrated
