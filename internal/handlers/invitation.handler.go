package handlers

import (
	"errors"
	"my-go-app/internal/authz"
	"my-go-app/internal/dto"
	"my-go-app/internal/models"
	"my-go-app/internal/services"
	"my-go-app/pkg/utils"

	"github.com/gofiber/fiber/v3"
)

type InvitationHandler struct {
	invitationService services.InvitationServiceInterface
}

func NewInvitationHandler(invitationService services.InvitationServiceInterface) *InvitationHandler {
	return &InvitationHandler{
		invitationService: invitationService,
	}
}

func (h *InvitationHandler) Create(c fiber.Ctx) error {
	userID, err := utils.RequireUserID(c)
	if err != nil {
		return err
	}

	req := new(models.CreateInvitationRequest)
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	if req.Role == "" {
		req.Role = models.RoleEditor
	}

	invitation, err := h.invitationService.Create(c.RequestCtx(), req.ProjectID, req.Email, req.Role, userID)
	if err != nil {
		if errors.Is(err, authz.ErrForbidden) {
			return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
		}
		return utils.SendError(c, fiber.StatusBadRequest, "Failed to create invitation", err, userID)
	}

	return utils.SendJSON(c, fiber.StatusCreated, invitation)
}

func (h *InvitationHandler) ListByProject(c fiber.Ctx) error {
	userID, ids, err := utils.RequireUserAndParams(c, "projectid")
	if err != nil {
		return err
	}
	projectID := ids[0]

	if err := h.invitationService.RequireProjectOwnership(c.RequestCtx(), projectID, userID); err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	invitations, err := h.invitationService.ListByProject(c.RequestCtx(), projectID, userID)
	if err != nil {
		return utils.RespondError(c, err, "", "Failed to retrieve invitations")
	}

	return utils.SendJSON(c, fiber.StatusOK, invitations)
}

func (h *InvitationHandler) ListPendingByProject(c fiber.Ctx) error {
	userID, ids, err := utils.RequireUserAndParams(c, "projectid")
	if err != nil {
		return err
	}
	projectID := ids[0]

	if err := h.invitationService.RequireProjectOwnership(c.RequestCtx(), projectID, userID); err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	invitations, err := h.invitationService.ListPendingByProject(c.RequestCtx(), projectID, userID)
	if err != nil {
		return utils.RespondError(c, err, "", "Failed to retrieve pending invitations")
	}

	return utils.SendJSON(c, fiber.StatusOK, invitations)
}

func (h *InvitationHandler) Accept(c fiber.Ctx) error {
	userID, err := utils.RequireUserID(c)
	if err != nil {
		return err
	}

	req := new(models.AcceptInvitationRequest)
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	if err := h.invitationService.Accept(c.RequestCtx(), req.Token, userID); err != nil {
		if errors.Is(err, authz.ErrForbidden) {
			return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
		}
		return utils.SendError(c, fiber.StatusBadRequest, "Failed to accept invitation", err, userID)
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{"message": "Invitation accepted successfully"})
}

func (h *InvitationHandler) Cancel(c fiber.Ctx) error {
	userID, ids, err := utils.RequireUserAndParams(c, "invitationid")
	if err != nil {
		return err
	}
	invitationID := ids[0]

	invitation, err := h.invitationService.GetByID(c.RequestCtx(), invitationID)
	if err != nil {
		return utils.RespondError(c, err, "Invitation not found", "Failed to retrieve invitation")
	}

	if err := h.invitationService.RequireProjectOwnership(c.RequestCtx(), invitation.ProjectId, userID); err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	if err := h.invitationService.Cancel(c.RequestCtx(), invitationID); err != nil {
		return utils.RespondError(c, err, "", "Failed to cancel invitation")
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{"message": "Invitation cancelled successfully"})
}

func (h *InvitationHandler) UpdateStatus(c fiber.Ctx) error {
	userID, ids, err := utils.RequireUserAndParams(c, "invitationid")
	if err != nil {
		return err
	}
	invitationID := ids[0]

	req := new(dto.UpdateInvitationStatusRequest)
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	invitation, err := h.invitationService.GetByID(c.RequestCtx(), invitationID)
	if err != nil {
		return utils.RespondError(c, err, "Invitation not found", "Failed to retrieve invitation")
	}

	if err := h.invitationService.RequireProjectOwnership(c.RequestCtx(), invitation.ProjectId, userID); err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	if err := h.invitationService.UpdateStatus(c.RequestCtx(), invitationID, req.Status); err != nil {
		return utils.RespondError(c, err, "", "Failed to update invitation status")
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{"message": "Invitation status updated successfully"})
}

func (h *InvitationHandler) Delete(c fiber.Ctx) error {
	userID, ids, err := utils.RequireUserAndParams(c, "invitationid")
	if err != nil {
		return err
	}
	invitationID := ids[0]

	invitation, err := h.invitationService.GetByID(c.RequestCtx(), invitationID)
	if err != nil {
		return utils.RespondError(c, err, "Invitation not found", "Failed to retrieve invitation")
	}

	if err := h.invitationService.RequireProjectOwnership(c.RequestCtx(), invitation.ProjectId, userID); err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	if err := h.invitationService.Delete(c.RequestCtx(), invitationID); err != nil {
		return utils.RespondError(c, err, "", "Failed to delete invitation")
	}

	return utils.SendNoContent(c)
}

// fiber:context-methods migrated
