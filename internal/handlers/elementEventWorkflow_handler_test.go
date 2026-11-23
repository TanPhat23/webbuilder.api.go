package handlers

import (
	"context"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"
)

type elementEventWorkflowSvcStub struct {
	services.ElementEventWorkflowServiceInterface
	err error
	// nilOnGet makes GetElementEventWorkflowByID answer (nil, nil).
	nilOnGet bool

	created    *models.ElementEventWorkflow
	updatedID  string
	updated    *models.ElementEventWorkflow
	deletedID  string
	calls      []string
	byPage     string
	delElement string
	delFlow    string
}

func (s *elementEventWorkflowSvcStub) CreateElementEventWorkflow(_ context.Context, e *models.ElementEventWorkflow) (*models.ElementEventWorkflow, error) {
	s.created = e
	if s.err != nil {
		return nil, s.err
	}
	e.Id = "eew-new"
	return e, nil
}
func (s *elementEventWorkflowSvcStub) GetElementEventWorkflowByID(_ context.Context, id string) (*models.ElementEventWorkflow, error) {
	if s.err != nil {
		return nil, s.err
	}
	if s.nilOnGet {
		return nil, nil
	}
	return &models.ElementEventWorkflow{Id: id, ElementId: "e1", WorkflowId: "w1", EventName: "onClick"}, nil
}
func (s *elementEventWorkflowSvcStub) list(call string) ([]models.ElementEventWorkflow, error) {
	s.calls = append(s.calls, call)
	if s.err != nil {
		return nil, s.err
	}
	return []models.ElementEventWorkflow{{Id: "l1"}, {Id: "l2"}}, nil
}
func (s *elementEventWorkflowSvcStub) GetElementEventWorkflowsByPageID(_ context.Context, p string) ([]models.ElementEventWorkflow, error) {
	s.byPage = strings.Clone(p)
	return s.list("page")
}
func (s *elementEventWorkflowSvcStub) GetElementEventWorkflowsByElementID(_ context.Context, id string) ([]models.ElementEventWorkflow, error) {
	return s.list("element:" + strings.Clone(id))
}
func (s *elementEventWorkflowSvcStub) GetElementEventWorkflowsByWorkflowID(_ context.Context, id string) ([]models.ElementEventWorkflow, error) {
	return s.list("workflow:" + strings.Clone(id))
}
func (s *elementEventWorkflowSvcStub) GetElementEventWorkflowsByFilters(_ context.Context, e, w, n string) ([]models.ElementEventWorkflow, error) {
	return s.list(strings.Clone("filters:" + e + "|" + w + "|" + n))
}
func (s *elementEventWorkflowSvcStub) GetAllElementEventWorkflows(context.Context) ([]models.ElementEventWorkflow, error) {
	return s.list("all")
}
func (s *elementEventWorkflowSvcStub) UpdateElementEventWorkflow(_ context.Context, id string, e *models.ElementEventWorkflow) (*models.ElementEventWorkflow, error) {
	s.updatedID, s.updated = strings.Clone(id), e
	return e, nil
}
func (s *elementEventWorkflowSvcStub) DeleteElementEventWorkflow(_ context.Context, id string) error {
	s.deletedID = strings.Clone(id)
	return nil
}
func (s *elementEventWorkflowSvcStub) DeleteElementEventWorkflowsByElementID(_ context.Context, id string) error {
	s.delElement = strings.Clone(id)
	return s.err
}
func (s *elementEventWorkflowSvcStub) DeleteElementEventWorkflowsByWorkflowID(_ context.Context, id string) error {
	s.delFlow = strings.Clone(id)
	return s.err
}

// elementEventWorkflowFailingStub fails the mutating calls only, so the
// handlers' second service call can be exercised after a successful lookup.
type elementEventWorkflowFailingStub struct {
	elementEventWorkflowSvcStub
	updateErr, deleteErr error
}

func (s *elementEventWorkflowFailingStub) UpdateElementEventWorkflow(context.Context, string, *models.ElementEventWorkflow) (*models.ElementEventWorkflow, error) {
	return nil, s.updateErr
}
func (s *elementEventWorkflowFailingStub) DeleteElementEventWorkflow(context.Context, string) error {
	return s.deleteErr
}

type elementEventWorkflowPageStub struct {
	services.PageServiceInterface
	err error
}

func (s *elementEventWorkflowPageStub) GetPageByID(_ context.Context, id string) (*models.Page, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &models.Page{Id: id, ProjectId: "proj-1"}, nil
}

type elementEventWorkflowProjectStub struct {
	services.ProjectServiceInterface
	err        error
	gotProject string
	gotUser    string
}

func (s *elementEventWorkflowProjectStub) GetProjectWithAccess(_ context.Context, projectID, userID string) (*models.Project, error) {
	s.gotProject, s.gotUser = strings.Clone(projectID), strings.Clone(userID)
	if s.err != nil {
		return nil, s.err
	}
	return &models.Project{ID: projectID}, nil
}

func elementEventWorkflowKit(userID string, svc services.ElementEventWorkflowServiceInterface, pages services.PageServiceInterface, projects services.ProjectServiceInterface) *fiber.App {
	h := NewElementEventWorkflowHandler(svc, projects, pages)
	return newKitApp(userID, func(app *fiber.App) {
		app.Post("/eew", h.CreateElementEventWorkflow)
		app.Get("/eew", h.GetElementEventWorkflows)
		app.Get("/eew/page/:pageId", h.GetElementEventWorkflowsByPageId)
		app.Delete("/eew/element/:elementId", h.DeleteElementEventWorkflowsByElement)
		app.Delete("/eew/workflow/:workflowId", h.DeleteElementEventWorkflowsByWorkflow)
		app.Get("/eew/:id", h.GetElementEventWorkflowByID)
		app.Patch("/eew/:id", h.UpdateElementEventWorkflow)
		app.Delete("/eew/:id", h.DeleteElementEventWorkflow)
		app.Get("/noparam/get", h.GetElementEventWorkflowByID)
		app.Get("/noparam/page", h.GetElementEventWorkflowsByPageId)
		app.Patch("/noparam/update", h.UpdateElementEventWorkflow)
		app.Delete("/noparam/delete", h.DeleteElementEventWorkflow)
		app.Delete("/noparam/element", h.DeleteElementEventWorkflowsByElement)
		app.Delete("/noparam/workflow", h.DeleteElementEventWorkflowsByWorkflow)
	})
}

func elementEventWorkflowDefaultKit(userID string, svc services.ElementEventWorkflowServiceInterface) *fiber.App {
	return elementEventWorkflowKit(userID, svc, &elementEventWorkflowPageStub{}, &elementEventWorkflowProjectStub{})
}

func TestElementEventWorkflowHandlerStatuses(t *testing.T) {
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
		{"create ok", "u1", "POST", "/eew", `{"elementId":"e1","workflowId":"w1","eventName":"onClick"}`, nil, false, 201, "eew-new"},
		{"create unauthenticated", "", "POST", "/eew", `{"elementId":"e1","workflowId":"w1","eventName":"onClick"}`, nil, false, 401, "Unauthorized"},
		{"create bad json", "u1", "POST", "/eew", `{`, nil, false, 0, ""},
		{"create missing event", "u1", "POST", "/eew", `{"elementId":"e1","workflowId":"w1"}`, nil, false, 0, ""},
		{"create internal", "u1", "POST", "/eew", `{"elementId":"e1","workflowId":"w1","eventName":"x"}`, errInternal, false, 500, "Failed to create element event workflow"},

		{"get ok", "u1", "GET", "/eew/l1", "", nil, false, 200, `"id":"l1"`},
		{"get unauthenticated", "", "GET", "/eew/l1", "", nil, false, 401, "Unauthorized"},
		{"get missing param", "u1", "GET", "/noparam/get", "", nil, false, 400, "id is required"},
		{"get not found", "u1", "GET", "/eew/l1", "", nf, false, 404, "Element event workflow not found"},
		{"get nil result", "u1", "GET", "/eew/l1", "", nil, true, 404, "Element event workflow not found"},
		{"get internal", "u1", "GET", "/eew/l1", "", errInternal, false, 500, "Failed to retrieve element event workflow"},

		{"list all", "u1", "GET", "/eew", "", nil, false, 200, `"count":2`},
		{"list unauthenticated", "", "GET", "/eew", "", nil, false, 401, "Unauthorized"},
		{"list by element", "u1", "GET", "/eew?elementId=e1", "", nil, false, 200, `"count":2`},
		{"list by workflow", "u1", "GET", "/eew?workflowId=w1", "", nil, false, 200, `"count":2`},
		{"list by filters", "u1", "GET", "/eew?elementId=e1&eventName=onClick", "", nil, false, 200, `"count":2`},
		{"list by event name only", "u1", "GET", "/eew?eventName=onClick", "", nil, false, 200, `"count":2`},
		{"list by element internal", "u1", "GET", "/eew?elementId=e1", "", errInternal, false, 500, "Failed to retrieve element event workflows"},
		{"list by workflow internal", "u1", "GET", "/eew?workflowId=w1", "", errInternal, false, 500, "Failed to retrieve element event workflows"},
		{"list by filters internal", "u1", "GET", "/eew?eventName=x", "", errInternal, false, 500, "Failed to retrieve element event workflows"},
		{"list all internal", "u1", "GET", "/eew", "", errInternal, false, 500, "Failed to retrieve element event workflows"},

		{"page ok", "u1", "GET", "/eew/page/pg1", "", nil, false, 200, `"count":2`},
		{"page unauthenticated", "", "GET", "/eew/page/pg1", "", nil, false, 401, "Unauthorized"},
		{"page missing param", "u1", "GET", "/noparam/page", "", nil, false, 400, "pageId is required"},
		{"page workflows internal", "u1", "GET", "/eew/page/pg1", "", errInternal, false, 500, "Failed to retrieve element event workflows"},

		{"update ok", "u1", "PATCH", "/eew/l1", `{"eventName":"onHover"}`, nil, false, 200, "onHover"},
		{"update unauthenticated", "", "PATCH", "/eew/l1", `{"eventName":"onHover"}`, nil, false, 401, "Unauthorized"},
		{"update missing param", "u1", "PATCH", "/noparam/update", `{"eventName":"x"}`, nil, false, 400, "id is required"},
		{"update lookup not found", "u1", "PATCH", "/eew/l1", `{"eventName":"x"}`, nf, false, 404, "Element event workflow not found"},
		{"update lookup internal", "u1", "PATCH", "/eew/l1", `{"eventName":"x"}`, errInternal, false, 500, "Failed to retrieve element event workflow"},
		{"update nil lookup", "u1", "PATCH", "/eew/l1", `{"eventName":"x"}`, nil, true, 404, "Element event workflow not found"},
		{"update bad json", "u1", "PATCH", "/eew/l1", `[`, nil, false, 0, ""},
		{"update empty event", "u1", "PATCH", "/eew/l1", `{"eventName":""}`, nil, false, 0, ""},

		{"delete ok", "u1", "DELETE", "/eew/l1", "", nil, false, 204, ""},
		{"delete unauthenticated", "", "DELETE", "/eew/l1", "", nil, false, 401, "Unauthorized"},
		{"delete missing param", "u1", "DELETE", "/noparam/delete", "", nil, false, 400, "id is required"},
		{"delete lookup not found", "u1", "DELETE", "/eew/l1", "", nf, false, 404, "Element event workflow not found"},
		{"delete lookup internal", "u1", "DELETE", "/eew/l1", "", errInternal, false, 500, "Failed to retrieve element event workflow"},
		{"delete nil lookup", "u1", "DELETE", "/eew/l1", "", nil, true, 404, "Element event workflow not found"},

		{"delete by element ok", "u1", "DELETE", "/eew/element/e1", "", nil, false, 200, "Element event workflows deleted successfully"},
		{"delete by element unauthenticated", "", "DELETE", "/eew/element/e1", "", nil, false, 401, "Unauthorized"},
		{"delete by element missing param", "u1", "DELETE", "/noparam/element", "", nil, false, 400, "elementId is required"},
		{"delete by element internal", "u1", "DELETE", "/eew/element/e1", "", errInternal, false, 500, "Failed to delete element event workflows"},
		{"delete by workflow ok", "u1", "DELETE", "/eew/workflow/w1", "", nil, false, 200, "Element event workflows deleted successfully"},
		{"delete by workflow missing param", "u1", "DELETE", "/noparam/workflow", "", nil, false, 400, "workflowId is required"},
		{"delete by workflow internal", "u1", "DELETE", "/eew/workflow/w1", "", errInternal, false, 500, "Failed to delete element event workflows"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := &elementEventWorkflowSvcStub{err: tc.err, nilOnGet: tc.nilOnGet}
			r := kitSend(t, elementEventWorkflowDefaultKit(tc.user, svc), tc.method, tc.path, tc.body)
			if tc.contains == "" {
				kitExpect(t, r, tc.want)
			} else {
				kitExpect(t, r, tc.want, tc.contains)
			}
		})
	}
}

func TestElementEventWorkflowCreateMapsRequestToModel(t *testing.T) {
	svc := &elementEventWorkflowSvcStub{}
	kitExpect(t, kitSend(t, elementEventWorkflowDefaultKit("u1", svc), "POST", "/eew", `{"elementId":"e5","workflowId":"w6","eventName":"onLoad"}`), 201, `"id":"eew-new"`)
	if svc.created == nil || svc.created.ElementId != "e5" || svc.created.WorkflowId != "w6" || svc.created.EventName != "onLoad" {
		t.Fatalf("created %+v", svc.created)
	}
}

func TestElementEventWorkflowListPicksTheMatchingQuery(t *testing.T) {
	cases := []struct{ query, call string }{
		{"", "all"},
		{"?elementId=e1", "element:e1"},
		{"?workflowId=w1", "workflow:w1"},
		{"?elementId=e1&workflowId=w1", "filters:e1|w1|"},
		{"?elementId=e1&eventName=onClick", "filters:e1||onClick"},
		{"?eventName=onClick", "filters:||onClick"},
		{"?elementId=e1&workflowId=w1&eventName=onClick", "filters:e1|w1|onClick"},
	}
	for _, c := range cases {
		svc := &elementEventWorkflowSvcStub{}
		r := kitSend(t, elementEventWorkflowDefaultKit("u1", svc), "GET", "/eew"+c.query, "")
		kitExpect(t, r, 200)
		if len(svc.calls) != 1 || svc.calls[0] != c.call {
			t.Errorf("query %q called %v, want %q", c.query, svc.calls, c.call)
		}
		if b := r.object(t); b["count"] != float64(2) || len(b["data"].([]any)) != 2 {
			t.Errorf("query %q body %v", c.query, b)
		}
	}
}

func TestElementEventWorkflowByPageChecksProjectAccess(t *testing.T) {
	svc := &elementEventWorkflowSvcStub{}
	projects := &elementEventWorkflowProjectStub{}
	app := elementEventWorkflowKit("u9", svc, &elementEventWorkflowPageStub{}, projects)
	kitExpect(t, kitSend(t, app, "GET", "/eew/page/pg7", ""), 200, `"count":2`)
	if projects.gotProject != "proj-1" || projects.gotUser != "u9" || svc.byPage != "pg7" {
		t.Fatalf("project=%q user=%q page=%q", projects.gotProject, projects.gotUser, svc.byPage)
	}

	cases := []struct {
		name     string
		pageErr  error
		projErr  error
		want     int
		contains string
	}{
		{"page missing", gorm.ErrRecordNotFound, nil, 404, "Page not found"},
		{"page internal", errInternal, nil, 500, "Failed to retrieve page"},
		{"project forbidden", nil, repositories.ErrProjectUnauthorized, 403, "You do not have access to this project"},
		{"project missing", nil, repositories.ErrProjectNotFound, 404, "Project not found"},
		{"project internal", nil, errInternal, 500, "Failed to verify project access"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := &elementEventWorkflowSvcStub{}
			app := elementEventWorkflowKit("u9", svc, &elementEventWorkflowPageStub{err: tc.pageErr}, &elementEventWorkflowProjectStub{err: tc.projErr})
			kitExpect(t, kitSend(t, app, "GET", "/eew/page/pg7", ""), tc.want, tc.contains)
			if len(svc.calls) != 0 {
				t.Fatalf("workflows must not be fetched after a failed check, calls=%v", svc.calls)
			}
		})
	}
}

func TestElementEventWorkflowUpdateOnlyChangesEventName(t *testing.T) {
	svc := &elementEventWorkflowSvcStub{}
	r := kitSend(t, elementEventWorkflowDefaultKit("u1", svc), "PATCH", "/eew/l4", `{"eventName":"onHover","elementId":"spoofed","workflowId":"spoofed"}`)
	kitExpect(t, r, 200, `"eventName":"onHover"`)
	if svc.updatedID != "l4" || svc.updated == nil || svc.updated.EventName != "onHover" || svc.updated.ElementId != "e1" || svc.updated.WorkflowId != "w1" {
		t.Fatalf("id=%q updated=%+v", svc.updatedID, svc.updated)
	}
}

func TestElementEventWorkflowMutationErrorsAfterLookup(t *testing.T) {
	cases := []struct {
		name     string
		err      error
		want     int
		contains string
	}{
		{"not found", gorm.ErrRecordNotFound, 404, "Element event workflow not found"},
		{"internal", errInternal, 500, "Failed to"},
	}
	for _, tc := range cases {
		t.Run("update "+tc.name, func(t *testing.T) {
			svc := &elementEventWorkflowFailingStub{updateErr: tc.err}
			kitExpect(t, kitSend(t, elementEventWorkflowDefaultKit("u1", svc), "PATCH", "/eew/l1", `{"eventName":"x"}`), tc.want, tc.contains)
		})
		t.Run("delete "+tc.name, func(t *testing.T) {
			svc := &elementEventWorkflowFailingStub{deleteErr: tc.err}
			kitExpect(t, kitSend(t, elementEventWorkflowDefaultKit("u1", svc), "DELETE", "/eew/l1", ""), tc.want, tc.contains)
		})
	}
}

func TestElementEventWorkflowDeleteForwardsIDs(t *testing.T) {
	svc := &elementEventWorkflowSvcStub{}
	app := elementEventWorkflowDefaultKit("u1", svc)
	kitExpect(t, kitSend(t, app, "DELETE", "/eew/l3", ""), 204)
	kitExpect(t, kitSend(t, app, "DELETE", "/eew/element/e3", ""), 200)
	kitExpect(t, kitSend(t, app, "DELETE", "/eew/workflow/w3", ""), 200)
	if svc.deletedID != "l3" || svc.delElement != "e3" || svc.delFlow != "w3" {
		t.Fatalf("deleted=%q element=%q workflow=%q", svc.deletedID, svc.delElement, svc.delFlow)
	}
}
