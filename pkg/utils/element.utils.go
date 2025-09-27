package utils

import (
	"encoding/json"
	"log"
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

	// Debug: log counts to aid diagnosing missing children
	totalElements := len(elements)
	totalRoots := len(rootElements)
	totalParentKeys := len(childrenMap)
	totalChildEntries := 0
	for _, arr := range childrenMap {
		totalChildEntries += len(arr)
	}
	log.Printf("BuildElementTree: totalElements=%d totalRoots=%d parentKeys=%d totalChildEntries=%d", totalElements, totalRoots, totalParentKeys, totalChildEntries)

	// Use fan-in/fan-out pattern for concurrent subtree building
	return buildElementTreeConcurrent(rootElements, childrenMap)
}

// buildElementTreeConcurrent uses fan-in/fan-out pattern to build element trees concurrently
func buildElementTreeConcurrent(rootElements []models.EditorElement, childrenMap map[string][]models.EditorElement) []models.EditorElement {
	if len(rootElements) == 0 {
		return rootElements
	}

	// Channel for results (fan-in)
	results := make(chan elementResult, len(rootElements))

	// Fan-out: start goroutines for each root element
	for i, rootElement := range rootElements {
		go func(index int, element models.EditorElement) {
			builtElement := buildElementWithChildren(element, childrenMap)
			results <- elementResult{index: index, element: builtElement}
		}(i, rootElement)
	}

	// Fan-in: collect results in correct order
	builtRootElements := make([]models.EditorElement, len(rootElements))
	for i := 0; i < len(rootElements); i++ {
		result := <-results
		builtRootElements[result.index] = result.element
	}

	close(results)
	return builtRootElements
}

// elementResult holds the result of building an element tree with its original index
type elementResult struct {
	index   int
	element models.EditorElement
}

func buildElementWithChildren(element models.EditorElement, childrenMap map[string][]models.EditorElement) models.EditorElement {
	baseElement := element.GetElement()
	// Guard against nil base element - return as-is to avoid panics.
	if baseElement == nil {
		return element
	}
	children := childrenMap[baseElement.Id]

	builtChildren := make([]models.EditorElement, len(children))
	for i, child := range children {
		builtChildren[i] = buildElementWithChildren(child, childrenMap)
	}

	switch baseElement.Type {
	case "Frame":
		frameElement := &models.FrameElement{Element: baseElement}
		frameElement.Elements = make([]any, len(builtChildren))
		for i, child := range builtChildren {
			frameElement.Elements[i] = child
		}
		return frameElement
	case "Section":
		sectionElement := &models.SectionElement{Element: baseElement}
		sectionElement.Elements = make([]any, len(builtChildren))
		for i, child := range builtChildren {
			sectionElement.Elements[i] = child
		}
		return sectionElement
	case "Carousel":
		carouselElement := &models.CarouselElement{Element: baseElement}
		carouselElement.Elements = make([]any, len(builtChildren))
		for i, child := range builtChildren {
			carouselElement.Elements[i] = child
		}
		return carouselElement
	case "Form":
		formElement := &models.FormElement{Element: baseElement}
		formElement.Elements = make([]any, len(builtChildren))
		for i, child := range builtChildren {
			formElement.Elements[i] = child
		}
		return formElement
	case "List":
		listElement := &models.ListElement{Element: baseElement}
		listElement.Elements = make([]any, len(builtChildren))
		for i, child := range builtChildren {
			listElement.Elements[i] = child
		}
		return listElement
	case "Select":
		selectElement := &models.SelectElement{Element: baseElement}
		selectElement.Elements = make([]any, len(builtChildren))
		for i, child := range builtChildren {
			selectElement.Elements[i] = child
		}
		return selectElement
	default:
		return element
	}
}

func GetChildrenFromEditorElement(element models.EditorElement) []any {
	if element == nil {
		return nil
	}
	switch e := element.(type) {
	case *models.FrameElement:
		if e == nil {
			return nil
		}
		return e.Elements
	case models.FrameElement:
		return e.Elements
	case *models.SectionElement:
		if e == nil {
			return nil
		}
		return e.Elements
	case models.SectionElement:
		return e.Elements
	case *models.CarouselElement:
		if e == nil {
			return nil
		}
		return e.Elements
	case models.CarouselElement:
		return e.Elements
	case *models.FormElement:
		if e == nil {
			return nil
		}
		return e.Elements
	case models.FormElement:
		return e.Elements
	case *models.ListElement:
		if e == nil {
			return nil
		}
		return e.Elements
	case models.ListElement:
		return e.Elements
	case *models.SelectElement:
		if e == nil {
			return nil
		}
		return e.Elements
	case models.SelectElement:
		return e.Elements
	default:
		return nil
	}
}

func ConvertToEditorElement(v any) (models.EditorElement, error) {
	if ee, ok := v.(models.EditorElement); ok {
		return ee, nil
	}

	unmarshalBase := func(b []byte) (models.Element, error) {
		var base models.Element
		if err := json.Unmarshal(b, &base); err != nil {
			return models.Element{}, err
		}
		return base, nil
	}

	constructors := map[string]func([]byte) (models.EditorElement, error){
		"Frame": func(b []byte) (models.EditorElement, error) {
			var fe models.FrameElement
			if err := json.Unmarshal(b, &fe); err != nil {
				return nil, err
			}
			return &fe, nil
		},
		"Section": func(b []byte) (models.EditorElement, error) {
			var se models.SectionElement
			if err := json.Unmarshal(b, &se); err != nil {
				return nil, err
			}
			return &se, nil
		},
		"Carousel": func(b []byte) (models.EditorElement, error) {
			var ce models.CarouselElement
			if err := json.Unmarshal(b, &ce); err != nil {
				return nil, err
			}
			return &ce, nil
		},
		"Form": func(b []byte) (models.EditorElement, error) {
			var fo models.FormElement
			if err := json.Unmarshal(b, &fo); err != nil {
				return nil, err
			}
			return &fo, nil
		},
		"List": func(b []byte) (models.EditorElement, error) {
			var le models.ListElement
			if err := json.Unmarshal(b, &le); err != nil {
				return nil, err
			}
			return &le, nil
		},
		"Select": func(b []byte) (models.EditorElement, error) {
			var se models.SelectElement
			if err := json.Unmarshal(b, &se); err != nil {
				return nil, err
			}
			return &se, nil
		},
		"Button": func(b []byte) (models.EditorElement, error) {
			var be models.ButtonElement
			if err := json.Unmarshal(b, &be); err != nil {
				return nil, err
			}
			return &be, nil
		},
		"Input": func(b []byte) (models.EditorElement, error) {
			var ie models.InputElement
			if err := json.Unmarshal(b, &ie); err != nil {
				return nil, err
			}
			return &ie, nil
		},
	}

	var raw []byte
	var err error
	switch t := v.(type) {
	case map[string]any:
		raw, err = json.Marshal(t)
		if err != nil {
			return nil, err
		}
	default:
		raw, err = json.Marshal(t)
		if err != nil {
			return nil, err
		}
	}

	base, err := unmarshalBase(raw)
	if err != nil {
		return nil, err
	}

	if ctor, ok := constructors[base.Type]; ok {
		return ctor(raw)
	}

	return &base, nil
}
