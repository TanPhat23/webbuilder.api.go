package usecase

import (
	"context"
	"my-go-app/internal/realtime/domain/entity"
	"my-go-app/internal/realtime/domain/logic"
	domainRepo "my-go-app/internal/realtime/domain/repository"
)

/*
ElementUseCase contains operations related to element lifecycle.

Methods are intentionally documented at the interface to preserve the
public contract for implementations and to keep implementation files
comment-free per your request.
*/
type ElementUseCase interface {
	// All mutating operations are applied atomically against the page document and return the
	// per-page sequence number assigned to the change. On error nothing is written and the
	// sequence number is not advanced.

	// CreateElement creates a new element in the given project/page and refreshes the page cache.
	CreateElement(ctx context.Context, projectID string, element *entity.Element) (int64, error)
	CreateElementAt(ctx context.Context, projectID string, element *entity.Element, position *int) (int64, error)
	// UpdateElement replaces an existing element and refreshes the page cache.
	UpdateElement(ctx context.Context, projectID string, pageID string, element *entity.Element) (int64, error)
	// PatchElement merges granular updates into an element: styles and settings are deep-merged
	// (a nil value deletes the key); properties set the matching element fields
	// (name, content, href, src, tailwindStyles, isLocked).
	PatchElement(ctx context.Context, projectID string, pageID string, elementID string, properties map[string]any, styles map[string]any, settings map[string]any) (int64, error)
	// RemoveElement deletes an element and refreshes the page cache.
	RemoveElement(ctx context.Context, projectID string, pageID string, elementID string) (int64, error)
	// MoveElement moves the element under a new parent at the given index and refreshes the page cache.
	MoveElement(ctx context.Context, projectID string, pageID string, elementID string, parentID string, index int) (*entity.Element, int64, error)
	// GetElementsByPageID returns all elements that belong to the specified page.
	GetElementsByPageID(ctx context.Context, projectID string, pageID string) ([]entity.Element, error)
	// GetElementsByProjectID returns all elements across the specified project.
	GetElementsByProjectID(ctx context.Context, projectID string) ([]entity.Element, error)
	// SyncPage forces a refresh of the page cache from the DB and returns the latest elements.
	SyncPage(ctx context.Context, projectID string, pageID string) ([]entity.Element, error)
	// SyncPageWithSeq is SyncPage that also returns the page sequence number matching the returned elements.
	SyncPageWithSeq(ctx context.Context, projectID string, pageID string) ([]entity.Element, int64, error)
}

/*
DocumentUseCase contains operations related to page retrieval and cache management.

Method comments are kept here to document the exact responsibilities of the
DocumentUseCase so implementations elsewhere can remain comment-free.
*/
type DocumentUseCase interface {
	// GetDocument returns the cached snapshot for a page, or builds and returns it if not present.
	GetDocument(ctx context.Context, projectID string, pageID string) ([]entity.Element, error)
	// RefreshCache fetches elements for a page from the DB and updates the cache.
	RefreshCache(ctx context.Context, projectID string, pageID string) error
	// GetDocumentWithSeq returns the cached document together with its page sequence number (consistent snapshot).
	GetDocumentWithSeq(ctx context.Context, projectID string, pageID string) ([]entity.Element, int64, error)
	// MutateDocument atomically applies mutate to the page document and returns the new page sequence number.
	MutateDocument(ctx context.Context, projectID string, pageID string, mutate domainRepo.DocumentMutateFunc) (int64, error)
	// AddElementsToDocument sets the cached snapshot for the given page.
	// It is not safe for read-modify-write; use MutateDocument for that.
	AddElementsToDocument(ctx context.Context, projectID string, pageID string, elements []entity.Element) error
}

type PersistUseCase interface {
	Start(ctx context.Context)

	MarkDirty(ctx context.Context, projectID, pageID string) error

	Flush(ctx context.Context) error
}

type WorkflowUseCase interface {
	// CreateWorkflow creates a new workflow for a project.
	CreateWorkflow(ctx context.Context, projectID string, workflow *entity.EventWorkflow) error
	// UpdateWorkflow updates an existing workflow.
	UpdateWorkflow(ctx context.Context, projectID string, workflow *entity.EventWorkflow) error
	// RemoveWorkflow deletes a workflow.
	RemoveWorkflow(ctx context.Context, projectID string, workflowID string) error
	// GetWorkflowByID retrieves a single workflow.
	GetWorkflowByID(ctx context.Context, projectID string, workflowID string) (*entity.EventWorkflow, error)
	// GetWorkflowsByProject retrieves all workflows for a project.
	GetWorkflowsByProject(ctx context.Context, projectID string) ([]entity.EventWorkflow, error)
}

type ElementWorkflowsUseCase interface {
	// AddElementWorkflow creates a new link between an element and a workflow.
	CreateElementWorkflow(ctx context.Context, projectID string, workflow *entity.ElementEventWorkflow) error
	// GetElementWorkflowByElementID retrieves a singular event workflow for a specific element.
	GetElementWorkflowByElementID(ctx context.Context, projectID string, elementID string) (*entity.ElementEventWorkflow, error)
	// GetWorkflowsByProject retrieves all element-workflow mappings for an entire project.
	GetWorkflowsByProject(ctx context.Context, projectID string) ([]entity.ElementEventWorkflow, error)
	// GetWorkflowsByPage retrieves all element-workflow mappings for a specific page.
	GetWorkflowsByPage(ctx context.Context, projectID string, pageID string) ([]entity.ElementEventWorkflow, error)
	// UpdateElementWorkflow updates an existing element-workflow mapping.
	UpdateElementWorkflow(ctx context.Context, projectID string, workflow *entity.ElementEventWorkflow) error
	// RemoveElementWorkflow deletes workflow mappings for a specific element.
	RemoveElementWorkflow(ctx context.Context, projectID string, elementID string) error
}

// LogicUseCase edits and reads page/project logic on behalf of a user. Errors
// carry AppError codes: ACCESS_DENIED, NOT_FOUND or VALIDATION_ERROR.
type LogicUseCase interface {
	Update(ctx context.Context, projectID, pageID, userID string, patch map[string]any) (logic.Revisioned, error)
	Get(ctx context.Context, projectID, pageID, userID string) (project logic.Revisioned, page logic.Revisioned, err error)
	Access(ctx context.Context, projectID, userID string) (read bool, write bool, err error)
}

// UseCase provides a central container for all use case interfaces.
type UseCase struct {
	Element          ElementUseCase
	Document         DocumentUseCase
	Persist          PersistUseCase
	Workflow         WorkflowUseCase
	ElementWorkflows ElementWorkflowsUseCase
	Logic            LogicUseCase
}
