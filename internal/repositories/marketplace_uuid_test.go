package repositories

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"my-go-app/internal/models"
)

func TestCloneIDsPrepareMappingsAndFailure(t *testing.T) {
	pages := []models.Page{{Id: "old-page"}}
	elements := []models.Element{{Id: "old-parent"}, {Id: "old-child"}}
	project, pageIDs, elementIDs, err := prepareCloneIDs(pages, elements, func() (string, error) { id, err := uuid.NewV7(); return id.String(), err })
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{project, pageIDs["old-page"], elementIDs["old-parent"], elementIDs["old-child"]} {
		parsed, err := uuid.Parse(id)
		if err != nil || parsed.Version() != 7 {
			t.Fatalf("bad new ID %q: %v", id, err)
		}
	}
	calls := 0
	_, _, _, err = prepareCloneIDs(pages, elements, func() (string, error) {
		calls++
		if calls == 3 {
			return "", errors.New("entropy failure")
		}
		return "generated", nil
	})
	if err == nil || calls != 3 {
		t.Fatalf("expected generation failure after two IDs, got %v after %d calls", err, calls)
	}
}
