package handlers

import (
	"log"
	"my-go-app/internal/models"
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
	userID := c.Locals("userId").(string)
	if userID == "" {
		return utils.SendError(c, fiber.StatusUnauthorized, "User not authenticated", nil)
	}

	projectID := c.Params("projectid")
	if projectID == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "Project ID is required", nil)
	}

	// Check if user has access to the project
	_, err := h.projectRepo.GetProjectWithAccess(c.Context(), projectID, userID)
	if err != nil {
		if err.Error() == "project not found" {
			return utils.SendError(c, fiber.StatusNotFound, "Project not found", err, userID)
		}
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	collaborators, err := h.collaboratorRepo.GetCollaboratorsByProject(c.Context(), projectID)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve collaborators", err)
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{"collaborators": collaborators})
}

func (h *CollaboratorHandler) GetCollaboratorByID(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)
	if userID == "" {
		return utils.SendError(c, fiber.StatusUnauthorized, "User not authenticated", nil)
	}

	collaboratorID := c.Params("collaboratorid")
	if collaboratorID == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "Collaborator ID is required", nil)
	}

	collaborator, err := h.collaboratorRepo.GetCollaboratorByID(c.Context(), collaboratorID)
	if err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Collaborator not found", err)
	}

	_, err = h.projectRepo.GetProjectWithAccess(c.Context(), collaborator.ProjectId, userID)
	if err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}
	log.Printf("%v", collaborator)

	return utils.SendJSON(c, fiber.StatusOK, collaborator)
}

func (h *CollaboratorHandler) UpdateCollaboratorRole(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)
	if userID == "" {
		return utils.SendError(c, fiber.StatusUnauthorized, "User not authenticated", nil)
	}

	collaboratorID := c.Params("collaboratorid")
	if collaboratorID == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "Collaborator ID is required", nil)
	}

	var req struct {
		Role models.CollaboratorRole `json:"role" validate:"required"`
	}

	if err := utils.ValidateJSONBody(c, &req); err != nil {
		return err
	}

	// Get collaborator to check project
	collaborator, err := h.collaboratorRepo.GetCollaboratorByID(c.Context(), collaboratorID)
	if err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Collaborator not found", err)
	}

	// Check if user is owner of the project
	project, err := h.projectRepo.GetProjectByID(c.Context(), collaborator.ProjectId, userID)
	if err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Only project owner can update roles", err, userID)
	}

	if project.OwnerId != userID {
		return utils.SendError(c, fiber.StatusForbidden, "Only project owner can update roles", nil, userID)
	}

	err = h.collaboratorRepo.UpdateCollaboratorRole(c.Context(), collaboratorID, req.Role)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to update collaborator role", err)
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{"message": "Role updated successfully"})
}

func (h *CollaboratorHandler) DeleteCollaborator(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)
	if userID == "" {
		return utils.SendError(c, fiber.StatusUnauthorized, "User not authenticated", nil)
	}

	collaboratorID := c.Params("collaboratorid")
	if collaboratorID == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "Collaborator ID is required", nil)
	}

	// Get collaborator to check project
	collaborator, err := h.collaboratorRepo.GetCollaboratorByID(c.Context(), collaboratorID)
	if err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Collaborator not found", err)
	}

	// Check if user is owner of the project or the collaborator themselves
	project, err := h.projectRepo.GetProjectByID(c.Context(), collaborator.ProjectId, userID)
	if err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	if project.OwnerId != userID && collaborator.UserId != userID {
		return utils.SendError(c, fiber.StatusForbidden, "Only project owner or the collaborator can remove access", nil, userID)
	}

	err = h.collaboratorRepo.DeleteCollaborator(c.Context(), collaboratorID)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to delete collaborator", err)
	}

	return utils.SendNoContent(c)
}
