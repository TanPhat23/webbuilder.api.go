package handlers

import (
	"context"
	"encoding/json"
	"log"
	"my-go-app/internal/dto"
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

func NewSnapshotHandler(
	snapshotRepo repositories.SnapshotRepositoryInterface,
	elementRepo repositories.ElementRepositoryInterface,
	projectRepo repositories.ProjectRepositoryInterface,
) *SnapshotHandler {
	return &SnapshotHandler{
		snapshotRepo: snapshotRepo,
		elementRepo:  elementRepo,
		projectRepo:  projectRepo,
	}
}

func (h *SnapshotHandler) SaveSnapshot(c *fiber.Ctx) error {
	ids, err := utils.MustParams(c, "projectid")
	if err != nil {
		return err
	}
	projectID := ids[0]

	var req dto.SaveSnapshotRequest
	if err := utils.ValidateJSONBody(c, &req); err != nil {
		return err
	}

	if req.Id == "" {
		req.Id = uuid.NewString()
	}

	elements, err := h.processElements(req.Elements)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid element structure", err)
	}

	snapshot, err := h.buildSnapshot(projectID, req, elements)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to build snapshot", err)
	}

	if err := h.saveAndSyncSnapshot(c.Context(), projectID, snapshot, elements); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to save and sync snapshot", err)
	}

	return utils.SendSuccess(c, fiber.StatusCreated, "Snapshot saved successfully")
}

func (h *SnapshotHandler) GetSnapshots(c *fiber.Ctx) error {
	ids, err := utils.MustParams(c, "projectid")
	if err != nil {
		return err
	}
	projectID := ids[0]

	snapshots, err := h.snapshotRepo.GetSnapshotsByProjectID(c.Context(), projectID)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to get snapshots")
	}

	return utils.SendJSON(c, fiber.StatusOK, snapshots)
}

func (h *SnapshotHandler) GetSnapshotByID(c *fiber.Ctx) error {
	ids, err := utils.MustParams(c, "snapshotid")
	if err != nil {
		return err
	}
	snapshotID := ids[0]

	snapshot, err := h.snapshotRepo.GetSnapshotByID(c.Context(), snapshotID)
	if err != nil {
		return utils.HandleRepoError(c, err, "Snapshot not found", "Failed to get snapshot")
	}

	return utils.SendJSON(c, fiber.StatusOK, snapshot)
}

func (h *SnapshotHandler) DeleteSnapshot(c *fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "projectid", "snapshotid")
	if err != nil {
		return err
	}
	projectID, snapshotID := ids[0], ids[1]

	snapshot, err := h.snapshotRepo.GetSnapshotByID(c.Context(), snapshotID)
	if err != nil {
		return utils.HandleRepoError(c, err, "Snapshot not found", "Failed to retrieve snapshot")
	}

	if snapshot.ProjectId != projectID {
		return utils.SendError(c, fiber.StatusBadRequest, "Snapshot does not belong to the specified project", nil)
	}

	if _, err := h.projectRepo.GetProjectByID(c.Context(), projectID, userID); err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied or project not found", err)
	}

	if snapshot.Type == "working" {
		return utils.SendError(c, fiber.StatusBadRequest, "Cannot delete working snapshots", nil)
	}

	if err := h.snapshotRepo.DeleteSnapshot(c.Context(), snapshotID); err != nil {
		return utils.HandleRepoError(c, err, "Snapshot not found", "Failed to delete snapshot")
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Snapshot deleted successfully")
}

// ── private helpers ──────────────────────────────────────────────────────────

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

func (h *SnapshotHandler) buildSnapshot(projectID string, req dto.SaveSnapshotRequest, elements []models.EditorElement) (models.Snapshot, error) {
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
		ProjectId: projectID,
		Name:      req.Name,
		Type:      snapshotType,
		Elements:  elementsJSON,
		Timestamp: timestamp.UnixMilli(),
	}, nil
}

func (h *SnapshotHandler) saveAndSyncSnapshot(ctx context.Context, projectID string, snapshot models.Snapshot, elements []models.EditorElement) error {
	if err := h.snapshotRepo.SaveSnapshot(ctx, projectID, &snapshot); err != nil {
		log.Printf("Error saving snapshot for project %s: %v", projectID, err)
		return err
	}

	log.Printf("Syncing %d elements for project %s", len(elements), projectID)
	if err := h.elementRepo.ReplaceElements(ctx, projectID, elements); err != nil {
		log.Printf("Error syncing elements for project %s: %v", projectID, err)
		return err
	}

	log.Printf("Successfully synced elements for project %s", projectID)
	return nil
}