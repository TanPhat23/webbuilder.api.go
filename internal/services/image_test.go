package services

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/config"

	"my-go-app/internal/models"
)

// imageFakeFile satisfies multipart.File.
type imageFakeFile struct{ *bytes.Reader }

func (imageFakeFile) Close() error { return nil }

// imageCloudinary returns a CloudinaryService pointed at a local server. The
// handler decides the response for upload vs destroy requests.
func imageCloudinary(t *testing.T, uploadStatus int, destroyed *[]string) *CloudinaryService {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/destroy"):
			body, _ := io.ReadAll(r.Body)
			if destroyed != nil && strings.Contains(string(body), "pub1") {
				*destroyed = append(*destroyed, "pub1")
			}
			_, _ = w.Write([]byte(`{"result":"ok"}`))
		case strings.HasSuffix(r.URL.Path, "/upload"):
			if uploadStatus != http.StatusOK {
				w.WriteHeader(uploadStatus)
				_, _ = w.Write([]byte(`{"error":{"message":"nope"}}`))
				return
			}
			_, _ = w.Write([]byte(`{"public_id":"pub1","secure_url":"https://cdn/x.png","url":"http://cdn/x.png"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)
	cfg, err := config.NewFromParams("cloud", "key", "secret")
	if err != nil {
		t.Fatal(err)
	}
	cfg.API.UploadPrefix = srv.URL
	cld, err := cloudinary.NewFromConfiguration(*cfg)
	if err != nil {
		t.Fatal(err)
	}
	return &CloudinaryService{cld: cld}
}

func TestImageService_CreateImage(t *testing.T) {
	ctx := context.Background()

	t.Run("validation", func(t *testing.T) {
		repo := &fkImageRepo{}
		s := NewImageService(repo, nil)
		_, err := s.CreateImage(ctx, models.Image{ImageLink: "l"})
		wantErr(t, err, "userId is required")
		_, err = s.CreateImage(ctx, models.Image{UserId: "u"})
		wantErr(t, err, "image link is required")
		wantCalls(t, repo.Calls)
	})
	t.Run("fills id and timestamps", func(t *testing.T) {
		repo := &fkImageRepo{OnCreateImage: func(_ context.Context, i models.Image) (*models.Image, error) {
			return &i, nil
		}}
		got, err := NewImageService(repo, nil).CreateImage(ctx, models.Image{UserId: "u", ImageLink: "l"})
		wantNoErr(t, err)
		if got.ImageId == "" || got.CreatedAt.IsZero() || !got.UpdatedAt.Equal(got.CreatedAt) {
			t.Fatalf("got = %+v", got)
		}
	})
	t.Run("keeps provided id and timestamps", func(t *testing.T) {
		repo := &fkImageRepo{OnCreateImage: func(_ context.Context, i models.Image) (*models.Image, error) {
			return &i, nil
		}}
		in := models.Image{ImageId: "keep", UserId: "u", ImageLink: "l"}
		in.CreatedAt = in.CreatedAt.AddDate(2000, 0, 0)
		in.UpdatedAt = in.CreatedAt.AddDate(0, 1, 0)
		got, err := NewImageService(repo, nil).CreateImage(ctx, in)
		wantNoErr(t, err)
		if got.ImageId != "keep" || !got.CreatedAt.Equal(in.CreatedAt) || !got.UpdatedAt.Equal(in.UpdatedAt) {
			t.Fatalf("got = %+v", got)
		}
	})
	t.Run("repo error", func(t *testing.T) {
		repo := &fkImageRepo{OnCreateImage: func(context.Context, models.Image) (*models.Image, error) { return nil, errBoom }}
		_, err := NewImageService(repo, nil).CreateImage(ctx, models.Image{UserId: "u", ImageLink: "l"})
		if !errors.Is(err, errBoom) {
			t.Fatalf("err = %v", err)
		}
	})
}

func TestImageService_Uploads(t *testing.T) {
	ctx := context.Background()
	name := "pic"
	file := func() imageFakeFile { return imageFakeFile{bytes.NewReader([]byte("data"))} }

	type variant struct {
		name string
		call func(ctx context.Context, s *ImageService, uid string) (*models.ImageUploadResponse, error)
	}
	variants := []variant{
		{"file", func(ctx context.Context, s *ImageService, uid string) (*models.ImageUploadResponse, error) {
			return s.CreateUploadedImage(ctx, uid, "a.png", &name, file())
		}},
		{"base64", func(ctx context.Context, s *ImageService, uid string) (*models.ImageUploadResponse, error) {
			return s.CreateBase64UploadedImage(ctx, uid, "data:image/png;base64,AAAA", &name)
		}},
	}
	for _, v := range variants {
		t.Run(v.name, func(t *testing.T) {
			t.Run("missing user", func(t *testing.T) {
				_, err := v.call(ctx, NewImageService(&fkImageRepo{}, nil), "")
				wantErr(t, err, "userId is required")
			})
			t.Run("cloudinary failure wrapped, no repo call", func(t *testing.T) {
				repo := &fkImageRepo{}
				s := NewImageService(repo, imageCloudinary(t, http.StatusOK, nil))
				cctx, cancel := context.WithCancel(ctx)
				cancel()
				_, err := v.call(cctx, s, "u")
				wantErr(t, err, "failed to upload image to Cloudinary")
				wantCalls(t, repo.Calls)
			})
			t.Run("success", func(t *testing.T) {
				var created models.Image
				repo := &fkImageRepo{OnCreateImage: func(_ context.Context, i models.Image) (*models.Image, error) {
					created = i
					return &i, nil
				}}
				s := NewImageService(repo, imageCloudinary(t, http.StatusOK, nil))
				res, err := v.call(ctx, s, "u")
				wantNoErr(t, err)
				if res.ImageLink != "https://cdn/x.png" || res.ImageId == "" || res.ImageName != &name ||
					created.UserId != "u" || created.ImageLink != "https://cdn/x.png" {
					t.Fatalf("res=%+v created=%+v", res, created)
				}
			})
			t.Run("db failure rolls back cloudinary upload", func(t *testing.T) {
				var destroyed []string
				repo := &fkImageRepo{OnCreateImage: func(context.Context, models.Image) (*models.Image, error) { return nil, errBoom }}
				s := NewImageService(repo, imageCloudinary(t, http.StatusOK, &destroyed))
				_, err := v.call(ctx, s, "u")
				if !errors.Is(err, errBoom) {
					t.Fatalf("err = %v", err)
				}
				if len(destroyed) != 1 || destroyed[0] != "pub1" {
					t.Fatalf("destroyed = %v", destroyed)
				}
			})
		})
	}

	t.Run("file nil and data empty", func(t *testing.T) {
		s := NewImageService(&fkImageRepo{}, nil)
		_, err := s.CreateUploadedImage(ctx, "u", "a.png", nil, nil)
		wantErr(t, err, "image file is required")
		_, err = s.CreateBase64UploadedImage(ctx, "u", "", nil)
		wantErr(t, err, "image data is required")
	})
}

func TestImageService_Queries(t *testing.T) {
	ctx := context.Background()

	t.Run("GetImagesByUserID", func(t *testing.T) {
		repo := &fkImageRepo{}
		s := NewImageService(repo, nil)
		_, err := s.GetImagesByUserID(ctx, "")
		wantErr(t, err, "userId is required")
		repo.OnGetImagesByUserID = func(_ context.Context, u string) ([]models.Image, error) {
			return []models.Image{{UserId: u}}, nil
		}
		res, err := s.GetImagesByUserID(ctx, "u")
		wantNoErr(t, err)
		if len(res) != 1 || res[0].UserId != "u" {
			t.Fatalf("res = %+v", res)
		}
		repo.OnGetImagesByUserID = func(context.Context, string) ([]models.Image, error) { return nil, errBoom }
		if _, err := s.GetImagesByUserID(ctx, "u"); !errors.Is(err, errBoom) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("GetImageByID", func(t *testing.T) {
		repo := &fkImageRepo{}
		s := NewImageService(repo, nil)
		_, err := s.GetImageByID(ctx, "", "u")
		wantErr(t, err, "image id is required")
		_, err = s.GetImageByID(ctx, "i", "")
		wantErr(t, err, "userId is required")
		repo.OnGetImageByID = func(_ context.Context, i, u string) (*models.Image, error) {
			return &models.Image{ImageId: i, UserId: u}, nil
		}
		img, err := s.GetImageByID(ctx, "i", "u")
		wantNoErr(t, err)
		if img.ImageId != "i" || img.UserId != "u" {
			t.Fatalf("img = %+v", img)
		}
	})
	t.Run("GetAllImages", func(t *testing.T) {
		repo := &fkImageRepo{}
		s := NewImageService(repo, nil)
		_, err := s.GetAllImages(ctx, -1, 0)
		wantErr(t, err, "limit cannot be negative")
		_, err = s.GetAllImages(ctx, 0, -1)
		wantErr(t, err, "offset cannot be negative")
		wantCalls(t, repo.Calls)
		repo.OnGetAllImages = func(_ context.Context, l, o int) ([]models.Image, error) {
			if l != 5 || o != 10 {
				t.Errorf("l=%d o=%d", l, o)
			}
			return nil, nil
		}
		_, err = s.GetAllImages(ctx, 5, 10)
		wantNoErr(t, err)
	})
}

func TestImageService_Deletes(t *testing.T) {
	ctx := context.Background()
	ops := []struct {
		name string
		call func(s *ImageService, i, u string) error
		hook func(r *fkImageRepo, err error)
	}{
		{"DeleteImage", func(s *ImageService, i, u string) error { return s.DeleteImage(ctx, i, u) },
			func(r *fkImageRepo, err error) {
				r.OnDeleteImage = func(context.Context, string, string) error { return err }
			}},
		{"SoftDeleteImage", func(s *ImageService, i, u string) error { return s.SoftDeleteImage(ctx, i, u) },
			func(r *fkImageRepo, err error) {
				r.OnSoftDeleteImage = func(context.Context, string, string) error { return err }
			}},
	}
	for _, o := range ops {
		t.Run(o.name, func(t *testing.T) {
			repo := &fkImageRepo{}
			s := NewImageService(repo, nil)
			wantErr(t, o.call(s, "", "u"), "image id is required")
			wantErr(t, o.call(s, "i", ""), "userId is required")
			wantCalls(t, repo.Calls)
			o.hook(repo, nil)
			wantNoErr(t, o.call(s, "i", "u"))
			o.hook(repo, errBoom)
			if err := o.call(s, "i", "u"); !errors.Is(err, errBoom) {
				t.Fatalf("err = %v", err)
			}
		})
	}
}
