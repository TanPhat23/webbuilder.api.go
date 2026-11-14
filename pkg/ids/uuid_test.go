package ids

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

type failedReader struct{}

func (failedReader) Read([]byte) (int, error) { return 0, errors.New("entropy unavailable") }

func TestNewV7(t *testing.T) {
	id, err := NewV7()
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := uuid.Parse(id)
	if err != nil || parsed.Version() != 7 {
		t.Fatalf("expected UUIDv7, got %q (%v)", id, err)
	}
}

func TestNewV7PropagatesEntropyFailure(t *testing.T) {
	uuid.SetRand(failedReader{})
	t.Cleanup(func() { uuid.SetRand(nil) })
	id, err := NewV7()
	if err == nil || id != "" {
		t.Fatalf("expected error and empty ID, got %q, %v", id, err)
	}
}
