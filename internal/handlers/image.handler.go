package handlers

import (
	"context"
	"errors"
	"my-go-app/internal/dto"
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

	var imageNamePtr *string
	if imageName := c.FormValue("imageName"); imageName != "" {
		imageNamePtr = &imageName
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	uploadResult, err := h.cloudinaryService.UploadImage(ctx, file, fileHeader.Filename, "webbuilder/"+userID)
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

	created, err := h.imageRepository.CreateImage(ctx, image)
	if err != nil {
		_ = h.cloudinaryService.DeleteImage(ctx, uploadResult.PublicID)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to save image metadata", err)
	}

	return utils.SendJSON(c, fiber.StatusCreated, models.ImageUploadResponse{
		ImageId:   created.ImageId,
		ImageLink: created.ImageLink,
		ImageName: created.ImageName,
		CreatedAt: created.CreatedAt,
	})
}

func (h *ImageHandler) UploadBase64Image(c *fiber.Ctx) error {
	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	var req dto.UploadBase64ImageRequest
	if err := utils.ValidateAndParseBody(c, &req); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	uploadResult, err := h.cloudinaryService.UploadBase64Image(ctx, req.ImageData, "webbuilder/"+userID)
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

	created, err := h.imageRepository.CreateImage(ctx, image)
	if err != nil {
		_ = h.cloudinaryService.DeleteImage(ctx, uploadResult.PublicID)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to save image metadata", err)
	}

	return utils.SendJSON(c, fiber.StatusCreated, models.ImageUploadResponse{
		ImageId:   created.ImageId,
		ImageLink: created.ImageLink,
		ImageName: created.ImageName,
		CreatedAt: created.CreatedAt,
	})
}

func (h *ImageHandler) GetUserImages(c *fiber.Ctx) error {
	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	ctx := c.Context()

	images, err := h.imageRepository.GetImagesByUserID(ctx, userID)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to retrieve images")
	}

	return utils.SendJSON(c, fiber.StatusOK, images)
}

func (h *ImageHandler) GetImageByID(c *fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "imageid")
	if err != nil {
		return err
	}
	imageID := ids[0]

	ctx := c.Context()

	image, err := h.imageRepository.GetImageByID(ctx, imageID, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrImageNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Image not found")
		}
		return utils.HandleRepoError(c, err, "Image not found", "Failed to retrieve image")
	}

	return utils.SendJSON(c, fiber.StatusOK, image)
}

func (h *ImageHandler) DeleteImage(c *fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "imageid")
	if err != nil {
		return err
	}
	imageID := ids[0]

	ctx := c.Context()

	// Verify the image exists and belongs to the user before soft-deleting.
	if _, err := h.imageRepository.GetImageByID(ctx, imageID, userID); err != nil {
		if errors.Is(err, repositories.ErrImageNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Image not found")
		}
		return utils.HandleRepoError(c, err, "Image not found", "Failed to retrieve image")
	}

	if err := h.imageRepository.SoftDeleteImage(ctx, imageID, userID); err != nil {
		if errors.Is(err, repositories.ErrImageNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Image not found")
		}
		return utils.HandleRepoError(c, err, "", "Failed to delete image")
	}

	return utils.SendNoContent(c)
}