package dto

type CreateContentTypeRequest struct {
	ProjectID        *string `json:"projectId"        validate:"omitempty,min=1"`
	Name             string  `json:"name"             validate:"required,min=1,max=255"`
	Slug             *string `json:"slug"             validate:"omitempty,min=1,max=255"`
	Description      *string `json:"description"      validate:"omitempty,min=1,max=1000"`
	PublicRead       *bool   `json:"publicRead"`
	PublicCreate     *bool   `json:"publicCreate"`
	DefaultPublished *bool   `json:"defaultPublished"`
	RateLimitPerMin  *int    `json:"rateLimitPerMin"  validate:"omitempty,min=1,max=1000"`
}

type UpdateContentTypeRequest struct {
	Name             *string `json:"name"             validate:"omitempty,min=1,max=255"`
	Slug             *string `json:"slug"             validate:"omitempty,min=1,max=255"`
	Description      *string `json:"description"      validate:"omitempty,min=1,max=1000"`
	PublicRead       *bool   `json:"publicRead"`
	PublicCreate     *bool   `json:"publicCreate"`
	DefaultPublished *bool   `json:"defaultPublished"`
	RateLimitPerMin  *int    `json:"rateLimitPerMin"  validate:"omitempty,min=1,max=1000"`
}
