package dto

type CreateProjectRequest struct {
	Styles      any     `json:"styles"      validate:"omitempty"`
	Header      any     `json:"header"      validate:"omitempty"`
	Name        string  `json:"name"        validate:"required,min=1,max=255"`
	Description *string `json:"description" validate:"omitempty,min=1,max=1000"`
	Subdomain   *string `json:"subdomain"   validate:"omitempty,alphanum,min=3,max=63"`
	Published   bool    `json:"published"`
}

type UpdateProjectRequest struct {
	Styles      any     `json:"styles"      validate:"omitempty"`
	Header      any     `json:"header"      validate:"omitempty"`
	Name        *string `json:"name"        validate:"omitempty,min=1,max=255"`
	Description *string `json:"description" validate:"omitempty,min=1,max=1000"`
	Subdomain   *string `json:"subdomain"   validate:"omitempty,alphanum,min=3,max=63"`
	Published   *bool   `json:"published"`
}
