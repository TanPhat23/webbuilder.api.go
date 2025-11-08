package handlers

import (
	"log"
	"my-go-app/internal/models"
	"my-go-app/internal/services"
	"my-go-app/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

type InvitationHandler struct {
	invitationService services.InvitationServiceInterface
}

func NewInvitationHandler(invitationService services.InvitationServiceInterface) *InvitationHandler {
	return &InvitationHandler{
		invitationService: invitationService,
	}
}

func (h *InvitationHandler) CreateInvitation(c *fiber.Ctx) error {
	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	var req models.CreateInvitationRequest
	if err := utils.ValidateJSONBody(c, &req); err != nil {
		return err
	}

	// Default role to editor if not provided
	if req.Role == "" {
		req.Role = models.RoleEditor
	}

	invitation, err := h.invitationService.CreateInvitation(c.Context(), req.ProjectID, req.Email, req.Role, userID)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Failed to create invitation", err, userID)
	}

	return utils.SendJSON(c, fiber.StatusCreated, invitation)
}

func (h *InvitationHandler) GetInvitationsByProject(c *fiber.Ctx) error {
	projectID, err := utils.ValidateRequiredParam(c, "projectid")
	if err != nil {
		return err
	}

	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	// Check ownership
	if err := h.invitationService.CheckProjectOwnership(c.Context(), projectID, userID); err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	invitations, err := h.invitationService.GetInvitationsByProject(c.Context(), projectID)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve invitations", err, userID)
	}

	return utils.SendJSON(c, fiber.StatusOK, invitations)
}

func (h *InvitationHandler) GetPendingInvitationsByProject(c *fiber.Ctx) error {
	projectID, err := utils.ValidateRequiredParam(c, "projectid")
	if err != nil {
		return err
	}

	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	// Check ownership
	if err := h.invitationService.CheckProjectOwnership(c.Context(), projectID, userID); err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	invitations, err := h.invitationService.GetPendingInvitationsByProject(c.Context(), projectID)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve pending invitations", err, userID)
	}

	log.Printf("Pending invitations for project %s: %v\n", projectID, invitations)
	return utils.SendJSON(c, fiber.StatusOK, invitations)
}

func (h *InvitationHandler) AcceptInvitation(c *fiber.Ctx) error {
	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	var req models.AcceptInvitationRequest
	if err := utils.ValidateJSONBody(c, &req); err != nil {
		return err
	}

	err = h.invitationService.AcceptInvitation(c.Context(), req.Token, userID)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Failed to accept invitation", err, userID)
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{"message": "Invitation accepted successfully"})
}

func (h *InvitationHandler) CancelInvitation(c *fiber.Ctx) error {
	invitationID, err := utils.ValidateRequiredParam(c, "invitationid")
	if err != nil {
		return err
	}

	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	// Get invitation to check project ownership
	invitation, err := h.invitationService.GetInvitationByID(c.Context(), invitationID)
	if err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Invitation not found", err, userID)
	}

	// Check ownership of the project
	if err := h.invitationService.CheckProjectOwnership(c.Context(), invitation.ProjectId, userID); err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	err = h.invitationService.CancelInvitation(c.Context(), invitationID)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to cancel invitation", err, userID)
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{"message": "Invitation cancelled successfully"})
}

func (h *InvitationHandler) UpdateInvitationStatus(c *fiber.Ctx) error {
	invitationID, err := utils.ValidateRequiredParam(c, "invitationid")
	if err != nil {
		return err
	}

	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	var req struct {
		Status models.InvitationStatus `json:"status" validate:"required"`
	}
	if err := utils.ValidateJSONBody(c, &req); err != nil {
		return err
	}

	// Get invitation to check project ownership
	invitation, err := h.invitationService.GetInvitationByID(c.Context(), invitationID)
	if err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Invitation not found", err, userID)
	}

	// Check ownership of the project
	if err := h.invitationService.CheckProjectOwnership(c.Context(), invitation.ProjectId, userID); err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	err = h.invitationService.UpdateInvitationStatus(c.Context(), invitationID, req.Status)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to update invitation status", err, userID)
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{"message": "Invitation status updated successfully"})
}

func (h *InvitationHandler) DeleteInvitation(c *fiber.Ctx) error {
	invitationID, err := utils.ValidateRequiredParam(c, "invitationid")
	if err != nil {
		return err
	}

	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	// Get invitation to check project ownership
	invitation, err := h.invitationService.GetInvitationByID(c.Context(), invitationID)
	if err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Invitation not found", err, userID)
	}

	// Check ownership of the project
	if err := h.invitationService.CheckProjectOwnership(c.Context(), invitation.ProjectId, userID); err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	err = h.invitationService.DeleteInvitation(c.Context(), invitationID)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to delete invitation", err, userID)
	}

	return utils.SendNoContent(c)
}
