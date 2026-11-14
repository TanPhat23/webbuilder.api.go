package unit

import (
	"context"
	"reflect"
	"testing"

	"my-go-app/internal/realtime/domain/entity"
	domainRepo "my-go-app/internal/realtime/domain/repository"
	"my-go-app/internal/realtime/usecase"
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
	// return a shallow copy to mimic typical behaviour
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
		_ = f.persist.MarkDirty(ctx, projectID, pageID)
	}
	return f.seq, nil
}

func (f *fakeDocUseCase) AddElementsToDocument(ctx context.Context, projectID string, pageID string, elements []entity.Element) error {
	// store a copy for assertions
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

func TestMoveElement_ReordersAcrossParents(t *testing.T) {
	ctx := context.Background()

	// initial hierarchical tree:
	// p1
	//  - c1 (order 0)
	//  - c2 (order 1)
	// p2
	//  - c3 (order 0)
	//  - c4 (order 1)
	initial := []entity.Element{
		{
			ID:    "p1",
			Order: 0,
			Elements: []entity.Element{
				{ID: "c1", ParentID: strPtr("p1"), Order: 0},
				{ID: "c2", ParentID: strPtr("p1"), Order: 1},
			},
		},
		{
			ID:    "p2",
			Order: 1,
			Elements: []entity.Element{
				{ID: "c3", ParentID: strPtr("p2"), Order: 0},
				{ID: "c4", ParentID: strPtr("p2"), Order: 1},
			},
		},
	}

	doc := &fakeDocUseCase{doc: initial}
	prst := &fakePersistUseCase{}
	doc.persist = prst
	uc := usecase.NewElementUseCase(&domainRepo.Repository{}, doc, prst)

	moved, _, err := uc.MoveElement(ctx, "proj1", "page1", "c1", "p2", 1)
	if err != nil {
		t.Fatalf("unexpected error from MoveElement: %v", err)
	}
	if moved == nil {
		t.Fatalf("expected moved element to be returned")
	}
	if moved.ID != "c1" {
		t.Fatalf("expected moved ID 'c1', got %q", moved.ID)
	}
	if moved.ParentID == nil || *moved.ParentID != "p2" {
		t.Fatalf("expected parent 'p2', got %v", moved.ParentID)
	}
	if moved.Order != 1 {
		t.Fatalf("expected moved Order 1, got %d", moved.Order)
	}

	// Ensure AddElementsToDocument was called with an updated tree
	if doc.added == nil {
		t.Fatalf("expected AddElementsToDocument to be called")
	}

	p2 := findElementByID(doc.added, "p2")
	if p2 == nil {
		t.Fatalf("expected to find p2 in updated tree")
	}
	gotIDs := idsFromElements(p2.Elements)
	wantIDs := []string{"c3", "c1", "c4"}
	if !reflect.DeepEqual(gotIDs, wantIDs) {
		t.Fatalf("unexpected children for p2: got %v want %v", gotIDs, wantIDs)
	}
	for i := range p2.Elements {
		if p2.Elements[i].Order != i {
			t.Fatalf("unexpected Order on p2.Elements[%d] (id=%s): got %d want %d", i, p2.Elements[i].ID, p2.Elements[i].Order, i)
		}
	}

	p1 := findElementByID(doc.added, "p1")
	if p1 == nil {
		t.Fatalf("expected to find p1 in updated tree")
	}
	gotIDs = idsFromElements(p1.Elements)
	wantIDs = []string{"c2"}
	if !reflect.DeepEqual(gotIDs, wantIDs) {
		t.Fatalf("unexpected children for p1: got %v want %v", gotIDs, wantIDs)
	}
	for i := range p1.Elements {
		if p1.Elements[i].Order != i {
			t.Fatalf("unexpected Order on p1.Elements[%d] (id=%s): got %d want %d", i, p1.Elements[i].ID, p1.Elements[i].Order, i)
		}
	}

	// Persist should have been asked to mark the page dirty
	if len(prst.calls) != 1 {
		t.Fatalf("expected MarkDirty to be called once, got %d", len(prst.calls))
	}
	if prst.calls[0].projectID != "proj1" || prst.calls[0].pageID != "page1" {
		t.Fatalf("unexpected MarkDirty args: %#v", prst.calls[0])
	}
}

func TestUpdateElement_UseCase(t *testing.T) {
	ctx := context.Background()

	initial := []entity.Element{
		{
			ID: "root",
			Elements: []entity.Element{
				{ID: "child", ParentID: strPtr("root"), Order: 0, Type: "oldType"},
			},
		},
	}

	doc := &fakeDocUseCase{doc: initial}
	prst := &fakePersistUseCase{}
	doc.persist = prst
	uc := usecase.NewElementUseCase(&domainRepo.Repository{}, doc, prst)

	_, err := uc.UpdateElement(ctx, "proj1", "page1", &entity.Element{ID: "child", Type: "newType"})
	if err != nil {
		t.Fatalf("unexpected error from UpdateElement: %v", err)
	}

	if doc.added == nil {
		t.Fatalf("expected AddElementsToDocument to be called")
	}
	child := findElementByID(doc.added, "child")
	if child == nil {
		t.Fatalf("expected child in saved tree")
	}
	if child.Type != "newType" {
		t.Fatalf("expected updated Type 'newType' got %q", child.Type)
	}
	if len(prst.calls) != 1 {
		t.Fatalf("expected MarkDirty to be called once, got %d", len(prst.calls))
	}
	if prst.calls[0].projectID != "proj1" || prst.calls[0].pageID != "page1" {
		t.Fatalf("unexpected MarkDirty args: %#v", prst.calls[0])
	}
}

func TestRemoveElement_UseCase(t *testing.T) {
	ctx := context.Background()

	initial := []entity.Element{
		{
			ID: "root",
			Elements: []entity.Element{
				{
					ID:       "child",
					ParentID: strPtr("root"),
					Elements: []entity.Element{
						{ID: "grand", ParentID: strPtr("child")},
					},
				},
			},
		},
	}

	doc := &fakeDocUseCase{doc: initial}
	prst := &fakePersistUseCase{}
	doc.persist = prst
	uc := usecase.NewElementUseCase(&domainRepo.Repository{}, doc, prst)

	if _, err := uc.RemoveElement(ctx, "proj2", "page2", "grand"); err != nil {
		t.Fatalf("unexpected error from RemoveElement: %v", err)
	}

	if doc.added == nil {
		t.Fatalf("expected AddElementsToDocument to be called")
	}
	child := findElementByID(doc.added, "child")
	if child == nil {
		t.Fatalf("expected child to remain")
	}
	if len(child.Elements) != 0 {
		t.Fatalf("expected grandchild to be removed; got %v", idsFromElements(child.Elements))
	}
	if len(prst.calls) != 1 {
		t.Fatalf("expected MarkDirty to be called once, got %d", len(prst.calls))
	}
	if prst.calls[0].projectID != "proj2" || prst.calls[0].pageID != "page2" {
		t.Fatalf("unexpected MarkDirty args: %#v", prst.calls[0])
	}
}

func TestCreateElement_UseCase(t *testing.T) {
	ctx := context.Background()

	parentID := "8a2f5f8c-3b45-4b1e-a2eb-1234567890ab"
	initial := []entity.Element{
		{
			ID: parentID,
			Elements: []entity.Element{
				{ID: "child", ParentID: strPtr(parentID), Order: 0},
			},
		},
	}

	doc := &fakeDocUseCase{doc: initial}
	prst := &fakePersistUseCase{}
	doc.persist = prst
	uc := usecase.NewElementUseCase(&domainRepo.Repository{}, doc, prst)

	newEl := entity.Element{
		ID:     "",
		Type:   "container",
		PageID: "page3",
		Order:  1,
	}

	if _, err := uc.CreateElement(ctx, "proj3", &newEl); err != nil {
		t.Fatalf("unexpected error from CreateElement: %v", err)
	}

	if doc.added == nil {
		t.Fatalf("expected AddElementsToDocument to be called")
	}

	found := false
	for _, a := range doc.added {
		if a.Type == "container" && a.PageID == "page3" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected new element to be saved in cached tree: %v", doc.added)
	}

	if len(prst.calls) != 1 {
		t.Fatalf("expected MarkDirty to be called once, got %d", len(prst.calls))
	}
	if prst.calls[0].projectID != "proj3" || prst.calls[0].pageID != "page3" {
		t.Fatalf("unexpected MarkDirty args: %#v", prst.calls[0])
	}
}

func TestCreateElement_EmptyParentStringTreatedAsRoot(t *testing.T) {
	ctx := context.Background()

	parentID := "11111111-1111-1111-1111-111111111111"
	initial := []entity.Element{
		{
			ID:       parentID,
			Elements: []entity.Element{},
		},
	}

	doc := &fakeDocUseCase{doc: initial}
	prst := &fakePersistUseCase{}
	doc.persist = prst
	uc := usecase.NewElementUseCase(&domainRepo.Repository{}, doc, prst)

	// parentId is an empty string - should be accepted and treated as no parent,
	// which results in the element being appended at root-level (top-level element).
	newEl := entity.Element{Type: "button", PageID: "pageTest", ParentID: strPtr("")}
	if _, err := uc.CreateElement(ctx, "projTest", &newEl); err != nil {
		t.Fatalf("unexpected error from CreateElement with empty parentId: %v", err)
	}

	if doc.added == nil {
		t.Fatalf("expected AddElementsToDocument to be called")
	}

	// The new element should appear at the top-level elements (no parent)
	found := false
	for _, el := range doc.added {
		if el.Type == "button" && el.PageID == "pageTest" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected new top-level element to be present; got %v", doc.added)
	}

	if len(prst.calls) != 1 {
		t.Fatalf("expected MarkDirty to be called once, got %d", len(prst.calls))
	}
}

func TestCreateElement_InsertUnderParent(t *testing.T) {
	ctx := context.Background()

	initial := []entity.Element{
		{
			ID: "11111111-1111-1111-1111-111111111111",
			Elements: []entity.Element{
				{ID: "child", ParentID: strPtr("11111111-1111-1111-1111-111111111111"), Order: 0},
			},
		},
	}

	doc := &fakeDocUseCase{doc: initial}
	prst := &fakePersistUseCase{}
	doc.persist = prst
	uc := usecase.NewElementUseCase(&domainRepo.Repository{}, doc, prst)

	// create an element under parent 'root'
	el := entity.Element{
		ID:       "",
		Type:     "button",
		PageID:   "page3",
		ParentID: strPtr("root"),
	}
	if _, err := uc.CreateElement(ctx, "proj3", &el); err != nil {
		t.Fatalf("unexpected error from CreateElement: %v", err)
	}

	if doc.added == nil {
		t.Fatalf("expected AddElementsToDocument to be called")
	}

	parent := findElementByID(doc.added, "root")
	if parent == nil {
		t.Fatalf("expected root to exist in saved tree")
	}

	found := false
	for _, c := range parent.Elements {
		if c.Type == "button" && c.PageID == "page3" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected new child to be under parent; got %v", parent.Elements)
	}

	if len(prst.calls) != 1 {
		t.Fatalf("expected MarkDirty to be called once, got %d", len(prst.calls))
	}
	if prst.calls[0].projectID != "proj3" || prst.calls[0].pageID != "page3" {
		t.Fatalf("unexpected MarkDirty args: %#v", prst.calls[0])
	}
}
