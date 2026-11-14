package handlers

import (
	"context"
	"encoding/json"

	wsentity "my-go-app/internal/realtime/delivery/ws/entity"
	domainEntity "my-go-app/internal/realtime/domain/entity"
	domainErrors "my-go-app/internal/realtime/domain/errors"
)

func (h *Handler) handleWorkflowCreate(ctx context.Context, client *wsentity.Client, msg domainEntity.Message) error {
	var payload domainEntity.PayloadWorkflow
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return domainErrors.Wrap(domainErrors.CodeValidation, err)
	}

	if err := h.uc.Workflow.CreateWorkflow(ctx, client.ProjectID, &payload.Workflow); err != nil {
		return err
	}

	broadcastMsg, _ := client.Factory().CreateWorkflow(payload.Workflow)
	h.hub.Broadcast(client.RoomID(), broadcastMsg, client)
	return nil
}

func (h *Handler) handleWorkflowUpdate(ctx context.Context, client *wsentity.Client, msg domainEntity.Message) error {
	var payload domainEntity.PayloadWorkflow
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return domainErrors.Wrap(domainErrors.CodeValidation, err)
	}

	if err := h.uc.Workflow.UpdateWorkflow(ctx, client.ProjectID, &payload.Workflow); err != nil {
		return err
	}

	broadcastMsg, _ := client.Factory().UpdateWorkflow(payload.Workflow)
	h.hub.Broadcast(client.RoomID(), broadcastMsg, client)
	return nil
}

func (h *Handler) handleWorkflowDelete(ctx context.Context, client *wsentity.Client, msg domainEntity.Message) error {
	var payload domainEntity.PayloadWorkflowDelete
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return domainErrors.Wrap(domainErrors.CodeValidation, err)
	}

	if err := h.uc.Workflow.RemoveWorkflow(ctx, client.ProjectID, payload.ID); err != nil {
		return err
	}

	broadcastMsg, _ := client.Factory().DeleteWorkflow(payload.ID)
	h.hub.Broadcast(client.RoomID(), broadcastMsg, client)
	return nil
}

func (h *Handler) handleElementWorkflowCreate(ctx context.Context, client *wsentity.Client, msg domainEntity.Message) error {
	var payload domainEntity.PayloadElementWorkflow
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return domainErrors.Wrap(domainErrors.CodeValidation, err)
	}

	if err := h.uc.ElementWorkflows.CreateElementWorkflow(ctx, client.ProjectID, &payload.Workflow); err != nil {
		return err
	}

	broadcastMsg, _ := client.Factory().CreateElementWorkflow(payload.Workflow)
	h.hub.Broadcast(client.RoomID(), broadcastMsg, client)
	return nil
}

func (h *Handler) handleElementWorkflowUpdate(ctx context.Context, client *wsentity.Client, msg domainEntity.Message) error {
	var payload domainEntity.PayloadElementWorkflow
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return domainErrors.Wrap(domainErrors.CodeValidation, err)
	}

	if err := h.uc.ElementWorkflows.UpdateElementWorkflow(ctx, client.ProjectID, &payload.Workflow); err != nil {
		return err
	}

	broadcastMsg, _ := client.Factory().UpdateElementWorkflow(payload.Workflow)
	h.hub.Broadcast(client.RoomID(), broadcastMsg, client)
	return nil
}

func (h *Handler) handleElementWorkflowDelete(ctx context.Context, client *wsentity.Client, msg domainEntity.Message) error {
	var payload domainEntity.PayloadElementWorkflowDelete
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return domainErrors.Wrap(domainErrors.CodeValidation, err)
	}

	if err := h.uc.ElementWorkflows.RemoveElementWorkflow(ctx, client.ProjectID, payload.ElementID); err != nil {
		return err
	}

	broadcastMsg, _ := client.Factory().DeleteElementWorkflow(payload.ElementID)
	h.hub.Broadcast(client.RoomID(), broadcastMsg, client)
	return nil
}
