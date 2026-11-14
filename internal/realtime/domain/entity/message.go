package entity

import (
	"encoding/json"
	"time"
)

// MessageType defines the action being performed via WebSocket
type MessageType string

const (
	// Presence & Connection
	TypeJoin     MessageType = "join"     // User joined page
	TypeLeave    MessageType = "leave"    // User left page
	TypePresence MessageType = "presence" // Cursor movement/selection

	// Element Operations
	TypeElementCreate  MessageType = "element:create"  // New element added
	TypeElementUpdate  MessageType = "element:update"  // Granular update (styles, props)
	TypeElementDelete  MessageType = "element:delete"  // Element removed
	TypeElementMove    MessageType = "element:move"    // Element reordered or reparented
	TypeElementContent MessageType = "element:content" // Real-time text/content editing

	// Collaboration Features
	TypeCommentCreate MessageType = "comment:create" // New comment added
	TypeCommentUpdate MessageType = "comment:update" // Comment resolved/edited
	TypeCommentDelete MessageType = "comment:delete" // Comment removed

	// Workflow Operations
	TypeWorkflowCreate MessageType = "workflow:create"
	TypeWorkflowUpdate MessageType = "workflow:update"
	TypeWorkflowDelete MessageType = "workflow:delete"

	// Element Workflow Operations
	TypeElementWorkflowCreate MessageType = "element-workflow:create"
	TypeElementWorkflowUpdate MessageType = "element-workflow:update"
	TypeElementWorkflowDelete MessageType = "element-workflow:delete"

	// Logic (variables and flows) of the project or one page
	TypeLogicUpdate MessageType = "logic:update" // Patch a logic document / its broadcast
	TypeLogicMove   MessageType = "logic:move"   // Atomically move a definition between scopes
	TypeLogicSync   MessageType = "logic:sync"   // Current project and page logic

	// Synchronization
	TypeSync        MessageType = "sync:page"    // Initial state request/response
	TypeSyncProject MessageType = "sync:project" // Project-wide sync
	TypeSnapshot    MessageType = "snapshot"     // Save/Load full state
	TypeError       MessageType = "error"        // Server-side error reporting
)

// Message is the top-level envelope for all WebSocket communication
type Message struct {
	Type      MessageType     `json:"type"`
	ProjectID string          `json:"projectId"`
	PageID    string          `json:"pageId"`
	UserID    string          `json:"userId,omitempty"`
	Timestamp int64           `json:"timestamp"`
	Payload   json.RawMessage `json:"payload,omitempty"`

	// ClientOpID is an optional client-generated operation id. Clients may set it on element
	// operations; the server echoes it on the resulting broadcast (the sender's ack) and on
	// error replies so the client can match them to the pending operation.
	ClientOpID string `json:"clientOpId,omitempty"`
	// RequestID is an optional client-generated request id. The server echoes it on the reply
	// to that request (the element broadcast, the sync response, or an error) so the client can
	// resolve its pending request.
	RequestID string `json:"requestId,omitempty"`
	// Seq is the per-page, strictly increasing sequence number assigned by the server to
	// element:create/update/move/delete broadcasts. Zero (omitted) for all other messages.
	Seq int64 `json:"seq,omitempty"`
}

type PayloadSyncProject struct {
	Elements         []Element              `json:"elements,omitempty"`
	Workflows        []EventWorkflow        `json:"workflows,omitempty"`
	ElementWorkflows []ElementEventWorkflow `json:"elementWorkflows,omitempty"`
	Users            []UserPresence         `json:"users,omitempty"`
}

// PayloadPresence tracks user activity for "Who's here" and cursor pins
type PayloadPresence struct {
	UserID    string         `json:"userId"`
	UserName  string         `json:"userName"`
	CursorX   float64        `json:"cursorX"`
	CursorY   float64        `json:"cursorY"`
	ElementID string         `json:"elementId,omitempty"`
	Meta      map[string]any `json:"meta,omitempty"`
}

// PayloadElementCreate used when a new component is dragged onto the canvas
type PayloadElementCreate struct {
	Element  Element `json:"element"`
	Position *int    `json:"position,omitempty"`
}

// PayloadElementUpdate for granular patches to avoid sending the whole object
//
// Semantics are merge, not replace: Styles and Settings are deep-merged into the stored maps
// (a null value deletes the key). Properties set the element's own fields (name, content,
// href, src, tailwindStyles, isLocked). The top-level Name/Content/Src/Href/TailwindStyles
// fields are the form the web client sends and are equivalent to the same keys in Properties.
type PayloadElementUpdate struct {
	ID             string         `json:"id"`
	Properties     map[string]any `json:"properties,omitempty"`
	Styles         map[string]any `json:"styles,omitempty"`
	Settings       map[string]any `json:"settings,omitempty"`
	Name           *string        `json:"name,omitempty"`
	Content        *string        `json:"content,omitempty"`
	Src            *string        `json:"src,omitempty"`
	Href           *string        `json:"href,omitempty"`
	TailwindStyles *string        `json:"tailwindStyles,omitempty"`
}

// PayloadElementMove for drag-and-drop reordering
type PayloadElementMove struct {
	ID       string  `json:"id"`
	ParentID *string `json:"parentId"` // Null if root
	Order    int     `json:"order"`
	OrderKey string  `json:"orderKey,omitempty"`
}

// PayloadElementDelete
type PayloadElementDelete struct {
	ID string `json:"id"`
}

// PayloadElementContent for real-time text input synchronization
type PayloadElementContent struct {
	ID      string `json:"id"`
	Content string `json:"content"`
}

// PayloadComment for real-time feedback
type PayloadComment struct {
	ElementID string         `json:"elementId"`
	Comment   ElementComment `json:"comment"`
}

// PayloadWorkflow
type PayloadWorkflow struct {
	Workflow EventWorkflow `json:"workflow"`
}

// PayloadWorkflowDelete
type PayloadWorkflowDelete struct {
	ID string `json:"id"`
}

// PayloadElementWorkflow
type PayloadElementWorkflow struct {
	Workflow ElementEventWorkflow `json:"workflow"`
}

// PayloadElementWorkflowDelete
type PayloadElementWorkflowDelete struct {
	ElementID string `json:"elementId"`
}

// PayloadSync provides the initial state to a joining client
type PayloadSync struct {
	Elements         []Element              `json:"elements"`
	ElementWorkflows []ElementEventWorkflow `json:"elementWorkflows,omitempty"`
	Users            []UserPresence         `json:"users"`
	// Seq is the page's current sequence number at the time of the snapshot (0 if no ops yet).
	// Element ops with seq <= Seq are already contained in Elements.
	Seq int64 `json:"seq"`
}

// Logic scopes for logic:update.
const (
	LogicScopeProject = "project"
	LogicScopePage    = "page"
)

// PayloadLogicUpdate is what a client sends: a merge patch (null deletes a key)
// for the project's logic, or for the page named by the message's pageId.
type PayloadLogicUpdate struct {
	Scope     string         `json:"scope"`
	Patch     map[string]any `json:"patch"`
	NullPaths [][]string     `json:"nullPaths,omitempty"`
}

// PayloadLogicChanged is what the server broadcasts after a committed update:
// the complete resulting document for the scope, so applying it is idempotent.
// PageID is set only for page scope.
// Revision increases with every commit of that document.
type PayloadLogicChanged struct {
	Scope    string         `json:"scope"`
	PageID   string         `json:"pageId,omitempty"`
	Logic    map[string]any `json:"logic"`
	Revision int64          `json:"revision"`
}

// PayloadLogicSync carries both documents that apply to the joined page, with
// their revisions so a client can ignore a sync older than what it has.
type PayloadLogicSync struct {
	Project         map[string]any `json:"project"`
	Page            map[string]any `json:"page,omitempty"`
	PageID          string         `json:"pageId,omitempty"`
	ProjectRevision int64          `json:"projectRevision"`
	PageRevision    int64          `json:"pageRevision"`
}

// PayloadSnapshot allows the client to request a forced save to DB
type PayloadSnapshot struct {
	Name string       `json:"name"`
	Type SnapshotType `json:"type"` // "working" or "version"
}

// PayloadError
type PayloadError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	// ClientOpID echoes the failed operation's clientOpId, when the client supplied one.
	ClientOpID string `json:"clientOpId,omitempty"`
}

// Presence represents cursor/selection state for a user in a sync message.
type Presence struct {
	CursorX   float64 `json:"cursorX"`
	CursorY   float64 `json:"cursorY"`
	ElementID *string `json:"elementId,omitempty"`
}

// UserPresence represents a synchronized user for `sync:page` payloads.
// Matches the client-side SyncUser interface shape:
//
//	{
//	  userId: string;
//	  userName: string;
//	  presence: { cursorX:number, cursorY:number, elementId?: string | null }
//	}
type UserPresence struct {
	UserID   string   `json:"userId"`
	UserName string   `json:"userName"`
	Presence Presence `json:"presence"`
}

// NewMessage helper to construct a standard response
func NewMessage(msgType MessageType, projectID, pageID string, payload any) (*Message, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return &Message{
		Type:      msgType,
		ProjectID: projectID,
		PageID:    pageID,
		Timestamp: time.Now().UnixMilli(),
		Payload:   data,
	}, nil
}
