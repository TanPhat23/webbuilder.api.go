package dto

// UpdateProjectRequest contains the fields that can be patched on a project.
type UpdateProjectRequest struct {
	Name        *string `json:"name"        validate:"omitempty,min=1,max=255"`
	Description *string `json:"description" validate:"omitempty,max=1000"`
	Published   *bool   `json:"published"`
	Subdomain   *string `json:"subdomain"   validate:"omitempty,alphanum,min=3,max=63"`
	Styles      any     `json:"styles"`
	Header      any     `json:"header"`
}
type CreateProjectRequest struct {
	Name        string `json:"name"        validate:"required,min=1,max=255"`
	Description *string `json:"description" validate:"omitempty,max=1000"`
	Published   bool   `json:"published"`
	Subdomain   *string `json:"subdomain"   validate:"omitempty,alphanum,min=3,max=63"`
	Styles      any     `json:"styles"`
	Header      any     `json:"header"`
}