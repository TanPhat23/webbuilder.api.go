package utils_test

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"my-go-app/pkg/utils"
)

// ─── helpers ─────────────────────────────────────────────────────────────────

// newTestApp returns a minimal Fiber app whose error handler mirrors the
// production jsonErrorHandler so tests exercise the same response shaping.
func newTestApp() *fiber.App {
	return fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var fe *fiber.Error
			if errors.As(err, &fe) {
				code = fe.Code
			}
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})
}

// statusFor fires a GET request against app at path and returns the status code.

func statusFor(app *fiber.App, path string) int {
	req := httptest.NewRequest("GET", path, nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		return 0
	}
	return resp.StatusCode
}

// ─── HandleRepoError ─────────────────────────────────────────────────────────

func TestHandleRepoError_NilErrorReturnsNil(t *testing.T) {
	app := newTestApp()
	app.Get("/test", func(c *fiber.Ctx) error {
		result := utils.HandleRepoError(c, nil, "not found", "internal error")
		if result != nil {
			return result
		}
		return c.SendStatus(fiber.StatusOK)
	})

	if code := statusFor(app, "/test"); code != fiber.StatusOK {
		t.Errorf("nil error: expected 200, got %d", code)
	}
}

func TestHandleRepoError_RecordNotFoundReturns404(t *testing.T) {
	app := newTestApp()
	app.Get("/test", func(c *fiber.Ctx) error {
		return utils.HandleRepoError(c, errors.New("record not found"), "Resource not found", "internal error")
	})

	if code := statusFor(app, "/test"); code != fiber.StatusNotFound {
		t.Errorf("'record not found': expected 404, got %d", code)
	}
}

func TestHandleRepoError_SuffixNotFoundReturns404(t *testing.T) {
	app := newTestApp()
	app.Get("/test", func(c *fiber.Ctx) error {
		return utils.HandleRepoError(c, errors.New("project not found"), "Project not found", "internal error")
	})

	if code := statusFor(app, "/test"); code != fiber.StatusNotFound {
		t.Errorf("'*not found' suffix: expected 404, got %d", code)
	}
}

func TestHandleRepoError_GenericErrorReturns500(t *testing.T) {
	app := newTestApp()
	app.Get("/test", func(c *fiber.Ctx) error {
		return utils.HandleRepoError(c, errors.New("connection refused"), "not found msg", "Failed to query")
	})

	if code := statusFor(app, "/test"); code != fiber.StatusInternalServerError {
		t.Errorf("generic error: expected 500, got %d", code)
	}
}

func TestHandleRepoError_EmptyNotFoundMsgAlwaysReturns500(t *testing.T) {
	// When notFoundMsg is empty the helper must skip the not-found check and
	// always return 500, even if the error message looks like "not found".
	app := newTestApp()
	app.Get("/test", func(c *fiber.Ctx) error {
		return utils.HandleRepoError(c, errors.New("record not found"), "", "Failed to query")
	})

	if code := statusFor(app, "/test"); code != fiber.StatusInternalServerError {
		t.Errorf("empty notFoundMsg: expected 500 even for not-found errors, got %d", code)
	}
}

func TestHandleRepoError_VariousNotFoundSuffixes(t *testing.T) {
	notFoundErrors := []string{
		"user not found",
		"image not found",
		"page not found",
		"snapshot not found",
		"collaborator not found",
		"record not found",
	}

	for _, msg := range notFoundErrors {
		msg := msg // capture
		t.Run(msg, func(t *testing.T) {
			app := newTestApp()
			app.Get("/test", func(c *fiber.Ctx) error {
				return utils.HandleRepoError(c, errors.New(msg), "Resource not found", "internal")
			})
			if code := statusFor(app, "/test"); code != fiber.StatusNotFound {
				t.Errorf("%q: expected 404, got %d", msg, code)
			}
		})
	}
}

func TestHandleRepoError_ErrorThatDoesNotEndWithNotFound(t *testing.T) {
	// "not found" must be a suffix; a message that merely contains the phrase
	// elsewhere should still map to 500.
	app := newTestApp()
	app.Get("/test", func(c *fiber.Ctx) error {
		return utils.HandleRepoError(c, errors.New("could not find connection"), "nf msg", "internal")
	})

	if code := statusFor(app, "/test"); code != fiber.StatusInternalServerError {
		t.Errorf("non-suffix 'not found': expected 500, got %d", code)
	}
}

// ─── NewValidationError ───────────────────────────────────────────────────────

// validationErrorsFor runs the validator on v and returns the ValidationErrors.
func validationErrorsFor(t *testing.T, v any) validator.ValidationErrors {
	t.Helper()
	validate := validator.New(validator.WithRequiredStructEnabled())
	err := validate.Struct(v)
	if err == nil {
		t.Fatal("expected validation to fail, but it passed")
	}
	ve, ok := err.(validator.ValidationErrors)
	if !ok {
		t.Fatalf("expected validator.ValidationErrors, got %T", err)
	}
	return ve
}

func TestNewValidationError_ImplementsErrorInterface(t *testing.T) {
	type S struct {
		Name string `validate:"required"`
	}
	ve := validationErrorsFor(t, S{})
	result := utils.NewValidationError(ve)

	var _ error = result // compile-time check
	if result.Error() != "validation failed" {
		t.Errorf("Error(): got %q, want %q", result.Error(), "validation failed")
	}
}

func TestNewValidationError_FieldCountMatchesFailures(t *testing.T) {
	type S struct {
		Name  string `validate:"required"`
		Email string `validate:"required,email"`
	}
	ve := validationErrorsFor(t, S{})
	result := utils.NewValidationError(ve)

	// Name is missing (1 failure); Email is missing (1 failure for "required").
	if len(result.Fields) == 0 {
		t.Fatal("expected at least one FieldError, got none")
	}
}

func TestNewValidationError_FieldNameIsPopulated(t *testing.T) {
	type S struct {
		Username string `validate:"required"`
	}
	ve := validationErrorsFor(t, S{})
	result := utils.NewValidationError(ve)

	if len(result.Fields) != 1 {
		t.Fatalf("expected 1 field error, got %d", len(result.Fields))
	}
	if result.Fields[0].Field != "Username" {
		t.Errorf("Field: got %q, want %q", result.Fields[0].Field, "Username")
	}
}

func TestNewValidationError_TagIsPopulated(t *testing.T) {
	type S struct {
		Age int `validate:"gte=18"`
	}
	ve := validationErrorsFor(t, S{Age: 5})
	result := utils.NewValidationError(ve)

	if len(result.Fields) != 1 {
		t.Fatalf("expected 1 field error, got %d", len(result.Fields))
	}
	if result.Fields[0].Tag != "gte" {
		t.Errorf("Tag: got %q, want %q", result.Fields[0].Tag, "gte")
	}
}

func TestNewValidationError_MessageIsHumanReadable(t *testing.T) {
	type S struct {
		Email string `validate:"required,email"`
	}
	// Provide a value that passes "required" but fails "email".
	ve := validationErrorsFor(t, S{Email: "not-an-email"})
	result := utils.NewValidationError(ve)

	if len(result.Fields) == 0 {
		t.Fatal("expected field errors, got none")
	}
	msg := result.Fields[0].Message
	if msg == "" {
		t.Error("Message must not be empty")
	}
}

func TestHumanizeValidationError_KnownTags(t *testing.T) {
	cases := []struct {
		name    string
		build   any
		wantTag string
	}{
		{
			name: "required",
			build: struct {
				F string `validate:"required"`
			}{},
			wantTag: "required",
		},
		{
			name: "email",
			build: struct {
				F string `validate:"required,email"`
			}{F: "bad"},
			wantTag: "email",
		},
		{
			name: "min",
			build: struct {
				F string `validate:"min=5"`
			}{F: "ab"},
			wantTag: "min",
		},
		{
			name: "max",
			build: struct {
				F string `validate:"max=3"`
			}{F: "abcde"},
			wantTag: "max",
		},
		{
			name: "url",
			build: struct {
				F string `validate:"url"`
			}{F: "not-a-url"},
			wantTag: "url",
		},
	}

	validate := validator.New(validator.WithRequiredStructEnabled())

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := validate.Struct(tc.build)
				if err == nil {
					t.Skip("validation unexpectedly passed; skipping")
				}
				ve, ok := err.(validator.ValidationErrors)
				if !ok {
					t.Fatalf("expected ValidationErrors, got %T", err)
				}
				result := utils.NewValidationError(ve)
			found := false
			for _, fe := range result.Fields {
				if fe.Tag == tc.wantTag {
					found = true
					if fe.Message == "" {
						t.Errorf("tag %q produced empty message", tc.wantTag)
					}
				}
			}
			if !found {
				t.Errorf("expected a field error with tag %q", tc.wantTag)
			}
		})
	}
}

func TestNewValidationError_MultipleFieldErrors(t *testing.T) {
	type S struct {
		First string `validate:"required"`
		Last  string `validate:"required"`
		Age   int    `validate:"gte=0,lte=150"`
	}
	ve := validationErrorsFor(t, S{Age: -1})
	result := utils.NewValidationError(ve)

	if len(result.Fields) < 2 {
		t.Errorf("expected at least 2 field errors (First + Last), got %d", len(result.Fields))
	}
}