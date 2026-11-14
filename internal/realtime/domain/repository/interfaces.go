package repository

import (
	"context"

	"my-go-app/internal/realtime/domain/entity"
	"my-go-app/internal/realtime/domain/logic"
)

// DocumentLoadFunc loads a page document from the source of truth (the database)
// when it is not present in the cache.
type DocumentLoadFunc func(ctx context.Context) ([]entity.Element, error)

// DocumentMutateFunc transforms the current page document into its next version.
// It may be invoked more than once for a single update (optimistic retries), so it
// must derive its result only from the elements it receives. Returning an error
// aborts the update: nothing is written and the page sequence is not advanced.
type DocumentMutateFunc func(elements []entity.Element) ([]entity.Element, error)

type ElementRepository interface {
	CreateElement(ctx context.Context, projectID string, pageID string, element *entity.Element) error
	CreateElements(ctx context.Context, projectID string, pageID string, element []entity.Element) error
	UpdateElement(ctx context.Context, projectID string, pageID string, element *entity.Element) error
	UpdateElements(ctx context.Context, projectID string, pageID string, elements []entity.Element) error
	UpsertElements(ctx context.Context, projectID string, pageID string, elements []entity.Element) error
	DeleteElement(ctx context.Context, projectID string, pageID string, elementID string) error
	GetElementByID(ctx context.Context, projectID string, pageID string, elementID string) (*entity.Element, error)
	GetElementsByPageID(ctx context.Context, projectID string, pageID string) ([]entity.Element, error)
	GetElementsByProjectID(ctx context.Context, projectID string) ([]entity.Element, error)
}

type SnapshotRepository interface {
	SaveSnapshot(ctx context.Context, pageID, projectID, name string, snapshotType entity.SnapshotType, elements []entity.Element) error
	GetSnapshot(ctx context.Context, pageID, projecID string, snapshotType entity.SnapshotType) (*entity.Snapshot, error)
	DeleteSnapshot(ctx context.Context, pageID, projectID string, snapshotType entity.SnapshotType) error
}

type DocumentRepository interface {
	SetDocument(ctx context.Context, projectID string, pageID string, elements []entity.Element) error
	GetDocument(ctx context.Context, projectID string, pageID string) ([]entity.Element, error)
	DeleteDocument(ctx context.Context, projectID string, pageID string) error

	// SetDocumentIfAbsent populates the cache only when no document is cached yet, so a cache
	// fill can never overwrite a concurrent mutation.
	SetDocumentIfAbsent(ctx context.Context, projectID string, pageID string, elements []entity.Element) error
	// GetDocumentWithSeq returns the cached document and the page sequence number as one
	// consistent snapshot. The elements are nil on a cache miss; seq is 0 if the page has no ops.
	GetDocumentWithSeq(ctx context.Context, projectID string, pageID string) ([]entity.Element, int64, error)
	// UpdateDocument atomically applies mutate to the current document (loading it with load on a
	// cache miss), stores the result, marks the page dirty and increments the per-page sequence
	// number. It is safe across server instances. It returns the new sequence number.
	UpdateDocument(ctx context.Context, projectID string, pageID string, load DocumentLoadFunc, mutate DocumentMutateFunc) (int64, error)

	MarkPageDirty(ctx context.Context, projectID string, pageId string) error
	PopDirtyPages(ctx context.Context, limit int) ([]entity.DirtyPage, error)
}

type WorkflowRepository interface {
	CreateWorkflow(ctx context.Context, projectID string, workflow *entity.EventWorkflow) error
	UpdateWorkflow(ctx context.Context, projectID string, workflow *entity.EventWorkflow) error
	DeleteWorkflow(ctx context.Context, projectID string, workflowID string) error
	GetWorkflowByID(ctx context.Context, projectID string, workflowID string) (*entity.EventWorkflow, error)
	GetWorkflowsByProjectID(ctx context.Context, projectID string) ([]entity.EventWorkflow, error)
}

type ProjectRepository interface {
	CreateProject(project *entity.Project) error
	GetProjectByID(projectID string) (*entity.Project, error)
	UpdateProject(project *entity.Project) error
	DeleteProject(projectID string) error
}

type ElementWorkflowRepository interface {
	CreateElementWorkflow(ctx context.Context, projectID string, workflow *entity.ElementEventWorkflow) error
	UpdateElementWorkflow(ctx context.Context, projectID string, workflow *entity.ElementEventWorkflow) error
	DeleteElementWorkflow(ctx context.Context, projectID string, elementID string) error
	GetElementWorkflowByElementID(ctx context.Context, projectID string, elementID string) (*entity.ElementEventWorkflow, error)
	GetElementWorkflowsByProjectID(ctx context.Context, projectID string) ([]entity.ElementEventWorkflow, error)
	GetElementWorkflowsByPageID(ctx context.Context, projectID string, pageID string) ([]entity.ElementEventWorkflow, error)
}

// LogicRepository stores the project's and pages' logic documents.
type LogicRepository interface {
	Move(ctx context.Context, projectID, pageID, userID, kind, name, to string) (project logic.Revisioned, page logic.Revisioned, err error)
	// Apply merges patch into the project (pageID empty) or page logic after
	// checking the user may edit, and returns the resulting document and its new revision.
	Apply(ctx context.Context, projectID, pageID, userID string, patch map[string]any, nullPaths ...[][]string) (logic.Revisioned, error)
	// Get returns the project and page documents and revisions after checking the user may read.
	Get(ctx context.Context, projectID, pageID, userID string) (project logic.Revisioned, page logic.Revisioned, err error)
	// Access reports whether the user may read and may edit the project.
	Access(ctx context.Context, projectID, userID string) (read bool, write bool, err error)
}

type Repository struct {
	LogicRepo           LogicRepository
	ElementRepo         ElementRepository
	SnapshotRepo        SnapshotRepository
	DocumentRepo        DocumentRepository
	ProjectRepo         ProjectRepository
	WorkflowRepo        WorkflowRepository
	ElementWorkflowRepo ElementWorkflowRepository
}
