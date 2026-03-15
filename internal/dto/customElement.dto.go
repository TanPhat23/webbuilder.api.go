package dto

import "encoding/json"

// CreateCustomElementRequest contains the required fields to create a new custom element.
type CreateCustomElementRequest struct {
	Name         string          `json:"name"         validate:"required,min=1,max=255"`
	Structure    json.RawMessage `json:"structure"    validate:"required"`
	TypeId       *string         `json:"typeId"`
	Description  *string         `json:"description"  validate:"omitempty,max=1000"`
	Category     *string         `json:"category"     validate:"omitempty,max=100"`
	Icon         *string         `json:"icon"         validate:"omitempty,max=255"`
	Thumbnail    *string         `json:"thumbnail"    validate:"omitempty,max=255"`
	DefaultProps json.RawMessage `json:"defaultProps"`
	Tags         *string         `json:"tags"         validate:"omitempty,max=500"`
	IsPublic     bool            `json:"isPublic"`
	Version      string          `json:"version"      validate:"omitempty"`
}

// UpdateCustomElementRequest contains the patchable fields on a custom element.
type UpdateCustomElementRequest struct {
	Name         *string         `json:"name"         validate:"omitempty,min=1,max=255"`
	Structure    json.RawMessage `json:"structure"`
	TypeId       *string         `json:"typeId"`
	Description  *string         `json:"description"  validate:"omitempty,max=1000"`
	Category     *string         `json:"category"     validate:"omitempty,max=100"`
	Icon         *string         `json:"icon"         validate:"omitempty,max=255"`
	Thumbnail    *string         `json:"thumbnail"    validate:"omitempty,max=255"`
	DefaultProps json.RawMessage `json:"defaultProps"`
	Tags         *string         `json:"tags"         validate:"omitempty,max=500"`
	IsPublic     *bool           `json:"isPublic"`
	Version      *string         `json:"version"      validate:"omitempty"`
}

// DuplicateCustomElementRequest contains the required fields to duplicate a custom element.
type DuplicateCustomElementRequest struct {
	NewName string `json:"newName" validate:"required,min=1,max=255"`
}