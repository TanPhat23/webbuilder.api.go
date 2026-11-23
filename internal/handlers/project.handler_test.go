package handlers

import (
	"context"
	"fmt"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"
)

type projectHandlerStub struct {
	services.ProjectServiceInterface
	err error

	gotProject, gotUser string
	gotPublic           string
	sent                *models.Project
}

func (s *projectHandlerStub) GetProjectsByUserID(_ context.Context, userID string) ([]models.Project, error) {
	s.gotUser = userID
	if s.err != nil {
		return nil, s.err
	}
	return []models.Project{{ID: "p1", Name: "Site", OwnerId: userID}}, nil
}

func (s *projectHandlerStub) GetProjectWithAccess(_ context.Context, projectID, userID string) (*models.Project, error) {
	s.gotProject, s.gotUser = projectID, userID
	if s.err != nil {
		return nil, s.err
	}
	return &models.Project{ID: projectID, Name: "Site"}, nil
}

func (s *projectHandlerStub) GetPublicProjectByID(_ context.Context, projectID string) (*models.Project, error) {
	s.gotPublic = projectID
	if s.err != nil {
		return nil, s.err
	}
	return &models.Project{ID: projectID, Name: "Public site", Published: true}, nil
}

func (s *projectHandlerStub) GetProjectPages(_ context.Context, projectID, userID string) ([]models.Page, error) {
	s.gotProject, s.gotUser = projectID, userID
	if s.err != nil {
		return nil, s.err
	}
	return []models.Page{{Id: "pg1", Name: "Home"}}, nil
}

func (s *projectHandlerStub) DeleteProject(_ context.Context, projectID, userID string) error {
	s.gotProject, s.gotUser = projectID, userID
	return s.err
}

func (s *projectHandlerStub) UpdateProject(_ context.Context, projectID, userID string, p *models.Project) (*models.Project, error) {
	s.gotProject, s.gotUser, s.sent = projectID, userID, p
	if s.err != nil {
		return nil, s.err
	}
	return &models.Project{ID: projectID, Name: "updated"}, nil
}

func (s *projectHandlerStub) CreateProjectWithPage(_ context.Context, p *models.Project) (*models.Project, error) {
	s.sent = p
	if s.err != nil {
		return nil, s.err
	}
	created := *p
	if created.ID == "" {
		created.ID = "generated"
	}
	return &created, nil
}

func (s *projectHandlerStub) GetProjectAccess(_ context.Context, projectID, userID string) (*models.ProjectAccess, error) {
	s.gotProject, s.gotUser = projectID, userID
	if s.err != nil {
		return nil, s.err
	}
	return &models.ProjectAccess{UserId: userID, ProjectId: projectID, Role: models.RoleViewer, IsCollaborator: true}, nil
}

func projectHandlerApp(user string, svc *projectHandlerStub) *fiber.App {
	h := NewProjectHandler(svc)
	return newKitApp(user, func(app *fiber.App) {
		app.Get("/projects", h.GetProjectsByUser)
		app.Get("/my-projects", h.GetProjectByUserID)
		app.Post("/projects", h.CreateProject)
		app.Get("/public/projects/:projectid", h.GetPublicProjectByID)
		app.Get("/projects/:projectid/pages", h.GetProjectPages)
		app.Get("/projects/:projectid/access", h.GetProjectAccess)
		app.Get("/projects/:projectid", h.GetProjectByID)
		app.Patch("/projects/:projectid", h.UpdateProject)
		app.Delete("/projects/:projectid", h.DeleteProject)
		app.Get("/noid", h.GetProjectByID)
		app.Get("/noid/public", h.GetPublicProjectByID)
		app.Get("/noid/pages", h.GetProjectPages)
		app.Get("/noid/access", h.GetProjectAccess)
		app.Patch("/noid", h.UpdateProject)
		app.Delete("/noid", h.DeleteProject)
	})
}

func expectKitStatus(t *testing.T, r kitReply, want int, expect string) {
	t.Helper()
	if expect == "" {
		kitExpect(t, r, want)
		return
	}
	kitExpect(t, r, want, expect)
}

func TestProjectListHandlers(t *testing.T) {
	for _, path := range []string{"/projects", "/my-projects"} {
		svc := &projectHandlerStub{}
		kitExpect(t, kitSend(t, projectHandlerApp("u1", svc), "GET", path, ""), 200, "Site", "u1")
		if svc.gotUser != "u1" {
			t.Fatalf("user = %q", svc.gotUser)
		}
		kitExpect(t, kitSend(t, projectHandlerApp("", &projectHandlerStub{}), "GET", path, ""), 401)
	}
	kitExpect(t, kitSend(t, projectHandlerApp("u1", &projectHandlerStub{err: errInternal}), "GET", "/projects", ""), 500, "Failed to retrieve projects")
	kitExpect(t, kitSend(t, projectHandlerApp("u1", &projectHandlerStub{err: errInternal}), "GET", "/my-projects", ""), 500, "Failed to retrieve projects by user ID")
}

func TestProjectReadHandlerStatuses(t *testing.T) {
	cases := []struct {
		name, user, path string
		err              error
		want             int
		expect           string
	}{
		{"get ok", "u1", "/projects/p1", nil, 200, "Site"},
		{"get anonymous", "", "/projects/p1", nil, 401, ""},
		{"get missing id", "u1", "/noid", nil, 400, "projectid is required"},
		{"get stranger", "u1", "/projects/p1", repositories.ErrProjectUnauthorized, 403, "You do not have access to this project"},
		{"get missing", "u1", "/projects/p1", repositories.ErrProjectNotFound, 404, "Project not found"},
		{"get failure", "u1", "/projects/p1", errInternal, 500, "Failed to retrieve project"},
		{"public ok", "", "/public/projects/p1", nil, 200, "Public site"},
		{"public missing id", "", "/noid/public", nil, 400, "projectid is required"},
		{"public not found", "", "/public/projects/p1", gorm.ErrRecordNotFound, 404, "Project not found"},
		{"public failure", "", "/public/projects/p1", errInternal, 500, "Failed to retrieve project"},
		{"pages ok", "u1", "/projects/p1/pages", nil, 200, "Home"},
		{"pages anonymous", "", "/projects/p1/pages", nil, 401, ""},
		{"pages missing id", "u1", "/noid/pages", nil, 400, "projectid is required"},
		{"pages stranger", "u1", "/projects/p1/pages", repositories.ErrProjectUnauthorized, 403, "You do not have access to this project"},
		{"pages failure", "u1", "/projects/p1/pages", errInternal, 500, "Failed to retrieve project pages"},
		{"access ok", "u1", "/projects/p1/access", nil, 200, "viewer"},
		{"access anonymous", "", "/projects/p1/access", nil, 401, ""},
		{"access missing id", "u1", "/noid/access", nil, 400, "projectid is required"},
		{"access not found", "u1", "/projects/p1/access", fmt.Errorf("lookup: %w", repositories.ErrProjectNotFound), 404, "Project not found"},
		{"access failure", "u1", "/projects/p1/access", errInternal, 500, "Failed to check project access"},
	}
	for _, tc := range cases {
		svc := &projectHandlerStub{err: tc.err}
		r := kitSend(t, projectHandlerApp(tc.user, svc), "GET", tc.path, "")
		expectKitStatus(t, r, tc.want, tc.expect)
		assertReadCallTracked(t, tc.name, tc.user, tc.want, svc)
	}
}

func assertReadCallTracked(t *testing.T, name, user string, want int, svc *projectHandlerStub) {
	t.Helper()
	if want == 200 && user != "" {
		if svc.gotProject != "p1" || svc.gotUser != "u1" {
			t.Fatalf("%s: service got project=%q user=%q", name, svc.gotProject, svc.gotUser)
		}
	}
	if name == "public ok" && svc.gotPublic != "p1" {
		t.Fatalf("public id = %q", svc.gotPublic)
	}
}

func TestDeleteProject(t *testing.T) {
	svc := &projectHandlerStub{}
	kitExpect(t, kitSend(t, projectHandlerApp("u1", svc), "DELETE", "/projects/p1", ""), 204)
	if svc.gotProject != "p1" || svc.gotUser != "u1" {
		t.Fatalf("service got project=%q user=%q", svc.gotProject, svc.gotUser)
	}
	kitExpect(t, kitSend(t, projectHandlerApp("", &projectHandlerStub{}), "DELETE", "/projects/p1", ""), 401)
	kitExpect(t, kitSend(t, projectHandlerApp("u1", &projectHandlerStub{}), "DELETE", "/noid", ""), 400, "projectid is required")
	kitExpect(t, kitSend(t, projectHandlerApp("u1", &projectHandlerStub{err: repositories.ErrProjectUnauthorized}), "DELETE", "/projects/p1", ""), 403, "You do not have access to this project")
	kitExpect(t, kitSend(t, projectHandlerApp("u1", &projectHandlerStub{err: errInternal}), "DELETE", "/projects/p1", ""), 500, "Failed to delete project")
}

func TestUpdateProjectMapsEveryField(t *testing.T) {
	svc := &projectHandlerStub{}
	body := `{"name":"Renamed","description":"About","published":true,"subdomain":"mysite","styles":{"bg":"red"},"header":{"h":1}}`
	kitExpect(t, kitSend(t, projectHandlerApp("u1", svc), "PATCH", "/projects/p1", body), 200, "updated")
	p := svc.sent
	if svc.gotProject != "p1" || svc.gotUser != "u1" || p == nil {
		t.Fatalf("service got project=%q user=%q sent=%v", svc.gotProject, svc.gotUser, p)
	}
	assertUpdatedProjectScalars(t, p)
	assertUpdatedProjectJSON(t, p)
}

func assertUpdatedProjectScalars(t *testing.T, p *models.Project) {
	t.Helper()
	if p.Name != "Renamed" || !p.Published {
		t.Fatalf("scalar fields = %+v", p)
	}
	if p.Description == nil || *p.Description != "About" {
		t.Fatalf("description = %v", p.Description)
	}
	if p.Subdomain == nil || *p.Subdomain != "mysite" {
		t.Fatalf("subdomain = %v", p.Subdomain)
	}
}

func assertUpdatedProjectJSON(t *testing.T, p *models.Project) {
	t.Helper()
	if p.Styles == nil || string(*p.Styles) != `{"bg":"red"}` {
		t.Fatalf("styles = %v", p.Styles)
	}
	if p.Header == nil || string(*p.Header) != `{"h":1}` {
		t.Fatalf("header = %v", p.Header)
	}
}

func TestUpdateProjectPartialAndUnpublish(t *testing.T) {
	svc := &projectHandlerStub{}
	kitExpect(t, kitSend(t, projectHandlerApp("u1", svc), "PATCH", "/projects/p1", `{"published":false}`), 200)
	if svc.sent.Published || svc.sent.Name != "" {
		t.Fatalf("only published should be set: %+v", svc.sent)
	}
	if svc.sent.Description != nil || svc.sent.Subdomain != nil {
		t.Fatalf("description/subdomain must be nil: %+v", svc.sent)
	}
	if svc.sent.Styles != nil || svc.sent.Header != nil {
		t.Fatalf("styles/header must be nil: %+v", svc.sent)
	}
	// A non-object style value is kept as raw JSON.
	svc = &projectHandlerStub{}
	kitExpect(t, kitSend(t, projectHandlerApp("u1", svc), "PATCH", "/projects/p1", `{"styles":"dark","header":["a"]}`), 200)
	if string(*svc.sent.Styles) != `"dark"` || string(*svc.sent.Header) != `["a"]` {
		t.Fatalf("styles=%s header=%s", *svc.sent.Styles, *svc.sent.Header)
	}
}

func TestUpdateProjectRejections(t *testing.T) {
	cases := []struct {
		name, user, path, body string
		err                    error
		want                   int
		expect                 string
	}{
		{"anonymous", "", "/projects/p1", `{"name":"a"}`, nil, 401, ""},
		{"missing id", "u1", "/noid", `{"name":"a"}`, nil, 400, "projectid is required"},
		{"bad json", "u1", "/projects/p1", `{`, nil, 0, ""},
		{"empty name", "u1", "/projects/p1", `{"name":""}`, nil, 0, ""},
		{"short subdomain", "u1", "/projects/p1", `{"subdomain":"ab"}`, nil, 0, ""},
		{"nothing to update", "u1", "/projects/p1", `{}`, nil, 400, "No valid fields to update"},
		{"unknown fields only", "u1", "/projects/p1", `{"ownerId":"x"}`, nil, 400, "No valid fields to update"},
		{"stranger", "u1", "/projects/p1", `{"name":"a"}`, repositories.ErrProjectUnauthorized, 403, "You do not have access to this project"},
		{"missing project", "u1", "/projects/p1", `{"name":"a"}`, repositories.ErrProjectNotFound, 404, "Project not found"},
		{"failure", "u1", "/projects/p1", `{"name":"a"}`, errInternal, 500, "Failed to update project"},
	}
	for _, tc := range cases {
		svc := &projectHandlerStub{err: tc.err}
		r := kitSend(t, projectHandlerApp(tc.user, svc), "PATCH", tc.path, tc.body)
		expectKitStatus(t, r, tc.want, tc.expect)
		if tc.err == nil && svc.sent != nil {
			t.Fatalf("%s: service reached for a rejected request: %+v", tc.name, svc.sent)
		}
	}
}

func TestCreateProjectDefaults(t *testing.T) {
	svc := &projectHandlerStub{}
	r := kitSend(t, projectHandlerApp("u1", svc), "POST", "/projects", `{"name":"Blog"}`)
	kitExpect(t, r, 201, "Blog", "generated")
	p := svc.sent
	if p.OwnerId != "u1" || p.Name != "Blog" || p.Published {
		t.Fatalf("sent %+v", p)
	}
	if p.Header != nil || p.Styles == nil || string(*p.Styles) != "{}" {
		t.Fatalf("styles/header mismatch %+v", p)
	}
}

func TestCreateProjectForwardsEveryField(t *testing.T) {
	svc := &projectHandlerStub{}
	id := uuid.NewString()
	body := `{"id":"` + id + `","name":"Blog","description":"d","subdomain":"myblog","published":true,"styles":{"a":1},"header":{"b":2}}`
	r := kitSend(t, projectHandlerApp("u1", svc), "POST", "/projects", body)
	kitExpect(t, r, 201, id)
	assertCreatedProjectForwarded(t, svc.sent, id)
	var got models.Project
	r.decode(t, &got)
	if got.ID != id || got.OwnerId != "u1" {
		t.Fatalf("response %+v", got)
	}
}

func assertCreatedProjectForwarded(t *testing.T, p *models.Project, id string) {
	t.Helper()
	if p.ID != id || !p.Published {
		t.Fatalf("sent %+v", p)
	}
	if p.Description == nil || *p.Description != "d" {
		t.Fatalf("description mismatch %+v", p)
	}
	if p.Subdomain == nil || *p.Subdomain != "myblog" {
		t.Fatalf("subdomain mismatch %+v", p)
	}
	if p.Styles == nil || string(*p.Styles) != `{"a":1}` {
		t.Fatalf("styles mismatch %+v", p)
	}
	if p.Header == nil || string(*p.Header) != `{"b":2}` {
		t.Fatalf("header mismatch %+v", p)
	}
}

func TestCreateProjectRejections(t *testing.T) {
	for _, tc := range []struct {
		name, user, body string
		err              error
		want             int
		expect           string
	}{
		{"anonymous", "", `{"name":"a"}`, nil, 401, ""},
		{"bad json", "u1", `{`, nil, 0, ""},
		{"missing name", "u1", `{}`, nil, 0, ""},
		{"id not a uuid", "u1", `{"id":"x","name":"a"}`, nil, 0, ""},
		{"non alphanumeric subdomain", "u1", `{"name":"a","subdomain":"my-site"}`, nil, 0, ""},
		{"failure", "u1", `{"name":"a"}`, errInternal, 500, "Failed to create project"},
	} {
		svc := &projectHandlerStub{err: tc.err}
		r := kitSend(t, projectHandlerApp(tc.user, svc), "POST", "/projects", tc.body)
		expectKitStatus(t, r, tc.want, tc.expect)
		if tc.err == nil && svc.sent != nil {
			t.Fatalf("%s: project created for a rejected request", tc.name)
		}
	}
}
