package handlers

import (
	"my-go-app/internal/dto"
	"my-go-app/internal/repositories"
	"my-go-app/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

type CollaboratorHandler struct {
	collaboratorRepo repositories.CollaboratorRepositoryInterface
	projectRepo      repositories.ProjectRepositoryInterface
}

func NewCollaboratorHandler(
	collaboratorRepo repositories.CollaboratorRepositoryInterface,
	projectRepo repositories.ProjectRepositoryInterface,
) *CollaboratorHandler {
	return &CollaboratorHandler{
		collaboratorRepo: collaboratorRepo,
		projectRepo:      projectRepo,
	}
}

func (h *CollaboratorHandler) GetCollaboratorsByProject(c *fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "projectid")
	if err != nil {
		return err
	}
	projectID := ids[0]

	if _, err := h.projectRepo.GetProjectWithAccess(c.Context(), projectID, userID); err != nil {
		return utils.HandleRepoError(c, err, "Project not found", "Access denied")
	}

	collaborators, err := h.collaboratorRepo.GetCollaboratorsByProject(c.Context(), projectID)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to retrieve collaborators")
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{"collaborators": collaborators})
}

func (h *CollaboratorHandler) GetCollaboratorByID(c *fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "collaboratorid")
	if err != nil {
		return err
	}
	collaboratorID := ids[0]

	collaborator, err := h.collaboratorRepo.GetCollaboratorByID(c.Context(), collaboratorID)
	if err != nil {
		return utils.HandleRepoError(c, err, "Collaborator not found", "Failed to retrieve collaborator")
	}

	if _, err := h.projectRepo.GetProjectWithAccess(c.Context(), collaborator.ProjectId, userID); err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	return utils.SendJSON(c, fiber.StatusOK, collaborator)
}

func (h *CollaboratorHandler) UpdateCollaboratorRole(c *fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "collaboratorid")
	if err != nil {
		return err
	}
	collaboratorID := ids[0]

	var req dto.UpdateCollaboratorRoleRequest
	if err := utils.ValidateAndParseBody(c, &req); err != nil {
		return err
	}

	collaborator, err := h.collaboratorRepo.GetCollaboratorByID(c.Context(), collaboratorID)
	if err != nil {
		return utils.HandleRepoError(c, err, "Collaborator not found", "Failed to retrieve collaborator")
	}

	project, err := h.projectRepo.GetProjectByID(c.Context(), collaborator.ProjectId, userID)
	if err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Only project owner can update roles", err, userID)
	}

	if project.OwnerId != userID {
		return utils.SendError(c, fiber.StatusForbidden, "Only project owner can update roles", nil, userID)
	}

	if err := h.collaboratorRepo.UpdateCollaboratorRole(c.Context(), collaboratorID, req.Role); err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to update collaborator role")
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{"message": "Role updated successfully"})
}

func (h *CollaboratorHandler) DeleteCollaborator(c *fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "collaboratorid")
	if err != nil {
		return err
	}
	collaboratorID := ids[0]

	collaborator, err := h.collaboratorRepo.GetCollaboratorByID(c.Context(), collaboratorID)
	if err != nil {
		return utils.HandleRepoError(c, err, "Collaborator not found", "Failed to retrieve collaborator")
	}

	project, err := h.projectRepo.GetProjectByID(c.Context(), collaborator.ProjectId, userID)
	if err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	if project.OwnerId != userID && collaborator.UserId != userID {
		return utils.SendError(c, fiber.StatusForbidden, "Only project owner or the collaborator can remove access", nil, userID)
	}

	if err := h.collaboratorRepo.DeleteCollaborator(c.Context(), collaboratorID); err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to delete collaborator")
	}

	return utils.SendNoContent(c)
}