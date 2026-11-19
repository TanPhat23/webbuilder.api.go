package routes

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/pkg/configs"
)

// UUID_GORM_TEST_DSN must point to a disposable, migrated PostgreSQL database.
func openLibraryTestDB(t *testing.T) (*gorm.DB, [2]string) {
	t.Helper()
	dsn := os.Getenv("UUID_GORM_TEST_DSN")
	if dsn == "" {
		t.Skip("requires dedicated migrated PostgreSQL database")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { sqlDB.Close() })
	users := [2]string{uuid.NewString(), uuid.NewString()}
	for _, id := range users {
		if err := db.Exec(`INSERT INTO "User" ("Id", "Email", "CreatedAt", "UpdatedAt") VALUES (?, ?, NOW(), NOW())`, id, id+"@example.invalid").Error; err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := db.Exec(`DELETE FROM "User" WHERE "Id" = ?`, id).Error; err != nil {
				t.Error(err)
			}
		})
	}
	return db, users
}

func testRouteResponse(t *testing.T, app *fiber.App, req *http.Request, want int) []byte {
	t.Helper()
	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != want {
		t.Fatalf("%s %s: got %d (%s), want %d", req.Method, req.URL, res.StatusCode, body, want)
	}
	return body
}

func TestCustomElementAccountLibraryCRUD(t *testing.T) {
	db, users := openLibraryTestDB(t)
	app := fiber.New(configs.FiberConfig())
	group := app.Group("/api/v1", func(c fiber.Ctx) error {
		c.Locals("userId", c.Get("Test-User"))
		return c.Next()
	})
	CustomElementRoutes(group, &repositories.RepositoriesInterface{
		CustomElementRepository: repositories.NewCustomElementRepository(db),
	})

	const path = "/api/v1/custom-elements"
	request := func(method, path, user, body string, want int) []byte {
		t.Helper()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Test-User", user)
		return testRouteResponse(t, app, req, want)
	}
	readElement := func(body []byte) models.CustomElement {
		t.Helper()
		var element models.CustomElement
		if err := json.Unmarshal(body, &element); err != nil {
			t.Fatal(err)
		}
		return element
	}
	create := func(user, name string, public bool) models.CustomElement {
		t.Helper()
		input := map[string]any{
			"name": name, "description": "Saved card",
			"structure": json.RawMessage(`{"type":"Section","children":[{"type":"Text","content":"Hello"}]}`),
			"isPublic":  public, "version": "2.0.0", "userId": "ignored",
		}
		body, err := json.Marshal(input)
		if err != nil {
			t.Fatal(err)
		}
		element := readElement(request(http.MethodPost, path, user, string(body), 201))
		if element.Id == "" || element.UserId != user || element.Name != name || element.IsPublic != public ||
			element.Version != "2.0.0" || element.Description == nil || *element.Description != "Saved card" ||
			!strings.Contains(string(element.Structure), "Hello") {
			t.Fatalf("unexpected snapshot: %+v", element)
		}
		return element
	}
	checkList := func(query, user string, wantIDs ...string) {
		t.Helper()
		body := request(http.MethodGet, path+query, user, "", 200)
		var items []models.CustomElement
		if err := json.Unmarshal(body, &items); err != nil || items == nil || len(items) != len(wantIDs) {
			t.Fatalf("unexpected list: %s (%v)", body, err)
		}
		for i, id := range wantIDs {
			if items[i].Id != id || items[i].UserId != user {
				t.Fatalf("unexpected list: %s", body)
			}
		}
	}

	request(http.MethodGet, path, "", "", 401)
	private := create(users[0], "Card", false)
	other := create(users[1], "Card", true)
	ownPublic := create(users[0], "Navbar", true)
	checkList("", users[0], ownPublic.Id, private.Id)
	checkList("?isPublic=true", users[0], ownPublic.Id)
	checkList("?isPublic=false", users[0], private.Id)
	checkList("", users[1], other.Id)
	request(http.MethodPost, path, users[0], `{"name":"card","structure":{},"version":"1.0.0"}`, 409)
	request(http.MethodGet, path+"/"+private.Id, users[1], "", 404)
	request(http.MethodDelete, path+"/"+ownPublic.Id, users[1], "", 403)
	request(http.MethodPatch, path+"/"+ownPublic.Id, users[1], `{"name":"Another"}`, 403)
	request(http.MethodPatch, path+"/"+ownPublic.Id, users[0], `{"name":"CARD"}`, 409)
	updated := readElement(request(http.MethodPatch, path+"/"+ownPublic.Id, users[0],
		`{"name":"Header","isPublic":false,"structure":{"type":"Frame"}}`, 200))
	if updated.Name != "Header" || updated.IsPublic || !strings.Contains(string(updated.Structure), "Frame") {
		t.Fatalf("updated snapshot = %+v", updated)
	}
	got := readElement(request(http.MethodGet, path+"/"+updated.Id, users[0], "", 200))
	if got.Id != updated.Id || got.Name != "Header" || got.IsPublic {
		t.Fatalf("stored snapshot = %+v", got)
	}
	request(http.MethodDelete, path+"/"+private.Id, users[0], "", 200)
	request(http.MethodDelete, path+"/"+ownPublic.Id, users[0], "", 200)
	request(http.MethodGet, path+"/"+private.Id, users[0], "", 404)
	checkList("", users[0])
	checkList("", users[1], other.Id)
}
