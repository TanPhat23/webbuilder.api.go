package repositories_test

import (
	"context"
	"errors"
	"testing"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/tests/testutil"
)

func TestMockImageRepository_DefaultsReturnSentinel(t *testing.T) {
	repo := &testutil.MockImageRepository{}
	_, err := repo.GetImageByID(context.Background(), "img-1", "u1")
	if !errors.Is(err, repositories.ErrImageNotFound) {
		t.Errorf("GetImageByID default: want ErrImageNotFound, got %v", err)
	}
}

func TestMockImageRepository_CreateImageDefaultReturnsSameImage(t *testing.T) {
	repo := &testutil.MockImageRepository{}
	input := models.Image{ImageId: "img-1", UserId: "u1", ImageLink: "https://example.com/img.png"}
	got, err := repo.CreateImage(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ImageId != input.ImageId {
		t.Errorf("ImageId: got %q, want %q", got.ImageId, input.ImageId)
	}
}

func TestMockImageRepository_CreateImageFnAssignsID(t *testing.T) {
	repo := &testutil.MockImageRepository{
		CreateImageFn: func(_ context.Context, img models.Image) (*models.Image, error) {
			img.ImageId = "img-generated"
			return &img, nil
		},
	}
	got, err := repo.CreateImage(context.Background(), models.Image{UserId: "u1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ImageId != "img-generated" {
		t.Errorf("ImageId: got %q, want %q", got.ImageId, "img-generated")
	}
}

func TestMockImageRepository_GetImagesByUserIDDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockImageRepository{}
	images, err := repo.GetImagesByUserID(context.Background(), "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(images) != 0 {
		t.Errorf("expected empty slice, got %d", len(images))
	}
}

func TestMockImageRepository_GetImagesByUserIDFnFilters(t *testing.T) {
	all := []models.Image{
		{ImageId: "img-1", UserId: "u1"},
		{ImageId: "img-2", UserId: "u2"},
		{ImageId: "img-3", UserId: "u1"},
	}
	repo := &testutil.MockImageRepository{
		GetImagesByUserIDFn: func(_ context.Context, userID string) ([]models.Image, error) {
			var out []models.Image
			for _, img := range all {
				if img.UserId == userID {
					out = append(out, img)
				}
			}
			return out, nil
		},
	}

	got, err := repo.GetImagesByUserID(context.Background(), "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 images for u1, got %d", len(got))
	}
	for _, img := range got {
		if img.UserId != "u1" {
			t.Errorf("unexpected UserId %q in result", img.UserId)
		}
	}
}

func TestMockImageRepository_GetImageByIDFnReturnsImage(t *testing.T) {
	want := &models.Image{ImageId: "img-1", UserId: "u1", ImageLink: "https://example.com/a.png"}
	repo := &testutil.MockImageRepository{
		GetImageByIDFn: func(_ context.Context, imageID, userID string) (*models.Image, error) {
			if imageID == "img-1" && userID == "u1" {
				return want, nil
			}
			return nil, repositories.ErrImageNotFound
		},
	}

	got, err := repo.GetImageByID(context.Background(), "img-1", "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ImageId != want.ImageId {
		t.Errorf("ImageId: got %q, want %q", got.ImageId, want.ImageId)
	}

	_, err = repo.GetImageByID(context.Background(), "img-missing", "u1")
	if !errors.Is(err, repositories.ErrImageNotFound) {
		t.Errorf("missing image: want ErrImageNotFound, got %v", err)
	}
}

func TestMockImageRepository_DeleteImageDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockImageRepository{}
	if err := repo.DeleteImage(context.Background(), "img-1", "u1"); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestMockImageRepository_DeleteImageFnCalled(t *testing.T) {
	var capturedImageID, capturedUserID string
	repo := &testutil.MockImageRepository{
		DeleteImageFn: func(_ context.Context, imageID, userID string) error {
			capturedImageID = imageID
			capturedUserID = userID
			return nil
		},
	}

	if err := repo.DeleteImage(context.Background(), "img-42", "u1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedImageID != "img-42" {
		t.Errorf("imageID: got %q, want %q", capturedImageID, "img-42")
	}
	if capturedUserID != "u1" {
		t.Errorf("userID: got %q, want %q", capturedUserID, "u1")
	}
}

func TestMockImageRepository_SoftDeleteImageDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockImageRepository{}
	if err := repo.SoftDeleteImage(context.Background(), "img-1", "u1"); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestMockImageRepository_SoftDeleteImageFnCalled(t *testing.T) {
	called := false
	repo := &testutil.MockImageRepository{
		SoftDeleteImageFn: func(_ context.Context, imageID, userID string) error {
			called = true
			if imageID == "" || userID == "" {
				return errors.New("imageID and userID are required")
			}
			return nil
		},
	}

	if err := repo.SoftDeleteImage(context.Background(), "img-1", "u1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("SoftDeleteImageFn was not called")
	}
}

func TestMockImageRepository_GetAllImagesDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockImageRepository{}
	images, err := repo.GetAllImages(context.Background(), 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(images) != 0 {
		t.Errorf("expected empty slice, got %d", len(images))
	}
}

func TestMockImageRepository_GetAllImagesFnPaginates(t *testing.T) {
	all := make([]models.Image, 10)
	for i := range all {
		all[i] = models.Image{ImageId: "img-" + string(rune('0'+i))}
	}
	repo := &testutil.MockImageRepository{
		GetAllImagesFn: func(_ context.Context, limit, offset int) ([]models.Image, error) {
			end := offset + limit
			if end > len(all) {
				end = len(all)
			}
			if offset >= len(all) {
				return []models.Image{}, nil
			}
			return all[offset:end], nil
		},
	}

	page1, err := repo.GetAllImages(context.Background(), 3, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(page1) != 3 {
		t.Errorf("page1: expected 3, got %d", len(page1))
	}

	page2, err := repo.GetAllImages(context.Background(), 3, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(page2) != 3 {
		t.Errorf("page2: expected 3, got %d", len(page2))
	}

	if page1[0].ImageId == page2[0].ImageId {
		t.Error("pages should contain different images")
	}
}