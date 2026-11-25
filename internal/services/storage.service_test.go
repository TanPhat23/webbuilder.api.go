package services

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
)

const testPNGHeader = "\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDRimgdata"

type storageFakeFile struct {
	*bytes.Reader
	closed bool
	err    error
}

func (f *storageFakeFile) Close() error { f.closed = true; return nil }
func (f *storageFakeFile) Read(p []byte) (int, error) {
	if f.err != nil {
		return 0, f.err
	}
	return f.Reader.Read(p)
}
func (f *storageFakeFile) ReadAt(p []byte, off int64) (int, error) {
	return f.Reader.ReadAt(p, off)
}

func TestParseS3Endpoint(t *testing.T) {
	cases := []struct {
		raw      string
		override string
		wantHost string
		wantSSL  bool
		wantErr  bool
	}{
		{"https://t3.storageapi.dev", "", "t3.storageapi.dev", true, false},
		{"http://localhost:9000", "", "localhost:9000", false, false},
		{"t3.storageapi.dev", "", "t3.storageapi.dev", true, false},
		{"localhost:9000", "false", "localhost:9000", false, false},
		{"", "", "", false, true},
		{"https://host/badpath", "", "", false, true},
		{"ftp://host", "", "", false, true},
	}

	for _, tc := range cases {
		host, ssl, err := ParseS3Endpoint(tc.raw, tc.override)
		if (err != nil) != tc.wantErr {
			t.Fatalf("ParseS3Endpoint(%q, %q) err=%v wantErr=%v", tc.raw, tc.override, err, tc.wantErr)
		}
		if host != tc.wantHost || ssl != tc.wantSSL {
			t.Fatalf("ParseS3Endpoint(%q, %q) = (%q, %v), want (%q, %v)", tc.raw, tc.override, host, ssl, tc.wantHost, tc.wantSSL)
		}
	}
}

func TestNewS3StorageConfigFromEnv(t *testing.T) {
	t.Run("missing env", func(t *testing.T) {
		for _, k := range []string{
			"S3_ENDPOINT", "RAILWAY_BUCKET_ENDPOINT", "MINIO_ENDPOINT",
			"S3_BUCKET", "RAILWAY_BUCKET_NAME", "MINIO_BUCKET",
			"S3_ACCESS_KEY_ID", "RAILWAY_BUCKET_ACCESS_KEY_ID", "MINIO_ACCESS_KEY",
			"S3_SECRET_ACCESS_KEY", "RAILWAY_BUCKET_SECRET_ACCESS_KEY", "MINIO_SECRET_KEY",
		} {
			t.Setenv(k, "")
		}
		if _, err := NewS3StorageConfigFromEnv(); err == nil {
			t.Fatal("expected error when credentials are missing")
		}
	})

	t.Run("railway variables", func(t *testing.T) {
		t.Setenv("S3_ENDPOINT", "")
		t.Setenv("S3_BUCKET", "")
		t.Setenv("S3_ACCESS_KEY_ID", "")
		t.Setenv("S3_SECRET_ACCESS_KEY", "")
		t.Setenv("RAILWAY_BUCKET_ENDPOINT", "https://t3.storageapi.dev")
		t.Setenv("RAILWAY_BUCKET_NAME", "stackable-box-123")
		t.Setenv("RAILWAY_BUCKET_ACCESS_KEY_ID", "ak")
		t.Setenv("RAILWAY_BUCKET_SECRET_ACCESS_KEY", "sk")
		t.Setenv("RAILWAY_BUCKET_REGION", "auto")
		t.Setenv("RAILWAY_BUCKET_URL_STYLE", "virtual-host")

		cfg, err := NewS3StorageConfigFromEnv()
		wantNoErr(t, err)
		if cfg.Endpoint != "t3.storageapi.dev" || !cfg.UseSSL || cfg.Bucket != "stackable-box-123" ||
			cfg.Region != "auto" || cfg.BucketLookup != minio.BucketLookupDNS {
			t.Fatalf("unexpected cfg: %+v", cfg)
		}
	})
}

func TestPublicAPIURL(t *testing.T) {
	t.Setenv("API_PUBLIC_URL", "https://api.example.com/")
	t.Setenv("RAILWAY_PUBLIC_DOMAIN", "ignored.up.railway.app")
	if got := PublicAPIURL(); got != "https://api.example.com" {
		t.Fatalf("got %q", got)
	}

	t.Setenv("API_PUBLIC_URL", "")
	t.Setenv("RAILWAY_PUBLIC_DOMAIN", "webbuilderapigo-production.up.railway.app")
	if got := PublicAPIURL(); got != "https://webbuilderapigo-production.up.railway.app" {
		t.Fatalf("got %q", got)
	}

	t.Setenv("RAILWAY_PUBLIC_DOMAIN", "")
	t.Setenv("PORT", "9090")
	if got := PublicAPIURL(); got != "http://localhost:9090" {
		t.Fatalf("got %q", got)
	}
}

func TestS3StorageService_Operations(t *testing.T) {
	var seenMethod, seenPath, seenContentType, seenBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenMethod = r.Method
		seenPath = r.URL.Path
		seenContentType = r.Header.Get("Content-Type")
		b, _ := io.ReadAll(r.Body)
		seenBody = string(b)
		w.Header().Set("ETag", `"etag-1"`)
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	svc, err := NewS3StorageServiceWithConfig(S3StorageConfig{
		Endpoint:     strings.TrimPrefix(srv.URL, "http://"),
		Bucket:       "stackable-box",
		AccessKeyID:  "ak",
		SecretKey:    "sk",
		Region:       "auto",
		UseSSL:       false,
		BucketLookup: minio.BucketLookupPath,
	})
	wantNoErr(t, err)

	ctx := t.Context()
	if err := svc.PutObject(ctx, "images/u1/img1", strings.NewReader(testPNGHeader), int64(len(testPNGHeader)), "image/png"); err != nil {
		t.Fatalf("PutObject: %v", err)
	}
	if seenMethod != http.MethodPut || seenPath != "/stackable-box/images/u1/img1" || seenContentType != "image/png" || !strings.Contains(seenBody, testPNGHeader) {
		t.Fatalf("put mismatch method=%q path=%q ct=%q body=%q", seenMethod, seenPath, seenContentType, seenBody)
	}

	presigned, err := svc.PresignedGetURL(ctx, "images/u1/img1", 5*time.Minute)
	wantNoErr(t, err)
	if !strings.Contains(presigned, "/stackable-box/images/u1/img1") || !strings.Contains(presigned, "X-Amz-Signature=") {
		t.Fatalf("unexpected presigned url: %q", presigned)
	}

	if err := svc.DeleteObject(ctx, "images/u1/img1"); err != nil {
		t.Fatalf("DeleteObject: %v", err)
	}
	if seenMethod != http.MethodDelete || seenPath != "/stackable-box/images/u1/img1" {
		t.Fatalf("delete mismatch method=%q path=%q", seenMethod, seenPath)
	}
}

func TestImageValidators(t *testing.T) {
	hdr := func(name string, size int64) *multipart.FileHeader {
		return &multipart.FileHeader{Filename: name, Size: size}
	}
	for _, n := range []string{"a.jpg", "a.JPEG", "a.png", "a.gif", "a.webp", "a.bmp"} {
		wantNoErr(t, ValidateImageFile(hdr(n, 100)))
	}
	wantNoErr(t, ValidateImageFile(hdr("a.png", 10*1024*1024)))
	wantErr(t, ValidateImageFile(hdr("a.png", 10*1024*1024+1)), "exceeds maximum")
	wantErr(t, ValidateImageFile(hdr("a.exe", 1)), "invalid file type: .exe")
	wantErr(t, ValidateImageFile(hdr("a.svg", 1)), "invalid file type: .svg")
	wantErr(t, ValidateImageFile(hdr("noext", 1)), "invalid file type")

	ct, err := DetectImageContentType(&storageFakeFile{Reader: bytes.NewReader([]byte(testPNGHeader))})
	wantNoErr(t, err)
	if ct != "image/png" {
		t.Fatalf("got %q, want image/png", ct)
	}
	wantNoErr(t, ValidateImageContent(&storageFakeFile{Reader: bytes.NewReader([]byte("\xff\xd8\xff\xe0JFIF"))}))
	wantNoErr(t, ValidateImageContent(&storageFakeFile{Reader: bytes.NewReader([]byte("GIF89a"))}))
	wantErr(t, ValidateImageContent(&storageFakeFile{Reader: bytes.NewReader([]byte("<svg><script>alert(1)</script></svg>"))}), "invalid image content type")
	wantErr(t, ValidateImageContent(nil), "image file is required")
	wantErr(t, ValidateImageContent(&storageFakeFile{Reader: bytes.NewReader([]byte("")), err: errors.New("boom")}), "failed to read image content")

	f := &storageFakeFile{Reader: bytes.NewReader([]byte("hello"))}
	b, err := ReadFileContent(f)
	wantNoErr(t, err)
	if string(b) != "hello" || !f.closed {
		t.Fatalf("b=%q closed=%v", b, f.closed)
	}
}
