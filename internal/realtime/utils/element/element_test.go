package element

import (
	"testing"

	"my-go-app/internal/realtime/domain/entity"
)

func strPtr(s string) *string { return &s }

func TestParentsEqual(t *testing.T) {
	if !ParentsEqual(nil, nil) {
		t.Fatal("nil parents should match")
	}
	if ParentsEqual(nil, strPtr("a")) || ParentsEqual(strPtr("a"), nil) {
		t.Fatal("nil and non-nil should not match")
	}
	if !ParentsEqual(strPtr("a"), strPtr("a")) || ParentsEqual(strPtr("a"), strPtr("b")) {
		t.Fatal("pointer comparison mismatch")
	}
}

func TestCompareParent(t *testing.T) {
	if !CompareParent(nil, "") || CompareParent(nil, "a") {
		t.Fatal("unexpected CompareParent nil result")
	}
	if !CompareParent(strPtr("a"), "a") || CompareParent(strPtr("a"), "b") {
		t.Fatal("unexpected CompareParent value result")
	}
}

func TestNormalizeOrder(t *testing.T) {
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
	if items[0].Elements[0].Order != 0 {
		t.Fatalf("nested child order not normalized: %+v", items[0].Elements[0])
	}
	if items[0].Elements[0].OrderKey == "" {
		t.Fatalf("nested child order key missing: %+v", items[0].Elements[0])
	}
}

func TestInsertAt(t *testing.T) {
	items := []entity.Element{
		{ID: "e1", Type: "Text", PageID: "p1"},
		{ID: "e2", Type: "Text", PageID: "p1", OrderKey: "a1"},
		{ID: "e3", Type: "Text", PageID: "p1", OrderKey: "a0"},
	}
	if err := NormalizeOrder(items); err != nil {
		t.Fatalf("NormalizeOrder error: %v", err)
	}
	inserted, err := InsertAt(items, entity.Element{ID: "mid", Type: "Text", PageID: "p1"}, 1)
	if err != nil {
		t.Fatalf("InsertAt error: %v", err)
	}
	if len(inserted) != 4 || inserted[1].ID != "mid" {
		t.Fatalf("unexpected inserted items: %+v", inserted)
	}
	if inserted[1].Order != 1 || inserted[3].Order != 3 {
		t.Fatalf("unexpected inserted order: %+v", inserted)
	}
	if inserted[0].OrderKey >= inserted[1].OrderKey {
		t.Fatalf("inserted[0] key >= inserted[1]: %q >= %q", inserted[0].OrderKey, inserted[1].OrderKey)
	}
	if inserted[1].OrderKey >= inserted[2].OrderKey {
		t.Fatalf("inserted[1] key >= inserted[2]: %q >= %q", inserted[1].OrderKey, inserted[2].OrderKey)
	}
	if _, err := InsertAt(items, entity.Element{ID: "bad"}, -1); err == nil {
		t.Fatal("expected out of range error")
	}
}

func TestBuildElementTreeAndFlatten(t *testing.T) {
	flat := []entity.Element{
		{ID: "r2", Type: "Frame", PageID: "p1", Order: 1, OrderKey: "a1"},
		{ID: "c2", Type: "Text", PageID: "p1", ParentID: strPtr("r1"), Order: 1, OrderKey: "b1"},
		{ID: "c1", Type: "Text", PageID: "p1", ParentID: strPtr("r1"), Order: 0, OrderKey: "b0"},
		{ID: "r1", Type: "Frame", PageID: "p1", Order: 0, OrderKey: "a0"},
		{ID: "orphan", Type: "Text", PageID: "p1", ParentID: strPtr("missing"), Order: 2, OrderKey: "a2"},
	}
	tree := BuildElementTree(flat)
	if len(tree) != 3 || tree[0].ID != "r1" {
		t.Fatalf("unexpected roots: %+v", tree)
	}
	if tree[1].ID != "r2" || tree[2].ID != "orphan" {
		t.Fatalf("unexpected root order: %+v", tree)
	}
	if len(tree[0].Elements) != 2 {
		t.Fatalf("expected 2 children on r1: %+v", tree[0].Elements)
	}
	if tree[0].Elements[0].ID != "c1" || tree[0].Elements[1].ID != "c2" {
		t.Fatalf("unexpected children order: %+v", tree[0].Elements)
	}
	if got := FlattenElements(tree); len(got) != 5 {
		t.Fatalf("expected 5 flattened elements, got %d", len(got))
	}
}

func TestTreeCreateFindSwapUpdateDelete(t *testing.T) {
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
}

func TestTreeChildrenAndDetach(t *testing.T) {
	root := entity.Element{ID: "r1", Type: "Frame", PageID: "p1", Elements: []entity.Element{{ID: "c3"}, {ID: "c1"}}}
	children := Children(&root.Elements, "")
	if len(*children) != 2 {
		t.Fatalf("expected 2 root children, got %d", len(*children))
	}
	detached := Detach(&root.Elements, "c3")
	if detached == nil || detached.ID != "c3" {
		t.Fatalf("unexpected Detach result: %+v", detached)
	}
	if len(root.Elements) != 1 || root.Elements[0].Order != 0 {
		t.Fatalf("unexpected remaining elements: %+v", root.Elements)
	}
	if Detach(&root.Elements, "missing") != nil {
		t.Fatal("expected nil for missing Detach")
	}
}

func TestMergeMap(t *testing.T) {
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
	if _, hasColor := style["color"]; hasColor {
		t.Fatalf("expected color removed: %v", style)
	}
	if style["margin"] != "4px" || style["padding"] != "8px" {
		t.Fatalf("unexpected merged style: %v", style)
	}
}

func TestApplyPropertiesValid(t *testing.T) {
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
	if *el.Name != "Hero" || *el.TailwindStyles != "p-4" {
		t.Fatalf("unexpected element strings after ApplyProperties: %+v", el)
	}
	if !el.IsLocked {
		t.Fatalf("expected IsLocked true: %+v", el)
	}
	if err := ApplyProperties(el, map[string]any{"content": nil}); err != nil || el.Content != nil {
		t.Fatalf("expected nil clear, got err=%v content=%v", err, el.Content)
	}
}

func TestApplyPropertiesErrors(t *testing.T) {
	el := &entity.Element{}
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

func TestMergeMap_CapsDepthAtMaxMergeDepth(t *testing.T) {
	makeDeep := func(depth int, leafKey string, leafVal any) map[string]any {
		cur := map[string]any{leafKey: leafVal}
		for i := depth; i >= 1; i-- {
			cur = map[string]any{"n": cur}
		}
		return cur
	}
	dst := makeDeep(MaxMergeDepth+2, "keep", 1)
	patch := makeDeep(MaxMergeDepth+2, "added", 2)
	merged := MergeMap(dst, patch)

	// Walk down MaxMergeDepth levels to verify recursion stopped at MaxMergeDepth.
	cur := merged
	for i := 0; i < MaxMergeDepth; i++ {
		cur = cur["n"].(map[string]any)
	}
	sub := cur["n"].(map[string]any)
	if _, hasKeep := sub["keep"]; hasKeep {
		t.Fatalf("expected merge recursion to stop at depth %d", MaxMergeDepth)
	}
}

func TestFlattenElements_CapsDepthAndNodeCount(t *testing.T) {
	deep := entity.Element{ID: "leaf"}
	for i := 0; i < MaxElementDepth+5; i++ {
		deep = entity.Element{ID: "n", Elements: []entity.Element{deep}}
	}
	flat := FlattenElements([]entity.Element{deep})
	if len(flat) != MaxElementDepth {
		t.Fatalf("expected FlattenElements to cap at %d depth, got %d", MaxElementDepth, len(flat))
	}

	wide := make([]entity.Element, MaxElementNodes+50)
	for i := range wide {
		wide[i] = entity.Element{ID: "x"}
	}
	if got := len(FlattenElements(wide)); got != MaxElementNodes {
		t.Fatalf("expected FlattenElements to cap at %d nodes, got %d", MaxElementNodes, got)
	}
}
