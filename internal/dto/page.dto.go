package dto

import "encoding/json"

// CreatePageRequest contains the required fields to create a new page.
type CreatePageRequest struct {
	Name   string          `json:"name"   validate:"required,min=1,max=255"`
	Type   string          `json:"type"   validate:"required"`
	Styles json.RawMessage `json:"styles,omitempty"`
}

// UpdatePageRequest contains the patchable fields on a page.
type UpdatePageRequest struct {
	Name   *string         `json:"name"   validate:"omitempty,min=1,max=255"`
	Type   *string         `json:"type"   validate:"omitempty"`
	Styles json.RawMessage `json:"styles,omitempty"`
}