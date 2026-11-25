package configs

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"

	"my-go-app/pkg/utils"
)

func TestCorsConfigUsesEnvOverridesAndDefaults(t *testing.T) {
	t.Setenv("ENV", "development")
	t.Setenv("ALLOWED_ORIGINS", "")
	def := CorsConfig()
	if !def.AllowCredentials || len(def.AllowOrigins) != 2 {
		t.Fatalf("unexpected default cors config: %+v", def)
	}

	t.Setenv("ALLOWED_ORIGINS", " https://a.example.com , https://b.example.com ")
	custom := CorsConfig()
	if len(custom.AllowOrigins) != 2 || custom.AllowOrigins[0] != "https://a.example.com" || custom.AllowOrigins[1] != "https://b.example.com" {
		t.Fatalf("unexpected trimmed cors origins: %+v", custom.AllowOrigins)
	}

	for _, env := range []string{"production", "prod"} {
		t.Setenv("ENV", env)
		t.Setenv("ALLOWED_ORIGINS", "")
		if _, err := AllowedOrigins(); err == nil {
			t.Fatalf("expected AllowedOrigins error in %s with empty ALLOWED_ORIGINS", env)
		}
		t.Setenv("ALLOWED_ORIGINS", "https://app.example.com")
		origins, err := AllowedOrigins()
		if err != nil || len(origins) != 1 || origins[0] != "https://app.example.com" {
			t.Fatalf("unexpected AllowedOrigins in %s: origins=%v err=%v", env, origins, err)
		}
	}
}

func TestFiberConfigHardeningDefaults(t *testing.T) {
	cfg := FiberConfig()
	if cfg.ReadTimeout != 15*time.Second || cfg.WriteTimeout != 30*time.Second || cfg.IdleTimeout != 60*time.Second {
		t.Fatalf("unexpected timeouts: read=%v write=%v idle=%v", cfg.ReadTimeout, cfg.WriteTimeout, cfg.IdleTimeout)
	}
	if cfg.BodyLimit != 4*1024*1024 {
		t.Fatalf("unexpected body limit: %d", cfg.BodyLimit)
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
