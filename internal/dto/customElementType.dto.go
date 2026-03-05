package dto

// CreateCustomElementTypeRequest contains the required fields to create a new custom element type.
type CreateCustomElementTypeRequest struct {
	Name        string  `json:"name"        validate:"required,min=1,max=255"`
	Description *string `json:"description" validate:"omitempty,max=1000"`
	Category    *string `json:"category"    validate:"omitempty,max=100"`
	Icon        *string `json:"icon"        validate:"omitempty,max=255"`
}

// UpdateCustomElementTypeRequest contains the patchable fields on a custom element type.
type UpdateCustomElementTypeRequest struct {
	Name        *string `json:"name"        validate:"omitempty,min=1,max=255"`
	Description *string `json:"description" validate:"omitempty,max=1000"`
	Category    *string `json:"category"    validate:"omitempty,max=100"`
	Icon        *string `json:"icon"        validate:"omitempty,max=255"`
}