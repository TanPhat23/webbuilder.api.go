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
	Calls                       []string
	OnGetElements               func(ctx context.Context, projectID string, pageID ...string) ([]models.EditorElement, error)
	OnReplaceElements           func(ctx context.Context, projectID string, elements []models.EditorElement) error
	OnGetElementByID            func(ctx context.Context, elementID string) (*models.Element, error)
	OnGetElementsByPageID       func(ctx context.Context, pageID string) ([]models.Element, error)
	OnGetElementsByPageIds      func(ctx context.Context, projectID string, pageIDs []string) ([]models.EditorElement, error)
	OnGetChildElements          func(ctx context.Context, parentID string) ([]models.Element, error)
	OnGetRootElements           func(ctx context.Context, projectID string) ([]models.Element, error)
	OnCreateElement             func(ctx context.Context, element *models.Element) error
	OnUpdateElement             func(ctx context.Context, element *models.Element) error
	OnUpdateEventWorkflows      func(ctx context.Context, elementID string, workflows []byte) error
	OnDeleteElementByID         func(ctx context.Context, elementID string) error
	OnDeleteElementsByPageID    func(ctx context.Context, pageID string) error
	OnDeleteElementsByProjectID func(ctx context.Context, projectID string) error
	OnCountElementsByProjectID  func(ctx context.Context, projectID string) (int64, error)
	OnGetElementWithRelations   func(ctx context.Context, elementID string) (*models.Element, error)
	OnGetElementsByIDs          func(ctx context.Context, elementIDs []string) ([]models.Element, error)
}

func (f *fkElementRepo) GetElements(ctx context.Context, projectID string, pageID ...string) ([]models.EditorElement, error) {
	f.Calls = append(f.Calls, "GetElements")
	if f.OnGetElements == nil {
		panic("fkElementRepo: unexpected call to GetElements")
	}
	return f.OnGetElements(ctx, projectID, pageID...)
}

func (f *fkElementRepo) ReplaceElements(ctx context.Context, projectID string, elements []models.EditorElement) error {
	f.Calls = append(f.Calls, "ReplaceElements")
	if f.OnReplaceElements == nil {
		panic("fkElementRepo: unexpected call to ReplaceElements")
	}
	return f.OnReplaceElements(ctx, projectID, elements)
}

func (f *fkElementRepo) GetElementByID(ctx context.Context, elementID string) (*models.Element, error) {
	f.Calls = append(f.Calls, "GetElementByID")
	if f.OnGetElementByID == nil {
		panic("fkElementRepo: unexpected call to GetElementByID")
	}
	return f.OnGetElementByID(ctx, elementID)
}

func (f *fkElementRepo) GetElementsByPageID(ctx context.Context, pageID string) ([]models.Element, error) {
	f.Calls = append(f.Calls, "GetElementsByPageID")
	if f.OnGetElementsByPageID == nil {
		panic("fkElementRepo: unexpected call to GetElementsByPageID")
	}
	return f.OnGetElementsByPageID(ctx, pageID)
}

func (f *fkElementRepo) GetElementsByPageIds(ctx context.Context, projectID string, pageIDs []string) ([]models.EditorElement, error) {
	f.Calls = append(f.Calls, "GetElementsByPageIds")
	if f.OnGetElementsByPageIds == nil {
		panic("fkElementRepo: unexpected call to GetElementsByPageIds")
	}
	return f.OnGetElementsByPageIds(ctx, projectID, pageIDs)
}

func (f *fkElementRepo) GetChildElements(ctx context.Context, parentID string) ([]models.Element, error) {
	f.Calls = append(f.Calls, "GetChildElements")
	if f.OnGetChildElements == nil {
		panic("fkElementRepo: unexpected call to GetChildElements")
	}
	return f.OnGetChildElements(ctx, parentID)
}

func (f *fkElementRepo) GetRootElements(ctx context.Context, projectID string) ([]models.Element, error) {
	f.Calls = append(f.Calls, "GetRootElements")
	if f.OnGetRootElements == nil {
		panic("fkElementRepo: unexpected call to GetRootElements")
	}
	return f.OnGetRootElements(ctx, projectID)
}

func (f *fkElementRepo) CreateElement(ctx context.Context, element *models.Element) error {
	f.Calls = append(f.Calls, "CreateElement")
	if f.OnCreateElement == nil {
		panic("fkElementRepo: unexpected call to CreateElement")
	}
	return f.OnCreateElement(ctx, element)
}

func (f *fkElementRepo) UpdateElement(ctx context.Context, element *models.Element) error {
	f.Calls = append(f.Calls, "UpdateElement")
	if f.OnUpdateElement == nil {
		panic("fkElementRepo: unexpected call to UpdateElement")
	}
	return f.OnUpdateElement(ctx, element)
}

func (f *fkElementRepo) UpdateEventWorkflows(ctx context.Context, elementID string, workflows []byte) error {
	f.Calls = append(f.Calls, "UpdateEventWorkflows")
	if f.OnUpdateEventWorkflows == nil {
		panic("fkElementRepo: unexpected call to UpdateEventWorkflows")
	}
	return f.OnUpdateEventWorkflows(ctx, elementID, workflows)
}

func (f *fkElementRepo) DeleteElementByID(ctx context.Context, elementID string) error {
	f.Calls = append(f.Calls, "DeleteElementByID")
	if f.OnDeleteElementByID == nil {
		panic("fkElementRepo: unexpected call to DeleteElementByID")
	}
	return f.OnDeleteElementByID(ctx, elementID)
}

func (f *fkElementRepo) DeleteElementsByPageID(ctx context.Context, pageID string) error {
	f.Calls = append(f.Calls, "DeleteElementsByPageID")
	if f.OnDeleteElementsByPageID == nil {
		panic("fkElementRepo: unexpected call to DeleteElementsByPageID")
	}
	return f.OnDeleteElementsByPageID(ctx, pageID)
}

func (f *fkElementRepo) DeleteElementsByProjectID(ctx context.Context, projectID string) error {
	f.Calls = append(f.Calls, "DeleteElementsByProjectID")
	if f.OnDeleteElementsByProjectID == nil {
		panic("fkElementRepo: unexpected call to DeleteElementsByProjectID")
	}
	return f.OnDeleteElementsByProjectID(ctx, projectID)
}

func (f *fkElementRepo) CountElementsByProjectID(ctx context.Context, projectID string) (int64, error) {
	f.Calls = append(f.Calls, "CountElementsByProjectID")
	if f.OnCountElementsByProjectID == nil {
		panic("fkElementRepo: unexpected call to CountElementsByProjectID")
	}
	return f.OnCountElementsByProjectID(ctx, projectID)
}

func (f *fkElementRepo) GetElementWithRelations(ctx context.Context, elementID string) (*models.Element, error) {
	f.Calls = append(f.Calls, "GetElementWithRelations")
	if f.OnGetElementWithRelations == nil {
		panic("fkElementRepo: unexpected call to GetElementWithRelations")
	}
	return f.OnGetElementWithRelations(ctx, elementID)
}

func (f *fkElementRepo) GetElementsByIDs(ctx context.Context, elementIDs []string) ([]models.Element, error) {
	f.Calls = append(f.Calls, "GetElementsByIDs")
	if f.OnGetElementsByIDs == nil {
		panic("fkElementRepo: unexpected call to GetElementsByIDs")
	}
	return f.OnGetElementsByIDs(ctx, elementIDs)
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
	Calls                              []string
	OnCreateElementComment             func(ctx context.Context, comment *models.ElementComment) (*models.ElementComment, error)
	OnGetElementCommentByID            func(ctx context.Context, id string) (*models.ElementComment, error)
	OnGetElementComments               func(ctx context.Context, elementID string, filter *models.ElementCommentFilter) ([]models.ElementComment, error)
	OnUpdateElementComment             func(ctx context.Context, id string, updates map[string]any) (*models.ElementComment, error)
	OnDeleteElementComment             func(ctx context.Context, id string) error
	OnGetElementCommentsByAuthorID     func(ctx context.Context, authorID string, limit int, offset int) ([]models.ElementComment, error)
	OnCountElementComments             func(ctx context.Context, elementID string) (int64, error)
	OnToggleResolvedStatus             func(ctx context.Context, id string) (*models.ElementComment, error)
	OnDeleteElementCommentsByElementID func(ctx context.Context, elementID string) error
	OnGetElementCommentsByProjectID    func(ctx context.Context, projectID string, limit int, offset int) ([]models.ElementComment, error)
	OnCountElementCommentsByProjectID  func(ctx context.Context, projectID string) (int64, error)
}

func (f *fkElementCommentRepo) CreateElementComment(ctx context.Context, comment *models.ElementComment) (*models.ElementComment, error) {
	f.Calls = append(f.Calls, "CreateElementComment")
	if f.OnCreateElementComment == nil {
		panic("fkElementCommentRepo: unexpected call to CreateElementComment")
	}
	return f.OnCreateElementComment(ctx, comment)
}

func (f *fkElementCommentRepo) GetElementCommentByID(ctx context.Context, id string) (*models.ElementComment, error) {
	f.Calls = append(f.Calls, "GetElementCommentByID")
	if f.OnGetElementCommentByID == nil {
		panic("fkElementCommentRepo: unexpected call to GetElementCommentByID")
	}
	return f.OnGetElementCommentByID(ctx, id)
}

func (f *fkElementCommentRepo) GetElementComments(ctx context.Context, elementID string, filter *models.ElementCommentFilter) ([]models.ElementComment, error) {
	f.Calls = append(f.Calls, "GetElementComments")
	if f.OnGetElementComments == nil {
		panic("fkElementCommentRepo: unexpected call to GetElementComments")
	}
	return f.OnGetElementComments(ctx, elementID, filter)
}

func (f *fkElementCommentRepo) UpdateElementComment(ctx context.Context, id string, updates map[string]any) (*models.ElementComment, error) {
	f.Calls = append(f.Calls, "UpdateElementComment")
	if f.OnUpdateElementComment == nil {
		panic("fkElementCommentRepo: unexpected call to UpdateElementComment")
	}
	return f.OnUpdateElementComment(ctx, id, updates)
}

func (f *fkElementCommentRepo) DeleteElementComment(ctx context.Context, id string) error {
	f.Calls = append(f.Calls, "DeleteElementComment")
	if f.OnDeleteElementComment == nil {
		panic("fkElementCommentRepo: unexpected call to DeleteElementComment")
	}
	return f.OnDeleteElementComment(ctx, id)
}

func (f *fkElementCommentRepo) GetElementCommentsByAuthorID(ctx context.Context, authorID string, limit int, offset int) ([]models.ElementComment, error) {
	f.Calls = append(f.Calls, "GetElementCommentsByAuthorID")
	if f.OnGetElementCommentsByAuthorID == nil {
		panic("fkElementCommentRepo: unexpected call to GetElementCommentsByAuthorID")
	}
	return f.OnGetElementCommentsByAuthorID(ctx, authorID, limit, offset)
}

func (f *fkElementCommentRepo) CountElementComments(ctx context.Context, elementID string) (int64, error) {
	f.Calls = append(f.Calls, "CountElementComments")
	if f.OnCountElementComments == nil {
		panic("fkElementCommentRepo: unexpected call to CountElementComments")
	}
	return f.OnCountElementComments(ctx, elementID)
}

func (f *fkElementCommentRepo) ToggleResolvedStatus(ctx context.Context, id string) (*models.ElementComment, error) {
	f.Calls = append(f.Calls, "ToggleResolvedStatus")
	if f.OnToggleResolvedStatus == nil {
		panic("fkElementCommentRepo: unexpected call to ToggleResolvedStatus")
	}
	return f.OnToggleResolvedStatus(ctx, id)
}

func (f *fkElementCommentRepo) DeleteElementCommentsByElementID(ctx context.Context, elementID string) error {
	f.Calls = append(f.Calls, "DeleteElementCommentsByElementID")
	if f.OnDeleteElementCommentsByElementID == nil {
		panic("fkElementCommentRepo: unexpected call to DeleteElementCommentsByElementID")
	}
	return f.OnDeleteElementCommentsByElementID(ctx, elementID)
}

func (f *fkElementCommentRepo) GetElementCommentsByProjectID(ctx context.Context, projectID string, limit int, offset int) ([]models.ElementComment, error) {
	f.Calls = append(f.Calls, "GetElementCommentsByProjectID")
	if f.OnGetElementCommentsByProjectID == nil {
		panic("fkElementCommentRepo: unexpected call to GetElementCommentsByProjectID")
	}
	return f.OnGetElementCommentsByProjectID(ctx, projectID, limit, offset)
}

func (f *fkElementCommentRepo) CountElementCommentsByProjectID(ctx context.Context, projectID string) (int64, error) {
	f.Calls = append(f.Calls, "CountElementCommentsByProjectID")
	if f.OnCountElementCommentsByProjectID == nil {
		panic("fkElementCommentRepo: unexpected call to CountElementCommentsByProjectID")
	}
	return f.OnCountElementCommentsByProjectID(ctx, projectID)
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
	Calls                     []string
	OnGetPublicProjectByID    func(ctx context.Context, projectID string) (*models.Project, error)
	OnGetProjectByID          func(ctx context.Context, projectID string, userID string) (*models.Project, error)
	OnGetProjectWithAccess    func(ctx context.Context, projectID string, userID string) (*models.Project, error)
	OnGetProjectsByUserID     func(ctx context.Context, userID string) ([]models.Project, error)
	OnGetCollaboratorProjects func(ctx context.Context, userID string) ([]models.Project, error)
	OnGetProjectPages         func(ctx context.Context, projectID string, userID string) ([]models.Page, error)
	OnCreateProject           func(ctx context.Context, project *models.Project) error
	OnCreateProjectWithPage   func(ctx context.Context, project *models.Project, page *models.Page) error
	OnGetProjectAccess        func(ctx context.Context, projectID string, userID string) (*models.ProjectAccess, error)
	OnUpdateProject           func(ctx context.Context, projectID string, userID string, updates map[string]any) (*models.Project, error)
	OnUpdateProjectFields     func(ctx context.Context, projectID string, userID string, updates map[string]any) (*models.Project, error)
	OnTransferOwnership       func(ctx context.Context, projectID string, currentOwnerID string, newOwnerID string) error
	OnDeleteProject           func(ctx context.Context, projectID string, userID string) error
	OnHardDeleteProject       func(ctx context.Context, projectID string, userID string) error
	OnRestoreProject          func(ctx context.Context, projectID string, userID string) error
	OnExistsForUser           func(ctx context.Context, projectID string, userID string) (bool, error)
	OnGetProjectWithLock      func(ctx context.Context, projectID string, userID string) (*models.Project, error)
}

func (f *fkProjectRepo) GetPublicProjectByID(ctx context.Context, projectID string) (*models.Project, error) {
	f.Calls = append(f.Calls, "GetPublicProjectByID")
	if f.OnGetPublicProjectByID == nil {
		panic("fkProjectRepo: unexpected call to GetPublicProjectByID")
	}
	return f.OnGetPublicProjectByID(ctx, projectID)
}

func (f *fkProjectRepo) GetProjectByID(ctx context.Context, projectID string, userID string) (*models.Project, error) {
	f.Calls = append(f.Calls, "GetProjectByID")
	if f.OnGetProjectByID == nil {
		panic("fkProjectRepo: unexpected call to GetProjectByID")
	}
	return f.OnGetProjectByID(ctx, projectID, userID)
}

func (f *fkProjectRepo) GetProjectWithAccess(ctx context.Context, projectID string, userID string) (*models.Project, error) {
	f.Calls = append(f.Calls, "GetProjectWithAccess")
	if f.OnGetProjectWithAccess == nil {
		panic("fkProjectRepo: unexpected call to GetProjectWithAccess")
	}
	return f.OnGetProjectWithAccess(ctx, projectID, userID)
}

func (f *fkProjectRepo) GetProjectsByUserID(ctx context.Context, userID string) ([]models.Project, error) {
	f.Calls = append(f.Calls, "GetProjectsByUserID")
	if f.OnGetProjectsByUserID == nil {
		panic("fkProjectRepo: unexpected call to GetProjectsByUserID")
	}
	return f.OnGetProjectsByUserID(ctx, userID)
}

func (f *fkProjectRepo) GetCollaboratorProjects(ctx context.Context, userID string) ([]models.Project, error) {
	f.Calls = append(f.Calls, "GetCollaboratorProjects")
	if f.OnGetCollaboratorProjects == nil {
		panic("fkProjectRepo: unexpected call to GetCollaboratorProjects")
	}
	return f.OnGetCollaboratorProjects(ctx, userID)
}

func (f *fkProjectRepo) GetProjectPages(ctx context.Context, projectID string, userID string) ([]models.Page, error) {
	f.Calls = append(f.Calls, "GetProjectPages")
	if f.OnGetProjectPages == nil {
		panic("fkProjectRepo: unexpected call to GetProjectPages")
	}
	return f.OnGetProjectPages(ctx, projectID, userID)
}

func (f *fkProjectRepo) CreateProject(ctx context.Context, project *models.Project) error {
	f.Calls = append(f.Calls, "CreateProject")
	if f.OnCreateProject == nil {
		panic("fkProjectRepo: unexpected call to CreateProject")
	}
	return f.OnCreateProject(ctx, project)
}

func (f *fkProjectRepo) CreateProjectWithPage(ctx context.Context, project *models.Project, page *models.Page) error {
	f.Calls = append(f.Calls, "CreateProjectWithPage")
	if f.OnCreateProjectWithPage == nil {
		panic("fkProjectRepo: unexpected call to CreateProjectWithPage")
	}
	return f.OnCreateProjectWithPage(ctx, project, page)
}

func (f *fkProjectRepo) GetProjectAccess(ctx context.Context, projectID string, userID string) (*models.ProjectAccess, error) {
	f.Calls = append(f.Calls, "GetProjectAccess")
	if f.OnGetProjectAccess == nil {
		panic("fkProjectRepo: unexpected call to GetProjectAccess")
	}
	return f.OnGetProjectAccess(ctx, projectID, userID)
}

func (f *fkProjectRepo) UpdateProject(ctx context.Context, projectID string, userID string, updates map[string]any) (*models.Project, error) {
	f.Calls = append(f.Calls, "UpdateProject")
	if f.OnUpdateProject == nil {
		panic("fkProjectRepo: unexpected call to UpdateProject")
	}
	return f.OnUpdateProject(ctx, projectID, userID, updates)
}

func (f *fkProjectRepo) UpdateProjectFields(ctx context.Context, projectID string, userID string, updates map[string]any) (*models.Project, error) {
	f.Calls = append(f.Calls, "UpdateProjectFields")
	if f.OnUpdateProjectFields != nil {
		return f.OnUpdateProjectFields(ctx, projectID, userID, updates)
	}
	if f.OnUpdateProject != nil {
		return f.OnUpdateProject(ctx, projectID, userID, updates)
	}
	panic("fkProjectRepo: unexpected call to UpdateProjectFields")
}

func (f *fkProjectRepo) TransferOwnership(ctx context.Context, projectID string, currentOwnerID string, newOwnerID string) error {
	f.Calls = append(f.Calls, "TransferOwnership")
	if f.OnTransferOwnership == nil {
		panic("fkProjectRepo: unexpected call to TransferOwnership")
	}
	return f.OnTransferOwnership(ctx, projectID, currentOwnerID, newOwnerID)
}

func (f *fkProjectRepo) DeleteProject(ctx context.Context, projectID string, userID string) error {
	f.Calls = append(f.Calls, "DeleteProject")
	if f.OnDeleteProject == nil {
		panic("fkProjectRepo: unexpected call to DeleteProject")
	}
	return f.OnDeleteProject(ctx, projectID, userID)
}

func (f *fkProjectRepo) HardDeleteProject(ctx context.Context, projectID string, userID string) error {
	f.Calls = append(f.Calls, "HardDeleteProject")
	if f.OnHardDeleteProject == nil {
		panic("fkProjectRepo: unexpected call to HardDeleteProject")
	}
	return f.OnHardDeleteProject(ctx, projectID, userID)
}

func (f *fkProjectRepo) RestoreProject(ctx context.Context, projectID string, userID string) error {
	f.Calls = append(f.Calls, "RestoreProject")
	if f.OnRestoreProject == nil {
		panic("fkProjectRepo: unexpected call to RestoreProject")
	}
	return f.OnRestoreProject(ctx, projectID, userID)
}

func (f *fkProjectRepo) ExistsForUser(ctx context.Context, projectID string, userID string) (bool, error) {
	f.Calls = append(f.Calls, "ExistsForUser")
	if f.OnExistsForUser == nil {
		panic("fkProjectRepo: unexpected call to ExistsForUser")
	}
	return f.OnExistsForUser(ctx, projectID, userID)
}

func (f *fkProjectRepo) GetProjectWithLock(ctx context.Context, projectID string, userID string) (*models.Project, error) {
	f.Calls = append(f.Calls, "GetProjectWithLock")
	if f.OnGetProjectWithLock == nil {
		panic("fkProjectRepo: unexpected call to GetProjectWithLock")
	}
	return f.OnGetProjectWithLock(ctx, projectID, userID)
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
	Calls                     []string
	OnSaveSnapshot            func(ctx context.Context, projectID string, snapshot *models.Snapshot) error
	OnGetSnapshotsByProjectID func(ctx context.Context, projectID string) ([]models.Snapshot, error)
	OnGetSnapshotByID         func(ctx context.Context, snapshotID string) (*models.Snapshot, error)
	OnDeleteSnapshot          func(ctx context.Context, snapshotID string) error
}

func (f *fkSnapshotRepo) SaveSnapshot(ctx context.Context, projectID string, snapshot *models.Snapshot) error {
	f.Calls = append(f.Calls, "SaveSnapshot")
	if f.OnSaveSnapshot == nil {
		panic("fkSnapshotRepo: unexpected call to SaveSnapshot")
	}
	return f.OnSaveSnapshot(ctx, projectID, snapshot)
}

func (f *fkSnapshotRepo) GetSnapshotsByProjectID(ctx context.Context, projectID string) ([]models.Snapshot, error) {
	f.Calls = append(f.Calls, "GetSnapshotsByProjectID")
	if f.OnGetSnapshotsByProjectID == nil {
		panic("fkSnapshotRepo: unexpected call to GetSnapshotsByProjectID")
	}
	return f.OnGetSnapshotsByProjectID(ctx, projectID)
}

func (f *fkSnapshotRepo) GetSnapshotByID(ctx context.Context, snapshotID string) (*models.Snapshot, error) {
	f.Calls = append(f.Calls, "GetSnapshotByID")
	if f.OnGetSnapshotByID == nil {
		panic("fkSnapshotRepo: unexpected call to GetSnapshotByID")
	}
	return f.OnGetSnapshotByID(ctx, snapshotID)
}

func (f *fkSnapshotRepo) DeleteSnapshot(ctx context.Context, snapshotID string) error {
	f.Calls = append(f.Calls, "DeleteSnapshot")
	if f.OnDeleteSnapshot == nil {
		panic("fkSnapshotRepo: unexpected call to DeleteSnapshot")
	}
	return f.OnDeleteSnapshot(ctx, snapshotID)
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
	Calls                   []string
	OnGetPagesByProjectID   func(ctx context.Context, projectID string) ([]models.Page, error)
	OnGetPageByID           func(ctx context.Context, pageID string, projectID string) (*models.Page, error)
	OnCreatePage            func(ctx context.Context, page *models.Page) error
	OnUpdatePage            func(ctx context.Context, page *models.Page) error
	OnUpdatePageFields      func(ctx context.Context, pageID string, updates map[string]any) error
	OnDeletePage            func(ctx context.Context, pageID string) error
	OnDeletePageByProjectID func(ctx context.Context, pageID string, projectID string, userID string) error
}

func (f *fkPageRepo) GetPagesByProjectID(ctx context.Context, projectID string) ([]models.Page, error) {
	f.Calls = append(f.Calls, "GetPagesByProjectID")
	if f.OnGetPagesByProjectID == nil {
		panic("fkPageRepo: unexpected call to GetPagesByProjectID")
	}
	return f.OnGetPagesByProjectID(ctx, projectID)
}

func (f *fkPageRepo) GetPageByID(ctx context.Context, pageID string, projectID string) (*models.Page, error) {
	f.Calls = append(f.Calls, "GetPageByID")
	if f.OnGetPageByID == nil {
		panic("fkPageRepo: unexpected call to GetPageByID")
	}
	return f.OnGetPageByID(ctx, pageID, projectID)
}

func (f *fkPageRepo) CreatePage(ctx context.Context, page *models.Page) error {
	f.Calls = append(f.Calls, "CreatePage")
	if f.OnCreatePage == nil {
		panic("fkPageRepo: unexpected call to CreatePage")
	}
	return f.OnCreatePage(ctx, page)
}

func (f *fkPageRepo) UpdatePage(ctx context.Context, page *models.Page) error {
	f.Calls = append(f.Calls, "UpdatePage")
	if f.OnUpdatePage == nil {
		panic("fkPageRepo: unexpected call to UpdatePage")
	}
	return f.OnUpdatePage(ctx, page)
}

func (f *fkPageRepo) UpdatePageFields(ctx context.Context, pageID string, updates map[string]any) error {
	f.Calls = append(f.Calls, "UpdatePageFields")
	if f.OnUpdatePageFields == nil {
		panic("fkPageRepo: unexpected call to UpdatePageFields")
	}
	return f.OnUpdatePageFields(ctx, pageID, updates)
}

func (f *fkPageRepo) DeletePage(ctx context.Context, pageID string) error {
	f.Calls = append(f.Calls, "DeletePage")
	if f.OnDeletePage == nil {
		panic("fkPageRepo: unexpected call to DeletePage")
	}
	return f.OnDeletePage(ctx, pageID)
}

func (f *fkPageRepo) DeletePageByProjectID(ctx context.Context, pageID string, projectID string, userID string) error {
	f.Calls = append(f.Calls, "DeletePageByProjectID")
	if f.OnDeletePageByProjectID == nil {
		panic("fkPageRepo: unexpected call to DeletePageByProjectID")
	}
	return f.OnDeletePageByProjectID(ctx, pageID, projectID, userID)
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
	Calls                      []string
	OnGetContentTypes          func(ctx context.Context) ([]models.ContentType, error)
	OnGetContentTypesByProject func(ctx context.Context, projectID string) ([]models.ContentType, error)
	OnGetContentTypesByUser    func(ctx context.Context, userID string) ([]models.ContentType, error)
	OnGetContentTypeByID                 func(ctx context.Context, id string) (*models.ContentType, error)
	OnGetContentTypeByProjectAndSlugOrID func(ctx context.Context, projectID, slugOrID string) (*models.ContentType, error)
	OnCreateContentType                  func(ctx context.Context, ct *models.ContentType) (*models.ContentType, error)
	OnUpdateContentType        func(ctx context.Context, id string, updates map[string]any) (*models.ContentType, error)
	OnDeleteContentType        func(ctx context.Context, id string) error
}

func (f *fkContentTypeRepo) GetContentTypes(ctx context.Context) ([]models.ContentType, error) {
	f.Calls = append(f.Calls, "GetContentTypes")
	if f.OnGetContentTypes == nil {
		panic("fkContentTypeRepo: unexpected call to GetContentTypes")
	}
	return f.OnGetContentTypes(ctx)
}

func (f *fkContentTypeRepo) GetContentTypesByProject(ctx context.Context, projectID string) ([]models.ContentType, error) {
	f.Calls = append(f.Calls, "GetContentTypesByProject")
	if f.OnGetContentTypesByProject == nil {
		panic("fkContentTypeRepo: unexpected call to GetContentTypesByProject")
	}
	return f.OnGetContentTypesByProject(ctx, projectID)
}

func (f *fkContentTypeRepo) GetContentTypesByUser(ctx context.Context, userID string) ([]models.ContentType, error) {
	f.Calls = append(f.Calls, "GetContentTypesByUser")
	if f.OnGetContentTypesByUser == nil {
		panic("fkContentTypeRepo: unexpected call to GetContentTypesByUser")
	}
	return f.OnGetContentTypesByUser(ctx, userID)
}

func (f *fkContentTypeRepo) GetContentTypeByID(ctx context.Context, id string) (*models.ContentType, error) {
	f.Calls = append(f.Calls, "GetContentTypeByID")
	if f.OnGetContentTypeByID == nil {
		panic("fkContentTypeRepo: unexpected call to GetContentTypeByID")
	}
	return f.OnGetContentTypeByID(ctx, id)
}

func (f *fkContentTypeRepo) GetContentTypeByProjectAndSlugOrID(ctx context.Context, projectID, slugOrID string) (*models.ContentType, error) {
	f.Calls = append(f.Calls, "GetContentTypeByProjectAndSlugOrID")
	if f.OnGetContentTypeByProjectAndSlugOrID == nil {
		panic("fkContentTypeRepo: unexpected call to GetContentTypeByProjectAndSlugOrID")
	}
	return f.OnGetContentTypeByProjectAndSlugOrID(ctx, projectID, slugOrID)
}

func (f *fkContentTypeRepo) CreateContentType(ctx context.Context, ct *models.ContentType) (*models.ContentType, error) {
	f.Calls = append(f.Calls, "CreateContentType")
	if f.OnCreateContentType == nil {
		panic("fkContentTypeRepo: unexpected call to CreateContentType")
	}
	return f.OnCreateContentType(ctx, ct)
}

func (f *fkContentTypeRepo) UpdateContentType(ctx context.Context, id string, updates map[string]any) (*models.ContentType, error) {
	f.Calls = append(f.Calls, "UpdateContentType")
	if f.OnUpdateContentType == nil {
		panic("fkContentTypeRepo: unexpected call to UpdateContentType")
	}
	return f.OnUpdateContentType(ctx, id, updates)
}

func (f *fkContentTypeRepo) DeleteContentType(ctx context.Context, id string) error {
	f.Calls = append(f.Calls, "DeleteContentType")
	if f.OnDeleteContentType == nil {
		panic("fkContentTypeRepo: unexpected call to DeleteContentType")
	}
	return f.OnDeleteContentType(ctx, id)
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
	Calls                           []string
	OnGetContentFieldsByContentType func(ctx context.Context, contentTypeID string) ([]models.ContentField, error)
	OnGetContentFieldByID           func(ctx context.Context, id string) (*models.ContentField, error)
	OnCreateContentField            func(ctx context.Context, cf *models.ContentField) (*models.ContentField, error)
	OnUpdateContentField            func(ctx context.Context, id string, updates map[string]any) (*models.ContentField, error)
	OnDeleteContentField            func(ctx context.Context, id string) error
}

func (f *fkContentFieldRepo) GetContentFieldsByContentType(ctx context.Context, contentTypeID string) ([]models.ContentField, error) {
	f.Calls = append(f.Calls, "GetContentFieldsByContentType")
	if f.OnGetContentFieldsByContentType == nil {
		panic("fkContentFieldRepo: unexpected call to GetContentFieldsByContentType")
	}
	return f.OnGetContentFieldsByContentType(ctx, contentTypeID)
}

func (f *fkContentFieldRepo) GetContentFieldByID(ctx context.Context, id string) (*models.ContentField, error) {
	f.Calls = append(f.Calls, "GetContentFieldByID")
	if f.OnGetContentFieldByID == nil {
		panic("fkContentFieldRepo: unexpected call to GetContentFieldByID")
	}
	return f.OnGetContentFieldByID(ctx, id)
}

func (f *fkContentFieldRepo) CreateContentField(ctx context.Context, cf *models.ContentField) (*models.ContentField, error) {
	f.Calls = append(f.Calls, "CreateContentField")
	if f.OnCreateContentField == nil {
		panic("fkContentFieldRepo: unexpected call to CreateContentField")
	}
	return f.OnCreateContentField(ctx, cf)
}

func (f *fkContentFieldRepo) UpdateContentField(ctx context.Context, id string, updates map[string]any) (*models.ContentField, error) {
	f.Calls = append(f.Calls, "UpdateContentField")
	if f.OnUpdateContentField == nil {
		panic("fkContentFieldRepo: unexpected call to UpdateContentField")
	}
	return f.OnUpdateContentField(ctx, id, updates)
}

func (f *fkContentFieldRepo) DeleteContentField(ctx context.Context, id string) error {
	f.Calls = append(f.Calls, "DeleteContentField")
	if f.OnDeleteContentField == nil {
		panic("fkContentFieldRepo: unexpected call to DeleteContentField")
	}
	return f.OnDeleteContentField(ctx, id)
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
	Calls                          []string
	OnGetContentItemsByContentType func(ctx context.Context, contentTypeID string) ([]models.ContentItem, error)
	OnGetContentItemByID           func(ctx context.Context, id string) (*models.ContentItem, error)
	OnGetContentItemBySlug         func(ctx context.Context, contentTypeID string, slug string) (*models.ContentItem, error)
	OnGetPublicContentItems        func(ctx context.Context, contentTypeID string, limit int, sortBy string, sortOrder string) ([]models.ContentItem, error)
	OnCreateContentItem            func(ctx context.Context, ci *models.ContentItem, fieldValues []models.ContentFieldValue) (*models.ContentItem, error)
	OnUpdateContentItem            func(ctx context.Context, id string, updates map[string]any, fieldValues []models.ContentFieldValue) (*models.ContentItem, error)
	OnDeleteContentItem            func(ctx context.Context, id string) error
}

func (f *fkContentItemRepo) GetContentItemsByContentType(ctx context.Context, contentTypeID string) ([]models.ContentItem, error) {
	f.Calls = append(f.Calls, "GetContentItemsByContentType")
	if f.OnGetContentItemsByContentType == nil {
		panic("fkContentItemRepo: unexpected call to GetContentItemsByContentType")
	}
	return f.OnGetContentItemsByContentType(ctx, contentTypeID)
}

func (f *fkContentItemRepo) GetContentItemByID(ctx context.Context, id string) (*models.ContentItem, error) {
	f.Calls = append(f.Calls, "GetContentItemByID")
	if f.OnGetContentItemByID == nil {
		panic("fkContentItemRepo: unexpected call to GetContentItemByID")
	}
	return f.OnGetContentItemByID(ctx, id)
}

func (f *fkContentItemRepo) GetContentItemBySlug(ctx context.Context, contentTypeID string, slug string) (*models.ContentItem, error) {
	f.Calls = append(f.Calls, "GetContentItemBySlug")
	if f.OnGetContentItemBySlug == nil {
		panic("fkContentItemRepo: unexpected call to GetContentItemBySlug")
	}
	return f.OnGetContentItemBySlug(ctx, contentTypeID, slug)
}

func (f *fkContentItemRepo) GetPublicContentItems(ctx context.Context, contentTypeID string, limit int, sortBy string, sortOrder string) ([]models.ContentItem, error) {
	f.Calls = append(f.Calls, "GetPublicContentItems")
	if f.OnGetPublicContentItems == nil {
		panic("fkContentItemRepo: unexpected call to GetPublicContentItems")
	}
	return f.OnGetPublicContentItems(ctx, contentTypeID, limit, sortBy, sortOrder)
}

func (f *fkContentItemRepo) CreateContentItem(ctx context.Context, ci *models.ContentItem, fieldValues []models.ContentFieldValue) (*models.ContentItem, error) {
	f.Calls = append(f.Calls, "CreateContentItem")
	if f.OnCreateContentItem == nil {
		panic("fkContentItemRepo: unexpected call to CreateContentItem")
	}
	return f.OnCreateContentItem(ctx, ci, fieldValues)
}

func (f *fkContentItemRepo) UpdateContentItem(ctx context.Context, id string, updates map[string]any, fieldValues []models.ContentFieldValue) (*models.ContentItem, error) {
	f.Calls = append(f.Calls, "UpdateContentItem")
	if f.OnUpdateContentItem == nil {
		panic("fkContentItemRepo: unexpected call to UpdateContentItem")
	}
	return f.OnUpdateContentItem(ctx, id, updates, fieldValues)
}

func (f *fkContentItemRepo) DeleteContentItem(ctx context.Context, id string) error {
	f.Calls = append(f.Calls, "DeleteContentItem")
	if f.OnDeleteContentItem == nil {
		panic("fkContentItemRepo: unexpected call to DeleteContentItem")
	}
	return f.OnDeleteContentItem(ctx, id)
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
	Calls               []string
	OnCreateImage       func(ctx context.Context, image models.Image) (*models.Image, error)
	OnGetImagesByUserID func(ctx context.Context, userID string) ([]models.Image, error)
	OnGetImageByID      func(ctx context.Context, imageID string, userID string) (*models.Image, error)
	OnDeleteImage       func(ctx context.Context, imageID string, userID string) error
	OnSoftDeleteImage   func(ctx context.Context, imageID string, userID string) error
	OnGetAllImages      func(ctx context.Context, limit int, offset int) ([]models.Image, error)
}

func (f *fkImageRepo) CreateImage(ctx context.Context, image models.Image) (*models.Image, error) {
	f.Calls = append(f.Calls, "CreateImage")
	if f.OnCreateImage == nil {
		panic("fkImageRepo: unexpected call to CreateImage")
	}
	return f.OnCreateImage(ctx, image)
}

func (f *fkImageRepo) GetImagesByUserID(ctx context.Context, userID string) ([]models.Image, error) {
	f.Calls = append(f.Calls, "GetImagesByUserID")
	if f.OnGetImagesByUserID == nil {
		panic("fkImageRepo: unexpected call to GetImagesByUserID")
	}
	return f.OnGetImagesByUserID(ctx, userID)
}

func (f *fkImageRepo) GetImageByID(ctx context.Context, imageID string, userID string) (*models.Image, error) {
	f.Calls = append(f.Calls, "GetImageByID")
	if f.OnGetImageByID == nil {
		panic("fkImageRepo: unexpected call to GetImageByID")
	}
	return f.OnGetImageByID(ctx, imageID, userID)
}

func (f *fkImageRepo) DeleteImage(ctx context.Context, imageID string, userID string) error {
	f.Calls = append(f.Calls, "DeleteImage")
	if f.OnDeleteImage == nil {
		panic("fkImageRepo: unexpected call to DeleteImage")
	}
	return f.OnDeleteImage(ctx, imageID, userID)
}

func (f *fkImageRepo) SoftDeleteImage(ctx context.Context, imageID string, userID string) error {
	f.Calls = append(f.Calls, "SoftDeleteImage")
	if f.OnSoftDeleteImage == nil {
		panic("fkImageRepo: unexpected call to SoftDeleteImage")
	}
	return f.OnSoftDeleteImage(ctx, imageID, userID)
}

func (f *fkImageRepo) GetAllImages(ctx context.Context, limit int, offset int) ([]models.Image, error) {
	f.Calls = append(f.Calls, "GetAllImages")
	if f.OnGetAllImages == nil {
		panic("fkImageRepo: unexpected call to GetAllImages")
	}
	return f.OnGetAllImages(ctx, limit, offset)
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
	Calls                     []string
	OnGetCustomElements       func(ctx context.Context, userID string, isPublic *bool) ([]models.CustomElement, error)
	OnGetCustomElementByID    func(ctx context.Context, id string, userID string) (*models.CustomElement, error)
	OnCreateCustomElement     func(ctx context.Context, customElement *models.CustomElement) (*models.CustomElement, error)
	OnUpdateCustomElement     func(ctx context.Context, id string, userID string, updates map[string]any) (*models.CustomElement, error)
	OnDeleteCustomElement     func(ctx context.Context, id string, userID string) error
	OnGetPublicCustomElements func(ctx context.Context, category *string, limit int, offset int) ([]models.CustomElement, error)
	OnDuplicateCustomElement  func(ctx context.Context, id string, userID string, newName string) (*models.CustomElement, error)
}

func (f *fkCustomElementRepo) GetCustomElements(ctx context.Context, userID string, isPublic *bool) ([]models.CustomElement, error) {
	f.Calls = append(f.Calls, "GetCustomElements")
	if f.OnGetCustomElements == nil {
		panic("fkCustomElementRepo: unexpected call to GetCustomElements")
	}
	return f.OnGetCustomElements(ctx, userID, isPublic)
}

func (f *fkCustomElementRepo) GetCustomElementByID(ctx context.Context, id string, userID string) (*models.CustomElement, error) {
	f.Calls = append(f.Calls, "GetCustomElementByID")
	if f.OnGetCustomElementByID == nil {
		panic("fkCustomElementRepo: unexpected call to GetCustomElementByID")
	}
	return f.OnGetCustomElementByID(ctx, id, userID)
}

func (f *fkCustomElementRepo) CreateCustomElement(ctx context.Context, customElement *models.CustomElement) (*models.CustomElement, error) {
	f.Calls = append(f.Calls, "CreateCustomElement")
	if f.OnCreateCustomElement == nil {
		panic("fkCustomElementRepo: unexpected call to CreateCustomElement")
	}
	return f.OnCreateCustomElement(ctx, customElement)
}

func (f *fkCustomElementRepo) UpdateCustomElement(ctx context.Context, id string, userID string, updates map[string]any) (*models.CustomElement, error) {
	f.Calls = append(f.Calls, "UpdateCustomElement")
	if f.OnUpdateCustomElement == nil {
		panic("fkCustomElementRepo: unexpected call to UpdateCustomElement")
	}
	return f.OnUpdateCustomElement(ctx, id, userID, updates)
}

func (f *fkCustomElementRepo) DeleteCustomElement(ctx context.Context, id string, userID string) error {
	f.Calls = append(f.Calls, "DeleteCustomElement")
	if f.OnDeleteCustomElement == nil {
		panic("fkCustomElementRepo: unexpected call to DeleteCustomElement")
	}
	return f.OnDeleteCustomElement(ctx, id, userID)
}

func (f *fkCustomElementRepo) GetPublicCustomElements(ctx context.Context, category *string, limit int, offset int) ([]models.CustomElement, error) {
	f.Calls = append(f.Calls, "GetPublicCustomElements")
	if f.OnGetPublicCustomElements == nil {
		panic("fkCustomElementRepo: unexpected call to GetPublicCustomElements")
	}
	return f.OnGetPublicCustomElements(ctx, category, limit, offset)
}

func (f *fkCustomElementRepo) DuplicateCustomElement(ctx context.Context, id string, userID string, newName string) (*models.CustomElement, error) {
	f.Calls = append(f.Calls, "DuplicateCustomElement")
	if f.OnDuplicateCustomElement == nil {
		panic("fkCustomElementRepo: unexpected call to DuplicateCustomElement")
	}
	return f.OnDuplicateCustomElement(ctx, id, userID, newName)
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
	Calls                        []string
	OnGetCustomElementTypes      func(ctx context.Context) ([]models.CustomElementType, error)
	OnGetCustomElementTypeByID   func(ctx context.Context, id string) (*models.CustomElementType, error)
	OnGetCustomElementTypeByName func(ctx context.Context, name string) (*models.CustomElementType, error)
	OnCreateCustomElementType    func(ctx context.Context, customElementType *models.CustomElementType) (*models.CustomElementType, error)
	OnUpdateCustomElementType    func(ctx context.Context, id string, updates map[string]any) (*models.CustomElementType, error)
	OnDeleteCustomElementType    func(ctx context.Context, id string) error
}

func (f *fkCustomElementTypeRepo) GetCustomElementTypes(ctx context.Context) ([]models.CustomElementType, error) {
	f.Calls = append(f.Calls, "GetCustomElementTypes")
	if f.OnGetCustomElementTypes == nil {
		panic("fkCustomElementTypeRepo: unexpected call to GetCustomElementTypes")
	}
	return f.OnGetCustomElementTypes(ctx)
}

func (f *fkCustomElementTypeRepo) GetCustomElementTypeByID(ctx context.Context, id string) (*models.CustomElementType, error) {
	f.Calls = append(f.Calls, "GetCustomElementTypeByID")
	if f.OnGetCustomElementTypeByID == nil {
		panic("fkCustomElementTypeRepo: unexpected call to GetCustomElementTypeByID")
	}
	return f.OnGetCustomElementTypeByID(ctx, id)
}

func (f *fkCustomElementTypeRepo) GetCustomElementTypeByName(ctx context.Context, name string) (*models.CustomElementType, error) {
	f.Calls = append(f.Calls, "GetCustomElementTypeByName")
	if f.OnGetCustomElementTypeByName == nil {
		panic("fkCustomElementTypeRepo: unexpected call to GetCustomElementTypeByName")
	}
	return f.OnGetCustomElementTypeByName(ctx, name)
}

func (f *fkCustomElementTypeRepo) CreateCustomElementType(ctx context.Context, customElementType *models.CustomElementType) (*models.CustomElementType, error) {
	f.Calls = append(f.Calls, "CreateCustomElementType")
	if f.OnCreateCustomElementType == nil {
		panic("fkCustomElementTypeRepo: unexpected call to CreateCustomElementType")
	}
	return f.OnCreateCustomElementType(ctx, customElementType)
}

func (f *fkCustomElementTypeRepo) UpdateCustomElementType(ctx context.Context, id string, updates map[string]any) (*models.CustomElementType, error) {
	f.Calls = append(f.Calls, "UpdateCustomElementType")
	if f.OnUpdateCustomElementType == nil {
		panic("fkCustomElementTypeRepo: unexpected call to UpdateCustomElementType")
	}
	return f.OnUpdateCustomElementType(ctx, id, updates)
}

func (f *fkCustomElementTypeRepo) DeleteCustomElementType(ctx context.Context, id string) error {
	f.Calls = append(f.Calls, "DeleteCustomElementType")
	if f.OnDeleteCustomElementType == nil {
		panic("fkCustomElementTypeRepo: unexpected call to DeleteCustomElementType")
	}
	return f.OnDeleteCustomElementType(ctx, id)
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
	Calls                            []string
	OnCreateInvitation               func(ctx context.Context, invitation *models.Invitation) (*models.Invitation, error)
	OnGetInvitationsByProject        func(ctx context.Context, projectID string) ([]models.Invitation, error)
	OnGetInvitationByID              func(ctx context.Context, id string) (*models.Invitation, error)
	OnGetInvitationByToken           func(ctx context.Context, token string) (*models.Invitation, error)
	OnAcceptInvitation               func(ctx context.Context, token string, userID string) error
	OnDeleteInvitation               func(ctx context.Context, id string) error
	OnUpdateInvitationStatus         func(ctx context.Context, id string, status models.InvitationStatus) error
	OnCancelInvitation               func(ctx context.Context, id string) error
	OnGetPendingInvitationsByProject func(ctx context.Context, projectID string) ([]models.Invitation, error)
}

func (f *fkInvitationRepo) CreateInvitation(ctx context.Context, invitation *models.Invitation) (*models.Invitation, error) {
	f.Calls = append(f.Calls, "CreateInvitation")
	if f.OnCreateInvitation == nil {
		panic("fkInvitationRepo: unexpected call to CreateInvitation")
	}
	return f.OnCreateInvitation(ctx, invitation)
}

func (f *fkInvitationRepo) GetInvitationsByProject(ctx context.Context, projectID string) ([]models.Invitation, error) {
	f.Calls = append(f.Calls, "GetInvitationsByProject")
	if f.OnGetInvitationsByProject == nil {
		panic("fkInvitationRepo: unexpected call to GetInvitationsByProject")
	}
	return f.OnGetInvitationsByProject(ctx, projectID)
}

func (f *fkInvitationRepo) GetInvitationByID(ctx context.Context, id string) (*models.Invitation, error) {
	f.Calls = append(f.Calls, "GetInvitationByID")
	if f.OnGetInvitationByID == nil {
		panic("fkInvitationRepo: unexpected call to GetInvitationByID")
	}
	return f.OnGetInvitationByID(ctx, id)
}

func (f *fkInvitationRepo) GetInvitationByToken(ctx context.Context, token string) (*models.Invitation, error) {
	f.Calls = append(f.Calls, "GetInvitationByToken")
	if f.OnGetInvitationByToken == nil {
		panic("fkInvitationRepo: unexpected call to GetInvitationByToken")
	}
	return f.OnGetInvitationByToken(ctx, token)
}

func (f *fkInvitationRepo) AcceptInvitation(ctx context.Context, token string, userID string) error {
	f.Calls = append(f.Calls, "AcceptInvitation")
	if f.OnAcceptInvitation == nil {
		panic("fkInvitationRepo: unexpected call to AcceptInvitation")
	}
	return f.OnAcceptInvitation(ctx, token, userID)
}

func (f *fkInvitationRepo) DeleteInvitation(ctx context.Context, id string) error {
	f.Calls = append(f.Calls, "DeleteInvitation")
	if f.OnDeleteInvitation == nil {
		panic("fkInvitationRepo: unexpected call to DeleteInvitation")
	}
	return f.OnDeleteInvitation(ctx, id)
}

func (f *fkInvitationRepo) UpdateInvitationStatus(ctx context.Context, id string, status models.InvitationStatus) error {
	f.Calls = append(f.Calls, "UpdateInvitationStatus")
	if f.OnUpdateInvitationStatus == nil {
		panic("fkInvitationRepo: unexpected call to UpdateInvitationStatus")
	}
	return f.OnUpdateInvitationStatus(ctx, id, status)
}

func (f *fkInvitationRepo) CancelInvitation(ctx context.Context, id string) error {
	f.Calls = append(f.Calls, "CancelInvitation")
	if f.OnCancelInvitation == nil {
		panic("fkInvitationRepo: unexpected call to CancelInvitation")
	}
	return f.OnCancelInvitation(ctx, id)
}

func (f *fkInvitationRepo) GetPendingInvitationsByProject(ctx context.Context, projectID string) ([]models.Invitation, error) {
	f.Calls = append(f.Calls, "GetPendingInvitationsByProject")
	if f.OnGetPendingInvitationsByProject == nil {
		panic("fkInvitationRepo: unexpected call to GetPendingInvitationsByProject")
	}
	return f.OnGetPendingInvitationsByProject(ctx, projectID)
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
	Calls                               []string
	OnCreateCollaborator                func(ctx context.Context, collaborator *models.Collaborator) (*models.Collaborator, error)
	OnGetCollaboratorsByProject         func(ctx context.Context, projectID string) ([]models.Collaborator, error)
	OnGetCollaboratorByID               func(ctx context.Context, id string) (*models.Collaborator, error)
	OnUpdateCollaboratorRole            func(ctx context.Context, id string, role models.CollaboratorRole) error
	OnDeleteCollaborator                func(ctx context.Context, id string) error
	OnIsCollaborator                    func(ctx context.Context, projectID string, userID string) (bool, error)
	OnGetCollaboratorByProjectAndUser   func(ctx context.Context, projectID string, userID string) (*models.Collaborator, error)
	OnDeleteCollaboratorByProjectAndUser func(ctx context.Context, projectID string, userID string) (bool, error)
}

func (f *fkCollaboratorRepo) CreateCollaborator(ctx context.Context, collaborator *models.Collaborator) (*models.Collaborator, error) {
	f.Calls = append(f.Calls, "CreateCollaborator")
	if f.OnCreateCollaborator == nil {
		panic("fkCollaboratorRepo: unexpected call to CreateCollaborator")
	}
	return f.OnCreateCollaborator(ctx, collaborator)
}

func (f *fkCollaboratorRepo) GetCollaboratorsByProject(ctx context.Context, projectID string) ([]models.Collaborator, error) {
	f.Calls = append(f.Calls, "GetCollaboratorsByProject")
	if f.OnGetCollaboratorsByProject == nil {
		panic("fkCollaboratorRepo: unexpected call to GetCollaboratorsByProject")
	}
	return f.OnGetCollaboratorsByProject(ctx, projectID)
}

func (f *fkCollaboratorRepo) GetCollaboratorByID(ctx context.Context, id string) (*models.Collaborator, error) {
	f.Calls = append(f.Calls, "GetCollaboratorByID")
	if f.OnGetCollaboratorByID == nil {
		panic("fkCollaboratorRepo: unexpected call to GetCollaboratorByID")
	}
	return f.OnGetCollaboratorByID(ctx, id)
}

func (f *fkCollaboratorRepo) UpdateCollaboratorRole(ctx context.Context, id string, role models.CollaboratorRole) error {
	f.Calls = append(f.Calls, "UpdateCollaboratorRole")
	if f.OnUpdateCollaboratorRole == nil {
		panic("fkCollaboratorRepo: unexpected call to UpdateCollaboratorRole")
	}
	return f.OnUpdateCollaboratorRole(ctx, id, role)
}

func (f *fkCollaboratorRepo) DeleteCollaborator(ctx context.Context, id string) error {
	f.Calls = append(f.Calls, "DeleteCollaborator")
	if f.OnDeleteCollaborator == nil {
		panic("fkCollaboratorRepo: unexpected call to DeleteCollaborator")
	}
	return f.OnDeleteCollaborator(ctx, id)
}

func (f *fkCollaboratorRepo) IsCollaborator(ctx context.Context, projectID string, userID string) (bool, error) {
	f.Calls = append(f.Calls, "IsCollaborator")
	if f.OnIsCollaborator == nil {
		panic("fkCollaboratorRepo: unexpected call to IsCollaborator")
	}
	return f.OnIsCollaborator(ctx, projectID, userID)
}

func (f *fkCollaboratorRepo) GetCollaboratorByProjectAndUser(ctx context.Context, projectID string, userID string) (*models.Collaborator, error) {
	f.Calls = append(f.Calls, "GetCollaboratorByProjectAndUser")
	if f.OnGetCollaboratorByProjectAndUser == nil {
		return nil, nil
	}
	return f.OnGetCollaboratorByProjectAndUser(ctx, projectID, userID)
}

func (f *fkCollaboratorRepo) DeleteCollaboratorByProjectAndUser(ctx context.Context, projectID string, userID string) (bool, error) {
	f.Calls = append(f.Calls, "DeleteCollaboratorByProjectAndUser")
	if f.OnDeleteCollaboratorByProjectAndUser == nil {
		panic("fkCollaboratorRepo: unexpected call to DeleteCollaboratorByProjectAndUser")
	}
	return f.OnDeleteCollaboratorByProjectAndUser(ctx, projectID, userID)
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
	Calls                                      []string
	OnCreateEventWorkflow                      func(ctx context.Context, workflow *models.EventWorkflow) (*models.EventWorkflow, error)
	OnGetEventWorkflowByID                     func(ctx context.Context, id string) (*models.EventWorkflow, error)
	OnGetEventWorkflowsByProjectID             func(ctx context.Context, projectID string) ([]models.EventWorkflow, error)
	OnGetEventWorkflowsByProjectIDWithElements func(ctx context.Context, projectID string) ([]models.EventWorkflow, error)
	OnGetEnabledEventWorkflowsByProjectID      func(ctx context.Context, projectID string) ([]models.EventWorkflow, error)
	OnGetEventWorkflowsByName                  func(ctx context.Context, projectID string, name string) ([]models.EventWorkflow, error)
	OnUpdateEventWorkflow                      func(ctx context.Context, id string, workflow *models.EventWorkflow) (*models.EventWorkflow, error)
	OnUpdateEventWorkflowEnabled               func(ctx context.Context, id string, enabled bool) error
	OnDeleteEventWorkflow                      func(ctx context.Context, id string) error
	OnDeleteEventWorkflowsByProjectID          func(ctx context.Context, projectID string) error
	OnCountEventWorkflowsByProjectID           func(ctx context.Context, projectID string) (int64, error)
	OnCheckIfWorkflowNameExists                func(ctx context.Context, projectID string, name string, excludeID string) (bool, error)
	OnGetEventWorkflowsWithFilters             func(ctx context.Context, projectID string, enabled *bool, searchName string) ([]models.EventWorkflow, error)
}

func (f *fkEventWorkflowRepo) CreateEventWorkflow(ctx context.Context, workflow *models.EventWorkflow) (*models.EventWorkflow, error) {
	f.Calls = append(f.Calls, "CreateEventWorkflow")
	if f.OnCreateEventWorkflow == nil {
		panic("fkEventWorkflowRepo: unexpected call to CreateEventWorkflow")
	}
	return f.OnCreateEventWorkflow(ctx, workflow)
}

func (f *fkEventWorkflowRepo) GetEventWorkflowByID(ctx context.Context, id string) (*models.EventWorkflow, error) {
	f.Calls = append(f.Calls, "GetEventWorkflowByID")
	if f.OnGetEventWorkflowByID == nil {
		panic("fkEventWorkflowRepo: unexpected call to GetEventWorkflowByID")
	}
	return f.OnGetEventWorkflowByID(ctx, id)
}

func (f *fkEventWorkflowRepo) GetEventWorkflowsByProjectID(ctx context.Context, projectID string) ([]models.EventWorkflow, error) {
	f.Calls = append(f.Calls, "GetEventWorkflowsByProjectID")
	if f.OnGetEventWorkflowsByProjectID == nil {
		panic("fkEventWorkflowRepo: unexpected call to GetEventWorkflowsByProjectID")
	}
	return f.OnGetEventWorkflowsByProjectID(ctx, projectID)
}

func (f *fkEventWorkflowRepo) GetEventWorkflowsByProjectIDWithElements(ctx context.Context, projectID string) ([]models.EventWorkflow, error) {
	f.Calls = append(f.Calls, "GetEventWorkflowsByProjectIDWithElements")
	if f.OnGetEventWorkflowsByProjectIDWithElements == nil {
		panic("fkEventWorkflowRepo: unexpected call to GetEventWorkflowsByProjectIDWithElements")
	}
	return f.OnGetEventWorkflowsByProjectIDWithElements(ctx, projectID)
}

func (f *fkEventWorkflowRepo) GetEnabledEventWorkflowsByProjectID(ctx context.Context, projectID string) ([]models.EventWorkflow, error) {
	f.Calls = append(f.Calls, "GetEnabledEventWorkflowsByProjectID")
	if f.OnGetEnabledEventWorkflowsByProjectID == nil {
		panic("fkEventWorkflowRepo: unexpected call to GetEnabledEventWorkflowsByProjectID")
	}
	return f.OnGetEnabledEventWorkflowsByProjectID(ctx, projectID)
}

func (f *fkEventWorkflowRepo) GetEventWorkflowsByName(ctx context.Context, projectID string, name string) ([]models.EventWorkflow, error) {
	f.Calls = append(f.Calls, "GetEventWorkflowsByName")
	if f.OnGetEventWorkflowsByName == nil {
		panic("fkEventWorkflowRepo: unexpected call to GetEventWorkflowsByName")
	}
	return f.OnGetEventWorkflowsByName(ctx, projectID, name)
}

func (f *fkEventWorkflowRepo) UpdateEventWorkflow(ctx context.Context, id string, workflow *models.EventWorkflow) (*models.EventWorkflow, error) {
	f.Calls = append(f.Calls, "UpdateEventWorkflow")
	if f.OnUpdateEventWorkflow == nil {
		panic("fkEventWorkflowRepo: unexpected call to UpdateEventWorkflow")
	}
	return f.OnUpdateEventWorkflow(ctx, id, workflow)
}

func (f *fkEventWorkflowRepo) UpdateEventWorkflowEnabled(ctx context.Context, id string, enabled bool) error {
	f.Calls = append(f.Calls, "UpdateEventWorkflowEnabled")
	if f.OnUpdateEventWorkflowEnabled == nil {
		panic("fkEventWorkflowRepo: unexpected call to UpdateEventWorkflowEnabled")
	}
	return f.OnUpdateEventWorkflowEnabled(ctx, id, enabled)
}

func (f *fkEventWorkflowRepo) DeleteEventWorkflow(ctx context.Context, id string) error {
	f.Calls = append(f.Calls, "DeleteEventWorkflow")
	if f.OnDeleteEventWorkflow == nil {
		panic("fkEventWorkflowRepo: unexpected call to DeleteEventWorkflow")
	}
	return f.OnDeleteEventWorkflow(ctx, id)
}

func (f *fkEventWorkflowRepo) DeleteEventWorkflowsByProjectID(ctx context.Context, projectID string) error {
	f.Calls = append(f.Calls, "DeleteEventWorkflowsByProjectID")
	if f.OnDeleteEventWorkflowsByProjectID == nil {
		panic("fkEventWorkflowRepo: unexpected call to DeleteEventWorkflowsByProjectID")
	}
	return f.OnDeleteEventWorkflowsByProjectID(ctx, projectID)
}

func (f *fkEventWorkflowRepo) CountEventWorkflowsByProjectID(ctx context.Context, projectID string) (int64, error) {
	f.Calls = append(f.Calls, "CountEventWorkflowsByProjectID")
	if f.OnCountEventWorkflowsByProjectID == nil {
		panic("fkEventWorkflowRepo: unexpected call to CountEventWorkflowsByProjectID")
	}
	return f.OnCountEventWorkflowsByProjectID(ctx, projectID)
}

func (f *fkEventWorkflowRepo) CheckIfWorkflowNameExists(ctx context.Context, projectID string, name string, excludeID string) (bool, error) {
	f.Calls = append(f.Calls, "CheckIfWorkflowNameExists")
	if f.OnCheckIfWorkflowNameExists == nil {
		panic("fkEventWorkflowRepo: unexpected call to CheckIfWorkflowNameExists")
	}
	return f.OnCheckIfWorkflowNameExists(ctx, projectID, name, excludeID)
}

func (f *fkEventWorkflowRepo) GetEventWorkflowsWithFilters(ctx context.Context, projectID string, enabled *bool, searchName string) ([]models.EventWorkflow, error) {
	f.Calls = append(f.Calls, "GetEventWorkflowsWithFilters")
	if f.OnGetEventWorkflowsWithFilters == nil {
		panic("fkEventWorkflowRepo: unexpected call to GetEventWorkflowsWithFilters")
	}
	return f.OnGetEventWorkflowsWithFilters(ctx, projectID, enabled, searchName)
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
	Calls                                     []string
	OnCreateElementEventWorkflow              func(ctx context.Context, eew *models.ElementEventWorkflow) (*models.ElementEventWorkflow, error)
	OnGetElementEventWorkflowByID             func(ctx context.Context, id string) (*models.ElementEventWorkflow, error)
	OnGetAllElementEventWorkflows             func(ctx context.Context) ([]models.ElementEventWorkflow, error)
	OnGetElementEventWorkflowsByElementID     func(ctx context.Context, elementID string) ([]models.ElementEventWorkflow, error)
	OnGetElementEventWorkflowsByWorkflowID    func(ctx context.Context, workflowID string) ([]models.ElementEventWorkflow, error)
	OnGetElementEventWorkflowsByEventName     func(ctx context.Context, eventName string) ([]models.ElementEventWorkflow, error)
	OnGetElementEventWorkflowsByFilters       func(ctx context.Context, elementID string, workflowID string, eventName string) ([]models.ElementEventWorkflow, error)
	OnUpdateElementEventWorkflow              func(ctx context.Context, id string, eew *models.ElementEventWorkflow) (*models.ElementEventWorkflow, error)
	OnDeleteElementEventWorkflow              func(ctx context.Context, id string) error
	OnDeleteElementEventWorkflowsByElementID  func(ctx context.Context, elementID string) error
	OnDeleteElementEventWorkflowsByWorkflowID func(ctx context.Context, workflowID string) error
	OnGetElementEventWorkflowsByPageID        func(ctx context.Context, pageID string) ([]models.ElementEventWorkflow, error)
	OnCheckIfWorkflowLinkedToElement          func(ctx context.Context, elementID string, workflowID string, eventName string) (bool, error)
}

func (f *fkElementEventWorkflowRepo) CreateElementEventWorkflow(ctx context.Context, eew *models.ElementEventWorkflow) (*models.ElementEventWorkflow, error) {
	f.Calls = append(f.Calls, "CreateElementEventWorkflow")
	if f.OnCreateElementEventWorkflow == nil {
		panic("fkElementEventWorkflowRepo: unexpected call to CreateElementEventWorkflow")
	}
	return f.OnCreateElementEventWorkflow(ctx, eew)
}

func (f *fkElementEventWorkflowRepo) GetElementEventWorkflowByID(ctx context.Context, id string) (*models.ElementEventWorkflow, error) {
	f.Calls = append(f.Calls, "GetElementEventWorkflowByID")
	if f.OnGetElementEventWorkflowByID == nil {
		panic("fkElementEventWorkflowRepo: unexpected call to GetElementEventWorkflowByID")
	}
	return f.OnGetElementEventWorkflowByID(ctx, id)
}

func (f *fkElementEventWorkflowRepo) GetAllElementEventWorkflows(ctx context.Context) ([]models.ElementEventWorkflow, error) {
	f.Calls = append(f.Calls, "GetAllElementEventWorkflows")
	if f.OnGetAllElementEventWorkflows == nil {
		panic("fkElementEventWorkflowRepo: unexpected call to GetAllElementEventWorkflows")
	}
	return f.OnGetAllElementEventWorkflows(ctx)
}

func (f *fkElementEventWorkflowRepo) GetElementEventWorkflowsByElementID(ctx context.Context, elementID string) ([]models.ElementEventWorkflow, error) {
	f.Calls = append(f.Calls, "GetElementEventWorkflowsByElementID")
	if f.OnGetElementEventWorkflowsByElementID == nil {
		panic("fkElementEventWorkflowRepo: unexpected call to GetElementEventWorkflowsByElementID")
	}
	return f.OnGetElementEventWorkflowsByElementID(ctx, elementID)
}

func (f *fkElementEventWorkflowRepo) GetElementEventWorkflowsByWorkflowID(ctx context.Context, workflowID string) ([]models.ElementEventWorkflow, error) {
	f.Calls = append(f.Calls, "GetElementEventWorkflowsByWorkflowID")
	if f.OnGetElementEventWorkflowsByWorkflowID == nil {
		panic("fkElementEventWorkflowRepo: unexpected call to GetElementEventWorkflowsByWorkflowID")
	}
	return f.OnGetElementEventWorkflowsByWorkflowID(ctx, workflowID)
}

func (f *fkElementEventWorkflowRepo) GetElementEventWorkflowsByEventName(ctx context.Context, eventName string) ([]models.ElementEventWorkflow, error) {
	f.Calls = append(f.Calls, "GetElementEventWorkflowsByEventName")
	if f.OnGetElementEventWorkflowsByEventName == nil {
		panic("fkElementEventWorkflowRepo: unexpected call to GetElementEventWorkflowsByEventName")
	}
	return f.OnGetElementEventWorkflowsByEventName(ctx, eventName)
}

func (f *fkElementEventWorkflowRepo) GetElementEventWorkflowsByFilters(ctx context.Context, elementID string, workflowID string, eventName string) ([]models.ElementEventWorkflow, error) {
	f.Calls = append(f.Calls, "GetElementEventWorkflowsByFilters")
	if f.OnGetElementEventWorkflowsByFilters == nil {
		panic("fkElementEventWorkflowRepo: unexpected call to GetElementEventWorkflowsByFilters")
	}
	return f.OnGetElementEventWorkflowsByFilters(ctx, elementID, workflowID, eventName)
}

func (f *fkElementEventWorkflowRepo) UpdateElementEventWorkflow(ctx context.Context, id string, eew *models.ElementEventWorkflow) (*models.ElementEventWorkflow, error) {
	f.Calls = append(f.Calls, "UpdateElementEventWorkflow")
	if f.OnUpdateElementEventWorkflow == nil {
		panic("fkElementEventWorkflowRepo: unexpected call to UpdateElementEventWorkflow")
	}
	return f.OnUpdateElementEventWorkflow(ctx, id, eew)
}

func (f *fkElementEventWorkflowRepo) DeleteElementEventWorkflow(ctx context.Context, id string) error {
	f.Calls = append(f.Calls, "DeleteElementEventWorkflow")
	if f.OnDeleteElementEventWorkflow == nil {
		panic("fkElementEventWorkflowRepo: unexpected call to DeleteElementEventWorkflow")
	}
	return f.OnDeleteElementEventWorkflow(ctx, id)
}

func (f *fkElementEventWorkflowRepo) DeleteElementEventWorkflowsByElementID(ctx context.Context, elementID string) error {
	f.Calls = append(f.Calls, "DeleteElementEventWorkflowsByElementID")
	if f.OnDeleteElementEventWorkflowsByElementID == nil {
		panic("fkElementEventWorkflowRepo: unexpected call to DeleteElementEventWorkflowsByElementID")
	}
	return f.OnDeleteElementEventWorkflowsByElementID(ctx, elementID)
}

func (f *fkElementEventWorkflowRepo) DeleteElementEventWorkflowsByWorkflowID(ctx context.Context, workflowID string) error {
	f.Calls = append(f.Calls, "DeleteElementEventWorkflowsByWorkflowID")
	if f.OnDeleteElementEventWorkflowsByWorkflowID == nil {
		panic("fkElementEventWorkflowRepo: unexpected call to DeleteElementEventWorkflowsByWorkflowID")
	}
	return f.OnDeleteElementEventWorkflowsByWorkflowID(ctx, workflowID)
}

func (f *fkElementEventWorkflowRepo) GetElementEventWorkflowsByPageID(ctx context.Context, pageID string) ([]models.ElementEventWorkflow, error) {
	f.Calls = append(f.Calls, "GetElementEventWorkflowsByPageID")
	if f.OnGetElementEventWorkflowsByPageID == nil {
		panic("fkElementEventWorkflowRepo: unexpected call to GetElementEventWorkflowsByPageID")
	}
	return f.OnGetElementEventWorkflowsByPageID(ctx, pageID)
}

func (f *fkElementEventWorkflowRepo) CheckIfWorkflowLinkedToElement(ctx context.Context, elementID string, workflowID string, eventName string) (bool, error) {
	f.Calls = append(f.Calls, "CheckIfWorkflowLinkedToElement")
	if f.OnCheckIfWorkflowLinkedToElement == nil {
		panic("fkElementEventWorkflowRepo: unexpected call to CheckIfWorkflowLinkedToElement")
	}
	return f.OnCheckIfWorkflowLinkedToElement(ctx, elementID, workflowID, eventName)
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
	Calls               []string
	OnGetUserByID       func(ctx context.Context, userID string) (*models.User, error)
	OnGetUserByEmail    func(ctx context.Context, email string) (*models.User, error)
	OnGetUserByUsername func(ctx context.Context, username string) (*models.User, error)
	OnSearchUsers       func(ctx context.Context, query string) ([]models.User, error)
}

func (f *fkUserRepo) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	f.Calls = append(f.Calls, "GetUserByID")
	if f.OnGetUserByID == nil {
		panic("fkUserRepo: unexpected call to GetUserByID")
	}
	return f.OnGetUserByID(ctx, userID)
}

func (f *fkUserRepo) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	f.Calls = append(f.Calls, "GetUserByEmail")
	if f.OnGetUserByEmail == nil {
		panic("fkUserRepo: unexpected call to GetUserByEmail")
	}
	return f.OnGetUserByEmail(ctx, email)
}

func (f *fkUserRepo) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	f.Calls = append(f.Calls, "GetUserByUsername")
	if f.OnGetUserByUsername == nil {
		panic("fkUserRepo: unexpected call to GetUserByUsername")
	}
	return f.OnGetUserByUsername(ctx, username)
}

func (f *fkUserRepo) SearchUsers(ctx context.Context, query string) ([]models.User, error) {
	f.Calls = append(f.Calls, "SearchUsers")
	if f.OnSearchUsers == nil {
		panic("fkUserRepo: unexpected call to SearchUsers")
	}
	return f.OnSearchUsers(ctx, query)
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
