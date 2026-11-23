package utils

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
)

func TestSendErrorClientErrorIncludesDetail(t *testing.T) {
	app := fiber.New()
	app.Use(requestid.New())
	app.Get("/", func(c fiber.Ctx) error {
		return SendError(c, fiber.StatusBadRequest, "Bad input", errors.New("field x invalid"), "user-7")
	})
	status, body := call(t, app, "GET", "/")
	if status != 400 {
		t.Fatalf("status = %d", status)
	}
	m := decodeMap(t, body)
	if m["error"] != "Bad input" || m["message"] != "field x invalid" || m["userId"] != "user-7" {
		t.Errorf("unexpected body %s", body)
	}
	if id, _ := m["requestId"].(string); id == "" {
		t.Errorf("requestId missing: %s", body)
	}
}

func TestSendErrorServerErrorHidesInternals(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		return SendError(c, fiber.StatusInternalServerError, "Failed", errors.New("pq: password authentication failed"))
	})
	status, body := call(t, app, "GET", "/")
	if status != 500 {
		t.Fatalf("status = %d", status)
	}
	if strings.Contains(body, "password") {
		t.Fatalf("5xx must not echo the underlying error: %s", body)
	}
	m := decodeMap(t, body)
	if _, ok := m["message"]; ok {
		t.Errorf("message field should be omitted: %s", body)
	}
	if _, ok := m["userId"]; ok {
		t.Errorf("userId should be omitted when not supplied: %s", body)
	}
	if _, ok := m["requestId"]; ok {
		t.Errorf("requestId should be omitted without the middleware: %s", body)
	}
}

func TestSendErrorNilError(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		return SendError(c, fiber.StatusNotFound, "Missing", nil)
	})
	status, body := call(t, app, "GET", "/")
	m := decodeMap(t, body)
	if status != 404 || m["error"] != "Missing" {
		t.Fatalf("got %d %s", status, body)
	}
	if _, ok := m["message"]; ok {
		t.Errorf("no error means no message: %s", body)
	}
}

func TestLogRequestErrorLevels(t *testing.T) {
	cases := []struct {
		name      string
		status    int
		err       error
		locals    any
		wantLevel string
		wantMsg   string
		wantUser  bool
		wantErr   bool
	}{
		{"5xx is error level", 503, errors.New("boom"), "u1", "ERROR", "request failed", true, true},
		{"4xx is warn level", 404, nil, nil, "WARN", "request rejected", false, false},
		{"empty user id omitted", 400, errors.New("bad"), "", "WARN", "request rejected", false, true},
		{"non-string user id omitted", 400, nil, 42, "WARN", "request rejected", false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			buf := captureLogs(t)
			app := fiber.New()
			app.Get("/things", func(c fiber.Ctx) error {
				if tc.locals != nil {
					c.Locals("userId", tc.locals)
				}
				LogRequestError(c, tc.status, "msg", tc.err)
				return c.SendStatus(http.StatusOK)
			})
			call(t, app, "GET", "/things")
			m := decodeMap(t, strings.TrimSpace(buf.String()))
			if m["level"] != tc.wantLevel || m["msg"] != tc.wantMsg {
				t.Errorf("level/msg = %v/%v, want %s/%s", m["level"], m["msg"], tc.wantLevel, tc.wantMsg)
			}
			if m["method"] != "GET" || m["path"] != "/things" || m["message"] != "msg" || m["status"] != float64(tc.status) {
				t.Errorf("request attributes wrong: %s", buf.String())
			}
			if _, ok := m["userId"]; ok != tc.wantUser {
				t.Errorf("userId present = %v, want %v", ok, tc.wantUser)
			}
			if _, ok := m["error"]; ok != tc.wantErr {
				t.Errorf("error present = %v, want %v", ok, tc.wantErr)
			}
		})
	}
}

func TestSendSuccessAndHelpers(t *testing.T) {
	app := fiber.New()
	app.Get("/with", func(c fiber.Ctx) error {
		return SendSuccess(c, fiber.StatusCreated, "Created", map[string]int{"n": 1})
	})
	app.Get("/without", func(c fiber.Ctx) error { return SendSuccess(c, fiber.StatusOK, "Done") })
	app.Get("/json", func(c fiber.Ctx) error { return SendJSON(c, fiber.StatusAccepted, []int{1, 2}) })
	app.Get("/empty", func(c fiber.Ctx) error { return SendNoContent(c) })

	status, body := call(t, app, "GET", "/with")
	m := decodeMap(t, body)
	if status != 201 || m["message"] != "Created" || m["data"].(map[string]any)["n"] != float64(1) {
		t.Errorf("with data: %d %s", status, body)
	}

	status, body = call(t, app, "GET", "/without")
	m = decodeMap(t, body)
	if _, ok := m["data"]; status != 200 || ok || m["message"] != "Done" {
		t.Errorf("without data: %d %s", status, body)
	}

	if status, body = call(t, app, "GET", "/json"); status != 202 || body != "[1,2]" {
		t.Errorf("SendJSON: %d %s", status, body)
	}
	if status, body = call(t, app, "GET", "/empty"); status != 204 || body != "" {
		t.Errorf("SendNoContent: %d %q", status, body)
	}
}

func TestRequestIDWithAndWithoutMiddleware(t *testing.T) {
	plain := fiber.New()
	plain.Get("/", func(c fiber.Ctx) error { return c.SendString("[" + RequestID(c) + "]") })
	if _, body := call(t, plain, "GET", "/"); body != "[]" {
		t.Errorf("without middleware want empty id, got %q", body)
	}

	with := fiber.New()
	with.Use(requestid.New())
	with.Get("/", func(c fiber.Ctx) error { return c.SendString(RequestID(c)) })
	if _, body := call(t, with, "GET", "/"); body == "" {
		t.Error("with middleware want a generated id")
	}
}
