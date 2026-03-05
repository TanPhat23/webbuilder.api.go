package handlers

import (
	"log"
	"my-go-app/internal/dto"
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
	if err := utils.ValidateAndParseBody(c, &req); err != nil {
		return err
	}

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
	userID, ids, err := utils.MustUserAndParams(c, "projectid")
	if err != nil {
		return err
	}
	projectID := ids[0]

	if err := h.invitationService.CheckProjectOwnership(c.Context(), projectID, userID); err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	invitations, err := h.invitationService.GetInvitationsByProject(c.Context(), projectID)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to retrieve invitations")
	}

	return utils.SendJSON(c, fiber.StatusOK, invitations)
}

func (h *InvitationHandler) GetPendingInvitationsByProject(c *fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "projectid")
	if err != nil {
		return err
	}
	projectID := ids[0]

	if err := h.invitationService.CheckProjectOwnership(c.Context(), projectID, userID); err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	invitations, err := h.invitationService.GetPendingInvitationsByProject(c.Context(), projectID)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to retrieve pending invitations")
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
	if err := utils.ValidateAndParseBody(c, &req); err != nil {
		return err
	}

	if err := h.invitationService.AcceptInvitation(c.Context(), req.Token, userID); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Failed to accept invitation", err, userID)
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{"message": "Invitation accepted successfully"})
}

func (h *InvitationHandler) CancelInvitation(c *fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "invitationid")
	if err != nil {
		return err
	}
	invitationID := ids[0]

	invitation, err := h.invitationService.GetInvitationByID(c.Context(), invitationID)
	if err != nil {
		return utils.HandleRepoError(c, err, "Invitation not found", "Failed to retrieve invitation")
	}

	if err := h.invitationService.CheckProjectOwnership(c.Context(), invitation.ProjectId, userID); err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	if err := h.invitationService.CancelInvitation(c.Context(), invitationID); err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to cancel invitation")
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{"message": "Invitation cancelled successfully"})
}

func (h *InvitationHandler) UpdateInvitationStatus(c *fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "invitationid")
	if err != nil {
		return err
	}
	invitationID := ids[0]

	var req dto.UpdateInvitationStatusRequest
	if err := utils.ValidateAndParseBody(c, &req); err != nil {
		return err
	}

	invitation, err := h.invitationService.GetInvitationByID(c.Context(), invitationID)
	if err != nil {
		return utils.HandleRepoError(c, err, "Invitation not found", "Failed to retrieve invitation")
	}

	if err := h.invitationService.CheckProjectOwnership(c.Context(), invitation.ProjectId, userID); err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	if err := h.invitationService.UpdateInvitationStatus(c.Context(), invitationID, req.Status); err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to update invitation status")
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{"message": "Invitation status updated successfully"})
}

func (h *InvitationHandler) DeleteInvitation(c *fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "invitationid")
	if err != nil {
		return err
	}
	invitationID := ids[0]

	invitation, err := h.invitationService.GetInvitationByID(c.Context(), invitationID)
	if err != nil {
		return utils.HandleRepoError(c, err, "Invitation not found", "Failed to retrieve invitation")
	}

	if err := h.invitationService.CheckProjectOwnership(c.Context(), invitation.ProjectId, userID); err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	if err := h.invitationService.DeleteInvitation(c.Context(), invitationID); err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to delete invitation")
	}

	return utils.SendNoContent(c)
}