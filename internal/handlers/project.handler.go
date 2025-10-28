package handlers

import (
	"encoding/json"
	"my-go-app/internal/repositories"
	"my-go-app/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

type ProjectHandler struct {
	projectRepository repositories.ProjectRepositoryInterface
}

func NewProjectHandler(projectRepo repositories.ProjectRepositoryInterface) *ProjectHandler {
	return &ProjectHandler{
		projectRepository: projectRepo,
	}
}

func (h *ProjectHandler) GetProject(c *fiber.Ctx) error {
	userID, _ := c.Locals("userId").(string)
	projects, err := h.projectRepository.GetProjects(c.Context())
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve projects", err, userID)
	}
	return utils.SendJSON(c, fiber.StatusOK, projects)
}

func (h *ProjectHandler) GetProjectByID(c *fiber.Ctx) error {
	projectID, err := utils.ValidateRequiredParam(c, "projectid")
	if err != nil {
		return err
	}

	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	project, err := h.projectRepository.GetProjectWithAccess(c.Context(), projectID, userID)
	if err != nil {
		if err.Error() == "project not found" || err.Error() == "project unauthorized" {
			return utils.SendError(c, fiber.StatusNotFound, "Project not found", err, userID)
		}
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve project", err, userID)
	}
	return utils.SendJSON(c, fiber.StatusOK, project)
}

func (h *ProjectHandler) GetPublicProjectByID(c *fiber.Ctx) error {
	projectID, err := utils.ValidateRequiredParam(c, "projectid")
	if err != nil {
		return err
	}

	project, err := h.projectRepository.GetPublicProjectByID(c.Context(), projectID)
	if err != nil {
		if err.Error() == "project not found" {
			return utils.SendError(c, fiber.StatusNotFound, "Project not found", err, "")
		}
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve project", err, "")
	}
	return utils.SendJSON(c, fiber.StatusOK, project)
}

func (h *ProjectHandler) GetProjectPages(c *fiber.Ctx) error {
	projectID, err := utils.ValidateRequiredParam(c, "projectid")
	if err != nil {
		return err
	}

	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	// Check access first
	_, err = h.projectRepository.GetProjectWithAccess(c.Context(), projectID, userID)
	if err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	pages, err := h.projectRepository.GetProjectPages(c.Context(), projectID, userID)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve project pages", err, userID)
	}
	return utils.SendJSON(c, fiber.StatusOK, pages)
}

func (h *ProjectHandler) GetProjectByUserID(c *fiber.Ctx) error {
	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	projects, err := h.projectRepository.GetProjectsByUserID(c.Context(), userID)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve projects by user ID", err, userID)
	}
	return utils.SendJSON(c, fiber.StatusOK, projects)
}

func (h *ProjectHandler) DeleteProject(c *fiber.Ctx) error {
	projectID, err := utils.ValidateRequiredParam(c, "projectid")
	if err != nil {
		return err
	}

	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	err = h.projectRepository.DeleteProject(c.Context(), projectID, userID)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to delete project", err)
	}
	return utils.SendNoContent(c)
}

func (h *ProjectHandler) UpdateProject(c *fiber.Ctx) error {
	projectID, err := utils.ValidateRequiredParam(c, "projectid")
	if err != nil {
		return err
	}

	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	var updates map[string]any
	if err := utils.ValidateJSONBody(c, &updates); err != nil {
		return err
	}

	columnUpdates, err := h.buildColumnUpdates(updates)
	if err != nil {
		return err
	}

	updatedProject, err := h.projectRepository.UpdateProject(c.Context(), projectID, userID, columnUpdates)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to update project", err)
	}

	return utils.SendJSON(c, fiber.StatusOK, updatedProject)
}

func (h *ProjectHandler) buildColumnUpdates(updates map[string]any) (map[string]any, error) {
	columnUpdates := make(map[string]any)
	for k, v := range updates {
		switch k {
		case "name":
			columnUpdates["Name"] = v
		case "description":
			columnUpdates["Description"] = v
		case "styles":
			stylesJSON, err := json.Marshal(v)
			if err != nil {
				return nil, fiber.NewError(fiber.StatusBadRequest, "Invalid styles format: "+err.Error())
			}
			columnUpdates["Styles"] = json.RawMessage(stylesJSON)
		case "header":
			headerJSON, err := json.Marshal(v)
			if err != nil {
				return nil, fiber.NewError(fiber.StatusBadRequest, "Invalid header format: "+err.Error())
			}
			columnUpdates["Header"] = json.RawMessage(headerJSON)
		case "published":
			columnUpdates["Published"] = v
		case "subdomain":
			columnUpdates["Subdomain"] = v
		default:
			columnUpdates[k] = v
		}
	}
	return columnUpdates, nil
}
