package element

import (
	"testing"

	"my-go-app/internal/realtime/domain/entity"
)

func strPtr(s string) *string { return &s }

func TestParentComparisons(t *testing.T) {
	if !ParentsEqual(nil, nil) || ParentsEqual(nil, strPtr("a")) || ParentsEqual(strPtr("a"), nil) {
		t.Fatal("unexpected nil parent comparison")
	}
	if !ParentsEqual(strPtr("a"), strPtr("a")) || ParentsEqual(strPtr("a"), strPtr("b")) {
		t.Fatal("unexpected pointer parent comparison")
	}
	if !CompareParent(nil, "") || CompareParent(nil, "a") || !CompareParent(strPtr("a"), "a") || CompareParent(strPtr("a"), "b") {
		t.Fatal("unexpected CompareParent result")
	}
}

func TestNormalizeOrderAndInsertAt(t *testing.T) {
	items := []entity.Element{
		{ID: "e1", Type: "Text", PageID: "p1", Elements: []entity.Element{{ID: "c1", Type: "Text", PageID: "p1"}}},
		{ID: "e2", Type: "Text", PageID: "p1", OrderKey: "a1"},
		{ID: "e3", Type: "Text", PageID: "p1", OrderKey: "a0"},
	}
	if err := NormalizeOrder(items); err != nil {
		t.Fatalf("NormalizeOrder error: %v", err)
	}
	for i := range items {
		if items[i].Order != i || items[i].OrderKey == "" {
			t.Fatalf("element %d not normalized: %+v", i, items[i])
		}
		if i > 0 && items[i].OrderKey <= items[i-1].OrderKey {
			t.Fatalf("keys not strictly increasing: %q <= %q", items[i].OrderKey, items[i-1].OrderKey)
		}
	}
	if items[0].Elements[0].OrderKey == "" || items[0].Elements[0].Order != 0 {
		t.Fatalf("nested child not normalized: %+v", items[0].Elements[0])
	}

	inserted, err := InsertAt(items, entity.Element{ID: "mid", Type: "Text", PageID: "p1"}, 1)
	if err != nil {
		t.Fatalf("InsertAt error: %v", err)
	}
	if len(inserted) != 4 || inserted[1].ID != "mid" || inserted[1].Order != 1 || inserted[3].Order != 3 {
		t.Fatalf("unexpected inserted order: %+v", inserted)
	}
	if !(inserted[0].OrderKey < inserted[1].OrderKey && inserted[1].OrderKey < inserted[2].OrderKey) {
		t.Fatalf("inserted key out of order: %q, %q, %q", inserted[0].OrderKey, inserted[1].OrderKey, inserted[2].OrderKey)
	}
	if _, err := InsertAt(items, entity.Element{ID: "bad"}, -1); err == nil {
		t.Fatal("expected out of range error")
	}
}

func TestTreeOperationsBuildFlattenMutateAndDetach(t *testing.T) {
	flat := []entity.Element{
		{ID: "r2", Type: "Frame", PageID: "p1", Order: 1, OrderKey: "a1"},
		{ID: "c2", Type: "Text", PageID: "p1", ParentID: strPtr("r1"), Order: 1, OrderKey: "b1"},
		{ID: "c1", Type: "Text", PageID: "p1", ParentID: strPtr("r1"), Order: 0, OrderKey: "b0"},
		{ID: "r1", Type: "Frame", PageID: "p1", Order: 0, OrderKey: "a0"},
		{ID: "orphan", Type: "Text", PageID: "p1", ParentID: strPtr("missing"), Order: 2, OrderKey: "a2"},
	}
	tree := BuildElementTree(flat)
	if len(tree) != 3 || tree[0].ID != "r1" || tree[1].ID != "r2" || tree[2].ID != "orphan" {
		t.Fatalf("unexpected roots: %+v", tree)
	}
	if len(tree[0].Elements) != 2 || tree[0].Elements[0].ID != "c1" || tree[0].Elements[1].ID != "c2" {
		t.Fatalf("unexpected children order: %+v", tree[0].Elements)
	}

	if got := FlattenElements(tree); len(got) != 5 {
		t.Fatalf("expected 5 flattened elements, got %d", len(got))
	}

	root := entity.Element{ID: "r1", Type: "Frame", PageID: "p1", Elements: []entity.Element{{ID: "c1"}, {ID: "c2"}}}
	if !CreateElement(&root, entity.Element{ID: "c3", ParentID: strPtr("r1")}) || len(root.Elements) != 3 {
		t.Fatal("CreateElement failed")
	}
	if FindByID([]entity.Element{root}, "c3") == nil || FindByID([]entity.Element{root}, "none") != nil {
		t.Fatal("FindByID mismatch")
	}
	if !SwapElements(&root, "c1", "c3") || root.Elements[0].ID != "c3" {
		t.Fatal("SwapElements failed")
	}
	if !UpdateElement(&root, entity.Element{ID: "c3", Type: "Button"}) || root.Elements[0].Type != "Button" {
		t.Fatal("UpdateElement failed")
	}
	if !DeleteElement(&root, "c2") || len(root.Elements) != 2 {
		t.Fatal("DeleteElement failed")
	}

	children := Children(&root.Elements, "")
	if len(*children) != 2 {
		t.Fatalf("expected 2 root children, got %d", len(*children))
	}
	detached := Detach(&root.Elements, "c3")
	if detached == nil || detached.ID != "c3" || len(root.Elements) != 1 || root.Elements[0].Order != 0 {
		t.Fatalf("unexpected Detach result: %+v remaining=%+v", detached, root.Elements)
	}
	if Detach(&root.Elements, "missing") != nil {
		t.Fatal("expected nil for missing Detach")
	}
}

func TestMergeMapAndApplyProperties(t *testing.T) {
	dst := map[string]any{
		"keep":   1,
		"remove": "gone",
		"style":  map[string]any{"color": "red", "margin": "4px"},
	}
	patch := map[string]any{
		"remove": nil,
		"added":  map[string]any{"skip": nil, "x": 2},
		"style":  map[string]any{"color": nil, "padding": "8px"},
	}
	merged := MergeMap(dst, patch)
	if _, ok := merged["remove"]; ok {
		t.Fatal("nil should delete key")
	}
	if dst["remove"] != "gone" {
		t.Fatal("dst should not be mutated")
	}
	style := merged["style"].(map[string]any)
	if _, hasColor := style["color"]; hasColor || style["margin"] != "4px" || style["padding"] != "8px" {
		t.Fatalf("unexpected merged style: %v", style)
	}

	el := &entity.Element{}
	if err := ApplyProperties(el, map[string]any{
		"name":           "Hero",
		"content":        "Hello",
		"href":           "/go",
		"src":            "/img.png",
		"tailwindStyles": "p-4",
		"isLocked":       true,
	}); err != nil {
		t.Fatalf("ApplyProperties error: %v", err)
	}
	if *el.Name != "Hero" || *el.TailwindStyles != "p-4" || !el.IsLocked {
		t.Fatalf("unexpected element after ApplyProperties: %+v", el)
	}
	if err := ApplyProperties(el, map[string]any{"content": nil}); err != nil || el.Content != nil {
		t.Fatalf("expected nil clear, got err=%v content=%v", err, el.Content)
	}
	if err := ApplyProperties(el, map[string]any{"name": 123}); err == nil {
		t.Fatal("expected non-string error")
	}
	if err := ApplyProperties(el, map[string]any{"isLocked": "yes"}); err == nil {
		t.Fatal("expected non-bool error")
	}
	if err := ApplyProperties(el, map[string]any{"unknown": "x"}); err == nil {
		t.Fatal("expected unsupported key error")
	}
}
