package dto

// CreateContentTypeRequest contains the required fields to create a new content type.
type CreateContentTypeRequest struct {
	Name        string  `json:"name"        validate:"required,min=1,max=255"`
	Description *string `json:"description" validate:"omitempty,max=1000"`
}

// UpdateContentTypeRequest contains the patchable fields on a content type.
type UpdateContentTypeRequest struct {
	Name        *string `json:"name"        validate:"omitempty,min=1,max=255"`
	Description *string `json:"description" validate:"omitempty,max=1000"`
}