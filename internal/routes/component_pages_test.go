package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/pkg/configs"
)

func componentPageTestApp(t *testing.T) (*fiber.App, string, *gorm.DB) {
	t.Helper()
	db, users := openLibraryTestDB(t)
	owner := users[0]
	projectID := owner + "-project"
	if err := db.Exec(`INSERT INTO "Project" ("Id", "Name", "OwnerId", "Published", "CreatedAt", "UpdatedAt") VALUES (?, 'Site', ?, true, NOW(), NOW())`, projectID, owner).Error; err != nil {
		t.Fatal(err)
	}
	app := fiber.New(configs.FiberConfig())
	api := app.Group("/api/v1")
	editor := app.Group("/api/v1", func(c fiber.Ctx) error {
		c.Locals("userId", owner)
		return c.Next()
	})
	repos := &repositories.RepositoriesInterface{
		PageRepository:    repositories.NewPageRepository(db),
		ProjectRepository: repositories.NewProjectRepository(db),
		ElementRepository: repositories.NewElementRepository(db),
	}
	PageRoutes(api, editor, repos, nil)
	ProjectRoutes(api, editor, repos)
	ElementRoutes(api, editor, repos, nil)
	return app, projectID, db
}

func TestComponentPagesAreEditorOnly(t *testing.T) {
	app, projectID, _ := componentPageTestApp(t)
	request := func(method, path, body string, want int) []byte {
		t.Helper()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		return testRouteResponse(t, app, req, want)
	}

	private := "/api/v1/pages/" + projectID
	public := "/api/v1/pages/public/" + projectID
	var home, component models.Page
	for _, input := range []struct {
		body string
		page *models.Page
	}{
		{`{"name":"Home","type":"sp","styles":{}}`, &home},
		{`{"name":"Card","type":"cp","styles":{}}`, &component},
	} {
		if err := json.Unmarshal(request(http.MethodPost, private, input.body, 201), input.page); err != nil {
			t.Fatal(err)
		}
	}

	assertPages := func(path string, want ...models.Page) {
		t.Helper()
		var pages []models.Page
		if err := json.Unmarshal(request(http.MethodGet, path, "", 200), &pages); err != nil {
			t.Fatal(err)
		}
		if len(pages) != len(want) {
			t.Fatalf("%s returned %d pages, want %d", path, len(pages), len(want))
		}
		for i, page := range pages {
			if page.Id != want[i].Id || page.Type != want[i].Type {
				t.Fatalf("%s page %d = %s (%s), want %s (%s)", path, i, page.Id, page.Type, want[i].Id, want[i].Type)
			}
		}
	}

	assertPages(public, home)
	assertPages(private, home, component)
	assertPages("/api/v1/projects/"+projectID+"/pages", home, component)
	request(http.MethodGet, private+"/"+component.Id, "", 200)
	request(http.MethodGet, public+"/"+home.Id, "", 200)
	request(http.MethodGet, public+"/"+component.Id, "", 404)
}

func TestComponentPageUpdatesAndDeletion(t *testing.T) {
	app, projectID, db := componentPageTestApp(t)
	path := "/api/v1/pages/" + projectID
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{"name":"Card","type":"cp"}`))
	req.Header.Set("Content-Type", "application/json")
	var page models.Page
	if err := json.Unmarshal(testRouteResponse(t, app, req, 201), &page); err != nil {
		t.Fatal(err)
	}
	req = httptest.NewRequest(http.MethodPatch, path+"/"+page.Id, strings.NewReader(`{"name":"Pricing card","type":"cp"}`))
	req.Header.Set("Content-Type", "application/json")
	if err := json.Unmarshal(testRouteResponse(t, app, req, 200), &page); err != nil || page.Name != "Pricing card" || page.Type != "cp" {
		t.Fatalf("updated page = %+v, %v", page, err)
	}
	if err := db.Exec(`INSERT INTO "Element" ("Id", "PageId", "Type") VALUES ('root-' || ?, ?, 'Section')`, page.Id, page.Id).Error; err != nil {
		t.Fatal(err)
	}
	elements := "/api/v1/elements/public/" + projectID
	byPages := "/api/v1/elements/public/by-pages?projectId=" + projectID + "&pageIds=" + page.Id
	for _, route := range []string{elements, byPages} {
		body := testRouteResponse(t, app, httptest.NewRequest(http.MethodGet, route, nil), 200)
		if !strings.Contains(string(body), "root-"+page.Id) {
			t.Fatalf("master elements missing from %s: %s", route, body)
		}
	}
	testRouteResponse(t, app, httptest.NewRequest(http.MethodDelete, path+"/"+page.Id, nil), 204)
	for _, route := range []string{path, "/api/v1/projects/" + projectID + "/pages", elements, byPages} {
		body := testRouteResponse(t, app, httptest.NewRequest(http.MethodGet, route, nil), 200)
		if strings.Contains(string(body), page.Id) {
			t.Fatalf("deleted master returned by %s: %s", route, body)
		}
	}
}
