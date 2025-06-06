package handlers

import (
	"log"
	"my-go-app/internal/database"

	"github.com/gofiber/fiber/v2"
)

type ElmentHandler struct {
}

func NewElementHandler() *ElmentHandler {
	return &ElmentHandler{}
}

func (h *ElmentHandler) GetElements(c *fiber.Ctx) error {
	projectID := c.Params("projectid")
	repo := database.GetRepositories()

	elements, err := repo.ElementRepository.GetElements(projectID)
	if err != nil {
		log.Println("Error retrieving elements:", err)
		log.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve elements",
			"errorMessage": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(elements)
}
