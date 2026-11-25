package utils

import (
	"errors"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func fiberErr(t *testing.T, err error, code int) *fiber.Error {
	t.Helper()
	var fe *fiber.Error
	if !errors.As(err, &fe) {
		t.Fatalf("want *fiber.Error, got %T (%v)", err, err)
	}
	if fe.Code != code {
		t.Fatalf("want status %d, got %d (%s)", code, fe.Code, fe.Message)
	}
	return fe
}

// paramApp registers /:projectid/:pageid and runs fn inside the handler,
// optionally after setting the userId local.
func paramApp(t *testing.T, userID any, fn func(c fiber.Ctx) error) *fiber.App {
	t.Helper()
	app := fiber.New()
	app.Get("/:projectid/:pageid", func(c fiber.Ctx) error {
		if userID != nil {
			c.Locals("userId", userID)
		}
		return fn(c)
	})
	return app
}

func TestValidateRequiredParam(t *testing.T) {
	var got string
	var gotErr error
	app := paramApp(t, nil, func(c fiber.Ctx) error {
		got, gotErr = RequireParam(c, "projectid")
		return c.SendStatus(http.StatusOK)
	})
	call(t, app, "GET", "/p1/pg1")
	if gotErr != nil || got != "p1" {
		t.Fatalf("present param: %q %v", got, gotErr)
	}

	app = paramApp(t, nil, func(c fiber.Ctx) error {
		got, gotErr = RequireParam(c, "nope")
		return c.SendStatus(http.StatusOK)
	})
	call(t, app, "GET", "/p1/pg1")
	if got != "" {
		t.Errorf("missing param should return empty string, got %q", got)
	}
	if fe := fiberErr(t, gotErr, 400); fe.Message != "nope is required" {
		t.Errorf("message = %q", fe.Message)
	}
}

func TestMustParams(t *testing.T) {
	var ids []string
	var err error
	app := paramApp(t, nil, func(c fiber.Ctx) error {
		ids, err = RequireParams(c, "pageid", "projectid")
		return c.SendStatus(http.StatusOK)
	})
	call(t, app, "GET", "/p1/pg1")
	if err != nil || len(ids) != 2 || ids[0] != "pg1" || ids[1] != "p1" {
		t.Fatalf("order must follow names: %v %v", ids, err)
	}

	app = paramApp(t, nil, func(c fiber.Ctx) error {
		ids, err = RequireParams(c, "projectid", "missing", "pageid")
		return c.SendStatus(http.StatusOK)
	})
	call(t, app, "GET", "/p1/pg1")
	if ids != nil {
		t.Errorf("no values on failure, got %v", ids)
	}
	if fe := fiberErr(t, err, 400); fe.Message != "missing is required" {
		t.Errorf("must report the first missing param: %q", fe.Message)
	}

	app = paramApp(t, nil, func(c fiber.Ctx) error {
		ids, err = RequireParams(c)
		return c.SendStatus(http.StatusOK)
	})
	call(t, app, "GET", "/p1/pg1")
	if err != nil || len(ids) != 0 {
		t.Errorf("no names: %v %v", ids, err)
	}
}

func TestValidateUserID(t *testing.T) {
	cases := []struct {
		name   string
		locals any
		want   string
		fail   bool
	}{
		{"set", "user_1", "user_1", false},
		{"unset", nil, "", true},
		{"empty", "", "", true},
		{"wrong type", 123, "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var got string
			var err error
			app := paramApp(t, tc.locals, func(c fiber.Ctx) error {
				got, err = RequireUserID(c)
				return c.SendStatus(http.StatusOK)
			})
			call(t, app, "GET", "/a/b")
			if tc.fail {
				fiberErr(t, err, 401)
				if got != "" {
					t.Errorf("got %q on failure", got)
				}
				return
			}
			if err != nil || got != tc.want {
				t.Errorf("got %q %v", got, err)
			}
		})
	}
}

func TestMustUserAndParams(t *testing.T) {
	var user string
	var ids []string
	var err error
	run := func(locals any, names ...string) {
		app := paramApp(t, locals, func(c fiber.Ctx) error {
			user, ids, err = RequireUserAndParams(c, names...)
			return c.SendStatus(http.StatusOK)
		})
		call(t, app, "GET", "/p1/pg1")
	}

	run("u1", "projectid", "pageid")
	if err != nil || user != "u1" || len(ids) != 2 || ids[0] != "p1" || ids[1] != "pg1" {
		t.Fatalf("happy path: %q %v %v", user, ids, err)
	}

	// The user check runs first, so an unauthenticated caller gets 401 even
	// when a param is also missing.
	run(nil, "missing")
	fiberErr(t, err, 401)
	if user != "" || ids != nil {
		t.Errorf("no partial results on failure: %q %v", user, ids)
	}

	run("u1", "missing")
	fiberErr(t, err, 400)
	if user != "" || ids != nil {
		t.Errorf("no partial results on failure: %q %v", user, ids)
	}
}

func TestRequireUpdates(t *testing.T) {
	fiberErr(t, RequireUpdates(nil), 400)
	fiberErr(t, RequireUpdates(map[string]any{}), 400)
	if err := RequireUpdates(map[string]any{"Name": "x"}); err != nil {
		t.Errorf("non-empty updates rejected: %v", err)
	}
}

type validateTarget struct {
	Name  string `validate:"required"`
	Email string `validate:"required,email"`
}

func TestValidateStruct(t *testing.T) {
	if err := ValidateStruct(validateTarget{Name: "a", Email: "a@b.co"}); err != nil {
		t.Fatalf("valid struct rejected: %v", err)
	}
	if err := ValidateStruct(&validateTarget{Name: "a", Email: "a@b.co"}); err != nil {
		t.Fatalf("valid pointer rejected: %v", err)
	}

	err := ValidateStruct(validateTarget{Email: "not-an-email"})
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("want *ValidationError, got %T (%v)", err, err)
	}
	if len(ve.Fields) != 2 || ve.Fields[0].Field != "Name" || ve.Fields[0].Tag != "required" || ve.Fields[1].Field != "Email" || ve.Fields[1].Tag != "email" {
		t.Errorf("unexpected fields: %+v", ve.Fields)
	}

	// validator returns InvalidValidationError (not ValidationErrors) for
	// non-struct input; that must surface as a 400 rather than a panic.
	fiberErr(t, ValidateStruct("just a string"), 400)
	fiberErr(t, ValidateStruct(nil), 400)
}

func TestValidateCollaboratorRole(t *testing.T) {
	for _, role := range []string{"owner", "editor", "viewer"} {
		if err := ValidateCollaboratorRole(role); err != nil {
			t.Errorf("%s rejected: %v", role, err)
		}
	}
	for _, role := range []string{"", "admin", "Owner", " editor"} {
		fe := fiberErr(t, ValidateCollaboratorRole(role), 400)
		if fe.Message != "Invalid collaborator role: "+role {
			t.Errorf("role %q message = %q", role, fe.Message)
		}
	}
}

func TestValidateCollaboratorID(t *testing.T) {
	if err := ValidateCollaboratorID("c1"); err != nil {
		t.Errorf("non-empty id rejected: %v", err)
	}
	fiberErr(t, ValidateCollaboratorID(""), 400)
}
