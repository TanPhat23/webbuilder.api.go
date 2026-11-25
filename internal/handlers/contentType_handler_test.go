package handlers

import (
	"context"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"

	"my-go-app/internal/models"
	"my-go-app/internal/services"
)

type contentTypeSvcStub struct {
	services.ContentTypeServiceInterface
	err     error
	gotID   string
	gotType *models.ContentType
	deleted string
}

func (s *contentTypeSvcStub) List(context.Context) ([]models.ContentType, error) {
	return []models.ContentType{{Id: "ct1", Name: "Blog"}}, s.err
}
func (s *contentTypeSvcStub) GetByID(_ context.Context, id string) (*models.ContentType, error) {
	s.gotID = id
	if s.err != nil {
		return nil, s.err
	}
	return &models.ContentType{Id: id, Name: "Blog"}, nil
}
func (s *contentTypeSvcStub) Create(_ context.Context, ct *models.ContentType) (*models.ContentType, error) {
	s.gotType = ct
	if s.err != nil {
		return nil, s.err
	}
	ct.Id = "new-id"
	return ct, nil
}
func (s *contentTypeSvcStub) Update(_ context.Context, id string, ct *models.ContentType) (*models.ContentType, error) {
	s.gotID, s.gotType = id, ct
	if s.err != nil {
		return nil, s.err
	}
	return &models.ContentType{Id: id, Name: ct.Name, Description: ct.Description}, nil
}
func (s *contentTypeSvcStub) Delete(_ context.Context, id string) error {
	s.deleted = id
	return s.err
}

func (s *contentTypeSvcStub) ListForUser(_ context.Context, _, _ string) ([]models.ContentType, error) {
	return s.List(context.Background())
}
func (s *contentTypeSvcStub) GetByIDForUser(_ context.Context, id, _ string) (*models.ContentType, error) {
	return s.GetByID(context.Background(), id)
}
func (s *contentTypeSvcStub) CreateForUser(_ context.Context, ct *models.ContentType, _ string) (*models.ContentType, error) {
	return s.Create(context.Background(), ct)
}
func (s *contentTypeSvcStub) PatchForUser(_ context.Context, id, _ string, updates map[string]any) (*models.ContentType, error) {
	ct := &models.ContentType{}
	if name, ok := updates["name"].(string); ok {
		ct.Name = name
	}
	if desc, ok := updates["description"].(*string); ok {
		ct.Description = desc
	}
	return s.Update(context.Background(), id, ct)
}
func (s *contentTypeSvcStub) UpdateForUser(_ context.Context, id, _ string, ct *models.ContentType) (*models.ContentType, error) {
	return s.Update(context.Background(), id, ct)
}
func (s *contentTypeSvcStub) DeleteForUser(_ context.Context, id, _ string) error {
	return s.Delete(context.Background(), id)
}

func contentTypeKit(svc *contentTypeSvcStub) *fiber.App {
	h := NewContentTypeHandler(svc)
	return newKitApp("u1", func(app *fiber.App) {
		app.Get("/content-types", h.List)
		app.Post("/content-types", h.Create)
		app.Get("/content-types/:id", h.GetByID)
		app.Patch("/content-types/:id", h.Update)
		app.Delete("/content-types/:id", h.Delete)
	})
}

func TestContentTypeHandlerStatuses(t *testing.T) {
	cases := []struct {
		name         string
		method, path string
		body         string
		err          error
		want         int
		contains     string
	}{
		{"list ok", "GET", "/content-types", "", nil, 200, "Blog"},
		{"list internal", "GET", "/content-types", "", errInternal, 500, "Failed to retrieve content types"},
		{"get ok", "GET", "/content-types/ct1", "", nil, 200, "ct1"},
		{"get missing", "GET", "/content-types/ct1", "", gorm.ErrRecordNotFound, 404, "Content type not found"},
		{"get internal", "GET", "/content-types/ct1", "", errInternal, 500, "Failed to retrieve content type"},
		{"create ok", "POST", "/content-types", `{"name":"News"}`, nil, 201, "new-id"},
		{"create bad json", "POST", "/content-types", `{"name":`, nil, 0, ""},
		{"create invalid", "POST", "/content-types", `{"description":"x"}`, nil, 0, ""},
		{"create internal", "POST", "/content-types", `{"name":"News"}`, errInternal, 500, "Failed to create content type"},
		{"update ok", "PATCH", "/content-types/ct1", `{"name":"Posts"}`, nil, 200, "Posts"},
		{"update nothing", "PATCH", "/content-types/ct1", `{}`, nil, 400, "At least one field must be updated"},
		{"update bad json", "PATCH", "/content-types/ct1", `nope`, nil, 0, ""},
		{"update missing", "PATCH", "/content-types/ct1", `{"name":"Posts"}`, gorm.ErrRecordNotFound, 404, "Content type not found"},
		{"update internal", "PATCH", "/content-types/ct1", `{"name":"Posts"}`, errInternal, 500, "Failed to update content type"},
		{"delete ok", "DELETE", "/content-types/ct1", "", nil, 204, ""},
		{"delete missing", "DELETE", "/content-types/ct1", "", gorm.ErrRecordNotFound, 404, "Content type not found"},
		{"delete internal", "DELETE", "/content-types/ct1", "", errInternal, 500, "Failed to delete content type"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := &contentTypeSvcStub{err: tc.err}
			r := kitSend(t, contentTypeKit(svc), tc.method, tc.path, tc.body)
			if tc.contains == "" {
				kitExpect(t, r, tc.want)
			} else {
				kitExpect(t, r, tc.want, tc.contains)
			}
		})
	}
}

func TestContentTypeHandlerPassesInputToService(t *testing.T) {
	svc := &contentTypeSvcStub{}
	app := contentTypeKit(svc)

	kitExpect(t, kitSend(t, app, http.MethodPost, "/content-types", `{"name":"News","description":"d"}`), 201)
	if svc.gotType == nil || svc.gotType.Name != "News" || svc.gotType.Description == nil || *svc.gotType.Description != "d" {
		t.Fatalf("service received %+v", svc.gotType)
	}

	kitExpect(t, kitSend(t, app, http.MethodPatch, "/content-types/abc", `{"description":"only"}`), 200)
	if svc.gotID != "abc" || svc.gotType.Name != "" || svc.gotType.Description == nil || *svc.gotType.Description != "only" {
		t.Fatalf("partial update sent id=%q type=%+v", svc.gotID, svc.gotType)
	}

	kitExpect(t, kitSend(t, app, http.MethodDelete, "/content-types/gone", ""), 204)
	if svc.deleted != "gone" {
		t.Fatalf("deleted %q", svc.deleted)
	}
}
