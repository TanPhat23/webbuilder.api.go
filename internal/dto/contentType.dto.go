package dto

type CreateContentTypeRequest struct {
	Name        string  `json:"name"        validate:"required,min=1,max=255"`
	Description *string `json:"description" validate:"omitempty,min=1,max=1000"`
}

type UpdateContentTypeRequest struct {
	Name        *string `json:"name"        validate:"omitempty,min=1,max=255"`
	Description *string `json:"description" validate:"omitempty,min=1,max=1000"`
}
