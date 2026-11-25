package handlers

import (
	"my-go-app/internal/dto"
	"my-go-app/internal/models"
	"my-go-app/internal/services"
	"my-go-app/pkg/utils"

	"github.com/gofiber/fiber/v3"
)

type CollaboratorHandler struct {
	collaboratorService services.CollaboratorServiceInterface
}

func NewCollaboratorHandler(collaboratorService services.CollaboratorServiceInterface) *CollaboratorHandler {
	return &CollaboratorHandler{
		collaboratorService: collaboratorService,
	}
}

func (h *CollaboratorHandler) ListByProject(c fiber.Ctx) error {
	userID, ids, err := utils.RequireUserAndParams(c, "projectid")
	if err != nil {
		return err
	}
	projectID := ids[0]

	collaborators, err := h.collaboratorService.ListByProject(c.RequestCtx(), projectID, userID)
	if err != nil {
		return utils.RespondError(c, err, "Project not found", "Failed to retrieve collaborators")
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{"collaborators": collaborators})
}

func (h *CollaboratorHandler) GetByID(c fiber.Ctx) error {
	userID, ids, err := utils.RequireUserAndParams(c, "collaboratorid")
	if err != nil {
		return err
	}
	collaboratorID := ids[0]

	collaborator, err := h.collaboratorService.GetByID(c.RequestCtx(), collaboratorID, userID)
	if err != nil {
		return utils.RespondError(c, err, "Collaborator not found", "Failed to retrieve collaborator")
	}

	return utils.SendJSON(c, fiber.StatusOK, collaborator)
}

func (h *CollaboratorHandler) UpdateRole(c fiber.Ctx) error {
	userID, ids, err := utils.RequireUserAndParams(c, "collaboratorid")
	if err != nil {
		return err
	}
	collaboratorID := ids[0]

	req := new(dto.UpdateCollaboratorRoleRequest)
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	if err := h.collaboratorService.UpdateRole(c.RequestCtx(), collaboratorID, userID, req.Role); err != nil {
		return utils.RespondError(c, err, "Collaborator not found", "Failed to update collaborator role")
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{"message": "Role updated successfully"})
}

func (h *CollaboratorHandler) Delete(c fiber.Ctx) error {
	userID, ids, err := utils.RequireUserAndParams(c, "collaboratorid")
	if err != nil {
		return err
	}
	collaboratorID := ids[0]

	if err := h.collaboratorService.Delete(c.RequestCtx(), collaboratorID, userID); err != nil {
		return utils.RespondError(c, err, "Collaborator not found", "Failed to delete collaborator")
	}

	return utils.SendNoContent(c)
}

func (h *CollaboratorHandler) LeaveProject(c fiber.Ctx) error {
	userID, ids, err := utils.RequireUserAndParams(c, "projectid")
	if err != nil {
		return err
	}
	projectID := ids[0]

	if err := h.collaboratorService.LeaveProject(c.RequestCtx(), projectID, userID); err != nil {
		return utils.RespondError(c, err, "Collaborator membership not found", "Failed to leave project")
	}

	return utils.SendNoContent(c)
}

func (h *CollaboratorHandler) Create(c fiber.Ctx) error {
	userID, err := utils.RequireUserID(c)
	if err != nil {
		return err
	}

	req := new(dto.CreateCollaboratorRequest)
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	collaborator := &models.Collaborator{
		ProjectId: req.ProjectID,
		UserId:    req.UserID,
		Role:      req.Role,
	}

	created, err := h.collaboratorService.Create(c.RequestCtx(), userID, collaborator)
	if err != nil {
		return utils.RespondError(c, err, "Project not found", "Failed to create collaborator")
	}

	return utils.SendJSON(c, fiber.StatusCreated, created)
}

// fiber:context-methods migrated
