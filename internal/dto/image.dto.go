package dto

type UploadBase64ImageRequest struct {
	ImageData string  `json:"imageData" validate:"required,min=1"`
	ImageName *string `json:"imageName" validate:"omitempty,min=1,max=255"`
}
