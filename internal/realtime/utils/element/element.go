package element

import (
	"fmt"
	"sort"

	"my-go-app/internal/realtime/domain/entity"
)

const (
	MaxElementDepth = 32
	MaxElementNodes = 2000
	MaxMergeDepth   = 16
)

func ParentsEqual(p1, p2 *string) bool {
	if p1 == nil && p2 == nil {
		return true
	}
	if p1 == nil || p2 == nil {
		return false
	}
	return *p1 == *p2
}

func HasParent(p *string, id string) bool {
	if p == nil {
		return id == ""
	}
	return *p == id
}

// Flatten flattens nested Elements for database operations, bounded to MaxElementDepth and MaxElementNodes.
func Flatten(elements []entity.Element) []entity.Element {
	result := make([]entity.Element, 0, len(elements))
	flatten(elements, &result, 1)
	return result
}

func flatten(elements []entity.Element, result *[]entity.Element, depth int) {
	if len(elements) == 0 || depth > MaxElementDepth || len(*result) >= MaxElementNodes {
		return
	}
	for _, el := range elements {
		if len(*result) >= MaxElementNodes {
			return
		}
		*result = append(*result, el)
		if len(el.Elements) > 0 {
			flatten(el.Elements, result, depth+1)
		}
	}
}

// BuildTree converts a flat list of elements into a hierarchical tree based on ParentID.
func BuildTree(elements []entity.Element) []entity.Element {
	if len(elements) == 0 {
		return nil
	}
	elementMap := indexFlatElements(elements)
	childrenMap, roots := groupElementsByParent(elements, elementMap)
	for pid, cids := range childrenMap {
		sortElementIDs(cids, elementMap)
		childrenMap[pid] = cids
	}
	sortElementIDs(roots, elementMap)
	return assembleElementForest(roots, elementMap, childrenMap)
}

func indexFlatElements(elements []entity.Element) map[string]entity.Element {
	elementMap := make(map[string]entity.Element, len(elements))
	for _, el := range elements {
		el.Elements = nil
		elementMap[el.ID] = el
	}
	return elementMap
}

func groupElementsByParent(elements []entity.Element, elementMap map[string]entity.Element) (map[string][]string, []string) {
	childrenMap := make(map[string][]string)
	roots := make([]string, 0)
	for _, el := range elements {
		if el.ParentID == nil || *el.ParentID == "" {
			roots = append(roots, el.ID)
			continue
		}
		pID := *el.ParentID
		if _, ok := elementMap[pID]; ok {
			childrenMap[pID] = append(childrenMap[pID], el.ID)
		} else {
			roots = append(roots, el.ID)
		}
	}
	return childrenMap, roots
}

func assembleElementForest(roots []string, elementMap map[string]entity.Element, childrenMap map[string][]string) []entity.Element {
	visiting := make(map[string]bool)
	nodes := 0
	var build func(id string, depth int) entity.Element
	build = func(id string, depth int) entity.Element {
		node := elementMap[id]
		nodes++
		if depth >= MaxElementDepth || nodes >= MaxElementNodes || visiting[id] {
			return node
		}
		visiting[id] = true
		for _, cid := range childrenMap[id] {
			if nodes >= MaxElementNodes {
				break
			}
			node.Elements = append(node.Elements, build(cid, depth+1))
		}
		delete(visiting, id)
		return node
	}
	result := make([]entity.Element, 0, len(roots))
	for _, rid := range roots {
		if nodes >= MaxElementNodes {
			break
		}
		result = append(result, build(rid, 1))
	}
	return result
}

func sortElementIDs(ids []string, elementMap map[string]entity.Element) {
	keyed := true
	for _, id := range ids {
		if elementMap[id].OrderKey == "" {
			keyed = false
			break
		}
	}
	sort.SliceStable(ids, func(i, j int) bool {
		a := elementMap[ids[i]]
		b := elementMap[ids[j]]
		if keyed {
			if a.OrderKey != b.OrderKey {
				return a.OrderKey < b.OrderKey
			}
			return a.ID < b.ID
		}
		if a.Order != b.Order {
			return a.Order < b.Order
		}
		return a.ID < b.ID
	})
}

func Update(root *entity.Element, updated entity.Element) bool {
	return updateElementAtDepth(root, updated, 1)
}

func updateElementAtDepth(root *entity.Element, updated entity.Element, depth int) bool {
	if root == nil || depth > MaxElementDepth {
		return false
	}
	if root.ID == updated.ID {
		*root = updated
		return true
	}
	for i := range root.Elements {
		if updateElementAtDepth(&root.Elements[i], updated, depth+1) {
			return true
		}
	}
	return false
}

func Delete(root *entity.Element, id string) bool {
	return deleteElementAtDepth(root, id, 1)
}

func deleteElementAtDepth(root *entity.Element, id string, depth int) bool {
	if root == nil || depth > MaxElementDepth {
		return false
	}
	for i := 0; i < len(root.Elements); i++ {
		if root.Elements[i].ID == id {
			root.Elements = append(root.Elements[:i], root.Elements[i+1:]...)
			return true
		}
	}
	for i := range root.Elements {
		if deleteElementAtDepth(&root.Elements[i], id, depth+1) {
			return true
		}
	}
	return false
}

func Swap(root *entity.Element, id1, id2 string) bool {
	return swapElementsAtDepth(root, id1, id2, 1)
}

func swapElementsAtDepth(root *entity.Element, id1, id2 string, depth int) bool {
	if root == nil || depth > MaxElementDepth {
		return false
	}
	idx1, idx2 := findSwapIndices(root.Elements, id1, id2)
	if idx1 != -1 && idx2 != -1 {
		root.Elements[idx1], root.Elements[idx2] = root.Elements[idx2], root.Elements[idx1]
		return true
	}
	for i := range root.Elements {
		if swapElementsAtDepth(&root.Elements[i], id1, id2, depth+1) {
			return true
		}
	}
	return false
}

func findSwapIndices(children []entity.Element, id1, id2 string) (int, int) {
	idx1, idx2 := -1, -1
	for i := range children {
		switch children[i].ID {
		case id1:
			idx1 = i
		case id2:
			idx2 = i
		}
	}
	return idx1, idx2
}

func InsertChild(root *entity.Element, newElement entity.Element) bool {
	return createElementAtDepth(root, newElement, 1)
}

func createElementAtDepth(root *entity.Element, newElement entity.Element, depth int) bool {
	if root == nil || depth >= MaxElementDepth {
		return false
	}
	if matchesParent(root.ID, newElement.ParentID) {
		newElement.Order = len(root.Elements)
		root.Elements = append(root.Elements, newElement)
		return true
	}
	for i := range root.Elements {
		if createElementAtDepth(&root.Elements[i], newElement, depth+1) {
			return true
		}
	}
	return false
}

func matchesParent(rootID string, parentID *string) bool {
	if parentID == nil {
		return rootID == ""
	}
	return *parentID == rootID
}

// FindByID returns a pointer to the element with the given id anywhere in the tree, or nil.
// The pointer aliases the tree, so mutating through it edits the tree in place.
func FindByID(elements []entity.Element, id string) *entity.Element {
	return findByIDAtDepth(elements, id, 1)
}

func findByIDAtDepth(elements []entity.Element, id string, depth int) *entity.Element {
	if depth > MaxElementDepth {
		return nil
	}
	for i := range elements {
		if elements[i].ID == id {
			return &elements[i]
		}
		if found := findByIDAtDepth(elements[i].Elements, id, depth+1); found != nil {
			return found
		}
	}
	return nil
}

// MergeMap deep-merges patch into dst and returns the result without mutating either input.
// A nil value in patch deletes the key; nested maps are merged recursively up to MaxMergeDepth;
// any other value (including arrays) replaces the existing one.
func MergeMap(dst, patch map[string]any) map[string]any {
	return mergeMapDepth(dst, patch, 0)
}

func mergeMapDepth(dst, patch map[string]any, depth int) map[string]any {
	out := make(map[string]any, len(dst)+len(patch))
	for k, v := range dst {
		out[k] = v
	}
	for k, v := range patch {
		if v == nil {
			delete(out, k)
			continue
		}
		if pm, ok := v.(map[string]any); ok {
			if depth >= MaxMergeDepth {
				out[k] = pm
				continue
			}
			if dm, ok := out[k].(map[string]any); ok {
				out[k] = mergeMapDepth(dm, pm, depth+1)
				continue
			}
			// Replacing a non-map (or missing) value: still drop nil entries inside the new map.
			out[k] = mergeMapDepth(nil, pm, depth+1)
			continue
		}
		out[k] = v
	}
	return out
}

// ApplyProperties applies an element-update `properties` map to the element's own fields.
// Supported keys: name, content, href, src, tailwindStyles (string or null) and isLocked (bool).
// Unknown keys or wrongly typed values are rejected; properties are never stored in Settings.
func ApplyProperties(el *entity.Element, props map[string]any) error {
	stringTargets := map[string]**string{
		"name":           &el.Name,
		"content":        &el.Content,
		"href":           &el.Href,
		"src":            &el.Src,
		"tailwindStyles": &el.TailwindStyles,
	}
	for k, v := range props {
		if dst, ok := stringTargets[k]; ok {
			if v == nil {
				*dst = nil
				continue
			}
			str, ok := v.(string)
			if !ok {
				return fmt.Errorf("property %q must be a string or null", k)
			}
			*dst = &str
			continue
		}
		if k == "isLocked" {
			b, ok := v.(bool)
			if !ok {
				return fmt.Errorf("property %q must be a boolean", k)
			}
			el.IsLocked = b
			continue
		}
		return fmt.Errorf("unsupported property %q", k)
	}
	return nil
}
