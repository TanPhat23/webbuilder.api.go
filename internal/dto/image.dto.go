package dto

type UploadBase64ImageRequest struct {
	ImageData string  `json:"imageData" validate:"required,min=1"`
	ImageName *string `json:"imageName" validate:"omitempty,min=1,max=255"`
}

type CreateCategoryRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}

type CreateTagRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}
