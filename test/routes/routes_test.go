package routes

import (
	"regexp"
	"testing"

	"my-go-app/internal/repositories"
	"my-go-app/internal/routes"

	"github.com/gofiber/fiber/v3"
)

var paramRe = regexp.MustCompile(`:[A-Za-z]+`)

func registered(t *testing.T) map[string][]string {
	t.Helper()
	app := fiber.New()
	routes.SetupRoutes(app, &repositories.RepositoriesInterface{}, nil, nil)

	got := map[string][]string{}
	for _, r := range app.GetRoutes(true) {
		key := r.Method + " " + paramRe.ReplaceAllString(r.Path, ":p")
		got[key] = append(got[key], r.Path)
	}
	return got
}

// Frontend endpoints the editor depends on must be routed.
func TestEditorRoutesExist(t *testing.T) {
	got := registered(t)
	for _, want := range []string{
		"GET /api/v1/projects/:p",
		"GET /api/v1/projects/:p/pages",
		"GET /api/v1/projects/:p/access",
		"GET /api/v1/projects/:p/event-workflows",
		"GET /api/v1/element-event-workflows/page/:p",
		"GET /api/v1/event-workflows/:p",
		"GET /api/v1/collaborators/project/:p",
		"GET /api/v1/projects/:p/comments",
	} {
		if _, ok := got[want]; !ok {
			t.Errorf("route not registered: %s", want)
		}
	}
}

// Two routes with the same method and shape shadow each other: the first one
// registered wins and the second is unreachable.
func TestNoShadowedRoutes(t *testing.T) {
	for key, paths := range registered(t) {
		if len(paths) > 1 {
			t.Errorf("%s is registered more than once: %v", key, paths)
		}
	}
}
