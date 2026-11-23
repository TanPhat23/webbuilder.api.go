package handlers

import (
	"context"
	"testing"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"

	"my-go-app/internal/models"
	"my-go-app/internal/services"
)

type userHandlerStub struct {
	services.UserServiceInterface
	err error
	got string
}

func (s *userHandlerStub) SearchUsers(_ context.Context, q string) ([]models.User, error) {
	s.got = q
	if s.err != nil {
		return nil, s.err
	}
	return []models.User{{Id: "u1", Email: "ann@example.com"}, {Id: "u2", Email: "andy@example.com"}}, nil
}

func (s *userHandlerStub) GetUserByEmail(_ context.Context, email string) (*models.User, error) {
	s.got = email
	if s.err != nil {
		return nil, s.err
	}
	return &models.User{Id: "u1", Email: email}, nil
}

func (s *userHandlerStub) GetUserByUsername(_ context.Context, username string) (*models.User, error) {
	s.got = username
	if s.err != nil {
		return nil, s.err
	}
	return &models.User{Id: "u1", Email: username + "@example.com"}, nil
}

func userHandlerApp(svc *userHandlerStub) *fiber.App {
	h := NewUserHandler(svc)
	return newKitApp("", func(app *fiber.App) {
		app.Get("/users/search", h.SearchUsers)
		app.Get("/users/email/:email", h.GetUserByEmail)
		app.Get("/users/username/:username", h.GetUserByUsername)
		app.Get("/noemail", h.GetUserByEmail)
		app.Get("/nousername", h.GetUserByUsername)
	})
}

func TestSearchUsers(t *testing.T) {
	svc := &userHandlerStub{}
	r := kitSend(t, userHandlerApp(svc), "GET", "/users/search?q=%20an%20", "")
	kitExpect(t, r, 200, "ann@example.com", "andy@example.com")
	if svc.got != "an" {
		t.Fatalf("query = %q, want it trimmed to \"an\"", svc.got)
	}
	for _, path := range []string{"/users/search", "/users/search?q=", "/users/search?q=%20%20"} {
		svc := &userHandlerStub{}
		kitExpect(t, kitSend(t, userHandlerApp(svc), "GET", path, ""), 400, "Query parameter 'q' is required")
		if svc.got != "" {
			t.Fatalf("%s reached the service", path)
		}
	}
	kitExpect(t, kitSend(t, userHandlerApp(&userHandlerStub{err: errInternal}), "GET", "/users/search?q=an", ""), 500, "Failed to search users")
}

func TestGetUserByEmailAndUsername(t *testing.T) {
	cases := []struct {
		name, path string
		err        error
		want       int
		expect     string
		wantArg    string
	}{
		{"email ok", "/users/email/ann@example.com", nil, 200, "ann@example.com", "ann@example.com"},
		{"email missing param", "/noemail", nil, 400, "email is required", ""},
		{"email not found", "/users/email/x@y.z", gorm.ErrRecordNotFound, 404, "User not found", "x@y.z"},
		{"email failure", "/users/email/x@y.z", errInternal, 500, "Failed to retrieve user", "x@y.z"},
		{"username ok", "/users/username/ann", nil, 200, "ann@example.com", "ann"},
		{"username missing param", "/nousername", nil, 400, "username is required", ""},
		{"username not found", "/users/username/ann", gorm.ErrRecordNotFound, 404, "User not found", "ann"},
		{"username failure", "/users/username/ann", errInternal, 500, "Failed to retrieve user", "ann"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := &userHandlerStub{err: tc.err}
			kitExpect(t, kitSend(t, userHandlerApp(svc), "GET", tc.path, ""), tc.want, tc.expect)
			if svc.got != tc.wantArg {
				t.Fatalf("service got %q, want %q", svc.got, tc.wantArg)
			}
		})
	}
}
