package dto

type CreateCustomElementTypeRequest struct {
	Name        string  `json:"name"        validate:"required,min=1,max=255"`
	Description *string `json:"description" validate:"omitempty,min=1,max=1000"`
	Category    *string `json:"category"    validate:"omitempty,min=1,max=100"`
	Icon        *string `json:"icon"        validate:"omitempty,min=1,max=255"`
}

type UpdateCustomElementTypeRequest struct {
	Name        *string `json:"name"        validate:"omitempty,min=1,max=255"`
	Description *string `json:"description" validate:"omitempty,min=1,max=1000"`
	Category    *string `json:"category"    validate:"omitempty,min=1,max=100"`
	Icon        *string `json:"icon"        validate:"omitempty,min=1,max=255"`
}
