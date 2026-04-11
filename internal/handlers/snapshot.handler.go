package handlers

import (
	"context"
	"encoding/json"
	"log"
	"my-go-app/internal/dto"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"
	"my-go-app/pkg/utils"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type SnapshotHandler struct {
	snapshotService services.SnapshotServiceInterface
	elementRepo     repositories.ElementRepositoryInterface
}

func NewSnapshotHandler(
	snapshotService services.SnapshotServiceInterface,
	elementRepo repositories.ElementRepositoryInterface,
) *SnapshotHandler {
	return &SnapshotHandler{
		snapshotService: snapshotService,
		elementRepo:     elementRepo,
	}
}

func (h *SnapshotHandler) SaveSnapshot(c fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	ids, err := utils.MustParams(c, "projectid")
	if err != nil {
		return err
	}
	projectID := ids[0]

	req := new(dto.SaveSnapshotRequest)
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	if req.Id == "" {
		req.Id = uuid.NewString()
	}

	elements, err := h.processElements(req.Elements)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid element structure", err)
	}

	snapshot, err := h.buildSnapshot(projectID, *req, elements)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to build snapshot", err)
	}

	if err := h.saveSnapshot(c.RequestCtx(), projectID, userID, snapshot, elements); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to save and sync snapshot", err)
	}

	return utils.SendSuccess(c, fiber.StatusCreated, "Snapshot saved successfully")
}

func (h *SnapshotHandler) GetSnapshots(c fiber.Ctx) error {
	ids, err := utils.MustParams(c, "projectid")
	if err != nil {
		return err
	}
	projectID := ids[0]

	snapshots, err := h.snapshotService.GetSnapshotsByProjectID(c.RequestCtx(), projectID)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to get snapshots")
	}

	return utils.SendJSON(c, fiber.StatusOK, snapshots)
}

func (h *SnapshotHandler) GetSnapshotByID(c fiber.Ctx) error {
	ids, err := utils.MustParams(c, "snapshotid")
	if err != nil {
		return err
	}
	snapshotID := ids[0]

	snapshot, err := h.snapshotService.GetSnapshotByID(c.RequestCtx(), snapshotID)
	if err != nil {
		return utils.HandleRepoError(c, err, "Snapshot not found", "Failed to get snapshot")
	}

	return utils.SendJSON(c, fiber.StatusOK, snapshot)
}

func (h *SnapshotHandler) DeleteSnapshot(c fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "projectid", "snapshotid")
	if err != nil {
		return err
	}
	projectID, snapshotID := ids[0], ids[1]

	if err := h.snapshotService.DeleteSnapshotWithAccess(c.RequestCtx(), snapshotID, projectID, userID); err != nil {
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

func (h *SnapshotHandler) saveSnapshot(ctx context.Context, projectID, userID string, snapshot models.Snapshot, elements []models.EditorElement) error {
	if err := h.snapshotService.SaveSnapshot(ctx, projectID, userID, &snapshot); err != nil {
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

// fiber:context-methods migrated
