package unit

import (
	"context"
	"encoding/json"
	"sync"
	"testing"

	"my-go-app/internal/realtime/domain/entity"
	domainRepo "my-go-app/internal/realtime/domain/repository"
	"my-go-app/internal/realtime/usecase"
	elementutils "my-go-app/internal/realtime/utils/element"
)

func TestMoveAssignsFractionalKeyAndKeepsLegacyOrder(t *testing.T) {
	docs := newMemDocRepo()
	initial := []entity.Element{{ID: "a", Order: 0}, {ID: "b", Order: 1}, {ID: "c", Order: 2}}
	if err := docs.SetDocument(context.Background(), testProject, testPage, initial); err != nil {
		t.Fatal(err)
	}
	repo := &domainRepo.Repository{DocumentRepo: docs, ElementRepo: &memElementRepo{}}
	uc := usecase.NewElementUseCase(repo, usecase.NewDocumentUseCase(repo), nil)
	moved, _, err := uc.MoveElement(context.Background(), testProject, testPage, "c", "", 1)
	if err != nil {
		t.Fatal(err)
	}
	got := docs.snapshot(testProject, testPage)
	if ids := idsFromElements(got); len(ids) != 3 || ids[0] != "a" || ids[1] != "c" || ids[2] != "b" {
		t.Fatalf("unexpected order %v", ids)
	}
	last := ""
	for i, el := range got {
		data, _ := json.Marshal(el)
		var fields map[string]any
		_ = json.Unmarshal(data, &fields)
		key, _ := fields["orderKey"].(string)
		if key == "" || key <= last {
			t.Fatalf("keys are not strictly increasing: %s", data)
		}
		if el.Order != i {
			t.Fatalf("order = %d, want %d", el.Order, i)
		}
		last = key
	}
	data, _ := json.Marshal(moved)
	var fields map[string]any
	_ = json.Unmarshal(data, &fields)
	if fields["orderKey"] == nil {
		t.Fatal("move result must contain committed key")
	}
}

func TestReloadSortsKeyedSiblingsByKey(t *testing.T) {
	got := elementutils.BuildElementTree([]entity.Element{
		{ID: "a", Order: 0, OrderKey: "a1"},
		{ID: "b", Order: 1, OrderKey: "a0"},
	})
	if ids := idsFromElements(got); ids[0] != "b" || ids[1] != "a" {
		t.Fatalf("keyed order ignored: %v", ids)
	}
}

func TestMoveRejectsNegativeIndexAndInvalidParent(t *testing.T) {
	doc := &fakeDocUseCase{doc: []entity.Element{{ID: "a", Elements: []entity.Element{{ID: "child", ParentID: strPtr("a")}}}}}
	uc := usecase.NewElementUseCase(&domainRepo.Repository{}, doc, nil)
	for _, tc := range []struct {
		parent string
		index  int
	}{{"", -1}, {"missing", 0}, {"child", 0}, {"a", 0}} {
		if _, _, err := uc.MoveElement(context.Background(), "p", "page", "a", tc.parent, tc.index); err == nil {
			t.Errorf("move to %q index %d must fail", tc.parent, tc.index)
		}
	}
}

func TestPositionedCreateAndLegacyAppend(t *testing.T) {
	ctx := context.Background()
	docs := newMemDocRepo()
	initial := []entity.Element{{ID: "a", Order: 0}, {ID: "b", Order: 1}}
	if err := docs.SetDocument(ctx, testProject, testPage, initial); err != nil {
		t.Fatal(err)
	}
	repo := &domainRepo.Repository{DocumentRepo: docs, ElementRepo: &memElementRepo{}}
	uc := usecase.NewElementUseCase(repo, usecase.NewDocumentUseCase(repo), nil)
	first := entity.Element{Type: "text", PageID: testPage}
	position := 1
	if _, err := uc.CreateElementAt(ctx, testProject, &first, &position); err != nil {
		t.Fatal(err)
	}
	if first.Order != 1 || first.OrderKey == "" {
		t.Fatalf("create result missing authoritative order: %+v", first)
	}
	last := entity.Element{Type: "text", PageID: testPage, Order: 0}
	if _, err := uc.CreateElement(ctx, testProject, &last); err != nil {
		t.Fatal(err)
	}
	got := docs.snapshot(testProject, testPage)
	ids := idsFromElements(got)
	if len(ids) != 4 || ids[0] != "a" || ids[1] != first.ID || ids[2] != "b" || ids[3] != last.ID {
		t.Fatalf("unexpected order %v", ids)
	}
	for i, el := range got {
		if el.Order != i || el.OrderKey == "" || (i > 0 && got[i-1].OrderKey >= el.OrderKey) {
			t.Fatalf("bad sibling at %d: %+v", i, el)
		}
	}
}

func TestSnapshotNormalizesLegacyTreeWithoutChangingSequence(t *testing.T) {
	ctx := context.Background()
	docs := newMemDocRepo()
	if err := docs.SetDocument(ctx, testProject, testPage, []entity.Element{{ID: "b", Order: 0}, {ID: "a", Order: 1}}); err != nil {
		t.Fatal(err)
	}
	repo := &domainRepo.Repository{DocumentRepo: docs, ElementRepo: &memElementRepo{}}
	uc := usecase.NewElementUseCase(repo, usecase.NewDocumentUseCase(repo), nil)
	got, seq, err := uc.SyncPageWithSeq(ctx, testProject, testPage)
	if err != nil {
		t.Fatal(err)
	}
	if seq != 0 || got[0].ID != "b" || got[1].ID != "a" || got[0].OrderKey == "" || got[0].OrderKey >= got[1].OrderKey {
		t.Fatalf("bad legacy snapshot: seq=%d elements=%+v", seq, got)
	}
}

func TestConcurrentCreatesAllocateDistinctKeys(t *testing.T) {
	ctx := context.Background()
	docs := newMemDocRepo()
	repo := &domainRepo.Repository{DocumentRepo: docs, ElementRepo: &memElementRepo{}}
	uc := usecase.NewElementUseCase(repo, usecase.NewDocumentUseCase(repo), nil)
	const count = 16
	var wg sync.WaitGroup
	errs := make(chan error, count)
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			el := entity.Element{Type: "text", PageID: testPage}
			position := 0
			_, err := uc.CreateElementAt(ctx, testProject, &el, &position)
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	got := docs.snapshot(testProject, testPage)
	if len(got) != count {
		t.Fatalf("got %d elements", len(got))
	}
	for i, el := range got {
		if el.Order != i || el.OrderKey == "" || (i > 0 && got[i-1].OrderKey >= el.OrderKey) {
			t.Fatalf("bad sibling at %d: %+v", i, el)
		}
	}
}
