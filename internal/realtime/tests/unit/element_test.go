package unit

import (
	"reflect"
	"testing"

	"my-go-app/internal/realtime/domain/entity"
	"my-go-app/internal/realtime/utils/element"
)

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func idsFromElements(elements []entity.Element) []string {
	ids := make([]string, 0, len(elements))
	for _, e := range elements {
		ids = append(ids, e.ID)
	}
	return ids
}

func makeEl(id string, children ...entity.Element) entity.Element {
	return entity.Element{
		ID:       id,
		Elements: children,
	}
}

type flattenCase struct {
	name    string
	input   []entity.Element
	wantIDs []string
}

func assertFlattenCase(t *testing.T, tc flattenCase) {
	t.Helper()
	var origFirstRootChildren int
	if len(tc.input) > 0 {
		origFirstRootChildren = len(tc.input[0].Elements)
	}
	got := element.Flatten(tc.input)
	if len(tc.wantIDs) == 0 {
		if len(got) != 0 {
			t.Fatalf("expected empty result; got %d elements: %v", len(got), got)
		}
		return
	}
	gotIDs := idsFromElements(got)
	if !reflect.DeepEqual(gotIDs, tc.wantIDs) {
		t.Fatalf("unexpected flattened order/ids\ngot:  %v\nwant: %v", gotIDs, tc.wantIDs)
	}
	if len(tc.input) > 0 && origFirstRootChildren != len(tc.input[0].Elements) {
		t.Fatalf("input mutated: expected first root child count %d, got %d", origFirstRootChildren, len(tc.input[0].Elements))
	}
	assertNestedPreserved(t, tc.name, gotIDs, got)
}

func assertNestedPreserved(t *testing.T, caseName string, gotIDs []string, got []entity.Element) {
	t.Helper()
	if caseName != "nested elements (multiple roots and nested children)" {
		return
	}
	for i, id := range gotIDs {
		if id == "r1c1" {
			if len(got[i].Elements) != 2 {
				t.Fatalf("expected r1c1 to have 2 child elements preserved, got %d", len(got[i].Elements))
			}
			return
		}
	}
	t.Fatalf("r1c1 not found in flattened result: %v", gotIDs)
}

func TestFlattenElements_TableDriven(t *testing.T) {
	cases := []flattenCase{
		{name: "nil input returns empty", input: nil, wantIDs: []string{}},
		{name: "single element", input: []entity.Element{makeEl("root")}, wantIDs: []string{"root"}},
		{
			name: "nested elements (multiple roots and nested children)",
			input: []entity.Element{
				makeEl("root1", makeEl("r1c1", makeEl("r1c1a"), makeEl("r1c1b")), makeEl("r1c2")),
				makeEl("root2", makeEl("r2c1")),
			},
			wantIDs: []string{"root1", "r1c1", "r1c1a", "r1c1b", "r1c2", "root2", "r2c1"},
		},
		{name: "deep chain", input: []entity.Element{makeEl("a", makeEl("b", makeEl("c")))}, wantIDs: []string{"a", "b", "c"}},
	}
	for _, tc := range cases {
		assertFlattenCase(t, tc)
	}
}

func TestFlattenElements_PreservesProperties(t *testing.T) {
	root := entity.Element{
		ID:   "withprops",
		Type: "button",
		Settings: map[string]any{
			"foo": "bar",
		},
		Elements: []entity.Element{
			{ID: "child"},
		},
	}

	got := element.Flatten([]entity.Element{root})

	if len(got) != 2 {
		t.Fatalf("expected 2 elements in flattened result, got %d", len(got))
	}

	// root should appear first
	rootGot := got[0]
	if rootGot.ID != "withprops" {
		t.Fatalf("expected root id %q, got %q", "withprops", rootGot.ID)
	}
	if rootGot.Type != "button" {
		t.Fatalf("expected root type %q, got %q", "button", rootGot.Type)
	}
	if rootGot.Settings == nil || rootGot.Settings["foo"] != "bar" {
		t.Fatalf("expected setting foo=bar, got %v", rootGot.Settings)
	}
}

// findElementByID searches the tree and returns a pointer to the element with the given id (or nil if not found).
func findElementByID(elements []entity.Element, id string) *entity.Element {
	for i := range elements {
		if elements[i].ID == id {
			return &elements[i]
		}
		if el := findElementByID(elements[i].Elements, id); el != nil {
			return el
		}
	}
	return nil
}

func TestBuildElementTree_Basic(t *testing.T) {
	elements := []entity.Element{
		{ID: "root"},
		{ID: "child", ParentID: strPtr("root")},
	}

	got := element.BuildTree(elements)

	if len(got) != 1 {
		t.Fatalf("expected 1 root element, got %d", len(got))
	}
	if got[0].ID != "root" {
		t.Fatalf("expected root id %q, got %q", "root", got[0].ID)
	}
	if len(got[0].Elements) != 1 {
		t.Fatalf("expected 1 child under root, got %d", len(got[0].Elements))
	}
	if got[0].Elements[0].ID != "child" {
		t.Fatalf("expected child id %q, got %q", "child", got[0].Elements[0].ID)
	}

	// ensure input wasn't mutated
	if len(elements[0].Elements) != 0 {
		t.Fatalf("expected input element's Elements to remain unchanged; got %d", len(elements[0].Elements))
	}
}

func assertTreeChildrenIDs(t *testing.T, roots []entity.Element, parentID string, want []string) {
	t.Helper()
	node := findElementByID(roots, parentID)
	if node == nil || len(node.Elements) != len(want) {
		t.Fatalf("expected %s to have %d children, got %v", parentID, len(want), node)
	}
	if gotIDs := idsFromElements(node.Elements); !reflect.DeepEqual(gotIDs, want) {
		t.Fatalf("unexpected %s children ids: got %v want %v", parentID, gotIDs, want)
	}
}

func TestBuildElementTree_MultipleRootsAndSiblings(t *testing.T) {
	elements := []entity.Element{
		{ID: "r1"},
		{ID: "r2"},
		{ID: "r1c1", ParentID: strPtr("r1")},
		{ID: "r1c2", ParentID: strPtr("r1")},
		{ID: "r2c1", ParentID: strPtr("r2")},
	}

	got := element.BuildTree(elements)
	if len(got) != 2 || got[0].ID != "r1" || got[1].ID != "r2" {
		t.Fatalf("unexpected root order/ids: %v", idsFromElements(got))
	}
	assertTreeChildrenIDs(t, got, "r1", []string{"r1c1", "r1c2"})
	assertTreeChildrenIDs(t, got, "r2", []string{"r2c1"})
}

func TestBuildElementTree_OrphanBecomesRoot(t *testing.T) {
	elements := []entity.Element{
		{ID: "orphan", ParentID: strPtr("missing")},
	}

	got := element.BuildTree(elements)
	if len(got) != 1 || got[0].ID != "orphan" {
		t.Fatalf("expected orphan to become root, got %v", idsFromElements(got))
	}
}

func TestBuildElementTree_NoDuplicateRoots(t *testing.T) {
	elements := []entity.Element{
		{ID: "r"},
	}

	got := element.BuildTree(elements)
	if len(got) != 1 {
		t.Fatalf("expected 1 root, got %d (duplicate roots)", len(got))
	}
}

func TestBuildElementTree_SortsChildrenByOrder(t *testing.T) {
	elements := []entity.Element{
		{ID: "root"},
		{ID: "c1", ParentID: strPtr("root"), Order: 2},
		{ID: "c2", ParentID: strPtr("root"), Order: 1},
	}

	got := element.BuildTree(elements)
	root := findElementByID(got, "root")
	if root == nil {
		t.Fatalf("root not found in tree: %v", idsFromElements(got))
	}
	if len(root.Elements) != 2 {
		t.Fatalf("expected 2 children, got %d", len(root.Elements))
	}
	if root.Elements[0].ID != "c2" || root.Elements[1].ID != "c1" {
		t.Fatalf("children not sorted by Order: %v", idsFromElements(root.Elements))
	}
}

func TestBuildElementTree_SortsRootsByOrderAndTieBreakers(t *testing.T) {
	elements := []entity.Element{
		{ID: "rootA", Order: 2},
		{ID: "rootB", Order: 1},
		{ID: "rootC", Order: 1},
		{ID: "rootD", Order: 1},
	}

	got := element.BuildTree(elements)

	ids := idsFromElements(got)
	// Elements should be ordered by Order, then by ID for deterministic tie-breaking.
	expected := []string{"rootB", "rootC", "rootD", "rootA"}
	if !reflect.DeepEqual(ids, expected) {
		t.Fatalf("unexpected root order/ids: got %v, want %v", ids, expected)
	}
}

func TestUpdateElement_ReplacesNested(t *testing.T) {
	root := makeEl("root",
		makeEl("child", makeEl("grand")),
	)

	updated := entity.Element{ID: "grand", Type: "updated"}

	ok := element.Update(&root, updated)
	if !ok {
		t.Fatalf("expected UpdateElement to return true")
	}

	el := findElementByID([]entity.Element{root}, "grand")
	if el == nil || el.Type != "updated" {
		t.Fatalf("expected grandchild to be updated; got %#v", el)
	}
}

func TestDeleteElement_DeletesNested(t *testing.T) {
	root := makeEl("root",
		makeEl("child", makeEl("grand")),
	)

	if !element.Delete(&root, "grand") {
		t.Fatalf("expected DeleteElement to return true")
	}

	if findElementByID([]entity.Element{root}, "grand") != nil {
		t.Fatalf("expected grandchild to be deleted")
	}

	child := findElementByID([]entity.Element{root}, "child")
	if child == nil {
		t.Fatalf("expected child to remain")
	}
	if len(child.Elements) != 0 {
		t.Fatalf("expected child to have 0 children after deletion; got %d", len(child.Elements))
	}
}

func TestDeleteElement_DeletesImmediateChild(t *testing.T) {
	root := makeEl("root", makeEl("child"))

	ok := element.Delete(&root, "child")
	if !ok {
		t.Fatalf("expected DeleteElement to return true")
	}
	if len(root.Elements) != 0 {
		t.Fatalf("expected child to be removed; got %v", idsFromElements(root.Elements))
	}
}

func TestSwapElements_SwapsImmediateChildren(t *testing.T) {
	root := makeEl("root", makeEl("a"), makeEl("b"), makeEl("c"))

	ok := element.Swap(&root, "a", "c")
	if !ok {
		t.Fatalf("expected Swap to return true")
	}

	ids := idsFromElements(root.Elements)
	expected := []string{"c", "b", "a"}
	if !reflect.DeepEqual(ids, expected) {
		t.Fatalf("unexpected order: got %v, want %v", ids, expected)
	}
}

func TestSwapElements_SwapsNestedChildren(t *testing.T) {
	root := makeEl("root", makeEl("parent", makeEl("a"), makeEl("b"), makeEl("x")))

	ok := element.Swap(&root, "a", "b")
	if !ok {
		t.Fatalf("expected Swap to return true")
	}

	p := findElementByID([]entity.Element{root}, "parent")
	if p == nil {
		t.Fatalf("expected parent to exist")
	}
	ids := idsFromElements(p.Elements)
	expected := []string{"b", "a", "x"}
	if !reflect.DeepEqual(ids, expected) {
		t.Fatalf("unexpected nested order: got %v want %v", ids, expected)
	}
}
