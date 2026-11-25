package handlers

import (
	"context"
	"encoding/json"

	wsentity "my-go-app/internal/realtime/delivery/ws/entity"
	domainEntity "my-go-app/internal/realtime/domain/entity"
	domainErrors "my-go-app/internal/realtime/domain/errors"
	"my-go-app/internal/realtime/utils/validation"
)

func (h *Handler) joinPage(ctx context.Context, client *wsentity.Client, msg domainEntity.Message) error {
	if err := h.authorizeJoin(ctx, client, msg.PageID); err != nil {
		return err
	}

	oldPage := client.PageID()
	if oldPage != "" && oldPage != msg.PageID {
		_ = h.hub.Detach(ctx, client)
	}
	client.SetPageID(msg.PageID)
	if err := h.hub.Register(ctx, client); err != nil {
		return err
	}

	if err := h.sendPageSync(ctx, client, msg.RequestID); err != nil {
		return err
	}
	if err := h.sendLogicSync(ctx, client, msg.RequestID); err != nil {
		h.logger.Errorf("logic sync on join failed (project=%s page=%s): %v", client.ProjectID, client.PageID(), err)
	}
	return h.sendProjectSync(ctx, client, msg.RequestID)
}

func (h *Handler) authorizeJoin(ctx context.Context, client *wsentity.Client, pageID string) error {
	if pageID == "" {
		return domainErrors.New(domainErrors.CodeValidation, "pageId is required to join", nil)
	}
	if !validation.IsSafeID(pageID) {
		return domainErrors.New(domainErrors.CodeValidation, "invalid pageId", nil)
	}
	if h.pageChecker == nil {
		return nil
	}
	owns, err := h.pageChecker.PageBelongsToProject(ctx, client.ProjectID, pageID)
	if err != nil {
		return domainErrors.New(domainErrors.CodeInternal, "could not verify page access", err)
	}
	if !owns {
		return domainErrors.New(domainErrors.CodeAccessDenied, "page does not belong to this project", nil)
	}
	return nil
}

func (h *Handler) sendPageSync(ctx context.Context, client *wsentity.Client, requestID string) error {
	elements, seq, err := h.uc.Element.SyncPageWithSeq(ctx, client.ProjectID, client.PageID())
	if err != nil {
		return err
	}
	users := h.hub.RoomParticipants(client.RoomID())
	syncMsg, _ := client.NewMessageFactory().NewPageSyncAtMessage(elements, nil, users, seq)
	syncMsg.RequestID = requestID
	_ = client.Send(syncMsg)
	return nil
}

// sendProjectSync sends project state only to the joining client, not to every room.
// Broadcasting sync:project on every join creates a message storm during rapid
// reconnects; other clients already have the project state and receive granular
// updates via element/workflow operation broadcasts.
func (h *Handler) sendProjectSync(ctx context.Context, client *wsentity.Client, requestID string) error {
	elements, err := h.uc.Element.ListElementsByProject(ctx, client.ProjectID)
	if err != nil {
		return err
	}
	syncMsg, _ := client.NewMessageFactory().NewProjectSyncMessage(elements, nil, nil, nil)
	syncMsg.RequestID = requestID
	_ = client.Send(syncMsg)
	return nil
}

func (h *Handler) leavePage(ctx context.Context, client *wsentity.Client, msg domainEntity.Message) error {
	if client.PageID() != "" {
		_ = h.hub.Detach(ctx, client)
		client.SetPageID("")
	}
	return nil
}

func (h *Handler) updatePresence(ctx context.Context, client *wsentity.Client, msg domainEntity.Message) error {
	var payload domainEntity.PayloadPresence
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return domainErrors.Wrap(domainErrors.CodeValidation, err)
	}

	// Store the client's latest presence locally so sync responses reflect it,
	// then broadcast the presence to other clients.
	client.SetPresence(payload)

	broadcastMsg, _ := client.NewMessageFactory().NewPresenceMessage(payload)
	h.hub.Broadcast(client.RoomID(), broadcastMsg, client)
	return nil
}

func (h *Handler) syncPage(ctx context.Context, client *wsentity.Client, msg domainEntity.Message) error {
	pageID := client.PageID()
	if pageID == "" {
		return domainErrors.ErrNotJoined
	}

	elements, seq, err := h.uc.Element.SyncPageWithSeq(ctx, client.ProjectID, pageID)
	if err != nil {
		return err
	}

	users := h.hub.RoomParticipants(client.RoomID())
	syncMsg, _ := client.NewMessageFactory().NewPageSyncAtMessage(elements, nil, users, seq)
	syncMsg.RequestID = msg.RequestID
	_ = client.Send(syncMsg)
	return nil
}

func (h *Handler) syncProject(ctx context.Context, client *wsentity.Client, msg domainEntity.Message) error {
	elements, err := h.uc.Element.ListElementsByProject(ctx, client.ProjectID)
	if err != nil {
		return err
	}

	syncMsg, err := client.NewMessageFactory().NewProjectSyncMessage(elements, nil, nil, nil)
	if err != nil {
		return err
	}

	syncMsg.RequestID = msg.RequestID
	return client.Send(syncMsg)
}
