package usecase

import (
	"context"
	"fmt"

	domainEntity "my-go-app/internal/realtime/domain/entity"
	domainErrors "my-go-app/internal/realtime/domain/errors"
	domainRepo "my-go-app/internal/realtime/domain/repository"
	domainUsecase "my-go-app/internal/realtime/domain/usecase"
	elementutils "my-go-app/internal/realtime/utils/element"
	"my-go-app/internal/realtime/utils/validation"

	"github.com/google/uuid"
)

type elementUseCase struct {
	repo           *domainRepo.Repository
	docUseCase     domainUsecase.DocumentUseCase
	persistUseCase domainUsecase.PersistUseCase
}

// NewElementUseCase creates a new instance of elementUseCase with persistence integration.
func NewElementUseCase(
	repo *domainRepo.Repository,
	docUseCase domainUsecase.DocumentUseCase,
	persistUseCase domainUsecase.PersistUseCase,
) domainUsecase.ElementUseCase {
	return &elementUseCase{
		repo:           repo,
		docUseCase:     docUseCase,
		persistUseCase: persistUseCase,
	}
}

// SyncPage ensures the cache is up to date for a specific page.
func (uc *elementUseCase) SyncPage(ctx context.Context, projectID string, pageID string) ([]domainEntity.Element, error) {
	if err := uc.docUseCase.EnsureDocumentCached(ctx, projectID, pageID); err != nil {
		return nil, fmt.Errorf("failed to refresh cache: %w", err)

	}
	return uc.docUseCase.GetDocument(ctx, projectID, pageID)
}

// SyncPageWithSeq is SyncPage plus the page sequence number belonging to the returned elements.
func (uc *elementUseCase) SyncPageWithSeq(ctx context.Context, projectID string, pageID string) ([]domainEntity.Element, int64, error) {
	if err := uc.docUseCase.EnsureDocumentCached(ctx, projectID, pageID); err != nil {
		return nil, 0, fmt.Errorf("failed to refresh cache: %w", err)
	}
	elements, seq, err := uc.docUseCase.GetDocumentWithSeq(ctx, projectID, pageID)
	if err != nil {
		return nil, 0, err
	}
	if err := elementutils.NormalizeOrder(elements); err != nil {
		return nil, 0, err
	}
	return elements, seq, nil
}

func (uc *elementUseCase) ListElementsByPage(ctx context.Context, projectID string, pageID string) ([]domainEntity.Element, error) {
	return uc.docUseCase.GetDocument(ctx, projectID, pageID)
}

func (uc *elementUseCase) ListElementsByProject(ctx context.Context, projectID string) ([]domainEntity.Element, error) {
	flattenElements, err := uc.repo.ElementRepo.ListElementsByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return elementutils.BuildTree(flattenElements), nil
}

func (uc *elementUseCase) CreateElement(ctx context.Context, projectID string, element *domainEntity.Element) (int64, error) {
	return uc.CreateElementAt(ctx, projectID, element, nil)
}

func (uc *elementUseCase) CreateElementAt(ctx context.Context, projectID string, element *domainEntity.Element, position *int) (int64, error) {
	if err := validation.ElementForCreate(element); err != nil {
		return 0, domainErrors.Wrap(domainErrors.CodeValidation, err)
	}
	if element.ID == "" {
		element.ID = uuid.NewString()
	}
	// The closure may be retried by Redis; never use output of a previous attempt as input.
	input := *element
	return uc.docUseCase.UpdateDocument(ctx, projectID, input.PageID, func(elements []domainEntity.Element) ([]domainEntity.Element, error) {
		updatedTree, created, err := insertElementIntoTree(elements, input, position)
		if err != nil {
			return nil, err
		}
		*element = created
		return updatedTree, nil
	})
}

func insertElementIntoTree(elements []domainEntity.Element, input domainEntity.Element, position *int) ([]domainEntity.Element, domainEntity.Element, error) {
	if err := elementutils.NormalizeOrder(elements); err != nil {
		return nil, domainEntity.Element{}, err
	}
	if elementutils.FindByID(elements, input.ID) != nil {
		return nil, domainEntity.Element{}, domainErrors.New(domainErrors.CodeValidation, "element already exists", nil)
	}
	parentID := ""
	if input.ParentID != nil {
		parentID = *input.ParentID
	}
	elements, children, err := ensureTargetChildren(&elements, parentID)
	if err != nil {
		return nil, domainEntity.Element{}, err
	}
	index, err := resolveInsertIndex(len(*children), position)
	if err != nil {
		return nil, domainEntity.Element{}, err
	}
	created := input
	if parentID == "" {
		created.ParentID = nil
	}
	if err := elementutils.NormalizeOrder(created.Elements); err != nil {
		return nil, domainEntity.Element{}, err
	}
	updated, err := elementutils.InsertAt(*children, created, index)
	if err != nil {
		return nil, domainEntity.Element{}, err
	}
	*children = updated
	return elements, updated[index], nil
}

func ensureTargetChildren(elements *[]domainEntity.Element, parentID string) ([]domainEntity.Element, *[]domainEntity.Element, error) {
	children := elementutils.FindChildren(elements, parentID)
	if children == nil && parentID == "root" {
		extended, err := elementutils.InsertAt(*elements, domainEntity.Element{ID: "root"}, len(*elements))
		if err != nil {
			return nil, nil, err
		}
		*elements = extended
		children = &(*elements)[len(*elements)-1].Elements
	}
	if children == nil {
		return nil, nil, domainErrors.New(domainErrors.CodeNotFound, "parent not found", nil)
	}
	return *elements, children, nil
}

func resolveInsertIndex(count int, position *int) (int, error) {
	index := count
	if position != nil {
		index = *position
	}
	if index < 0 || index > count {
		return 0, domainErrors.New(domainErrors.CodeValidation, "position out of range", nil)
	}
	return index, nil
}

func (uc *elementUseCase) UpdateElement(ctx context.Context, projectID string, pageID string, element *domainEntity.Element) (int64, error) {
	if element == nil || element.ID == "" {
		return 0, domainErrors.New(domainErrors.CodeValidation, "element id is required", nil)
	}

	return uc.docUseCase.UpdateDocument(ctx, projectID, pageID, func(elements []domainEntity.Element) ([]domainEntity.Element, error) {
		for i := range elements {
			if elementutils.Update(&elements[i], *element) {
				return elements, nil
			}
		}
		return nil, domainErrors.New(domainErrors.CodeNotFound, "element not found in cache", nil)
	})
}

// PatchElement merges the given changes into the element in place:
//   - styles and settings are deep-merged into the existing maps (a null value deletes the key);
//   - properties are applied to the element's own fields (see elementutils.ApplyProperties);
//     they are never written into Settings.
func (uc *elementUseCase) PatchElement(ctx context.Context, projectID string, pageID string, elementID string, properties map[string]any, styles map[string]any, settings map[string]any) (int64, error) {
	return uc.docUseCase.MutateElement(ctx, projectID, pageID, elementID, func(target *domainEntity.Element) error {
		if len(properties) > 0 {
			if err := elementutils.ApplyProperties(target, properties); err != nil {
				return domainErrors.Wrap(domainErrors.CodeValidation, err)
			}
		}
		if len(styles) > 0 {
			target.Styles = elementutils.MergeMap(target.Styles, styles)
		}
		if len(settings) > 0 {
			target.Settings = elementutils.MergeMap(target.Settings, settings)
		}
		return nil
	})
}

func (uc *elementUseCase) MoveElement(ctx context.Context, projectID string, pageID string, elementID string, parentID string, index int) (*domainEntity.Element, int64, error) {
	if index < 0 {
		return nil, 0, domainErrors.New(domainErrors.CodeValidation, "position out of range", nil)
	}
	var result *domainEntity.Element
	seq, err := uc.docUseCase.UpdateDocument(ctx, projectID, pageID, func(elements []domainEntity.Element) ([]domainEntity.Element, error) {
		updated, committed, err := moveElementInTree(elements, elementID, parentID, index)
		if err != nil {
			return nil, err
		}
		result = committed
		return updated, nil
	})
	if err != nil {
		return nil, 0, err
	}
	return result, seq, nil
}

func moveElementInTree(elements []domainEntity.Element, elementID, parentID string, index int) ([]domainEntity.Element, *domainEntity.Element, error) {
	if err := elementutils.NormalizeOrder(elements); err != nil {
		return nil, nil, err
	}
	if err := validateMoveParent(elements, elementID, parentID); err != nil {
		return nil, nil, err
	}
	moved := elementutils.Detach(&elements, elementID)
	children := elementutils.FindChildren(&elements, parentID)
	if children == nil {
		return nil, nil, domainErrors.New(domainErrors.CodeNotFound, "parent not found", nil)
	}
	if index > len(*children) {
		return nil, nil, domainErrors.New(domainErrors.CodeValidation, "position out of range", nil)
	}
	if parentID == "" {
		moved.ParentID = nil
	} else {
		parent := parentID
		moved.ParentID = &parent
	}
	updated, err := elementutils.InsertAt(*children, *moved, index)
	if err != nil {
		return nil, nil, err
	}
	*children = updated
	committed := updated[index]
	return elements, &committed, nil
}

func validateMoveParent(elements []domainEntity.Element, elementID, parentID string) error {
	target := elementutils.FindByID(elements, elementID)
	if target == nil {
		return domainErrors.New(domainErrors.CodeNotFound, "element not found", nil)
	}
	if parentID == "" {
		return nil
	}
	if elementID == parentID || elementutils.FindByID(target.Elements, parentID) != nil {
		return domainErrors.New(domainErrors.CodeValidation, "invalid parent", nil)
	}
	if elementutils.FindByID(elements, parentID) == nil {
		return domainErrors.New(domainErrors.CodeNotFound, "parent not found", nil)
	}
	return nil
}

func (uc *elementUseCase) DeleteElement(ctx context.Context, projectID string, pageID string, elementID string) (int64, error) {
	return uc.docUseCase.UpdateDocument(ctx, projectID, pageID, func(elements []domainEntity.Element) ([]domainEntity.Element, error) {
		newRoots := make([]domainEntity.Element, 0, len(elements))
		removed := false
		for _, r := range elements {
			if r.ID == elementID {
				removed = true
				continue
			}
			newRoots = append(newRoots, r)
		}
		if removed {
			return newRoots, nil
		}
		for i := range elements {
			if elementutils.Delete(&elements[i], elementID) {
				return elements, nil
			}
		}
		return nil, domainErrors.New(domainErrors.CodeNotFound, "element not found", nil)
	})
}
