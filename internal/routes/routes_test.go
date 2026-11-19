package routes

import (
	"strings"
	"testing"

	"my-go-app/internal/repositories"
	"my-go-app/internal/services"
	"my-go-app/pkg/configs"

	"github.com/gofiber/fiber/v3"
)

// buildApp registers every route against empty dependencies; no handler is invoked,
// so nothing is touched.
func buildApp() *fiber.App {
	app := fiber.New(configs.FiberConfig())
	SetupRoutes(app, repositories.NewRepositories(nil), &services.CloudinaryService{}, &services.InvitationService{}, nil)
	return app
}

func segments(path string) []string {
	return strings.FieldsFunc(path, func(r rune) bool { return r == '/' })
}

// covers reports whether the earlier pattern would match every URL the later
// pattern can receive, which makes the later route unreachable.
func covers(earlier, later string) bool {
	e, l := segments(earlier), segments(later)
	if len(e) != len(l) {
		return false
	}
	for i := range e {
		if strings.HasPrefix(e[i], ":") {
			continue
		}
		if e[i] != l[i] || strings.HasPrefix(l[i], ":") {
			return false
		}
	}
	return true
}

func TestNoRouteIsShadowed(t *testing.T) {
	routes := buildApp().GetRoutes()
	for i, r := range routes {
		if r.Method == "HEAD" || r.Method == "USE" {
			continue
		}
		for _, e := range routes[:i] {
			if e.Method == r.Method && covers(e.Path, r.Path) {
				t.Errorf("%s %s is unreachable: %s is registered first", r.Method, r.Path, e.Path)
				break
			}
		}
	}
}
