package unit

import (
	"context"
	"errors"
	"testing"

	"my-go-app/internal/realtime/domain/entity"
	repo "my-go-app/internal/realtime/domain/repository"
	"my-go-app/internal/realtime/usecase"
)

type poppedDocument struct {
	*memDocRepo
	popped bool
}

func (d *poppedDocument) PopDirtyPages(_ context.Context, _ int) ([]entity.DirtyPage, error) {
	if d.popped {
		return nil, nil
	}
	d.popped = true
	return []entity.DirtyPage{{ProjectID: "proj", PageID: "page"}}, nil
}

type failingElements struct{ repo.ElementRepository }

func (failingElements) UpdateElements(context.Context, string, string, []entity.Element) error {
	return errors.New("database unavailable")
}
func TestFailedPersistenceIsMarkedDirtyForRetry(t *testing.T) {
	docs := &poppedDocument{memDocRepo: newMemDocRepo()}
	if err := docs.SetDocument(context.Background(), "proj", "page", []entity.Element{}); err != nil {
		t.Fatal(err)
	}
	persist := usecase.NewPersistenceUsecase(&repo.Repository{DocumentRepo: docs, ElementRepo: failingElements{}}, &mockLogger{})
	if err := persist.Flush(context.Background()); err == nil {
		t.Fatal("expected persistence failure")
	}
	if docs.dirty[memKey("proj", "page")] == 0 {
		t.Fatal("failed page was lost from the dirty queue")
	}
}

type countingElements struct {
	repo.ElementRepository
	calls int
}

func (e *countingElements) UpdateElements(context.Context, string, string, []entity.Element) error {
	e.calls++
	return errors.New("unexpected database write")
}
func TestMissingCachedDocumentDoesNotDeletePersistedPage(t *testing.T) {
	docs := &poppedDocument{memDocRepo: newMemDocRepo()}
	elements := &countingElements{}
	persist := usecase.NewPersistenceUsecase(&repo.Repository{DocumentRepo: docs, ElementRepo: elements}, &mockLogger{})
	if err := persist.Flush(context.Background()); err == nil {
		t.Fatal("missing document must fail")
	}
	if elements.calls != 0 {
		t.Fatal("missing cache was treated as an empty page")
	}
	if docs.dirty[memKey("proj", "page")] == 0 {
		t.Fatal("missing page was not requeued")
	}
}
