package handlers

import (
	"encoding/json"
	"my-go-app/internal/repositories"

	"github.com/gofiber/fiber/v2"
)

type ProjectHandler struct{
	projectRepository repositories.ProjectRepositoryInterface
}

func NewProjectHandler(projectRepo repositories.ProjectRepositoryInterface) *ProjectHandler {
	return &ProjectHandler{
		projectRepository: projectRepo,
	}
}

func (h *ProjectHandler) GetProject(c *fiber.Ctx) error {
	projects, err := h.projectRepository.GetProjects()
	userID, _ := c.Locals("userId").(string)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to retrieve projects",
			"errorMessage": err.Error(),
			"userId":       userID,
		})
	}
	return c.Status(fiber.StatusOK).JSON(projects)
}

func (h *ProjectHandler) GetProjectByID(c *fiber.Ctx) error {
	projectID := c.Params("projectid")
	if projectID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Project ID is required",
			"errorMessage": "Missing projectid parameter in URL",
			"userId":       c.Locals("userId"),
		})
	}

	userID, ok := c.Locals("userId").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":        "Unauthorized",
			"errorMessage": "You must be logged in to access this resource",
			"userId":       userID,
		})
	}

	project, err := h.projectRepository.GetProjectByID(projectID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to retrieve project",
			"errorMessage": err.Error(),
			"userId":       userID,
		})
	}
	if project == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":        "Project not found",
			"userId":       userID,
		})
	}
	return c.Status(fiber.StatusOK).JSON(project)
}

func (h *ProjectHandler) GetProjectPages(c *fiber.Ctx) error {
	projectID := c.Params("projectid")
	if projectID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Project ID is required",
			"errorMessage": "Missing projectid parameter in URL",
			"userId":       c.Locals("userId"),
		})
	}

	userID, ok := c.Locals("userId").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":        "Unauthorized",
			"errorMessage": "You must be logged in to access this resource",
			"userId":       userID,
		})
	}

	pages, err := h.projectRepository.GetProjectPages(projectID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to retrieve project pages",
			"errorMessage": err.Error(),
			"userId":       userID,
		})
	}
	return c.Status(fiber.StatusOK).JSON(pages)
}

func (h *ProjectHandler) GetProjectByUserID(c *fiber.Ctx) error {
	userID, ok := c.Locals("userId").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":        "Unauthorized",
			"errorMessage": "You must be logged in to access this resource",
			"userId":       userID,
		})
	}

	projects, err := h.projectRepository.GetProjectsByUserID(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to retrieve projects by user ID",
			"errorMessage": err.Error(),
			"userId":       userID,
		})
	}
	return c.Status(fiber.StatusOK).JSON(projects)
}

func (h *ProjectHandler) DeleteProject(c *fiber.Ctx) error {
	projectID := c.Params("projectid")
	if projectID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Project ID is required",
			"errorMessage": "Missing projectid parameter in URL",
		})
	}

	userID, ok := c.Locals("userId").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":        "Unauthorized",
			"errorMessage": "You must be logged in to access this resource",
		})
	}

	err := h.projectRepository.DeleteProject(projectID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to delete project",
			"errorMessage": err.Error(),
		})
	}
	return c.Status(fiber.StatusNoContent).Send(nil)
}

func (h *ProjectHandler) UpdateProject(c *fiber.Ctx) error {
	projectID := c.Params("projectid")
	if projectID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Project ID is required",
			"errorMessage": "Missing projectid parameter in URL",
		})
	}

	userID, ok := c.Locals("userId").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":        "Unauthorized",
			"errorMessage": "You must be logged in to access this resource",
		})
	}

	var updates map[string]any
	if err := json.Unmarshal(c.Body(), &updates); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Invalid JSON body",
			"errorMessage": err.Error(),
		})
	}

	columnUpdates := make(map[string]any)
	for k, v := range updates {
		switch k {
		case "name":
			columnUpdates["Name"] = v
		case "description":
			columnUpdates["Description"] = v
		case "styles":
			if stylesJSON, err := json.Marshal(v); err == nil {
				columnUpdates["Styles"] = json.RawMessage(stylesJSON)
			} else {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":        "Invalid styles format",
					"errorMessage": err.Error(),
				})
			}
		case "header":
			if headerJSON, err := json.Marshal(v); err == nil {
				columnUpdates["Header"] = json.RawMessage(headerJSON)
			} else {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":        "Invalid header format",
					"errorMessage": err.Error(),
				})
			}
		case "published":
			columnUpdates["Published"] = v
		case "subdomain":
			columnUpdates["Subdomain"] = v
		case "updatedAt":
			columnUpdates["UpdatedAt"] = v
		default:
			columnUpdates[k] = v
		}
	}

	updatedProject, err := h.projectRepository.UpdateProject(projectID, userID, columnUpdates)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to update project",
			"errorMessage": err.Error(),
		})
	}
	if updatedProject == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Project not found or not updated",
		})
	}
	return c.Status(fiber.StatusOK).JSON(updatedProject)
}
