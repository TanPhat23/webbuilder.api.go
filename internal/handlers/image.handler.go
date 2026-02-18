package handlers

import (
	"context"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"
	"my-go-app/pkg/utils"
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

// UploadImage handles image upload to Cloudinary and saves metadata to database.
func (h *ImageHandler) UploadImage(c *fiber.Ctx) error {
	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	fileHeader, err := c.FormFile("image")
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "No file uploaded", err)
	}

	if err := services.ValidateImageFile(fileHeader); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid image file", err)
	}

	file, err := fileHeader.Open()
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to open uploaded file", err)
	}
	defer file.Close()

	imageName := c.FormValue("imageName")
	var imageNamePtr *string
	if imageName != "" {
		imageNamePtr = &imageName
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	folder := "webbuilder/" + userID
	uploadResult, err := h.cloudinaryService.UploadImage(ctx, file, fileHeader.Filename, folder)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to upload image to Cloudinary", err)
	}

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
		_ = h.cloudinaryService.DeleteImage(ctx, uploadResult.PublicID)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to save image metadata", err)
	}

	return utils.SendJSON(c, fiber.StatusCreated, models.ImageUploadResponse{
		ImageId:   createdImage.ImageId,
		ImageLink: createdImage.ImageLink,
		ImageName: createdImage.ImageName,
		CreatedAt: createdImage.CreatedAt,
	})
}

// GetUserImages retrieves all images for the authenticated user.
func (h *ImageHandler) GetUserImages(c *fiber.Ctx) error {
	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	images, err := h.imageRepository.GetImagesByUserID(userID)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve images", err)
	}

	return utils.SendJSON(c, fiber.StatusOK, images)
}

// GetImageByID retrieves a specific image by ID.
func (h *ImageHandler) GetImageByID(c *fiber.Ctx) error {
	imageID, err := utils.ValidateRequiredParam(c, "imageid")
	if err != nil {
		return err
	}

	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	image, err := h.imageRepository.GetImageByID(imageID, userID)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve image", err)
	}
	if image == nil {
		return utils.SendError(c, fiber.StatusNotFound, "Image not found", nil)
	}

	return utils.SendJSON(c, fiber.StatusOK, image)
}

// DeleteImage deletes an image from both Cloudinary and database.
func (h *ImageHandler) DeleteImage(c *fiber.Ctx) error {
	imageID, err := utils.ValidateRequiredParam(c, "imageid")
	if err != nil {
		return err
	}

	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	image, err := h.imageRepository.GetImageByID(imageID, userID)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve image", err)
	}
	if image == nil {
		return utils.SendError(c, fiber.StatusNotFound, "Image not found", nil)
	}

	if err := h.imageRepository.SoftDeleteImage(imageID, userID); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to delete image", err)
	}

	return utils.SendNoContent(c)
}

// UploadBase64Image handles base64 image upload.
func (h *ImageHandler) UploadBase64Image(c *fiber.Ctx) error {
	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	var req struct {
		ImageData string  `json:"imageData" validate:"required"`
		ImageName *string `json:"imageName"`
	}

	if err := utils.ValidateAndParseBody(c, &req); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	folder := "webbuilder/" + userID
	uploadResult, err := h.cloudinaryService.UploadBase64Image(ctx, req.ImageData, folder)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to upload image to Cloudinary", err)
	}

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
		_ = h.cloudinaryService.DeleteImage(ctx, uploadResult.PublicID)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to save image metadata", err)
	}

	return utils.SendJSON(c, fiber.StatusCreated, models.ImageUploadResponse{
		ImageId:   createdImage.ImageId,
		ImageLink: createdImage.ImageLink,
		ImageName: createdImage.ImageName,
		CreatedAt: createdImage.CreatedAt,
	})
}
