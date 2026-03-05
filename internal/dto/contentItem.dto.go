package dto

import "my-go-app/internal/models"

// UpdateContentItemRequest contains the patchable fields on a content item.
type UpdateContentItemRequest struct {
	Published   *bool                      `json:"published"`
	Slug        *string                    `json:"slug"        validate:"omitempty,min=1,max=255"`
	Title       *string                    `json:"title"       validate:"omitempty,min=1,max=255"`
	FieldValues []models.ContentFieldValue `json:"fieldValues"`
}