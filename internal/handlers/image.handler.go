package handlers

import (
	"context"
	"errors"
	"my-go-app/internal/dto"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"
	"strings"

	"my-go-app/pkg/utils"
	"time"

	"github.com/gofiber/fiber/v3"
)

type ImageHandler struct {
	imageService services.ImageServiceInterface
}

func NewImageHandler(imageService services.ImageServiceInterface) *ImageHandler {
	return &ImageHandler{
		imageService: imageService,
	}
}

func (h *ImageHandler) UploadImage(c fiber.Ctx) error {
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

	if err := services.ValidateImageContent(file); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid image file", err)
	}

	var imageNamePtr *string
	if imageName := c.FormValue("imageName"); imageName != "" {
		imageNamePtr = &imageName
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	uploadResponse, err := h.imageService.CreateUploadedImage(ctx, userID, fileHeader.Filename, imageNamePtr, file)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to save image metadata", err)
	}

	return utils.SendJSON(c, fiber.StatusCreated, uploadResponse)
}

func (h *ImageHandler) UploadBase64Image(c fiber.Ctx) error {
	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	req := new(dto.UploadBase64ImageRequest)
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	uploadResponse, err := h.imageService.CreateBase64UploadedImage(ctx, userID, req.ImageData, req.ImageName)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to save image metadata", err)
	}

	return utils.SendJSON(c, fiber.StatusCreated, uploadResponse)
}

func (h *ImageHandler) GetUserImages(c fiber.Ctx) error {
	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	ctx := c.RequestCtx()

	images, err := h.imageService.GetImagesByUserID(ctx, userID)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to retrieve images")
	}

	return utils.SendJSON(c, fiber.StatusOK, images)
}

func (h *ImageHandler) GetImageByID(c fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "imageid")
	if err != nil {
		return err
	}
	imageID := ids[0]

	ctx := c.RequestCtx()

	image, err := h.imageService.GetImageByID(ctx, imageID, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrImageNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Image not found")
		}
		return utils.HandleRepoError(c, err, "Image not found", "Failed to retrieve image")
	}

	return utils.SendJSON(c, fiber.StatusOK, image)
}

// RedirectPublicImage issues a short-lived 302 redirect to a presigned storage URL.
func (h *ImageHandler) RedirectPublicImage(c fiber.Ctx) error {
	imageID := c.Params("imageid")
	if imageID == "" || len(imageID) > 128 || strings.ContainsAny(imageID, "/\\") {
		return fiber.NewError(fiber.StatusBadRequest, "invalid imageid")
	}

	targetURL, err := h.imageService.GetPublicImageURL(c.RequestCtx(), imageID)
	if err != nil {
		if errors.Is(err, repositories.ErrImageNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Image not found")
		}
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to resolve image URL", err)
	}

	c.Set("Cache-Control", "private, max-age=240")
	return c.Redirect().Status(fiber.StatusFound).To(targetURL)
}

func (h *ImageHandler) DeleteImage(c fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "imageid")
	if err != nil {
		return err
	}
	imageID := ids[0]

	ctx := c.RequestCtx()

	if err := h.imageService.DeleteImage(ctx, imageID, userID); err != nil {
		if errors.Is(err, repositories.ErrImageNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Image not found")
		}
		return utils.HandleRepoError(c, err, "", "Failed to delete image")
	}

	return utils.SendNoContent(c)
}

// fiber:context-methods migrated
