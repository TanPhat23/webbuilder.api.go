package unit

import (
	"context"
	"reflect"
	"testing"

	"my-go-app/internal/realtime/domain/entity"
	domainErrors "my-go-app/internal/realtime/domain/errors"
	domainRepo "my-go-app/internal/realtime/domain/repository"
	domainUsecase "my-go-app/internal/realtime/domain/usecase"
	"my-go-app/internal/realtime/usecase"
	elementutils "my-go-app/internal/realtime/utils/element"
)

// fakeDocUseCase is a small in-memory implementation of DocumentUseCase used for tests.
type fakeDocUseCase struct {
	doc   []entity.Element
	added []entity.Element
	seq   int64
	// persist mimics the repository marking the page dirty inside the same atomic update.
	persist *fakePersistUseCase
}

func (f *fakeDocUseCase) GetDocument(ctx context.Context, projectID string, pageID string) ([]entity.Element, error) {
	out := make([]entity.Element, len(f.doc))
	copy(out, f.doc)
	return out, nil
}

func (f *fakeDocUseCase) RefreshCache(ctx context.Context, projectID string, pageID string) error {
	return nil
}

func (f *fakeDocUseCase) GetDocumentWithSeq(ctx context.Context, projectID string, pageID string) ([]entity.Element, int64, error) {
	return f.doc, f.seq, nil
}

func (f *fakeDocUseCase) MutateDocument(ctx context.Context, projectID string, pageID string, mutate domainRepo.DocumentMutateFunc) (int64, error) {
	out := make([]entity.Element, len(f.doc))
	copy(out, f.doc)
	updated, err := mutate(out)
	if err != nil {
		return 0, err
	}
	f.added = make([]entity.Element, len(updated))
	copy(f.added, updated)
	f.seq++
	if f.persist != nil {
		if err := f.persist.MarkDirty(ctx, projectID, pageID); err != nil {
			return 0, err
		}
	}
	return f.seq, nil
}

func (f *fakeDocUseCase) MutateElement(ctx context.Context, projectID string, pageID string, elementID string, mutate domainRepo.ElementMutateFunc) (int64, error) {
	return f.MutateDocument(ctx, projectID, pageID, func(elements []entity.Element) ([]entity.Element, error) {
		target := elementutils.FindByID(elements, elementID)
		if target == nil {
			return nil, domainErrors.New(domainErrors.CodeNotFound, "element not found in cache", nil)
		}
		return elements, mutate(target)
	})
}

func (f *fakeDocUseCase) AddElementsToDocument(ctx context.Context, projectID string, pageID string, elements []entity.Element) error {
	f.added = make([]entity.Element, len(elements))
	copy(f.added, elements)
	return nil
}

// fakePersistUseCase is a minimal PersistUseCase that records MarkDirty calls.
type fakePersistUseCase struct {
	calls []struct {
		projectID string
		pageID    string
	}
}

func (f *fakePersistUseCase) Start(ctx context.Context) {}

func (f *fakePersistUseCase) MarkDirty(ctx context.Context, projectID, pageID string) error {
	f.calls = append(f.calls, struct {
		projectID string
		pageID    string
	}{projectID: projectID, pageID: pageID})
	return nil
}

func (f *fakePersistUseCase) Flush(ctx context.Context) error { return nil }

func newUnitElementUseCase(initial []entity.Element) (*fakeDocUseCase, *fakePersistUseCase, domainUsecase.ElementUseCase) {
	doc := &fakeDocUseCase{doc: initial}
	prst := &fakePersistUseCase{}
	doc.persist = prst
	return doc, prst, usecase.NewElementUseCase(&domainRepo.Repository{}, doc, prst)
}

func assertMarkedDirtyOnce(t *testing.T, prst *fakePersistUseCase, projID, pageID string) {
	t.Helper()
	if len(prst.calls) != 1 {
		t.Fatalf("expected MarkDirty to be called once, got %d", len(prst.calls))
	}
	if prst.calls[0].projectID != projID || prst.calls[0].pageID != pageID {
		t.Fatalf("unexpected MarkDirty args: %#v", prst.calls[0])
	}
}

func assertOrderedChildren(t *testing.T, added []entity.Element, parentID string, wantIDs []string) {
	t.Helper()
	parent := findElementByID(added, parentID)
	if parent == nil {
		t.Fatalf("expected to find %s in updated tree", parentID)
	}
	gotIDs := idsFromElements(parent.Elements)
	if !reflect.DeepEqual(gotIDs, wantIDs) {
		t.Fatalf("unexpected children for %s: got %v want %v", parentID, gotIDs, wantIDs)
	}
	for i := range parent.Elements {
		if parent.Elements[i].Order != i {
			t.Fatalf("unexpected Order on %s.Elements[%d] (id=%s): got %d want %d", parentID, i, parent.Elements[i].ID, parent.Elements[i].Order, i)
		}
	}
}

func hasElementWithTypeAndPage(els []entity.Element, typ, pageID string) bool {
	for _, el := range els {
		if el.Type == typ && el.PageID == pageID {
			return true
		}
	}
	return false
}

func TestMoveElement_ReordersAcrossParents(t *testing.T) {
	initial := []entity.Element{
		{ID: "p1", Order: 0, Elements: []entity.Element{{ID: "c1", ParentID: strPtr("p1"), Order: 0}, {ID: "c2", ParentID: strPtr("p1"), Order: 1}}},
		{ID: "p2", Order: 1, Elements: []entity.Element{{ID: "c3", ParentID: strPtr("p2"), Order: 0}, {ID: "c4", ParentID: strPtr("p2"), Order: 1}}},
	}
	doc, prst, uc := newUnitElementUseCase(initial)

	moved, _, err := uc.MoveElement(context.Background(), "proj1", "page1", "c1", "p2", 1)
	if err != nil || moved == nil {
		t.Fatalf("MoveElement moved=%+v err=%v", moved, err)
	}
	if moved.ID != "c1" || moved.Order != 1 {
		t.Fatalf("expected moved ID 'c1' order 1, got %+v", moved)
	}
	if moved.ParentID == nil || *moved.ParentID != "p2" {
		t.Fatalf("expected parent 'p2', got %v", moved.ParentID)
	}
	if doc.added == nil {
		t.Fatalf("expected AddElementsToDocument to be called")
	}
	assertOrderedChildren(t, doc.added, "p2", []string{"c3", "c1", "c4"})
	assertOrderedChildren(t, doc.added, "p1", []string{"c2"})
	assertMarkedDirtyOnce(t, prst, "proj1", "page1")
}

func TestUpdateElement_UseCase(t *testing.T) {
	initial := []entity.Element{
		{ID: "root", Elements: []entity.Element{{ID: "child", ParentID: strPtr("root"), Order: 0, Type: "oldType"}}},
	}
	doc, prst, uc := newUnitElementUseCase(initial)

	if _, err := uc.UpdateElement(context.Background(), "proj1", "page1", &entity.Element{ID: "child", Type: "newType"}); err != nil {
		t.Fatalf("unexpected error from UpdateElement: %v", err)
	}
	if doc.added == nil {
		t.Fatalf("expected AddElementsToDocument to be called")
	}
	child := findElementByID(doc.added, "child")
	if child == nil || child.Type != "newType" {
		t.Fatalf("expected child with Type 'newType', got %+v", child)
	}
	assertMarkedDirtyOnce(t, prst, "proj1", "page1")
}

func TestRemoveElement_UseCase(t *testing.T) {
	initial := []entity.Element{
		{ID: "root", Elements: []entity.Element{{ID: "child", ParentID: strPtr("root"), Elements: []entity.Element{{ID: "grand", ParentID: strPtr("child")}}}}},
	}
	doc, prst, uc := newUnitElementUseCase(initial)

	if _, err := uc.RemoveElement(context.Background(), "proj2", "page2", "grand"); err != nil {
		t.Fatalf("unexpected error from RemoveElement: %v", err)
	}
	if doc.added == nil {
		t.Fatalf("expected AddElementsToDocument to be called")
	}
	child := findElementByID(doc.added, "child")
	if child == nil || len(child.Elements) != 0 {
		t.Fatalf("expected grandchild removed from child: %+v", child)
	}
	assertMarkedDirtyOnce(t, prst, "proj2", "page2")
}

func TestCreateElement_UseCase(t *testing.T) {
	parentID := "8a2f5f8c-3b45-4b1e-a2eb-1234567890ab"
	initial := []entity.Element{
		{ID: parentID, Elements: []entity.Element{{ID: "child", ParentID: strPtr(parentID), Order: 0}}},
	}
	doc, prst, uc := newUnitElementUseCase(initial)

	newEl := entity.Element{ID: "", Type: "container", PageID: "page3", Order: 1}
	if _, err := uc.CreateElement(context.Background(), "proj3", &newEl); err != nil {
		t.Fatalf("unexpected error from CreateElement: %v", err)
	}
	if doc.added == nil || !hasElementWithTypeAndPage(doc.added, "container", "page3") {
		t.Fatalf("expected new container on page3 in cached tree: %v", doc.added)
	}
	assertMarkedDirtyOnce(t, prst, "proj3", "page3")
}

func TestCreateElement_EmptyParentStringTreatedAsRoot(t *testing.T) {
	parentID := "11111111-1111-1111-1111-111111111111"
	doc, prst, uc := newUnitElementUseCase([]entity.Element{{ID: parentID, Elements: []entity.Element{}}})

	// parentId is an empty string - should be accepted and treated as no parent,
	// which results in the element being appended at root-level (top-level element).
	newEl := entity.Element{Type: "button", PageID: "pageTest", ParentID: strPtr("")}
	if _, err := uc.CreateElement(context.Background(), "projTest", &newEl); err != nil {
		t.Fatalf("unexpected error from CreateElement with empty parentId: %v", err)
	}
	if doc.added == nil || !hasElementWithTypeAndPage(doc.added, "button", "pageTest") {
		t.Fatalf("expected new top-level element to be present; got %v", doc.added)
	}
	if len(prst.calls) != 1 {
		t.Fatalf("expected MarkDirty to be called once, got %d", len(prst.calls))
	}
}

func TestCreateElement_InsertUnderParent(t *testing.T) {
	initial := []entity.Element{
		{
			ID: "11111111-1111-1111-1111-111111111111",
			Elements: []entity.Element{
				{ID: "child", ParentID: strPtr("11111111-1111-1111-1111-111111111111"), Order: 0},
			},
		},
	}
	doc, prst, uc := newUnitElementUseCase(initial)

	el := entity.Element{ID: "", Type: "button", PageID: "page3", ParentID: strPtr("root")}
	if _, err := uc.CreateElement(context.Background(), "proj3", &el); err != nil {
		t.Fatalf("unexpected error from CreateElement: %v", err)
	}
	if doc.added == nil {
		t.Fatalf("expected AddElementsToDocument to be called")
	}
	parent := findElementByID(doc.added, "root")
	if parent == nil || !hasElementWithTypeAndPage(parent.Elements, "button", "page3") {
		t.Fatalf("expected new child under root; got parent=%v", parent)
	}
	assertMarkedDirtyOnce(t, prst, "proj3", "page3")
}
