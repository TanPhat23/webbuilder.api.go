package configs

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"

	"my-go-app/pkg/utils"
)

func TestCorsConfigUsesEnvOverridesAndDefaults(t *testing.T) {
	t.Setenv("ALLOWED_ORIGINS", "")
	def := CorsConfig()
	if !def.AllowCredentials || len(def.AllowOrigins) != 3 {
		t.Fatalf("unexpected default cors config: %+v", def)
	}

	t.Setenv("ALLOWED_ORIGINS", " https://a.example.com , https://b.example.com ")
	custom := CorsConfig()
	if len(custom.AllowOrigins) != 2 || custom.AllowOrigins[0] != "https://a.example.com" || custom.AllowOrigins[1] != "https://b.example.com" {
		t.Fatalf("unexpected trimmed cors origins: %+v", custom.AllowOrigins)
	}
}

func TestLoggerConfigAndErrorHandlerShapes(t *testing.T) {
	if LoggerConfig().Format == "" {
		t.Fatal("expected non-empty logger format")
	}

	app := fiber.New(FiberConfig())
	app.Get("/val", func(c fiber.Ctx) error {
		return &utils.ValidationError{
			Fields: []utils.FieldError{{Field: "name", Tag: "required", Message: "name is required"}},
		}
	})
	app.Get("/fiber", func(c fiber.Ctx) error {
		return fiber.NewError(fiber.StatusNotFound, "Missing item")
	})
	app.Get("/boom", func(c fiber.Ctx) error {
		return errors.New("secret db failure")
	})

	valRes, err := app.Test(httptest.NewRequest("GET", "/val", nil))
	if err != nil || valRes.StatusCode != 422 {
		t.Fatalf("validation status = %d err = %v", valRes.StatusCode, err)
	}
	var valBody map[string]any
	if err := json.NewDecoder(valRes.Body).Decode(&valBody); err != nil || valBody["error"] != "Validation failed" {
		t.Fatalf("unexpected validation payload: %v err=%v", valBody, err)
	}

	fiberRes, err := app.Test(httptest.NewRequest("GET", "/fiber", nil))
	if err != nil || fiberRes.StatusCode != 404 {
		t.Fatalf("fiber status = %d err = %v", fiberRes.StatusCode, err)
	}

	boomRes, err := app.Test(httptest.NewRequest("GET", "/boom", nil))
	if err != nil || boomRes.StatusCode != 500 {
		t.Fatalf("boom status = %d err = %v", boomRes.StatusCode, err)
	}
	var boomBody map[string]any
	if err := json.NewDecoder(boomRes.Body).Decode(&boomBody); err != nil || boomBody["error"] != "Internal server error" {
		t.Fatalf("unexpected boom payload: %v err=%v", boomBody, err)
	}
}
