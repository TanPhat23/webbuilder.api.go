package services

import (
	"context"
	"encoding/json"
	"log"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/pkg/utils"
	"my-go-app/proto"
)

type ElementService struct {
	proto.UnimplementedElementServiceServer
	snapshotRepo repositories.SnapshotRepositoryInterface
	elementRepo  repositories.ElementRepositoryInterface
}

func NewElementService(snapshotRepo repositories.SnapshotRepositoryInterface, elementRepo repositories.ElementRepositoryInterface) *ElementService {
	return &ElementService{
		snapshotRepo: snapshotRepo,
		elementRepo:  elementRepo,
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
