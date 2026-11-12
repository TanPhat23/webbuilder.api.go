package dto

import "encoding/json"

type CreateCustomElementRequest struct {
	Structure    json.RawMessage `json:"structure"    validate:"required"`
	DefaultProps json.RawMessage `json:"defaultProps" validate:"omitempty"`
	Name         string          `json:"name"         validate:"required,min=1,max=255"`
	Version      string          `json:"version"      validate:"required,min=1,max=50"`
	TypeId       *string         `json:"typeId"       validate:"omitempty,min=1,max=255"`
	Description  *string         `json:"description"  validate:"omitempty,min=1,max=1000"`
	Category     *string         `json:"category"     validate:"omitempty,min=1,max=100"`
	Icon         *string         `json:"icon"         validate:"omitempty,min=1,max=255"`
	Thumbnail    *string         `json:"thumbnail"    validate:"omitempty,min=1,max=255"`
	Tags         *string         `json:"tags"         validate:"omitempty,min=1,max=500"`
	IsPublic     bool            `json:"isPublic"`
}

type UpdateCustomElementRequest struct {
	Structure    json.RawMessage `json:"structure"    validate:"omitempty"`
	DefaultProps json.RawMessage `json:"defaultProps" validate:"omitempty"`
	Name         *string         `json:"name"         validate:"omitempty,min=1,max=255"`
	TypeId       *string         `json:"typeId"       validate:"omitempty,min=1,max=255"`
	Description  *string         `json:"description"  validate:"omitempty,min=1,max=1000"`
	Category     *string         `json:"category"     validate:"omitempty,min=1,max=100"`
	Icon         *string         `json:"icon"         validate:"omitempty,min=1,max=255"`
	Thumbnail    *string         `json:"thumbnail"    validate:"omitempty,min=1,max=255"`
	Tags         *string         `json:"tags"         validate:"omitempty,min=1,max=500"`
	IsPublic     *bool           `json:"isPublic"`
	Version      *string         `json:"version"      validate:"omitempty,min=1,max=50"`
}

type DuplicateCustomElementRequest struct {
	NewName string `json:"newName" validate:"required,min=1,max=255"`
}
