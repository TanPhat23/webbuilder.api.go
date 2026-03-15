package handlers

import (
	"encoding/json"
	"my-go-app/internal/dto"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/pkg/utils"
	"time"

	"github.com/gofiber/fiber/v2"
)

type EventWorkflowHandler struct {
	workflowRepo repositories.EventWorkflowRepositoryInterface
	projectRepo  repositories.ProjectRepositoryInterface
	elementRepo  repositories.ElementRepositoryInterface
	eewRepo      repositories.ElementEventWorkflowRepositoryInterface
}

func NewEventWorkflowHandler(
	workflowRepo repositories.EventWorkflowRepositoryInterface,
	projectRepo repositories.ProjectRepositoryInterface,
	elementRepo repositories.ElementRepositoryInterface,
	eewRepo repositories.ElementEventWorkflowRepositoryInterface,
) *EventWorkflowHandler {
	return &EventWorkflowHandler{
		workflowRepo: workflowRepo,
		projectRepo:  projectRepo,
		elementRepo:  elementRepo,
		eewRepo:      eewRepo,
	}
}

func (h *EventWorkflowHandler) CreateEventWorkflow(c *fiber.Ctx) error {
	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	var req dto.CreateEventWorkflowRequest
	if err := utils.ValidateAndParseBody(c, &req); err != nil {
		return err
	}

	if _, err := h.projectRepo.GetProjectWithAccess(c.Context(), req.ProjectID, userID); err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied to project", err, userID)
	}

	exists, err := h.workflowRepo.CheckIfWorkflowNameExists(c.Context(), req.ProjectID, req.Name, "")
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to check workflow name")
	}
	if exists {
		return fiber.NewError(fiber.StatusBadRequest, "Workflow with this name already exists in the project")
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	now := time.Now()

	handlers := req.Handlers
	if len(handlers) == 0 {
		handlers = json.RawMessage("[]")
	}

	canvasData := req.CanvasData
	if len(canvasData) == 0 {
		canvasData = json.RawMessage("{}")
	}

	workflow := &models.EventWorkflow{
		ProjectId:   req.ProjectID,
		Name:        req.Name,
		Description: req.Description,
		CanvasData:  canvasData,
		Handlers:    handlers,
		Enabled:     enabled,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	created, err := h.workflowRepo.CreateEventWorkflow(c.Context(), workflow)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to create event workflow")
	}

	return utils.SendJSON(c, fiber.StatusCreated, created)
}

func (h *EventWorkflowHandler) GetEventWorkflowByID(c *fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "id")
	if err != nil {
		return err
	}
	workflowID := ids[0]

	workflow, err := h.workflowRepo.GetEventWorkflowByID(c.Context(), workflowID)
	if err != nil {
		return utils.HandleRepoError(c, err, "Event workflow not found", "Failed to retrieve event workflow")
	}
	if workflow == nil {
		return fiber.NewError(fiber.StatusNotFound, "Event workflow not found")
	}

	if _, err := h.projectRepo.GetProjectWithAccess(c.Context(), workflow.ProjectId, userID); err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	return utils.SendJSON(c, fiber.StatusOK, workflow)
}

func (h *EventWorkflowHandler) GetEventWorkflowsByProject(c *fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "projectid")
	if err != nil {
		return err
	}
	projectID := ids[0]

	if _, err := h.projectRepo.GetProjectWithAccess(c.Context(), projectID, userID); err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied to project", err, userID)
	}

	var enabledPtr *bool
	if enabled := c.Query("enabled"); enabled == "true" {
		v := true
		enabledPtr = &v
	} else if enabled == "false" {
		v := false
		enabledPtr = &v
	}

	workflows, err := h.workflowRepo.GetEventWorkflowsWithFilters(c.Context(), projectID, enabledPtr, c.Query("search"))
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to retrieve event workflows")
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{"data": workflows, "count": len(workflows)})
}

func (h *EventWorkflowHandler) UpdateEventWorkflow(c *fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "id")
	if err != nil {
		return err
	}
	workflowID := ids[0]

	existing, err := h.workflowRepo.GetEventWorkflowByID(c.Context(), workflowID)
	if err != nil {
		return utils.HandleRepoError(c, err, "Event workflow not found", "Failed to retrieve event workflow")
	}
	if existing == nil {
		return fiber.NewError(fiber.StatusNotFound, "Event workflow not found")
	}

	if _, err := h.projectRepo.GetProjectWithAccess(c.Context(), existing.ProjectId, userID); err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	var req dto.UpdateEventWorkflowRequest
	if err := utils.ValidateAndParseBody(c, &req); err != nil {
		return err
	}

	if req.Name != "" && req.Name != existing.Name {
		exists, err := h.workflowRepo.CheckIfWorkflowNameExists(c.Context(), existing.ProjectId, req.Name, ids[0])
		if err != nil {
			return utils.HandleRepoError(c, err, "", "Failed to check workflow name")
		}
		if exists {
			return fiber.NewError(fiber.StatusBadRequest, "Workflow with this name already exists in the project")
		}
		existing.Name = req.Name
	}

	if req.Description != nil      { existing.Description = req.Description }
	if len(req.CanvasData) > 0     { existing.CanvasData = req.CanvasData }
	if len(req.Handlers) > 0       { existing.Handlers = req.Handlers }
	if req.Enabled != nil          { existing.Enabled = *req.Enabled }

	updated, err := h.workflowRepo.UpdateEventWorkflow(c.Context(), workflowID, existing)
	if err != nil {
		return utils.HandleRepoError(c, err, "Event workflow not found", "Failed to update event workflow")
	}

	return utils.SendJSON(c, fiber.StatusOK, updated)
}

func (h *EventWorkflowHandler) UpdateEventWorkflowEnabled(c *fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "id")
	if err != nil {
		return err
	}
	workflowID := ids[0]

	existing, err := h.workflowRepo.GetEventWorkflowByID(c.Context(), workflowID)
	if err != nil {
		return utils.HandleRepoError(c, err, "Event workflow not found", "Failed to retrieve event workflow")
	}
	if existing == nil {
		return fiber.NewError(fiber.StatusNotFound, "Event workflow not found")
	}

	if _, err := h.projectRepo.GetProjectWithAccess(c.Context(), existing.ProjectId, userID); err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	var req dto.UpdateEventWorkflowEnabledRequest
	if err := utils.ValidateAndParseBody(c, &req); err != nil {
		return err
	}

	if err := h.workflowRepo.UpdateEventWorkflowEnabled(c.Context(), workflowID, *req.Enabled); err != nil {
		return utils.HandleRepoError(c, err, "Event workflow not found", "Failed to update event workflow status")
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{"enabled": *req.Enabled})
}

func (h *EventWorkflowHandler) DeleteEventWorkflow(c *fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "id")
	if err != nil {
		return err
	}
	workflowID := ids[0]

	workflow, err := h.workflowRepo.GetEventWorkflowByID(c.Context(), workflowID)
	if err != nil {
		return utils.HandleRepoError(c, err, "Event workflow not found", "Failed to retrieve event workflow")
	}
	if workflow == nil {
		return fiber.NewError(fiber.StatusNotFound, "Event workflow not found")
	}

	if _, err := h.projectRepo.GetProjectWithAccess(c.Context(), workflow.ProjectId, userID); err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	if err := h.workflowRepo.DeleteEventWorkflow(c.Context(), workflowID); err != nil {
		return utils.HandleRepoError(c, err, "Event workflow not found", "Failed to delete event workflow")
	}

	return utils.SendNoContent(c)
}

func (h *EventWorkflowHandler) GetEventWorkflowElements(c *fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "id")
	if err != nil {
		return err
	}
	workflowID := ids[0]

	workflow, err := h.workflowRepo.GetEventWorkflowByID(c.Context(), workflowID)
	if err != nil {
		return utils.HandleRepoError(c, err, "Event workflow not found", "Failed to retrieve event workflow")
	}
	if workflow == nil {
		return fiber.NewError(fiber.StatusNotFound, "Event workflow not found")
	}

	if _, err := h.projectRepo.GetProjectWithAccess(c.Context(), workflow.ProjectId, userID); err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	elements, err := h.eewRepo.GetElementEventWorkflowsByWorkflowID(c.Context(), workflowID)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to retrieve elements")
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{"data": elements, "count": len(elements)})
}