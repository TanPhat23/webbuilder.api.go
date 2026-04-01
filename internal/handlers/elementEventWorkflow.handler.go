package handlers

import (
	"my-go-app/internal/dto"
	"my-go-app/internal/models"
	"my-go-app/internal/services"
	"my-go-app/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

type ElementEventWorkflowHandler struct {
	elementEventWorkflowService services.ElementEventWorkflowServiceInterface
	projectService              services.ProjectServiceInterface
	pageService                 services.PageServiceInterface
}

func NewElementEventWorkflowHandler(
	elementEventWorkflowService services.ElementEventWorkflowServiceInterface,
	projectService services.ProjectServiceInterface,
	pageService services.PageServiceInterface,
) *ElementEventWorkflowHandler {
	return &ElementEventWorkflowHandler{
		elementEventWorkflowService: elementEventWorkflowService,
		projectService:              projectService,
		pageService:                 pageService,
	}
}

// CreateElementEventWorkflow creates a new element event workflow association.
func (h *ElementEventWorkflowHandler) CreateElementEventWorkflow(c *fiber.Ctx) error {
	_, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	var req dto.CreateElementEventWorkflowRequest
	if err := utils.ValidateAndParseBody(c, &req); err != nil {
		return err
	}

	created, err := h.elementEventWorkflowService.CreateElementEventWorkflow(c.Context(), &models.ElementEventWorkflow{
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
	_, ids, err := utils.MustUserAndParams(c, "id")
	if err != nil {
		return err
	}
	eewID := ids[0]

	eew, err := h.elementEventWorkflowService.GetElementEventWorkflowByID(c.Context(), eewID)
	if err != nil {
		return utils.HandleRepoError(c, err, "Element event workflow not found", "Failed to retrieve element event workflow")
	}
	if eew == nil {
		return fiber.NewError(fiber.StatusNotFound, "Element event workflow not found")
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

	page, err := h.pageService.GetPageByID(c.Context(), pageID)
	if err != nil {
		return utils.HandleRepoError(c, err, "Page not found", "Failed to retrieve page")
	}

	if _, err := h.projectService.GetProjectWithAccess(c.Context(), page.ProjectId, userID); err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Access denied to project", err, userID)
	}

	eews, err := h.elementEventWorkflowService.GetElementEventWorkflowsByPageID(c.Context(), pageID)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to retrieve element event workflows")
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{"data": eews, "count": len(eews)})
}

// GetElementEventWorkflows retrieves element event workflows with optional filters.
func (h *ElementEventWorkflowHandler) GetElementEventWorkflows(c *fiber.Ctx) error {
	_, err := utils.ValidateUserID(c)
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
		eews, fetchErr = h.elementEventWorkflowService.GetElementEventWorkflowsByElementID(c.Context(), elementID)

	case workflowID != "" && elementID == "" && eventName == "":
		eews, fetchErr = h.elementEventWorkflowService.GetElementEventWorkflowsByWorkflowID(c.Context(), workflowID)

	default:
		if elementID != "" || workflowID != "" || eventName != "" {
			eews, fetchErr = h.elementEventWorkflowService.GetElementEventWorkflowsByFilters(c.Context(), elementID, workflowID, eventName)
		} else {
			eews, fetchErr = h.elementEventWorkflowService.GetAllElementEventWorkflows(c.Context())
		}
		if fetchErr != nil {
			return utils.HandleRepoError(c, fetchErr, "", "Failed to retrieve element event workflows")
		}
	}

	if fetchErr != nil {
		return utils.HandleRepoError(c, fetchErr, "", "Failed to retrieve element event workflows")
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{"data": eews, "count": len(eews)})
}

// UpdateElementEventWorkflow updates the event name of an element event workflow.
func (h *ElementEventWorkflowHandler) UpdateElementEventWorkflow(c *fiber.Ctx) error {
	_, ids, err := utils.MustUserAndParams(c, "id")
	if err != nil {
		return err
	}
	eewID := ids[0]

	existing, err := h.elementEventWorkflowService.GetElementEventWorkflowByID(c.Context(), eewID)
	if err != nil {
		return utils.HandleRepoError(c, err, "Element event workflow not found", "Failed to retrieve element event workflow")
	}
	if existing == nil {
		return fiber.NewError(fiber.StatusNotFound, "Element event workflow not found")
	}

	var req dto.UpdateElementEventWorkflowRequest
	if err := utils.ValidateAndParseBody(c, &req); err != nil {
		return err
	}

	existing.EventName = req.EventName

	updated, err := h.elementEventWorkflowService.UpdateElementEventWorkflow(c.Context(), eewID, existing)
	if err != nil {
		return utils.HandleRepoError(c, err, "Element event workflow not found", "Failed to update element event workflow")
	}

	return utils.SendJSON(c, fiber.StatusOK, updated)
}

// DeleteElementEventWorkflow deletes an element event workflow.
func (h *ElementEventWorkflowHandler) DeleteElementEventWorkflow(c *fiber.Ctx) error {
	_, ids, err := utils.MustUserAndParams(c, "id")
	if err != nil {
		return err
	}
	eewID := ids[0]

	eew, err := h.elementEventWorkflowService.GetElementEventWorkflowByID(c.Context(), eewID)
	if err != nil {
		return utils.HandleRepoError(c, err, "Element event workflow not found", "Failed to retrieve element event workflow")
	}
	if eew == nil {
		return fiber.NewError(fiber.StatusNotFound, "Element event workflow not found")
	}

	if err := h.elementEventWorkflowService.DeleteElementEventWorkflow(c.Context(), eewID); err != nil {
		return utils.HandleRepoError(c, err, "Element event workflow not found", "Failed to delete element event workflow")
	}

	return utils.SendNoContent(c)
}

// DeleteElementEventWorkflowsByElement deletes all event workflow links for a specific element.
func (h *ElementEventWorkflowHandler) DeleteElementEventWorkflowsByElement(c *fiber.Ctx) error {
	_, ids, err := utils.MustUserAndParams(c, "elementId")
	if err != nil {
		return err
	}
	elementID := ids[0]

	if err := h.elementEventWorkflowService.DeleteElementEventWorkflowsByElementID(c.Context(), elementID); err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to delete element event workflows")
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Element event workflows deleted successfully")
}

// DeleteElementEventWorkflowsByWorkflow deletes all element associations for a specific workflow.
func (h *ElementEventWorkflowHandler) DeleteElementEventWorkflowsByWorkflow(c *fiber.Ctx) error {
	_, ids, err := utils.MustUserAndParams(c, "workflowId")
	if err != nil {
		return err
	}
	workflowID := ids[0]

	if err := h.elementEventWorkflowService.DeleteElementEventWorkflowsByWorkflowID(c.Context(), workflowID); err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to delete element event workflows")
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Element event workflows deleted successfully")
}