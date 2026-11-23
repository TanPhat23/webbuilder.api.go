package handlers

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"

	"my-go-app/internal/models"
	"my-go-app/internal/services"
)

type eventWorkflowSvcStub struct {
	services.EventWorkflowServiceInterface
	err      error
	nilOnGet bool
	// emptyCreate makes CreateEventWorkflow return a workflow without JSON fields.
	emptyCreate bool

	created    *models.EventWorkflow
	createUser string
	updatedID  string
	updated    *models.EventWorkflow
	enabledID  string
	enabled    *bool
	deletedID  string
	fProject   string
	fUser      string
	fEnabled   *bool
	fSearch    string
}

func (s *eventWorkflowSvcStub) CreateEventWorkflow(_ context.Context, w *models.EventWorkflow, userID string) (*models.EventWorkflow, error) {
	s.created, s.createUser = w, strings.Clone(userID)
	if s.err != nil {
		return nil, s.err
	}
	w.Id = "wf-new"
	if s.emptyCreate {
		w.CanvasData, w.Handlers = nil, nil
	}
	return w, nil
}
func (s *eventWorkflowSvcStub) GetEventWorkflowByID(_ context.Context, id string) (*models.EventWorkflow, error) {
	if s.err != nil {
		return nil, s.err
	}
	if s.nilOnGet {
		return nil, nil
	}
	return &models.EventWorkflow{Id: id, Name: "existing", Enabled: true}, nil
}
func (s *eventWorkflowSvcStub) GetEventWorkflowsWithFilters(_ context.Context, projectID, userID string, enabled *bool, search string) ([]models.EventWorkflow, error) {
	s.fProject, s.fUser, s.fEnabled, s.fSearch = strings.Clone(projectID), strings.Clone(userID), enabled, strings.Clone(search)
	if s.err != nil {
		return nil, s.err
	}
	return []models.EventWorkflow{{Id: "w1"}, {Id: "w2"}, {Id: "w3"}}, nil
}
func (s *eventWorkflowSvcStub) UpdateEventWorkflow(_ context.Context, id string, w *models.EventWorkflow) (*models.EventWorkflow, error) {
	s.updatedID, s.updated = strings.Clone(id), w
	if s.err != nil && !s.nilOnGet {
		// GetEventWorkflowByID shares err; tests that reach here use updateErr instead.
		return w, nil
	}
	w.Id = id
	return w, nil
}
func (s *eventWorkflowSvcStub) UpdateEventWorkflowEnabled(_ context.Context, id string, enabled bool) error {
	s.enabledID, s.enabled = strings.Clone(id), &enabled
	return s.err
}
func (s *eventWorkflowSvcStub) DeleteEventWorkflow(_ context.Context, id string) error {
	s.deletedID = strings.Clone(id)
	return s.err
}

// eventWorkflowUpdateFailStub lets the lookup succeed and fails the update.
type eventWorkflowUpdateFailStub struct {
	eventWorkflowSvcStub
	updateErr error
}

func (s *eventWorkflowUpdateFailStub) UpdateEventWorkflow(context.Context, string, *models.EventWorkflow) (*models.EventWorkflow, error) {
	return nil, s.updateErr
}

func eventWorkflowKit(userID string, svc services.EventWorkflowServiceInterface, links services.ElementEventWorkflowServiceInterface) *fiber.App {
	h := NewEventWorkflowHandler(svc, links)
	return newKitApp(userID, func(app *fiber.App) {
		app.Post("/workflows", h.CreateEventWorkflow)
		app.Get("/projects/:projectid/workflows", h.GetEventWorkflowsByProject)
		app.Get("/workflows/:id/elements", h.GetEventWorkflowElements)
		app.Patch("/workflows/:id/enabled", h.UpdateEventWorkflowEnabled)
		app.Get("/workflows/:id", h.GetEventWorkflowByID)
		app.Put("/workflows/:id", h.UpdateEventWorkflow)
		app.Delete("/workflows/:id", h.DeleteEventWorkflow)
		app.Get("/noparam/get", h.GetEventWorkflowByID)
		app.Get("/noparam/project", h.GetEventWorkflowsByProject)
		app.Put("/noparam/update", h.UpdateEventWorkflow)
		app.Patch("/noparam/enabled", h.UpdateEventWorkflowEnabled)
		app.Delete("/noparam/delete", h.DeleteEventWorkflow)
		app.Get("/noparam/elements", h.GetEventWorkflowElements)
	})
}

func eventWorkflowDefaultKit(userID string, svc services.EventWorkflowServiceInterface, linkErr error) *fiber.App {
	return eventWorkflowKit(userID, svc, &elementEventWorkflowSvcStub{err: linkErr})
}

func TestEventWorkflowHandlerStatuses(t *testing.T) {
	nf := gorm.ErrRecordNotFound
	cases := []struct {
		name         string
		user         string
		method, path string
		body         string
		err          error
		nilOnGet     bool
		want         int
		contains     string
	}{
		{"create ok", "u1", "POST", "/workflows", `{"projectId":"p1","name":"Flow"}`, nil, false, 201, "wf-new"},
		{"create unauthenticated", "", "POST", "/workflows", `{"projectId":"p1","name":"Flow"}`, nil, false, 401, "Unauthorized"},
		{"create bad json", "u1", "POST", "/workflows", `{`, nil, false, 0, ""},
		{"create missing name", "u1", "POST", "/workflows", `{"projectId":"p1"}`, nil, false, 0, ""},
		{"create project missing", "u1", "POST", "/workflows", `{"projectId":"p1","name":"Flow"}`, nf, false, 404, "Project not found"},
		{"create internal", "u1", "POST", "/workflows", `{"projectId":"p1","name":"Flow"}`, errInternal, false, 500, "Failed to create event workflow"},

		{"get ok", "", "GET", "/workflows/w1", "", nil, false, 200, `"id":"w1"`},
		{"get missing param", "", "GET", "/noparam/get", "", nil, false, 400, "id is required"},
		{"get not found", "", "GET", "/workflows/w1", "", nf, false, 404, "Event workflow not found"},
		{"get nil result", "", "GET", "/workflows/w1", "", nil, true, 404, "Event workflow not found"},
		{"get internal", "", "GET", "/workflows/w1", "", errInternal, false, 500, "Failed to retrieve event workflow"},

		{"project list ok", "u1", "GET", "/projects/p1/workflows", "", nil, false, 200, `"count":3`},
		{"project list unauthenticated", "", "GET", "/projects/p1/workflows", "", nil, false, 401, "Unauthorized"},
		{"project list missing param", "u1", "GET", "/noparam/project", "", nil, false, 400, "projectid is required"},
		{"project list not found", "u1", "GET", "/projects/p1/workflows", "", nf, false, 404, "Project not found"},
		{"project list internal", "u1", "GET", "/projects/p1/workflows", "", errInternal, false, 500, "Failed to retrieve event workflows"},

		{"update ok", "", "PUT", "/workflows/w1", `{"name":"Renamed"}`, nil, false, 200, "Renamed"},
		{"update missing param", "", "PUT", "/noparam/update", `{"name":"x"}`, nil, false, 400, "id is required"},
		{"update lookup not found", "", "PUT", "/workflows/w1", `{"name":"x"}`, nf, false, 404, "Event workflow not found"},
		{"update lookup internal", "", "PUT", "/workflows/w1", `{"name":"x"}`, errInternal, false, 500, "Failed to retrieve event workflow"},
		{"update nil lookup", "", "PUT", "/workflows/w1", `{"name":"x"}`, nil, true, 404, "Event workflow not found"},
		{"update bad json", "", "PUT", "/workflows/w1", `[`, nil, false, 0, ""},
		{"update empty name", "", "PUT", "/workflows/w1", `{"name":""}`, nil, false, 0, ""},

		{"enabled ok", "", "PATCH", "/workflows/w1/enabled", `{"enabled":false}`, nil, false, 200, `"enabled":false`},
		{"enabled missing param", "", "PATCH", "/noparam/enabled", `{"enabled":true}`, nil, false, 400, "id is required"},
		{"enabled bad json", "", "PATCH", "/workflows/w1/enabled", `{`, nil, false, 0, ""},
		{"enabled missing value", "", "PATCH", "/workflows/w1/enabled", `{}`, nil, false, 0, ""},
		{"enabled not found", "", "PATCH", "/workflows/w1/enabled", `{"enabled":true}`, nf, false, 404, "Event workflow not found"},
		{"enabled internal", "", "PATCH", "/workflows/w1/enabled", `{"enabled":true}`, errInternal, false, 500, "Failed to update event workflow status"},

		{"delete ok", "", "DELETE", "/workflows/w1", "", nil, false, 204, ""},
		{"delete missing param", "", "DELETE", "/noparam/delete", "", nil, false, 400, "id is required"},
		{"delete not found", "", "DELETE", "/workflows/w1", "", nf, false, 404, "Event workflow not found"},
		{"delete internal", "", "DELETE", "/workflows/w1", "", errInternal, false, 500, "Failed to delete event workflow"},

		{"elements ok", "", "GET", "/workflows/w1/elements", "", nil, false, 200, `"count":2`},
		{"elements missing param", "", "GET", "/noparam/elements", "", nil, false, 400, "id is required"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := &eventWorkflowSvcStub{err: tc.err, nilOnGet: tc.nilOnGet}
			r := kitSend(t, eventWorkflowDefaultKit(tc.user, svc, nil), tc.method, tc.path, tc.body)
			if tc.contains == "" {
				kitExpect(t, r, tc.want)
			} else {
				kitExpect(t, r, tc.want, tc.contains)
			}
		})
	}
}

func TestEventWorkflowElementsInternalError(t *testing.T) {
	r := kitSend(t, eventWorkflowDefaultKit("", &eventWorkflowSvcStub{}, errInternal), "GET", "/workflows/w1/elements", "")
	kitExpect(t, r, 500, "Failed to retrieve elements")
}

func TestEventWorkflowUpdateFailureAfterLookup(t *testing.T) {
	for _, tc := range []struct {
		err      error
		want     int
		contains string
	}{
		{gorm.ErrRecordNotFound, 404, "Event workflow not found"},
		{errInternal, 500, "Failed to update event workflow"},
	} {
		svc := &eventWorkflowUpdateFailStub{updateErr: tc.err}
		kitExpect(t, kitSend(t, eventWorkflowDefaultKit("", svc, nil), "PUT", "/workflows/w1", `{"name":"x"}`), tc.want, tc.contains)
	}
}

func TestEventWorkflowCreateMapsRequestAndDefaultsJSON(t *testing.T) {
	svc := &eventWorkflowSvcStub{emptyCreate: true}
	r := kitSend(t, eventWorkflowDefaultKit("u5", svc, nil), "POST", "/workflows",
		`{"projectId":"p2","name":"Flow","description":"desc","canvasData":{"nodes":[]},"handlers":[{"a":1}],"enabled":true}`)
	kitExpect(t, r, 201)
	c := svc.created
	if svc.createUser != "u5" || c == nil || c.ProjectId != "p2" || c.Name != "Flow" || c.Description == nil || *c.Description != "desc" || !c.Enabled {
		t.Fatalf("user=%q created=%+v", svc.createUser, c)
	}
	if string(c.CanvasData) != "{}" && string(c.CanvasData) != `{"nodes":[]}` {
		t.Fatalf("canvas %s", c.CanvasData)
	}
	// The service returned no JSON fields, so the handler must fill in defaults.
	body := r.object(t)
	if string(mustMarshal(t, body["canvasData"])) != "{}" || string(mustMarshal(t, body["handlers"])) != "[]" {
		t.Fatalf("defaults not applied: %s", r.body)
	}
}

func TestEventWorkflowCreateKeepsProvidedJSONAndDisabledByDefault(t *testing.T) {
	svc := &eventWorkflowSvcStub{}
	r := kitSend(t, eventWorkflowDefaultKit("u5", svc, nil), "POST", "/workflows",
		`{"projectId":"p2","name":"Flow","canvasData":{"nodes":[1]},"handlers":[{"a":1}]}`)
	kitExpect(t, r, 201)
	if svc.created.Enabled {
		t.Fatal("workflows must be created disabled unless requested")
	}
	body := r.object(t)
	if string(mustMarshal(t, body["canvasData"])) != `{"nodes":[1]}` || string(mustMarshal(t, body["handlers"])) != `[{"a":1}]` {
		t.Fatalf("body %s", r.body)
	}
}

func mustMarshal(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestEventWorkflowProjectListParsesFilters(t *testing.T) {
	cases := []struct {
		query   string
		enabled *bool
		search  string
	}{
		{"", nil, ""},
		{"?enabled=true", boolPtr(true), ""},
		{"?enabled=false&search=login", boolPtr(false), "login"},
		{"?enabled=banana", nil, ""},
	}
	for _, c := range cases {
		svc := &eventWorkflowSvcStub{}
		r := kitSend(t, eventWorkflowDefaultKit("u2", svc, nil), "GET", "/projects/p4/workflows"+c.query, "")
		kitExpect(t, r, 200, `"count":3`)
		if svc.fProject != "p4" || svc.fUser != "u2" || svc.fSearch != c.search {
			t.Errorf("query %q: project=%q user=%q search=%q", c.query, svc.fProject, svc.fUser, svc.fSearch)
		}
		if (svc.fEnabled == nil) != (c.enabled == nil) || (c.enabled != nil && *svc.fEnabled != *c.enabled) {
			t.Errorf("query %q: enabled=%v want %v", c.query, svc.fEnabled, c.enabled)
		}
	}
}

func TestEventWorkflowUpdateMapsBody(t *testing.T) {
	svc := &eventWorkflowSvcStub{}
	r := kitSend(t, eventWorkflowDefaultKit("", svc, nil), "PUT", "/workflows/w8",
		`{"name":"New","description":"d","canvasData":{"k":1},"handlers":[],"enabled":true}`)
	kitExpect(t, r, 200, `"name":"New"`)
	u := svc.updated
	if svc.updatedID != "w8" || u == nil || u.Name != "New" || u.Description == nil || *u.Description != "d" || !u.Enabled || string(u.CanvasData) != `{"k":1}` {
		t.Fatalf("id=%q updated=%+v", svc.updatedID, u)
	}
	// An omitted enabled value preserves the existing setting.
	kitExpect(t, kitSend(t, eventWorkflowDefaultKit("", svc, nil), "PUT", "/workflows/w8", `{"name":"New"}`), 200)
	if !svc.updated.Enabled {
		t.Fatal("omitted enabled must preserve the existing value")
	}
}

func TestEventWorkflowEnabledAndDeleteForwardIDs(t *testing.T) {
	svc := &eventWorkflowSvcStub{}
	app := eventWorkflowDefaultKit("", svc, nil)
	kitExpect(t, kitSend(t, app, "PATCH", "/workflows/w2/enabled", `{"enabled":true}`), 200, `"enabled":true`)
	if svc.enabledID != "w2" || svc.enabled == nil || !*svc.enabled {
		t.Fatalf("enabled id=%q value=%v", svc.enabledID, svc.enabled)
	}
	kitExpect(t, kitSend(t, app, "DELETE", "/workflows/w2", ""), 204)
	if svc.deletedID != "w2" {
		t.Fatalf("deleted %q", svc.deletedID)
	}
}

func TestEventWorkflowEmptyUpdates(t *testing.T) {
	svc := &eventWorkflowSvcStub{}
	app := eventWorkflowDefaultKit("", svc, nil)
	kitExpect(t, kitSend(t, app, "PUT", "/workflows/w1", `{}`), 200)
	if svc.updated == nil || svc.updated.Name != "existing" || !svc.updated.Enabled {
		t.Fatalf("update = %+v", svc.updated)
	}
	kitExpect(t, kitSend(t, app, "PATCH", "/workflows/w1/enabled", `{}`), 400)
}
