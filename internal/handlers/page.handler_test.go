package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"
)

// pageHandlerStub records what reaches the page service. getErrs are returned
// by successive GetByID calls (nil once exhausted).
type pageHandlerStub struct {
	services.PageServiceInterface
	err     error
	getErrs []error

	gotPageID, gotProjectID, gotUser string
	gotPublicProject, gotPublicPage  string
	created                          *models.Page
	updates                          map[string]any
	getCalls                         int
}

func (s *pageHandlerStub) DeleteForUser(_ context.Context, pageID, projectID, userID string) error {
	s.gotPageID, s.gotProjectID, s.gotUser = pageID, projectID, userID
	return s.err
}

func (s *pageHandlerStub) ListByProject(_ context.Context, projectID, userID string) ([]models.Page, error) {
	s.gotProjectID, s.gotUser = projectID, userID
	if s.err != nil {
		return nil, s.err
	}
	return []models.Page{{Id: "pg1", ProjectId: projectID, Name: "Home"}}, nil
}

func (s *pageHandlerStub) GetByID(_ context.Context, pageID string) (*models.Page, error) {
	s.gotPageID = pageID
	call := s.getCalls
	s.getCalls++
	if call < len(s.getErrs) && s.getErrs[call] != nil {
		return nil, s.getErrs[call]
	}
	if s.err != nil && len(s.getErrs) == 0 {
		return nil, s.err
	}
	return &models.Page{Id: pageID, Name: fmt.Sprintf("read-%d", call)}, nil
}

func (s *pageHandlerStub) GetPublicByID(_ context.Context, projectID, pageID string) (*models.Page, error) {
	s.gotPublicProject, s.gotPublicPage = projectID, pageID
	if s.err != nil {
		return nil, s.err
	}
	return &models.Page{Id: pageID, ProjectId: projectID, Name: "Public"}, nil
}

func (s *pageHandlerStub) CreateForUser(_ context.Context, page *models.Page, userID string) (*models.Page, error) {
	s.created, s.gotUser = page, userID
	if s.err != nil {
		return nil, s.err
	}
	return page, nil
}

func (s *pageHandlerStub) GetByIDForUser(_ context.Context, projectID, pageID, userID string) (*models.Page, error) {
	s.gotProjectID, s.gotUser = projectID, userID
	return s.GetByID(context.Background(), pageID)
}

func (s *pageHandlerStub) UpdateFields(_ context.Context, pageID string, updates map[string]any) error {
	s.gotPageID, s.updates = pageID, updates
	return s.err
}

func (s *pageHandlerStub) UpdateFieldsForUser(_ context.Context, projectID, pageID, userID string, updates map[string]any) error {
	s.gotProjectID, s.gotUser = projectID, userID
	return s.UpdateFields(context.Background(), pageID, updates)
}

func pageHandlerApp(user string, svc services.PageServiceInterface) *fiber.App {
	h := NewPageHandler(svc)
	return newKitApp(user, func(app *fiber.App) {
		app.Get("/projects/:projectid/pages", h.ListByProject)
		app.Get("/public/projects/:projectid/pages", h.ListPublic)
		app.Get("/public/projects/:projectid/pages/:pageid", h.GetPublicByID)
		app.Post("/projects/:projectid/pages", h.Create)
		app.Get("/projects/:projectid/pages/:pageid", h.GetByID)
		app.Patch("/projects/:projectid/pages/:pageid", h.Update)
		app.Delete("/projects/:projectid/pages/:pageid", h.Delete)
		app.Get("/nopage/:projectid", h.GetByID)
		app.Patch("/nopage/:projectid", h.Update)
		app.Delete("/nopage/:projectid", h.Delete)
		app.Get("/nopage/public/:projectid", h.GetPublicByID)
		app.Get("/noproject", h.ListByProject)
		app.Post("/noproject", h.Create)
	})
}

func TestDeletePage(t *testing.T) {
	svc := &pageHandlerStub{}
	kitExpect(t, kitSend(t, pageHandlerApp("u1", svc), "DELETE", "/projects/p1/pages/pg1", ""), 204)
	if svc.gotPageID != "pg1" || svc.gotProjectID != "p1" || svc.gotUser != "u1" {
		t.Fatalf("service got page=%q project=%q user=%q", svc.gotPageID, svc.gotProjectID, svc.gotUser)
	}
	kitExpect(t, kitSend(t, pageHandlerApp("", &pageHandlerStub{}), "DELETE", "/projects/p1/pages/pg1", ""), 401)
	kitExpect(t, kitSend(t, pageHandlerApp("u1", &pageHandlerStub{}), "DELETE", "/nopage/p1", ""), 400, "pageid is required")
	kitExpect(t, kitSend(t, pageHandlerApp("u1", &pageHandlerStub{err: gorm.ErrRecordNotFound}), "DELETE", "/projects/p1/pages/pg1", ""), 404, "Page not found or not owned by user")
	kitExpect(t, kitSend(t, pageHandlerApp("u1", &pageHandlerStub{err: errInternal}), "DELETE", "/projects/p1/pages/pg1", ""), 500, "Failed to delete page")
}

func TestGetPagesByProject(t *testing.T) {
	t.Run("private read passes the signed-in user", func(t *testing.T) {
		svc := &pageHandlerStub{}
		kitExpect(t, kitSend(t, pageHandlerApp("u1", svc), "GET", "/projects/p1/pages", ""), 200, "Home", "p1")
		if svc.gotProjectID != "p1" || svc.gotUser != "u1" {
			t.Fatalf("service got project=%q user=%q", svc.gotProjectID, svc.gotUser)
		}
	})
	t.Run("private read without a user passes an empty user", func(t *testing.T) {
		svc := &pageHandlerStub{}
		kitExpect(t, kitSend(t, pageHandlerApp("", svc), "GET", "/projects/p1/pages", ""), 200)
		if svc.gotUser != "" {
			t.Fatalf("user = %q", svc.gotUser)
		}
	})
	t.Run("public read never passes the user", func(t *testing.T) {
		svc := &pageHandlerStub{}
		kitExpect(t, kitSend(t, pageHandlerApp("u1", svc), "GET", "/public/projects/p1/pages", ""), 200, "Home")
		if svc.gotProjectID != "p1" || svc.gotUser != "" {
			t.Fatalf("service got project=%q user=%q", svc.gotProjectID, svc.gotUser)
		}
	})
	t.Run("missing project id", func(t *testing.T) {
		kitExpect(t, kitSend(t, pageHandlerApp("u1", &pageHandlerStub{}), "GET", "/noproject", ""), 400, "projectid is required")
	})
	t.Run("not found", func(t *testing.T) {
		kitExpect(t, kitSend(t, pageHandlerApp("u1", &pageHandlerStub{err: gorm.ErrRecordNotFound}), "GET", "/projects/p1/pages", ""), 404, "Project not found")
	})
	t.Run("failure", func(t *testing.T) {
		kitExpect(t, kitSend(t, pageHandlerApp("u1", &pageHandlerStub{err: errInternal}), "GET", "/projects/p1/pages", ""), 500, "Failed to retrieve pages")
	})
}

func TestGetPageByID(t *testing.T) {
	svc := &pageHandlerStub{}
	kitExpect(t, kitSend(t, pageHandlerApp("u1", svc), "GET", "/projects/p1/pages/pg1", ""), 200, `"id":"pg1"`)
	if svc.gotPageID != "pg1" {
		t.Fatalf("page id = %q", svc.gotPageID)
	}
	kitExpect(t, kitSend(t, pageHandlerApp("", &pageHandlerStub{}), "GET", "/projects/p1/pages/pg1", ""), 401)
	kitExpect(t, kitSend(t, pageHandlerApp("u1", &pageHandlerStub{}), "GET", "/nopage/p1", ""), 400, "pageid is required")
	kitExpect(t, kitSend(t, pageHandlerApp("u1", &pageHandlerStub{err: gorm.ErrRecordNotFound}), "GET", "/projects/p1/pages/pg1", ""), 404, "Page not found")
	kitExpect(t, kitSend(t, pageHandlerApp("u1", &pageHandlerStub{err: errInternal}), "GET", "/projects/p1/pages/pg1", ""), 500, "Failed to retrieve page")
}

func TestGetPublicPageByID(t *testing.T) {
	svc := &pageHandlerStub{}
	kitExpect(t, kitSend(t, pageHandlerApp("", svc), "GET", "/public/projects/p1/pages/pg1", ""), 200, "Public", "p1")
	if svc.gotPublicProject != "p1" || svc.gotPublicPage != "pg1" {
		t.Fatalf("service got project=%q page=%q", svc.gotPublicProject, svc.gotPublicPage)
	}
	kitExpect(t, kitSend(t, pageHandlerApp("", &pageHandlerStub{}), "GET", "/nopage/public/p1", ""), 400, "pageid is required")
	kitExpect(t, kitSend(t, pageHandlerApp("", &pageHandlerStub{err: gorm.ErrRecordNotFound}), "GET", "/public/projects/p1/pages/pg1", ""), 404, "Page not found")
	kitExpect(t, kitSend(t, pageHandlerApp("", &pageHandlerStub{err: errInternal}), "GET", "/public/projects/p1/pages/pg1", ""), 500, "Failed to retrieve page")
}

func TestCreatePage(t *testing.T) {
	t.Run("generates an id and persists the request", func(t *testing.T) {
		svc := &pageHandlerStub{}
		r := kitSend(t, pageHandlerApp("u1", svc), "POST", "/projects/p1/pages", `{"name":"About","type":"page","styles":{"color":"red"}}`)
		kitExpect(t, r, 201, "About")
		p := svc.created
		if p == nil || p.Name != "About" || p.Type != "page" || p.ProjectId != "p1" || svc.gotUser != "u1" {
			t.Fatalf("created %+v by %q", p, svc.gotUser)
		}
		if _, err := uuid.Parse(p.Id); err != nil {
			t.Fatalf("generated id %q is not a uuid", p.Id)
		}
		if string(p.Styles) != `{"color":"red"}` {
			t.Fatalf("styles = %s", p.Styles)
		}
		if p.CreatedAt.IsZero() || !p.CreatedAt.Equal(p.UpdatedAt) {
			t.Fatalf("audit timestamps not set: %v %v", p.CreatedAt, p.UpdatedAt)
		}
		var body models.Page
		r.decode(t, &body)
		if body.Id != p.Id {
			t.Fatalf("response id %q != persisted %q", body.Id, p.Id)
		}
	})
	t.Run("keeps a client chosen id", func(t *testing.T) {
		svc := &pageHandlerStub{}
		id := uuid.NewString()
		kitExpect(t, kitSend(t, pageHandlerApp("u1", svc), "POST", "/projects/p1/pages", `{"id":"`+id+`","name":"About","type":"page"}`), 201, id)
		if svc.created.Id != id {
			t.Fatalf("id = %q, want %q", svc.created.Id, id)
		}
	})
	rejected := []struct {
		name, user, path, body string
		want                   int
		expect                 string
	}{
		{"anonymous", "", "/projects/p1/pages", `{"name":"a","type":"page"}`, 401, ""},
		{"missing project", "u1", "/noproject", `{"name":"a","type":"page"}`, 400, "projectid is required"},
		{"bad json", "u1", "/projects/p1/pages", `{`, 0, ""},
		{"missing name", "u1", "/projects/p1/pages", `{"type":"page"}`, 0, ""},
		{"missing type", "u1", "/projects/p1/pages", `{"name":"a"}`, 0, ""},
		{"id not a uuid", "u1", "/projects/p1/pages", `{"id":"nope","name":"a","type":"page"}`, 0, ""},
	}
	for _, tc := range rejected {
		t.Run(tc.name, func(t *testing.T) {
			svc := &pageHandlerStub{}
			r := kitSend(t, pageHandlerApp(tc.user, svc), "POST", tc.path, tc.body)
			if tc.expect == "" {
				kitExpect(t, r, tc.want)
			} else {
				kitExpect(t, r, tc.want, tc.expect)
			}
			if svc.created != nil {
				t.Fatal("page was created")
			}
		})
	}
	for _, tc := range []struct {
		name   string
		err    error
		want   int
		expect string
	}{
		{"forbidden", repositories.ErrProjectUnauthorized, 403, "You cannot add pages to this project"},
		{"duplicate name", fmt.Errorf("insert: %w", repositories.ErrPageNameTaken), 409, "A page with this name already exists in the project"},
		{"failure", errInternal, 500, "Failed to create page"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			kitExpect(t, kitSend(t, pageHandlerApp("u1", &pageHandlerStub{err: tc.err}), "POST", "/projects/p1/pages", `{"name":"a","type":"page"}`), tc.want, tc.expect)
		})
	}
}

func TestUpdatePage(t *testing.T) {
	t.Run("updates only the supplied fields and returns the re-read page", func(t *testing.T) {
		svc := &pageHandlerStub{}
		r := kitSend(t, pageHandlerApp("u1", svc), "PATCH", "/projects/p1/pages/pg1", `{"name":"New","type":"landing","styles":{"a":1}}`)
		kitExpect(t, r, 200, `"id":"pg1"`, "read-0")
		want := map[string]any{"Name": "New", "Type": "landing", "Styles": json.RawMessage(`{"a":1}`)}
		if !reflect.DeepEqual(svc.updates, want) {
			t.Fatalf("updates = %#v, want %#v", svc.updates, want)
		}
	})
	t.Run("a single field", func(t *testing.T) {
		svc := &pageHandlerStub{}
		kitExpect(t, kitSend(t, pageHandlerApp("u1", svc), "PATCH", "/projects/p1/pages/pg1", `{"name":"Only"}`), 200)
		if want := map[string]any{"Name": "Only"}; !reflect.DeepEqual(svc.updates, want) {
			t.Fatalf("updates = %#v, want %#v", svc.updates, want)
		}
	})
	t.Run("ignores fields outside the allowlist", func(t *testing.T) {
		svc := &pageHandlerStub{}
		kitExpect(t, kitSend(t, pageHandlerApp("u1", svc), "PATCH", "/projects/p1/pages/pg1", `{"id":"other","projectId":"x"}`), 400, "No valid fields to update")
		if svc.updates != nil {
			t.Fatalf("service updated %v", svc.updates)
		}
	})
	t.Run("missing page id", func(t *testing.T) {
		kitExpect(t, kitSend(t, pageHandlerApp("u1", &pageHandlerStub{}), "PATCH", "/nopage/p1", `{"name":"a"}`), 400, "pageid is required")
	})
	t.Run("page missing", func(t *testing.T) {
		svc := &pageHandlerUpdateFails{updateErr: gorm.ErrRecordNotFound}
		kitExpect(t, kitSend(t, pageHandlerApp("u1", svc), "PATCH", "/projects/p1/pages/pg1", `{"name":"a"}`), 404, "Page not found")
	})
	t.Run("invalid dynamic slug on static page returns 400", func(t *testing.T) {
		svc := &pageHandlerUpdateFails{updateErr: services.ErrInvalidPageSlug}
		kitExpect(t, kitSend(t, pageHandlerApp("u1", svc), "PATCH", "/projects/p1/pages/pg1", `{"slug":"blog/:slug"}`), 400, "dynamic")
	})
	t.Run("update failure", func(t *testing.T) {
		svc := &pageHandlerUpdateFails{updateErr: errInternal}
		kitExpect(t, kitSend(t, pageHandlerApp("u1", svc), "PATCH", "/projects/p1/pages/pg1", `{"name":"a"}`), 500, "Failed to update page")
	})
	for _, body := range []string{`{`, `{"name":""}`, `{"type":""}`} {
		t.Run("invalid body "+body, func(t *testing.T) {
			svc := &pageHandlerStub{}
			kitExpect(t, kitSend(t, pageHandlerApp("u1", svc), "PATCH", "/projects/p1/pages/pg1", body), 0)
			if svc.updates != nil {
				t.Fatal("invalid body was applied")
			}
		})
	}
}

// updateFailingPageStub fails UpdateFields only, so the post read stays separate.
type pageHandlerUpdateFails struct {
	pageHandlerStub
	updateErr error
}

func (s *pageHandlerUpdateFails) UpdateFields(ctx context.Context, id string, u map[string]any) error {
	_ = s.pageHandlerStub.UpdateFields(ctx, id, u)
	return s.updateErr
}

func (s *pageHandlerUpdateFails) UpdateFieldsForUser(ctx context.Context, projectID, id, userID string, u map[string]any) error {
	_ = s.pageHandlerStub.UpdateFieldsForUser(ctx, projectID, id, userID, u)
	return s.updateErr
}

func TestUpdatePageServiceFailures(t *testing.T) {
	app := func(svc services.PageServiceInterface) *fiber.App {
		h := NewPageHandler(svc)
		return newKitApp("u1", func(a *fiber.App) { a.Patch("/projects/:projectid/pages/:pageid", h.Update) })
	}
	kitExpect(t, kitSend(t, app(&pageHandlerUpdateFails{updateErr: gorm.ErrRecordNotFound}), "PATCH", "/projects/p1/pages/pg1", `{"name":"a"}`), 404, "Page not found")
	kitExpect(t, kitSend(t, app(&pageHandlerUpdateFails{updateErr: errInternal}), "PATCH", "/projects/p1/pages/pg1", `{"name":"a"}`), 500, "Failed to update page")
	reread := &pageHandlerStub{getErrs: []error{errInternal}}
	kitExpect(t, kitSend(t, app(reread), "PATCH", "/projects/p1/pages/pg1", `{"name":"a"}`), 500, "Failed to fetch updated page")
	if reread.updates == nil {
		t.Fatal("update was not applied before the re-read failed")
	}
}
