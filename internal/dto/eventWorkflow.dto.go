package dto

import "encoding/json"

type CreateEventWorkflowRequest struct {
	CanvasData  json.RawMessage `json:"canvasData"  validate:"omitempty"`
	Handlers    json.RawMessage `json:"handlers"    validate:"omitempty"`
	ProjectID   string          `json:"projectId"   validate:"required,min=1"`
	Name        string          `json:"name"        validate:"required,min=1,max=255"`
	Description *string         `json:"description" validate:"omitempty,min=1,max=1000"`
	Enabled     *bool           `json:"enabled"`
}

type UpdateEventWorkflowRequest struct {
	CanvasData  json.RawMessage `json:"canvasData"  validate:"omitempty"`
	Handlers    json.RawMessage `json:"handlers"    validate:"omitempty"`
	Name        *string         `json:"name"        validate:"omitempty,min=1,max=255"`
	Description *string         `json:"description" validate:"omitempty,min=1,max=1000"`
	Enabled     *bool           `json:"enabled"`
}

type UpdateEventWorkflowEnabledRequest struct {
	Enabled *bool `json:"enabled" validate:"required"`
}
