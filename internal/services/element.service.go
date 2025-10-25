package services

import (
	"context"
	"encoding/json"
	"log"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/pkg/utils"
	"my-go-app/proto/element"
)

type ElementService struct {
	element.UnimplementedElementSeriviceServer
	snapshotRepo repositories.SnapshotRepositoryInterface
	elementRepo  repositories.ElementRepositoryInterface
}

func NewElementService(snapshotRepo repositories.SnapshotRepositoryInterface, elementRepo repositories.ElementRepositoryInterface) *ElementService {
	return &ElementService{
		snapshotRepo: snapshotRepo,
		elementRepo:  elementRepo,
	}
}

func (s *ElementService) SaveSnapshot(ctx context.Context, req *element.SaveSnapshotRequest) (*element.SaveSnapshotResponse, error) {
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

	return &element.SaveSnapshotResponse{Message: "Snapshot saved successfully"}, nil
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
