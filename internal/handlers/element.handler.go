package handlers

import (
	"encoding/json"
	"log"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

type ElementHandler struct {
	elementRepo repositories.ElementRepositoryInterface
}

func NewElementHandler(elementRepo repositories.ElementRepositoryInterface) *ElementHandler {
	return &ElementHandler{
		elementRepo: elementRepo,
	}
}

func (h *ElementHandler) GetElements(c *fiber.Ctx) error {
	projectID := c.Params("projectid")

	elements, err := h.elementRepo.GetElements(projectID)
	if err != nil {
		log.Println("Error retrieving elements:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to retrieve elements",
			"errorMessage": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(elements)
}

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

	var rawSlice []any
	if err := json.Unmarshal(body, &rawSlice); err != nil {
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
		editorElements = append(editorElements, ee)
	}

	if err := h.elementRepo.CreateElement(editorElements, projectId); err != nil {
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

func (h *ElementHandler) InsertElementAfter(c *fiber.Ctx) error {
	projectId := c.Params("projectid")
	previousElementId := c.Params("previouselementid")
	if projectId == "" || previousElementId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Project ID and Previous Element ID are required",
			"errorMessage": "Missing projectid or previouselementid parameter in URL",
		})
	}

	var rawElement any
	if err := c.BodyParser(&rawElement); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Invalid request body",
			"errorMessage": err.Error(),
		})
	}

	newElement, err := utils.ConvertToEditorElement(rawElement)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Invalid element structure",
			"errorMessage": err.Error(),
		})
	}

	if err := h.elementRepo.InsertElementAfter(projectId, previousElementId, newElement); err != nil {
		log.Println("Error inserting element:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to insert element",
			"errorMessage": err.Error(),
		})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Element inserted successfully",
	})
}
