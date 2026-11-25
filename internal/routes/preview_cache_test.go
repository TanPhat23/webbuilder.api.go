package routes

import (
	"context"
	"io"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	mockrepo "my-go-app/test/mockrepo"
)

type previewElementRepo struct {
	repositories.ElementRepositoryInterface
	label string
}

func (r previewElementRepo) List(_ context.Context, _ string, _ ...string) ([]models.EditorElement, error) {
	return []models.EditorElement{&models.Element{Id: r.label + " elements"}}, nil
}

func TestPublicPreviewCacheAcrossAPIInstances(t *testing.T) {
	addr := os.Getenv("WEBBUILDER_TEST_REDIS_ADDR")
	if addr == "" {
		t.Skip("set WEBBUILDER_TEST_REDIS_ADDR to test against Redis")
	}
	first := redis.NewClient(&redis.Options{Addr: addr})
	second := redis.NewClient(&redis.Options{Addr: addr})
	t.Cleanup(func() { first.Close(); second.Close() })
	if err := first.Ping(context.Background()).Err(); err != nil {
		t.Fatal(err)
	}

	projectID := uuid.NewString()
	updatedAt := time.Now()
	published := true
	projects := mockrepo.NewMockProjectRepo().SetGetPublicByID(func(_ context.Context, id string) (*models.Project, error) {
		if !published || id != projectID {
			return nil, repositories.ErrProjectNotFound
		}
		return &models.Project{ID: id, Published: true, AuditFields: &models.AuditFields{UpdatedAt: updatedAt}}, nil
	})

	newAPI := func(client *redis.Client, label string) *fiber.App {
		app := fiber.New()
		pages := mockrepo.NewMockPageRepo().SetListByProject(func(_ context.Context, id string) ([]models.Page, error) {
			return []models.Page{{Name: label + " pages", ProjectId: id}}, nil
		})
		Register(app, &repositories.Repositories{
			ProjectRepository: projects,
			PageRepository:    pages,
			ElementRepository: previewElementRepo{label: label},
		}, nil, nil, client)
		return app
	}
	a, b := newAPI(first, "first"), newAPI(second, "second")

	check := func(app *fiber.App, path string, status int, body string) {
		t.Helper()
		res, err := app.Test(httptest.NewRequest("GET", path, nil))
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		data, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}
		if res.StatusCode != status || (body != "" && !strings.Contains(string(data), body)) {
			t.Fatalf("GET %s = %d %s, want %d %s", path, res.StatusCode, data, status, body)
		}
	}

	pages := "/api/v1/pages/public/" + projectID
	elements := "/api/v1/elements/public/" + projectID
	check(a, pages+"?previewCache=1", 200, `"first pages"`)
	check(b, pages+"?previewCache=1", 200, `"first pages"`)
	check(b, elements+"?previewCache=1", 200, `"second elements"`)
	check(b, pages, 200, `"second pages"`)
	check(b, "/api/v1/pages/"+projectID+"?previewCache=1", 401, "")

	keys, err := first.Keys(context.Background(), "webbuilder:preview:v1:*:"+projectID+":*").Result()
	if err != nil || len(keys) != 2 {
		t.Fatalf("want two preview keys, got %v (%v)", keys, err)
	}
	t.Cleanup(func() { first.Del(context.Background(), keys...) })
	for _, key := range keys {
		ttl, err := second.TTL(context.Background(), key).Result()
		if err != nil || ttl <= 0 || ttl > 30*time.Second {
			t.Fatalf("unexpected TTL %v (%v)", ttl, err)
		}
	}

	published = false
	check(b, pages+"?previewCache=1", 404, "")
	check(b, elements+"?previewCache=1", 404, "")
	published = true
	updatedAt = updatedAt.Add(time.Second)
	check(b, pages+"?previewCache=1", 200, `"second pages"`)

	second.Close()
	check(b, elements+"?previewCache=1", 200, `"second elements"`)
}
