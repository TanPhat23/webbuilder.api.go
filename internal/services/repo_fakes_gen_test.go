// Code generated for tests from internal/repositories/interface.repository.go. DO NOT EDIT.
// Each fake records the names of called methods in Calls and delegates to the
// matching On<Method> hook. Calling a method without a hook panics, so a test
// fails loudly when the unit under test touches a repository it should not.

package services

import (
	"context"

	"my-go-app/internal/models"
)

type fkElementRepo struct {
	Calls                  []string
	OnList                 func(ctx context.Context, projectID string, pageID ...string) ([]models.EditorElement, error)
	OnReplace              func(ctx context.Context, projectID string, elements []models.EditorElement) error
	OnGetByID              func(ctx context.Context, elementID string) (*models.Element, error)
	OnListByPage           func(ctx context.Context, pageID string) ([]models.Element, error)
	OnListByPageIDs        func(ctx context.Context, projectID string, pageIDs []string) ([]models.EditorElement, error)
	OnListChildren         func(ctx context.Context, parentID string) ([]models.Element, error)
	OnListRoots            func(ctx context.Context, projectID string) ([]models.Element, error)
	OnCreate               func(ctx context.Context, element *models.Element) error
	OnUpdate               func(ctx context.Context, element *models.Element) error
	OnUpdateEventWorkflows func(ctx context.Context, elementID string, workflows []byte) error
	OnDelete               func(ctx context.Context, elementID string) error
	OnDeleteByPage         func(ctx context.Context, pageID string) error
	OnDeleteByProject      func(ctx context.Context, projectID string) error
	OnCountByProject       func(ctx context.Context, projectID string) (int64, error)
	OnGetWithRelations     func(ctx context.Context, elementID string) (*models.Element, error)
	OnListByIDs            func(ctx context.Context, elementIDs []string) ([]models.Element, error)
}

func (f *fkElementRepo) List(ctx context.Context, projectID string, pageID ...string) ([]models.EditorElement, error) {
	f.Calls = append(f.Calls, "List")
	if f.OnList == nil {
		panic("fkElementRepo: unexpected call to List")
	}
	return f.OnList(ctx, projectID, pageID...)
}

func (f *fkElementRepo) Replace(ctx context.Context, projectID string, elements []models.EditorElement) error {
	f.Calls = append(f.Calls, "Replace")
	if f.OnReplace == nil {
		panic("fkElementRepo: unexpected call to Replace")
	}
	return f.OnReplace(ctx, projectID, elements)
}

func (f *fkElementRepo) GetByID(ctx context.Context, elementID string) (*models.Element, error) {
	f.Calls = append(f.Calls, "GetByID")
	if f.OnGetByID == nil {
		panic("fkElementRepo: unexpected call to GetByID")
	}
	return f.OnGetByID(ctx, elementID)
}

func (f *fkElementRepo) ListByPage(ctx context.Context, pageID string) ([]models.Element, error) {
	f.Calls = append(f.Calls, "ListByPage")
	if f.OnListByPage == nil {
		panic("fkElementRepo: unexpected call to ListByPage")
	}
	return f.OnListByPage(ctx, pageID)
}

func (f *fkElementRepo) ListByPageIDs(ctx context.Context, projectID string, pageIDs []string) ([]models.EditorElement, error) {
	f.Calls = append(f.Calls, "ListByPageIDs")
	if f.OnListByPageIDs == nil {
		panic("fkElementRepo: unexpected call to ListByPageIDs")
	}
	return f.OnListByPageIDs(ctx, projectID, pageIDs)
}

func (f *fkElementRepo) ListChildren(ctx context.Context, parentID string) ([]models.Element, error) {
	f.Calls = append(f.Calls, "ListChildren")
	if f.OnListChildren == nil {
		panic("fkElementRepo: unexpected call to ListChildren")
	}
	return f.OnListChildren(ctx, parentID)
}

func (f *fkElementRepo) ListRoots(ctx context.Context, projectID string) ([]models.Element, error) {
	f.Calls = append(f.Calls, "ListRoots")
	if f.OnListRoots == nil {
		panic("fkElementRepo: unexpected call to ListRoots")
	}
	return f.OnListRoots(ctx, projectID)
}

func (f *fkElementRepo) Create(ctx context.Context, element *models.Element) error {
	f.Calls = append(f.Calls, "Create")
	if f.OnCreate == nil {
		panic("fkElementRepo: unexpected call to Create")
	}
	return f.OnCreate(ctx, element)
}

func (f *fkElementRepo) Update(ctx context.Context, element *models.Element) error {
	f.Calls = append(f.Calls, "Update")
	if f.OnUpdate == nil {
		panic("fkElementRepo: unexpected call to Update")
	}
	return f.OnUpdate(ctx, element)
}

func (f *fkElementRepo) UpdateEventWorkflows(ctx context.Context, elementID string, workflows []byte) error {
	f.Calls = append(f.Calls, "UpdateEventWorkflows")
	if f.OnUpdateEventWorkflows == nil {
		panic("fkElementRepo: unexpected call to UpdateEventWorkflows")
	}
	return f.OnUpdateEventWorkflows(ctx, elementID, workflows)
}

func (f *fkElementRepo) Delete(ctx context.Context, elementID string) error {
	f.Calls = append(f.Calls, "Delete")
	if f.OnDelete == nil {
		panic("fkElementRepo: unexpected call to Delete")
	}
	return f.OnDelete(ctx, elementID)
}

func (f *fkElementRepo) DeleteByPage(ctx context.Context, pageID string) error {
	f.Calls = append(f.Calls, "DeleteByPage")
	if f.OnDeleteByPage == nil {
		panic("fkElementRepo: unexpected call to DeleteByPage")
	}
	return f.OnDeleteByPage(ctx, pageID)
}

func (f *fkElementRepo) DeleteByProject(ctx context.Context, projectID string) error {
	f.Calls = append(f.Calls, "DeleteByProject")
	if f.OnDeleteByProject == nil {
		panic("fkElementRepo: unexpected call to DeleteByProject")
	}
	return f.OnDeleteByProject(ctx, projectID)
}

func (f *fkElementRepo) CountByProject(ctx context.Context, projectID string) (int64, error) {
	f.Calls = append(f.Calls, "CountByProject")
	if f.OnCountByProject == nil {
		panic("fkElementRepo: unexpected call to CountByProject")
	}
	return f.OnCountByProject(ctx, projectID)
}

func (f *fkElementRepo) GetWithRelations(ctx context.Context, elementID string) (*models.Element, error) {
	f.Calls = append(f.Calls, "GetWithRelations")
	if f.OnGetWithRelations == nil {
		panic("fkElementRepo: unexpected call to GetWithRelations")
	}
	return f.OnGetWithRelations(ctx, elementID)
}

func (f *fkElementRepo) ListByIDs(ctx context.Context, elementIDs []string) ([]models.Element, error) {
	f.Calls = append(f.Calls, "ListByIDs")
	if f.OnListByIDs == nil {
		panic("fkElementRepo: unexpected call to ListByIDs")
	}
	return f.OnListByIDs(ctx, elementIDs)
}

// Called reports whether the named method was invoked at least once.
func (f *fkElementRepo) Called(name string) bool {
	for _, c := range f.Calls {
		if c == name {
			return true
		}
	}
	return false
}

type fkElementCommentRepo struct {
	Calls             []string
	OnCreate          func(ctx context.Context, comment *models.ElementComment) (*models.ElementComment, error)
	OnGetByID         func(ctx context.Context, id string) (*models.ElementComment, error)
	OnList            func(ctx context.Context, elementID string, filter *models.ElementCommentFilter) ([]models.ElementComment, error)
	OnUpdate          func(ctx context.Context, id string, updates map[string]any) (*models.ElementComment, error)
	OnDelete          func(ctx context.Context, id string) error
	OnListByAuthor    func(ctx context.Context, authorID string, limit int, offset int) ([]models.ElementComment, error)
	OnCount           func(ctx context.Context, elementID string) (int64, error)
	OnToggleResolved  func(ctx context.Context, id string) (*models.ElementComment, error)
	OnDeleteByElement func(ctx context.Context, elementID string) error
	OnListByProject   func(ctx context.Context, projectID string, limit int, offset int) ([]models.ElementComment, error)
	OnCountByProject  func(ctx context.Context, projectID string) (int64, error)
}

func (f *fkElementCommentRepo) Create(ctx context.Context, comment *models.ElementComment) (*models.ElementComment, error) {
	f.Calls = append(f.Calls, "Create")
	if f.OnCreate == nil {
		panic("fkElementCommentRepo: unexpected call to Create")
	}
	return f.OnCreate(ctx, comment)
}

func (f *fkElementCommentRepo) GetByID(ctx context.Context, id string) (*models.ElementComment, error) {
	f.Calls = append(f.Calls, "GetByID")
	if f.OnGetByID == nil {
		panic("fkElementCommentRepo: unexpected call to GetByID")
	}
	return f.OnGetByID(ctx, id)
}

func (f *fkElementCommentRepo) List(ctx context.Context, elementID string, filter *models.ElementCommentFilter) ([]models.ElementComment, error) {
	f.Calls = append(f.Calls, "List")
	if f.OnList == nil {
		panic("fkElementCommentRepo: unexpected call to List")
	}
	return f.OnList(ctx, elementID, filter)
}

func (f *fkElementCommentRepo) Update(ctx context.Context, id string, updates map[string]any) (*models.ElementComment, error) {
	f.Calls = append(f.Calls, "Update")
	if f.OnUpdate == nil {
		panic("fkElementCommentRepo: unexpected call to Update")
	}
	return f.OnUpdate(ctx, id, updates)
}

func (f *fkElementCommentRepo) Delete(ctx context.Context, id string) error {
	f.Calls = append(f.Calls, "Delete")
	if f.OnDelete == nil {
		panic("fkElementCommentRepo: unexpected call to Delete")
	}
	return f.OnDelete(ctx, id)
}

func (f *fkElementCommentRepo) ListByAuthor(ctx context.Context, authorID string, limit int, offset int) ([]models.ElementComment, error) {
	f.Calls = append(f.Calls, "ListByAuthor")
	if f.OnListByAuthor == nil {
		panic("fkElementCommentRepo: unexpected call to ListByAuthor")
	}
	return f.OnListByAuthor(ctx, authorID, limit, offset)
}

func (f *fkElementCommentRepo) Count(ctx context.Context, elementID string) (int64, error) {
	f.Calls = append(f.Calls, "Count")
	if f.OnCount == nil {
		panic("fkElementCommentRepo: unexpected call to Count")
	}
	return f.OnCount(ctx, elementID)
}

func (f *fkElementCommentRepo) ToggleResolved(ctx context.Context, id string) (*models.ElementComment, error) {
	f.Calls = append(f.Calls, "ToggleResolved")
	if f.OnToggleResolved == nil {
		panic("fkElementCommentRepo: unexpected call to ToggleResolved")
	}
	return f.OnToggleResolved(ctx, id)
}

func (f *fkElementCommentRepo) DeleteByElement(ctx context.Context, elementID string) error {
	f.Calls = append(f.Calls, "DeleteByElement")
	if f.OnDeleteByElement == nil {
		panic("fkElementCommentRepo: unexpected call to DeleteByElement")
	}
	return f.OnDeleteByElement(ctx, elementID)
}

func (f *fkElementCommentRepo) ListByProject(ctx context.Context, projectID string, limit int, offset int) ([]models.ElementComment, error) {
	f.Calls = append(f.Calls, "ListByProject")
	if f.OnListByProject == nil {
		panic("fkElementCommentRepo: unexpected call to ListByProject")
	}
	return f.OnListByProject(ctx, projectID, limit, offset)
}

func (f *fkElementCommentRepo) CountByProject(ctx context.Context, projectID string) (int64, error) {
	f.Calls = append(f.Calls, "CountByProject")
	if f.OnCountByProject == nil {
		panic("fkElementCommentRepo: unexpected call to CountByProject")
	}
	return f.OnCountByProject(ctx, projectID)
}

// Called reports whether the named method was invoked at least once.
func (f *fkElementCommentRepo) Called(name string) bool {
	for _, c := range f.Calls {
		if c == name {
			return true
		}
	}
	return false
}

type fkProjectRepo struct {
	Calls                  []string
	OnGetPublicByID        func(ctx context.Context, projectID string) (*models.Project, error)
	OnGetPublicBySubdomain func(ctx context.Context, subdomain string) (*models.Project, error)
	OnGetByID              func(ctx context.Context, projectID string, userID string) (*models.Project, error)
	OnGetWithAccess        func(ctx context.Context, projectID string, userID string) (*models.Project, error)
	OnListByUser           func(ctx context.Context, userID string) ([]models.Project, error)
	OnListByCollaborator   func(ctx context.Context, userID string) ([]models.Project, error)
	OnListPages            func(ctx context.Context, projectID string, userID string) ([]models.Page, error)
	OnCreate               func(ctx context.Context, project *models.Project) error
	OnCreateWithPage       func(ctx context.Context, project *models.Project, page *models.Page) error
	OnGetAccess            func(ctx context.Context, projectID string, userID string) (*models.ProjectAccess, error)
	OnUpdate               func(ctx context.Context, projectID string, userID string, updates map[string]any) (*models.Project, error)
	OnUpdateFields         func(ctx context.Context, projectID string, userID string, updates map[string]any) (*models.Project, error)
	OnTransferOwnership    func(ctx context.Context, projectID string, currentOwnerID string, newOwnerID string) error
	OnDelete               func(ctx context.Context, projectID string, userID string) error
	OnHardDelete           func(ctx context.Context, projectID string, userID string) error
	OnRestore              func(ctx context.Context, projectID string, userID string) error
	OnExistsForUser        func(ctx context.Context, projectID string, userID string) (bool, error)
	OnGetWithLock          func(ctx context.Context, projectID string, userID string) (*models.Project, error)
}

func (f *fkProjectRepo) GetPublicByID(ctx context.Context, projectID string) (*models.Project, error) {
	f.Calls = append(f.Calls, "GetPublicByID")
	if f.OnGetPublicByID == nil {
		panic("fkProjectRepo: unexpected call to GetPublicByID")
	}
	return f.OnGetPublicByID(ctx, projectID)
}

func (f *fkProjectRepo) GetPublicBySubdomain(ctx context.Context, subdomain string) (*models.Project, error) {
	f.Calls = append(f.Calls, "GetPublicBySubdomain")
	if f.OnGetPublicBySubdomain == nil {
		panic("fkProjectRepo: unexpected call to GetPublicBySubdomain")
	}
	return f.OnGetPublicBySubdomain(ctx, subdomain)
}

func (f *fkProjectRepo) GetByID(ctx context.Context, projectID string, userID string) (*models.Project, error) {
	f.Calls = append(f.Calls, "GetByID")
	if f.OnGetByID == nil {
		panic("fkProjectRepo: unexpected call to GetByID")
	}
	return f.OnGetByID(ctx, projectID, userID)
}

func (f *fkProjectRepo) GetWithAccess(ctx context.Context, projectID string, userID string) (*models.Project, error) {
	f.Calls = append(f.Calls, "GetWithAccess")
	if f.OnGetWithAccess == nil {
		panic("fkProjectRepo: unexpected call to GetWithAccess")
	}
	return f.OnGetWithAccess(ctx, projectID, userID)
}

func (f *fkProjectRepo) ListByUser(ctx context.Context, userID string) ([]models.Project, error) {
	f.Calls = append(f.Calls, "ListByUser")
	if f.OnListByUser == nil {
		panic("fkProjectRepo: unexpected call to ListByUser")
	}
	return f.OnListByUser(ctx, userID)
}

func (f *fkProjectRepo) ListByCollaborator(ctx context.Context, userID string) ([]models.Project, error) {
	f.Calls = append(f.Calls, "ListByCollaborator")
	if f.OnListByCollaborator == nil {
		panic("fkProjectRepo: unexpected call to ListByCollaborator")
	}
	return f.OnListByCollaborator(ctx, userID)
}

func (f *fkProjectRepo) ListPages(ctx context.Context, projectID string, userID string) ([]models.Page, error) {
	f.Calls = append(f.Calls, "ListPages")
	if f.OnListPages == nil {
		panic("fkProjectRepo: unexpected call to ListPages")
	}
	return f.OnListPages(ctx, projectID, userID)
}

func (f *fkProjectRepo) Create(ctx context.Context, project *models.Project) error {
	f.Calls = append(f.Calls, "Create")
	if f.OnCreate == nil {
		panic("fkProjectRepo: unexpected call to Create")
	}
	return f.OnCreate(ctx, project)
}

func (f *fkProjectRepo) CreateWithPage(ctx context.Context, project *models.Project, page *models.Page) error {
	f.Calls = append(f.Calls, "CreateWithPage")
	if f.OnCreateWithPage == nil {
		panic("fkProjectRepo: unexpected call to CreateWithPage")
	}
	return f.OnCreateWithPage(ctx, project, page)
}

func (f *fkProjectRepo) GetAccess(ctx context.Context, projectID string, userID string) (*models.ProjectAccess, error) {
	f.Calls = append(f.Calls, "GetAccess")
	if f.OnGetAccess == nil {
		panic("fkProjectRepo: unexpected call to GetAccess")
	}
	return f.OnGetAccess(ctx, projectID, userID)
}

func (f *fkProjectRepo) Update(ctx context.Context, projectID string, userID string, updates map[string]any) (*models.Project, error) {
	f.Calls = append(f.Calls, "Update")
	if f.OnUpdate == nil {
		panic("fkProjectRepo: unexpected call to Update")
	}
	return f.OnUpdate(ctx, projectID, userID, updates)
}

func (f *fkProjectRepo) UpdateFields(ctx context.Context, projectID string, userID string, updates map[string]any) (*models.Project, error) {
	f.Calls = append(f.Calls, "UpdateFields")
	if f.OnUpdateFields != nil {
		return f.OnUpdateFields(ctx, projectID, userID, updates)
	}
	if f.OnUpdate != nil {
		return f.OnUpdate(ctx, projectID, userID, updates)
	}
	panic("fkProjectRepo: unexpected call to UpdateFields")
}

func (f *fkProjectRepo) TransferOwnership(ctx context.Context, projectID string, currentOwnerID string, newOwnerID string) error {
	f.Calls = append(f.Calls, "TransferOwnership")
	if f.OnTransferOwnership == nil {
		panic("fkProjectRepo: unexpected call to TransferOwnership")
	}
	return f.OnTransferOwnership(ctx, projectID, currentOwnerID, newOwnerID)
}

func (f *fkProjectRepo) Delete(ctx context.Context, projectID string, userID string) error {
	f.Calls = append(f.Calls, "Delete")
	if f.OnDelete == nil {
		panic("fkProjectRepo: unexpected call to Delete")
	}
	return f.OnDelete(ctx, projectID, userID)
}

func (f *fkProjectRepo) HardDelete(ctx context.Context, projectID string, userID string) error {
	f.Calls = append(f.Calls, "HardDelete")
	if f.OnHardDelete == nil {
		panic("fkProjectRepo: unexpected call to HardDelete")
	}
	return f.OnHardDelete(ctx, projectID, userID)
}

func (f *fkProjectRepo) Restore(ctx context.Context, projectID string, userID string) error {
	f.Calls = append(f.Calls, "Restore")
	if f.OnRestore == nil {
		panic("fkProjectRepo: unexpected call to Restore")
	}
	return f.OnRestore(ctx, projectID, userID)
}

func (f *fkProjectRepo) ExistsForUser(ctx context.Context, projectID string, userID string) (bool, error) {
	f.Calls = append(f.Calls, "ExistsForUser")
	if f.OnExistsForUser == nil {
		panic("fkProjectRepo: unexpected call to ExistsForUser")
	}
	return f.OnExistsForUser(ctx, projectID, userID)
}

func (f *fkProjectRepo) GetWithLock(ctx context.Context, projectID string, userID string) (*models.Project, error) {
	f.Calls = append(f.Calls, "GetWithLock")
	if f.OnGetWithLock == nil {
		panic("fkProjectRepo: unexpected call to GetWithLock")
	}
	return f.OnGetWithLock(ctx, projectID, userID)
}

// Called reports whether the named method was invoked at least once.
func (f *fkProjectRepo) Called(name string) bool {
	for _, c := range f.Calls {
		if c == name {
			return true
		}
	}
	return false
}

type fkSnapshotRepo struct {
	Calls           []string
	OnSave          func(ctx context.Context, projectID string, snapshot *models.Snapshot) error
	OnListByProject func(ctx context.Context, projectID string) ([]models.Snapshot, error)
	OnGetByID       func(ctx context.Context, snapshotID string) (*models.Snapshot, error)
	OnDelete        func(ctx context.Context, snapshotID string) error
}

func (f *fkSnapshotRepo) Save(ctx context.Context, projectID string, snapshot *models.Snapshot) error {
	f.Calls = append(f.Calls, "Save")
	if f.OnSave == nil {
		panic("fkSnapshotRepo: unexpected call to Save")
	}
	return f.OnSave(ctx, projectID, snapshot)
}

func (f *fkSnapshotRepo) ListByProject(ctx context.Context, projectID string) ([]models.Snapshot, error) {
	f.Calls = append(f.Calls, "ListByProject")
	if f.OnListByProject == nil {
		panic("fkSnapshotRepo: unexpected call to ListByProject")
	}
	return f.OnListByProject(ctx, projectID)
}

func (f *fkSnapshotRepo) GetByID(ctx context.Context, snapshotID string) (*models.Snapshot, error) {
	f.Calls = append(f.Calls, "GetByID")
	if f.OnGetByID == nil {
		panic("fkSnapshotRepo: unexpected call to GetByID")
	}
	return f.OnGetByID(ctx, snapshotID)
}

func (f *fkSnapshotRepo) Delete(ctx context.Context, snapshotID string) error {
	f.Calls = append(f.Calls, "Delete")
	if f.OnDelete == nil {
		panic("fkSnapshotRepo: unexpected call to Delete")
	}
	return f.OnDelete(ctx, snapshotID)
}

// Called reports whether the named method was invoked at least once.
func (f *fkSnapshotRepo) Called(name string) bool {
	for _, c := range f.Calls {
		if c == name {
			return true
		}
	}
	return false
}

type fkPageRepo struct {
	Calls           []string
	OnListByProject func(ctx context.Context, projectID string) ([]models.Page, error)
	OnGetByID       func(ctx context.Context, pageID string, projectID string) (*models.Page, error)
	OnCreate        func(ctx context.Context, page *models.Page) error
	OnUpdate        func(ctx context.Context, page *models.Page) error
	OnUpdateFields  func(ctx context.Context, pageID string, updates map[string]any) error
	OnDelete        func(ctx context.Context, pageID string) error
	OnDeleteForUser func(ctx context.Context, pageID string, projectID string, userID string) error
}

func (f *fkPageRepo) ListByProject(ctx context.Context, projectID string) ([]models.Page, error) {
	f.Calls = append(f.Calls, "ListByProject")
	if f.OnListByProject == nil {
		panic("fkPageRepo: unexpected call to ListByProject")
	}
	return f.OnListByProject(ctx, projectID)
}

func (f *fkPageRepo) GetByID(ctx context.Context, pageID string, projectID string) (*models.Page, error) {
	f.Calls = append(f.Calls, "GetByID")
	if f.OnGetByID == nil {
		panic("fkPageRepo: unexpected call to GetByID")
	}
	return f.OnGetByID(ctx, pageID, projectID)
}

func (f *fkPageRepo) Create(ctx context.Context, page *models.Page) error {
	f.Calls = append(f.Calls, "Create")
	if f.OnCreate == nil {
		panic("fkPageRepo: unexpected call to Create")
	}
	return f.OnCreate(ctx, page)
}

func (f *fkPageRepo) Update(ctx context.Context, page *models.Page) error {
	f.Calls = append(f.Calls, "Update")
	if f.OnUpdate == nil {
		panic("fkPageRepo: unexpected call to Update")
	}
	return f.OnUpdate(ctx, page)
}

func (f *fkPageRepo) UpdateFields(ctx context.Context, pageID string, updates map[string]any) error {
	f.Calls = append(f.Calls, "UpdateFields")
	if f.OnUpdateFields == nil {
		panic("fkPageRepo: unexpected call to UpdateFields")
	}
	return f.OnUpdateFields(ctx, pageID, updates)
}

func (f *fkPageRepo) Delete(ctx context.Context, pageID string) error {
	f.Calls = append(f.Calls, "Delete")
	if f.OnDelete == nil {
		panic("fkPageRepo: unexpected call to Delete")
	}
	return f.OnDelete(ctx, pageID)
}

func (f *fkPageRepo) DeleteForUser(ctx context.Context, pageID string, projectID string, userID string) error {
	f.Calls = append(f.Calls, "DeleteForUser")
	if f.OnDeleteForUser == nil {
		panic("fkPageRepo: unexpected call to DeleteForUser")
	}
	return f.OnDeleteForUser(ctx, pageID, projectID, userID)
}

// Called reports whether the named method was invoked at least once.
func (f *fkPageRepo) Called(name string) bool {
	for _, c := range f.Calls {
		if c == name {
			return true
		}
	}
	return false
}

type fkContentTypeRepo struct {
	Calls                     []string
	OnList                    func(ctx context.Context) ([]models.ContentType, error)
	OnListByProject           func(ctx context.Context, projectID string) ([]models.ContentType, error)
	OnListByUser              func(ctx context.Context, userID string) ([]models.ContentType, error)
	OnGetByID                 func(ctx context.Context, id string) (*models.ContentType, error)
	OnGetByProjectAndSlugOrID func(ctx context.Context, projectID, slugOrID string) (*models.ContentType, error)
	OnCreate                  func(ctx context.Context, ct *models.ContentType) (*models.ContentType, error)
	OnUpdate                  func(ctx context.Context, id string, updates map[string]any) (*models.ContentType, error)
	OnDelete                  func(ctx context.Context, id string) error
}

func (f *fkContentTypeRepo) List(ctx context.Context) ([]models.ContentType, error) {
	f.Calls = append(f.Calls, "List")
	if f.OnList == nil {
		panic("fkContentTypeRepo: unexpected call to List")
	}
	return f.OnList(ctx)
}

func (f *fkContentTypeRepo) ListByProject(ctx context.Context, projectID string) ([]models.ContentType, error) {
	f.Calls = append(f.Calls, "ListByProject")
	if f.OnListByProject == nil {
		panic("fkContentTypeRepo: unexpected call to ListByProject")
	}
	return f.OnListByProject(ctx, projectID)
}

func (f *fkContentTypeRepo) ListByUser(ctx context.Context, userID string) ([]models.ContentType, error) {
	f.Calls = append(f.Calls, "ListByUser")
	if f.OnListByUser == nil {
		panic("fkContentTypeRepo: unexpected call to ListByUser")
	}
	return f.OnListByUser(ctx, userID)
}

func (f *fkContentTypeRepo) GetByID(ctx context.Context, id string) (*models.ContentType, error) {
	f.Calls = append(f.Calls, "GetByID")
	if f.OnGetByID == nil {
		panic("fkContentTypeRepo: unexpected call to GetByID")
	}
	return f.OnGetByID(ctx, id)
}

func (f *fkContentTypeRepo) GetByProjectAndSlugOrID(ctx context.Context, projectID, slugOrID string) (*models.ContentType, error) {
	f.Calls = append(f.Calls, "GetByProjectAndSlugOrID")
	if f.OnGetByProjectAndSlugOrID == nil {
		panic("fkContentTypeRepo: unexpected call to GetByProjectAndSlugOrID")
	}
	return f.OnGetByProjectAndSlugOrID(ctx, projectID, slugOrID)
}

func (f *fkContentTypeRepo) Create(ctx context.Context, ct *models.ContentType) (*models.ContentType, error) {
	f.Calls = append(f.Calls, "Create")
	if f.OnCreate == nil {
		panic("fkContentTypeRepo: unexpected call to Create")
	}
	return f.OnCreate(ctx, ct)
}

func (f *fkContentTypeRepo) Update(ctx context.Context, id string, updates map[string]any) (*models.ContentType, error) {
	f.Calls = append(f.Calls, "Update")
	if f.OnUpdate == nil {
		panic("fkContentTypeRepo: unexpected call to Update")
	}
	return f.OnUpdate(ctx, id, updates)
}

func (f *fkContentTypeRepo) Delete(ctx context.Context, id string) error {
	f.Calls = append(f.Calls, "Delete")
	if f.OnDelete == nil {
		panic("fkContentTypeRepo: unexpected call to Delete")
	}
	return f.OnDelete(ctx, id)
}

// Called reports whether the named method was invoked at least once.
func (f *fkContentTypeRepo) Called(name string) bool {
	for _, c := range f.Calls {
		if c == name {
			return true
		}
	}
	return false
}

type fkContentFieldRepo struct {
	Calls               []string
	OnListByContentType func(ctx context.Context, contentTypeID string) ([]models.ContentField, error)
	OnGetByID           func(ctx context.Context, id string) (*models.ContentField, error)
	OnCreate            func(ctx context.Context, cf *models.ContentField) (*models.ContentField, error)
	OnUpdate            func(ctx context.Context, id string, updates map[string]any) (*models.ContentField, error)
	OnDelete            func(ctx context.Context, id string) error
}

func (f *fkContentFieldRepo) ListByContentType(ctx context.Context, contentTypeID string) ([]models.ContentField, error) {
	f.Calls = append(f.Calls, "ListByContentType")
	if f.OnListByContentType == nil {
		panic("fkContentFieldRepo: unexpected call to ListByContentType")
	}
	return f.OnListByContentType(ctx, contentTypeID)
}

func (f *fkContentFieldRepo) GetByID(ctx context.Context, id string) (*models.ContentField, error) {
	f.Calls = append(f.Calls, "GetByID")
	if f.OnGetByID == nil {
		panic("fkContentFieldRepo: unexpected call to GetByID")
	}
	return f.OnGetByID(ctx, id)
}

func (f *fkContentFieldRepo) Create(ctx context.Context, cf *models.ContentField) (*models.ContentField, error) {
	f.Calls = append(f.Calls, "Create")
	if f.OnCreate == nil {
		panic("fkContentFieldRepo: unexpected call to Create")
	}
	return f.OnCreate(ctx, cf)
}

func (f *fkContentFieldRepo) Update(ctx context.Context, id string, updates map[string]any) (*models.ContentField, error) {
	f.Calls = append(f.Calls, "Update")
	if f.OnUpdate == nil {
		panic("fkContentFieldRepo: unexpected call to Update")
	}
	return f.OnUpdate(ctx, id, updates)
}

func (f *fkContentFieldRepo) Delete(ctx context.Context, id string) error {
	f.Calls = append(f.Calls, "Delete")
	if f.OnDelete == nil {
		panic("fkContentFieldRepo: unexpected call to Delete")
	}
	return f.OnDelete(ctx, id)
}

// Called reports whether the named method was invoked at least once.
func (f *fkContentFieldRepo) Called(name string) bool {
	for _, c := range f.Calls {
		if c == name {
			return true
		}
	}
	return false
}

type fkContentItemRepo struct {
	Calls               []string
	OnListByContentType func(ctx context.Context, contentTypeID string) ([]models.ContentItem, error)
	OnGetByID           func(ctx context.Context, id string) (*models.ContentItem, error)
	OnGetBySlug         func(ctx context.Context, contentTypeID string, slug string) (*models.ContentItem, error)
	OnListPublic        func(ctx context.Context, contentTypeID string, limit int, sortBy string, sortOrder string) ([]models.ContentItem, error)
	OnCreate            func(ctx context.Context, ci *models.ContentItem, fieldValues []models.ContentFieldValue) (*models.ContentItem, error)
	OnUpdate            func(ctx context.Context, id string, updates map[string]any, fieldValues []models.ContentFieldValue) (*models.ContentItem, error)
	OnDelete            func(ctx context.Context, id string) error
}

func (f *fkContentItemRepo) ListByContentType(ctx context.Context, contentTypeID string) ([]models.ContentItem, error) {
	f.Calls = append(f.Calls, "ListByContentType")
	if f.OnListByContentType == nil {
		panic("fkContentItemRepo: unexpected call to ListByContentType")
	}
	return f.OnListByContentType(ctx, contentTypeID)
}

func (f *fkContentItemRepo) GetByID(ctx context.Context, id string) (*models.ContentItem, error) {
	f.Calls = append(f.Calls, "GetByID")
	if f.OnGetByID == nil {
		panic("fkContentItemRepo: unexpected call to GetByID")
	}
	return f.OnGetByID(ctx, id)
}

func (f *fkContentItemRepo) GetBySlug(ctx context.Context, contentTypeID string, slug string) (*models.ContentItem, error) {
	f.Calls = append(f.Calls, "GetBySlug")
	if f.OnGetBySlug == nil {
		panic("fkContentItemRepo: unexpected call to GetBySlug")
	}
	return f.OnGetBySlug(ctx, contentTypeID, slug)
}

func (f *fkContentItemRepo) ListPublic(ctx context.Context, contentTypeID string, limit int, sortBy string, sortOrder string) ([]models.ContentItem, error) {
	f.Calls = append(f.Calls, "ListPublic")
	if f.OnListPublic == nil {
		panic("fkContentItemRepo: unexpected call to ListPublic")
	}
	return f.OnListPublic(ctx, contentTypeID, limit, sortBy, sortOrder)
}

func (f *fkContentItemRepo) Create(ctx context.Context, ci *models.ContentItem, fieldValues []models.ContentFieldValue) (*models.ContentItem, error) {
	f.Calls = append(f.Calls, "Create")
	if f.OnCreate == nil {
		panic("fkContentItemRepo: unexpected call to Create")
	}
	return f.OnCreate(ctx, ci, fieldValues)
}

func (f *fkContentItemRepo) Update(ctx context.Context, id string, updates map[string]any, fieldValues []models.ContentFieldValue) (*models.ContentItem, error) {
	f.Calls = append(f.Calls, "Update")
	if f.OnUpdate == nil {
		panic("fkContentItemRepo: unexpected call to Update")
	}
	return f.OnUpdate(ctx, id, updates, fieldValues)
}

func (f *fkContentItemRepo) Delete(ctx context.Context, id string) error {
	f.Calls = append(f.Calls, "Delete")
	if f.OnDelete == nil {
		panic("fkContentItemRepo: unexpected call to Delete")
	}
	return f.OnDelete(ctx, id)
}

// Called reports whether the named method was invoked at least once.
func (f *fkContentItemRepo) Called(name string) bool {
	for _, c := range f.Calls {
		if c == name {
			return true
		}
	}
	return false
}

type fkImageRepo struct {
	Calls           []string
	OnCreate        func(ctx context.Context, image models.Image) (*models.Image, error)
	OnListByUser    func(ctx context.Context, userID string) ([]models.Image, error)
	OnGetByID       func(ctx context.Context, imageID string, userID string) (*models.Image, error)
	OnGetPublicByID func(ctx context.Context, imageID string) (*models.Image, error)
	OnDelete        func(ctx context.Context, imageID string, userID string) error
	OnSoftDelete    func(ctx context.Context, imageID string, userID string) error
	OnListAll       func(ctx context.Context, limit int, offset int) ([]models.Image, error)
}

func (f *fkImageRepo) Create(ctx context.Context, image models.Image) (*models.Image, error) {
	f.Calls = append(f.Calls, "Create")
	if f.OnCreate == nil {
		panic("fkImageRepo: unexpected call to Create")
	}
	return f.OnCreate(ctx, image)
}

func (f *fkImageRepo) ListByUser(ctx context.Context, userID string) ([]models.Image, error) {
	f.Calls = append(f.Calls, "ListByUser")
	if f.OnListByUser == nil {
		panic("fkImageRepo: unexpected call to ListByUser")
	}
	return f.OnListByUser(ctx, userID)
}

func (f *fkImageRepo) GetByID(ctx context.Context, imageID string, userID string) (*models.Image, error) {
	f.Calls = append(f.Calls, "GetByID")
	if f.OnGetByID == nil {
		panic("fkImageRepo: unexpected call to GetByID")
	}
	return f.OnGetByID(ctx, imageID, userID)
}

func (f *fkImageRepo) GetPublicByID(ctx context.Context, imageID string) (*models.Image, error) {
	f.Calls = append(f.Calls, "GetPublicByID")
	if f.OnGetPublicByID == nil {
		panic("fkImageRepo: unexpected call to GetPublicByID")
	}
	return f.OnGetPublicByID(ctx, imageID)
}

func (f *fkImageRepo) Delete(ctx context.Context, imageID string, userID string) error {
	f.Calls = append(f.Calls, "Delete")
	if f.OnDelete == nil {
		panic("fkImageRepo: unexpected call to Delete")
	}
	return f.OnDelete(ctx, imageID, userID)
}

func (f *fkImageRepo) SoftDelete(ctx context.Context, imageID string, userID string) error {
	f.Calls = append(f.Calls, "SoftDelete")
	if f.OnSoftDelete == nil {
		panic("fkImageRepo: unexpected call to SoftDelete")
	}
	return f.OnSoftDelete(ctx, imageID, userID)
}

func (f *fkImageRepo) ListAll(ctx context.Context, limit int, offset int) ([]models.Image, error) {
	f.Calls = append(f.Calls, "ListAll")
	if f.OnListAll == nil {
		panic("fkImageRepo: unexpected call to ListAll")
	}
	return f.OnListAll(ctx, limit, offset)
}

// Called reports whether the named method was invoked at least once.
func (f *fkImageRepo) Called(name string) bool {
	for _, c := range f.Calls {
		if c == name {
			return true
		}
	}
	return false
}

type fkCustomElementRepo struct {
	Calls        []string
	OnList       func(ctx context.Context, userID string, isPublic *bool) ([]models.CustomElement, error)
	OnGetByID    func(ctx context.Context, id string, userID string) (*models.CustomElement, error)
	OnCreate     func(ctx context.Context, customElement *models.CustomElement) (*models.CustomElement, error)
	OnUpdate     func(ctx context.Context, id string, userID string, updates map[string]any) (*models.CustomElement, error)
	OnDelete     func(ctx context.Context, id string, userID string) error
	OnListPublic func(ctx context.Context, category *string, limit int, offset int) ([]models.CustomElement, error)
	OnDuplicate  func(ctx context.Context, id string, userID string, newName string) (*models.CustomElement, error)
}

func (f *fkCustomElementRepo) List(ctx context.Context, userID string, isPublic *bool) ([]models.CustomElement, error) {
	f.Calls = append(f.Calls, "List")
	if f.OnList == nil {
		panic("fkCustomElementRepo: unexpected call to List")
	}
	return f.OnList(ctx, userID, isPublic)
}

func (f *fkCustomElementRepo) GetByID(ctx context.Context, id string, userID string) (*models.CustomElement, error) {
	f.Calls = append(f.Calls, "GetByID")
	if f.OnGetByID == nil {
		panic("fkCustomElementRepo: unexpected call to GetByID")
	}
	return f.OnGetByID(ctx, id, userID)
}

func (f *fkCustomElementRepo) Create(ctx context.Context, customElement *models.CustomElement) (*models.CustomElement, error) {
	f.Calls = append(f.Calls, "Create")
	if f.OnCreate == nil {
		panic("fkCustomElementRepo: unexpected call to Create")
	}
	return f.OnCreate(ctx, customElement)
}

func (f *fkCustomElementRepo) Update(ctx context.Context, id string, userID string, updates map[string]any) (*models.CustomElement, error) {
	f.Calls = append(f.Calls, "Update")
	if f.OnUpdate == nil {
		panic("fkCustomElementRepo: unexpected call to Update")
	}
	return f.OnUpdate(ctx, id, userID, updates)
}

func (f *fkCustomElementRepo) Delete(ctx context.Context, id string, userID string) error {
	f.Calls = append(f.Calls, "Delete")
	if f.OnDelete == nil {
		panic("fkCustomElementRepo: unexpected call to Delete")
	}
	return f.OnDelete(ctx, id, userID)
}

func (f *fkCustomElementRepo) ListPublic(ctx context.Context, category *string, limit int, offset int) ([]models.CustomElement, error) {
	f.Calls = append(f.Calls, "ListPublic")
	if f.OnListPublic == nil {
		panic("fkCustomElementRepo: unexpected call to ListPublic")
	}
	return f.OnListPublic(ctx, category, limit, offset)
}

func (f *fkCustomElementRepo) Duplicate(ctx context.Context, id string, userID string, newName string) (*models.CustomElement, error) {
	f.Calls = append(f.Calls, "Duplicate")
	if f.OnDuplicate == nil {
		panic("fkCustomElementRepo: unexpected call to Duplicate")
	}
	return f.OnDuplicate(ctx, id, userID, newName)
}

// Called reports whether the named method was invoked at least once.
func (f *fkCustomElementRepo) Called(name string) bool {
	for _, c := range f.Calls {
		if c == name {
			return true
		}
	}
	return false
}

type fkCustomElementTypeRepo struct {
	Calls       []string
	OnList      func(ctx context.Context) ([]models.CustomElementType, error)
	OnGetByID   func(ctx context.Context, id string) (*models.CustomElementType, error)
	OnGetByName func(ctx context.Context, name string) (*models.CustomElementType, error)
	OnCreate    func(ctx context.Context, customElementType *models.CustomElementType) (*models.CustomElementType, error)
	OnUpdate    func(ctx context.Context, id string, updates map[string]any) (*models.CustomElementType, error)
	OnDelete    func(ctx context.Context, id string) error
}

func (f *fkCustomElementTypeRepo) List(ctx context.Context) ([]models.CustomElementType, error) {
	f.Calls = append(f.Calls, "List")
	if f.OnList == nil {
		panic("fkCustomElementTypeRepo: unexpected call to List")
	}
	return f.OnList(ctx)
}

func (f *fkCustomElementTypeRepo) GetByID(ctx context.Context, id string) (*models.CustomElementType, error) {
	f.Calls = append(f.Calls, "GetByID")
	if f.OnGetByID == nil {
		panic("fkCustomElementTypeRepo: unexpected call to GetByID")
	}
	return f.OnGetByID(ctx, id)
}

func (f *fkCustomElementTypeRepo) GetByName(ctx context.Context, name string) (*models.CustomElementType, error) {
	f.Calls = append(f.Calls, "GetByName")
	if f.OnGetByName == nil {
		panic("fkCustomElementTypeRepo: unexpected call to GetByName")
	}
	return f.OnGetByName(ctx, name)
}

func (f *fkCustomElementTypeRepo) Create(ctx context.Context, customElementType *models.CustomElementType) (*models.CustomElementType, error) {
	f.Calls = append(f.Calls, "Create")
	if f.OnCreate == nil {
		panic("fkCustomElementTypeRepo: unexpected call to Create")
	}
	return f.OnCreate(ctx, customElementType)
}

func (f *fkCustomElementTypeRepo) Update(ctx context.Context, id string, updates map[string]any) (*models.CustomElementType, error) {
	f.Calls = append(f.Calls, "Update")
	if f.OnUpdate == nil {
		panic("fkCustomElementTypeRepo: unexpected call to Update")
	}
	return f.OnUpdate(ctx, id, updates)
}

func (f *fkCustomElementTypeRepo) Delete(ctx context.Context, id string) error {
	f.Calls = append(f.Calls, "Delete")
	if f.OnDelete == nil {
		panic("fkCustomElementTypeRepo: unexpected call to Delete")
	}
	return f.OnDelete(ctx, id)
}

// Called reports whether the named method was invoked at least once.
func (f *fkCustomElementTypeRepo) Called(name string) bool {
	for _, c := range f.Calls {
		if c == name {
			return true
		}
	}
	return false
}

type fkInvitationRepo struct {
	Calls                  []string
	OnCreate               func(ctx context.Context, invitation *models.Invitation) (*models.Invitation, error)
	OnListByProject        func(ctx context.Context, projectID string) ([]models.Invitation, error)
	OnGetByID              func(ctx context.Context, id string) (*models.Invitation, error)
	OnGetByToken           func(ctx context.Context, token string) (*models.Invitation, error)
	OnAccept               func(ctx context.Context, token string, userID string) error
	OnDelete               func(ctx context.Context, id string) error
	OnUpdateStatus         func(ctx context.Context, id string, status models.InvitationStatus) error
	OnCancel               func(ctx context.Context, id string) error
	OnListPendingByProject func(ctx context.Context, projectID string) ([]models.Invitation, error)
}

func (f *fkInvitationRepo) Create(ctx context.Context, invitation *models.Invitation) (*models.Invitation, error) {
	f.Calls = append(f.Calls, "Create")
	if f.OnCreate == nil {
		panic("fkInvitationRepo: unexpected call to Create")
	}
	return f.OnCreate(ctx, invitation)
}

func (f *fkInvitationRepo) ListByProject(ctx context.Context, projectID string) ([]models.Invitation, error) {
	f.Calls = append(f.Calls, "ListByProject")
	if f.OnListByProject == nil {
		panic("fkInvitationRepo: unexpected call to ListByProject")
	}
	return f.OnListByProject(ctx, projectID)
}

func (f *fkInvitationRepo) GetByID(ctx context.Context, id string) (*models.Invitation, error) {
	f.Calls = append(f.Calls, "GetByID")
	if f.OnGetByID == nil {
		panic("fkInvitationRepo: unexpected call to GetByID")
	}
	return f.OnGetByID(ctx, id)
}

func (f *fkInvitationRepo) GetByToken(ctx context.Context, token string) (*models.Invitation, error) {
	f.Calls = append(f.Calls, "GetByToken")
	if f.OnGetByToken == nil {
		panic("fkInvitationRepo: unexpected call to GetByToken")
	}
	return f.OnGetByToken(ctx, token)
}

func (f *fkInvitationRepo) Accept(ctx context.Context, token string, userID string) error {
	f.Calls = append(f.Calls, "Accept")
	if f.OnAccept == nil {
		panic("fkInvitationRepo: unexpected call to Accept")
	}
	return f.OnAccept(ctx, token, userID)
}

func (f *fkInvitationRepo) Delete(ctx context.Context, id string) error {
	f.Calls = append(f.Calls, "Delete")
	if f.OnDelete == nil {
		panic("fkInvitationRepo: unexpected call to Delete")
	}
	return f.OnDelete(ctx, id)
}

func (f *fkInvitationRepo) UpdateStatus(ctx context.Context, id string, status models.InvitationStatus) error {
	f.Calls = append(f.Calls, "UpdateStatus")
	if f.OnUpdateStatus == nil {
		panic("fkInvitationRepo: unexpected call to UpdateStatus")
	}
	return f.OnUpdateStatus(ctx, id, status)
}

func (f *fkInvitationRepo) Cancel(ctx context.Context, id string) error {
	f.Calls = append(f.Calls, "Cancel")
	if f.OnCancel == nil {
		panic("fkInvitationRepo: unexpected call to Cancel")
	}
	return f.OnCancel(ctx, id)
}

func (f *fkInvitationRepo) ListPendingByProject(ctx context.Context, projectID string) ([]models.Invitation, error) {
	f.Calls = append(f.Calls, "ListPendingByProject")
	if f.OnListPendingByProject == nil {
		panic("fkInvitationRepo: unexpected call to ListPendingByProject")
	}
	return f.OnListPendingByProject(ctx, projectID)
}

// Called reports whether the named method was invoked at least once.
func (f *fkInvitationRepo) Called(name string) bool {
	for _, c := range f.Calls {
		if c == name {
			return true
		}
	}
	return false
}

type fkCollaboratorRepo struct {
	Calls                    []string
	OnCreate                 func(ctx context.Context, collaborator *models.Collaborator) (*models.Collaborator, error)
	OnListByProject          func(ctx context.Context, projectID string) ([]models.Collaborator, error)
	OnGetByID                func(ctx context.Context, id string) (*models.Collaborator, error)
	OnUpdateRole             func(ctx context.Context, id string, role models.CollaboratorRole) error
	OnDelete                 func(ctx context.Context, id string) error
	OnIsCollaborator         func(ctx context.Context, projectID string, userID string) (bool, error)
	OnGetByProjectAndUser    func(ctx context.Context, projectID string, userID string) (*models.Collaborator, error)
	OnDeleteByProjectAndUser func(ctx context.Context, projectID string, userID string) (bool, error)
}

func (f *fkCollaboratorRepo) Create(ctx context.Context, collaborator *models.Collaborator) (*models.Collaborator, error) {
	f.Calls = append(f.Calls, "Create")
	if f.OnCreate == nil {
		panic("fkCollaboratorRepo: unexpected call to Create")
	}
	return f.OnCreate(ctx, collaborator)
}

func (f *fkCollaboratorRepo) ListByProject(ctx context.Context, projectID string) ([]models.Collaborator, error) {
	f.Calls = append(f.Calls, "ListByProject")
	if f.OnListByProject == nil {
		panic("fkCollaboratorRepo: unexpected call to ListByProject")
	}
	return f.OnListByProject(ctx, projectID)
}

func (f *fkCollaboratorRepo) GetByID(ctx context.Context, id string) (*models.Collaborator, error) {
	f.Calls = append(f.Calls, "GetByID")
	if f.OnGetByID == nil {
		panic("fkCollaboratorRepo: unexpected call to GetByID")
	}
	return f.OnGetByID(ctx, id)
}

func (f *fkCollaboratorRepo) UpdateRole(ctx context.Context, id string, role models.CollaboratorRole) error {
	f.Calls = append(f.Calls, "UpdateRole")
	if f.OnUpdateRole == nil {
		panic("fkCollaboratorRepo: unexpected call to UpdateRole")
	}
	return f.OnUpdateRole(ctx, id, role)
}

func (f *fkCollaboratorRepo) Delete(ctx context.Context, id string) error {
	f.Calls = append(f.Calls, "Delete")
	if f.OnDelete == nil {
		panic("fkCollaboratorRepo: unexpected call to Delete")
	}
	return f.OnDelete(ctx, id)
}

func (f *fkCollaboratorRepo) IsCollaborator(ctx context.Context, projectID string, userID string) (bool, error) {
	f.Calls = append(f.Calls, "IsCollaborator")
	if f.OnIsCollaborator == nil {
		panic("fkCollaboratorRepo: unexpected call to IsCollaborator")
	}
	return f.OnIsCollaborator(ctx, projectID, userID)
}

func (f *fkCollaboratorRepo) GetByProjectAndUser(ctx context.Context, projectID string, userID string) (*models.Collaborator, error) {
	f.Calls = append(f.Calls, "GetByProjectAndUser")
	if f.OnGetByProjectAndUser == nil {
		return nil, nil
	}
	return f.OnGetByProjectAndUser(ctx, projectID, userID)
}

func (f *fkCollaboratorRepo) DeleteByProjectAndUser(ctx context.Context, projectID string, userID string) (bool, error) {
	f.Calls = append(f.Calls, "DeleteByProjectAndUser")
	if f.OnDeleteByProjectAndUser == nil {
		panic("fkCollaboratorRepo: unexpected call to DeleteByProjectAndUser")
	}
	return f.OnDeleteByProjectAndUser(ctx, projectID, userID)
}

// Called reports whether the named method was invoked at least once.
func (f *fkCollaboratorRepo) Called(name string) bool {
	for _, c := range f.Calls {
		if c == name {
			return true
		}
	}
	return false
}

type fkEventWorkflowRepo struct {
	Calls                       []string
	OnCreate                    func(ctx context.Context, workflow *models.EventWorkflow) (*models.EventWorkflow, error)
	OnGetByID                   func(ctx context.Context, id string) (*models.EventWorkflow, error)
	OnListByProject             func(ctx context.Context, projectID string) ([]models.EventWorkflow, error)
	OnListByProjectWithElements func(ctx context.Context, projectID string) ([]models.EventWorkflow, error)
	OnListEnabledByProject      func(ctx context.Context, projectID string) ([]models.EventWorkflow, error)
	OnListByName                func(ctx context.Context, projectID string, name string) ([]models.EventWorkflow, error)
	OnUpdate                    func(ctx context.Context, id string, workflow *models.EventWorkflow) (*models.EventWorkflow, error)
	OnSetEnabled                func(ctx context.Context, id string, enabled bool) error
	OnDelete                    func(ctx context.Context, id string) error
	OnDeleteByProject           func(ctx context.Context, projectID string) error
	OnCountByProject            func(ctx context.Context, projectID string) (int64, error)
	OnExistsByName              func(ctx context.Context, projectID string, name string, excludeID string) (bool, error)
	OnListWithFilters           func(ctx context.Context, projectID string, enabled *bool, searchName string) ([]models.EventWorkflow, error)
}

func (f *fkEventWorkflowRepo) Create(ctx context.Context, workflow *models.EventWorkflow) (*models.EventWorkflow, error) {
	f.Calls = append(f.Calls, "Create")
	if f.OnCreate == nil {
		panic("fkEventWorkflowRepo: unexpected call to Create")
	}
	return f.OnCreate(ctx, workflow)
}

func (f *fkEventWorkflowRepo) GetByID(ctx context.Context, id string) (*models.EventWorkflow, error) {
	f.Calls = append(f.Calls, "GetByID")
	if f.OnGetByID == nil {
		panic("fkEventWorkflowRepo: unexpected call to GetByID")
	}
	return f.OnGetByID(ctx, id)
}

func (f *fkEventWorkflowRepo) ListByProject(ctx context.Context, projectID string) ([]models.EventWorkflow, error) {
	f.Calls = append(f.Calls, "ListByProject")
	if f.OnListByProject == nil {
		panic("fkEventWorkflowRepo: unexpected call to ListByProject")
	}
	return f.OnListByProject(ctx, projectID)
}

func (f *fkEventWorkflowRepo) ListByProjectWithElements(ctx context.Context, projectID string) ([]models.EventWorkflow, error) {
	f.Calls = append(f.Calls, "ListByProjectWithElements")
	if f.OnListByProjectWithElements == nil {
		panic("fkEventWorkflowRepo: unexpected call to ListByProjectWithElements")
	}
	return f.OnListByProjectWithElements(ctx, projectID)
}

func (f *fkEventWorkflowRepo) ListEnabledByProject(ctx context.Context, projectID string) ([]models.EventWorkflow, error) {
	f.Calls = append(f.Calls, "ListEnabledByProject")
	if f.OnListEnabledByProject == nil {
		panic("fkEventWorkflowRepo: unexpected call to ListEnabledByProject")
	}
	return f.OnListEnabledByProject(ctx, projectID)
}

func (f *fkEventWorkflowRepo) ListByName(ctx context.Context, projectID string, name string) ([]models.EventWorkflow, error) {
	f.Calls = append(f.Calls, "ListByName")
	if f.OnListByName == nil {
		panic("fkEventWorkflowRepo: unexpected call to ListByName")
	}
	return f.OnListByName(ctx, projectID, name)
}

func (f *fkEventWorkflowRepo) Update(ctx context.Context, id string, workflow *models.EventWorkflow) (*models.EventWorkflow, error) {
	f.Calls = append(f.Calls, "Update")
	if f.OnUpdate == nil {
		panic("fkEventWorkflowRepo: unexpected call to Update")
	}
	return f.OnUpdate(ctx, id, workflow)
}

func (f *fkEventWorkflowRepo) SetEnabled(ctx context.Context, id string, enabled bool) error {
	f.Calls = append(f.Calls, "SetEnabled")
	if f.OnSetEnabled == nil {
		panic("fkEventWorkflowRepo: unexpected call to SetEnabled")
	}
	return f.OnSetEnabled(ctx, id, enabled)
}

func (f *fkEventWorkflowRepo) Delete(ctx context.Context, id string) error {
	f.Calls = append(f.Calls, "Delete")
	if f.OnDelete == nil {
		panic("fkEventWorkflowRepo: unexpected call to Delete")
	}
	return f.OnDelete(ctx, id)
}

func (f *fkEventWorkflowRepo) DeleteByProject(ctx context.Context, projectID string) error {
	f.Calls = append(f.Calls, "DeleteByProject")
	if f.OnDeleteByProject == nil {
		panic("fkEventWorkflowRepo: unexpected call to DeleteByProject")
	}
	return f.OnDeleteByProject(ctx, projectID)
}

func (f *fkEventWorkflowRepo) CountByProject(ctx context.Context, projectID string) (int64, error) {
	f.Calls = append(f.Calls, "CountByProject")
	if f.OnCountByProject == nil {
		panic("fkEventWorkflowRepo: unexpected call to CountByProject")
	}
	return f.OnCountByProject(ctx, projectID)
}

func (f *fkEventWorkflowRepo) ExistsByName(ctx context.Context, projectID string, name string, excludeID string) (bool, error) {
	f.Calls = append(f.Calls, "ExistsByName")
	if f.OnExistsByName == nil {
		panic("fkEventWorkflowRepo: unexpected call to ExistsByName")
	}
	return f.OnExistsByName(ctx, projectID, name, excludeID)
}

func (f *fkEventWorkflowRepo) ListWithFilters(ctx context.Context, projectID string, enabled *bool, searchName string) ([]models.EventWorkflow, error) {
	f.Calls = append(f.Calls, "ListWithFilters")
	if f.OnListWithFilters == nil {
		panic("fkEventWorkflowRepo: unexpected call to ListWithFilters")
	}
	return f.OnListWithFilters(ctx, projectID, enabled, searchName)
}

// Called reports whether the named method was invoked at least once.
func (f *fkEventWorkflowRepo) Called(name string) bool {
	for _, c := range f.Calls {
		if c == name {
			return true
		}
	}
	return false
}

type fkElementEventWorkflowRepo struct {
	Calls              []string
	OnCreate           func(ctx context.Context, eew *models.ElementEventWorkflow) (*models.ElementEventWorkflow, error)
	OnGetByID          func(ctx context.Context, id string) (*models.ElementEventWorkflow, error)
	OnListAll          func(ctx context.Context) ([]models.ElementEventWorkflow, error)
	OnListByElement    func(ctx context.Context, elementID string) ([]models.ElementEventWorkflow, error)
	OnListByWorkflow   func(ctx context.Context, workflowID string) ([]models.ElementEventWorkflow, error)
	OnListByEvent      func(ctx context.Context, eventName string) ([]models.ElementEventWorkflow, error)
	OnListByFilters    func(ctx context.Context, elementID string, workflowID string, eventName string) ([]models.ElementEventWorkflow, error)
	OnUpdate           func(ctx context.Context, id string, eew *models.ElementEventWorkflow) (*models.ElementEventWorkflow, error)
	OnDelete           func(ctx context.Context, id string) error
	OnDeleteByElement  func(ctx context.Context, elementID string) error
	OnDeleteByWorkflow func(ctx context.Context, workflowID string) error
	OnListByPage       func(ctx context.Context, pageID string) ([]models.ElementEventWorkflow, error)
	OnIsLinked         func(ctx context.Context, elementID string, workflowID string, eventName string) (bool, error)
}

func (f *fkElementEventWorkflowRepo) Create(ctx context.Context, eew *models.ElementEventWorkflow) (*models.ElementEventWorkflow, error) {
	f.Calls = append(f.Calls, "Create")
	if f.OnCreate == nil {
		panic("fkElementEventWorkflowRepo: unexpected call to Create")
	}
	return f.OnCreate(ctx, eew)
}

func (f *fkElementEventWorkflowRepo) GetByID(ctx context.Context, id string) (*models.ElementEventWorkflow, error) {
	f.Calls = append(f.Calls, "GetByID")
	if f.OnGetByID == nil {
		panic("fkElementEventWorkflowRepo: unexpected call to GetByID")
	}
	return f.OnGetByID(ctx, id)
}

func (f *fkElementEventWorkflowRepo) ListAll(ctx context.Context) ([]models.ElementEventWorkflow, error) {
	f.Calls = append(f.Calls, "ListAll")
	if f.OnListAll == nil {
		panic("fkElementEventWorkflowRepo: unexpected call to ListAll")
	}
	return f.OnListAll(ctx)
}

func (f *fkElementEventWorkflowRepo) ListByElement(ctx context.Context, elementID string) ([]models.ElementEventWorkflow, error) {
	f.Calls = append(f.Calls, "ListByElement")
	if f.OnListByElement == nil {
		panic("fkElementEventWorkflowRepo: unexpected call to ListByElement")
	}
	return f.OnListByElement(ctx, elementID)
}

func (f *fkElementEventWorkflowRepo) ListByWorkflow(ctx context.Context, workflowID string) ([]models.ElementEventWorkflow, error) {
	f.Calls = append(f.Calls, "ListByWorkflow")
	if f.OnListByWorkflow == nil {
		panic("fkElementEventWorkflowRepo: unexpected call to ListByWorkflow")
	}
	return f.OnListByWorkflow(ctx, workflowID)
}

func (f *fkElementEventWorkflowRepo) ListByEvent(ctx context.Context, eventName string) ([]models.ElementEventWorkflow, error) {
	f.Calls = append(f.Calls, "ListByEvent")
	if f.OnListByEvent == nil {
		panic("fkElementEventWorkflowRepo: unexpected call to ListByEvent")
	}
	return f.OnListByEvent(ctx, eventName)
}

func (f *fkElementEventWorkflowRepo) ListByFilters(ctx context.Context, elementID string, workflowID string, eventName string) ([]models.ElementEventWorkflow, error) {
	f.Calls = append(f.Calls, "ListByFilters")
	if f.OnListByFilters == nil {
		panic("fkElementEventWorkflowRepo: unexpected call to ListByFilters")
	}
	return f.OnListByFilters(ctx, elementID, workflowID, eventName)
}

func (f *fkElementEventWorkflowRepo) Update(ctx context.Context, id string, eew *models.ElementEventWorkflow) (*models.ElementEventWorkflow, error) {
	f.Calls = append(f.Calls, "Update")
	if f.OnUpdate == nil {
		panic("fkElementEventWorkflowRepo: unexpected call to Update")
	}
	return f.OnUpdate(ctx, id, eew)
}

func (f *fkElementEventWorkflowRepo) Delete(ctx context.Context, id string) error {
	f.Calls = append(f.Calls, "Delete")
	if f.OnDelete == nil {
		panic("fkElementEventWorkflowRepo: unexpected call to Delete")
	}
	return f.OnDelete(ctx, id)
}

func (f *fkElementEventWorkflowRepo) DeleteByElement(ctx context.Context, elementID string) error {
	f.Calls = append(f.Calls, "DeleteByElement")
	if f.OnDeleteByElement == nil {
		panic("fkElementEventWorkflowRepo: unexpected call to DeleteByElement")
	}
	return f.OnDeleteByElement(ctx, elementID)
}

func (f *fkElementEventWorkflowRepo) DeleteByWorkflow(ctx context.Context, workflowID string) error {
	f.Calls = append(f.Calls, "DeleteByWorkflow")
	if f.OnDeleteByWorkflow == nil {
		panic("fkElementEventWorkflowRepo: unexpected call to DeleteByWorkflow")
	}
	return f.OnDeleteByWorkflow(ctx, workflowID)
}

func (f *fkElementEventWorkflowRepo) ListByPage(ctx context.Context, pageID string) ([]models.ElementEventWorkflow, error) {
	f.Calls = append(f.Calls, "ListByPage")
	if f.OnListByPage == nil {
		panic("fkElementEventWorkflowRepo: unexpected call to ListByPage")
	}
	return f.OnListByPage(ctx, pageID)
}

func (f *fkElementEventWorkflowRepo) IsLinked(ctx context.Context, elementID string, workflowID string, eventName string) (bool, error) {
	f.Calls = append(f.Calls, "IsLinked")
	if f.OnIsLinked == nil {
		panic("fkElementEventWorkflowRepo: unexpected call to IsLinked")
	}
	return f.OnIsLinked(ctx, elementID, workflowID, eventName)
}

// Called reports whether the named method was invoked at least once.
func (f *fkElementEventWorkflowRepo) Called(name string) bool {
	for _, c := range f.Calls {
		if c == name {
			return true
		}
	}
	return false
}

type fkUserRepo struct {
	Calls           []string
	OnGetByID       func(ctx context.Context, userID string) (*models.User, error)
	OnGetByEmail    func(ctx context.Context, email string) (*models.User, error)
	OnGetByUsername func(ctx context.Context, username string) (*models.User, error)
	OnSearch        func(ctx context.Context, query string) ([]models.User, error)
}

func (f *fkUserRepo) GetByID(ctx context.Context, userID string) (*models.User, error) {
	f.Calls = append(f.Calls, "GetByID")
	if f.OnGetByID == nil {
		panic("fkUserRepo: unexpected call to GetByID")
	}
	return f.OnGetByID(ctx, userID)
}

func (f *fkUserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	f.Calls = append(f.Calls, "GetByEmail")
	if f.OnGetByEmail == nil {
		panic("fkUserRepo: unexpected call to GetByEmail")
	}
	return f.OnGetByEmail(ctx, email)
}

func (f *fkUserRepo) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	f.Calls = append(f.Calls, "GetByUsername")
	if f.OnGetByUsername == nil {
		panic("fkUserRepo: unexpected call to GetByUsername")
	}
	return f.OnGetByUsername(ctx, username)
}

func (f *fkUserRepo) Search(ctx context.Context, query string) ([]models.User, error) {
	f.Calls = append(f.Calls, "Search")
	if f.OnSearch == nil {
		panic("fkUserRepo: unexpected call to Search")
	}
	return f.OnSearch(ctx, query)
}

// Called reports whether the named method was invoked at least once.
func (f *fkUserRepo) Called(name string) bool {
	for _, c := range f.Calls {
		if c == name {
			return true
		}
	}
	return false
}
