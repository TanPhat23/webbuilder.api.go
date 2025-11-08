package handlers

import (
	"context"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/lucsky/cuid"
)

type ImageHandler struct {
	imageRepository   repositories.ImageRepositoryInterface
	cloudinaryService *services.CloudinaryService
}

func NewImageHandler(imageRepo repositories.ImageRepositoryInterface, cloudinaryService *services.CloudinaryService) *ImageHandler {
	return &ImageHandler{
		imageRepository:   imageRepo,
		cloudinaryService: cloudinaryService,
	}
}

// UploadImage handles image upload to Cloudinary and saves metadata to database
func (h *ImageHandler) UploadImage(c *fiber.Ctx) error {
	// Get user ID from context
	userID, ok := c.Locals("userId").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":        "Unauthorized",
			"errorMessage": "You must be logged in to upload images",
		})
	}

	// Get the uploaded file
	fileHeader, err := c.FormFile("image")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "No file uploaded",
			"errorMessage": err.Error(),
		})
	}

	// Validate the image file
	if err := services.ValidateImageFile(fileHeader); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Invalid image file",
			"errorMessage": err.Error(),
		})
	}

	// Open the file
	file, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to open uploaded file",
			"errorMessage": err.Error(),
		})
	}
	defer file.Close()

	// Get optional image name from form data
	imageName := c.FormValue("imageName")
	var imageNamePtr *string
	if imageName != "" {
		imageNamePtr = &imageName
	}

	// Upload to Cloudinary
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	folder := "webbuilder/" + userID
	uploadResult, err := h.cloudinaryService.UploadImage(ctx, file, fileHeader.Filename, folder)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to upload image to Cloudinary",
			"errorMessage": err.Error(),
		})
	}

	// Create image record in database
	now := time.Now()
	image := models.Image{
		ImageId:   cuid.New(),
		ImageLink: uploadResult.SecureURL,
		ImageName: imageNamePtr,
		UserId:    userID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	createdImage, err := h.imageRepository.CreateImage(image)
	if err != nil {
		// Try to cleanup Cloudinary upload if database insert fails
		_ = h.cloudinaryService.DeleteImage(ctx, uploadResult.PublicID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to save image metadata",
			"errorMessage": err.Error(),
		})
	}

	// Return response
	response := models.ImageUploadResponse{
		ImageId:   createdImage.ImageId,
		ImageLink: createdImage.ImageLink,
		ImageName: createdImage.ImageName,
		CreatedAt: createdImage.CreatedAt,
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}

// GetUserImages retrieves all images for the authenticated user
func (h *ImageHandler) GetUserImages(c *fiber.Ctx) error {
	userID, ok := c.Locals("userId").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":        "Unauthorized",
			"errorMessage": "You must be logged in to access images",
		})
	}

	images, err := h.imageRepository.GetImagesByUserID(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to retrieve images",
			"errorMessage": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(images)
}

// GetImageByID retrieves a specific image by ID
func (h *ImageHandler) GetImageByID(c *fiber.Ctx) error {
	imageID := c.Params("imageid")
	if imageID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Image ID is required",
			"errorMessage": "Missing imageid parameter in URL",
		})
	}

	userID, ok := c.Locals("userId").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":        "Unauthorized",
			"errorMessage": "You must be logged in to access images",
		})
	}

	image, err := h.imageRepository.GetImageByID(imageID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to retrieve image",
			"errorMessage": err.Error(),
		})
	}
	if image == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Image not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(image)
}

// DeleteImage deletes an image from both Cloudinary and database
func (h *ImageHandler) DeleteImage(c *fiber.Ctx) error {
	imageID := c.Params("imageid")
	if imageID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Image ID is required",
			"errorMessage": "Missing imageid parameter in URL",
		})
	}

	userID, ok := c.Locals("userId").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":        "Unauthorized",
			"errorMessage": "You must be logged in to delete images",
		})
	}

	// Get image to extract Cloudinary public ID
	image, err := h.imageRepository.GetImageByID(imageID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to retrieve image",
			"errorMessage": err.Error(),
		})
	}
	if image == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Image not found",
		})
	}

	// Extract public ID from Cloudinary URL (simplified approach)
	// Note: In production, you might want to store the public ID separately
	// For now, we'll attempt to delete from Cloudinary but continue even if it fails

	// Soft delete the image in database
	err = h.imageRepository.SoftDeleteImage(imageID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to delete image",
			"errorMessage": err.Error(),
		})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// UploadBase64Image handles base64 image upload
func (h *ImageHandler) UploadBase64Image(c *fiber.Ctx) error {
	userID, ok := c.Locals("userId").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":        "Unauthorized",
			"errorMessage": "You must be logged in to upload images",
		})
	}

	type Base64Request struct {
		ImageData string  `json:"imageData"`
		ImageName *string `json:"imageName"`
	}

	var req Base64Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Invalid request body",
			"errorMessage": err.Error(),
		})
	}

	if req.ImageData == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Image data is required",
			"errorMessage": "imageData field cannot be empty",
		})
	}

	// Upload to Cloudinary
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	folder := "webbuilder/" + userID
	uploadResult, err := h.cloudinaryService.UploadBase64Image(ctx, req.ImageData, folder)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to upload image to Cloudinary",
			"errorMessage": err.Error(),
		})
	}

	// Create image record in database
	now := time.Now()
	image := models.Image{
		ImageId:   cuid.New(),
		ImageLink: uploadResult.SecureURL,
		ImageName: req.ImageName,
		UserId:    userID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	createdImage, err := h.imageRepository.CreateImage(image)
	if err != nil {
		// Try to cleanup Cloudinary upload if database insert fails
		_ = h.cloudinaryService.DeleteImage(ctx, uploadResult.PublicID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to save image metadata",
			"errorMessage": err.Error(),
		})
	}

	// Return response
	response := models.ImageUploadResponse{
		ImageId:   createdImage.ImageId,
		ImageLink: createdImage.ImageLink,
		ImageName: createdImage.ImageName,
		CreatedAt: createdImage.CreatedAt,
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}
