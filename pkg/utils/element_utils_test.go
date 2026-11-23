package utils

import (
	"encoding/json"
	"fmt"
	"testing"

	"my-go-app/internal/models"
)

func el(id, typ string, parent *string) *models.Element {
	return &models.Element{Id: id, Type: typ, ParentId: parent}
}

func childIDs(t *testing.T, e models.EditorElement) []string {
	t.Helper()
	kids := GetChildrenFromEditorElement(e)
	ids := make([]string, len(kids))
	for i, k := range kids {
		ee, ok := k.(models.EditorElement)
		if !ok {
			t.Fatalf("child %d is %T, not an EditorElement", i, k)
		}
		ids[i] = ee.GetElement().Id
	}
	return ids
}

func assertStringIDs(t *testing.T, label string, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s = %v, want %v", label, got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("%s = %v, want %v", label, got, want)
		}
	}
}

func TestBuildElementTreeNestsChildrenUnderContainers(t *testing.T) {
	sec := "sec"
	frame := "frame"
	in := []models.EditorElement{
		el("sec", "Section", nil),
		el("frame", "Frame", &sec),
		el("text1", "Text", &frame),
		el("btn", "Button", &sec),
		el("sec2", "Section", nil),
	}

	roots := BuildElementTree(in)
	if len(roots) != 2 {
		t.Fatalf("roots must keep input order: %v", roots)
	}
	if roots[0].GetElement().Id != "sec" || roots[1].GetElement().Id != "sec2" {
		t.Fatalf("roots must keep input order: %v", roots)
	}
	assertStringIDs(t, "section children", childIDs(t, roots[0]), []string{"frame", "btn"})
	frameNode := GetChildrenFromEditorElement(roots[0])[0].(models.EditorElement)
	assertStringIDs(t, "frame children", childIDs(t, frameNode), []string{"text1"})
	if got := GetChildrenFromEditorElement(roots[1]); got != nil {
		t.Errorf("childless container should be returned as-is, got children %v", got)
	}
}

func TestBuildElementTreeIgnoresChildrenOfNonContainersAndOrphans(t *testing.T) {
	text := "text"
	ghost := "does-not-exist"
	in := []models.EditorElement{
		el("text", "Text", nil),
		el("under-text", "Button", &text),    // parent is not a container
		el("orphan", "Text", &ghost),         // parent missing from input
		(*models.Element)(nil),               // typed nil is skipped, not a panic
	}
	roots := BuildElementTree(in)
	if len(roots) != 1 || roots[0].GetElement().Id != "text" {
		t.Fatalf("only the true root survives: %v", roots)
	}
	if kids := GetChildrenFromEditorElement(roots[0]); kids != nil {
		t.Errorf("non-container must not gain children: %v", kids)
	}
}

func TestBuildElementTreeEmpty(t *testing.T) {
	if got := BuildElementTree(nil); len(got) != 0 {
		t.Errorf("nil input: %v", got)
	}
	if got := BuildElementTree([]models.EditorElement{}); len(got) != 0 {
		t.Errorf("empty input: %v", got)
	}
}

// Many roots exercise the goroutine fan-out; the index bookkeeping must put
// every result back in its original slot.
func TestBuildElementTreeConcurrentPreservesOrder(t *testing.T) {
	const n = 200
	var in []models.EditorElement
	for i := 0; i < n; i++ {
		id := fmt.Sprintf("s%03d", i)
		parent := id
		in = append(in, el(id, "Section", nil), el("c-"+id, "Text", &parent))
	}
	roots := BuildElementTree(in)
	if len(roots) != n {
		t.Fatalf("got %d roots", len(roots))
	}
	for i, r := range roots {
		id := fmt.Sprintf("s%03d", i)
		if r.GetElement().Id != id {
			t.Fatalf("root %d is %s, want %s", i, r.GetElement().Id, id)
		}
		if kids := childIDs(t, r); len(kids) != 1 || kids[0] != "c-"+id {
			t.Fatalf("root %s children = %v", id, kids)
		}
	}
}

func TestBuildElementWithChildrenNilBase(t *testing.T) {
	var nilEl *models.Element
	got := buildElementWithChildren(nilEl, nil)
	if got != models.EditorElement(nilEl) {
		t.Errorf("nil element must be returned unchanged, got %v", got)
	}
}

func TestElementWithChildrenJSON(t *testing.T) {
	sec := "sec"
	roots := BuildElementTree([]models.EditorElement{
		el("sec", "Section", nil),
		el("t", "Text", &sec),
	})
	b, err := json.Marshal(roots[0])
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	if m["id"] != "sec" || m["type"] != "Section" {
		t.Errorf("base fields must be flattened into the object: %s", b)
	}
	kids, ok := m["elements"].([]any)
	if !ok || len(kids) != 1 || kids[0].(map[string]any)["id"] != "t" {
		t.Errorf("children must serialise under \"elements\": %s", b)
	}
}

func TestElementWithChildrenAccessors(t *testing.T) {
	base := el("x", "Frame", nil)
	w := WrapElementWithChildren(base, []models.EditorElement{el("y", "Text", nil)})
	if w.GetElement() != base || w.GetType() != "Frame" {
		t.Errorf("accessors wrong: %v %q", w.GetElement(), w.GetType())
	}
	empty := &ElementWithChildren{}
	if empty.GetElement() != nil || empty.GetType() != "" {
		t.Errorf("nil base must give zero values, got %v %q", empty.GetElement(), empty.GetType())
	}
}

func TestConvertEditorElementsToAny(t *testing.T) {
	a, b := el("a", "Text", nil), el("b", "Text", nil)
	got := ConvertEditorElementsToAny([]models.EditorElement{a, b})
	if len(got) != 2 || got[0] != any(a) || got[1] != any(b) {
		t.Errorf("got %v", got)
	}
	if out := ConvertEditorElementsToAny(nil); out == nil || len(out) != 0 {
		t.Errorf("nil input should give empty non-nil slice, got %#v", out)
	}
}

func TestGetChildrenFromEditorElement(t *testing.T) {
	if got := GetChildrenFromEditorElement(nil); got != nil {
		t.Errorf("nil element: %v", got)
	}
	if got := GetChildrenFromEditorElement((*models.Element)(nil)); got != nil {
		t.Errorf("typed nil element: %v", got)
	}
	if got := GetChildrenFromEditorElement(el("t", "Text", nil)); got != nil {
		t.Errorf("non-container: %v", got)
	}
	if got := GetChildrenFromEditorElement(el("s", "Section", nil)); got != nil {
		t.Errorf("plain container carries no children: %v", got)
	}
	w := WrapElementWithChildren(el("s", "Section", nil), []models.EditorElement{el("c", "Text", nil)})
	if got := GetChildrenFromEditorElement(w); len(got) != 1 {
		t.Errorf("wrapper children = %v", got)
	}
}

type fakeEditorElement struct{ models.Element }

func TestConvertToEditorElementPassthroughAndLeafMap(t *testing.T) {
	in := el("a", "Text", nil)
	got, err := ConvertToEditorElement(in)
	if err != nil || got != models.EditorElement(in) {
		t.Fatalf("got %v %v", got, err)
	}

	got, err = ConvertToEditorElement(map[string]any{"id": "t1", "type": "Text", "content": "hi", "order": 2})
	if err != nil {
		t.Fatal(err)
	}
	e, ok := got.(*models.Element)
	if !ok || e.Id != "t1" {
		t.Fatalf("got %#v", got)
	}
	if e.GetType() != "Text" || *e.Content != "hi" || e.Order != 2 {
		t.Fatalf("got %#v", got)
	}
}

func TestConvertToEditorElementContainerWithNested(t *testing.T) {
	got, err := ConvertToEditorElement(map[string]any{
		"id": "s1", "type": "Section",
		"elements": []any{map[string]any{"id": "t1", "type": "Text"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	w, ok := got.(*ElementWithChildren)
	if !ok {
		t.Fatalf("want *ElementWithChildren, got %T", got)
	}
	if w.GetElement().Id != "s1" || len(w.Elements) != 1 {
		t.Fatalf("got %#v", w)
	}
	child, err := ConvertToEditorElement(w.Elements[0])
	if err != nil || child.GetElement().Id != "t1" {
		t.Fatalf("child conversion: %v %v", child, err)
	}
}

func TestConvertToEditorElementContainerFallbacks(t *testing.T) {
	got, err := ConvertToEditorElement(map[string]any{"id": "s1", "type": "Section", "elements": []any{}})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := got.(*models.Element); !ok {
		t.Fatalf("want *models.Element, got %T", got)
	}

	got, err = ConvertToEditorElement(map[string]any{"id": "s1", "type": "Section", "elements": "bogus"})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := got.(*models.Element); !ok || got.GetElement().Id != "s1" {
		t.Fatalf("got %#v", got)
	}

	got, err = ConvertToEditorElement(map[string]any{"id": "t", "type": "Text", "elements": []any{map[string]any{"id": "x"}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := got.(*models.Element); !ok {
		t.Fatalf("got %T", got)
	}
}

func TestConvertToEditorElementStructAndErrors(t *testing.T) {
	got, err := ConvertToEditorElement(struct {
		Id   string `json:"id"`
		Type string `json:"type"`
	}{"z", "Button"})
	if err != nil || got.GetElement().Id != "z" {
		t.Fatalf("got %v %v", got, err)
	}
	if !got.GetElement().IsButton() {
		t.Fatalf("expected Button, got %v", got.GetElement().Type)
	}

	if _, err := ConvertToEditorElement(make(chan int)); err == nil {
		t.Fatal("expected marshal error")
	}
	if _, err := ConvertToEditorElement(map[string]any{"id": 12, "type": "Text"}); err == nil {
		t.Fatal("expected unmarshal error for numeric id")
	}
}
