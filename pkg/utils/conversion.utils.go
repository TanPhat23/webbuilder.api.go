package utils

import (
	"encoding/json"
	"my-go-app/internal/models"
	"my-go-app/proto"
)

// ConvertElementToProto converts a Go Element model to protobuf Element message
func ConvertElementToProto(elem *models.Element) *proto.Element {
	if elem == nil {
		return nil
	}

	protoElem := &proto.Element{
		Id:        elem.Id,
		Type:      elem.Type,
		ProjectId: elem.ProjectId,
		Order:     int32(elem.Order),
	}

	// Handle optional string fields
	if elem.Content != nil {
		protoElem.Content = elem.Content
	}
	if elem.Name != nil {
		protoElem.Name = elem.Name
	}
	if elem.TailwindStyles != nil {
		protoElem.TailwindStyles = elem.TailwindStyles
	}
	if elem.Src != nil {
		protoElem.Src = elem.Src
	}
	if elem.Href != nil {
		protoElem.Href = elem.Href
	}
	if elem.ParentId != nil {
		protoElem.ParentId = elem.ParentId
	}
	if elem.PageId != nil {
		protoElem.PageId = elem.PageId
	}

	// Convert json.RawMessage fields to strings
	protoElem.Styles = string(elem.Styles)
	if elem.Settings != nil {
		settingsStr := string(*elem.Settings)
		protoElem.Settings = &settingsStr
	}

	// Convert EventWorkflows slice to JSON string for protobuf
	if len(elem.EventWorkflows) > 0 {
		workflowsJSON, err := json.Marshal(elem.EventWorkflows)
		if err == nil {
			workflowsStr := string(workflowsJSON)
			protoElem.EventWorkflows = &workflowsStr
		}
	}

	// Convert child elements recursively
	if len(elem.Elements) > 0 {
		protoElem.Elements = make([]*proto.Element, len(elem.Elements))
		for i, child := range elem.Elements {
			protoElem.Elements[i] = ConvertElementToProto(&child)
		}
	}

	return protoElem
}

// ConvertProtoToElement converts a protobuf Element message to Go Element model
func ConvertProtoToElement(protoElem *proto.Element) *models.Element {
	if protoElem == nil {
		return nil
	}

	elem := &models.Element{
		Id:        protoElem.Id,
		Type:      protoElem.Type,
		ProjectId: protoElem.ProjectId,
		Order:     int(protoElem.Order),
	}

	// Handle optional string fields
	if protoElem.Content != nil {
		elem.Content = protoElem.Content
	}
	if protoElem.Name != nil {
		elem.Name = protoElem.Name
	}
	if protoElem.TailwindStyles != nil {
		elem.TailwindStyles = protoElem.TailwindStyles
	}
	if protoElem.Src != nil {
		elem.Src = protoElem.Src
	}
	if protoElem.Href != nil {
		elem.Href = protoElem.Href
	}
	if protoElem.ParentId != nil {
		elem.ParentId = protoElem.ParentId
	}
	if protoElem.PageId != nil {
		elem.PageId = protoElem.PageId
	}

	// Convert string fields to json.RawMessage
	elem.Styles = json.RawMessage(protoElem.Styles)
	if protoElem.Settings != nil {
		settingsRaw := json.RawMessage(*protoElem.Settings)
		elem.Settings = &settingsRaw
	}

	// Convert EventWorkflows string to ElementEventWorkflow slice
	if protoElem.EventWorkflows != nil {
		var workflows []models.ElementEventWorkflow
		err := json.Unmarshal([]byte(*protoElem.EventWorkflows), &workflows)
		if err == nil {
			elem.EventWorkflows = workflows
		}
	}

	// Convert child elements recursively
	if len(protoElem.Elements) > 0 {
		elem.Elements = make([]models.Element, len(protoElem.Elements))
		for i, child := range protoElem.Elements {
			if converted := ConvertProtoToElement(child); converted != nil {
				elem.Elements[i] = *converted
			}
		}
	}

	return elem
}

// ConvertElementsToProto converts a slice of Go Element models to protobuf Element messages
func ConvertElementsToProto(elements []models.Element) []*proto.Element {
	if len(elements) == 0 {
		return nil
	}

	protoElements := make([]*proto.Element, len(elements))
	for i, elem := range elements {
		protoElements[i] = ConvertElementToProto(&elem)
	}
	return protoElements
}

// ConvertProtoElementsToModel converts a slice of protobuf Element messages to Go Element models
func ConvertProtoElementsToModel(protoElements []*proto.Element) []models.Element {
	if len(protoElements) == 0 {
		return nil
	}

	elements := make([]models.Element, 0, len(protoElements))
	for _, protoElem := range protoElements {
		if elem := ConvertProtoToElement(protoElem); elem != nil {
			elements = append(elements, *elem)
		}
	}
	return elements
}

// ConvertEventWorkflowsToString converts ElementEventWorkflow slice to string pointer for protobuf
func ConvertEventWorkflowsToString(eventWorkflows []models.ElementEventWorkflow) *string {
	if len(eventWorkflows) == 0 {
		return nil
	}
	workflowsJSON, err := json.Marshal(eventWorkflows)
	if err != nil {
		return nil
	}
	str := string(workflowsJSON)
	return &str
}

// ConvertStringToEventWorkflows converts string pointer from protobuf to ElementEventWorkflow slice
func ConvertStringToEventWorkflows(workflowsStr *string) []models.ElementEventWorkflow {
	if workflowsStr == nil || *workflowsStr == "" {
		return []models.ElementEventWorkflow{}
	}
	var workflows []models.ElementEventWorkflow
	err := json.Unmarshal([]byte(*workflowsStr), &workflows)
	if err != nil {
		return []models.ElementEventWorkflow{}
	}
	return workflows
}
