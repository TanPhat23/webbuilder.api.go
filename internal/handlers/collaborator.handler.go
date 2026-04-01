package handlers

import (
	"my-go-app/internal/dto"
	"my-go-app/internal/models"
	"my-go-app/internal/services"
	"my-go-app/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

type CollaboratorHandler struct {
	collaboratorService services.CollaboratorServiceInterface
}

func NewCollaboratorHandler(collaboratorService services.CollaboratorServiceInterface) *CollaboratorHandler {
	return &CollaboratorHandler{
		collaboratorService: collaboratorService,
	}
}

func (h *CollaboratorHandler) GetCollaboratorsByProject(c *fiber.Ctx) error {
	ids, err := utils.MustParams(c, "projectid")
	if err != nil {
		return err
	}
	projectID := ids[0]

	collaborators, err := h.collaboratorService.GetCollaboratorsByProject(c.Context(), projectID)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to retrieve collaborators")
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{"collaborators": collaborators})
}

func (h *CollaboratorHandler) GetCollaboratorByID(c *fiber.Ctx) error {
	ids, err := utils.MustParams(c, "collaboratorid")
	if err != nil {
		return err
	}
	collaboratorID := ids[0]

	collaborator, err := h.collaboratorService.GetCollaboratorByID(c.Context(), collaboratorID)
	if err != nil {
		return utils.HandleRepoError(c, err, "Collaborator not found", "Failed to retrieve collaborator")
	}

	return utils.SendJSON(c, fiber.StatusOK, collaborator)
}

func (h *CollaboratorHandler) UpdateCollaboratorRole(c *fiber.Ctx) error {
	ids, err := utils.MustParams(c, "collaboratorid")
	if err != nil {
		return err
	}
	collaboratorID := ids[0]

	var req dto.UpdateCollaboratorRoleRequest
	if err := utils.ValidateAndParseBody(c, &req); err != nil {
		return err
	}

	if err := h.collaboratorService.UpdateCollaboratorRole(c.Context(), collaboratorID, req.Role); err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to update collaborator role")
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{"message": "Role updated successfully"})
}

func (h *CollaboratorHandler) DeleteCollaborator(c *fiber.Ctx) error {
	ids, err := utils.MustParams(c, "collaboratorid")
	if err != nil {
		return err
	}
	collaboratorID := ids[0]

	if err := h.collaboratorService.DeleteCollaborator(c.Context(), collaboratorID); err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to delete collaborator")
	}

	return utils.SendNoContent(c)
}

func (h *CollaboratorHandler) CreateCollaborator(c *fiber.Ctx) error {
	var req dto.CreateCollaboratorRequest
	if err := utils.ValidateAndParseBody(c, &req); err != nil {
		return err
	}

	collaborator := &models.Collaborator{
		ProjectId: req.ProjectID,
		UserId:    req.UserID,
		Role:      req.Role,
	}

	created, err := h.collaboratorService.CreateCollaborator(c.Context(), collaborator)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to create collaborator")
	}

	return utils.SendJSON(c, fiber.StatusCreated, created)
}