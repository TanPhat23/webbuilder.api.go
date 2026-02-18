package handlers

import (
	"encoding/json"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/pkg/utils"
	"time"

	"github.com/gofiber/fiber/v2"
)

type EventWorkflowHandler struct {
	workflowRepo *repositories.EventWorkflowRepository
	projectRepo  repositories.ProjectRepositoryInterface
	elementRepo  repositories.ElementRepositoryInterface
	eewRepo      *repositories.ElementEventWorkflowRepository
}

func NewEventWorkflowHandler(
	workflowRepo *repositories.EventWorkflowRepository,
	projectRepo repositories.ProjectRepositoryInterface,
	elementRepo repositories.ElementRepositoryInterface,
	eewRepo *repositories.ElementEventWorkflowRepository,
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

	var req struct {
		ProjectID   string          `json:"projectId"   validate:"required"`
		Name        string          `json:"name"        validate:"required"`
		Description *string         `json:"description"`
		CanvasData  json.RawMessage `json:"canvasData"`
		Enabled     *bool           `json:"enabled"`
	}

	if err := utils.ValidateAndParseBody(c, &req); err != nil {
		return err
	}

	if _, err := h.projectRepo.GetProjectWithAccess(c.Context(), req.ProjectID, userID); err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied to project", err, userID)
	}

	exists, err := h.workflowRepo.CheckIfWorkflowNameExists(c.Context(), req.ProjectID, req.Name, "")
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to check workflow name", err)
	}
	if exists {
		return utils.SendError(c, fiber.StatusBadRequest, "Workflow with this name already exists in the project", nil)
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	now := time.Now()
	workflow := &models.EventWorkflow{
		ProjectId:   req.ProjectID,
		Name:        req.Name,
		Description: req.Description,
		CanvasData:  req.CanvasData,
		Enabled:     enabled,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	createdWorkflow, err := h.workflowRepo.CreateEventWorkflow(c.Context(), workflow)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to create event workflow", err)
	}

	return utils.SendJSON(c, fiber.StatusCreated, createdWorkflow)
}

// GetEventWorkflowByID retrieves a specific event workflow.
func (h *EventWorkflowHandler) GetEventWorkflowByID(c *fiber.Ctx) error {
	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	id, err := utils.ValidateRequiredParam(c, "id")
	if err != nil {
		return err
	}

	workflow, err := h.workflowRepo.GetEventWorkflowByID(c.Context(), id)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve event workflow", err)
	}
	if workflow == nil {
		return utils.SendError(c, fiber.StatusNotFound, "Event workflow not found", nil)
	}

	if _, err := h.projectRepo.GetProjectWithAccess(c.Context(), workflow.ProjectId, userID); err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	return utils.SendJSON(c, fiber.StatusOK, workflow)
}

// GetEventWorkflowsByProject retrieves all event workflows for a project.
func (h *EventWorkflowHandler) GetEventWorkflowsByProject(c *fiber.Ctx) error {
	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	projectID, err := utils.ValidateRequiredParam(c, "projectid")
	if err != nil {
		return err
	}

	if _, err := h.projectRepo.GetProjectWithAccess(c.Context(), projectID, userID); err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied to project", err, userID)
	}

	enabled := c.Query("enabled")
	searchName := c.Query("search")

	var enabledPtr *bool
	if enabled == "true" {
		v := true
		enabledPtr = &v
	} else if enabled == "false" {
		v := false
		enabledPtr = &v
	}

	workflows, err := h.workflowRepo.GetEventWorkflowsWithFilters(c.Context(), projectID, enabledPtr, searchName)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve event workflows", err)
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{"data": workflows, "count": len(workflows)})
}

// UpdateEventWorkflow updates an event workflow.
func (h *EventWorkflowHandler) UpdateEventWorkflow(c *fiber.Ctx) error {
	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	id, err := utils.ValidateRequiredParam(c, "id")
	if err != nil {
		return err
	}

	existingWorkflow, err := h.workflowRepo.GetEventWorkflowByID(c.Context(), id)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve event workflow", err)
	}
	if existingWorkflow == nil {
		return utils.SendError(c, fiber.StatusNotFound, "Event workflow not found", nil)
	}

	if _, err := h.projectRepo.GetProjectWithAccess(c.Context(), existingWorkflow.ProjectId, userID); err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	var req struct {
		Name        string          `json:"name"`
		Description *string         `json:"description"`
		CanvasData  json.RawMessage `json:"canvasData"`
		Enabled     *bool           `json:"enabled"`
	}

	if err := utils.ValidateAndParseBody(c, &req); err != nil {
		return err
	}

	if req.Name != "" && req.Name != existingWorkflow.Name {
		exists, err := h.workflowRepo.CheckIfWorkflowNameExists(c.Context(), existingWorkflow.ProjectId, req.Name, id)
		if err != nil {
			return utils.SendError(c, fiber.StatusInternalServerError, "Failed to check workflow name", err)
		}
		if exists {
			return utils.SendError(c, fiber.StatusBadRequest, "Workflow with this name already exists in the project", nil)
		}
		existingWorkflow.Name = req.Name
	}

	if req.Description != nil {
		existingWorkflow.Description = req.Description
	}
	if len(req.CanvasData) > 0 {
		existingWorkflow.CanvasData = req.CanvasData
	}
	if req.Enabled != nil {
		existingWorkflow.Enabled = *req.Enabled
	}

	updatedWorkflow, err := h.workflowRepo.UpdateEventWorkflow(c.Context(), id, existingWorkflow)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to update event workflow", err)
	}

	return utils.SendJSON(c, fiber.StatusOK, updatedWorkflow)
}

// UpdateEventWorkflowEnabled toggles the enabled status of an event workflow.
func (h *EventWorkflowHandler) UpdateEventWorkflowEnabled(c *fiber.Ctx) error {
	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	id, err := utils.ValidateRequiredParam(c, "id")
	if err != nil {
		return err
	}

	existingWorkflow, err := h.workflowRepo.GetEventWorkflowByID(c.Context(), id)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve event workflow", err)
	}
	if existingWorkflow == nil {
		return utils.SendError(c, fiber.StatusNotFound, "Event workflow not found", nil)
	}

	if _, err := h.projectRepo.GetProjectWithAccess(c.Context(), existingWorkflow.ProjectId, userID); err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	var req struct {
		Enabled bool `json:"enabled" validate:"required"`
	}

	if err := utils.ValidateAndParseBody(c, &req); err != nil {
		return err
	}

	if err := h.workflowRepo.UpdateEventWorkflowEnabled(c.Context(), id, req.Enabled); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to update event workflow status", err)
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{"enabled": req.Enabled})
}

// DeleteEventWorkflow deletes an event workflow.
func (h *EventWorkflowHandler) DeleteEventWorkflow(c *fiber.Ctx) error {
	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	id, err := utils.ValidateRequiredParam(c, "id")
	if err != nil {
		return err
	}

	workflow, err := h.workflowRepo.GetEventWorkflowByID(c.Context(), id)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve event workflow", err)
	}
	if workflow == nil {
		return utils.SendError(c, fiber.StatusNotFound, "Event workflow not found", nil)
	}

	if _, err := h.projectRepo.GetProjectWithAccess(c.Context(), workflow.ProjectId, userID); err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	if err := h.workflowRepo.DeleteEventWorkflow(c.Context(), id); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to delete event workflow", err)
	}

	return utils.SendNoContent(c)
}

// GetEventWorkflowElements retrieves all elements linked to a workflow.
func (h *EventWorkflowHandler) GetEventWorkflowElements(c *fiber.Ctx) error {
	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	id, err := utils.ValidateRequiredParam(c, "id")
	if err != nil {
		return err
	}

	workflow, err := h.workflowRepo.GetEventWorkflowByID(c.Context(), id)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve event workflow", err)
	}
	if workflow == nil {
		return utils.SendError(c, fiber.StatusNotFound, "Event workflow not found", nil)
	}

	if _, err := h.projectRepo.GetProjectWithAccess(c.Context(), workflow.ProjectId, userID); err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	elements, err := h.eewRepo.GetElementEventWorkflowsByWorkflowID(c.Context(), id)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve elements", err)
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{"data": elements, "count": len(elements)})
}
