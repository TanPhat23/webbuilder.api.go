package handlers

import (
	"my-go-app/internal/dto"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

type ElementEventWorkflowHandler struct {
	eewRepo      repositories.ElementEventWorkflowRepositoryInterface
	elementRepo  repositories.ElementRepositoryInterface
	workflowRepo repositories.EventWorkflowRepositoryInterface
	projectRepo  repositories.ProjectRepositoryInterface
	pageRepo     repositories.PageRepositoryInterface
}

func NewElementEventWorkflowHandler(
	eewRepo repositories.ElementEventWorkflowRepositoryInterface,
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

	var req dto.CreateElementEventWorkflowRequest
	if err := utils.ValidateAndParseBody(c, &req); err != nil {
		return err
	}

	element, err := h.elementRepo.GetElementByID(c.Context(), req.ElementID)
	if err != nil {
		return utils.HandleRepoError(c, err, "Element not found", "Failed to retrieve element")
	}

	if _, err := h.projectRepo.GetProjectWithAccess(c.Context(), element.Page.ProjectId, userID); err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied to project", err, userID)
	}

	workflow, err := h.workflowRepo.GetEventWorkflowByID(c.Context(), req.WorkflowID)
	if err != nil {
		return utils.HandleRepoError(c, err, "Workflow not found", "Failed to retrieve workflow")
	}

	if workflow.ProjectId != element.Page.ProjectId {
		return fiber.NewError(fiber.StatusBadRequest, "Workflow and element must belong to the same project")
	}

	exists, err := h.eewRepo.CheckIfWorkflowLinkedToElement(c.Context(), req.ElementID, req.WorkflowID, req.EventName)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to check existing association")
	}
	if exists {
		return fiber.NewError(fiber.StatusBadRequest, "This workflow is already linked to this element for this event")
	}

	created, err := h.eewRepo.CreateElementEventWorkflow(c.Context(), &models.ElementEventWorkflow{
		ElementId:  req.ElementID,
		WorkflowId: req.WorkflowID,
		EventName:  req.EventName,
	})
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to create element event workflow")
	}

	return utils.SendJSON(c, fiber.StatusCreated, created)
}

// GetElementEventWorkflowByID retrieves a specific element event workflow.
func (h *ElementEventWorkflowHandler) GetElementEventWorkflowByID(c *fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "id")
	if err != nil {
		return err
	}
	eewID := ids[0]

	eew, err := h.eewRepo.GetElementEventWorkflowByID(c.Context(), eewID)
	if err != nil {
		return utils.HandleRepoError(c, err, "Element event workflow not found", "Failed to retrieve element event workflow")
	}
	if eew == nil {
		return fiber.NewError(fiber.StatusNotFound, "Element event workflow not found")
	}

	element, err := h.elementRepo.GetElementByID(c.Context(), eew.ElementId)
	if err != nil {
		return utils.HandleRepoError(c, err, "Element not found", "Failed to retrieve element")
	}

	if _, err := h.projectRepo.GetProjectWithAccess(c.Context(), element.Page.ProjectId, userID); err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	return utils.SendJSON(c, fiber.StatusOK, eew)
}

// GetElementEventWorkflowsByPageId retrieves all element event workflows for a page.
func (h *ElementEventWorkflowHandler) GetElementEventWorkflowsByPageId(c *fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "pageId")
	if err != nil {
		return err
	}
	pageID := ids[0]

	page, err := h.pageRepo.GetPageByID(c.Context(), pageID, "")
	if err != nil {
		return utils.HandleRepoError(c, err, "Page not found", "Failed to retrieve page")
	}

	if _, err := h.projectRepo.GetProjectWithAccess(c.Context(), page.ProjectId, userID); err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied to project", err, userID)
	}

	eews, err := h.eewRepo.GetElementEventWorkflowsByPageID(c.Context(), pageID)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to retrieve element event workflows")
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{"data": eews, "count": len(eews)})
}

// GetElementEventWorkflows retrieves element event workflows with optional filters.
func (h *ElementEventWorkflowHandler) GetElementEventWorkflows(c *fiber.Ctx) error {
	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	elementID := c.Query("elementId")
	workflowID := c.Query("workflowId")
	eventName := c.Query("eventName")

	var (
		eews     []models.ElementEventWorkflow
		fetchErr error
	)

	switch {
	case elementID != "" && workflowID == "" && eventName == "":
		element, err := h.elementRepo.GetElementByID(c.Context(), elementID)
		if err != nil {
			return utils.HandleRepoError(c, err, "Element not found", "Failed to retrieve element")
		}
		if _, err := h.projectRepo.GetProjectWithAccess(c.Context(), element.Page.ProjectId, userID); err != nil {
			return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
		}
		eews, fetchErr = h.eewRepo.GetElementEventWorkflowsByElementID(c.Context(), elementID)

	case workflowID != "" && elementID == "" && eventName == "":
		workflow, err := h.workflowRepo.GetEventWorkflowByID(c.Context(), workflowID)
		if err != nil {
			return utils.HandleRepoError(c, err, "Workflow not found", "Failed to retrieve workflow")
		}
		if _, err := h.projectRepo.GetProjectWithAccess(c.Context(), workflow.ProjectId, userID); err != nil {
			return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
		}
		eews, fetchErr = h.eewRepo.GetElementEventWorkflowsByWorkflowID(c.Context(), workflowID)

	default:
		if elementID != "" || workflowID != "" || eventName != "" {
			eews, fetchErr = h.eewRepo.GetElementEventWorkflowsByFilters(c.Context(), elementID, workflowID, eventName)
		} else {
			eews, fetchErr = h.eewRepo.GetAllElementEventWorkflows(c.Context())
		}
		if fetchErr != nil {
			return utils.HandleRepoError(c, fetchErr, "", "Failed to retrieve element event workflows")
		}
		eews = h.filterByAccess(c, eews, userID)
	}

	if fetchErr != nil {
		return utils.HandleRepoError(c, fetchErr, "", "Failed to retrieve element event workflows")
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{"data": eews, "count": len(eews)})
}

// UpdateElementEventWorkflow updates the event name of an element event workflow.
func (h *ElementEventWorkflowHandler) UpdateElementEventWorkflow(c *fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "id")
	if err != nil {
		return err
	}
	eewID := ids[0]

	existing, err := h.eewRepo.GetElementEventWorkflowByID(c.Context(), eewID)
	if err != nil {
		return utils.HandleRepoError(c, err, "Element event workflow not found", "Failed to retrieve element event workflow")
	}
	if existing == nil {
		return fiber.NewError(fiber.StatusNotFound, "Element event workflow not found")
	}

	element, err := h.elementRepo.GetElementByID(c.Context(), existing.ElementId)
	if err != nil {
		return utils.HandleRepoError(c, err, "Element not found", "Failed to retrieve element")
	}

	if _, err := h.projectRepo.GetProjectWithAccess(c.Context(), element.Page.ProjectId, userID); err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	var req dto.UpdateElementEventWorkflowRequest
	if err := utils.ValidateAndParseBody(c, &req); err != nil {
		return err
	}

	if req.EventName != existing.EventName {
		exists, err := h.eewRepo.CheckIfWorkflowLinkedToElement(c.Context(), existing.ElementId, existing.WorkflowId, req.EventName)
		if err != nil {
			return utils.HandleRepoError(c, err, "", "Failed to check existing association")
		}
		if exists {
			return fiber.NewError(fiber.StatusBadRequest, "This workflow is already linked to this element for this event")
		}
	}

	existing.EventName = req.EventName

	updated, err := h.eewRepo.UpdateElementEventWorkflow(c.Context(), eewID, existing)
	if err != nil {
		return utils.HandleRepoError(c, err, "Element event workflow not found", "Failed to update element event workflow")
	}

	return utils.SendJSON(c, fiber.StatusOK, updated)
}

// DeleteElementEventWorkflow deletes an element event workflow.
func (h *ElementEventWorkflowHandler) DeleteElementEventWorkflow(c *fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "id")
	if err != nil {
		return err
	}
	eewID := ids[0]

	eew, err := h.eewRepo.GetElementEventWorkflowByID(c.Context(), eewID)
	if err != nil {
		return utils.HandleRepoError(c, err, "Element event workflow not found", "Failed to retrieve element event workflow")
	}
	if eew == nil {
		return fiber.NewError(fiber.StatusNotFound, "Element event workflow not found")
	}

	element, err := h.elementRepo.GetElementByID(c.Context(), eew.ElementId)
	if err != nil {
		return utils.HandleRepoError(c, err, "Element not found", "Failed to retrieve element")
	}

	if _, err := h.projectRepo.GetProjectWithAccess(c.Context(), element.Page.ProjectId, userID); err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	if err := h.eewRepo.DeleteElementEventWorkflow(c.Context(), eewID); err != nil {
		return utils.HandleRepoError(c, err, "Element event workflow not found", "Failed to delete element event workflow")
	}

	return utils.SendNoContent(c)
}

// DeleteElementEventWorkflowsByElement deletes all event workflow links for a specific element.
func (h *ElementEventWorkflowHandler) DeleteElementEventWorkflowsByElement(c *fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "elementId")
	if err != nil {
		return err
	}
	elementID := ids[0]

	element, err := h.elementRepo.GetElementByID(c.Context(), elementID)
	if err != nil {
		return utils.HandleRepoError(c, err, "Element not found", "Failed to retrieve element")
	}

	if _, err := h.projectRepo.GetProjectWithAccess(c.Context(), element.Page.ProjectId, userID); err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	if err := h.eewRepo.DeleteElementEventWorkflowsByElementID(c.Context(), elementID); err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to delete element event workflows")
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Element event workflows deleted successfully")
}

// DeleteElementEventWorkflowsByWorkflow deletes all element associations for a specific workflow.
func (h *ElementEventWorkflowHandler) DeleteElementEventWorkflowsByWorkflow(c *fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "workflowId")
	if err != nil {
		return err
	}
	workflowID := ids[0]

	workflow, err := h.workflowRepo.GetEventWorkflowByID(c.Context(), workflowID)
	if err != nil {
		return utils.HandleRepoError(c, err, "Workflow not found", "Failed to retrieve workflow")
	}

	if _, err := h.projectRepo.GetProjectWithAccess(c.Context(), workflow.ProjectId, userID); err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied", err, userID)
	}

	if err := h.eewRepo.DeleteElementEventWorkflowsByWorkflowID(c.Context(), workflowID); err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to delete element event workflows")
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Element event workflows deleted successfully")
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