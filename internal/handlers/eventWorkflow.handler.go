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
	userID := c.Locals("userId").(string)
	if userID == "" {
		return utils.SendError(c, fiber.StatusUnauthorized, "User not authenticated", nil)
	}

	var req struct {
		ProjectID   string           `json:"projectId" validate:"required"`
		Name        string           `json:"name" validate:"required"`
		Description *string          `json:"description"`
		CanvasData  json.RawMessage  `json:"canvasData"`
		Enabled     *bool            `json:"enabled"`
	}

	if err := utils.ValidateJSONBody(c, &req); err != nil {
		return err
	}

	// Verify user has access to the project
	_, err := h.projectRepo.GetProjectWithAccess(c.Context(), req.ProjectID, userID)
	if err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied to project", err, userID)
	}

	// Check if workflow name already exists
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

// GetEventWorkflowByID retrieves a specific event workflow
func (h *EventWorkflowHandler) GetEventWorkflowByID(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)
	if userID == "" {
		return utils.SendError(c, fiber.StatusUnauthorized, "User not authenticated", nil)
	}

	id := c.Params("id")
	if id == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "Event workflow ID is required", nil)
	}

	workflow, err := h.workflowRepo.GetEventWorkflowByID(c.Context(), id)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve event workflow", err)
	}

	if workflow == nil {
		return utils.SendError(c, fiber.StatusNotFound, "Event workflow not found", nil)
	}

	// Verify user has access to the project
	_, err = h.projectRepo.GetProjectWithAccess(c.Context(), workflow.ProjectId, userID)
	if err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	return utils.SendJSON(c, fiber.StatusOK, workflow)
}

// GetEventWorkflowsByProject retrieves all event workflows for a project
func (h *EventWorkflowHandler) GetEventWorkflowsByProject(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)
	if userID == "" {
		return utils.SendError(c, fiber.StatusUnauthorized, "User not authenticated", nil)
	}

	projectID := c.Params("projectid")
	if projectID == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "Project ID is required", nil)
	}

	// Verify user has access to the project
	_, err := h.projectRepo.GetProjectWithAccess(c.Context(), projectID, userID)
	if err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied to project", err, userID)
	}

	enabled := c.Query("enabled")
	searchName := c.Query("search")

	var enabledPtr *bool
	if enabled != "" {
		if enabled == "true" {
			trueVal := true
			enabledPtr = &trueVal
		} else if enabled == "false" {
			falseVal := false
			enabledPtr = &falseVal
		}
	}

	workflows, err := h.workflowRepo.GetEventWorkflowsWithFilters(c.Context(), projectID, enabledPtr, searchName)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve event workflows", err)
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{"data": workflows, "count": len(workflows)})
}

// UpdateEventWorkflow updates an event workflow
func (h *EventWorkflowHandler) UpdateEventWorkflow(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)
	if userID == "" {
		return utils.SendError(c, fiber.StatusUnauthorized, "User not authenticated", nil)
	}

	id := c.Params("id")
	if id == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "Event workflow ID is required", nil)
	}

	// Get existing workflow to verify access
	existingWorkflow, err := h.workflowRepo.GetEventWorkflowByID(c.Context(), id)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve event workflow", err)
	}

	if existingWorkflow == nil {
		return utils.SendError(c, fiber.StatusNotFound, "Event workflow not found", nil)
	}

	_, err = h.projectRepo.GetProjectWithAccess(c.Context(), existingWorkflow.ProjectId, userID)
	if err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	var req struct {
		Name        string           `json:"name"`
		Description *string          `json:"description"`
		CanvasData  json.RawMessage  `json:"canvasData"`
		Enabled     *bool            `json:"enabled"`
	}

	if err := utils.ValidateJSONBody(c, &req); err != nil {
		return err
	}

	// Check if new name conflicts
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

// UpdateEventWorkflowEnabled updates the enabled status of an event workflow
func (h *EventWorkflowHandler) UpdateEventWorkflowEnabled(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)
	if userID == "" {
		return utils.SendError(c, fiber.StatusUnauthorized, "User not authenticated", nil)
	}

	id := c.Params("id")
	if id == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "Event workflow ID is required", nil)
	}

	// Get existing workflow to verify access
	existingWorkflow, err := h.workflowRepo.GetEventWorkflowByID(c.Context(), id)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve event workflow", err)
	}

	if existingWorkflow == nil {
		return utils.SendError(c, fiber.StatusNotFound, "Event workflow not found", nil)
	}

	_, err = h.projectRepo.GetProjectWithAccess(c.Context(), existingWorkflow.ProjectId, userID)
	if err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	var req struct {
		Enabled bool `json:"enabled" validate:"required"`
	}

	if err := utils.ValidateJSONBody(c, &req); err != nil {
		return err
	}

	err = h.workflowRepo.UpdateEventWorkflowEnabled(c.Context(), id, req.Enabled)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to update event workflow status", err)
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{"enabled": req.Enabled})
}

// DeleteEventWorkflow deletes an event workflow
func (h *EventWorkflowHandler) DeleteEventWorkflow(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)
	if userID == "" {
		return utils.SendError(c, fiber.StatusUnauthorized, "User not authenticated", nil)
	}

	id := c.Params("id")
	if id == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "Event workflow ID is required", nil)
	}

	// Get workflow to verify access
	workflow, err := h.workflowRepo.GetEventWorkflowByID(c.Context(), id)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve event workflow", err)
	}

	if workflow == nil {
		return utils.SendError(c, fiber.StatusNotFound, "Event workflow not found", nil)
	}

	_, err = h.projectRepo.GetProjectWithAccess(c.Context(), workflow.ProjectId, userID)
	if err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	err = h.workflowRepo.DeleteEventWorkflow(c.Context(), id)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to delete event workflow", err)
	}

	return utils.SendNoContent(c)
}

// GetEventWorkflowElements retrieves all elements linked to a workflow
func (h *EventWorkflowHandler) GetEventWorkflowElements(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)
	if userID == "" {
		return utils.SendError(c, fiber.StatusUnauthorized, "User not authenticated", nil)
	}

	id := c.Params("id")
	if id == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "Event workflow ID is required", nil)
	}

	// Get workflow to verify access
	workflow, err := h.workflowRepo.GetEventWorkflowByID(c.Context(), id)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve event workflow", err)
	}

	if workflow == nil {
		return utils.SendError(c, fiber.StatusNotFound, "Event workflow not found", nil)
	}

	_, err = h.projectRepo.GetProjectWithAccess(c.Context(), workflow.ProjectId, userID)
	if err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	// Get element event workflows
	elements, err := h.eewRepo.GetElementEventWorkflowsByWorkflowID(c.Context(), id)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve elements", err)
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{"data": elements, "count": len(elements)})
}
