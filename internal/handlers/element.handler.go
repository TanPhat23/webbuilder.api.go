package handlers

import (
	"encoding/json"
	"log"
	"my-go-app/internal/database"
	"my-go-app/internal/models"
	"my-go-app/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

type ElementHandler struct {
}

func NewElementHandler() *ElementHandler {
	return &ElementHandler{}
}

func (h *ElementHandler) GetElements(c *fiber.Ctx) error {
	projectID := c.Params("projectid")
	repo := database.GetRepositories()

	elements, err := repo.ElementRepository.GetElements(projectID)
	if err != nil {
		log.Println("Error retrieving elements:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to retrieve elements",
			"errorMessage": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(elements)
}

// CreateElements parses incoming JSON into editor elements, converts each top-level
// item into a models.EditorElement (letting utils.ConvertToEditorElement handle child
// conversion lazily during persistence), and then hands the converted slice to the repository.
func (h *ElementHandler) CreateElements(c *fiber.Ctx) error {
	projectId := c.Params("projectid")
	if projectId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Project ID is required",
			"errorMessage": "Missing projectid parameter in URL",
		})
	}

	body := c.Body()
	if len(body) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Request body is required",
			"errorMessage": "Empty request body",
		})
	}

	// Try to unmarshal into an array of generic values first.
	var rawSlice []any
	if err := json.Unmarshal(body, &rawSlice); err != nil {
		// If not an array, try a single object and wrap it.
		var single any
		if err2 := json.Unmarshal(body, &single); err2 != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":        "Invalid JSON",
				"errorMessage": err2.Error(),
			})
		}
		rawSlice = []any{single}
	}

	var editorElements []models.EditorElement
	for _, item := range rawSlice {
		ee, err := utils.ConvertToEditorElement(item)
		if err != nil {
			log.Println("Error converting item to EditorElement:", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":        "Invalid element structure",
				"errorMessage": err.Error(),
			})
		}
		// Do not set ProjectId here; repository.saveElementRecursive will set it.
		editorElements = append(editorElements, ee)
	}

	repo := database.GetRepositories()
	if err := repo.ElementRepository.CreateElement(editorElements, projectId); err != nil {
		log.Println("Error creating elements:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to create elements",
			"errorMessage": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Elements created successfully",
	})
}
