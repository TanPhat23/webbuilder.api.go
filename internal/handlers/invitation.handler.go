package handlers

import (
	"my-go-app/internal/models"
	"my-go-app/internal/services"
	"my-go-app/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

type InvitationHandler struct {
	invitationService *services.InvitationService
}

func NewInvitationHandler(invitationService *services.InvitationService) *InvitationHandler {
	return &InvitationHandler{
		invitationService: invitationService,
	}
}

func (h *InvitationHandler) CreateInvitation(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)
	if userID == "" {
		return utils.SendError(c, fiber.StatusUnauthorized, "User not authenticated", nil)
	}

	var req struct {
		ProjectID string                `json:"projectId" validate:"required"`
		Email     string                `json:"email" validate:"required,email"`
		Role      models.CollaboratorRole `json:"role"`
	}

	if err := utils.ValidateJSONBody(c, &req); err != nil {
		return err
	}

	if req.Role == "" {
		req.Role = models.RoleEditor
	}

	invitation, err := h.invitationService.CreateInvitation(c.Context(), req.ProjectID, req.Email, req.Role, userID)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error(), err)
	}

	return utils.SendJSON(c, fiber.StatusCreated, invitation)
}

func (h *InvitationHandler) GetInvitationsByProject(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)
	if userID == "" {
		return utils.SendError(c, fiber.StatusUnauthorized, "User not authenticated", nil)
	}

	projectID := c.Params("projectid")
	if projectID == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "Project ID is required", nil)
	}

	// TODO: Check if user has access to the project

	invitations, err := h.invitationService.GetInvitationsByProject(c.Context(), projectID)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve invitations", err)
	}

	return utils.SendJSON(c, fiber.StatusOK, invitations)
}

func (h *InvitationHandler) AcceptInvitation(c *fiber.Ctx) error {
	var req struct {
		Token string `json:"token" validate:"required"`
	}

	if err := utils.ValidateJSONBody(c, &req); err != nil {
		return err
	}

	userID := c.Locals("userId").(string)
	if userID == "" {
		return utils.SendError(c, fiber.StatusUnauthorized, "User not authenticated", nil)
	}

	err := h.invitationService.AcceptInvitation(c.Context(), req.Token, userID)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Failed to accept invitation", err)
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{"message": "Invitation accepted successfully"})
}

func (h *InvitationHandler) DeleteInvitation(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)
	if userID == "" {
		return utils.SendError(c, fiber.StatusUnauthorized, "User not authenticated", nil)
	}

	invitationID := c.Params("invitationid")
	if invitationID == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "Invitation ID is required", nil)
	}

	// TODO: Check permissions

	err := h.invitationService.DeleteInvitation(c.Context(), invitationID)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to delete invitation", err)
	}

	return utils.SendNoContent(c)
}
