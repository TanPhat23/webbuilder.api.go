package handlers

import (
	"encoding/json"
	"errors"
	"my-go-app/internal/dto"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"
	"my-go-app/pkg/utils"

	"github.com/gofiber/fiber/v3"
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

func (h *ProjectHandler) GetProjectsByUser(c fiber.Ctx) error {
	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	projects, err := h.projectService.GetProjectsByUserID(c.RequestCtx(), userID)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to retrieve projects")
	}

	return utils.SendJSON(c, fiber.StatusOK, projects)
}

func (h *ProjectHandler) GetProjectByID(c fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "projectid")
	if err != nil {
		return err
	}
	projectID := ids[0]

	project, err := h.projectService.GetProjectWithAccess(c.RequestCtx(), projectID, userID)
	if err != nil {
		return handleProjectError(c, err, "Failed to retrieve project")
	}

	return utils.SendJSON(c, fiber.StatusOK, project)
}

func (h *ProjectHandler) GetPublicProjectByID(c fiber.Ctx) error {
	ids, err := utils.MustParams(c, "projectid")
	if err != nil {
		return err
	}
	projectID := ids[0]

	project, err := h.projectService.GetPublicProjectByID(c.RequestCtx(), projectID)
	if err != nil {
		return utils.HandleRepoError(c, err, "Project not found", "Failed to retrieve project")
	}

	return utils.SendJSON(c, fiber.StatusOK, project)
}

func (h *ProjectHandler) GetProjectPages(c fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "projectid")
	if err != nil {
		return err
	}
	projectID := ids[0]

	// The service verifies access, so a separate check here would repeat the query.
	pages, err := h.projectService.GetProjectPages(c.RequestCtx(), projectID, userID)
	if err != nil {
		return handleProjectError(c, err, "Failed to retrieve project pages")
	}

	return utils.SendJSON(c, fiber.StatusOK, pages)
}

func (h *ProjectHandler) GetProjectByUserID(c fiber.Ctx) error {
	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	projects, err := h.projectService.GetProjectsByUserID(c.RequestCtx(), userID)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to retrieve projects by user ID")
	}

	return utils.SendJSON(c, fiber.StatusOK, projects)
}

func (h *ProjectHandler) DeleteProject(c fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "projectid")
	if err != nil {
		return err
	}
	projectID := ids[0]

	if err := h.projectService.DeleteProject(c.RequestCtx(), projectID, userID); err != nil {
		return handleProjectError(c, err, "Failed to delete project")
	}

	return utils.SendNoContent(c)
}

func (h *ProjectHandler) UpdateProject(c fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "projectid")
	if err != nil {
		return err
	}
	projectID := ids[0]

	req := new(dto.UpdateProjectRequest)
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	columnUpdates, err := buildProjectColumnUpdates(req)
	if err != nil {
		return err
	}
	project := projectFromColumnUpdates(columnUpdates)

	updatedProject, err := h.projectService.UpdateProject(c.RequestCtx(), projectID, userID, project)
	if err != nil {
		return handleProjectError(c, err, "Failed to update project")
	}

	return utils.SendJSON(c, fiber.StatusOK, updatedProject)
}

func buildProjectColumnUpdates(req *dto.UpdateProjectRequest) (map[string]any, error) {
	rawBody := map[string]any{}
	utils.SetIfNotNil(rawBody, "name", req.Name)
	utils.SetIfNotNil(rawBody, "description", req.Description)
	utils.SetIfNotNil(rawBody, "published", req.Published)
	utils.SetIfNotNil(rawBody, "subdomain", req.Subdomain)
	if req.Styles != nil {
		rawBody["styles"] = req.Styles
	}
	if req.Header != nil {
		rawBody["header"] = req.Header
	}

	columnUpdates, err := utils.BuildColumnUpdates(rawBody, projectAllowedCols)
	if err != nil {
		return nil, err
	}
	if err := utils.RequireUpdates(columnUpdates); err != nil {
		return nil, err
	}
	return columnUpdates, nil
}

func projectFromColumnUpdates(columnUpdates map[string]any) *models.Project {
	project := &models.Project{}
	assignStringUpdate(columnUpdates, "Name", &project.Name)
	assignStringPtrUpdate(columnUpdates, "Description", &project.Description)
	assignStringPtrUpdate(columnUpdates, "Subdomain", &project.Subdomain)
	if published, ok := columnUpdates["Published"].(bool); ok {
		project.Published = published
	}
	assignRawJSONUpdate(columnUpdates, "Styles", &project.Styles)
	assignRawJSONUpdate(columnUpdates, "Header", &project.Header)
	return project
}

func assignStringUpdate(updates map[string]any, key string, dst *string) {
	if v, ok := updates[key].(string); ok {
		*dst = v
	}
}

func assignStringPtrUpdate(updates map[string]any, key string, dst **string) {
	if v, ok := updates[key].(string); ok {
		*dst = &v
	}
}

func assignRawJSONUpdate(updates map[string]any, key string, dst **json.RawMessage) {
	if raw, ok := updates[key]; ok {
		if msg, err := utils.MarshalRawJSON(raw); err == nil {
			*dst = msg
		}
	}
}

// CreateProject creates a project owned by the caller, with its first page.
func (h *ProjectHandler) CreateProject(c fiber.Ctx) error {
	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}
	req := new(dto.CreateProjectRequest)
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	styles := json.RawMessage(`{}`)
	if req.Styles != nil {
		msg, err := utils.MarshalRawJSON(req.Styles)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid styles")
		}
		styles = *msg
	}
	project := &models.Project{
		ID:          req.Id,
		Name:        req.Name,
		OwnerId:     userID,
		Description: req.Description,
		Subdomain:   req.Subdomain,
		Published:   req.Published,
		Styles:      &styles,
	}
	if req.Header != nil {
		msg, err := utils.MarshalRawJSON(req.Header)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid header")
		}
		project.Header = msg
	}

	created, err := h.projectService.CreateProjectWithPage(c.RequestCtx(), project)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to create project")
	}
	return utils.SendJSON(c, fiber.StatusCreated, created)
}

// GetProjectAccess returns the caller's role on a project. 404 means no
// access (deleted, missing, or not shared), so ids cannot be probed.
func (h *ProjectHandler) GetProjectAccess(c fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "projectid")
	if err != nil {
		return err
	}
	access, err := h.projectService.GetProjectAccess(c.RequestCtx(), ids[0], userID)
	if errors.Is(err, repositories.ErrProjectNotFound) {
		return utils.SendError(c, fiber.StatusNotFound, "Project not found", err, userID)
	}
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to check project access", err, userID)
	}
	return utils.SendJSON(c, fiber.StatusOK, access)
}
