package dto

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

func TestPageRequestUUIDValidationSupportsV4AndV7(t *testing.T) {
	validate := validator.New()
	for _, tc := range []struct {
		id    string
		valid bool
	}{{"", true}, {uuid.NewString(), true}, {uuid.Must(uuid.NewV7()).String(), true}, {"not-a-uuid", false}} {
		err := validate.Struct(CreatePageRequest{Id: tc.id, Name: "Page", Type: "page"})
		if (err == nil) != tc.valid {
			t.Errorf("id %q: unexpected validation result %v", tc.id, err)
		}
	}
}
