package entity

import (
	"encoding/json"
	"fmt"

	domainEntity "my-go-app/internal/realtime/domain/entity"
)

// MessageFactory provides a structured way to create WebSocket messages
// within a specific project and page context.
//
// Optionally carries a user context so created message envelopes can set
// the `userId` field automatically.
type MessageFactory struct {
	projectID string
	pageID    string
	userID    string
}

// NewMessageFactory initializes a factory with routing context.
// Optionally accepts a userID as a third argument.
func NewMessageFactory(projectID, pageID string, userID ...string) *MessageFactory {
	uid := ""
	if len(userID) > 0 {
		uid = userID[0]
	}
	return &MessageFactory{
		projectID: projectID,
		pageID:    pageID,
		userID:    uid,
	}
}

// CreateElement constructs a message for new element additions.
func (f *MessageFactory) NewElementCreatedMessage(element domainEntity.Element) (domainEntity.Message, error) {
	element = sanitizeElements([]domainEntity.Element{element})[0]
	return f.NewMessage(domainEntity.TypeElementCreate, domainEntity.PayloadElementCreate{
		Element: element,
	})
}

// UpdateElement constructs a message for granular property or style changes.
func (f *MessageFactory) NewElementUpdatedMessage(payload domainEntity.PayloadElementUpdate) (domainEntity.Message, error) {
	return f.NewMessage(domainEntity.TypeElementUpdate, payload)
}

// DeleteElement constructs a message for element removals.
func (f *MessageFactory) NewElementDeletedMessage(elementID string) (domainEntity.Message, error) {
	return f.NewMessage(domainEntity.TypeElementDelete, domainEntity.PayloadElementDelete{
		ID: elementID,
	})
}

// MoveElement constructs a message for reordering or reparenting elements.
func (f *MessageFactory) NewElementMovedMessage(payload domainEntity.PayloadElementMove) (domainEntity.Message, error) {
	return f.NewMessage(domainEntity.TypeElementMove, payload)
}

// SyncPage constructs a synchronization message with the full page state.
// Workflows (project-level) are intentionally excluded from page-level syncs.
//
// Sanitizes elements prior to creating the message so all required properties
// are present for the frontend (prevents missing properties from causing UI hangs).
func (f *MessageFactory) NewPageSyncMessage(elements []domainEntity.Element, elementWorkflows []domainEntity.ElementEventWorkflow, users []domainEntity.UserPresence) (domainEntity.Message, error) {
	return f.NewPageSyncAtMessage(elements, elementWorkflows, users, 0)
}

// NewPageSyncAtMessage is SyncPage that also carries the page's current op sequence number.
func (f *MessageFactory) NewPageSyncAtMessage(elements []domainEntity.Element, elementWorkflows []domainEntity.ElementEventWorkflow, users []domainEntity.UserPresence, seq int64) (domainEntity.Message, error) {
	sanitized := sanitizeElements(elements)
	return f.NewMessage(domainEntity.TypeSync, domainEntity.PayloadSync{
		Elements:         sanitized,
		ElementWorkflows: elementWorkflows,
		Users:            users,
		Seq:              seq,
	})
}

// sanitizeElements ensures element objects contain safe default properties
// (e.g., non-nil pointers and empty maps) so that JSON payloads always include
// the expected fields and the frontend doesn't attempt to access missing values.
func sanitizeElements(elements []domainEntity.Element) []domainEntity.Element {
	out := make([]domainEntity.Element, 0, len(elements))
	for _, el := range elements {
		e := el // copy to avoid mutating original

		// Ensure pointer/string fields exist (so `omitempty` won't drop them)
		if e.Name == nil {
			empty := ""
			e.Name = &empty
		}
		if e.Content == nil {
			empty := ""
			e.Content = &empty
		}
		if e.Href == nil {
			empty := ""
			e.Href = &empty
		}
		if e.Src == nil {
			empty := ""
			e.Src = &empty
		}
		if e.TailwindStyles == nil {
			empty := ""
			e.TailwindStyles = &empty
		}

		if e.Styles == nil {
			e.Styles = map[string]any{}
		}

		e.Elements = sanitizeElements(e.Elements)

		out = append(out, e)
	}
	return out
}

func (f *MessageFactory) NewProjectSyncMessage(elements []domainEntity.Element, workflows []domainEntity.EventWorkflow, elementWorkflows []domainEntity.ElementEventWorkflow, users []domainEntity.UserPresence) (domainEntity.Message, error) {
	return f.NewMessage(domainEntity.TypeSyncProject, domainEntity.PayloadSyncProject{
		Elements:         elements,
		Workflows:        workflows,
		ElementWorkflows: elementWorkflows,
		Users:            users,
	})
}

// CreateWorkflow constructs a message for new workflow additions.
func (f *MessageFactory) NewWorkflowCreatedMessage(workflow domainEntity.EventWorkflow) (domainEntity.Message, error) {
	return f.NewMessage(domainEntity.TypeWorkflowCreate, domainEntity.PayloadWorkflow{
		Workflow: workflow,
	})
}

// UpdateWorkflow constructs a message for workflow changes.
func (f *MessageFactory) NewWorkflowUpdatedMessage(workflow domainEntity.EventWorkflow) (domainEntity.Message, error) {
	return f.NewMessage(domainEntity.TypeWorkflowUpdate, domainEntity.PayloadWorkflow{
		Workflow: workflow,
	})
}

// DeleteWorkflow constructs a message for workflow removals.
func (f *MessageFactory) NewWorkflowDeletedMessage(workflowID string) (domainEntity.Message, error) {
	return f.NewMessage(domainEntity.TypeWorkflowDelete, domainEntity.PayloadWorkflowDelete{
		ID: workflowID,
	})
}

// CreateElementWorkflow constructs a message for new element-workflow links.
func (f *MessageFactory) NewElementWorkflowCreatedMessage(workflow domainEntity.ElementEventWorkflow) (domainEntity.Message, error) {
	return f.NewMessage(domainEntity.TypeElementWorkflowCreate, domainEntity.PayloadElementWorkflow{
		Workflow: workflow,
	})
}

// UpdateElementWorkflow constructs a message for element-workflow link changes.
func (f *MessageFactory) NewElementWorkflowUpdatedMessage(workflow domainEntity.ElementEventWorkflow) (domainEntity.Message, error) {
	return f.NewMessage(domainEntity.TypeElementWorkflowUpdate, domainEntity.PayloadElementWorkflow{
		Workflow: workflow,
	})
}

// DeleteElementWorkflow constructs a message for element-workflow link removals.
func (f *MessageFactory) NewElementWorkflowDeletedMessage(elementID string) (domainEntity.Message, error) {
	return f.NewMessage(domainEntity.TypeElementWorkflowDelete, domainEntity.PayloadElementWorkflowDelete{
		ElementID: elementID,
	})
}

// NewLogicChangedMessage constructs the broadcast for a committed logic update.
func (f *MessageFactory) NewLogicChangedMessage(payload domainEntity.PayloadLogicChanged) (domainEntity.Message, error) {
	return f.NewMessage(domainEntity.TypeLogicUpdate, payload)
}

// NewLogicSyncMessage constructs the current logic for a joining or resyncing client.
func (f *MessageFactory) NewLogicSyncMessage(payload domainEntity.PayloadLogicSync) (domainEntity.Message, error) {
	return f.NewMessage(domainEntity.TypeLogicSync, payload)
}

// NewPresenceMessage constructs a message for broadcasting live user activity.
func (f *MessageFactory) NewPresenceMessage(payload domainEntity.PayloadPresence) (domainEntity.Message, error) {
	return f.NewMessage(domainEntity.TypePresence, payload)
}

// TryErrorMessage constructs a standardized error message and returns an error if marshalling fails.
func (f *MessageFactory) TryErrorMessage(code, errMsg string) (domainEntity.Message, error) {
	payload := domainEntity.PayloadError{
		Code:    code,
		Message: errMsg,
	}
	return f.NewMessage(domainEntity.TypeError, payload)
}

// NewErrorMessageForOp constructs an error message that echoes the failed operation's clientOpId
// (both on the envelope and in the payload). With an empty clientOpId it equals Error.
func (f *MessageFactory) NewErrorMessageForOp(code, errMsg, clientOpID string) domainEntity.Message {
	return f.buildFallbackError(code, errMsg, clientOpID)
}

// Error constructs a standardized error message.
// Use TryErrorMessage if you need to propagate marshaling errors.
func (f *MessageFactory) NewErrorMessage(code, errMsg string) domainEntity.Message {
	return f.buildFallbackError(code, errMsg, "")
}

func (f *MessageFactory) buildFallbackError(code, errMsg, clientOpID string) domainEntity.Message {
	msg, err := f.NewMessage(domainEntity.TypeError, domainEntity.PayloadError{
		Code:       code,
		Message:    errMsg,
		ClientOpID: clientOpID,
	})
	if err != nil {
		return domainEntity.Message{
			Type:       domainEntity.TypeError,
			ProjectID:  f.projectID,
			PageID:     f.pageID,
			UserID:     f.userID,
			ClientOpID: clientOpID,
		}
	}
	msg.ClientOpID = clientOpID
	return msg
}

func (f *MessageFactory) NewMessage(msgType domainEntity.MessageType, payload any) (domainEntity.Message, error) {
	m, err := domainEntity.NewMessage(msgType, f.projectID, f.pageID, payload)
	if err != nil {
		return domainEntity.Message{}, fmt.Errorf("failed to create message: %w", err)
	}
	if payload == nil {
		m.Payload = nil
	}
	if f.userID != "" {
		m.UserID = f.userID
	}
	return *m, nil
}

// DecodePayload unmarshals raw payload data into a concrete type.
func DecodePayload[T any](m domainEntity.Message) (T, error) {
	var dst T
	if len(m.Payload) == 0 {
		return dst, nil
	}
	if err := json.Unmarshal(m.Payload, &dst); err != nil {
		return dst, fmt.Errorf("failed to decode payload: %w", err)
	}
	return dst, nil
}
