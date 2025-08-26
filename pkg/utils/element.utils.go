package utils

import (
	"my-go-app/internal/models"
)

func BuildElementTree(elements []models.EditorElement) []models.EditorElement {
	elementMap := make(map[string]models.EditorElement, len(elements))
	childrenMap := make(map[string][]models.EditorElement)
	rootElements := make([]models.EditorElement, 0, len(elements))

	// Build maps and identify roots
	for _, element := range elements {
		baseElement := element.GetElement()
		elementMap[baseElement.Id] = element

		if baseElement.ParentId == nil {
			rootElements = append(rootElements, element)
		} else {
			parentID := *baseElement.ParentId
			childrenMap[parentID] = append(childrenMap[parentID], element)
		}
	}

	builtRootElements := make([]models.EditorElement, len(rootElements))
	for i, rootElement := range rootElements {
		builtRootElements[i] = buildElementWithChildren(rootElement, childrenMap)
	}
	return builtRootElements
}

func buildElementWithChildren(element models.EditorElement, childrenMap map[string][]models.EditorElement) models.EditorElement {
	baseElement := element.GetElement()
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
