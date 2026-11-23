package handlers

import (
	"errors"
	"fmt"
	"testing"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"

	"my-go-app/internal/repositories"
)

func TestProjectAccessFiberError(t *testing.T) {
	cases := []struct {
		name        string
		err         error
		wantCode    int
		wantMessage string
	}{
		{"stranger", repositories.ErrProjectUnauthorized, 403, "You do not have access to this project"},
		{"wrapped stranger", fmt.Errorf("verify: %w", repositories.ErrProjectUnauthorized), 403, "You do not have access to this project"},
		{"missing", repositories.ErrProjectNotFound, 404, "Project not found"},
		{"database failure", errInternal, 500, "Failed to load"},
		{"no row is not an access error here", gorm.ErrRecordNotFound, 500, "Failed to load"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := projectAccessFiberError(tc.err, "Failed to load")
			var fe *fiber.Error
			if !errors.As(err, &fe) {
				t.Fatalf("got %T, want *fiber.Error", err)
			}
			if fe.Code != tc.wantCode || fe.Message != tc.wantMessage {
				t.Fatalf("got %d %q, want %d %q", fe.Code, fe.Message, tc.wantCode, tc.wantMessage)
			}
		})
	}
}
