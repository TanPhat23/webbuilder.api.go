package ids

import (
	"fmt"

	"github.com/google/uuid"
)

// NewV7 returns a time-ordered UUID or an error if secure randomness fails.
func NewV7() (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("generate UUIDv7: %w", err)
	}
	return id.String(), nil
}
