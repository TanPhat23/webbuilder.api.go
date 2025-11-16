package services

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/pkg/utils"
	"my-go-app/proto"
)

type ElementService struct {
	proto.UnimplementedElementServiceServer
	snapshotRepo                repositories.SnapshotRepositoryInterface
	elementRepo                 repositories.ElementRepositoryInterface
	eventWorkflowRepo           *repositories.EventWorkflowRepository
	elementEventWorkflowRepo    *repositories.ElementEventWorkflowRepository
}

func NewElementService(
	snapshotRepo repositories.SnapshotRepositoryInterface,
	elementRepo repositories.ElementRepositoryInterface,
	eventWorkflowRepo *repositories.EventWorkflowRepository,
	elementEventWorkflowRepo *repositories.ElementEventWorkflowRepository,
) *ElementService {
	return &ElementService{
		snapshotRepo:             snapshotRepo,
		elementRepo:              elementRepo,
		eventWorkflowRepo:        eventWorkflowRepo,
		elementEventWorkflowRepo: elementEventWorkflowRepo,
	}
}

func (s *ElementService) SaveSnapshot(ctx context.Context, req *proto.SaveSnapshotRequest) (*proto.SaveSnapshotResponse, error) {
	var rawElements []any
	err := json.Unmarshal([]byte(req.Elements), &rawElements)
	if err != nil {
		log.Printf("Error parsing elements JSON: %v", err)
		return nil, err
	}

	elements, err := s.convertRawElementsToEditorElements(rawElements)
	if err != nil {
		log.Printf("Error converting elements: %v", err)
		return nil, err
	}

	elementsJSON, err := json.Marshal(elements)
	if err != nil {
		log.Printf("Error marshaling elements: %v", err)
		return nil, err
	}

	snapshot := models.Snapshot{
		Id:        req.Id,
		ProjectId: req.ProjectId,
		Name:      req.Name,
		Type:      req.Type,
		Elements:  elementsJSON,
		Timestamp: req.Timestamp,
	}

	err = s.snapshotRepo.SaveSnapshot(ctx, req.ProjectId, &snapshot)
	if err != nil {
		log.Printf("Error saving snapshot: %v", err)
		return nil, err
	}

	err = s.elementRepo.ReplaceElements(ctx, req.ProjectId, elements)
	if err != nil {
		log.Printf("Error replacing elements: %v", err)
		return nil, err
	}

	return &proto.SaveSnapshotResponse{Message: "Snapshot saved successfully"}, nil
}

// GetElementEventWorkflows retrieves all event workflows linked to an element
func (s *ElementService) GetElementEventWorkflows(ctx context.Context, elementID string) ([]models.ElementEventWorkflow, error) {
	if s.elementEventWorkflowRepo == nil {
		log.Printf("Warning: ElementEventWorkflowRepository not initialized")
		return []models.ElementEventWorkflow{}, nil
	}

	workflows, err := s.elementEventWorkflowRepo.GetElementEventWorkflowsByElementID(ctx, elementID)
	if err != nil {
		log.Printf("Error retrieving event workflows for element %s: %v", elementID, err)
		return nil, err
	}

	return workflows, nil
}

// LinkElementToWorkflow creates an association between an element and a workflow
func (s *ElementService) LinkElementToWorkflow(ctx context.Context, elementID, workflowID, eventName string) (*models.ElementEventWorkflow, error) {
	if s.elementEventWorkflowRepo == nil {
		log.Printf("Error: ElementEventWorkflowRepository not initialized")
		return nil, errors.New("element event workflow repository not initialized")
	}

	eew := &models.ElementEventWorkflow{
		ElementId:  elementID,
		WorkflowId: workflowID,
		EventName:  eventName,
	}

	created, err := s.elementEventWorkflowRepo.CreateElementEventWorkflow(ctx, eew)
	if err != nil {
		log.Printf("Error linking element %s to workflow %s: %v", elementID, workflowID, err)
		return nil, err
	}

	return created, nil
}

// UnlinkElementFromWorkflow removes an association between an element and a workflow
func (s *ElementService) UnlinkElementFromWorkflow(ctx context.Context, elementID, workflowID, eventName string) error {
	if s.elementEventWorkflowRepo == nil {
		log.Printf("Error: ElementEventWorkflowRepository not initialized")
		return errors.New("element event workflow repository not initialized")
	}

	// Check if the association exists
	exists, err := s.elementEventWorkflowRepo.CheckIfWorkflowLinkedToElement(ctx, elementID, workflowID, eventName)
	if err != nil {
		log.Printf("Error checking workflow link: %v", err)
		return err
	}

	if !exists {
		log.Printf("Workflow not linked to element")
		return errors.New("workflow not linked to element")
	}

	// Delete all associations for this element-workflow-event combination
	err = s.elementEventWorkflowRepo.DeleteElementEventWorkflowsByElementID(ctx, elementID)
	if err != nil {
		log.Printf("Error unlinking element %s from workflow %s: %v", elementID, workflowID, err)
		return err
	}

	return nil
}

// GetProjectWorkflows retrieves all workflows for a project
func (s *ElementService) GetProjectWorkflows(ctx context.Context, projectID string) ([]models.EventWorkflow, error) {
	if s.eventWorkflowRepo == nil {
		log.Printf("Warning: EventWorkflowRepository not initialized")
		return []models.EventWorkflow{}, nil
	}

	workflows, err := s.eventWorkflowRepo.GetEventWorkflowsByProjectID(ctx, projectID)
	if err != nil {
		log.Printf("Error retrieving workflows for project %s: %v", projectID, err)
		return nil, err
	}

	return workflows, nil
}

// GetEnabledWorkflows retrieves all enabled workflows for a project
func (s *ElementService) GetEnabledWorkflows(ctx context.Context, projectID string) ([]models.EventWorkflow, error) {
	if s.eventWorkflowRepo == nil {
		log.Printf("Warning: EventWorkflowRepository not initialized")
		return []models.EventWorkflow{}, nil
	}

	workflows, err := s.eventWorkflowRepo.GetEnabledEventWorkflowsByProjectID(ctx, projectID)
	if err != nil {
		log.Printf("Error retrieving enabled workflows for project %s: %v", projectID, err)
		return nil, err
	}

	return workflows, nil
}

func (s *ElementService) GetWorkflowsByEvent(ctx context.Context, eventName string) ([]models.ElementEventWorkflow, error) {
	if s.elementEventWorkflowRepo == nil {
		log.Printf("Warning: ElementEventWorkflowRepository not initialized")
		return []models.ElementEventWorkflow{}, nil
	}

	workflows, err := s.elementEventWorkflowRepo.GetElementEventWorkflowsByEventName(ctx, eventName)
	if err != nil {
		log.Printf("Error retrieving workflows for event %s: %v", eventName, err)
		return nil, err
	}

	return workflows, nil
}

func (s *ElementService) convertRawElementsToEditorElements(rawElements []any) ([]models.EditorElement, error) {
	elements := make([]models.EditorElement, len(rawElements))
	for i, e := range rawElements {
		ee, err := utils.ConvertToEditorElement(e)
		if err != nil {
			return nil, err
		}
		elements[i] = ee
	}
	return elements, nil
}

func (s *ElementService) GetProjectElements(ctx context.Context, req *proto.ProjectElementsRequest) (*proto.ProjectElementsResponse, error) {
	elements, err := s.elementRepo.GetElements(ctx, req.ProjectId)
	if err != nil {
		log.Printf("Error retrieving elements for project %s: %v", req.ProjectId, err)
		return nil, err
	}

	protoElements := s.convertEditorElementsToProto(elements)

	return &proto.ProjectElementsResponse{Elements: protoElements}, nil
}

func (s *ElementService) convertEditorElementsToProto(elements []models.EditorElement) []*proto.Element {
	protoElements := make([]*proto.Element, 0, len(elements))

	for _, editorElem := range elements {
		if editorElem == nil {
			continue
		}

		elem := editorElem.GetElement()
		if elem == nil {
			continue
		}

		protoElem := &proto.Element{
			Id:        elem.Id,
			Type:      elem.Type,
			Content:   elem.Content,
			Name:      elem.Name,
			Styles:    string(elem.Styles),
			TailwindStyles: elem.TailwindStyles,
			Src:       elem.Src,
			Href:      elem.Href,
			ParentId:  elem.ParentId,
			PageId:    elem.PageId,
			ProjectId: elem.ProjectId,
			EventWorkflows: utils.ConvertEventWorkflowsToString(elem.EventWorkflows),
			Order:     int32(elem.Order),
		}

		if elem.Settings != nil {
			settingsStr := string(*elem.Settings)
			protoElem.Settings = &settingsStr
		}

		children := utils.GetChildrenFromEditorElement(editorElem)
		if len(children) > 0 {
			childElements := make([]models.EditorElement, 0, len(children))
			for _, child := range children {
				childElem, err := utils.ConvertToEditorElement(child)
				if err == nil {
					childElements = append(childElements, childElem)
				}
			}
			protoElem.Elements = s.convertEditorElementsToProto(childElements)
		}

		protoElements = append(protoElements, protoElem)
	}

	return protoElements
}
