package element

import (
	"fmt"

	"my-go-app/internal/realtime/domain/entity"
	"roci.dev/fracdex"
)

// NormalizeOrder repairs missing, malformed, duplicate or out-of-order keys without
// changing the existing sibling array order. It also maintains the legacy index.
func NormalizeOrder(elements []entity.Element) error {
	previous := ""
	valid := true
	for i := range elements {
		key := elements[i].OrderKey
		if key == "" || (i > 0 && key <= previous) {
			valid = false
		}
		if key != "" {
			if _, err := fracdex.KeyBetween(key, ""); err != nil {
				valid = false
			}
		}
		previous = key
	}
	if !valid {
		keys, err := fracdex.NKeysBetween("", "", uint(len(elements)))
		if err != nil {
			return fmt.Errorf("normalize sibling keys: %w", err)
		}
		for i := range elements {
			elements[i].OrderKey = keys[i]
		}
	}
	for i := range elements {
		elements[i].Order = i
		if err := NormalizeOrder(elements[i].Elements); err != nil {
			return err
		}
	}
	return nil
}

// InsertAt allocates a key between the current neighbors after legacy normalization.
func InsertAt(elements []entity.Element, element entity.Element, position int) ([]entity.Element, error) {
	if position < 0 || position > len(elements) {
		return nil, fmt.Errorf("position out of range")
	}
	left, right := "", ""
	if position > 0 {
		left = elements[position-1].OrderKey
	}
	if position < len(elements) {
		right = elements[position].OrderKey
	}
	key, err := fracdex.KeyBetween(left, right)
	if err != nil {
		return nil, err
	}
	element.OrderKey = key
	elements = append(elements, entity.Element{})
	copy(elements[position+1:], elements[position:])
	elements[position] = element
	for i := range elements {
		elements[i].Order = i
	}
	return elements, nil
}

func Children(elements *[]entity.Element, parentID string) *[]entity.Element {
	if parentID == "" {
		return elements
	}
	for i := range *elements {
		if (*elements)[i].ID == parentID {
			return &(*elements)[i].Elements
		}
		if found := ChildrenNested(&(*elements)[i].Elements, parentID); found != nil {
			return found
		}
	}
	return nil
}

func ChildrenNested(elements *[]entity.Element, parentID string) *[]entity.Element {
	for i := range *elements {
		if (*elements)[i].ID == parentID {
			return &(*elements)[i].Elements
		}
		if found := ChildrenNested(&(*elements)[i].Elements, parentID); found != nil {
			return found
		}
	}
	return nil
}

func Detach(elements *[]entity.Element, id string) *entity.Element {
	for i := range *elements {
		if (*elements)[i].ID == id {
			result := (*elements)[i]
			*elements = append((*elements)[:i], (*elements)[i+1:]...)
			for j := range *elements {
				(*elements)[j].Order = j
			}
			return &result
		}
		if found := Detach(&(*elements)[i].Elements, id); found != nil {
			return found
		}
	}
	return nil
}
