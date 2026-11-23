package handlers

import (
	"context"
	"testing"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"

	"my-go-app/internal/models"
	"my-go-app/internal/services"
)

type contentFieldSvcStub struct {
	services.ContentFieldServiceInterface
	err       error
	created   *models.ContentField
	updatedID string
	updates   map[string]any
	deleted   string
}

func (s *contentFieldSvcStub) GetContentFieldsByContentType(_ context.Context, typeID string) ([]models.ContentField, error) {
	return []models.ContentField{{Id: "f1", ContentTypeId: typeID, Name: "title"}}, s.err
}
func (s *contentFieldSvcStub) GetContentFieldByID(_ context.Context, id string) (*models.ContentField, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &models.ContentField{Id: id, Name: "title"}, nil
}
func (s *contentFieldSvcStub) CreateContentField(_ context.Context, f *models.ContentField) (*models.ContentField, error) {
	s.created = f
	if s.err != nil {
		return nil, s.err
	}
	f.Id = "f-new"
	return f, nil
}
func (s *contentFieldSvcStub) UpdateContentField(_ context.Context, id string, updates map[string]any) (*models.ContentField, error) {
	s.updatedID, s.updates = id, updates
	if s.err != nil {
		return nil, s.err
	}
	return &models.ContentField{Id: id, Name: "renamed"}, nil
}
func (s *contentFieldSvcStub) DeleteContentField(_ context.Context, id string) error {
	s.deleted = id
	return s.err
}

func contentFieldKit(svc *contentFieldSvcStub) *fiber.App {
	h := NewContentFieldHandler(svc)
	return newKitApp("", func(app *fiber.App) {
		app.Get("/types/:contentTypeId/fields", h.GetContentFieldsByContentType)
		app.Post("/types/:contentTypeId/fields", h.CreateContentField)
		app.Get("/fields/:fieldId", h.GetContentFieldByID)
		app.Patch("/fields/:fieldId", h.UpdateContentField)
		app.Delete("/fields/:fieldId", h.DeleteContentField)
	})
}

func TestContentFieldHandlerStatuses(t *testing.T) {
	cases := []struct {
		name         string
		method, path string
		body         string
		err          error
		want         int
		contains     string
	}{
		{"list ok", "GET", "/types/t1/fields", "", nil, 200, "\"contentTypeId\":\"t1\""},
		{"list internal", "GET", "/types/t1/fields", "", errInternal, 500, "Failed to retrieve content fields"},
		{"get ok", "GET", "/fields/f1", "", nil, 200, "f1"},
		{"get missing", "GET", "/fields/f1", "", gorm.ErrRecordNotFound, 404, "Content field not found"},
		{"get internal", "GET", "/fields/f1", "", errInternal, 500, "Failed to retrieve content field"},
		{"create ok", "POST", "/types/t1/fields", `{"name":"title","type":"text","required":true}`, nil, 201, "f-new"},
		{"create bad json", "POST", "/types/t1/fields", `{`, nil, 0, ""},
		{"create invalid", "POST", "/types/t1/fields", `{"name":"title"}`, nil, 0, ""},
		{"create internal", "POST", "/types/t1/fields", `{"name":"title","type":"text"}`, errInternal, 500, "Failed to create content field"},
		{"update ok", "PATCH", "/fields/f1", `{"name":"renamed"}`, nil, 200, "renamed"},
		{"update empty", "PATCH", "/fields/f1", `{}`, nil, 400, "No valid fields to update"},
		{"update bad json", "PATCH", "/fields/f1", `[`, nil, 0, ""},
		{"update missing", "PATCH", "/fields/f1", `{"name":"x"}`, gorm.ErrRecordNotFound, 404, "Content field not found"},
		{"update internal", "PATCH", "/fields/f1", `{"name":"x"}`, errInternal, 500, "Failed to update content field"},
		{"delete ok", "DELETE", "/fields/f1", "", nil, 204, ""},
		{"delete missing", "DELETE", "/fields/f1", "", gorm.ErrRecordNotFound, 404, "Content field not found"},
		{"delete internal", "DELETE", "/fields/f1", "", errInternal, 500, "Failed to delete content field"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := kitSend(t, contentFieldKit(&contentFieldSvcStub{err: tc.err}), tc.method, tc.path, tc.body)
			if tc.contains == "" {
				kitExpect(t, r, tc.want)
			} else {
				kitExpect(t, r, tc.want, tc.contains)
			}
		})
	}
}

func TestContentFieldCreateBindsParentFromPath(t *testing.T) {
	svc := &contentFieldSvcStub{}
	kitExpect(t, kitSend(t, contentFieldKit(svc), "POST", "/types/type-9/fields", `{"name":"body","type":"text","required":true,"contentTypeId":"spoofed"}`), 201)
	if svc.created == nil || svc.created.ContentTypeId != "type-9" || svc.created.Name != "body" || !svc.created.Required {
		t.Fatalf("created %+v", svc.created)
	}
}

func TestContentFieldUpdateOnlyForwardsAllowedColumns(t *testing.T) {
	svc := &contentFieldSvcStub{}
	kitExpect(t, kitSend(t, contentFieldKit(svc), "PATCH", "/fields/f9", `{"type":"number","required":false,"contentTypeId":"other"}`), 200)
	if svc.updatedID != "f9" || len(svc.updates) != 2 || svc.updates["Type"] != "number" || svc.updates["Required"] != false {
		t.Fatalf("id=%q updates=%v", svc.updatedID, svc.updates)
	}
}
