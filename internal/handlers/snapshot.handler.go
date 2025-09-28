package handlers

import (
	"encoding/json"
	"log"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/pkg/utils"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type SnapshotHandler struct {
	snapshotRepo repositories.SnapshotRepositoryInterface
	elementRepo  repositories.ElementRepositoryInterface
}

func NewSnapshotHandler(snapshotRepo repositories.SnapshotRepositoryInterface, elementRepo repositories.ElementRepositoryInterface) *SnapshotHandler {
	return &SnapshotHandler{
		snapshotRepo: snapshotRepo,
		elementRepo:  elementRepo,
	}
}

type SaveSnapshotRequest struct {
	Id        string                 `json:"id"`
	Elements  []any          `json:"elements"`
	Timestamp int64                  `json:"timestamp,omitempty"` // optional, will use current if not provided
}

func (h *SnapshotHandler) SaveSnapshot(c *fiber.Ctx) error {
	projectId, req, err := h.validateAndParseRequest(c)
	if err != nil {
		return err
	}

	elements, err := h.processElements(req.Elements)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Invalid element structure",
			"errorMessage": err.Error(),
		})
	}

	snapshot, err := h.createSnapshot(projectId, req, elements)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to create snapshot",
			"errorMessage": err.Error(),
		})
	}

	if err := h.saveAndSyncSnapshot(projectId, snapshot, elements); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to save and sync snapshot",
			"errorMessage": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Snapshot saved successfully",
	})
}

func (h *SnapshotHandler) validateAndParseRequest(c *fiber.Ctx) (string, SaveSnapshotRequest, error) {
	projectId := c.Params("projectid")
	if projectId == "" {
		return "", SaveSnapshotRequest{}, fiber.NewError(fiber.StatusBadRequest, "Project ID is required")
	}

	var req SaveSnapshotRequest
	if err := c.BodyParser(&req); err != nil {
		return "", SaveSnapshotRequest{}, fiber.NewError(fiber.StatusBadRequest, "Invalid request body: "+err.Error())
	}

	if req.Id == "" {
		req.Id = uuid.NewString()
	}

	return projectId, req, nil
}

func (h *SnapshotHandler) processElements(rawElements []any) ([]models.EditorElement, error) {
	elements := make([]models.EditorElement, len(rawElements))
	for i, e := range rawElements {
		ee, err := utils.ConvertToEditorElement(e)
		if err != nil {
			return nil, err
		}
		elements[i] = ee
	}
	return elements, nil
}

func (h *SnapshotHandler) createSnapshot(projectId string, req SaveSnapshotRequest, elements []models.EditorElement) (models.Snapshot, error) {
	timestamp := time.Now()
	if req.Timestamp != 0 {
		timestamp = time.UnixMilli(req.Timestamp)
	}

	elementsJSON, err := json.Marshal(elements)
	if err != nil {
		return models.Snapshot{}, err
	}

	return models.Snapshot{
		Id:        req.Id,
		ProjectId: projectId,
		Elements:  elementsJSON,
		Timestamp: timestamp.UnixMilli(),
	}, nil
}

func (h *SnapshotHandler) saveAndSyncSnapshot(projectId string, snapshot models.Snapshot, elements []models.EditorElement) error {
	if err := h.snapshotRepo.SaveSnapshot(projectId, snapshot); err != nil {
		log.Printf("Error saving snapshot for project %s: %v", projectId, err)
		return err
	}

	log.Printf("About to sync elements for project: %s with %d elements", projectId, len(elements))
	if err := h.elementRepo.ReplaceElements(projectId, elements); err != nil {
		log.Printf("Error syncing elements with snapshot for project %s: %v", projectId, err)
		return err
	}
	log.Printf("Successfully synced elements for project: %s", projectId)

	return nil
}
