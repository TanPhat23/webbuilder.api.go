package usecase

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"slices"
	"sync"
	"testing"

	domainEntity "my-go-app/internal/realtime/domain/entity"
	domainErrors "my-go-app/internal/realtime/domain/errors"
	"my-go-app/internal/realtime/domain/logic"
	domainRepo "my-go-app/internal/realtime/domain/repository"
	"my-go-app/internal/realtime/utils/logger"
)

type fakeElementRepo struct {
	byPage          []domainEntity.Element
	byProj          []domainEntity.Element
	updateErr       error
	updateCalls     int
	getByPageErr    error
	getByProjectErr error
}

func (f *fakeElementRepo) CreateElement(context.Context, string, string, *domainEntity.Element) error {
	return nil
}
func (f *fakeElementRepo) CreateElements(context.Context, string, string, []domainEntity.Element) error {
	return nil
}
func (f *fakeElementRepo) UpdateElement(context.Context, string, string, *domainEntity.Element) error {
	return nil
}
func (f *fakeElementRepo) UpdateElements(context.Context, string, string, []domainEntity.Element) error {
	f.updateCalls++
	return f.updateErr
}
func (f *fakeElementRepo) UpsertElements(context.Context, string, string, []domainEntity.Element) error {
	return nil
}
func (f *fakeElementRepo) DeleteElement(context.Context, string, string, string) error {
	return nil
}
func (f *fakeElementRepo) GetElementByID(context.Context, string, string, string) (*domainEntity.Element, error) {
	return nil, nil
}
func (f *fakeElementRepo) GetElementsByPageID(context.Context, string, string) ([]domainEntity.Element, error) {
	if f.getByPageErr != nil {
		return nil, f.getByPageErr
	}
	return slices.Clone(f.byPage), nil
}
func (f *fakeElementRepo) GetElementsByProjectID(context.Context, string) ([]domainEntity.Element, error) {
	if f.getByProjectErr != nil {
		return nil, f.getByProjectErr
	}
	return slices.Clone(f.byProj), nil
}

type fakeDocumentRepo struct {
	mu           sync.Mutex
	doc          []domainEntity.Element
	seq          int64
	dirty        []domainEntity.DirtyPage
	getErr       error
	setErr       error
	setAbsentErr error
	seqErr       error
	markErr      error
	popErr       error
	markedDirty  []string
}

func (f *fakeDocumentRepo) SetDocument(_ context.Context, _, _ string, elements []domainEntity.Element) error {
	if f.setErr != nil {
		return f.setErr
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.doc = slices.Clone(elements)
	return nil
}
func (f *fakeDocumentRepo) GetDocument(context.Context, string, string) ([]domainEntity.Element, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.doc == nil {
		return nil, nil
	}
	return slices.Clone(f.doc), nil
}
func (f *fakeDocumentRepo) DeleteDocument(context.Context, string, string) error { return nil }
func (f *fakeDocumentRepo) SetDocumentIfAbsent(_ context.Context, _, _ string, elements []domainEntity.Element) error {
	if f.setAbsentErr != nil {
		return f.setAbsentErr
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.doc == nil {
		f.doc = slices.Clone(elements)
	}
	return nil
}
func (f *fakeDocumentRepo) GetDocumentWithSeq(context.Context, string, string) ([]domainEntity.Element, int64, error) {
	if f.seqErr != nil {
		return nil, 0, f.seqErr
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.doc), f.seq, nil
}
func (f *fakeDocumentRepo) UpdateDocument(ctx context.Context, _, _ string, load domainRepo.DocumentLoadFunc, mutate domainRepo.DocumentMutateFunc) (int64, error) {
	f.mu.Lock()
	cur := slices.Clone(f.doc)
	f.mu.Unlock()
	if cur == nil && load != nil {
		loaded, err := load(ctx)
		if err != nil {
			return 0, err
		}
		cur = loaded
	}
	next, err := mutate(cur)
	if err != nil {
		return 0, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.doc = next
	f.seq++
	return f.seq, nil
}
func (f *fakeDocumentRepo) PatchDocumentElement(ctx context.Context, _, _, elementID string, load domainRepo.DocumentLoadFunc, mutate domainRepo.ElementMutateFunc) (int64, error) {
	return f.UpdateDocument(ctx, "", "", load, func(elements []domainEntity.Element) ([]domainEntity.Element, error) {
		for i := range elements {
			if elements[i].ID == elementID {
				if err := mutate(&elements[i]); err != nil {
					return nil, err
				}
				return elements, nil
			}
		}
		return nil, domainErrors.New(domainErrors.CodeNotFound, "not found", nil)
	})
}
func (f *fakeDocumentRepo) MarkPageDirty(_ context.Context, projectID, pageID string) error {
	if f.markErr != nil {
		return f.markErr
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.markedDirty = append(f.markedDirty, projectID+":"+pageID)
	return nil
}
func (f *fakeDocumentRepo) PopDirtyPages(context.Context, int) ([]domainEntity.DirtyPage, error) {
	if f.popErr != nil {
		return nil, f.popErr
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	out := f.dirty
	f.dirty = nil
	return out, nil
}

type fakeSnapshotRepo struct {
	saveErr error
	saved   int
}

func (f *fakeSnapshotRepo) SaveSnapshot(context.Context, string, string, string, domainEntity.SnapshotType, []domainEntity.Element) error {
	f.saved++
	return f.saveErr
}
func (f *fakeSnapshotRepo) GetSnapshot(context.Context, string, string, domainEntity.SnapshotType) (*domainEntity.Snapshot, error) {
	return nil, nil
}
func (f *fakeSnapshotRepo) DeleteSnapshot(context.Context, string, string, domainEntity.SnapshotType) error {
	return nil
}

type fakeWorkflowRepo struct {
	items map[string]*domainEntity.EventWorkflow
}

func (f *fakeWorkflowRepo) CreateWorkflow(_ context.Context, _ string, wf *domainEntity.EventWorkflow) error {
	if f.items == nil {
		f.items = map[string]*domainEntity.EventWorkflow{}
	}
	f.items[wf.ID] = wf
	return nil
}
func (f *fakeWorkflowRepo) UpdateWorkflow(_ context.Context, _ string, wf *domainEntity.EventWorkflow) error {
	f.items[wf.ID] = wf
	return nil
}
func (f *fakeWorkflowRepo) DeleteWorkflow(_ context.Context, _ string, id string) error {
	delete(f.items, id)
	return nil
}
func (f *fakeWorkflowRepo) GetWorkflowByID(_ context.Context, _ string, id string) (*domainEntity.EventWorkflow, error) {
	return f.items[id], nil
}
func (f *fakeWorkflowRepo) GetWorkflowsByProjectID(context.Context, string) ([]domainEntity.EventWorkflow, error) {
	out := make([]domainEntity.EventWorkflow, 0, len(f.items))
	for _, wf := range f.items {
		out = append(out, *wf)
	}
	return out, nil
}

type fakeElementWorkflowRepo struct {
	byElem map[string]*domainEntity.ElementEventWorkflow
}

func (f *fakeElementWorkflowRepo) CreateElementWorkflow(_ context.Context, _ string, wf *domainEntity.ElementEventWorkflow) error {
	if f.byElem == nil {
		f.byElem = map[string]*domainEntity.ElementEventWorkflow{}
	}
	f.byElem[wf.ElementID] = wf
	return nil
}
func (f *fakeElementWorkflowRepo) UpdateElementWorkflow(_ context.Context, _ string, wf *domainEntity.ElementEventWorkflow) error {
	f.byElem[wf.ElementID] = wf
	return nil
}
func (f *fakeElementWorkflowRepo) DeleteElementWorkflow(_ context.Context, _ string, elementID string) error {
	delete(f.byElem, elementID)
	return nil
}
func (f *fakeElementWorkflowRepo) GetElementWorkflowByElementID(_ context.Context, _ string, elementID string) (*domainEntity.ElementEventWorkflow, error) {
	return f.byElem[elementID], nil
}
func (f *fakeElementWorkflowRepo) GetElementWorkflowsByProjectID(context.Context, string) ([]domainEntity.ElementEventWorkflow, error) {
	out := make([]domainEntity.ElementEventWorkflow, 0, len(f.byElem))
	for _, wf := range f.byElem {
		out = append(out, *wf)
	}
	return out, nil
}
func (f *fakeElementWorkflowRepo) GetElementWorkflowsByPageID(context.Context, string, string) ([]domainEntity.ElementEventWorkflow, error) {
	return f.GetElementWorkflowsByProjectID(context.Background(), "")
}

type fakeLogicRepo struct {
	err       error
	readOnly  bool
	canWrite  bool
	accessErr error
}

func (f *fakeLogicRepo) Move(context.Context, string, string, string, string, string, string) (logic.Revisioned, logic.Revisioned, error) {
	if f.err != nil {
		return logic.Revisioned{}, logic.Revisioned{}, f.err
	}
	return logic.Revisioned{Revision: 2}, logic.Revisioned{Revision: 3}, nil
}
func (f *fakeLogicRepo) Apply(context.Context, string, string, string, map[string]any, ...[][]string) (logic.Revisioned, error) {
	if f.err != nil {
		return logic.Revisioned{}, f.err
	}
	return logic.Revisioned{Doc: logic.Empty(), Revision: 5}, nil
}
func (f *fakeLogicRepo) Get(context.Context, string, string, string) (logic.Revisioned, logic.Revisioned, error) {
	if f.err != nil {
		return logic.Revisioned{}, logic.Revisioned{}, f.err
	}
	return logic.Revisioned{Revision: 1}, logic.Revisioned{Revision: 2}, nil
}
func (f *fakeLogicRepo) Access(context.Context, string, string) (bool, bool, error) {
	return f.readOnly, f.canWrite, f.accessErr
}

func makeBundle() (*domainRepo.Repository, *fakeElementRepo, *fakeDocumentRepo, *fakeSnapshotRepo, *fakeLogicRepo) {
	elem := &fakeElementRepo{}
	doc := &fakeDocumentRepo{}
	snap := &fakeSnapshotRepo{}
	lgc := &fakeLogicRepo{}
	repo := &domainRepo.Repository{
		LogicRepo:           lgc,
		ElementRepo:         elem,
		SnapshotRepo:        snap,
		DocumentRepo:        doc,
		WorkflowRepo:        &fakeWorkflowRepo{},
		ElementWorkflowRepo: &fakeElementWorkflowRepo{},
	}
	return repo, elem, doc, snap, lgc
}

func TestDocumentAndElementUseCaseLifecycle(t *testing.T) {
	ctx := context.Background()
	repo, elemRepo, docRepo, _, _ := makeBundle()
	elemID1 := "11111111-1111-1111-1111-111111111111"
	elemID2 := "22222222-2222-2222-2222-222222222222"
	frameID := "33333333-3333-3333-3333-333333333333"
	elemRepo.byPage = []domainEntity.Element{{ID: elemID1, Type: "Text", PageID: "page-1"}}
	elemRepo.byProj = []domainEntity.Element{{ID: elemID1, Type: "Text", PageID: "page-1"}}

	uc := NewUseCase(repo, logger.New(io.Discard, "", 0))

	got, seq, err := uc.Element.SyncPageWithSeq(ctx, "proj-1", "page-1")
	if err != nil || len(got) != 1 || seq != 0 {
		t.Fatalf("SyncPageWithSeq=%v seq=%d err=%v", got, seq, err)
	}
	synced, err := uc.Element.SyncPage(ctx, "proj-1", "page-1")
	if err != nil || len(synced) != 1 {
		t.Fatalf("SyncPage err=%v len=%d", err, len(synced))
	}
	byProj, err := uc.Element.GetElementsByProjectID(ctx, "proj-1")
	if err != nil || len(byProj) != 1 {
		t.Fatalf("GetElementsByProjectID err=%v", err)
	}
	if byPage, err := uc.Element.GetElementsByPageID(ctx, "proj-1", "page-1"); err != nil || len(byPage) != 1 {
		t.Fatalf("GetElementsByPageID err=%v", err)
	}

	frame := domainEntity.Element{ID: frameID, Type: "Frame", PageID: "page-1"}
	if _, err := uc.Element.CreateElement(ctx, "proj-1", &frame); err != nil {
		t.Fatalf("CreateElement frame: %v", err)
	}
	child := domainEntity.Element{ID: elemID2, Type: "Text", PageID: "page-1"}
	if _, err := uc.Element.CreateElementAt(ctx, "proj-1", &child, nil); err != nil {
		t.Fatalf("CreateElementAt child: %v", err)
	}
	bad := domainEntity.Element{Type: ""}
	if _, err := uc.Element.CreateElement(ctx, "proj-1", &bad); !domainErrors.IsCode(err, domainErrors.CodeValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}

	newContent := "Updated copy"
	if _, err := uc.Element.UpdateElement(ctx, "proj-1", "page-1", &domainEntity.Element{ID: elemID1, Type: "Text", PageID: "page-1", Content: &newContent}); err != nil {
		t.Fatalf("UpdateElement: %v", err)
	}
	if _, err := uc.Element.PatchElement(ctx, "proj-1", "page-1", elemID1, map[string]any{"name": "Hero"}, map[string]any{"color": "red"}, map[string]any{"flag": true}); err != nil {
		t.Fatalf("PatchElement: %v", err)
	}

	moved, _, err := uc.Element.MoveElement(ctx, "proj-1", "page-1", elemID2, frameID, 0)
	if err != nil || moved.ParentID == nil || *moved.ParentID != frameID {
		t.Fatalf("MoveElement moved=%+v err=%v", moved, err)
	}
	if _, _, err := uc.Element.MoveElement(ctx, "proj-1", "page-1", frameID, elemID2, 0); !domainErrors.IsCode(err, domainErrors.CodeValidation) {
		t.Fatalf("expected cycle error, got %v", err)
	}
	if _, _, err := uc.Element.MoveElement(ctx, "proj-1", "page-1", elemID2, "", -1); !domainErrors.IsCode(err, domainErrors.CodeValidation) {
		t.Fatalf("expected negative index validation error, got %v", err)
	}

	if _, err := uc.Element.RemoveElement(ctx, "proj-1", "page-1", elemID2); err != nil {
		t.Fatalf("RemoveElement nested: %v", err)
	}
	if _, err := uc.Element.RemoveElement(ctx, "proj-1", "page-1", "nonexistent"); !domainErrors.IsCode(err, domainErrors.CodeNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}

	if err := uc.Document.AddElementsToDocument(ctx, "proj-1", "page-1", []domainEntity.Element{{ID: "root-2", PageID: "page-1"}}); err != nil {
		t.Fatalf("AddElementsToDocument: %v", err)
	}
	if gotDoc, _ := docRepo.GetDocument(ctx, "proj-1", "page-1"); len(gotDoc) != 1 || gotDoc[0].ID != "root-2" {
		t.Fatalf("expected doc replaced, got %+v", gotDoc)
	}
}

func TestPersistWorkflowAndLogicUseCases(t *testing.T) {
	ctx := context.Background()
	repo, elemRepo, docRepo, snapRepo, lgcRepo := makeBundle()
	docRepo.doc = []domainEntity.Element{{ID: "11111111-1111-1111-1111-111111111111", Type: "Text", PageID: "page-1"}}
	uc := NewUseCase(repo, logger.New(io.Discard, "", 0))

	if err := uc.Persist.MarkDirty(ctx, "proj-1", "page-1"); err != nil {
		t.Fatalf("MarkDirty: %v", err)
	}
	if len(docRepo.markedDirty) != 1 || docRepo.markedDirty[0] != "proj-1:page-1" {
		t.Fatalf("markedDirty=%v", docRepo.markedDirty)
	}

	docRepo.dirty = []domainEntity.DirtyPage{{ProjectID: "proj-1", PageID: "page-1"}}
	if err := uc.Persist.Flush(ctx); err != nil {
		t.Fatalf("Flush success: %v", err)
	}
	if elemRepo.updateCalls != 1 || snapRepo.saved != 1 {
		t.Fatalf("updateCalls=%d saved=%d", elemRepo.updateCalls, snapRepo.saved)
	}

	// Requeue on downstream failure
	docRepo.dirty = []domainEntity.DirtyPage{{ProjectID: "proj-1", PageID: "page-2"}}
	snapRepo.saveErr = errors.New("db down")
	if err := uc.Persist.Flush(ctx); err == nil {
		t.Fatal("expected flush error when snapshot save fails")
	}

	// Workflow & ElementWorkflow usecases
	wf := &domainEntity.EventWorkflow{ID: "wf-1", Name: "Click"}
	if err := uc.Workflow.CreateWorkflow(ctx, "proj-1", wf); err != nil {
		t.Fatalf("CreateWorkflow: %v", err)
	}
	wf.Name = "Updated Click"
	if err := uc.Workflow.UpdateWorkflow(ctx, "proj-1", wf); err != nil {
		t.Fatalf("UpdateWorkflow: %v", err)
	}
	if got, _ := uc.Workflow.GetWorkflowByID(ctx, "proj-1", "wf-1"); got == nil || got.Name != "Updated Click" {
		t.Fatalf("GetWorkflowByID=%+v", got)
	}
	if list, _ := uc.Workflow.GetWorkflowsByProject(ctx, "proj-1"); len(list) != 1 {
		t.Fatalf("GetWorkflowsByProject len=%d", len(list))
	}
	if err := uc.Workflow.RemoveWorkflow(ctx, "proj-1", "wf-1"); err != nil {
		t.Fatalf("RemoveWorkflow: %v", err)
	}

	ewf := &domainEntity.ElementEventWorkflow{ID: "ewf-1", ElementID: "el-1", WorkflowID: "wf-1", EventName: "click"}
	if err := uc.ElementWorkflows.CreateElementWorkflow(ctx, "proj-1", ewf); err != nil {
		t.Fatalf("CreateElementWorkflow: %v", err)
	}
	if err := uc.ElementWorkflows.UpdateElementWorkflow(ctx, "proj-1", ewf); err != nil {
		t.Fatalf("UpdateElementWorkflow: %v", err)
	}
	if got, _ := uc.ElementWorkflows.GetElementWorkflowByElementID(ctx, "proj-1", "el-1"); got == nil {
		t.Fatal("GetElementWorkflowByElementID returned nil")
	}
	if list, _ := uc.ElementWorkflows.GetWorkflowsByPage(ctx, "proj-1", "page-1"); len(list) != 1 {
		t.Fatalf("GetWorkflowsByPage len=%d", len(list))
	}
	if err := uc.ElementWorkflows.RemoveElementWorkflow(ctx, "proj-1", "el-1"); err != nil {
		t.Fatalf("RemoveElementWorkflow: %v", err)
	}

	// Logic usecase error mapping & success paths
	lgcRepo.readOnly, lgcRepo.canWrite = true, false
	if r, w, err := uc.Logic.Access(ctx, "proj-1", "u1"); err != nil || !r || w {
		t.Fatalf("Access r=%v w=%v err=%v", r, w, err)
	}
	if _, err := uc.Logic.Update(ctx, "proj-1", "page-1", "u1", map[string]any{}); err != nil {
		t.Fatalf("Logic.Update: %v", err)
	}
	if _, _, err := uc.Logic.Move(ctx, "proj-1", "page-1", "u1", "variable", "x", "project"); err != nil {
		t.Fatalf("Logic.Move: %v", err)
	}
	if _, _, err := uc.Logic.Get(ctx, "proj-1", "page-1", "u1"); err != nil {
		t.Fatalf("Logic.Get: %v", err)
	}

	for _, tc := range []struct {
		repoErr  error
		wantCode string
	}{
		{domainErrors.ErrUnauthorized, domainErrors.CodeAccessDenied},
		{sql.ErrNoRows, domainErrors.CodeNotFound},
		{domainErrors.ErrValidation, domainErrors.CodeValidation},
		{errors.New("boom"), domainErrors.CodeInternal},
	} {
		lgcRepo.err = tc.repoErr
		if _, err := uc.Logic.Update(ctx, "proj-1", "", "u1", nil); !domainErrors.IsCode(err, tc.wantCode) {
			t.Fatalf("repoErr=%v got=%v want=%s", tc.repoErr, err, tc.wantCode)
		}
	}
}
