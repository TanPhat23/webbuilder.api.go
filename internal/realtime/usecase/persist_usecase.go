package usecase

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"

	domainEntity "my-go-app/internal/realtime/domain/entity"
	domainRepo "my-go-app/internal/realtime/domain/repository"
	domainUseCase "my-go-app/internal/realtime/domain/usecase"
	elementUtils "my-go-app/internal/realtime/utils/element"
	"my-go-app/internal/realtime/utils/logger"
)

type PersistenceUseCase struct {
	repo   *domainRepo.Repository
	logger logger.Logger
}

func NewPersistenceUsecase(repo *domainRepo.Repository, logger logger.Logger) domainUseCase.PersistUseCase {
	return &PersistenceUseCase{
		repo:   repo,
		logger: logger,
	}
}

type Result struct {
	ProjectID, PageID string
	Err               error
}

func (uc *PersistenceUseCase) worker(ctx context.Context, tasks <-chan domainEntity.DirtyPage, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	for dirty := range tasks {
		elementTree, err := uc.repo.DocumentRepo.GetDocument(ctx, dirty.ProjectID, dirty.PageID)
		if err == nil && elementTree == nil {
			err = fmt.Errorf("dirty page document is missing from cache")
		}
		if err != nil {
			results <- Result{
				ProjectID: dirty.ProjectID,
				PageID:    dirty.PageID,
				Err:       err,
			}
			continue
		}
		flattenedEl := elementUtils.FlattenElements(elementTree)
		if err := uc.repo.ElementRepo.UpdateElements(ctx, dirty.ProjectID, dirty.PageID, flattenedEl); err != nil {
			results <- Result{
				ProjectID: dirty.ProjectID,
				PageID:    dirty.PageID,
				Err:       err,
			}
			continue
		}
		if err := uc.repo.SnapshotRepo.SaveSnapshot(ctx, dirty.ProjectID, dirty.PageID, "autosave", domainEntity.SnapshotTypes.Working, elementTree); err != nil {
			results <- Result{
				ProjectID: dirty.ProjectID,
				PageID:    dirty.PageID,
				Err:       err,
			}
			continue
		}
		results <- Result{
			ProjectID: dirty.ProjectID,
			PageID:    dirty.PageID,
			Err:       nil,
		}
	}
}

func (uc *PersistenceUseCase) Start(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := uc.Flush(ctx); err != nil {
				uc.logger.Errorf("Persistence background flush failed: %v", err)
			}
		}
	}
}

func (uc *PersistenceUseCase) Flush(ctx context.Context) error {
	dirtyPages, err := uc.repo.DocumentRepo.PopDirtyPages(ctx, 100)
	if err != nil {
		return err
	}
	if len(dirtyPages) == 0 {
		return nil
	}

	workerCount := runtime.NumCPU()
	if len(dirtyPages) < workerCount {
		workerCount = len(dirtyPages)
	}

	tasks := make(chan domainEntity.DirtyPage, len(dirtyPages))
	results := make(chan Result, len(dirtyPages))
	wg := sync.WaitGroup{}

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go uc.worker(ctx, tasks, results, &wg)
	}

	for _, dirtyPage := range dirtyPages {
		tasks <- dirtyPage
	}
	close(tasks)

	// Wait for workers to finish processing all tasks
	wg.Wait()
	close(results)

	// Collect results and log errors
	var lastErr error
	for result := range results {
		if result.Err != nil {
			uc.logger.Errorf("Error saving elements to database for pageId %s projectID %s: %v", result.PageID, result.ProjectID, result.Err)
			lastErr = errors.Join(lastErr, result.Err)
			// A flush can be interrupted by shutdown; requeue with a fresh context
			// so a transient database failure doesn't discard the pending page.
			retryCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			err := uc.repo.DocumentRepo.MarkPageDirty(retryCtx, result.ProjectID, result.PageID)
			cancel()
			if err != nil {
				lastErr = errors.Join(lastErr, fmt.Errorf("requeue page %s: %w", result.PageID, err))
			}
		}
	}
	return lastErr
}

func (uc *PersistenceUseCase) MarkDirty(ctx context.Context, projectID, pageID string) error {
	uc.logger.Debugf("MarkDirty called for project=%s page=%s", projectID, pageID)
	if err := uc.repo.DocumentRepo.MarkPageDirty(ctx, projectID, pageID); err != nil {
		uc.logger.Errorf("failed to mark page dirty: %v", err)
		return err
	}
	uc.logger.Debugf("MarkDirty succeeded for project=%s page=%s", projectID, pageID)
	return nil
}
