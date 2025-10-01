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

	// Helper functions for element operations
	func (h *ElementHandler) validateElementID(c *fiber.Ctx) (string, error) {
		elementId := c.Params("elementid")
		if elementId == "" {
			return "", fiber.NewError(fiber.StatusBadRequest, "Element ID is required")
		}
		return elementId, nil
	}

	func (h *ElementHandler) parseElementPayload(c *fiber.Ctx) (models.EditorElement, *string, error) {
		var payload map[string]any
		if err := c.BodyParser(&payload); err != nil {
			return nil, nil, fiber.NewError(fiber.StatusBadRequest, "Invalid request body: "+err.Error())
		}

		// Extract element data
		var rawElement any
		if elem, exists := payload["element"]; exists && elem != nil {
			rawElement = elem
		} else if data, exists := payload["data"]; exists && data != nil {
			rawElement = data
		} else {
			return nil, nil, fiber.NewError(fiber.StatusBadRequest, "Element data is required")
		}

		element, err := utils.ConvertToEditorElement(rawElement)
		if err != nil {
			return nil, nil, fiber.NewError(fiber.StatusBadRequest, "Invalid element structure: "+err.Error())
		}

		// Extract settings
		var settings *string
		if settingsVal, exists := payload["settings"]; exists {
			if settingsStr, ok := settingsVal.(string); ok {
				settings = &settingsStr
			}
		}

		return element, settings, nil
	}

	func (h *ElementHandler) handleRepositoryError(err error, operation string, elementId string) error {
		log.Printf("Error %s element %s: %v", operation, elementId, err)
			
		if err.Error() == "record not found" {
			return fiber.NewError(fiber.StatusNotFound, "Element not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to "+operation+" element: "+err.Error())
	}

	func (h *ElementHandler) UpdateElement(c *fiber.Ctx) error {
		elementId, err := h.validateElementID(c)
		if err != nil {
			return c.Status(err.(*fiber.Error).Code).JSON(fiber.Map{
				"error":        err.Error(),
				"errorMessage": err.Error(),
			})
		}

		var payload map[string]any
		if err := c.BodyParser(&payload); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":        "Invalid request body",
				"errorMessage": err.Error(),
			})
		}

		updates, ok := payload["updates"].(map[string]any)
		if !ok {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":        "Invalid request body",
				"errorMessage": "Missing or invalid 'updates' field",
			})
		}

		element, err := utils.ConvertToEditorElement(updates)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":        "Invalid element structure",
				"errorMessage": err.Error(),
			})
		}



		base := element.GetElement()
		if base == nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":        "Invalid element structure",
				"errorMessage": "Element must have valid base data",
			})
		}
		base.Id = elementId

		if err := h.elementRepo.UpdateElement(element); err != nil {
			repoErr := h.handleRepositoryError(err, "updating", elementId)
			return c.Status(repoErr.(*fiber.Error).Code).JSON(fiber.Map{
				"error":        repoErr.Error(),
				"errorMessage": repoErr.Error(),
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Element updated successfully",
			"data": fiber.Map{
				"id":   elementId,
				"type": base.Type,
			},
		})
	}

	func (h *ElementHandler) DeleteElement(c *fiber.Ctx) error {
		elementId, err := h.validateElementID(c)
		if err != nil {
			return c.Status(err.(*fiber.Error).Code).JSON(fiber.Map{
				"error":        err.Error(),
				"errorMessage": err.Error(),
			})
		}

		if err := h.elementRepo.DeleteElement(elementId); err != nil {
			repoErr := h.handleRepositoryError(err, "deleting", elementId)
			return c.Status(repoErr.(*fiber.Error).Code).JSON(fiber.Map{
				"error":        repoErr.Error(),
				"errorMessage": repoErr.Error(),
			})
		}

		return c.Status(fiber.StatusNoContent).Send(nil)
	}

	func (h *ElementHandler) SwapElements(c *fiber.Ctx) error {
		projectID := c.Params("projectid")

		var payload struct {
			ElementID1 string `json:"elementId1"`
			ElementID2 string `json:"elementId2"`
		}

		if err := c.BodyParser(&payload); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":        "Invalid request body",
				"errorMessage": err.Error(),
			})
		}

		if payload.ElementID1 == "" || payload.ElementID2 == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Element IDs are required",
			})
		}

		if err := h.elementRepo.SwapElements(projectID, payload.ElementID1, payload.ElementID2); err != nil {
			log.Println("Error swapping elements:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":        "Failed to swap elements",
				"errorMessage": err.Error(),
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Elements swapped successfully",
		})
	}
