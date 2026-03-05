package dto

// UploadBase64ImageRequest contains the fields required to upload a base64 encoded image.
type UploadBase64ImageRequest struct {
	ImageData string  `json:"imageData" validate:"required"`
	ImageName *string `json:"imageName"`
}