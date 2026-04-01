package handlers

import (
	"encoding/json"
	"my-go-app/internal/dto"
	"my-go-app/internal/models"
	"my-go-app/internal/services"
	"my-go-app/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

var projectAllowedCols = map[string]string{
	"name":        "Name",
	"description": "Description",
	"published":   "Published",
	"subdomain":   "Subdomain",
	"styles":      "Styles",
	"header":      "Header",
}

type ProjectHandler struct {
	projectService services.ProjectServiceInterface
}

func NewProjectHandler(projectService services.ProjectServiceInterface) *ProjectHandler {
	return &ProjectHandler{
		projectService: projectService,
	}
}

func (h *ProjectHandler) GetProjectsByUser(c *fiber.Ctx) error {
	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	projects, err := h.projectService.GetProjectsByUserID(c.Context(), userID)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to retrieve projects")
	}

	return utils.SendJSON(c, fiber.StatusOK, projects)
}

func (h *ProjectHandler) GetProjectByID(c *fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "projectid")
	if err != nil {
		return err
	}
	projectID := ids[0]

	project, err := h.projectService.GetProjectWithAccess(c.Context(), projectID, userID)
	if err != nil {
		return utils.HandleRepoError(c, err, "Project not found", "Failed to retrieve project")
	}

	return utils.SendJSON(c, fiber.StatusOK, project)
}

func (h *ProjectHandler) GetPublicProjectByID(c *fiber.Ctx) error {
	ids, err := utils.MustParams(c, "projectid")
	if err != nil {
		return err
	}
	projectID := ids[0]

	project, err := h.projectService.GetPublicProjectByID(c.Context(), projectID)
	if err != nil {
		return utils.HandleRepoError(c, err, "Project not found", "Failed to retrieve project")
	}

	return utils.SendJSON(c, fiber.StatusOK, project)
}

func (h *ProjectHandler) GetProjectPages(c *fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "projectid")
	if err != nil {
		return err
	}
	projectID := ids[0]

	if _, err := h.projectService.GetProjectWithAccess(c.Context(), projectID, userID); err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	pages, err := h.projectService.GetProjectPages(c.Context(), projectID)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to retrieve project pages")
	}

	return utils.SendJSON(c, fiber.StatusOK, pages)
}

func (h *ProjectHandler) GetProjectByUserID(c *fiber.Ctx) error {
	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	projects, err := h.projectService.GetProjectsByUserID(c.Context(), userID)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to retrieve projects by user ID")
	}

	return utils.SendJSON(c, fiber.StatusOK, projects)
}

func (h *ProjectHandler) DeleteProject(c *fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "projectid")
	if err != nil {
		return err
	}
	projectID := ids[0]

	if err := h.projectService.DeleteProject(c.Context(), projectID, userID); err != nil {
		return utils.HandleRepoError(c, err, "Project not found", "Failed to delete project")
	}

	return utils.SendNoContent(c)
}

func (h *ProjectHandler) UpdateProject(c *fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "projectid")
	if err != nil {
		return err
	}
	projectID := ids[0]

	var req dto.UpdateProjectRequest
	if err := utils.ValidateAndParseBody(c, &req); err != nil {
		return err
	}

	rawBody := map[string]any{}
	if req.Name != nil        { rawBody["name"] = *req.Name }
	if req.Description != nil { rawBody["description"] = *req.Description }
	if req.Published != nil   { rawBody["published"] = *req.Published }
	if req.Subdomain != nil   { rawBody["subdomain"] = *req.Subdomain }
	if req.Styles != nil      { rawBody["styles"] = req.Styles }
	if req.Header != nil      { rawBody["header"] = req.Header }

	columnUpdates, err := utils.BuildColumnUpdates(rawBody, projectAllowedCols)
	if err != nil {
		return err
	}
	if err := utils.RequireUpdates(columnUpdates); err != nil {
		return err
	}

	project := &models.Project{}

	if name, ok := columnUpdates["Name"]; ok {
		project.Name = name.(string)
	}
	if description, ok := columnUpdates["Description"]; ok {
		desc := description.(string)
		project.Description = &desc
	}
	if published, ok := columnUpdates["Published"]; ok {
		project.Published = published.(bool)
	}
	if subdomain, ok := columnUpdates["Subdomain"]; ok {
		sub := subdomain.(string)
		project.Subdomain = &sub
	}
	if styles, ok := columnUpdates["Styles"]; ok {
		if styleBytes, err := json.Marshal(styles); err == nil {
			msg := json.RawMessage(styleBytes)
			project.Styles = &msg
		}
	}
	if header, ok := columnUpdates["Header"]; ok {
		if headerBytes, err := json.Marshal(header); err == nil {
			msg := json.RawMessage(headerBytes)
			project.Header = &msg
		}
	}

	updatedProject, err := h.projectService.UpdateProject(c.Context(), projectID, userID, project)
	if err != nil {
		return utils.HandleRepoError(c, err, "Project not found", "Failed to update project")
	}

	return utils.SendJSON(c, fiber.StatusOK, updatedProject)
}