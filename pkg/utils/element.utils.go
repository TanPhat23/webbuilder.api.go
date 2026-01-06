package utils

import (
	"encoding/json"
	"my-go-app/internal/models"
)

func BuildElementTree(elements []models.EditorElement) []models.EditorElement {
	elementMap := make(map[string]models.EditorElement, len(elements))
	childrenMap := make(map[string][]models.EditorElement)
	rootElements := make([]models.EditorElement, 0, len(elements))

	for _, element := range elements {
		baseElement := element.GetElement()
		if baseElement == nil {
			continue
		}
		elementMap[baseElement.Id] = element

		if baseElement.ParentId == nil {
			rootElements = append(rootElements, element)
		} else {
			parentID := *baseElement.ParentId
			childrenMap[parentID] = append(childrenMap[parentID], element)
		}
	}

	totalChildEntries := 0
	for _, arr := range childrenMap {
		totalChildEntries += len(arr)
	}

	return buildElementTreeConcurrent(rootElements, childrenMap)
}

func buildElementTreeConcurrent(rootElements []models.EditorElement, childrenMap map[string][]models.EditorElement) []models.EditorElement {
	if len(rootElements) == 0 {
		return rootElements
	}

	results := make(chan ElementResult, len(rootElements))

	for i, rootElement := range rootElements {
		go func(index int, element models.EditorElement) {
			builtElement := buildElementWithChildren(element, childrenMap)
			results <- ElementResult{Index: index, Element: builtElement}
		}(i, rootElement)
	}

	builtRootElements := make([]models.EditorElement, len(rootElements))
	for range rootElements {
		result := <-results
		builtRootElements[result.Index] = result.Element
	}
	close(results)
	return builtRootElements
}

// ElementResult represents a built element with its index for concurrent processing
type ElementResult struct {
	Index   int
	Element models.EditorElement
}

// buildElementWithChildren builds element tree structure, attaching children to container elements
func buildElementWithChildren(element models.EditorElement, childrenMap map[string][]models.EditorElement) models.EditorElement {
	baseElement := element.GetElement()
	if baseElement == nil {
		return element
	}

	// Only process children for container element types
	if !baseElement.IsContainer() {
		return element
	}

	children := childrenMap[baseElement.Id]
	if len(children) == 0 {
		return element
	}

	// Build children recursively
	builtChildren := make([]models.EditorElement, len(children))
	for i, child := range children {
		builtChildren[i] = buildElementWithChildren(child, childrenMap)
	}

	// Attach built children to the element
	// This creates a wrapper that includes the Elements field for JSON serialization
	return WrapElementWithChildren(baseElement, builtChildren)
}

// WrapElementWithChildren wraps an element with its children for serialization
func WrapElementWithChildren(element *models.Element, children []models.EditorElement) models.EditorElement {
	return &ElementWithChildren{
		Element:  element,
		Elements: ConvertEditorElementsToAny(children),
	}
}

// ConvertEditorElementsToAny converts EditorElement slice to []any for serialization
func ConvertEditorElementsToAny(elements []models.EditorElement) []any {
	result := make([]any, len(elements))
	for i, elem := range elements {
		result[i] = elem
	}
	return result
}

// ElementWithChildren is a wrapper that adds Elements field for JSON serialization
type ElementWithChildren struct {
	*models.Element
	Elements []any `json:"elements,omitempty"`
}

func (e *ElementWithChildren) GetElement() *models.Element {
	return e.Element
}

func (e *ElementWithChildren) GetType() string {
	if e.Element == nil {
		return ""
	}
	return e.Element.Type
}

// GetChildrenFromEditorElement extracts children from an EditorElement
func GetChildrenFromEditorElement(element models.EditorElement) []any {
	if element == nil {
		return nil
	}

	// Check if it's our wrapper type with children
	if ewc, ok := element.(*ElementWithChildren); ok {
		return ewc.Elements
	}

	// Check if base element is a container
	baseElement := element.GetElement()
	if baseElement == nil || !baseElement.IsContainer() {
		return nil
	}

	// If no explicit elements field, return nil (should be populated if needed)
	return nil
}

// ConvertToEditorElement converts any value to EditorElement
func ConvertToEditorElement(v any) (models.EditorElement, error) {
	if ee, ok := v.(models.EditorElement); ok {
		return ee, nil
	}

	// Marshal to JSON bytes for consistent handling
	var raw []byte
	var err error

	switch t := v.(type) {
	case map[string]any:
		raw, err = json.Marshal(t)
	default:
		raw, err = json.Marshal(t)
	}

	if err != nil {
		return nil, err
	}

	// Unmarshal to Element to get the Type field
	var element models.Element
	if err := json.Unmarshal(raw, &element); err != nil {
		return nil, err
	}

	// If it's a container type, also try to unmarshal the Elements field
	if element.IsContainer() {
		// Try to parse elements if present
		var withElements ElementWithChildren
		if err := json.Unmarshal(raw, &withElements); err == nil && len(withElements.Elements) > 0 {
			withElements.Element = &element
			return &withElements, nil
		}
	}

	return &element, nil
}
