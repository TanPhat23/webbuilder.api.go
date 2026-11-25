package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"

	"my-go-app/internal/repositories"
)

func TestHandleProjectErrorMapsAccessErrors(t *testing.T) {
	dbFailure := errors.New(`pq: connection to 10.0.0.5 refused for user "admin"`)
	cases := []struct {
		name       string
		err        error
		wantStatus int
		wantError  string
	}{
		{"stranger", repositories.ErrProjectUnauthorized, 403, "You do not have access to this project"},
		{"wrapped stranger", fmt.Errorf("failed to verify project: %w", repositories.ErrProjectUnauthorized), 403, "You do not have access to this project"},
		{"missing", repositories.ErrProjectNotFound, 404, "Project not found"},
		{"no row", gorm.ErrRecordNotFound, 404, "Project not found"},
		{"database failure", dbFailure, 500, "Failed to retrieve project"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/", func(c fiber.Ctx) error { return respondProjectError(c, tc.err, "Failed to retrieve project") })
			res, err := app.Test(httptest.NewRequest(http.MethodGet, "/", nil))
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()
			raw, _ := io.ReadAll(res.Body)
			if res.StatusCode != tc.wantStatus {
				t.Fatalf("status = %d, want %d (%s)", res.StatusCode, tc.wantStatus, raw)
			}
			var body struct{ Error, Message string }
			if err := json.Unmarshal(raw, &body); err != nil {
				t.Fatal(err)
			}
			if body.Error != tc.wantError {
				t.Errorf("error = %q, want %q", body.Error, tc.wantError)
			}
			if tc.wantStatus >= 500 && (body.Message != "" || strings.Contains(string(raw), "10.0.0.5")) {
				t.Errorf("a 5xx body must not expose the cause: %s", raw)
			}
		})
	}
}
