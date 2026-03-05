package dto

import "encoding/json"

// CreateEventWorkflowRequest contains the required fields to create a new event workflow.
type CreateEventWorkflowRequest struct {
	ProjectID   string          `json:"projectId"   validate:"required"`
	Name        string          `json:"name"        validate:"required,min=1,max=255"`
	Description *string         `json:"description" validate:"omitempty,max=1000"`
	CanvasData  json.RawMessage `json:"canvasData"`
	Handlers    json.RawMessage `json:"handlers"`
	Enabled     *bool           `json:"enabled"`
}

// UpdateEventWorkflowRequest contains the patchable fields on an event workflow.
type UpdateEventWorkflowRequest struct {
	Name        string          `json:"name"        validate:"omitempty,min=1,max=255"`
	Description *string         `json:"description" validate:"omitempty,max=1000"`
	CanvasData  json.RawMessage `json:"canvasData"`
	Handlers    json.RawMessage `json:"handlers"`
	Enabled     *bool           `json:"enabled"`
}

// UpdateEventWorkflowEnabledRequest contains the enabled flag to toggle a workflow.
type UpdateEventWorkflowEnabledRequest struct {
	Enabled *bool `json:"enabled" validate:"required"`
}