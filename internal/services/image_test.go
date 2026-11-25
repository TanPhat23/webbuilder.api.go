package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"io"
	"testing"
	"time"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
)

// imageFakeFile satisfies multipart.File.
type imageFakeFile struct{ *bytes.Reader }

func (imageFakeFile) Close() error { return nil }

type fakeObjectStorage struct {
	putErr       error
	presignedErr error
	deleteErr    error

	putCalls     []fakePutCall
	deletedKeys  []string
	presignedKey string
	presignedTTL time.Duration
	presignedURL string
}

type fakePutCall struct {
	key         string
	contentType string
	size        int64
	body        []byte
}

func (f *fakeObjectStorage) PutObject(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if f.putErr != nil {
		return f.putErr
	}
	raw, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	f.putCalls = append(f.putCalls, fakePutCall{
		key:         key,
		contentType: contentType,
		size:        size,
		body:        raw,
	})
	return nil
}

func (f *fakeObjectStorage) DeleteObject(_ context.Context, key string) error {
	f.deletedKeys = append(f.deletedKeys, key)
	return f.deleteErr
}

func (f *fakeObjectStorage) PresignedGetURL(_ context.Context, key string, ttl time.Duration) (string, error) {
	f.presignedKey = key
	f.presignedTTL = ttl
	if f.presignedErr != nil {
		return "", f.presignedErr
	}
	if f.presignedURL != "" {
		return f.presignedURL, nil
	}
	return "https://t3.storageapi.dev/signed/" + key, nil
}

func TestImageService_CreateImage(t *testing.T) {
	ctx := context.Background()

	t.Run("validation", func(t *testing.T) {
		repo := &fkImageRepo{}
		s := NewImageService(repo, nil)
		_, err := s.Create(ctx, models.Image{ImageLink: "l"})
		wantErr(t, err, "userId is required")
		_, err = s.Create(ctx, models.Image{UserId: "u"})
		wantErr(t, err, "image link is required")
		wantCalls(t, repo.Calls)
	})
	t.Run("fills id and timestamps", func(t *testing.T) {
		repo := &fkImageRepo{OnCreate: func(_ context.Context, i models.Image) (*models.Image, error) {
			return &i, nil
		}}
		got, err := NewImageService(repo, nil).Create(ctx, models.Image{UserId: "u", ImageLink: "l"})
		wantNoErr(t, err)
		if got.ImageId == "" || got.CreatedAt.IsZero() || !got.UpdatedAt.Equal(got.CreatedAt) {
			t.Fatalf("got = %+v", got)
		}
	})
	t.Run("keeps provided id and timestamps", func(t *testing.T) {
		repo := &fkImageRepo{OnCreate: func(_ context.Context, i models.Image) (*models.Image, error) {
			return &i, nil
		}}
		in := models.Image{ImageId: "keep", UserId: "u", ImageLink: "l"}
		in.CreatedAt = in.CreatedAt.AddDate(2000, 0, 0)
		in.UpdatedAt = in.CreatedAt.AddDate(0, 1, 0)
		got, err := NewImageService(repo, nil).Create(ctx, in)
		wantNoErr(t, err)
		if got.ImageId != "keep" || !got.CreatedAt.Equal(in.CreatedAt) || !got.UpdatedAt.Equal(in.UpdatedAt) {
			t.Fatalf("got = %+v", got)
		}
	})
	t.Run("repo error", func(t *testing.T) {
		repo := &fkImageRepo{OnCreate: func(context.Context, models.Image) (*models.Image, error) { return nil, errBoom }}
		_, err := NewImageService(repo, nil).Create(ctx, models.Image{UserId: "u", ImageLink: "l"})
		if !errors.Is(err, errBoom) {
			t.Fatalf("err = %v", err)
		}
	})
}

func TestImageService_Uploads(t *testing.T) {
	ctx := context.Background()
	name := "pic"
	pngRaw := []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDRdata")
	pngB64 := "data:image/png;base64," + base64.StdEncoding.EncodeToString(pngRaw)
	file := func() imageFakeFile {
		return imageFakeFile{bytes.NewReader(pngRaw)}
	}

	type variant struct {
		name string
		call func(ctx context.Context, s *ImageService, uid string) (*models.ImageUploadResponse, error)
	}
	variants := []variant{
		{"file", func(ctx context.Context, s *ImageService, uid string) (*models.ImageUploadResponse, error) {
			return s.Upload(ctx, uid, &name, file())
		}},
		{"base64", func(ctx context.Context, s *ImageService, uid string) (*models.ImageUploadResponse, error) {
			return s.UploadBase64(ctx, uid, pngB64, &name)
		}},
	}
	for _, v := range variants {
		t.Run(v.name, func(t *testing.T) {
			t.Run("missing user", func(t *testing.T) {
				_, err := v.call(ctx, NewImageService(&fkImageRepo{}, nil), "")
				wantErr(t, err, "userId is required")
			})
			t.Run("storage failure wrapped, no repo call", func(t *testing.T) {
				repo := &fkImageRepo{}
				store := &fakeObjectStorage{putErr: errors.New("storage down")}
				s := NewImageServiceWithBaseURL(repo, store, "https://api.example.com")
				_, err := v.call(ctx, s, "u")
				wantErr(t, err, "failed to upload image to object storage")
				wantCalls(t, repo.Calls)
			})
			t.Run("success", func(t *testing.T) {
				var created models.Image
				repo := &fkImageRepo{OnCreate: func(_ context.Context, i models.Image) (*models.Image, error) {
					created = i
					return &i, nil
				}}
				store := &fakeObjectStorage{}
				s := NewImageServiceWithBaseURL(repo, store, "https://api.example.com/")
				res, err := v.call(ctx, s, "u")
				wantNoErr(t, err)
				wantLink := "https://api.example.com/api/v1/public/images/" + res.ImageId
				if res.ImageId == "" || res.ImageLink != wantLink || res.ImageName != &name ||
					created.UserId != "u" || created.ImageLink != wantLink {
					t.Fatalf("res=%+v created=%+v", res, created)
				}
				if len(store.putCalls) != 1 || store.putCalls[0].key != "images/u/"+res.ImageId ||
					store.putCalls[0].contentType != "image/png" || store.putCalls[0].size != int64(len(pngRaw)) {
					t.Fatalf("putCalls=%+v", store.putCalls)
				}
			})
			t.Run("db failure rolls back object upload", func(t *testing.T) {
				repo := &fkImageRepo{OnCreate: func(context.Context, models.Image) (*models.Image, error) { return nil, errBoom }}
				store := &fakeObjectStorage{}
				s := NewImageServiceWithBaseURL(repo, store, "https://api.example.com")
				_, err := v.call(ctx, s, "u")
				if !errors.Is(err, errBoom) {
					t.Fatalf("err = %v", err)
				}
				if len(store.putCalls) != 1 || len(store.deletedKeys) != 1 || store.deletedKeys[0] != store.putCalls[0].key {
					t.Fatalf("rollback mismatch putCalls=%+v deletedKeys=%v", store.putCalls, store.deletedKeys)
				}
			})
		})
	}

	t.Run("file nil and data empty", func(t *testing.T) {
		s := NewImageService(&fkImageRepo{}, nil)
		_, err := s.Upload(ctx, "u", nil, nil)
		wantErr(t, err, "image file is required")
		_, err = s.UploadBase64(ctx, "u", "", nil)
		wantErr(t, err, "image data is required")
	})

	t.Run("nil storage service returns explicit error", func(t *testing.T) {
		s := NewImageService(&fkImageRepo{}, nil)
		_, err := s.Upload(ctx, "u", nil, file())
		wantErr(t, err, "image upload service is not configured")
		_, err = s.UploadBase64(ctx, "u", pngB64, nil)
		wantErr(t, err, "image upload service is not configured")
	})

	t.Run("base64 rejects svg and invalid encoding", func(t *testing.T) {
		store := &fakeObjectStorage{}
		s := NewImageService(&fkImageRepo{}, store)
		svgB64 := "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte("<svg><script>alert(1)</script></svg>"))
		_, err := s.UploadBase64(ctx, "u", svgB64, nil)
		wantErr(t, err, "invalid image content type")
		_, err = s.UploadBase64(ctx, "u", "data:image/png;base64,!!!", nil)
		wantErr(t, err, "invalid base64 image encoding")
		if len(store.putCalls) != 0 {
			t.Fatal("invalid payload must not reach object storage")
		}
	})
}

func TestImageService_GetPublicImageURL(t *testing.T) {
	ctx := context.Background()

	repo := &fkImageRepo{
		OnGetPublicByID: func(_ context.Context, id string) (*models.Image, error) {
			if id == "missing" {
				return nil, repositories.ErrImageNotFound
			}
			return &models.Image{ImageId: id, UserId: "u1"}, nil
		},
	}
	store := &fakeObjectStorage{presignedURL: "https://t3.storageapi.dev/signed-url"}
	s := NewImageService(repo, store)

	if _, err := s.PublicURL(ctx, ""); err == nil {
		t.Fatal("expected error for empty image id")
	}
	if _, err := NewImageService(repo, nil).PublicURL(ctx, "img-1"); err == nil {
		t.Fatal("expected error when storage is nil")
	}
	if _, err := s.PublicURL(ctx, "missing"); !errors.Is(err, repositories.ErrImageNotFound) {
		t.Fatalf("err = %v, want ErrImageNotFound", err)
	}

	got, err := s.PublicURL(ctx, "img-1")
	wantNoErr(t, err)
	if got != "https://t3.storageapi.dev/signed-url" || store.presignedKey != "images/u1/img-1" || store.presignedTTL != presignedImageURLTTL {
		t.Fatalf("got=%q key=%q ttl=%v", got, store.presignedKey, store.presignedTTL)
	}
}

func TestImageService_Queries(t *testing.T) {
	ctx := context.Background()

	t.Run("ListByUser", func(t *testing.T) {
		repo := &fkImageRepo{}
		s := NewImageService(repo, nil)
		_, err := s.ListByUser(ctx, "")
		wantErr(t, err, "userId is required")
		repo.OnListByUser = func(_ context.Context, u string) ([]models.Image, error) {
			return []models.Image{{UserId: u}}, nil
		}
		res, err := s.ListByUser(ctx, "u")
		wantNoErr(t, err)
		if len(res) != 1 || res[0].UserId != "u" {
			t.Fatalf("res = %+v", res)
		}
		repo.OnListByUser = func(context.Context, string) ([]models.Image, error) { return nil, errBoom }
		if _, err := s.ListByUser(ctx, "u"); !errors.Is(err, errBoom) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("GetByID", func(t *testing.T) {
		repo := &fkImageRepo{}
		s := NewImageService(repo, nil)
		_, err := s.GetByID(ctx, "", "u")
		wantErr(t, err, "image id is required")
		_, err = s.GetByID(ctx, "i", "")
		wantErr(t, err, "userId is required")
		repo.OnGetByID = func(_ context.Context, i, u string) (*models.Image, error) {
			return &models.Image{ImageId: i, UserId: u}, nil
		}
		img, err := s.GetByID(ctx, "i", "u")
		wantNoErr(t, err)
		if img.ImageId != "i" || img.UserId != "u" {
			t.Fatalf("img = %+v", img)
		}
	})
	t.Run("ListAll", func(t *testing.T) {
		repo := &fkImageRepo{}
		s := NewImageService(repo, nil)
		_, err := s.ListAll(ctx, -1, 0)
		wantErr(t, err, "limit cannot be negative")
		_, err = s.ListAll(ctx, 0, -1)
		wantErr(t, err, "offset cannot be negative")
		wantCalls(t, repo.Calls)
		repo.OnListAll = func(_ context.Context, l, o int) ([]models.Image, error) {
			if l != 5 || o != 10 {
				t.Errorf("l=%d o=%d", l, o)
			}
			return nil, nil
		}
		_, err = s.ListAll(ctx, 5, 10)
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
		{"Delete", func(s *ImageService, i, u string) error { return s.Delete(ctx, i, u) },
			func(r *fkImageRepo, err error) {
				r.OnDelete = func(context.Context, string, string) error { return err }
			}},
		{"SoftDelete", func(s *ImageService, i, u string) error { return s.SoftDelete(ctx, i, u) },
			func(r *fkImageRepo, err error) {
				r.OnSoftDelete = func(context.Context, string, string) error { return err }
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

	t.Run("Delete removes object on success", func(t *testing.T) {
		repo := &fkImageRepo{OnDelete: func(context.Context, string, string) error { return nil }}
		store := &fakeObjectStorage{}
		s := NewImageService(repo, store)
		wantNoErr(t, s.Delete(ctx, "img1", "u1"))
		if len(store.deletedKeys) != 1 || store.deletedKeys[0] != "images/u1/img1" {
			t.Fatalf("deletedKeys = %v", store.deletedKeys)
		}
	})

	t.Run("Delete does not touch storage when db delete fails", func(t *testing.T) {
		repo := &fkImageRepo{OnDelete: func(context.Context, string, string) error { return repositories.ErrImageNotFound }}
		store := &fakeObjectStorage{}
		s := NewImageService(repo, store)
		if err := s.Delete(ctx, "img1", "u1"); !errors.Is(err, repositories.ErrImageNotFound) {
			t.Fatalf("err = %v", err)
		}
		if len(store.deletedKeys) != 0 {
			t.Fatalf("deletedKeys = %v, want empty", store.deletedKeys)
		}
	})
}
