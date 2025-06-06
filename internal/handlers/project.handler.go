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
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to retrieve projects",
			"errorMessage": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(projects)
}

func (h *ProjectHandler) GetProjectByID(c *fiber.Ctx) error {
	projectID := c.Params("projectid")

	repo := database.GetRepositories()
	authUserId := c.Get("userId")
	if authUserId == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":        "Unauthorized",
			"errorMessage": "You must be logged in to access this resource",
		})
	}
	project, err := repo.GetProjectByID(projectID, authUserId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to retrieve project",
			"errorMessage": err.Error(),
		})
	}
	if project == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Project not found",
		})
	}
	return c.Status(fiber.StatusOK).JSON(project)
}

func (h *ProjectHandler) GetProjectByUserID(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)
	repo := database.GetRepositories()

	projects, err := repo.GetProjectsByUserID(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to retrieve projects by user ID",
			"errorMessage": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(projects)
}
