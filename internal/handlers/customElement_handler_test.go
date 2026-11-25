package handlers

import (
	"context"
	"errors"
	"testing"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"
)

type customElementSvcStub struct {
	services.CustomElementServiceInterface
	err error

	listUser    string
	listPublic  *bool
	pubCategory *string
	pubLimit    int
	pubOffset   int
	getArgs     []string
	created     *models.CustomElement
	updateArgs  []string
	updates     map[string]any
	deleteArgs  []string
	dupArgs     []string
}

func (s *customElementSvcStub) List(_ context.Context, userID string, isPublic *bool) ([]models.CustomElement, error) {
	s.listUser, s.listPublic = userID, isPublic
	return []models.CustomElement{{Id: "e1", UserId: userID}}, s.err
}
func (s *customElementSvcStub) ListPublic(_ context.Context, category *string, limit, offset int) ([]models.CustomElement, error) {
	s.pubCategory, s.pubLimit, s.pubOffset = category, limit, offset
	return []models.CustomElement{}, s.err
}
func (s *customElementSvcStub) GetByID(_ context.Context, id, userID string) (*models.CustomElement, error) {
	s.getArgs = []string{id, userID}
	if s.err != nil {
		return nil, s.err
	}
	return &models.CustomElement{Id: id, UserId: userID}, nil
}
func (s *customElementSvcStub) Create(_ context.Context, e *models.CustomElement) (*models.CustomElement, error) {
	s.created = e
	if s.err != nil {
		return nil, s.err
	}
	return e, nil
}
func (s *customElementSvcStub) Update(_ context.Context, id, userID string, updates map[string]any) (*models.CustomElement, error) {
	s.updateArgs, s.updates = []string{id, userID}, updates
	if s.err != nil {
		return nil, s.err
	}
	return &models.CustomElement{Id: id, UserId: userID}, nil
}
func (s *customElementSvcStub) Delete(_ context.Context, id, userID string) error {
	s.deleteArgs = []string{id, userID}
	return s.err
}
func (s *customElementSvcStub) Duplicate(_ context.Context, id, userID, newName string) (*models.CustomElement, error) {
	s.dupArgs = []string{id, userID, newName}
	if s.err != nil {
		return nil, s.err
	}
	return &models.CustomElement{Id: "copy", UserId: userID, Name: newName}, nil
}

func customElementKit(user string, svc *customElementSvcStub) *fiber.App {
	h := NewCustomElementHandler(svc)
	return newKitApp(user, func(app *fiber.App) {
		app.Get("/custom-elements", h.List)
		app.Post("/custom-elements", h.Create)
		app.Get("/custom-elements/public", h.ListPublic)
		app.Get("/custom-elements/:id", h.GetByID)
		app.Patch("/custom-elements/:id", h.Update)
		app.Delete("/custom-elements/:id", h.Delete)
		app.Post("/custom-elements/:id/duplicate", h.Duplicate)
	})
}

const validElementBody = `{"name":"Card","structure":{"type":"div"},"version":"2.0.0"}`

func TestCustomElementHandlerStatuses(t *testing.T) {
	cases := []struct {
		name         string
		user         string
		method, path string
		body         string
		err          error
		want         int
		contains     string
	}{
		{"list ok", "u1", "GET", "/custom-elements", "", nil, 200, "e1"},
		{"list anonymous", "", "GET", "/custom-elements", "", nil, 401, "Unauthorized"},
		{"list internal", "u1", "GET", "/custom-elements", "", errInternal, 500, "Failed to retrieve custom elements"},
		{"public ok", "", "GET", "/custom-elements/public", "", nil, 200, "[]"},
		{"public internal", "", "GET", "/custom-elements/public", "", errInternal, 500, "Failed to retrieve public custom elements"},
		{"get ok", "u1", "GET", "/custom-elements/e1", "", nil, 200, "e1"},
		{"get anonymous", "", "GET", "/custom-elements/e1", "", nil, 401, "Unauthorized"},
		{"get missing", "u1", "GET", "/custom-elements/e1", "", errors.New("custom element not found"), 404, "Custom element not found"},
		{"get gorm missing", "u1", "GET", "/custom-elements/e1", "", gorm.ErrRecordNotFound, 404, "Custom element not found"},
		{"get internal", "u1", "GET", "/custom-elements/e1", "", errInternal, 500, "Failed to retrieve custom element"},
		{"create ok", "u1", "POST", "/custom-elements", validElementBody, nil, 201, "Card"},
		{"create anonymous", "", "POST", "/custom-elements", validElementBody, nil, 401, "Unauthorized"},
		{"create bad json", "u1", "POST", "/custom-elements", `{`, nil, 0, ""},
		{"create invalid", "u1", "POST", "/custom-elements", `{"structure":{}}`, nil, 0, ""},
		{"create duplicate", "u1", "POST", "/custom-elements", validElementBody, repositories.ErrCustomElementAlreadyExists, 409, "already exists"},
		{"create internal", "u1", "POST", "/custom-elements", validElementBody, errInternal, 500, "Failed to create custom element"},
		{"update ok", "u1", "PATCH", "/custom-elements/e1", `{"name":"New"}`, nil, 200, "e1"},
		{"update anonymous", "", "PATCH", "/custom-elements/e1", `{"name":"New"}`, nil, 401, "Unauthorized"},
		{"update nothing", "u1", "PATCH", "/custom-elements/e1", `{}`, nil, 400, "No valid fields to update"},
		{"update bad json", "u1", "PATCH", "/custom-elements/e1", `[`, nil, 0, ""},
		{"update duplicate name", "u1", "PATCH", "/custom-elements/e1", `{"name":"New"}`, repositories.ErrCustomElementAlreadyExists, 409, "already exists"},
		{"update not owner", "u1", "PATCH", "/custom-elements/e1", `{"name":"New"}`, errors.New("unauthorized to update custom element"), 403, "Unauthorized to update this custom element"},
		{"update missing", "u1", "PATCH", "/custom-elements/e1", `{"name":"New"}`, gorm.ErrRecordNotFound, 404, "Custom element not found"},
		{"update internal", "u1", "PATCH", "/custom-elements/e1", `{"name":"New"}`, errInternal, 500, "Failed to update custom element"},
		{"delete ok", "u1", "DELETE", "/custom-elements/e1", "", nil, 200, "Custom element deleted successfully"},
		{"delete anonymous", "", "DELETE", "/custom-elements/e1", "", nil, 401, "Unauthorized"},
		{"delete not owner", "u1", "DELETE", "/custom-elements/e1", "", errors.New("unauthorized"), 403, "Unauthorized to delete this custom element"},
		{"delete missing", "u1", "DELETE", "/custom-elements/e1", "", gorm.ErrRecordNotFound, 404, "Custom element not found"},
		{"delete internal", "u1", "DELETE", "/custom-elements/e1", "", errInternal, 500, "Failed to delete custom element"},
		{"duplicate ok", "u1", "POST", "/custom-elements/e1/duplicate", `{"newName":"Copy"}`, nil, 201, "Copy"},
		{"duplicate anonymous", "", "POST", "/custom-elements/e1/duplicate", `{"newName":"Copy"}`, nil, 401, "Unauthorized"},
		{"duplicate bad json", "u1", "POST", "/custom-elements/e1/duplicate", `{`, nil, 0, ""},
		{"duplicate invalid", "u1", "POST", "/custom-elements/e1/duplicate", `{}`, nil, 0, ""},
		{"duplicate missing", "u1", "POST", "/custom-elements/e1/duplicate", `{"newName":"Copy"}`, errors.New("custom element not found"), 404, "Custom element not found"},
		{"duplicate internal", "u1", "POST", "/custom-elements/e1/duplicate", `{"newName":"Copy"}`, errInternal, 500, "Failed to duplicate custom element"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := kitSend(t, customElementKit(tc.user, &customElementSvcStub{err: tc.err}), tc.method, tc.path, tc.body)
			if tc.contains == "" {
				kitExpect(t, r, tc.want)
			} else {
				kitExpect(t, r, tc.want, tc.contains)
			}
		})
	}
}

func TestCustomElementListFilters(t *testing.T) {
	for _, tc := range []struct {
		query string
		want  *bool
	}{{"", nil}, {"?isPublic=true", boolPtr(true)}, {"?isPublic=false", boolPtr(false)}, {"?isPublic=maybe", nil}} {
		svc := &customElementSvcStub{}
		kitExpect(t, kitSend(t, customElementKit("u1", svc), "GET", "/custom-elements"+tc.query, ""), 200)
		if svc.listUser != "u1" || (svc.listPublic == nil) != (tc.want == nil) || (tc.want != nil && *svc.listPublic != *tc.want) {
			t.Errorf("%q: user=%q isPublic=%v", tc.query, svc.listUser, svc.listPublic)
		}
	}
}

func TestPublicCustomElementPaging(t *testing.T) {
	for _, tc := range []struct {
		query         string
		limit, offset int
		category      string
	}{
		{"", 20, 0, ""},
		{"?limit=5&offset=10&category=forms", 5, 10, "forms"},
		{"?limit=x&offset=y", 20, 0, ""},
	} {
		svc := &customElementSvcStub{}
		kitExpect(t, kitSend(t, customElementKit("", svc), "GET", "/custom-elements/public"+tc.query, ""), 200)
		if svc.pubLimit != tc.limit || svc.pubOffset != tc.offset {
			t.Errorf("%q: limit=%d offset=%d", tc.query, svc.pubLimit, svc.pubOffset)
		}
		if (tc.category == "") != (svc.pubCategory == nil) || (svc.pubCategory != nil && *svc.pubCategory != tc.category) {
			t.Errorf("%q: category=%v", tc.query, svc.pubCategory)
		}
	}
}

func TestCreateCustomElementOwnsAndDefaults(t *testing.T) {
	svc := &customElementSvcStub{}
	body := `{"name":"Card","structure":{"type":"div"},"userId":"spoofed","isPublic":true,"tags":"a,b"}`
	kitExpect(t, kitSend(t, customElementKit("u1", svc), "POST", "/custom-elements", body), 201)
	e := svc.created
	if e.UserId != "u1" {
		t.Errorf("owner = %q, must come from the session, not the body", e.UserId)
	}
	if e.Id == "" || e.Version != "1.0.0" || !e.IsPublic || e.Name != "Card" || e.Tags == nil || *e.Tags != "a,b" {
		t.Errorf("created %+v", e)
	}
	if string(e.Structure) != `{"type":"div"}` {
		t.Errorf("structure = %s", e.Structure)
	}

	svc = &customElementSvcStub{}
	kitSend(t, customElementKit("u1", svc), "POST", "/custom-elements", validElementBody)
	if svc.created.Version != "2.0.0" {
		t.Errorf("explicit version overwritten: %q", svc.created.Version)
	}
}

func TestUpdateCustomElementBuildsUpdatesFromProvidedFields(t *testing.T) {
	svc := &customElementSvcStub{}
	body := `{"name":"N","typeId":"t","description":"d","category":"c","icon":"i","thumbnail":"th","tags":"x","isPublic":false,"version":"3","structure":{"a":1},"defaultProps":{"b":2},"userId":"evil"}`
	kitExpect(t, kitSend(t, customElementKit("u1", svc), "PATCH", "/custom-elements/e5", body), 200)
	if svc.updateArgs[0] != "e5" || svc.updateArgs[1] != "u1" {
		t.Fatalf("args %v", svc.updateArgs)
	}
	for _, k := range []string{"name", "typeId", "description", "category", "icon", "thumbnail", "tags", "isPublic", "version", "structure", "defaultProps"} {
		if _, ok := svc.updates[k]; !ok {
			t.Errorf("missing update %q in %v", k, svc.updates)
		}
	}
	if len(svc.updates) != 11 {
		t.Errorf("unexpected keys: %v", svc.updates)
	}
	if svc.updates["isPublic"] != false {
		t.Errorf("explicit false must be forwarded, got %v", svc.updates["isPublic"])
	}
}

func TestCustomElementActionsUseSessionUser(t *testing.T) {
	svc := &customElementSvcStub{}
	app := customElementKit("u7", svc)
	kitSend(t, app, "GET", "/custom-elements/e1", "")
	kitSend(t, app, "DELETE", "/custom-elements/e2", "")
	kitSend(t, app, "POST", "/custom-elements/e3/duplicate", `{"newName":"Twin"}`)
	if svc.getArgs[1] != "u7" || svc.deleteArgs[0] != "e2" || svc.deleteArgs[1] != "u7" {
		t.Errorf("get=%v delete=%v", svc.getArgs, svc.deleteArgs)
	}
	if len(svc.dupArgs) != 3 || svc.dupArgs[0] != "e3" || svc.dupArgs[1] != "u7" || svc.dupArgs[2] != "Twin" {
		t.Errorf("duplicate args %v", svc.dupArgs)
	}
}
