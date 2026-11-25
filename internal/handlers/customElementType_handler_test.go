package handlers

import (
	"context"
	"testing"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"

	"my-go-app/internal/models"
	"my-go-app/internal/services"
)

type customElementTypeSvcStub struct {
	services.CustomElementTypeServiceInterface
	err       error
	created   *models.CustomElementType
	updatedID string
	updates   map[string]any
	deleted   string
}

func (s *customElementTypeSvcStub) List(context.Context) ([]models.CustomElementType, error) {
	return []models.CustomElementType{{Id: "t1", Name: "Form"}}, s.err
}
func (s *customElementTypeSvcStub) GetByID(_ context.Context, id string) (*models.CustomElementType, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &models.CustomElementType{Id: id, Name: "Form"}, nil
}
func (s *customElementTypeSvcStub) Create(_ context.Context, t *models.CustomElementType) (*models.CustomElementType, error) {
	s.created = t
	if s.err != nil {
		return nil, s.err
	}
	return t, nil
}
func (s *customElementTypeSvcStub) Update(_ context.Context, id string, updates map[string]any) (*models.CustomElementType, error) {
	s.updatedID, s.updates = id, updates
	if s.err != nil {
		return nil, s.err
	}
	return &models.CustomElementType{Id: id, Name: "Renamed"}, nil
}
func (s *customElementTypeSvcStub) Delete(_ context.Context, id string) error {
	s.deleted = id
	return s.err
}

func customElementTypeKit(svc *customElementTypeSvcStub) *fiber.App {
	h := NewCustomElementTypeHandler(svc)
	return newKitApp("", func(app *fiber.App) {
		app.Get("/types", h.List)
		app.Post("/types", h.Create)
		app.Get("/types/:id", h.GetByID)
		app.Patch("/types/:id", h.Update)
		app.Delete("/types/:id", h.Delete)
	})
}

func TestCustomElementTypeHandlerStatuses(t *testing.T) {
	cases := []struct {
		name         string
		method, path string
		body         string
		err          error
		want         int
		contains     string
	}{
		{"list ok", "GET", "/types", "", nil, 200, "Form"},
		{"list internal", "GET", "/types", "", errInternal, 500, "Failed to retrieve custom element types"},
		{"get ok", "GET", "/types/t1", "", nil, 200, "t1"},
		{"get missing", "GET", "/types/t1", "", gorm.ErrRecordNotFound, 404, "Custom element type not found"},
		{"get internal", "GET", "/types/t1", "", errInternal, 500, "Failed to retrieve custom element type"},
		{"create ok", "POST", "/types", `{"name":"Form","icon":"i"}`, nil, 201, "Form"},
		{"create bad json", "POST", "/types", `{`, nil, 0, ""},
		{"create invalid", "POST", "/types", `{"icon":"i"}`, nil, 0, ""},
		{"create internal", "POST", "/types", `{"name":"Form"}`, errInternal, 500, "Failed to create custom element type"},
		{"update ok", "PATCH", "/types/t1", `{"name":"Renamed"}`, nil, 200, "Renamed"},
		{"update nothing", "PATCH", "/types/t1", `{}`, nil, 400, "No valid fields to update"},
		{"update bad json", "PATCH", "/types/t1", `[`, nil, 0, ""},
		{"update missing", "PATCH", "/types/t1", `{"name":"x"}`, gorm.ErrRecordNotFound, 404, "Custom element type not found"},
		{"update internal", "PATCH", "/types/t1", `{"name":"x"}`, errInternal, 500, "Failed to update custom element type"},
		{"delete ok", "DELETE", "/types/t1", "", nil, 200, "Custom element type deleted successfully"},
		{"delete missing", "DELETE", "/types/t1", "", gorm.ErrRecordNotFound, 404, "Custom element type not found"},
		{"delete internal", "DELETE", "/types/t1", "", errInternal, 500, "Failed to delete custom element type"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := kitSend(t, customElementTypeKit(&customElementTypeSvcStub{err: tc.err}), tc.method, tc.path, tc.body)
			if tc.contains == "" {
				kitExpect(t, r, tc.want)
			} else {
				kitExpect(t, r, tc.want, tc.contains)
			}
		})
	}
}

func TestCreateCustomElementTypeGeneratesID(t *testing.T) {
	svc := &customElementTypeSvcStub{}
	kitExpect(t, kitSend(t, customElementTypeKit(svc), "POST", "/types", `{"id":"client-id","name":"Form","description":"d","category":"c","icon":"i"}`), 201)
	c := svc.created
	if c.Id == "" || c.Id == "client-id" || c.Name != "Form" || *c.Description != "d" || *c.Category != "c" || *c.Icon != "i" {
		t.Fatalf("created %+v", c)
	}
}

func TestUpdateCustomElementTypeForwardsOnlyProvidedFields(t *testing.T) {
	svc := &customElementTypeSvcStub{}
	kitExpect(t, kitSend(t, customElementTypeKit(svc), "PATCH", "/types/t9", `{"icon":"star","category":"layout"}`), 200)
	if svc.updatedID != "t9" || len(svc.updates) != 2 || svc.updates["icon"] != "star" || svc.updates["category"] != "layout" {
		t.Fatalf("id=%q updates=%v", svc.updatedID, svc.updates)
	}
}
