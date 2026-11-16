// Package agentproto holds what the MCP side (internal/agent) and the realtime
// side (internal/realtime) agree on: Redis channel names and message shapes.
package agentproto

import "encoding/json"

const (
	RequestChannelPrefix  = "agent:req:" // + userID
	AckChannelPrefix      = "agent:ack:" // + requestID
	ResponseChannelPrefix = "agent:res:" // + requestID
)

const (
	CodeNoEditorSession = "no_editor_session"
	CodeEditorTimeout   = "editor_timeout"
	CodeInvalidEdit     = "invalid_edit"
	CodeForbidden       = "forbidden"
	CodeInternal        = "internal"
)

// Request is a tool call for the editor tab to run.
type Request struct {
	RequestID string          `json:"requestId"`
	UserID    string          `json:"userId"`
	ProjectID string          `json:"projectId"`
	PageID    string          `json:"pageId,omitempty"`
	Tool      string          `json:"tool"`
	Args      json.RawMessage `json:"args"`
}

// Issue points at the part of the arguments that is wrong, e.g. "operations[2].id".
type Issue struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}

type Error struct {
	Code    string  `json:"code"`
	Message string  `json:"message"`
	Issues  []Issue `json:"issues,omitempty"`
}

// Response is the tab's answer (or a relay failure).
type Response struct {
	RequestID string          `json:"requestId"`
	OK        bool            `json:"ok"`
	Result    json.RawMessage `json:"result,omitempty"`
	Error     *Error          `json:"error,omitempty"`
}
