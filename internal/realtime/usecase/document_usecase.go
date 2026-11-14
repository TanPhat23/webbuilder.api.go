package usecase

import (
	"context"
	"fmt"

	domainEntity "my-go-app/internal/realtime/domain/entity"
	domainRepo "my-go-app/internal/realtime/domain/repository"
	domainUsecase "my-go-app/internal/realtime/domain/usecase"
	elementutils "my-go-app/internal/realtime/utils/element"
)

type documentUseCase struct {
	repo *domainRepo.Repository
}

// NewDocumentUseCase creates a new instance of documentUseCase
func NewDocumentUseCase(repo *domainRepo.Repository) domainUsecase.DocumentUseCase {
	return &documentUseCase{
		repo: repo,
	}
}

func (uc *documentUseCase) GetDocument(ctx context.Context, projectID string, pageID string) ([]domainEntity.Element, error) {
	cachedElements, err := uc.repo.DocumentRepo.GetDocument(ctx, projectID, pageID)
	if err != nil {
		fmt.Printf("cache error during get: %v\n", err)
	}
	if cachedElements != nil {
		return cachedElements, nil
	}

	elements, err := uc.repo.ElementRepo.GetElementsByPageID(ctx, projectID, pageID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch page elements from db: %w", err)
	}

	// Build hierarchical tree from flat elements before caching/returning.
	tree := elementutils.BuildElementTree(elements)

	if err := uc.repo.DocumentRepo.SetDocumentIfAbsent(ctx, projectID, pageID, tree); err != nil {
		fmt.Printf("failed to update cache: %v\n", err)
	}

	return tree, nil
}
func (uc *documentUseCase) RefreshCache(ctx context.Context, projectID string, pageID string) error {
	cachedElements, err := uc.repo.DocumentRepo.GetDocument(ctx, projectID, pageID)
	if err != nil {
		return fmt.Errorf("failed to check cache during refresh: %w", err)
	}
	if cachedElements != nil {
		return nil
	}

	// Cache miss — populate from the database.
	elements, err := uc.repo.ElementRepo.GetElementsByPageID(ctx, projectID, pageID)
	if err != nil {
		return fmt.Errorf("failed to fetch elements for cache refresh: %w", err)
	}

	// Cache a hierarchical tree (not a flat list) so consumers get nested children directly.
	tree := elementutils.BuildElementTree(elements)

	// SET NX: never overwrite a document a concurrent mutation may have just written.
	if err := uc.repo.DocumentRepo.SetDocumentIfAbsent(ctx, projectID, pageID, tree); err != nil {
		return fmt.Errorf("failed to refresh cache: %w", err)
	}

	return nil
}

func (uc *documentUseCase) AddElementsToDocument(ctx context.Context, projectID string, pageID string, elements []domainEntity.Element) error {
	if err := uc.repo.DocumentRepo.SetDocument(ctx, projectID, pageID, elements); err != nil {
		return fmt.Errorf("failed to set document elements in cache: %w", err)
	}
	return nil
}

func (uc *documentUseCase) GetDocumentWithSeq(ctx context.Context, projectID string, pageID string) ([]domainEntity.Element, int64, error) {
	return uc.repo.DocumentRepo.GetDocumentWithSeq(ctx, projectID, pageID)
}

func (uc *documentUseCase) MutateDocument(ctx context.Context, projectID string, pageID string, mutate domainRepo.DocumentMutateFunc) (int64, error) {
	load := func(ctx context.Context) ([]domainEntity.Element, error) {
		elements, err := uc.repo.ElementRepo.GetElementsByPageID(ctx, projectID, pageID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch page elements from db: %w", err)
		}
		return elementutils.BuildElementTree(elements), nil
	}
	return uc.repo.DocumentRepo.UpdateDocument(ctx, projectID, pageID, load, mutate)
}
