package handlers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"
)

type imageHandlerStub struct {
	services.ImageServiceInterface
	err error

	gotUser, gotFile, gotID string
	gotName                 *string
	gotContent              string
	gotBase64               string
	calls                   int
}

func (s *imageHandlerStub) CreateUploadedImage(_ context.Context, userID, fileName string, imageName *string, file multipart.File) (*models.ImageUploadResponse, error) {
	s.calls++
	s.gotUser, s.gotFile, s.gotName = userID, fileName, imageName
	raw, _ := io.ReadAll(file)
	s.gotContent = string(raw)
	if s.err != nil {
		return nil, s.err
	}
	return &models.ImageUploadResponse{ImageId: "img-1", ImageLink: "https://cdn.example/x.png", ImageName: imageName, CreatedAt: time.Unix(0, 0).UTC()}, nil
}

func (s *imageHandlerStub) CreateBase64UploadedImage(_ context.Context, userID, data string, name *string) (*models.ImageUploadResponse, error) {
	s.calls++
	s.gotUser, s.gotBase64, s.gotName = userID, data, name
	if s.err != nil {
		return nil, s.err
	}
	return &models.ImageUploadResponse{ImageId: "img-2", ImageLink: "https://cdn.example/y.png", ImageName: name}, nil
}

func (s *imageHandlerStub) GetImagesByUserID(_ context.Context, userID string) ([]models.Image, error) {
	s.gotUser = userID
	if s.err != nil {
		return nil, s.err
	}
	return []models.Image{{ImageId: "img-1", UserId: userID, ImageLink: "https://cdn.example/x.png"}}, nil
}

func (s *imageHandlerStub) GetImageByID(_ context.Context, imageID, userID string) (*models.Image, error) {
	s.gotID, s.gotUser = imageID, userID
	if s.err != nil {
		return nil, s.err
	}
	return &models.Image{ImageId: imageID, UserId: userID}, nil
}

func (s *imageHandlerStub) DeleteImage(_ context.Context, imageID, userID string) error {
	s.gotID, s.gotUser = imageID, userID
	return s.err
}

func (s *imageHandlerStub) GetPublicImageURL(_ context.Context, imageID string) (string, error) {
	s.gotID = imageID
	if s.err != nil {
		return "", s.err
	}
	return "https://t3.storageapi.dev/signed/" + imageID, nil
}

func imageHandlerApp(userID string, svc *imageHandlerStub) *fiber.App {
	h := NewImageHandler(svc)
	return newKitApp(userID, func(app *fiber.App) {
		app.Get("/public/images/:imageid", h.RedirectPublicImage)
		app.Post("/images", h.UploadImage)
		app.Post("/images/base64", h.UploadBase64Image)
		app.Get("/images", h.GetUserImages)
		app.Get("/images/:imageid", h.GetImageByID)
		app.Delete("/images/:imageid", h.DeleteImage)
		// Registered without the parameter to exercise the missing-id branch.
		app.Get("/noid", h.GetImageByID)
		app.Delete("/noid", h.DeleteImage)
		app.Get("/public/noid", h.RedirectPublicImage)
	})
}

func imageHandlerUpload(t *testing.T, app *fiber.App, filename, content, name string) kitReply {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile("image", filename)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = fw.Write([]byte(content))
	if name != "" {
		_ = w.WriteField("imageName", name)
	}
	_ = w.Close()
	req := httptest.NewRequest("POST", "/images", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	return kitReply{status: res.StatusCode, body: string(raw)}
}

const (
	testPNGBytes  = "\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR"
	testJPEGBytes = "\xff\xd8\xff\xe0\x00\x10JFIF"
)

func TestUploadImagePassesFileAndNameToService(t *testing.T) {
	svc := &imageHandlerStub{}
	r := imageHandlerUpload(t, imageHandlerApp("u1", svc), "logo.png", testPNGBytes, "My Logo")
	kitExpect(t, r, 201, "img-1", "https://cdn.example/x.png", "My Logo")
	if svc.gotUser != "u1" || svc.gotFile != "logo.png" || svc.gotContent != testPNGBytes {
		t.Fatalf("service got user=%q file=%q content=%q", svc.gotUser, svc.gotFile, svc.gotContent)
	}
	if svc.gotName == nil || *svc.gotName != "My Logo" {
		t.Fatalf("image name = %v", svc.gotName)
	}
}

func TestUploadImageWithoutNameSendsNil(t *testing.T) {
	svc := &imageHandlerStub{}
	kitExpect(t, imageHandlerUpload(t, imageHandlerApp("u1", svc), "a.jpg", testJPEGBytes, ""), 201)
	if svc.gotName != nil {
		t.Fatalf("image name = %q, want nil", *svc.gotName)
	}
}

func TestUploadImageRejections(t *testing.T) {
	t.Run("anonymous", func(t *testing.T) {
		svc := &imageHandlerStub{}
		kitExpect(t, imageHandlerUpload(t, imageHandlerApp("", svc), "a.png", testPNGBytes, ""), 401)
		if svc.calls != 0 {
			t.Fatal("service called without a user")
		}
	})
	t.Run("no file", func(t *testing.T) {
		svc := &imageHandlerStub{}
		r := kitSend(t, imageHandlerApp("u1", svc), "POST", "/images", `{"a":1}`)
		kitExpect(t, r, 400, "No file uploaded")
		if svc.calls != 0 {
			t.Fatal("service called without a file")
		}
	})
	t.Run("bad extension", func(t *testing.T) {
		svc := &imageHandlerStub{}
		kitExpect(t, imageHandlerUpload(t, imageHandlerApp("u1", svc), "evil.exe", "x", ""), 400, "Invalid image file", "invalid file type")
		if svc.calls != 0 {
			t.Fatal("service called for an invalid file")
		}
	})
	t.Run("svg extension rejected", func(t *testing.T) {
		svc := &imageHandlerStub{}
		kitExpect(t, imageHandlerUpload(t, imageHandlerApp("u1", svc), "xss.svg", "<svg onload=alert(1)>", ""), 400, "Invalid image file", "invalid file type")
		if svc.calls != 0 {
			t.Fatal("service called for an svg file")
		}
	})
	t.Run("magic bytes mismatch", func(t *testing.T) {
		svc := &imageHandlerStub{}
		kitExpect(t, imageHandlerUpload(t, imageHandlerApp("u1", svc), "fake.png", "<script>alert(1)</script>", ""), 400, "Invalid image file", "invalid image content type")
		if svc.calls != 0 {
			t.Fatal("service called for disguised content")
		}
	})
	t.Run("service failure", func(t *testing.T) {
		svc := &imageHandlerStub{err: errInternal}
		kitExpect(t, imageHandlerUpload(t, imageHandlerApp("u1", svc), "a.png", testPNGBytes, ""), 500, "Failed to save image metadata")
	})
}

func TestUploadBase64Image(t *testing.T) {
	cases := []struct {
		name   string
		user   string
		body   string
		err    error
		want   int
		expect []string
	}{
		{"ok", "u1", `{"imageData":"QUJD","imageName":"pic"}`, nil, 201, []string{"img-2", "pic"}},
		{"anonymous", "", `{"imageData":"QUJD"}`, nil, 401, nil},
		{"bad json", "u1", `{"imageData":`, nil, 0, nil},
		{"missing data", "u1", `{"imageName":"pic"}`, nil, 0, nil},
		{"empty name", "u1", `{"imageData":"QUJD","imageName":""}`, nil, 0, nil},
		{"service failure", "u1", `{"imageData":"QUJD"}`, errInternal, 500, []string{"Failed to save image metadata"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := &imageHandlerStub{err: tc.err}
			kitExpect(t, kitSend(t, imageHandlerApp(tc.user, svc), "POST", "/images/base64", tc.body), tc.want, tc.expect...)
			if tc.want >= 400 && tc.err == nil && svc.calls != 0 {
				t.Fatal("service called for a rejected request")
			}
		})
	}
	svc := &imageHandlerStub{}
	kitExpect(t, kitSend(t, imageHandlerApp("u9", svc), "POST", "/images/base64", `{"imageData":"QUJD","imageName":"pic"}`), 201)
	if svc.gotUser != "u9" || svc.gotBase64 != "QUJD" || svc.gotName == nil || *svc.gotName != "pic" {
		t.Fatalf("service got user=%q data=%q name=%v", svc.gotUser, svc.gotBase64, svc.gotName)
	}
}

func TestImageReadAndDeleteHandlers(t *testing.T) {
	notFoundText := fmt.Errorf("lookup: %w", repositories.ErrImageNotFound)
	cases := []struct {
		name         string
		user         string
		method, path string
		err          error
		want         int
		expect       string
	}{
		{"list ok", "u1", "GET", "/images", nil, 200, "img-1"},
		{"list anonymous", "", "GET", "/images", nil, 401, ""},
		{"list failure", "u1", "GET", "/images", errInternal, 500, "Failed to retrieve images"},
		{"get ok", "u1", "GET", "/images/img-7", nil, 200, "img-7"},
		{"get anonymous", "", "GET", "/images/img-7", nil, 401, ""},
		{"get missing id", "u1", "GET", "/noid", nil, 400, "imageid is required"},
		{"get not found", "u1", "GET", "/images/img-7", notFoundText, 404, "Image not found"},
		{"get failure", "u1", "GET", "/images/img-7", errInternal, 500, "Failed to retrieve image"},
		{"delete ok", "u1", "DELETE", "/images/img-7", nil, 204, ""},
		{"delete anonymous", "", "DELETE", "/images/img-7", nil, 401, ""},
		{"delete missing id", "u1", "DELETE", "/noid", nil, 400, "imageid is required"},
		{"delete not found", "u1", "DELETE", "/images/img-7", repositories.ErrImageNotFound, 404, "Image not found"},
		{"delete failure", "u1", "DELETE", "/images/img-7", errInternal, 500, "Failed to delete image"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := &imageHandlerStub{err: tc.err}
			r := kitSend(t, imageHandlerApp(tc.user, svc), tc.method, tc.path, "")
			if tc.expect == "" {
				kitExpect(t, r, tc.want)
			} else {
				kitExpect(t, r, tc.want, tc.expect)
			}
			if tc.want < 400 && tc.path != "/images" && (svc.gotID != "img-7" || svc.gotUser != "u1") {
				t.Fatalf("service got id=%q user=%q", svc.gotID, svc.gotUser)
			}
		})
	}
}

func TestRedirectPublicImage(t *testing.T) {
	t.Run("redirects 302 with private cache", func(t *testing.T) {
		svc := &imageHandlerStub{}
		app := imageHandlerApp("", svc)
		req := httptest.NewRequest("GET", "/public/images/img-9", nil)
		res, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		if res.StatusCode != fiber.StatusFound {
			t.Fatalf("status = %d, want 302", res.StatusCode)
		}
		if loc := res.Header.Get("Location"); loc != "https://t3.storageapi.dev/signed/img-9" {
			t.Fatalf("Location = %q", loc)
		}
		if cc := res.Header.Get("Cache-Control"); cc != "private, max-age=240" {
			t.Fatalf("Cache-Control = %q", cc)
		}
	})

	t.Run("not found", func(t *testing.T) {
		svc := &imageHandlerStub{err: repositories.ErrImageNotFound}
		kitExpect(t, kitSend(t, imageHandlerApp("", svc), "GET", "/public/images/missing", ""), 404, "Image not found")
	})

	t.Run("missing param", func(t *testing.T) {
		svc := &imageHandlerStub{}
		kitExpect(t, kitSend(t, imageHandlerApp("", svc), "GET", "/public/noid", ""), 400, "invalid imageid")
	})

	t.Run("storage error", func(t *testing.T) {
		svc := &imageHandlerStub{err: errInternal}
		kitExpect(t, kitSend(t, imageHandlerApp("", svc), "GET", "/public/images/img-9", ""), 500, "Failed to resolve image URL")
	})
}
