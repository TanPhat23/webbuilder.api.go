package handlers

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"
)

type elementServiceSpy struct {
	services.ElementServiceInterface
	projectReads []string
	pageReads    [][]string
}

func (s *elementServiceSpy) GetElements(_ context.Context, projectID string, _ ...string) ([]models.EditorElement, error) {
	// Fiber reuses param buffers across requests, so keep a copy.
	s.projectReads = append(s.projectReads, strings.Clone(projectID))
	return []models.EditorElement{}, nil
}

func (s *elementServiceSpy) GetElementsByPageIds(_ context.Context, projectID string, pageIDs []string) ([]models.EditorElement, error) {
	read := []string{strings.Clone(projectID)}
	for _, id := range pageIDs {
		read = append(read, strings.Clone(id))
	}
	s.pageReads = append(s.pageReads, read)
	return []models.EditorElement{}, nil
}

// projectsFake models three projects: "mine" (user u1 may read it), "theirs"
// (exists, not u1's) and "draft"/"live" for the public checks.
type projectsFake struct{ calls int }

func (p *projectsFake) GetProjectWithAccess(_ context.Context, projectID, userID string) (*models.Project, error) {
	p.calls++
	switch {
	case projectID == "mine" && userID == "u1":
		return &models.Project{}, nil
	case projectID == "theirs":
		return nil, repositories.ErrProjectUnauthorized
	default:
		return nil, repositories.ErrProjectNotFound
	}
}

func (p *projectsFake) GetPublicProjectByID(_ context.Context, projectID string) (*models.Project, error) {
	p.calls++
	switch projectID {
	case "live":
		return &models.Project{Published: true}, nil
	case "draft":
		// The repository filters on Published, but the handler must not rely on it alone.
		return &models.Project{Published: false}, nil
	default:
		return nil, repositories.ErrProjectNotFound
	}
}

func elementAccessApp(spy *elementServiceSpy, projects *projectsFake, userID string) *fiber.App {
	app := fiber.New()
	h := NewElementHandler(spy, projects)
	api := app.Group("/api/v1", func(c fiber.Ctx) error {
		if userID != "" {
			c.Locals("userId", userID)
		}
		return c.Next()
	})
	api.Get("/elements/public/by-pages", h.GetPublicElementsByPageIds)
	api.Get("/elements/public/:projectid", h.GetPublicElements)
	api.Get("/elements/by-pages", h.GetElementsByPageIds)
	api.Get("/elements/:projectid", h.GetElements)
	return app
}

func status(t *testing.T, app *fiber.App, path string) int {
	t.Helper()
	res, err := app.Test(httptest.NewRequest(http.MethodGet, path, nil))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	_, _ = io.Copy(io.Discard, res.Body)
	return res.StatusCode
}

func TestPrivateElementReadsRequireProjectAccess(t *testing.T) {
	spy, projects := &elementServiceSpy{}, &projectsFake{}
	app := elementAccessApp(spy, projects, "u1")

	for _, tc := range []struct {
		path string
		want int
	}{
		{"/api/v1/elements/mine", 200},
		{"/api/v1/elements/theirs", 403},
		{"/api/v1/elements/missing", 404},
		{"/api/v1/elements/by-pages?projectId=mine&pageIds=a,b", 200},
		{"/api/v1/elements/by-pages?projectId=theirs&pageIds=a", 403},
		{"/api/v1/elements/by-pages?projectId=missing&pageIds=a", 404},
		{"/api/v1/elements/by-pages?pageIds=a", 400},
		{"/api/v1/elements/by-pages?projectId=mine", 400},
	} {
		if got := status(t, app, tc.path); got != tc.want {
			t.Errorf("GET %s = %d, want %d", tc.path, got, tc.want)
		}
	}
	if len(spy.projectReads) != 1 || spy.projectReads[0] != "mine" {
		t.Fatalf("elements were read for %v, only \"mine\" may be read", spy.projectReads)
	}
	if len(spy.pageReads) != 1 || strings.Join(spy.pageReads[0], ",") != "mine,a,b" {
		t.Fatalf("page reads = %v, want one read scoped to project mine", spy.pageReads)
	}
}

func TestPrivateElementReadsRequireLogin(t *testing.T) {
	spy := &elementServiceSpy{}
	app := elementAccessApp(spy, &projectsFake{}, "")
	for _, path := range []string{"/api/v1/elements/mine", "/api/v1/elements/by-pages?projectId=mine&pageIds=a"} {
		if got := status(t, app, path); got != 401 {
			t.Errorf("GET %s without a user = %d, want 401", path, got)
		}
	}
	if len(spy.projectReads)+len(spy.pageReads) != 0 {
		t.Fatal("elements were read without a signed-in user")
	}
}

func TestPublicElementReadsRequirePublishedProject(t *testing.T) {
	spy := &elementServiceSpy{}
	app := elementAccessApp(spy, &projectsFake{}, "")

	for _, tc := range []struct {
		path string
		want int
	}{
		{"/api/v1/elements/public/live", 200},
		{"/api/v1/elements/public/draft", 404},
		{"/api/v1/elements/public/missing", 404},
		{"/api/v1/elements/public/by-pages?projectId=live&pageIds=a", 200},
		{"/api/v1/elements/public/by-pages?projectId=draft&pageIds=a", 404},
		{"/api/v1/elements/public/by-pages?pageIds=a", 400},
	} {
		if got := status(t, app, tc.path); got != tc.want {
			t.Errorf("GET %s = %d, want %d", tc.path, got, tc.want)
		}
	}
	if len(spy.projectReads) != 1 || spy.projectReads[0] != "live" {
		t.Fatalf("public elements were read for %v, only \"live\" may be read", spy.projectReads)
	}
}

func TestPageIDListIsBounded(t *testing.T) {
	app := elementAccessApp(&elementServiceSpy{}, &projectsFake{}, "u1")
	ids := strings.TrimSuffix(strings.Repeat("p,", maxPageIDsPerRequest+1), ",")
	if got := status(t, app, "/api/v1/elements/by-pages?projectId=mine&pageIds="+ids); got != 400 {
		t.Fatalf("%d page ids = %d, want 400", maxPageIDsPerRequest+1, got)
	}
}
