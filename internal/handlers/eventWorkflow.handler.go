package handlers

import (
	"encoding/json"
	"my-go-app/internal/dto"
	"my-go-app/internal/models"
	"my-go-app/internal/services"
	"my-go-app/pkg/utils"

	"github.com/gofiber/fiber/v3"
)

type EventWorkflowHandler struct {
	eventWorkflowService        services.EventWorkflowServiceInterface
	elementEventWorkflowService services.ElementEventWorkflowServiceInterface
}

func NewEventWorkflowHandler(
	eventWorkflowService services.EventWorkflowServiceInterface,
	elementEventWorkflowService services.ElementEventWorkflowServiceInterface,
) *EventWorkflowHandler {
	return &EventWorkflowHandler{
		eventWorkflowService:        eventWorkflowService,
		elementEventWorkflowService: elementEventWorkflowService,
	}
}

func (h *EventWorkflowHandler) CreateEventWorkflow(c fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	req := new(dto.CreateEventWorkflowRequest)
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	workflow, err := h.eventWorkflowService.CreateEventWorkflow(c.RequestCtx(), userID, &models.EventWorkflow{
		ProjectId:   req.ProjectID,
		Name:        req.Name,
		Description: req.Description,
		CanvasData:  req.CanvasData,
		Handlers:    req.Handlers,
		Enabled:     req.Enabled != nil && *req.Enabled,
	})
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to create event workflow")
	}

	if workflow.CanvasData == nil {
		workflow.CanvasData = json.RawMessage("{}")
	}
	if workflow.Handlers == nil {
		workflow.Handlers = json.RawMessage("[]")
	}

	return utils.SendJSON(c, fiber.StatusCreated, workflow)
}

func (h *EventWorkflowHandler) GetEventWorkflowByID(c fiber.Ctx) error {
	ids, err := utils.MustParams(c, "id")
	if err != nil {
		return err
	}
	workflowID := ids[0]

	workflow, err := h.eventWorkflowService.GetEventWorkflowByID(c.RequestCtx(), workflowID)
	if err != nil {
		return utils.HandleRepoError(c, err, "Event workflow not found", "Failed to retrieve event workflow")
	}
	if workflow == nil {
		return fiber.NewError(fiber.StatusNotFound, "Event workflow not found")
	}

	return utils.SendJSON(c, fiber.StatusOK, workflow)
}

func (h *EventWorkflowHandler) GetEventWorkflowsByProject(c fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "projectid")
	if err != nil {
		return err
	}
	projectID := ids[0]

	var enabledPtr *bool
	if enabled := c.Query("enabled"); enabled == "true" {
		v := true
		enabledPtr = &v
	} else if enabled == "false" {
		v := false
		enabledPtr = &v
	}

	workflows, err := h.eventWorkflowService.GetEventWorkflowsWithFilters(c.RequestCtx(), projectID, userID, enabledPtr, c.Query("search"))
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to retrieve event workflows")
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{"data": workflows, "count": len(workflows)})
}

func (h *EventWorkflowHandler) UpdateEventWorkflow(c fiber.Ctx) error {
	ids, err := utils.MustParams(c, "id")
	if err != nil {
		return err
	}
	workflowID := ids[0]

	existing, err := h.eventWorkflowService.GetEventWorkflowByID(c.RequestCtx(), workflowID)
	if err != nil {
		return utils.HandleRepoError(c, err, "Event workflow not found", "Failed to retrieve event workflow")
	}
	if existing == nil {
		return fiber.NewError(fiber.StatusNotFound, "Event workflow not found")
	}

	req := new(dto.UpdateEventWorkflowRequest)
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	updated, err := h.eventWorkflowService.UpdateEventWorkflow(c.RequestCtx(), workflowID, &models.EventWorkflow{
		Name:        *req.Name,
		Description: req.Description,
		CanvasData:  req.CanvasData,
		Handlers:    req.Handlers,
		Enabled:     req.Enabled != nil && *req.Enabled,
	})
	if err != nil {
		return utils.HandleRepoError(c, err, "Event workflow not found", "Failed to update event workflow")
	}

	return utils.SendJSON(c, fiber.StatusOK, updated)
}

func (h *EventWorkflowHandler) UpdateEventWorkflowEnabled(c fiber.Ctx) error {
	ids, err := utils.MustParams(c, "id")
	if err != nil {
		return err
	}
	workflowID := ids[0]

	req := new(dto.UpdateEventWorkflowEnabledRequest)
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	if err := h.eventWorkflowService.UpdateEventWorkflowEnabled(c.RequestCtx(), workflowID, *req.Enabled); err != nil {
		return utils.HandleRepoError(c, err, "Event workflow not found", "Failed to update event workflow status")
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{"enabled": *req.Enabled})
}

func (h *EventWorkflowHandler) DeleteEventWorkflow(c fiber.Ctx) error {
	ids, err := utils.MustParams(c, "id")
	if err != nil {
		return err
	}
	workflowID := ids[0]

	if err := h.eventWorkflowService.DeleteEventWorkflow(c.RequestCtx(), workflowID); err != nil {
		return utils.HandleRepoError(c, err, "Event workflow not found", "Failed to delete event workflow")
	}

	return utils.SendNoContent(c)
}

func (h *EventWorkflowHandler) GetEventWorkflowElements(c fiber.Ctx) error {
	ids, err := utils.MustParams(c, "id")
	if err != nil {
		return err
	}
	workflowID := ids[0]

	elements, err := h.elementEventWorkflowService.GetElementEventWorkflowsByWorkflowID(c.RequestCtx(), workflowID)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to retrieve elements")
	}

	return utils.SendJSON(c, fiber.StatusOK, fiber.Map{"data": elements, "count": len(elements)})
}
