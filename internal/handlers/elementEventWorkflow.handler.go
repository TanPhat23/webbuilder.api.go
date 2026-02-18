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
	pageRepo     repositories.PageRepositoryInterface
}

func NewElementEventWorkflowHandler(
	eewRepo *repositories.ElementEventWorkflowRepository,
	elementRepo repositories.ElementRepositoryInterface,
	workflowRepo repositories.EventWorkflowRepositoryInterface,
	projectRepo repositories.ProjectRepositoryInterface,
	pageRepo repositories.PageRepositoryInterface,
) *ElementEventWorkflowHandler {
	return &ElementEventWorkflowHandler{
		eewRepo:      eewRepo,
		elementRepo:  elementRepo,
		workflowRepo: workflowRepo,
		projectRepo:  projectRepo,
		pageRepo:     pageRepo,
	}
}

// CreateElementEventWorkflow creates a new element event workflow association.
func (h *ElementEventWorkflowHandler) CreateElementEventWorkflow(c *fiber.Ctx) error {
	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	var req struct {
		ElementID  string `json:"elementId"  validate:"required"`
		WorkflowID string `json:"workflowId" validate:"required"`
		EventName  string `json:"eventName"  validate:"required"`
	}

	if err := utils.ValidateAndParseBody(c, &req); err != nil {
		return err
	}

	element, err := h.elementRepo.GetElementByID(c.Context(), req.ElementID)
	if err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Element not found", err)
	}

	if _, err := h.projectRepo.GetProjectWithAccess(c.Context(), element.Page.ProjectId, userID); err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied to project", err, userID)
	}

	workflow, err := h.workflowRepo.GetEventWorkflowByID(c.Context(), req.WorkflowID)
	if err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Workflow not found", err)
	}

	if workflow.ProjectId != element.Page.ProjectId {
		return utils.SendError(c, fiber.StatusBadRequest, "Workflow and element must belong to the same project", nil)
	}

	exists, err := h.eewRepo.CheckIfWorkflowLinkedToElement(c.Context(), req.ElementID, req.WorkflowID, req.EventName)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to check existing association", err)
	}
	if exists {
		return utils.SendError(c, fiber.StatusBadRequest, "This workflow is already linked to this element for this event", nil)
	}

	createdEEW, err := h.eewRepo.CreateElementEventWorkflow(c.Context(), &models.ElementEventWorkflow{
		ElementId:  req.ElementID,
		WorkflowId: req.WorkflowID,
		EventName:  req.EventName,
	})
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to create element event workflow", err)
	}

	return utils.SendJSON(c, fiber.StatusCreated, createdEEW)
}

// GetElementEventWorkflowByID retrieves a specific element event workflow.
func (h *ElementEventWorkflowHandler) GetElementEventWorkflowByID(c *fiber.Ctx) error {
	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	id, err := utils.ValidateRequiredParam(c, "id")
	if err != nil {
		return err
	}

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

	if _, err := h.projectRepo.GetProjectWithAccess(c.Context(), element.Page.ProjectId, userID); err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	return utils.SendJSON(c, fiber.StatusOK, eew)
}

// GetElementEventWorkflowsByPageId retrieves all element event workflows for a page.
func (h *ElementEventWorkflowHandler) GetElementEventWorkflowsByPageId(c *fiber.Ctx) error {
	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	pageID, err := utils.ValidateRequiredParam(c, "pageId")
	if err != nil {
		return err
	}

	page, err := h.pageRepo.GetPageByID(c.Context(), pageID, "")
	if err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Page not found", err)
	}

	if _, err := h.projectRepo.GetProjectWithAccess(c.Context(), page.ProjectId, userID); err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied to project", err, userID)
	}

	eews, err := h.eewRepo.GetElementEventWorkflowsByPageID(c.Context(), pageID)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve element event workflows", err)
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{"data": eews, "count": len(eews)})
}

// GetElementEventWorkflows retrieves element event workflows with optional filters.
// Exactly one of elementId, workflowId, or eventName may be provided; if none are
// provided all workflows accessible to the user are returned.
func (h *ElementEventWorkflowHandler) GetElementEventWorkflows(c *fiber.Ctx) error {
	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	elementID := c.Query("elementId")
	workflowID := c.Query("workflowId")
	eventName := c.Query("eventName")

	var (
		eews    []models.ElementEventWorkflow
		fetchErr error
	)

	switch {
	case elementID != "" && workflowID == "" && eventName == "":
		element, err := h.elementRepo.GetElementByID(c.Context(), elementID)
		if err != nil {
			return utils.SendError(c, fiber.StatusNotFound, "Element not found", err)
		}
		if _, err := h.projectRepo.GetProjectWithAccess(c.Context(), element.Page.ProjectId, userID); err != nil {
			return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
		}
		eews, fetchErr = h.eewRepo.GetElementEventWorkflowsByElementID(c.Context(), elementID)

	case workflowID != "" && elementID == "" && eventName == "":
		workflow, err := h.workflowRepo.GetEventWorkflowByID(c.Context(), workflowID)
		if err != nil {
			return utils.SendError(c, fiber.StatusNotFound, "Workflow not found", err)
		}
		if _, err := h.projectRepo.GetProjectWithAccess(c.Context(), workflow.ProjectId, userID); err != nil {
			return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
		}
		eews, fetchErr = h.eewRepo.GetElementEventWorkflowsByWorkflowID(c.Context(), workflowID)

	default:
		// eventName-only, combined filters, or no filters — fetch then apply access filter
		if elementID != "" || workflowID != "" || eventName != "" {
			eews, fetchErr = h.eewRepo.GetElementEventWorkflowsByFilters(c.Context(), elementID, workflowID, eventName)
		} else {
			eews, fetchErr = h.eewRepo.GetAllElementEventWorkflows(c.Context())
		}
		if fetchErr != nil {
			return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve element event workflows", fetchErr)
		}
		eews = h.filterByAccess(c, eews, userID)
	}

	if fetchErr != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve element event workflows", fetchErr)
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{"data": eews, "count": len(eews)})
}

// filterByAccess removes any EEW entries whose element's project the user cannot access.
func (h *ElementEventWorkflowHandler) filterByAccess(c *fiber.Ctx, eews []models.ElementEventWorkflow, userID string) []models.ElementEventWorkflow {
	filtered := make([]models.ElementEventWorkflow, 0, len(eews))
	for _, eew := range eews {
		element, err := h.elementRepo.GetElementByID(c.Context(), eew.ElementId)
		if err != nil {
			continue
		}
		if _, err := h.projectRepo.GetProjectWithAccess(c.Context(), element.Page.ProjectId, userID); err == nil {
			filtered = append(filtered, eew)
		}
	}
	return filtered
}

// UpdateElementEventWorkflow updates the event name of an element event workflow.
func (h *ElementEventWorkflowHandler) UpdateElementEventWorkflow(c *fiber.Ctx) error {
	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	id, err := utils.ValidateRequiredParam(c, "id")
	if err != nil {
		return err
	}

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

	if _, err := h.projectRepo.GetProjectWithAccess(c.Context(), element.Page.ProjectId, userID); err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	var req struct {
		EventName string `json:"eventName" validate:"required"`
	}

	if err := utils.ValidateAndParseBody(c, &req); err != nil {
		return err
	}

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

// DeleteElementEventWorkflow deletes an element event workflow.
func (h *ElementEventWorkflowHandler) DeleteElementEventWorkflow(c *fiber.Ctx) error {
	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	id, err := utils.ValidateRequiredParam(c, "id")
	if err != nil {
		return err
	}

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

	if _, err := h.projectRepo.GetProjectWithAccess(c.Context(), element.Page.ProjectId, userID); err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	if err := h.eewRepo.DeleteElementEventWorkflow(c.Context(), id); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to delete element event workflow", err)
	}

	return utils.SendNoContent(c)
}

// DeleteElementEventWorkflowsByElement deletes all event workflow links for a specific element.
func (h *ElementEventWorkflowHandler) DeleteElementEventWorkflowsByElement(c *fiber.Ctx) error {
	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	elementID, err := utils.ValidateRequiredParam(c, "elementId")
	if err != nil {
		return err
	}

	element, err := h.elementRepo.GetElementByID(c.Context(), elementID)
	if err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Element not found", err)
	}

	if _, err := h.projectRepo.GetProjectWithAccess(c.Context(), element.Page.ProjectId, userID); err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	if err := h.eewRepo.DeleteElementEventWorkflowsByElementID(c.Context(), elementID); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to delete element event workflows", err)
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Element event workflows deleted successfully")
}

// DeleteElementEventWorkflowsByWorkflow deletes all element associations for a specific workflow.
func (h *ElementEventWorkflowHandler) DeleteElementEventWorkflowsByWorkflow(c *fiber.Ctx) error {
	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	workflowID, err := utils.ValidateRequiredParam(c, "workflowId")
	if err != nil {
		return err
	}

	workflow, err := h.workflowRepo.GetEventWorkflowByID(c.Context(), workflowID)
	if err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Workflow not found", err)
	}

	if _, err := h.projectRepo.GetProjectWithAccess(c.Context(), workflow.ProjectId, userID); err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	if err := h.eewRepo.DeleteElementEventWorkflowsByWorkflowID(c.Context(), workflowID); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to delete element event workflows", err)
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Element event workflows deleted successfully")
}
