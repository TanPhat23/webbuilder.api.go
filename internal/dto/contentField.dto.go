package dto

// CreateContentFieldRequest contains the required fields to create a new content field.
type CreateContentFieldRequest struct {
	Name     string `json:"name"     validate:"required,min=1,max=255"`
	Type     string `json:"type"     validate:"required"`
	Required bool   `json:"required"`
}

// UpdateContentFieldRequest contains the patchable fields on a content field.
type UpdateContentFieldRequest struct {
	Name     *string `json:"name"     validate:"omitempty,min=1,max=255"`
	Type     *string `json:"type"     validate:"omitempty"`
	Required *bool   `json:"required"`
}