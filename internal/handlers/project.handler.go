package handlers

import (
	"my-go-app/internal/database"

	"github.com/gofiber/fiber/v2"
)

type ProjectHandler struct{}

func NewProjectHandler() *ProjectHandler {
	return &ProjectHandler{}
}

func (h *ProjectHandler) GetProject(c *fiber.Ctx) error {
	repo := database.GetRepositories()
	projects, err := repo.GetProjects()
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

	repo := database.GetRepositories()
	project, err := repo.GetProjectByID(projectID, userID)
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

	repo := database.GetRepositories()
	pages, err := repo.GetProjectPages(projectID, userID)
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

	repo := database.GetRepositories()
	projects, err := repo.GetProjectsByUserID(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to retrieve projects by user ID",
			"errorMessage": err.Error(),
			"userId":       userID,
		})
	}
	return c.Status(fiber.StatusOK).JSON(projects)
}
