package dto

import "my-go-app/internal/models"

type CreateContentItemRequest struct {
	FieldValues []models.ContentFieldValue `json:"fieldValues" validate:"omitempty"`
	Slug        string                     `json:"slug"        validate:"required,min=1,max=255"`
	Title       string                     `json:"title"       validate:"required,min=1,max=255"`
	Published   *bool                      `json:"published"`
}

type UpdateContentItemRequest struct {
	FieldValues []models.ContentFieldValue `json:"fieldValues" validate:"omitempty"`
	Slug        *string                    `json:"slug"        validate:"omitempty,min=1,max=255"`
	Title       *string                    `json:"title"       validate:"omitempty,min=1,max=255"`
	Published   *bool                      `json:"published"`
}
