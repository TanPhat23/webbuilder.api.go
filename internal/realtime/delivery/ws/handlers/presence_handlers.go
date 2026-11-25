package handlers

import (
	"context"
	"encoding/json"

	wsentity "my-go-app/internal/realtime/delivery/ws/entity"
	domainEntity "my-go-app/internal/realtime/domain/entity"
	domainErrors "my-go-app/internal/realtime/domain/errors"
	"my-go-app/internal/realtime/utils/validation"
)

func (h *Handler) handleJoin(ctx context.Context, client *wsentity.Client, msg domainEntity.Message) error {
	if msg.PageID == "" {
		return domainErrors.New(domainErrors.CodeValidation, "pageId is required to join", nil)
	}
	if !validation.IsValidSafeID(msg.PageID) {
		return domainErrors.New(domainErrors.CodeValidation, "invalid pageId", nil)
	}
	if h.pageChecker != nil {
		owns, err := h.pageChecker.PageBelongsToProject(ctx, client.ProjectID, msg.PageID)
		if err != nil {
			return domainErrors.New(domainErrors.CodeInternal, "could not verify page access", err)
		}
		if !owns {
			return domainErrors.New(domainErrors.CodeAccessDenied, "page does not belong to this project", nil)
		}
	}

	oldPage := client.GetPageID()
	if oldPage != "" && oldPage != msg.PageID {
		_ = h.hub.Detach(ctx, client)
	}

	client.SetPageID(msg.PageID)
	if err := h.hub.Register(ctx, client); err != nil {
		return err
	}

	pageID := client.GetPageID()
	elements, seq, err := h.uc.Element.SyncPageWithSeq(ctx, client.ProjectID, pageID)
	if err != nil {
		return err
	}

	users := h.hub.GetRoomParticipants(client.RoomID())
	syncMsg, _ := client.Factory().SyncPageAt(elements, nil, users, seq)
	syncMsg.RequestID = msg.RequestID
	_ = client.Send(syncMsg)

	if err := h.sendLogicSync(ctx, client, msg.RequestID); err != nil {
		h.logger.Errorf("logic sync on join failed (project=%s page=%s): %v", client.ProjectID, pageID, err)
	}

	projElements, err := h.uc.Element.GetElementsByProjectID(ctx, client.ProjectID)
	if err != nil {
		return err
	}

	// Send project sync only to the joining client — NOT broadcast to all rooms.
	// Broadcasting sync:project to every room on every join creates a message storm
	// during rapid reconnects (each reconnect triggers a full project sync to ALL
	// other clients across ALL pages). The joining client needs the project state;
	// other clients already have it and will receive granular updates via element/workflow
	// operation broadcasts.
	syncProjMsg, _ := client.Factory().SyncProject(projElements, nil, nil, nil)
	syncProjMsg.RequestID = msg.RequestID
	_ = client.Send(syncProjMsg)

	return nil
}

func (h *Handler) handleLeave(ctx context.Context, client *wsentity.Client, msg domainEntity.Message) error {
	if client.GetPageID() != "" {
		_ = h.hub.Detach(ctx, client)
		client.SetPageID("")
	}
	return nil
}

func (h *Handler) handlePresence(ctx context.Context, client *wsentity.Client, msg domainEntity.Message) error {
	var payload domainEntity.PayloadPresence
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return domainErrors.Wrap(domainErrors.CodeValidation, err)
	}

	// Store the client's latest presence locally so sync responses reflect it,
	// then broadcast the presence to other clients.
	client.SetPresence(payload)

	broadcastMsg, _ := client.Factory().BroadcastPresence(payload)
	h.hub.Broadcast(client.RoomID(), broadcastMsg, client)
	return nil
}

func (h *Handler) handleSync(ctx context.Context, client *wsentity.Client, msg domainEntity.Message) error {
	pageID := client.GetPageID()
	if pageID == "" {
		return domainErrors.ErrNotJoined
	}

	elements, seq, err := h.uc.Element.SyncPageWithSeq(ctx, client.ProjectID, pageID)
	if err != nil {
		return err
	}

	users := h.hub.GetRoomParticipants(client.RoomID())
	syncMsg, _ := client.Factory().SyncPageAt(elements, nil, users, seq)
	syncMsg.RequestID = msg.RequestID
	_ = client.Send(syncMsg)
	return nil
}

func (h *Handler) handleSyncProject(ctx context.Context, client *wsentity.Client, msg domainEntity.Message) error {
	elements, err := h.uc.Element.GetElementsByProjectID(ctx, client.ProjectID)
	if err != nil {
		return err
	}

	syncMsg, err := client.Factory().SyncProject(elements, nil, nil, nil)
	if err != nil {
		return err
	}

	syncMsg.RequestID = msg.RequestID
	return client.Send(syncMsg)
}
