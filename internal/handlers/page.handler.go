package handlers

import (
	"my-go-app/internal/repositories"

	"github.com/gofiber/fiber/v2"
)

type PageHandler struct {
	pageRepository repositories.PageRepositoryInterface
}

func NewPageHandler(pageRepo repositories.PageRepositoryInterface) *PageHandler {
	return &PageHandler{
		pageRepository: pageRepo,
	}
}

func (h *PageHandler) DeletePage(c *fiber.Ctx) error {
	projectID := c.Params("projectid")
	pageID := c.Params("pageid")

	if projectID == "" || pageID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Project ID and Page ID are required",
			"errorMessage": "Missing projectid or pageid parameter in URL",
		})
	}

	userID, ok := c.Locals("userId").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":        "Unauthorized",
			"errorMessage": "You must be logged in to access this resource",
		})
	}

	err := h.pageRepository.DeletePageByProjectID(pageID, projectID, userID)
	if err != nil {
		if err.Error() == "record not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Page not found or not owned by user",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to delete page",
			"errorMessage": err.Error(),
		})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}
