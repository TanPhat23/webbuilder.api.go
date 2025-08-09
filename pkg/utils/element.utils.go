package utils

import (
	"my-go-app/internal/models"
	"sync"
)

func BuildElementTree(elements []models.EditorElement) []models.EditorElement {
	elementMap := make(map[string]models.EditorElement, len(elements))
	childrenMap := make(map[string][]models.EditorElement)
	rootElements := make([]models.EditorElement, 0, len(elements))

	// Single pass to build maps and identify roots
	for _, element := range elements {
		baseElement := element.GetElement()
		elementMap[baseElement.ID] = element

		if baseElement.ParentID == nil {
			rootElements = append(rootElements, element)
		} else {
			parentID := *baseElement.ParentID
			childrenMap[parentID] = append(childrenMap[parentID], element)
		}
	}

	// Process roots concurrently with proper synchronization
	builtRootElements := make([]models.EditorElement, len(rootElements))
	var wg sync.WaitGroup

	for i, rootElement := range rootElements {
		wg.Add(1)
		go func(index int, element models.EditorElement) {
			defer wg.Done()
			builtRootElements[index] = buildElementWithChildren(element, childrenMap)
		}(i, rootElement)
	}

	wg.Wait()
	return builtRootElements
}

// buildElementWithChildren recursively builds an element with its nested children
func buildElementWithChildren(element models.EditorElement, childrenMap map[string][]models.EditorElement) models.EditorElement {
	baseElement := element.GetElement()
	children := childrenMap[baseElement.ID]

	if len(children) == 0 {
		return element
	}

	// Recursively build children
	builtChildren := make([]models.EditorElement, 0, len(children))
	var wg sync.WaitGroup
	for _, child := range children {
		wg.Add(1)
		builtChild := buildElementWithChildren(child, childrenMap)
		builtChildren = append(builtChildren, builtChild)
		wg.Done()
	}
	wg.Wait()

	// Create appropriate element type with children based on element type
	switch baseElement.Type {
	case "Frame":
		frameElement := &models.FrameElement{Element: baseElement}
		frameElement.Elements = make([]interface{}, len(builtChildren))
		for i, child := range builtChildren {
			frameElement.Elements[i] = child
		}
		return frameElement
	case "Section":
		sectionElement := &models.SectionElement{Element: baseElement}
		sectionElement.Elements = make([]interface{}, len(builtChildren))
		for i, child := range builtChildren {
			sectionElement.Elements[i] = child
		}
        return sectionElement
	case "Carousel":
		carouselElement := &models.CarouselElement{Element: baseElement}
		carouselElement.Elements = make([]interface{}, len(builtChildren))
		for i, child := range builtChildren {
			carouselElement.Elements[i] = child
		}
		return carouselElement

	case "Form":
		formElement := &models.FormElement{Element: baseElement}
		formElement.Elements = make([]interface{}, len(builtChildren))
		for i, child := range builtChildren {
			formElement.Elements[i] = child
		}
		return formElement

	case "List":
		listElement := &models.ListElement{Element: baseElement}
		listElement.Elements = make([]interface{}, len(builtChildren))
		for i, child := range builtChildren {
			listElement.Elements[i] = child
		}
		return listElement

	default:
		// For elements that don't have children containers, just return the element
		return element
	}
}

func ApplyElementSetting(element *models.Element, settings map[string]interface{}) models.EditorElement {
    if settings == nil {
        return element
    }

    switch element.Type {
    case "Carousel":
        return &models.CarouselElement{Element: *element, Settings: settings}
    case "Input":
        return &models.InputElement{Element: *element, Settings: settings}
    case "Select":
        return &models.SelectElement{Element: *element, Settings: settings}
    case "Form":
        return &models.FormElement{Element: *element, Settings: settings}
    case "Frame":
        return &models.FrameElement{Element: *element}
    case "List":
        return &models.ListElement{Element: *element}
    default:
        return element
    }
}

