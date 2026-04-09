package dto

import "encoding/json"

type CreatePageRequest struct {
	Styles json.RawMessage `json:"styles" validate:"omitempty"`
	Name   string          `json:"name"   validate:"required,min=1,max=255"`
	Type   string          `json:"type"   validate:"required,min=1,max=100"`
}

type UpdatePageRequest struct {
	Styles json.RawMessage `json:"styles" validate:"omitempty"`
	Name   *string         `json:"name"   validate:"omitempty,min=1,max=255"`
	Type   *string         `json:"type"   validate:"omitempty,min=1,max=100"`
}
