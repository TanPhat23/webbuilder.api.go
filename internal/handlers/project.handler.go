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
	"authConfig":  "AuthConfig",
}

type ProjectHandler struct {
	projectService services.ProjectServiceInterface
}

func NewProjectHandler(projectService services.ProjectServiceInterface) *ProjectHandler {
	return &ProjectHandler{
		projectService: projectService,
	}
}

func (h *ProjectHandler) ListByUser(c fiber.Ctx) error {
	userID, err := utils.RequireUserID(c)
	if err != nil {
		return err
	}

	projects, err := h.projectService.ListByUser(c.RequestCtx(), userID)
	if err != nil {
		return utils.RespondError(c, err, "", "Failed to retrieve projects")
	}

	return utils.SendJSON(c, fiber.StatusOK, projects)
}

func (h *ProjectHandler) GetByID(c fiber.Ctx) error {
	userID, ids, err := utils.RequireUserAndParams(c, "projectid")
	if err != nil {
		return err
	}
	projectID := ids[0]

	project, err := h.projectService.GetWithAccess(c.RequestCtx(), projectID, userID)
	if err != nil {
		return respondProjectError(c, err, "Failed to retrieve project")
	}

	return utils.SendJSON(c, fiber.StatusOK, project)
}

func (h *ProjectHandler) GetPublicByID(c fiber.Ctx) error {
	ids, err := utils.RequireParams(c, "projectid")
	if err != nil {
		return err
	}
	projectID := ids[0]

	project, err := h.projectService.GetPublicByID(c.RequestCtx(), projectID)
	if err != nil {
		return utils.RespondError(c, err, "Project not found", "Failed to retrieve project")
	}

	return utils.SendJSON(c, fiber.StatusOK, project)
}

func (h *ProjectHandler) GetPublicBySubdomain(c fiber.Ctx) error {
	ids, err := utils.RequireParams(c, "subdomain")
	if err != nil {
		return err
	}

	project, err := h.projectService.GetPublicBySubdomain(c.RequestCtx(), ids[0])
	if err != nil {
		return utils.RespondError(c, err, "Project not found", "Failed to retrieve project")
	}

	return utils.SendJSON(c, fiber.StatusOK, project)
}

func (h *ProjectHandler) ListPages(c fiber.Ctx) error {
	userID, ids, err := utils.RequireUserAndParams(c, "projectid")
	if err != nil {
		return err
	}
	projectID := ids[0]

	// The service verifies access, so a separate check here would repeat the query.
	pages, err := h.projectService.ListPages(c.RequestCtx(), projectID, userID)
	if err != nil {
		return respondProjectError(c, err, "Failed to retrieve project pages")
	}

	return utils.SendJSON(c, fiber.StatusOK, pages)
}

func (h *ProjectHandler) Delete(c fiber.Ctx) error {
	userID, ids, err := utils.RequireUserAndParams(c, "projectid")
	if err != nil {
		return err
	}
	projectID := ids[0]

	if err := h.projectService.Delete(c.RequestCtx(), projectID, userID); err != nil {
		return respondProjectError(c, err, "Failed to delete project")
	}

	return utils.SendNoContent(c)
}

func (h *ProjectHandler) Update(c fiber.Ctx) error {
	userID, ids, err := utils.RequireUserAndParams(c, "projectid")
	if err != nil {
		return err
	}
	projectID := ids[0]

	req := new(dto.UpdateProjectRequest)
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	columnUpdates, err := buildProjectColumnUpdates(req, c.Body())
	if err != nil {
		return err
	}

	updatedProject, err := h.projectService.UpdateFieldsForUser(c.RequestCtx(), projectID, userID, columnUpdates)
	if err != nil {
		return respondProjectError(c, err, "Failed to update project")
	}

	return utils.SendJSON(c, fiber.StatusOK, updatedProject)
}

func (h *ProjectHandler) TransferOwnership(c fiber.Ctx) error {
	userID, ids, err := utils.RequireUserAndParams(c, "projectid")
	if err != nil {
		return err
	}
	projectID := ids[0]

	req := new(dto.TransferOwnershipRequest)
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	if err := h.projectService.TransferOwnership(c.RequestCtx(), projectID, userID, req.NewOwnerID); err != nil {
		return respondProjectError(c, err, "Failed to transfer project ownership")
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Project ownership transferred successfully")
}

func hasExplicitNullField(raw []byte, field string) bool {
	if len(raw) == 0 {
		return false
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return false
	}
	val, exists := m[field]
	return exists && string(val) == "null"
}

func buildProjectColumnUpdates(req *dto.UpdateProjectRequest, body []byte) (map[string]any, error) {
	rawBody := map[string]any{}
	utils.SetIfNotNil(rawBody, "name", req.Name)
	utils.SetIfNotNil(rawBody, "description", req.Description)
	utils.SetIfNotNil(rawBody, "published", req.Published)
	if req.Subdomain != nil {
		normalized, ok, err := models.NormalizeSubdomain(*req.Subdomain)
		if err != nil {
			return nil, fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		if ok {
			rawBody["subdomain"] = normalized
		} else {
			rawBody["subdomain"] = nil
		}
	} else if hasExplicitNullField(body, "subdomain") {
		rawBody["subdomain"] = nil
	}
	if req.Styles != nil {
		rawBody["styles"] = req.Styles
	}
	if req.Header != nil {
		rawBody["header"] = req.Header
	}
	if len(req.AuthConfig) > 0 {
		validated, err := models.ValidateProjectAuthConfig(req.AuthConfig)
		if err != nil {
			return nil, fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		if validated == nil {
			rawBody["authConfig"] = nil
		} else {
			rawBody["authConfig"] = *validated
		}
	} else if hasExplicitNullField(body, "authConfig") {
		rawBody["authConfig"] = nil
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
	assignRawJSONUpdate(columnUpdates, "AuthConfig", &project.AuthConfig)
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

// Create creates a project owned by the caller, with its first page.
func (h *ProjectHandler) Create(c fiber.Ctx) error {
	userID, err := utils.RequireUserID(c)
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
	var normalizedSub *string
	if req.Subdomain != nil {
		sub, ok, err := models.NormalizeSubdomain(*req.Subdomain)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		if ok {
			normalizedSub = &sub
		}
	}

	project := &models.Project{
		ID:          req.Id,
		Name:        req.Name,
		OwnerId:     userID,
		Description: req.Description,
		Subdomain:   normalizedSub,
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

	created, err := h.projectService.CreateWithPage(c.RequestCtx(), project)
	if err != nil {
		return respondProjectError(c, err, "Failed to create project")
	}
	return utils.SendJSON(c, fiber.StatusCreated, created)
}

// GetAccess returns the caller's role on a project. 404 means no
// access (deleted, missing, or not shared), so ids cannot be probed.
func (h *ProjectHandler) GetAccess(c fiber.Ctx) error {
	userID, ids, err := utils.RequireUserAndParams(c, "projectid")
	if err != nil {
		return err
	}
	access, err := h.projectService.GetAccess(c.RequestCtx(), ids[0], userID)
	if errors.Is(err, repositories.ErrProjectNotFound) {
		return utils.SendError(c, fiber.StatusNotFound, "Project not found", err, userID)
	}
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to check project access", err, userID)
	}
	return utils.SendJSON(c, fiber.StatusOK, access)
}
