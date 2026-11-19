package handlers

import (
	"context"
	"errors"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"
	"my-go-app/pkg/utils"
	"strings"

	"github.com/gofiber/fiber/v3"
)

// maxPageIDsPerRequest bounds the IN list a client can make the database build.
const maxPageIDsPerRequest = 100

// PublishedProjectVerifiedKey is set in fiber locals by middleware that already
// confirmed the project is published, so the handler does not look it up again.
const PublishedProjectVerifiedKey = "publishedProjectVerified"

// ProjectAccess is the part of the project repository the element routes need to
// decide who may read a project's elements.
type ProjectAccess interface {
	GetProjectWithAccess(ctx context.Context, projectID, userID string) (*models.Project, error)
	GetPublicProjectByID(ctx context.Context, projectID string) (*models.Project, error)
}

type ElementHandler struct {
	elementService services.ElementServiceInterface
	projects       ProjectAccess
}

func NewElementHandler(elementService services.ElementServiceInterface, projects ProjectAccess) *ElementHandler {
	return &ElementHandler{
		elementService: elementService,
		projects:       projects,
	}
}

// GetElements returns a project's elements to its owner or a collaborator.
func (h *ElementHandler) GetElements(c fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "projectid")
	if err != nil {
		return err
	}
	projectID := ids[0]

	if err := h.requireAccess(c, projectID, userID); err != nil {
		return err
	}

	elements, err := h.elementService.GetElements(c.RequestCtx(), projectID)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to retrieve elements")
	}

	return utils.SendJSON(c, fiber.StatusOK, elements)
}

// GetPublicElements returns a published project's elements to anyone.
func (h *ElementHandler) GetPublicElements(c fiber.Ctx) error {
	projectID, err := utils.ValidateRequiredParam(c, "projectid")
	if err != nil {
		return err
	}
	if err := h.requirePublished(c, projectID); err != nil {
		return err
	}

	elements, err := h.elementService.GetElements(c.RequestCtx(), projectID)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to retrieve elements")
	}

	return utils.SendJSON(c, fiber.StatusOK, elements)
}

// GetElementsByPageIds returns the elements of some pages of one project to its
// owner or a collaborator. Query: projectId, pageIds (comma separated).
func (h *ElementHandler) GetElementsByPageIds(c fiber.Ctx) error {
	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}
	projectID, pageIDs, err := pageIDsFromQuery(c)
	if err != nil {
		return err
	}
	if err := h.requireAccess(c, projectID, userID); err != nil {
		return err
	}

	return h.sendPageElements(c, projectID, pageIDs)
}

// GetPublicElementsByPageIds is GetElementsByPageIds for a published project, open to anyone.
func (h *ElementHandler) GetPublicElementsByPageIds(c fiber.Ctx) error {
	projectID, pageIDs, err := pageIDsFromQuery(c)
	if err != nil {
		return err
	}
	if err := h.requirePublished(c, projectID); err != nil {
		return err
	}

	return h.sendPageElements(c, projectID, pageIDs)
}

func (h *ElementHandler) sendPageElements(c fiber.Ctx, projectID string, pageIDs []string) error {
	elements, err := h.elementService.GetElementsByPageIds(c.RequestCtx(), projectID, pageIDs)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to retrieve elements")
	}

	return utils.SendJSON(c, fiber.StatusOK, elements)
}

// requireAccess returns a fiber error unless userID owns or collaborates on the project.
// It returns the error instead of writing the response, so callers cannot mistake
// "already answered" for success.
func (h *ElementHandler) requireAccess(c fiber.Ctx, projectID, userID string) error {
	_, err := h.projects.GetProjectWithAccess(c.RequestCtx(), projectID, userID)
	switch {
	case err == nil:
		return nil
	case errors.Is(err, repositories.ErrProjectUnauthorized):
		return fiber.NewError(fiber.StatusForbidden, "You do not have access to this project")
	case errors.Is(err, repositories.ErrProjectNotFound):
		return fiber.NewError(fiber.StatusNotFound, "Project not found")
	default:
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to verify project access")
	}
}

// requirePublished returns a fiber error unless the project is published.
func (h *ElementHandler) requirePublished(c fiber.Ctx, projectID string) error {
	if verified, _ := c.Locals(PublishedProjectVerifiedKey).(bool); verified {
		return nil
	}
	project, err := h.projects.GetPublicProjectByID(c.RequestCtx(), projectID)
	switch {
	case errors.Is(err, repositories.ErrProjectNotFound):
		return fiber.NewError(fiber.StatusNotFound, "Project not found")
	case err != nil:
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to verify publication")
	case project == nil || !project.Published:
		return fiber.NewError(fiber.StatusNotFound, "Project not found")
	}
	return nil
}

func pageIDsFromQuery(c fiber.Ctx) (string, []string, error) {
	projectID := strings.TrimSpace(c.Query("projectId"))
	if projectID == "" {
		return "", nil, fiber.NewError(fiber.StatusBadRequest, "projectId query parameter is required")
	}

	pageIdsParam := c.Query("pageIds")
	if pageIdsParam == "" {
		return "", nil, fiber.NewError(fiber.StatusBadRequest, "pageIds query parameter is required")
	}

	var pageIDs []string
	for _, id := range strings.Split(pageIdsParam, ",") {
		if trimmed := strings.TrimSpace(id); trimmed != "" {
			pageIDs = append(pageIDs, trimmed)
		}
	}

	if len(pageIDs) == 0 {
		return "", nil, fiber.NewError(fiber.StatusBadRequest, "At least one valid pageId is required")
	}
	if len(pageIDs) > maxPageIDsPerRequest {
		return "", nil, fiber.NewError(fiber.StatusBadRequest, "Too many pageIds")
	}

	return projectID, pageIDs, nil
}
