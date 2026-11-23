package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestAppErrorError(t *testing.T) {
	tests := []struct {
		name string
		err  *AppError
		want string
	}{
		{"without cause", &AppError{Code: CodeNotFound, Message: "no page"}, "[NOT_FOUND] no page"},
		{"with cause", &AppError{Code: CodeInternal, Message: "db down", Err: errors.New("conn refused")}, "[INTERNAL_ERROR] db down: conn refused"},
	}
	for _, tt := range tests {
		if got := tt.err.Error(); got != tt.want {
			t.Errorf("%s: got %q want %q", tt.name, got, tt.want)
		}
	}
}

func TestNewAndUnwrap(t *testing.T) {
	cause := errors.New("cause")
	e := New(CodeValidation, "bad input", cause)
	if e.Code != CodeValidation || e.Message != "bad input" || e.Err != cause {
		t.Fatalf("unexpected %+v", e)
	}
	if e.Unwrap() != cause {
		t.Fatal("Unwrap must return the cause")
	}
	if !errors.Is(e, cause) {
		t.Fatal("errors.Is must reach the cause")
	}
	if New(CodeConflict, "m", nil).Unwrap() != nil {
		t.Fatal("nil cause must unwrap to nil")
	}
}

func TestWrap(t *testing.T) {
	if got := Wrap(CodeInternal, nil); got != nil {
		t.Fatalf("Wrap(nil) must be nil, got %+v", got)
	}
	cause := fmt.Errorf("ctx: %w", ErrNotFound)
	w := Wrap(CodeNotFound, cause)
	if w.Code != CodeNotFound || w.Message != cause.Error() {
		t.Fatalf("unexpected %+v", w)
	}
	if !errors.Is(w, ErrNotFound) {
		t.Fatal("sentinel must stay reachable through Wrap")
	}
}

func TestIsCode(t *testing.T) {
	app := New(CodeUnauthorized, "nope", nil)
	wrapped := fmt.Errorf("handler: %w", app)
	tests := []struct {
		name string
		err  error
		code string
		want bool
	}{
		{"direct match", app, CodeUnauthorized, true},
		{"wrapped match", wrapped, CodeUnauthorized, true},
		{"different code", app, CodeNotFound, false},
		{"non AppError", ErrNotFound, CodeNotFound, false},
		{"nil error", nil, CodeNotFound, false},
		{"empty code vs non-empty", app, "", false},
	}
	for _, tt := range tests {
		if got := IsCode(tt.err, tt.code); got != tt.want {
			t.Errorf("%s: got %v want %v", tt.name, got, tt.want)
		}
	}
}

func TestSentinelsAreDistinct(t *testing.T) {
	all := []error{ErrNotFound, ErrValidation, ErrInternal, ErrConflict, ErrUnauthorized, ErrNotJoined, ErrInvalidProject}
	for i, a := range all {
		for j, b := range all {
			if (i == j) != errors.Is(a, b) {
				t.Errorf("sentinel %d vs %d: errors.Is=%v", i, j, errors.Is(a, b))
			}
		}
	}
}
