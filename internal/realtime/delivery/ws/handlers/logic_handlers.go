package handlers

import (
	"context"
	"encoding/json"

	wsentity "my-go-app/internal/realtime/delivery/ws/entity"
	domainEntity "my-go-app/internal/realtime/domain/entity"
	domainErrors "my-go-app/internal/realtime/domain/errors"
)

// logicActorKey serializes every logic write of a project (project and page
// scope alike) on this instance, so commits and their broadcasts keep one order.
func logicActorKey(projectID string) string { return projectID + ":logic" }

// handleLogicUpdate applies a patch to the project's or the current page's logic.
// The database transaction checks permission, validates the whole project
// (names must be unique across scopes) and persists. Only after it commits is the
// full resulting document broadcast, including to the sender as its ack. A rejected
// patch changes nothing and is reported only to the sender.
func (h *Handler) handleLogicUpdate(ctx context.Context, client *wsentity.Client, msg domainEntity.Message) error {
	if h.uc.Logic == nil {
		return domainErrors.New(domainErrors.CodeInternal, "logic editing is not available", nil)
	}
	var payload domainEntity.PayloadLogicUpdate
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return domainErrors.Wrap(domainErrors.CodeValidation, err)
	}
	if payload.Patch == nil {
		return domainErrors.New(domainErrors.CodeValidation, "patch is required", nil)
	}

	pageID := ""
	switch payload.Scope {
	case domainEntity.LogicScopeProject:
	case domainEntity.LogicScopePage:
		pageID = msg.PageID
		if pageID == "" {
			return domainErrors.ErrNotJoined
		}
	default:
		return domainErrors.New(domainErrors.CodeValidation, "scope must be \"project\" or \"page\"", nil)
	}

	var opErr error
	if err := h.pages.Do(ctx, logicActorKey(client.ProjectID), func() {
		doc, err := h.uc.Logic.Update(ctx, client.ProjectID, pageID, client.User.ID, payload.Patch)
		if err != nil {
			opErr = err
			return
		}
		changed := domainEntity.PayloadLogicChanged{Scope: payload.Scope, PageID: pageID, Logic: doc.Doc, Revision: doc.Revision}
		factory := wsentity.NewMessageFactory(client.ProjectID, pageID, client.User.ID)
		out, err := factory.LogicChanged(changed)
		if err != nil {
			h.logger.Errorf("failed to build logic broadcast (project=%s page=%s): %v", client.ProjectID, pageID, err)
			opErr = domainErrors.Wrap(domainErrors.CodeInternal, err)
			return
		}
		out.ClientOpID = msg.ClientOpID
		out.RequestID = msg.RequestID
		if payload.Scope == domainEntity.LogicScopeProject {
			h.hub.BroadcastProject(client.ProjectID, out, nil)
		} else {
			h.hub.Broadcast(client.ProjectID+":"+pageID, out, nil)
		}
	}); err != nil {
		return err
	}
	return opErr
}

// handleLogicSync replies to the requester with the current logic documents.
func (h *Handler) handleLogicSync(ctx context.Context, client *wsentity.Client, msg domainEntity.Message) error {
	return h.sendLogicSync(ctx, client, msg.RequestID)
}

func (h *Handler) sendLogicSync(ctx context.Context, client *wsentity.Client, requestID string) error {
	if h.uc.Logic == nil {
		return nil
	}
	if client.PageID == "" {
		return domainErrors.ErrNotJoined
	}
	project, page, err := h.uc.Logic.Get(ctx, client.ProjectID, client.PageID, client.User.ID)
	if err != nil {
		return err
	}
	// This read runs outside the logic actor, so an update may be broadcast
	// before this reply is sent; the revisions let the client drop it then.
	out, err := client.Factory().LogicSync(domainEntity.PayloadLogicSync{
		Project: project.Doc, Page: page.Doc, PageID: client.PageID,
		ProjectRevision: project.Revision, PageRevision: page.Revision,
	})
	if err != nil {
		return domainErrors.Wrap(domainErrors.CodeInternal, err)
	}
	out.RequestID = requestID
	return client.Send(out)
}
