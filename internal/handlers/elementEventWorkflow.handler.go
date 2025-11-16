package handlers

import (
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

type ElementEventWorkflowHandler struct {
	eewRepo      *repositories.ElementEventWorkflowRepository
	elementRepo  repositories.ElementRepositoryInterface
	workflowRepo repositories.EventWorkflowRepositoryInterface
	projectRepo  repositories.ProjectRepositoryInterface
}

func NewElementEventWorkflowHandler(
	eewRepo *repositories.ElementEventWorkflowRepository,
	elementRepo repositories.ElementRepositoryInterface,
	workflowRepo repositories.EventWorkflowRepositoryInterface,
	projectRepo repositories.ProjectRepositoryInterface,
) *ElementEventWorkflowHandler {
	return &ElementEventWorkflowHandler{
		eewRepo:      eewRepo,
		elementRepo:  elementRepo,
		workflowRepo: workflowRepo,
		projectRepo:  projectRepo,
	}
}

// CreateElementEventWorkflow creates a new element event workflow association
func (h *ElementEventWorkflowHandler) CreateElementEventWorkflow(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)
	if userID == "" {
		return utils.SendError(c, fiber.StatusUnauthorized, "User not authenticated", nil)
	}

	var req struct {
		ElementID  string `json:"elementId" validate:"required"`
		WorkflowID string `json:"workflowId" validate:"required"`
		EventName  string `json:"eventName" validate:"required"`
	}

	if err := utils.ValidateJSONBody(c, &req); err != nil {
		return err
	}

	// Get element to verify it exists and check project access
	element, err := h.elementRepo.GetElementByID(c.Context(), req.ElementID)
	if err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Element not found", err)
	}

	// Verify user has access to the project
	_, err = h.projectRepo.GetProjectWithAccess(c.Context(), element.ProjectId, userID)
	if err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied to project", err, userID)
	}

	// Get workflow to verify it exists
	workflow, err := h.workflowRepo.GetEventWorkflowByID(c.Context(), req.WorkflowID)
	if err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Workflow not found", err)
	}

	// Verify workflow belongs to the same project
	if workflow.ProjectId != element.ProjectId {
		return utils.SendError(c, fiber.StatusBadRequest, "Workflow and element must belong to the same project", nil)
	}

	// Check if association already exists
	exists, err := h.eewRepo.CheckIfWorkflowLinkedToElement(c.Context(), req.ElementID, req.WorkflowID, req.EventName)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to check existing association", err)
	}

	if exists {
		return utils.SendError(c, fiber.StatusBadRequest, "This workflow is already linked to this element for this event", nil)
	}

	eew := &models.ElementEventWorkflow{
		ElementId:  req.ElementID,
		WorkflowId: req.WorkflowID,
		EventName:  req.EventName,
	}

	createdEEW, err := h.eewRepo.CreateElementEventWorkflow(c.Context(), eew)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to create element event workflow", err)
	}

	return utils.SendJSON(c, fiber.StatusCreated, createdEEW)
}

// GetElementEventWorkflowByID retrieves a specific element event workflow
func (h *ElementEventWorkflowHandler) GetElementEventWorkflowByID(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)
	if userID == "" {
		return utils.SendError(c, fiber.StatusUnauthorized, "User not authenticated", nil)
	}

	id := c.Params("id")
	if id == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "Element event workflow ID is required", nil)
	}

	eew, err := h.eewRepo.GetElementEventWorkflowByID(c.Context(), id)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve element event workflow", err)
	}

	if eew == nil {
		return utils.SendError(c, fiber.StatusNotFound, "Element event workflow not found", nil)
	}

	// Verify user has access to the project
	element, err := h.elementRepo.GetElementByID(c.Context(), eew.ElementId)
	if err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Element not found", err)
	}

	_, err = h.projectRepo.GetProjectWithAccess(c.Context(), element.ProjectId, userID)
	if err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	return utils.SendJSON(c, fiber.StatusOK, eew)
}

// GetElementEventWorkflows retrieves element event workflows with optional filters
func (h *ElementEventWorkflowHandler) GetElementEventWorkflows(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)
	if userID == "" {
		return utils.SendError(c, fiber.StatusUnauthorized, "User not authenticated", nil)
	}

	elementID := c.Query("elementId")
	workflowID := c.Query("workflowId")
	eventName := c.Query("eventName")

	eews := []models.ElementEventWorkflow{}
	var err error

	if elementID != "" && workflowID == "" && eventName == "" {
		// Get workflows for a specific element
		element, err := h.elementRepo.GetElementByID(c.Context(), elementID)
		if err != nil {
			return utils.SendError(c, fiber.StatusNotFound, "Element not found", err)
		}

		// Verify user has access to the project
		_, err = h.projectRepo.GetProjectWithAccess(c.Context(), element.ProjectId, userID)
		if err != nil {
			return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
		}

		eews, err = h.eewRepo.GetElementEventWorkflowsByElementID(c.Context(), elementID)
	} else if workflowID != "" && elementID == "" && eventName == "" {
		// Get elements for a specific workflow
		workflow, err := h.workflowRepo.GetEventWorkflowByID(c.Context(), workflowID)
		if err != nil {
			return utils.SendError(c, fiber.StatusNotFound, "Workflow not found", err)
		}

		// Verify user has access to the project
		_, err = h.projectRepo.GetProjectWithAccess(c.Context(), workflow.ProjectId, userID)
		if err != nil {
			return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
		}

		eews, err = h.eewRepo.GetElementEventWorkflowsByWorkflowID(c.Context(), workflowID)
	} else if eventName != "" && elementID == "" && workflowID == "" {
		// Get all workflows for a specific event
		eews, err = h.eewRepo.GetElementEventWorkflowsByEventName(c.Context(), eventName)

		// Filter results to only include workflows in projects the user has access to
		var filteredEEWs []models.ElementEventWorkflow
		for _, eew := range eews {
			element, errElement := h.elementRepo.GetElementByID(c.Context(), eew.ElementId)
			if errElement != nil {
				continue
			}

			_, errAccess := h.projectRepo.GetProjectWithAccess(c.Context(), element.ProjectId, userID)
			if errAccess == nil {
				filteredEEWs = append(filteredEEWs, eew)
			}
		}
		eews = filteredEEWs
	} else if elementID != "" || workflowID != "" || eventName != "" {
		// Use filters if any combination is provided
		eews, err = h.eewRepo.GetElementEventWorkflowsByFilters(c.Context(), elementID, workflowID, eventName)

		// Filter results to only include workflows in projects the user has access to
		var filteredEEWs []models.ElementEventWorkflow
		for _, eew := range eews {
			element, errElement := h.elementRepo.GetElementByID(c.Context(), eew.ElementId)
			if errElement != nil {
				continue
			}

			_, errAccess := h.projectRepo.GetProjectWithAccess(c.Context(), element.ProjectId, userID)
			if errAccess == nil {
				filteredEEWs = append(filteredEEWs, eew)
			}
		}
		eews = filteredEEWs
	} else {
		// Get all element event workflows
		eews, err = h.eewRepo.GetAllElementEventWorkflows(c.Context())

		// Filter results to only include workflows in projects the user has access to
		var filteredEEWs []models.ElementEventWorkflow
		for _, eew := range eews {
			element, errElement := h.elementRepo.GetElementByID(c.Context(), eew.ElementId)
			if errElement != nil {
				continue
			}

			_, errAccess := h.projectRepo.GetProjectWithAccess(c.Context(), element.ProjectId, userID)
			if errAccess == nil {
				filteredEEWs = append(filteredEEWs, eew)
			}
		}
		eews = filteredEEWs
	}

	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve element event workflows", err)
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{"data": eews, "count": len(eews)})
}

// UpdateElementEventWorkflow updates an element event workflow
func (h *ElementEventWorkflowHandler) UpdateElementEventWorkflow(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)
	if userID == "" {
		return utils.SendError(c, fiber.StatusUnauthorized, "User not authenticated", nil)
	}

	id := c.Params("id")
	if id == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "Element event workflow ID is required", nil)
	}

	// Get existing eew to verify access
	existingEEW, err := h.eewRepo.GetElementEventWorkflowByID(c.Context(), id)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve element event workflow", err)
	}

	if existingEEW == nil {
		return utils.SendError(c, fiber.StatusNotFound, "Element event workflow not found", nil)
	}

	element, err := h.elementRepo.GetElementByID(c.Context(), existingEEW.ElementId)
	if err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Element not found", err)
	}

	_, err = h.projectRepo.GetProjectWithAccess(c.Context(), element.ProjectId, userID)
	if err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	var req struct {
		EventName string `json:"eventName"`
	}

	if err := utils.ValidateJSONBody(c, &req); err != nil {
		return err
	}

	if req.EventName == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "EventName is required", nil)
	}

	// Check if new association would conflict
	if req.EventName != existingEEW.EventName {
		exists, err := h.eewRepo.CheckIfWorkflowLinkedToElement(c.Context(), existingEEW.ElementId, existingEEW.WorkflowId, req.EventName)
		if err != nil {
			return utils.SendError(c, fiber.StatusInternalServerError, "Failed to check existing association", err)
		}

		if exists {
			return utils.SendError(c, fiber.StatusBadRequest, "This workflow is already linked to this element for this event", nil)
		}
	}

	existingEEW.EventName = req.EventName

	updatedEEW, err := h.eewRepo.UpdateElementEventWorkflow(c.Context(), id, existingEEW)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to update element event workflow", err)
	}

	return utils.SendJSON(c, fiber.StatusOK, updatedEEW)
}

// DeleteElementEventWorkflow deletes an element event workflow
func (h *ElementEventWorkflowHandler) DeleteElementEventWorkflow(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)
	if userID == "" {
		return utils.SendError(c, fiber.StatusUnauthorized, "User not authenticated", nil)
	}

	id := c.Params("id")
	if id == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "Element event workflow ID is required", nil)
	}

	// Get eew to verify access
	eew, err := h.eewRepo.GetElementEventWorkflowByID(c.Context(), id)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve element event workflow", err)
	}

	if eew == nil {
		return utils.SendError(c, fiber.StatusNotFound, "Element event workflow not found", nil)
	}

	element, err := h.elementRepo.GetElementByID(c.Context(), eew.ElementId)
	if err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Element not found", err)
	}

	_, err = h.projectRepo.GetProjectWithAccess(c.Context(), element.ProjectId, userID)
	if err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	err = h.eewRepo.DeleteElementEventWorkflow(c.Context(), id)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to delete element event workflow", err)
	}

	return utils.SendNoContent(c)
}

// DeleteElementEventWorkflowsByElement deletes all event workflows for a specific element
func (h *ElementEventWorkflowHandler) DeleteElementEventWorkflowsByElement(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)
	if userID == "" {
		return utils.SendError(c, fiber.StatusUnauthorized, "User not authenticated", nil)
	}

	elementID := c.Params("elementId")
	if elementID == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "Element ID is required", nil)
	}

	element, err := h.elementRepo.GetElementByID(c.Context(), elementID)
	if err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Element not found", err)
	}

	_, err = h.projectRepo.GetProjectWithAccess(c.Context(), element.ProjectId, userID)
	if err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	err = h.eewRepo.DeleteElementEventWorkflowsByElementID(c.Context(), elementID)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to delete element event workflows", err)
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{"message": "Element event workflows deleted successfully"})
}

// DeleteElementEventWorkflowsByWorkflow deletes all element associations for a specific workflow
func (h *ElementEventWorkflowHandler) DeleteElementEventWorkflowsByWorkflow(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)
	if userID == "" {
		return utils.SendError(c, fiber.StatusUnauthorized, "User not authenticated", nil)
	}

	workflowID := c.Params("workflowId")
	if workflowID == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "Workflow ID is required", nil)
	}

	workflow, err := h.workflowRepo.GetEventWorkflowByID(c.Context(), workflowID)
	if err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Workflow not found", err)
	}

	_, err = h.projectRepo.GetProjectWithAccess(c.Context(), workflow.ProjectId, userID)
	if err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	err = h.eewRepo.DeleteElementEventWorkflowsByWorkflowID(c.Context(), workflowID)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to delete element event workflows", err)
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{"message": "Element event workflows deleted successfully"})
}
