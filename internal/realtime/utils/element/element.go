package element

import (
	"fmt"
	"sort"

	"my-go-app/internal/realtime/domain/entity"
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

func CompareParent(p *string, id string) bool {
	if p == nil {
		return id == ""
	}
	return *p == id
}

// FlattenElements flattens nested Elements for database operations.
func FlattenElements(elements []entity.Element) []entity.Element {
	result := make([]entity.Element, 0, len(elements))
	flatten(elements, &result)
	return result
}

func flatten(elements []entity.Element, result *[]entity.Element) {
	if len(elements) == 0 {
		return
	}
	for _, el := range elements {
		*result = append(*result, el)
		if len(el.Elements) > 0 {
			flatten(el.Elements, result)
		}
	}
}

// BuildElementTree converts a flat list of elements into a hierarchical tree based on ParentID.
func BuildElementTree(elements []entity.Element) []entity.Element {
	if len(elements) == 0 {
		return nil
	}

	elementMap := make(map[string]entity.Element, len(elements))
	for _, el := range elements {
		el.Elements = nil
		elementMap[el.ID] = el
	}

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

	for pid, cids := range childrenMap {
		keyed := true
		for _, id := range cids {
			if elementMap[id].OrderKey == "" {
				keyed = false
				break
			}
		}
		sort.SliceStable(cids, func(i, j int) bool {
			a := elementMap[cids[i]]
			b := elementMap[cids[j]]
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
		childrenMap[pid] = cids
	}

	keyed := true
	for _, id := range roots {
		if elementMap[id].OrderKey == "" {
			keyed = false
			break
		}
	}
	sort.SliceStable(roots, func(i, j int) bool {
		a := elementMap[roots[i]]
		b := elementMap[roots[j]]
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

	var build func(id string) entity.Element
	build = func(id string) entity.Element {
		node := elementMap[id]
		for _, cid := range childrenMap[id] {
			node.Elements = append(node.Elements, build(cid))
		}
		return node
	}

	result := make([]entity.Element, 0, len(roots))
	for _, rid := range roots {
		result = append(result, build(rid))
	}
	return result
}

func UpdateElement(root *entity.Element, updated entity.Element) bool {
	if root == nil {
		return false
	}
	if root.ID == updated.ID {
		*root = updated
		return true
	}
	for i := range root.Elements {
		if UpdateElement(&root.Elements[i], updated) {
			return true
		}
	}
	return false
}

func DeleteElement(root *entity.Element, id string) bool {
	if root == nil {
		return false
	}
	for i := 0; i < len(root.Elements); i++ {
		if root.Elements[i].ID == id {
			root.Elements = append(root.Elements[:i], root.Elements[i+1:]...)
			return true
		}
	}
	for i := range root.Elements {
		if DeleteElement(&root.Elements[i], id) {
			return true
		}
	}
	return false
}

func SwapElements(root *entity.Element, id1, id2 string) bool {
	if root == nil {
		return false
	}
	var idx1, idx2 = -1, -1
	for i := range root.Elements {
		if root.Elements[i].ID == id1 {
			idx1 = i
		} else if root.Elements[i].ID == id2 {
			idx2 = i
		}
	}
	if idx1 != -1 && idx2 != -1 {
		root.Elements[idx1], root.Elements[idx2] = root.Elements[idx2], root.Elements[idx1]
		return true
	}

	for i := range root.Elements {
		if SwapElements(&root.Elements[i], id1, id2) {
			return true
		}
	}
	return false
}

func CreateElement(root *entity.Element, newElement entity.Element) bool {
	if root == nil {
		return false
	}
	if (newElement.ParentID == nil && root.ID == "") || (newElement.ParentID != nil && *newElement.ParentID == root.ID) {
		newElement.Order = len(root.Elements)
		root.Elements = append(root.Elements, newElement)
		return true
	}
	for i := range root.Elements {
		if CreateElement(&root.Elements[i], newElement) {
			return true
		}
	}
	return false
}

// FindByID returns a pointer to the element with the given id anywhere in the tree, or nil.
// The pointer aliases the tree, so mutating through it edits the tree in place.
func FindByID(elements []entity.Element, id string) *entity.Element {
	for i := range elements {
		if elements[i].ID == id {
			return &elements[i]
		}
		if found := FindByID(elements[i].Elements, id); found != nil {
			return found
		}
	}
	return nil
}

// MergeMap deep-merges patch into dst and returns the result without mutating either input.
// A nil value in patch deletes the key; nested maps are merged recursively; any other value
// (including arrays) replaces the existing one.
func MergeMap(dst, patch map[string]any) map[string]any {
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
			if dm, ok := out[k].(map[string]any); ok {
				out[k] = MergeMap(dm, pm)
				continue
			}
			// Replacing a non-map (or missing) value: still drop nil entries inside the new map.
			out[k] = MergeMap(nil, pm)
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
	for k, v := range props {
		switch k {
		case "name", "content", "href", "src", "tailwindStyles":
			var ptr *string
			if v != nil {
				str, ok := v.(string)
				if !ok {
					return fmt.Errorf("property %q must be a string or null", k)
				}
				ptr = &str
			}
			switch k {
			case "name":
				el.Name = ptr
			case "content":
				el.Content = ptr
			case "href":
				el.Href = ptr
			case "src":
				el.Src = ptr
			case "tailwindStyles":
				el.TailwindStyles = ptr
			}
		case "isLocked":
			b, ok := v.(bool)
			if !ok {
				return fmt.Errorf("property %q must be a boolean", k)
			}
			el.IsLocked = b
		default:
			return fmt.Errorf("unsupported property %q", k)
		}
	}
	return nil
}
