package utils

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

type validationSample struct {
	Name   string `validate:"required"`
	Email  string `validate:"email"`
	Pass   string `validate:"min=8"`
	Short  string `validate:"max=3"`
	Code   string `validate:"len=4"`
	Site   string `validate:"url"`
	ID     string `validate:"uuid"`
	Kind   string `validate:"oneof=a b"`
	N1     int    `validate:"gt=5"`
	N2     int    `validate:"gte=5"`
	N3     int    `validate:"lt=5"`
	N4     int    `validate:"lte=5"`
	Alnum  string `validate:"alphanum"`
	Num    string `validate:"numeric"`
	Custom string `validate:"startswith=x"`
}

func TestNewValidationErrorHumanizesEveryKnownTag(t *testing.T) {
	bad := validationSample{
		Email: "nope", Pass: "abc", Short: "toolong", Code: "12", Site: "x", ID: "x", Kind: "z",
		N1: 1, N2: 1, N3: 9, N4: 9, Alnum: "a-b", Num: "abc", Custom: "y",
	}
	err := validator.New().Struct(bad)
	var verrs validator.ValidationErrors
	if !errors.As(err, &verrs) {
		t.Fatalf("expected validation errors, got %v", err)
	}

	got := NewValidationError(verrs)
	want := map[string][2]string{
		"Name":   {"required", "Name is required"},
		"Email":  {"email", "Email must be a valid email address"},
		"Pass":   {"min", "Pass must be at least 8 characters long"},
		"Short":  {"max", "Short must be at most 3 characters long"},
		"Code":   {"len", "Code must be exactly 4 characters long"},
		"Site":   {"url", "Site must be a valid URL"},
		"ID":     {"uuid", "ID must be a valid UUID"},
		"Kind":   {"oneof", "Kind must be one of: a b"},
		"N1":     {"gt", "N1 must be greater than 5"},
		"N2":     {"gte", "N2 must be greater than or equal to 5"},
		"N3":     {"lt", "N3 must be less than 5"},
		"N4":     {"lte", "N4 must be less than or equal to 5"},
		"Alnum":  {"alphanum", "Alnum must contain only alphanumeric characters"},
		"Num":    {"numeric", "Num must be a numeric value"},
		"Custom": {"startswith", "Custom failed validation on rule 'startswith'"},
	}
	if len(got.Fields) != len(want) {
		t.Fatalf("got %d fields, want %d: %+v", len(got.Fields), len(want), got.Fields)
	}
	for _, f := range got.Fields {
		w, ok := want[f.Field]
		if !ok {
			t.Errorf("unexpected field %q", f.Field)
			continue
		}
		if f.Tag != w[0] || f.Message != w[1] {
			t.Errorf("%s: got tag=%q message=%q, want tag=%q message=%q", f.Field, f.Tag, f.Message, w[0], w[1])
		}
	}
	if got.Error() != "validation failed" {
		t.Errorf("Error() = %q", got.Error())
	}
}

func TestNewValidationErrorWithNoErrorsHasEmptyFields(t *testing.T) {
	got := NewValidationError(nil)
	if got.Fields == nil || len(got.Fields) != 0 {
		t.Fatalf("want empty non-nil Fields (serialises as []), got %#v", got.Fields)
	}
}

func TestHandleRepoError(t *testing.T) {
	cases := []struct {
		name         string
		err          error
		notFoundMsg  string
		wantStatus   int
		wantError    string
		wantMessages bool // whether the raw error text is echoed to the client
	}{
		{"gorm record not found", gorm.ErrRecordNotFound, "User not found", http.StatusNotFound, "User not found", true},
		{"wrapped gorm not found", errors.Join(errors.New("ctx"), gorm.ErrRecordNotFound), "User not found", http.StatusNotFound, "User not found", true},
		{"message mentions not found", errors.New("project not found"), "Project missing", http.StatusNotFound, "Project missing", true},
		{"not found check disabled", gorm.ErrRecordNotFound, "", http.StatusInternalServerError, "Failed", false},
		{"other error", errors.New("connection refused"), "User not found", http.StatusInternalServerError, "Failed", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/", func(c fiber.Ctx) error {
				return RespondError(c, tc.err, tc.notFoundMsg, "Failed")
			})
			status, body := call(t, app, "GET", "/")
			if status != tc.wantStatus {
				t.Fatalf("status = %d, want %d (%s)", status, tc.wantStatus, body)
			}
			m := decodeMap(t, body)
			if m["error"] != tc.wantError {
				t.Errorf("error = %v, want %q", m["error"], tc.wantError)
			}
			_, hasMessage := m["message"]
			if hasMessage != tc.wantMessages {
				t.Errorf("message present = %v, want %v (%s)", hasMessage, tc.wantMessages, body)
			}
			if strings.Contains(body, "connection refused") {
				t.Errorf("internal error text leaked to the client: %s", body)
			}
		})
	}
}

func TestHandleRepoErrorNilErrorIsNoop(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		if err := RespondError(c, nil, "nf", "fail"); err != nil {
			return err
		}
		return c.SendString("untouched")
	})
	if status, body := call(t, app, "GET", "/"); status != 200 || body != "untouched" {
		t.Fatalf("nil error must not write a response, got %d %q", status, body)
	}
}
