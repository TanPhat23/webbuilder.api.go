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
func (f *MessageFactory) CreateElement(element domainEntity.Element) (domainEntity.Message, error) {
	element = sanitizeElements([]domainEntity.Element{element})[0]
	return f.Create(domainEntity.TypeElementCreate, domainEntity.PayloadElementCreate{
		Element: element,
	})
}

// UpdateElement constructs a message for granular property or style changes.
func (f *MessageFactory) UpdateElement(payload domainEntity.PayloadElementUpdate) (domainEntity.Message, error) {
	return f.Create(domainEntity.TypeElementUpdate, payload)
}

// DeleteElement constructs a message for element removals.
func (f *MessageFactory) DeleteElement(elementID string) (domainEntity.Message, error) {
	return f.Create(domainEntity.TypeElementDelete, domainEntity.PayloadElementDelete{
		ID: elementID,
	})
}

// MoveElement constructs a message for reordering or reparenting elements.
func (f *MessageFactory) MoveElement(payload domainEntity.PayloadElementMove) (domainEntity.Message, error) {
	return f.Create(domainEntity.TypeElementMove, payload)
}

// SyncPage constructs a synchronization message with the full page state.
// Workflows (project-level) are intentionally excluded from page-level syncs.
//
// Sanitizes elements prior to creating the message so all required properties
// are present for the frontend (prevents missing properties from causing UI hangs).
func (f *MessageFactory) SyncPage(elements []domainEntity.Element, elementWorkflows []domainEntity.ElementEventWorkflow, users []domainEntity.UserPresence) (domainEntity.Message, error) {
	return f.SyncPageAt(elements, elementWorkflows, users, 0)
}

// SyncPageAt is SyncPage that also carries the page's current op sequence number.
func (f *MessageFactory) SyncPageAt(elements []domainEntity.Element, elementWorkflows []domainEntity.ElementEventWorkflow, users []domainEntity.UserPresence, seq int64) (domainEntity.Message, error) {
	sanitized := sanitizeElements(elements)
	return f.Create(domainEntity.TypeSync, domainEntity.PayloadSync{
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

func (f *MessageFactory) SyncProject(elements []domainEntity.Element, workflows []domainEntity.EventWorkflow, elementWorkflows []domainEntity.ElementEventWorkflow, users []domainEntity.UserPresence) (domainEntity.Message, error) {
	return f.Create(domainEntity.TypeSyncProject, domainEntity.PayloadSyncProject{
		Elements:         elements,
		Workflows:        workflows,
		ElementWorkflows: elementWorkflows,
		Users:            users,
	})
}

// CreateWorkflow constructs a message for new workflow additions.
func (f *MessageFactory) CreateWorkflow(workflow domainEntity.EventWorkflow) (domainEntity.Message, error) {
	return f.Create(domainEntity.TypeWorkflowCreate, domainEntity.PayloadWorkflow{
		Workflow: workflow,
	})
}

// UpdateWorkflow constructs a message for workflow changes.
func (f *MessageFactory) UpdateWorkflow(workflow domainEntity.EventWorkflow) (domainEntity.Message, error) {
	return f.Create(domainEntity.TypeWorkflowUpdate, domainEntity.PayloadWorkflow{
		Workflow: workflow,
	})
}

// DeleteWorkflow constructs a message for workflow removals.
func (f *MessageFactory) DeleteWorkflow(workflowID string) (domainEntity.Message, error) {
	return f.Create(domainEntity.TypeWorkflowDelete, domainEntity.PayloadWorkflowDelete{
		ID: workflowID,
	})
}

// CreateElementWorkflow constructs a message for new element-workflow links.
func (f *MessageFactory) CreateElementWorkflow(workflow domainEntity.ElementEventWorkflow) (domainEntity.Message, error) {
	return f.Create(domainEntity.TypeElementWorkflowCreate, domainEntity.PayloadElementWorkflow{
		Workflow: workflow,
	})
}

// UpdateElementWorkflow constructs a message for element-workflow link changes.
func (f *MessageFactory) UpdateElementWorkflow(workflow domainEntity.ElementEventWorkflow) (domainEntity.Message, error) {
	return f.Create(domainEntity.TypeElementWorkflowUpdate, domainEntity.PayloadElementWorkflow{
		Workflow: workflow,
	})
}

// DeleteElementWorkflow constructs a message for element-workflow link removals.
func (f *MessageFactory) DeleteElementWorkflow(elementID string) (domainEntity.Message, error) {
	return f.Create(domainEntity.TypeElementWorkflowDelete, domainEntity.PayloadElementWorkflowDelete{
		ElementID: elementID,
	})
}

// LogicChanged constructs the broadcast for a committed logic update.
func (f *MessageFactory) LogicChanged(payload domainEntity.PayloadLogicChanged) (domainEntity.Message, error) {
	return f.Create(domainEntity.TypeLogicUpdate, payload)
}

// LogicSync constructs the current logic for a joining or resyncing client.
func (f *MessageFactory) LogicSync(payload domainEntity.PayloadLogicSync) (domainEntity.Message, error) {
	return f.Create(domainEntity.TypeLogicSync, payload)
}

// BroadcastPresence constructs a message for broadcasting live user activity.
func (f *MessageFactory) BroadcastPresence(payload domainEntity.PayloadPresence) (domainEntity.Message, error) {
	return f.Create(domainEntity.TypePresence, payload)
}

// ErrorWithErr constructs a standardized error message and returns an error if marshalling fails.
func (f *MessageFactory) ErrorWithErr(code, errMsg string) (domainEntity.Message, error) {
	payload := domainEntity.PayloadError{
		Code:    code,
		Message: errMsg,
	}
	return f.Create(domainEntity.TypeError, payload)
}

// ErrorForOp constructs an error message that echoes the failed operation's clientOpId
// (both on the envelope and in the payload). With an empty clientOpId it equals Error.
func (f *MessageFactory) ErrorForOp(code, errMsg, clientOpID string) domainEntity.Message {
	msg, _ := f.Create(domainEntity.TypeError, domainEntity.PayloadError{
		Code:       code,
		Message:    errMsg,
		ClientOpID: clientOpID,
	})
	msg.ClientOpID = clientOpID
	return msg
}

// Error constructs a standardized error message.
// Use ErrorWithErr if you need to propagate marshaling errors.
func (f *MessageFactory) Error(code, errMsg string) domainEntity.Message {
	msg, _ := f.ErrorWithErr(code, errMsg)
	return msg
}

func (f *MessageFactory) Create(msgType domainEntity.MessageType, payload any) (domainEntity.Message, error) {
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
