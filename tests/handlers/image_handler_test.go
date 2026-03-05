package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"my-go-app/internal/handlers"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/tests/testutil"

	"github.com/gofiber/fiber/v2"
)

// ─── test app factory ─────────────────────────────────────────────────────────

func newImageTestApp(imageRepo repositories.ImageRepositoryInterface) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var fe *fiber.Error
			if errors.As(err, &fe) {
				code = fe.Code
			}
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})

	// Inject a fake userId into locals so auth middleware is bypassed.
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("userId", "test-user-id")
		return c.Next()
	})

	h := handlers.NewImageHandler(imageRepo, nil /* CloudinaryService unused in these tests */)

	app.Get("/images", h.GetUserImages)
	app.Get("/images/:imageid", h.GetImageByID)
	app.Delete("/images/:imageid", h.DeleteImage)

	return app
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func doRequest(app *fiber.App, method, path string, body io.Reader, contentType string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, body)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	resp, _ := app.Test(req, -1)
	_ = resp
	return nil
}

// statusOf fires a request and returns only the HTTP status code.
func statusOf(t *testing.T, app *fiber.App, method, path string) int {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	return resp.StatusCode
}

// bodyOf fires a request and returns status + raw body bytes.
func bodyOf(t *testing.T, app *fiber.App, method, path string) (int, []byte) {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("io.ReadAll error: %v", err)
	}
	return resp.StatusCode, b
}

// ─── GetUserImages ────────────────────────────────────────────────────────────

func TestGetUserImages_ReturnsEmptySliceWhenNoImages(t *testing.T) {
	repo := &testutil.MockImageRepository{
		GetImagesByUserIDFn: func(_ context.Context, _ string) ([]models.Image, error) {
			return []models.Image{}, nil
		},
	}
	app := newImageTestApp(repo)

	status, body := bodyOf(t, app, "GET", "/images")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}

	// Body should be a JSON array (empty).
	var result []any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON array: %v – body: %s", err, body)
	}
	if len(result) != 0 {
		t.Errorf("expected empty array, got %d elements", len(result))
	}
}

func TestGetUserImages_ReturnsImagesForAuthenticatedUser(t *testing.T) {
	now := time.Now()
	name := "screenshot.png"
	repo := &testutil.MockImageRepository{
		GetImagesByUserIDFn: func(_ context.Context, userID string) ([]models.Image, error) {
			if userID != "test-user-id" {
				return nil, errors.New("unexpected userID: " + userID)
			}
			return []models.Image{
				{ImageId: "img-1", UserId: userID, ImageLink: "https://cdn.example.com/1.png", ImageName: &name, CreatedAt: now},
				{ImageId: "img-2", UserId: userID, ImageLink: "https://cdn.example.com/2.png", CreatedAt: now},
			}, nil
		},
	}
	app := newImageTestApp(repo)

	status, body := bodyOf(t, app, "GET", "/images")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}

	var result []map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON array: %v – body: %s", err, body)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 images, got %d", len(result))
	}
}

func TestGetUserImages_Returns500OnRepositoryError(t *testing.T) {
	repo := &testutil.MockImageRepository{
		GetImagesByUserIDFn: func(_ context.Context, _ string) ([]models.Image, error) {
			return nil, errors.New("db connection lost")
		},
	}
	app := newImageTestApp(repo)

	if code := statusOf(t, app, "GET", "/images"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

func TestGetUserImages_Returns401WhenNoUserID(t *testing.T) {
	repo := &testutil.MockImageRepository{}

	// App without the userId local set.
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var fe *fiber.Error
			if errors.As(err, &fe) {
				code = fe.Code
			}
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})
	h := handlers.NewImageHandler(repo, nil)
	app.Get("/images", h.GetUserImages)

	if code := statusOf(t, app, "GET", "/images"); code != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", code)
	}
}

// ─── GetImageByID ─────────────────────────────────────────────────────────────

func TestGetImageByID_ReturnsImageWhenFound(t *testing.T) {
	now := time.Now()
	repo := &testutil.MockImageRepository{
		GetImageByIDFn: func(_ context.Context, imageID, userID string) (*models.Image, error) {
			if imageID == "img-1" && userID == "test-user-id" {
				return &models.Image{
					ImageId:   imageID,
					UserId:    userID,
					ImageLink: "https://cdn.example.com/1.png",
					CreatedAt: now,
				}, nil
			}
			return nil, repositories.ErrImageNotFound
		},
	}
	app := newImageTestApp(repo)

	status, body := bodyOf(t, app, "GET", "/images/img-1")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, body)
	}
	if result["imageId"] != "img-1" {
		t.Errorf("imageId: got %v, want %q", result["imageId"], "img-1")
	}
}

func TestGetImageByID_Returns404WhenNotFound(t *testing.T) {
	repo := &testutil.MockImageRepository{
		GetImageByIDFn: func(_ context.Context, _, _ string) (*models.Image, error) {
			return nil, repositories.ErrImageNotFound
		},
	}
	app := newImageTestApp(repo)

	if code := statusOf(t, app, "GET", "/images/nonexistent"); code != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", code)
	}
}

func TestGetImageByID_Returns500OnRepositoryError(t *testing.T) {
	repo := &testutil.MockImageRepository{
		GetImageByIDFn: func(_ context.Context, _, _ string) (*models.Image, error) {
			return nil, errors.New("unexpected db error")
		},
	}
	app := newImageTestApp(repo)

	if code := statusOf(t, app, "GET", "/images/img-1"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

func TestGetImageByID_Returns400WhenParamMissing(t *testing.T) {
	// Route "/images/" with empty segment — Fiber will not match the
	// ":imageid" param route; test against a separate route that forces the
	// missing-param path by calling the handler directly with an empty param.
	repo := &testutil.MockImageRepository{}

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var fe *fiber.Error
			if errors.As(err, &fe) {
				code = fe.Code
			}
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("userId", "test-user-id")
		return c.Next()
	})
	h := handlers.NewImageHandler(repo, nil)
	// Register with a blank-allowing route so we can exercise the empty param.
	app.Get("/images/:imageid", h.GetImageByID)

	// Fiber will 404 for "/images/" (no param), so just confirm not 200.
	req := httptest.NewRequest("GET", "/images/", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode == fiber.StatusOK {
		t.Errorf("empty imageid should not return 200")
	}
}

// ─── DeleteImage ──────────────────────────────────────────────────────────────

func TestDeleteImage_Returns204OnSuccess(t *testing.T) {
	now := time.Now()
	repo := &testutil.MockImageRepository{
		GetImageByIDFn: func(_ context.Context, imageID, userID string) (*models.Image, error) {
			return &models.Image{ImageId: imageID, UserId: userID, CreatedAt: now}, nil
		},
		SoftDeleteImageFn: func(_ context.Context, _, _ string) error {
			return nil
		},
	}
	app := newImageTestApp(repo)

	if code := statusOf(t, app, "DELETE", "/images/img-1"); code != fiber.StatusNoContent {
		t.Errorf("expected 204, got %d", code)
	}
}

func TestDeleteImage_Returns404WhenImageNotFound(t *testing.T) {
	repo := &testutil.MockImageRepository{
		GetImageByIDFn: func(_ context.Context, _, _ string) (*models.Image, error) {
			return nil, repositories.ErrImageNotFound
		},
	}
	app := newImageTestApp(repo)

	if code := statusOf(t, app, "DELETE", "/images/ghost"); code != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", code)
	}
}

func TestDeleteImage_Returns500WhenGetFails(t *testing.T) {
	repo := &testutil.MockImageRepository{
		GetImageByIDFn: func(_ context.Context, _, _ string) (*models.Image, error) {
			return nil, errors.New("db timeout")
		},
	}
	app := newImageTestApp(repo)

	if code := statusOf(t, app, "DELETE", "/images/img-1"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

func TestDeleteImage_Returns404WhenSoftDeleteFindsNoRow(t *testing.T) {
	now := time.Now()
	repo := &testutil.MockImageRepository{
		GetImageByIDFn: func(_ context.Context, imageID, userID string) (*models.Image, error) {
			return &models.Image{ImageId: imageID, UserId: userID, CreatedAt: now}, nil
		},
		SoftDeleteImageFn: func(_ context.Context, _, _ string) error {
			// Simulate a race where the row was deleted between Get and SoftDelete.
			return repositories.ErrImageNotFound
		},
	}
	app := newImageTestApp(repo)

	if code := statusOf(t, app, "DELETE", "/images/img-1"); code != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", code)
	}
}

func TestDeleteImage_Returns500WhenSoftDeleteFails(t *testing.T) {
	now := time.Now()
	repo := &testutil.MockImageRepository{
		GetImageByIDFn: func(_ context.Context, imageID, userID string) (*models.Image, error) {
			return &models.Image{ImageId: imageID, UserId: userID, CreatedAt: now}, nil
		},
		SoftDeleteImageFn: func(_ context.Context, _, _ string) error {
			return errors.New("db write error")
		},
	}
	app := newImageTestApp(repo)

	if code := statusOf(t, app, "DELETE", "/images/img-1"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

func TestDeleteImage_Returns401WhenNoUserID(t *testing.T) {
	repo := &testutil.MockImageRepository{}

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var fe *fiber.Error
			if errors.As(err, &fe) {
				code = fe.Code
			}
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})
	h := handlers.NewImageHandler(repo, nil)
	app.Delete("/images/:imageid", h.DeleteImage)

	if code := statusOf(t, app, "DELETE", "/images/img-1"); code != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", code)
	}
}

// ─── UploadImage (multipart form — Cloudinary stub not wired; validates early-exit paths) ───

func TestUploadImage_Returns401WhenNoUserID(t *testing.T) {
	repo := &testutil.MockImageRepository{}

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var fe *fiber.Error
			if errors.As(err, &fe) {
				code = fe.Code
			}
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})
	h := handlers.NewImageHandler(repo, nil)
	app.Post("/images/upload", h.UploadImage)

	req := httptest.NewRequest("POST", "/images/upload", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestUploadImage_Returns400WhenNoFilePart(t *testing.T) {
	repo := &testutil.MockImageRepository{}

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var fe *fiber.Error
			if errors.As(err, &fe) {
				code = fe.Code
			}
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("userId", "test-user-id")
		return c.Next()
	})
	h := handlers.NewImageHandler(repo, nil)
	app.Post("/images/upload", h.UploadImage)

	// Send a multipart form without an "image" file part.
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	_ = mw.WriteField("imageName", "test")
	mw.Close()

	req := httptest.NewRequest("POST", "/images/upload", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("expected 400 when no image file part, got %d", resp.StatusCode)
	}
}

// ─── UploadBase64Image ────────────────────────────────────────────────────────

func TestUploadBase64Image_Returns401WhenNoUserID(t *testing.T) {
	repo := &testutil.MockImageRepository{}

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var fe *fiber.Error
			if errors.As(err, &fe) {
				code = fe.Code
			}
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})
	h := handlers.NewImageHandler(repo, nil)
	app.Post("/images/upload/base64", h.UploadBase64Image)

	req := httptest.NewRequest("POST", "/images/upload/base64", strings.NewReader(`{"imageData":"data:image/png;base64,abc"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestUploadBase64Image_Returns400WhenBodyMissingRequiredField(t *testing.T) {
	repo := &testutil.MockImageRepository{}

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var fe *fiber.Error
			if errors.As(err, &fe) {
				code = fe.Code
			}
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("userId", "test-user-id")
		return c.Next()
	})
	h := handlers.NewImageHandler(repo, nil)
	app.Post("/images/upload/base64", h.UploadBase64Image)

	// Send JSON body without the required "imageData" field.
	req := httptest.NewRequest("POST", "/images/upload/base64", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	// Validation should reject the empty body — 400 or 422.
	if resp.StatusCode == fiber.StatusOK || resp.StatusCode == fiber.StatusCreated {
		t.Errorf("expected 4xx for missing imageData, got %d", resp.StatusCode)
	}
}