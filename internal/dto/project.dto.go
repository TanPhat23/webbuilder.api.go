package dto

import "encoding/json"

type CreateProjectRequest struct {
	// Id lets the client choose the project id; generated when empty.
	Id          string  `json:"id"          validate:"omitempty,uuid"`
	Styles      any     `json:"styles"      validate:"omitempty"`
	Header      any     `json:"header"      validate:"omitempty"`
	Name        string  `json:"name"        validate:"required,min=1,max=255"`
	Description *string `json:"description" validate:"omitempty,min=1,max=1000"`
	Subdomain   *string `json:"subdomain"   validate:"omitempty,max=63"`
	Published   bool    `json:"published"`
}

type UpdateProjectRequest struct {
	Styles      any             `json:"styles"      validate:"omitempty"`
	Header      any             `json:"header"      validate:"omitempty"`
	AuthConfig  json.RawMessage `json:"authConfig"  validate:"omitempty"`
	Name        *string         `json:"name"        validate:"omitempty,min=1,max=255"`
	Description *string         `json:"description" validate:"omitempty,min=1,max=1000"`
	Subdomain   *string         `json:"subdomain"   validate:"omitempty,max=63"`
	Published   *bool           `json:"published"`
}

type TransferOwnershipRequest struct {
	NewOwnerID string `json:"newOwnerId" validate:"required,min=1"`
}
