package unit

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"

	"my-go-app/internal/realtime/domain/entity"
	domainErrors "my-go-app/internal/realtime/domain/errors"
	domainRepo "my-go-app/internal/realtime/domain/repository"
	domainUsecase "my-go-app/internal/realtime/domain/usecase"
	"my-go-app/internal/realtime/usecase"
)

const (
	testProject = "proj"
	testPage    = "page"
	testElemID  = "11111111-1111-1111-1111-111111111111"
)

// newPatchFixture wires the real element/document use cases over an in-memory repository
// holding a single element.
func newPatchFixture(t *testing.T) (domainUsecase.ElementUseCase, *memDocRepo) {
	t.Helper()
	docs := newMemDocRepo()
	base := entity.Element{
		ID:     testElemID,
		Type:   "text",
		PageID: testPage,
		Styles: map[string]any{"color": "black", "nested": map[string]any{"a": 1.0, "b": 2.0}},
		Settings: map[string]any{
			"keep": "me",
		},
		Content: strPtr("old"),
	}
	repo := &domainRepo.Repository{
		DocumentRepo: docs,
		ElementRepo:  &memElementRepo{elements: []entity.Element{base}},
	}
	docUC := usecase.NewDocumentUseCase(repo)
	return usecase.NewElementUseCase(repo, docUC, nil), docs
}

func currentElement(t *testing.T, docs *memDocRepo) entity.Element {
	t.Helper()
	el := findElementByID(docs.snapshot(testProject, testPage), testElemID)
	if el == nil {
		t.Fatalf("element missing from document")
	}
	return *el
}

func TestPatchElement_ConcurrentStyleKeysAllSurvive(t *testing.T) {
	uc, docs := newPatchFixture(t)
	ctx := context.Background()

	const n = 50
	var wg sync.WaitGroup
	errs := make(chan error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := uc.PatchElement(ctx, testProject, testPage, testElemID, nil, map[string]any{fmt.Sprintf("k%d", i): i}, nil)
			errs <- err
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("PatchElement failed: %v", err)
		}
	}

	el := currentElement(t, docs)
	for i := 0; i < n; i++ {
		if _, ok := el.Styles[fmt.Sprintf("k%d", i)]; !ok {
			t.Fatalf("style key k%d was lost; styles=%v", i, el.Styles)
		}
	}
	if el.Styles["color"] != "black" {
		t.Fatalf("pre-existing style key was lost: %v", el.Styles)
	}
	if got := docs.seq(testProject, testPage); got != n {
		t.Fatalf("expected seq %d after %d ops, got %d", n, n, got)
	}
}

func TestPatchElement_PropertiesDoNotLandInSettings(t *testing.T) {
	uc, docs := newPatchFixture(t)

	_, err := uc.PatchElement(context.Background(), testProject, testPage, testElemID,
		map[string]any{"content": "new", "src": "a.png", "isLocked": true}, nil, nil)
	if err != nil {
		t.Fatalf("PatchElement: %v", err)
	}

	el := currentElement(t, docs)
	if el.Content == nil || *el.Content != "new" {
		t.Fatalf("content not applied: %v", el.Content)
	}
	if el.Src == nil || *el.Src != "a.png" {
		t.Fatalf("src not applied: %v", el.Src)
	}
	if !el.IsLocked {
		t.Fatalf("isLocked not applied")
	}
	if !reflect.DeepEqual(el.Settings, map[string]any{"keep": "me"}) {
		t.Fatalf("settings were modified by properties: %v", el.Settings)
	}
}

func TestPatchElement_PropertiesAndSettingsTogether(t *testing.T) {
	uc, docs := newPatchFixture(t)

	_, err := uc.PatchElement(context.Background(), testProject, testPage, testElemID,
		map[string]any{"name": "Hero"}, nil, map[string]any{"extra": true})
	if err != nil {
		t.Fatalf("PatchElement: %v", err)
	}
	el := currentElement(t, docs)
	if el.Name == nil || *el.Name != "Hero" {
		t.Fatalf("name not applied: %v", el.Name)
	}
	if el.Settings["keep"] != "me" || el.Settings["extra"] != true {
		t.Fatalf("settings should be merged, got %v", el.Settings)
	}
}

func TestPatchElement_NullDeletesKeysAndDeepMerges(t *testing.T) {
	uc, docs := newPatchFixture(t)

	_, err := uc.PatchElement(context.Background(), testProject, testPage, testElemID, nil,
		map[string]any{"color": nil, "nested": map[string]any{"a": nil, "c": 3.0}}, nil)
	if err != nil {
		t.Fatalf("PatchElement: %v", err)
	}

	el := currentElement(t, docs)
	if _, ok := el.Styles["color"]; ok {
		t.Fatalf("null should delete style key, got %v", el.Styles)
	}
	want := map[string]any{"b": 2.0, "c": 3.0}
	if !reflect.DeepEqual(el.Styles["nested"], want) {
		t.Fatalf("nested styles: got %v want %v", el.Styles["nested"], want)
	}
}

func TestPatchElement_UnsupportedPropertyIsRejectedWithoutSeq(t *testing.T) {
	uc, docs := newPatchFixture(t)

	_, err := uc.PatchElement(context.Background(), testProject, testPage, testElemID,
		map[string]any{"bogus": 1}, map[string]any{"color": "red"}, nil)
	if !domainErrors.IsCode(err, domainErrors.CodeValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if got := docs.seq(testProject, testPage); got != 0 {
		t.Fatalf("seq must not advance on failure, got %d", got)
	}
	// Nothing may have been written (the document was never cached) or it is unchanged.
	if el := findElementByID(docs.snapshot(testProject, testPage), testElemID); el != nil && el.Styles["color"] != "black" {
		t.Fatalf("failed op must not partially apply, styles=%v", el.Styles)
	}
}

func TestSeq_IncrementsPerSuccessfulOpOnly(t *testing.T) {
	uc, docs := newPatchFixture(t)
	ctx := context.Background()

	seq, err := uc.PatchElement(ctx, testProject, testPage, testElemID, nil, map[string]any{"a": 1}, nil)
	if err != nil || seq != 1 {
		t.Fatalf("first op: seq=%d err=%v", seq, err)
	}

	// failure: unknown element
	if _, err := uc.PatchElement(ctx, testProject, testPage, "22222222-2222-2222-2222-222222222222", nil, map[string]any{"a": 1}, nil); !domainErrors.IsCode(err, domainErrors.CodeNotFound) {
		t.Fatalf("expected NOT_FOUND, got %v", err)
	}
	if _, err := uc.RemoveElement(ctx, testProject, testPage, "does-not-exist"); !domainErrors.IsCode(err, domainErrors.CodeNotFound) {
		t.Fatalf("expected NOT_FOUND from remove, got %v", err)
	}
	if _, _, err := uc.MoveElement(ctx, testProject, testPage, "does-not-exist", "", 0); err == nil {
		t.Fatalf("expected move of unknown element to fail")
	}
	if got := docs.seq(testProject, testPage); got != 1 {
		t.Fatalf("failures must not consume seq; got %d", got)
	}

	newEl := entity.Element{Type: "text", PageID: testPage}
	seq, err = uc.CreateElement(ctx, testProject, &newEl)
	if err != nil || seq != 2 {
		t.Fatalf("create: seq=%d err=%v", seq, err)
	}
	_, seq, err = uc.MoveElement(ctx, testProject, testPage, newEl.ID, "", 0)
	if err != nil || seq != 3 {
		t.Fatalf("move: seq=%d err=%v", seq, err)
	}
	seq, err = uc.RemoveElement(ctx, testProject, testPage, newEl.ID)
	if err != nil || seq != 4 {
		t.Fatalf("remove: seq=%d err=%v", seq, err)
	}

	// sync reports the same seq as the last op.
	_, syncSeq, err := uc.SyncPageWithSeq(ctx, testProject, testPage)
	if err != nil || syncSeq != 4 {
		t.Fatalf("sync seq=%d err=%v", syncSeq, err)
	}
}

func TestPatchElement_ErrorIsAppError(t *testing.T) {
	uc, _ := newPatchFixture(t)
	_, err := uc.PatchElement(context.Background(), testProject, testPage, "nope", nil, nil, nil)
	var appErr *domainErrors.AppError
	if !errors.As(err, &appErr) || appErr.Code != domainErrors.CodeNotFound {
		t.Fatalf("expected AppError NOT_FOUND, got %v", err)
	}
}
