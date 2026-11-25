package handlers

import (
	"my-go-app/internal/authz"
	"my-go-app/internal/dto"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"
	"my-go-app/pkg/utils"

	"github.com/gofiber/fiber/v3"
)

type ElementEventWorkflowHandler struct {
	elementEventWorkflowService services.ElementEventWorkflowServiceInterface
	projectService              services.ProjectServiceInterface
	pageService                 services.PageServiceInterface
	authorizer                  *authz.Authorizer
	eventWorkflowService        services.EventWorkflowServiceInterface
	elementRepo                 repositories.ElementRepositoryInterface
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

// WithWorkflowAuth enables project-scoped RBAC checks on ID and query-based lookups.
func (h *ElementEventWorkflowHandler) WithWorkflowAuth(
	a *authz.Authorizer,
	ews services.EventWorkflowServiceInterface,
	elements repositories.ElementRepositoryInterface,
) *ElementEventWorkflowHandler {
	h.authorizer = a
	h.eventWorkflowService = ews
	h.elementRepo = elements
	return h
}

// Create creates a new element event workflow association.
func (h *ElementEventWorkflowHandler) Create(c fiber.Ctx) error {
	_, err := utils.RequireUserID(c)
	if err != nil {
		return err
	}

	req := new(dto.CreateElementEventWorkflowRequest)
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	created, err := h.elementEventWorkflowService.Create(c.RequestCtx(), &models.ElementEventWorkflow{
		ElementId:  req.ElementID,
		WorkflowId: req.WorkflowID,
		EventName:  req.EventName,
	})
	if err != nil {
		return utils.RespondError(c, err, "", "Failed to create element event workflow")
	}

	return utils.SendJSON(c, fiber.StatusCreated, created)
}

// GetByID retrieves a specific element event workflow.
func (h *ElementEventWorkflowHandler) GetByID(c fiber.Ctx) error {
	userID, ids, err := utils.RequireUserAndParams(c, "id")
	if err != nil {
		return err
	}
	eewID := ids[0]

	eew, err := h.elementEventWorkflowService.GetByID(c.RequestCtx(), eewID)
	if err != nil {
		return utils.RespondError(c, err, "Element event workflow not found", "Failed to retrieve element event workflow")
	}
	if eew == nil {
		return fiber.NewError(fiber.StatusNotFound, "Element event workflow not found")
	}
	if h.authorizer != nil {
		if err := h.authorizeElementOrWorkflow(c, userID, eew.ElementId, eew.WorkflowId); err != nil {
			return err
		}
	}

	return utils.SendJSON(c, fiber.StatusOK, eew)
}

func (h *ElementEventWorkflowHandler) authorizeElementOrWorkflow(c fiber.Ctx, userID, elementID, workflowID string) error {
	if h.authorizer == nil {
		return nil
	}
	if elementID != "" && h.elementRepo != nil {
		el, err := h.elementRepo.GetByID(c.RequestCtx(), elementID)
		if err != nil {
			return utils.RespondError(c, err, "Element not found", "Failed to verify element access")
		}
		if el != nil && el.Page != nil && el.Page.ProjectId != "" {
			if _, err := h.authorizer.Authorize(c.RequestCtx(), el.Page.ProjectId, userID, authz.PermProjectView); err != nil {
				return utils.RespondError(c, err, "Project not found", "Failed to verify project access")
			}
			return nil
		}
	}
	if workflowID != "" && h.eventWorkflowService != nil {
		wf, err := h.eventWorkflowService.GetByID(c.RequestCtx(), workflowID)
		if err != nil {
			return utils.RespondError(c, err, "Event workflow not found", "Failed to verify workflow access")
		}
		if wf != nil && wf.ProjectId != "" {
			if _, err := h.authorizer.Authorize(c.RequestCtx(), wf.ProjectId, userID, authz.PermProjectView); err != nil {
				return utils.RespondError(c, err, "Project not found", "Failed to verify project access")
			}
			return nil
		}
	}
	return fiber.NewError(fiber.StatusForbidden, "Access denied")
}

// ListByPage retrieves all element event workflows for a page.
func (h *ElementEventWorkflowHandler) ListByPage(c fiber.Ctx) error {
	userID, ids, err := utils.RequireUserAndParams(c, "pageId")
	if err != nil {
		return err
	}
	pageID := ids[0]

	page, err := h.pageService.GetByID(c.RequestCtx(), pageID)
	if err != nil {
		return utils.RespondError(c, err, "Page not found", "Failed to retrieve page")
	}

	if _, err := h.projectService.GetWithAccess(c.RequestCtx(), page.ProjectId, userID); err != nil {
		return respondProjectError(c, err, "Failed to verify project access")
	}

	eews, err := h.elementEventWorkflowService.ListByPage(c.RequestCtx(), pageID)
	if err != nil {
		return utils.RespondError(c, err, "", "Failed to retrieve element event workflows")
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{"data": eews, "count": len(eews)})
}

// List retrieves element event workflows with optional filters.
func (h *ElementEventWorkflowHandler) List(c fiber.Ctx) error {
	userID, err := utils.RequireUserID(c)
	if err != nil {
		return err
	}

	elementID := c.Query("elementId")
	workflowID := c.Query("workflowId")
	eventName := c.Query("eventName")

	if h.authorizer != nil {
		if elementID == "" && workflowID == "" {
			return fiber.NewError(fiber.StatusBadRequest, "elementId or workflowId query parameter is required")
		}
		if err := h.authorizeElementOrWorkflow(c, userID, elementID, workflowID); err != nil {
			return err
		}
	}

	var (
		eews     []models.ElementEventWorkflow
		fetchErr error
	)

	switch {
	case elementID != "" && workflowID == "" && eventName == "":
		eews, fetchErr = h.elementEventWorkflowService.ListByElement(c.RequestCtx(), elementID)

	case workflowID != "" && elementID == "" && eventName == "":
		eews, fetchErr = h.elementEventWorkflowService.ListByWorkflow(c.RequestCtx(), workflowID)

	default:
		if elementID != "" || workflowID != "" || eventName != "" {
			eews, fetchErr = h.elementEventWorkflowService.ListByFilters(c.RequestCtx(), elementID, workflowID, eventName)
		} else {
			eews, fetchErr = h.elementEventWorkflowService.ListAll(c.RequestCtx())
		}
		if fetchErr != nil {
			return utils.RespondError(c, fetchErr, "", "Failed to retrieve element event workflows")
		}
	}

	if fetchErr != nil {
		return utils.RespondError(c, fetchErr, "", "Failed to retrieve element event workflows")
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{"data": eews, "count": len(eews)})
}

// Update updates the event name of an element event workflow.
func (h *ElementEventWorkflowHandler) Update(c fiber.Ctx) error {
	_, ids, err := utils.RequireUserAndParams(c, "id")
	if err != nil {
		return err
	}
	eewID := ids[0]

	existing, err := h.elementEventWorkflowService.GetByID(c.RequestCtx(), eewID)
	if err != nil {
		return utils.RespondError(c, err, "Element event workflow not found", "Failed to retrieve element event workflow")
	}
	if existing == nil {
		return fiber.NewError(fiber.StatusNotFound, "Element event workflow not found")
	}

	req := new(dto.UpdateElementEventWorkflowRequest)
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	existing.EventName = req.EventName

	updated, err := h.elementEventWorkflowService.Update(c.RequestCtx(), eewID, existing)
	if err != nil {
		return utils.RespondError(c, err, "Element event workflow not found", "Failed to update element event workflow")
	}

	return utils.SendJSON(c, fiber.StatusOK, updated)
}

// Delete deletes an element event workflow.
func (h *ElementEventWorkflowHandler) Delete(c fiber.Ctx) error {
	_, ids, err := utils.RequireUserAndParams(c, "id")
	if err != nil {
		return err
	}
	eewID := ids[0]

	eew, err := h.elementEventWorkflowService.GetByID(c.RequestCtx(), eewID)
	if err != nil {
		return utils.RespondError(c, err, "Element event workflow not found", "Failed to retrieve element event workflow")
	}
	if eew == nil {
		return fiber.NewError(fiber.StatusNotFound, "Element event workflow not found")
	}

	if err := h.elementEventWorkflowService.Delete(c.RequestCtx(), eewID); err != nil {
		return utils.RespondError(c, err, "Element event workflow not found", "Failed to delete element event workflow")
	}

	return utils.SendNoContent(c)
}

// DeleteByElement deletes all event workflow links for a specific element.
func (h *ElementEventWorkflowHandler) DeleteByElement(c fiber.Ctx) error {
	_, ids, err := utils.RequireUserAndParams(c, "elementId")
	if err != nil {
		return err
	}
	elementID := ids[0]

	if err := h.elementEventWorkflowService.DeleteByElement(c.RequestCtx(), elementID); err != nil {
		return utils.RespondError(c, err, "", "Failed to delete element event workflows")
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Element event workflows deleted successfully")
}

// DeleteByWorkflow deletes all element associations for a specific workflow.
func (h *ElementEventWorkflowHandler) DeleteByWorkflow(c fiber.Ctx) error {
	_, ids, err := utils.RequireUserAndParams(c, "workflowId")
	if err != nil {
		return err
	}
	workflowID := ids[0]

	if err := h.elementEventWorkflowService.DeleteByWorkflow(c.RequestCtx(), workflowID); err != nil {
		return utils.RespondError(c, err, "", "Failed to delete element event workflows")
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Element event workflows deleted successfully")
}

// fiber:context-methods migrated
