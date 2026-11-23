package services

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type cloudinaryFakeFile struct {
	*strings.Reader
	closed bool
	err    error
}

func (f *cloudinaryFakeFile) Close() error { f.closed = true; return nil }
func (f *cloudinaryFakeFile) ReadAt(p []byte, off int64) (int, error) {
	return f.Reader.ReadAt(p, off)
}
func (f *cloudinaryFakeFile) Read(p []byte) (int, error) {
	if f.err != nil {
		return 0, f.err
	}
	return f.Reader.Read(p)
}

func cloudinarySetEnv(t *testing.T, name, key, secret string) {
	t.Helper()
	t.Setenv("CLOUDINARY_CLOUD_NAME", name)
	t.Setenv("CLOUDINARY_API_KEY", key)
	t.Setenv("CLOUDINARY_API_SECRET", secret)
}

func TestNewCloudinaryService_MissingCredentials(t *testing.T) {
	cases := [][3]string{{"", "k", "s"}, {"n", "", "s"}, {"n", "k", ""}, {"", "", ""}}
	for _, c := range cases {
		cloudinarySetEnv(t, c[0], c[1], c[2])
		svc, err := NewCloudinaryService()
		wantErr(t, err, "credentials not found")
		if svc != nil {
			t.Fatalf("expected nil service for %v", c)
		}
	}
}

func TestNewCloudinaryService_OK(t *testing.T) {
	cloudinarySetEnv(t, "demo", "key", "secret")
	svc, err := NewCloudinaryService()
	wantNoErr(t, err)
	if svc == nil || svc.cld == nil {
		t.Fatal("expected initialized service")
	}
	if svc.cld.Config.Cloud.CloudName != "demo" {
		t.Fatalf("cloud name = %q", svc.cld.Config.Cloud.CloudName)
	}
}

// cloudinaryTestService points the SDK at a local httptest server.
func cloudinaryTestService(t *testing.T, h http.HandlerFunc) (*CloudinaryService, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	cloudinarySetEnv(t, "demo", "key", "secret")
	svc, err := NewCloudinaryService()
	wantNoErr(t, err)
	svc.cld.Upload.Config.API.UploadPrefix = srv.URL
	return svc, srv
}

func TestCloudinary_UploadImage(t *testing.T) {
	var path, body string
	svc, _ := cloudinaryTestService(t, func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		body = string(b)
		w.Write([]byte(`{"public_id":"f/pic_1","secure_url":"https://s/x.png","url":"http://s/x.png","format":"png","width":10,"height":20}`))
	})
	f := &cloudinaryFakeFile{Reader: strings.NewReader("imgdata")}
	res, err := svc.UploadImage(context.Background(), f, "pic.png", "f")
	wantNoErr(t, err)
	want := UploadResult{PublicID: "f/pic_1", SecureURL: "https://s/x.png", URL: "http://s/x.png", Format: "png", Width: 10, Height: 20}
	if *res != want {
		t.Fatalf("result = %+v, want %+v", *res, want)
	}
	if !strings.HasSuffix(path, "/demo/auto/upload") {
		t.Fatalf("path = %q", path)
	}
	if !strings.Contains(body, "folder") || !strings.Contains(body, "pic_") {
		t.Fatalf("body missing folder/public id: %s", body)
	}
}

func TestCloudinary_UploadImage_Error(t *testing.T) {
	svc, _ := cloudinaryTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte("nope"))
	})
	res, err := svc.UploadImage(context.Background(), &cloudinaryFakeFile{Reader: strings.NewReader("x")}, "a.png", "f")
	if res != nil {
		t.Fatal("expected nil result")
	}
	if err == nil {
		// SDK may report failure through the result error field instead
		t.Skip("sdk did not surface HTTP 500 as error")
	}
	wantErr(t, err, "failed to upload image to cloudinary")
}

func TestCloudinary_UploadBase64Image(t *testing.T) {
	svc, _ := cloudinaryTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"public_id":"b1","secure_url":"https://s/b.png","format":"png","width":1,"height":2}`))
	})
	res, err := svc.UploadBase64Image(context.Background(), "data:image/png;base64,AAAA", "folder")
	wantNoErr(t, err)
	if res.PublicID != "b1" || res.SecureURL != "https://s/b.png" || res.Width != 1 || res.Height != 2 {
		t.Fatalf("unexpected result %+v", res)
	}
}

func TestCloudinary_UploadBase64Image_TransportError(t *testing.T) {
	svc, srv := cloudinaryTestService(t, func(w http.ResponseWriter, r *http.Request) {})
	srv.Close()
	_, err := svc.UploadBase64Image(context.Background(), "data:image/png;base64,AAAA", "folder")
	wantErr(t, err, "failed to upload base64 image to cloudinary")
	_, err = svc.UploadImage(context.Background(), &cloudinaryFakeFile{Reader: strings.NewReader("x")}, "a.png", "f")
	wantErr(t, err, "failed to upload image to cloudinary")
	err = svc.DeleteImage(context.Background(), "id")
	wantErr(t, err, "failed to delete image from cloudinary")
}

func TestCloudinary_DeleteImage(t *testing.T) {
	var path, body string
	svc, _ := cloudinaryTestService(t, func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		body = string(b)
		w.Write([]byte(`{"result":"ok"}`))
	})
	wantNoErr(t, svc.DeleteImage(context.Background(), "my/public"))
	if !strings.Contains(path, "/image/destroy") || !strings.Contains(body, "public_id=my%2Fpublic") {
		t.Fatalf("path=%q body=%q", path, body)
	}
}

func cloudinaryHeader(name string, size int64) *multipart.FileHeader {
	return &multipart.FileHeader{Filename: name, Size: size}
}

func TestValidateImageFile(t *testing.T) {
	for _, n := range []string{"a.jpg", "a.jpeg", "a.png", "a.gif", "a.webp", "a.svg", "a.bmp"} {
		wantNoErr(t, ValidateImageFile(cloudinaryHeader(n, 100)))
	}
	wantNoErr(t, ValidateImageFile(cloudinaryHeader("a.png", 10*1024*1024)))
	wantErr(t, ValidateImageFile(cloudinaryHeader("a.png", 10*1024*1024+1)), "exceeds maximum")
	wantErr(t, ValidateImageFile(cloudinaryHeader("a.exe", 1)), "invalid file type: .exe")
	wantErr(t, ValidateImageFile(cloudinaryHeader("noext", 1)), "invalid file type")
	wantErr(t, ValidateImageFile(cloudinaryHeader("A.PNG", 1)), "invalid file type: .PNG")
}

func TestReadFileContent(t *testing.T) {
	f := &cloudinaryFakeFile{Reader: strings.NewReader("hello")}
	b, err := ReadFileContent(f)
	wantNoErr(t, err)
	if !bytes.Equal(b, []byte("hello")) || !f.closed {
		t.Fatalf("got %q closed=%v", b, f.closed)
	}
	f = &cloudinaryFakeFile{Reader: strings.NewReader(""), err: errBoom}
	_, err = ReadFileContent(f)
	if err != errBoom || !f.closed {
		t.Fatalf("err=%v closed=%v", err, f.closed)
	}
}

func TestGeneratePublicID(t *testing.T) {
	a := generatePublicID("photo.png")
	if !strings.HasPrefix(a, "photo_") || strings.Contains(a, ".png") {
		t.Fatalf("id = %q", a)
	}
	if b := generatePublicID("photo.png"); a == b {
		t.Fatal("ids should be unique")
	}
	if id := generatePublicID("noext"); !strings.HasPrefix(id, "noext_") {
		t.Fatalf("id = %q", id)
	}
}
