package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"

	"my-go-app/pkg/configs"
)

func init() {
	// The handlers log every rejected request; keep test output readable.
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
}

// errInternal looks like a driver failure that carries infrastructure details.
// Handlers must turn it into a generic 5xx and never echo it to the client.
var errInternal = errors.New(`pq: connection to 10.0.0.5 refused for user "admin"`)

const leakMarker = "10.0.0.5"

// newKitApp builds a Fiber app configured like production (JSON error handler
// and struct validator). A non-empty userID is installed as the authenticated
// user, the way the auth middleware does.
func newKitApp(userID string, register func(app *fiber.App)) *fiber.App {
	app := fiber.New(configs.FiberConfig())
	app.Use(func(c fiber.Ctx) error {
		if userID != "" {
			c.Locals("userId", userID)
		}
		return c.Next()
	})
	register(app)
	return app
}

type kitReply struct {
	status int
	body   string
}

func (r kitReply) decode(t *testing.T, into any) {
	t.Helper()
	if err := json.Unmarshal([]byte(r.body), into); err != nil {
		t.Fatalf("response is not the expected JSON (%v): %s", err, r.body)
	}
}

func (r kitReply) object(t *testing.T) map[string]any {
	t.Helper()
	var m map[string]any
	r.decode(t, &m)
	return m
}

// kitSend performs the request; a non-empty body is sent as JSON.
func kitSend(t *testing.T, app *fiber.App, method, target, body string) kitReply {
	t.Helper()
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, target, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	return kitReply{status: res.StatusCode, body: string(raw)}
}

// kitExpect asserts the status and that the body contains every fragment.
// Any 5xx body is additionally checked for leaked internals.
//
// status 0 means "the request must be rejected" (any 4xx/5xx). It is used for
// bodies that fail to parse or validate: the handlers return the binder's error
// unchanged and the production error handler currently renders it as a 500
// (see the report), so pinning the exact code here would enshrine that bug.
func kitExpect(t *testing.T, r kitReply, status int, fragments ...string) {
	t.Helper()
	if status == 0 {
		if r.status < 400 {
			t.Fatalf("status = %d, want a rejection (body %s)", r.status, r.body)
		}
	} else if r.status != status {
		t.Fatalf("status = %d, want %d (body %s)", r.status, status, r.body)
	}
	for _, f := range fragments {
		if !strings.Contains(r.body, f) {
			t.Errorf("body %s does not contain %q", r.body, f)
		}
	}
	if strings.Contains(r.body, leakMarker) && r.status >= 500 {
		t.Errorf("5xx body leaks internal error: %s", r.body)
	}
}

func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool    { return &b }
