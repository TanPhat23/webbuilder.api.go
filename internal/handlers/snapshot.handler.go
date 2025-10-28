package handlers

import (
	"context"
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
	projectRepo  repositories.ProjectRepositoryInterface
}

func NewSnapshotHandler(snapshotRepo repositories.SnapshotRepositoryInterface, elementRepo repositories.ElementRepositoryInterface, projectRepo repositories.ProjectRepositoryInterface) *SnapshotHandler {
	return &SnapshotHandler{
		snapshotRepo: snapshotRepo,
		elementRepo:  elementRepo,
		projectRepo:  projectRepo,
	}
}

type SaveSnapshotRequest struct {
	Id        string                 `json:"id"`
	Name      string                 `json:"name"`
	Type      string                 `json:"type,omitempty"` // "working" or "version"
	Elements  []any          					`json:"elements"`
	Timestamp int64                  `json:"timestamp,omitempty"` // optional, will use current if not provided
}

func (h *SnapshotHandler) SaveSnapshot(c *fiber.Ctx) error {
	projectId, req, err := h.validateAndParseRequest(c)
	if err != nil {
		return err
	}

	elements, err := h.processElements(req.Elements)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid element structure", err)
	}

	snapshot, err := h.createSnapshot(projectId, req, elements)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to create snapshot", err)
	}

	if err := h.saveAndSyncSnapshot(c.Context(), projectId, snapshot, elements); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to save and sync snapshot", err)
	}

	return utils.SendSuccess(c, fiber.StatusCreated, "Snapshot saved successfully")
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

	snapshotType := "working"
	if req.Type != "" {
		snapshotType = req.Type
	}

	return models.Snapshot{
		Id:        req.Id,
		ProjectId: projectId,
		Name:      req.Name,
		Type:      snapshotType,
		Elements:  elementsJSON,
		Timestamp: timestamp.UnixMilli(),
	}, nil
}

func (h *SnapshotHandler) saveAndSyncSnapshot(ctx context.Context, projectId string, snapshot models.Snapshot, elements []models.EditorElement) error {
	if err := h.snapshotRepo.SaveSnapshot(ctx, projectId, &snapshot); err != nil {
		log.Printf("Error saving snapshot for project %s: %v", projectId, err)
		return err
	}

	log.Printf("About to sync elements for project: %s with %d elements", projectId, len(elements))
	if err := h.elementRepo.ReplaceElements(ctx, projectId, elements); err != nil {
		log.Printf("Error syncing elements with snapshot for project %s: %v", projectId, err)
		return err
	}
	log.Printf("Successfully synced elements for project: %s", projectId)

	return nil
}

func (h *SnapshotHandler) GetSnapshots(c *fiber.Ctx) error {
	projectId, err := utils.ValidateRequiredParam(c, "projectid")
	if err != nil {
		return err
	}

	snapshots, err := h.snapshotRepo.GetSnapshotsByProjectID(c.Context(), projectId)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to get snapshots", err)
	}

	return utils.SendJSON(c, fiber.StatusOK, snapshots)
}

func (h *SnapshotHandler) GetSnapshotByID(c *fiber.Ctx) error {
	snapshotId, err := utils.ValidateRequiredParam(c, "snapshotid")
	if err != nil {
		return err
	}

	snapshot, err := h.snapshotRepo.GetSnapshotByID(c.Context(), snapshotId)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to get snapshot", err)
	}

	return utils.SendJSON(c, fiber.StatusOK, snapshot)
}

func (h *SnapshotHandler) DeleteSnapshot(c *fiber.Ctx) error {
	projectId, err := utils.ValidateRequiredParam(c, "projectid")
	if err != nil {
		return err
	}

	snapshotId, err := utils.ValidateRequiredParam(c, "snapshotid")
	if err != nil {
		return err
	}

	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	// Check if snapshot exists and belongs to the project
	snapshot, err := h.snapshotRepo.GetSnapshotByID(c.Context(), snapshotId)
	if err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Snapshot not found", err)
	}

	if snapshot.ProjectId != projectId {
		return utils.SendError(c, fiber.StatusBadRequest, "Snapshot does not belong to the specified project", nil)
	}

	// Validate that the user owns the project
	_, err = h.projectRepo.GetProjectByID(c.Context(), projectId, userID)
	if err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied or project not found", err)
	}

	// Prevent deleting working snapshots
	if snapshot.Type == "working" {
		return utils.SendError(c, fiber.StatusBadRequest, "Cannot delete working snapshots", nil)
	}

	err = h.snapshotRepo.DeleteSnapshot(c.Context(), snapshotId)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to delete snapshot", err)
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Snapshot deleted successfully")
}
