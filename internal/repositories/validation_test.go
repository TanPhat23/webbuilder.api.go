package repositories

import (
	"context"
	"strings"
	"testing"
	"time"

	"my-go-app/internal/models"
)

type repoValidationCase struct {
	name string
	want string
	call func() error
}

func assertRepoValidationCases(t *testing.T, fake *fakeDB, tests []repoValidationCase) {
	t.Helper()
	for _, tt := range tests {
		before := fake.queryCount()
		err := tt.call()
		if err == nil {
			t.Fatalf("%s: expected an error containing %q, got nil", tt.name, tt.want)
		}
		if !strings.Contains(err.Error(), tt.want) {
			t.Fatalf("%s: error = %q, want it to contain %q", tt.name, err, tt.want)
		}
		if n := fake.queryCount() - before; n != 0 {
			t.Fatalf("%s: validation must fail before querying, but %d statement(s) ran: %s", tt.name, n, fake.lastQuery())
		}
	}
}

func TestContentRepositoriesRejectInvalidArgsWithoutQuerying(t *testing.T) {
	ctx := context.Background()
	db, fake := newFakeDB(t)
	contentTypes := &ContentTypeRepository{db: db}
	contentFields := &ContentFieldRepository{db: db}
	contentItems := &ContentItemRepository{db: db}
	contentValues := &ContentFieldValueRepository{db: db}

	assertRepoValidationCases(t, fake, []repoValidationCase{
		{"ContentType get by id", "id is required", func() error { _, err := contentTypes.GetByID(ctx, ""); return err }},
		{"ContentType get by name", "name is required", func() error { _, err := contentTypes.GetByName(ctx, ""); return err }},
		{"ContentType create nil", "cannot be nil", func() error { _, err := contentTypes.Create(ctx, nil); return err }},
		{"ContentType create without name", "name is required", func() error { _, err := contentTypes.Create(ctx, &models.ContentType{}); return err }},
		{"ContentType update without id", "id is required", func() error { _, err := contentTypes.Update(ctx, "", map[string]any{"a": 1}); return err }},
		{"ContentType update without changes", "no updates provided", func() error { _, err := contentTypes.Update(ctx, "id", nil); return err }},
		{"ContentType delete", "id is required", func() error { return contentTypes.Delete(ctx, "") }},
		{"ContentType exists by name", "name is required", func() error { _, err := contentTypes.ExistsByName(ctx, ""); return err }},
		{"ContentType exists by id", "id is required", func() error { _, err := contentTypes.ExistsByID(ctx, ""); return err }},
		{"ContentField list", "contentTypeID is required", func() error { _, err := contentFields.ListByContentType(ctx, ""); return err }},
		{"ContentField get", "id is required", func() error { _, err := contentFields.GetByID(ctx, ""); return err }},
		{"ContentField create nil", "cannot be nil", func() error { _, err := contentFields.Create(ctx, nil); return err }},
		{"ContentField update without id", "id is required", func() error { _, err := contentFields.Update(ctx, "", map[string]any{"a": 1}); return err }},
		{"ContentField update without changes", "no updates provided", func() error { _, err := contentFields.Update(ctx, "id", map[string]any{}); return err }},
		{"ContentField delete", "id is required", func() error { return contentFields.Delete(ctx, "") }},
		{"ContentItem list", "contentTypeID is required", func() error { _, err := contentItems.ListByContentType(ctx, ""); return err }},
		{"ContentItem get", "id is required", func() error { _, err := contentItems.GetByID(ctx, ""); return err }},
		{"ContentItem create nil", "cannot be nil", func() error { _, err := contentItems.Create(ctx, nil, nil); return err }},
		{"ContentItem create without title", "title is required", func() error {
			_, err := contentItems.Create(ctx, &models.ContentItem{Slug: "s", ContentTypeId: "ct"}, nil)
			return err
		}},
		{"ContentItem create without slug", "slug is required", func() error {
			_, err := contentItems.Create(ctx, &models.ContentItem{Title: "t", ContentTypeId: "ct"}, nil)
			return err
		}},
		{"ContentItem create without type", "content type ID is required", func() error {
			_, err := contentItems.Create(ctx, &models.ContentItem{Title: "t", Slug: "s"}, nil)
			return err
		}},
		{"ContentItem update without id", "id is required", func() error { _, err := contentItems.Update(ctx, "", map[string]any{"a": 1}, nil); return err }},
		{"ContentItem update without changes", "no updates provided", func() error { _, err := contentItems.Update(ctx, "id", nil, nil); return err }},
		{"ContentItem delete", "id is required", func() error { return contentItems.Delete(ctx, "") }},
		{"ContentItem public list", "contentTypeID is required", func() error { _, err := contentItems.ListPublic(ctx, "", 1, "", ""); return err }},
		{"ContentItem by slug without type", "required", func() error { _, err := contentItems.GetBySlug(ctx, "", "slug"); return err }},
		{"ContentItem by slug without slug", "required", func() error { _, err := contentItems.GetBySlug(ctx, "ct", ""); return err }},
		{"ContentFieldValue list", "contentItemID is required", func() error { _, err := contentValues.ListByContentItem(ctx, ""); return err }},
		{"ContentFieldValue get", "id is required", func() error { _, err := contentValues.GetByID(ctx, ""); return err }},
		{"ContentFieldValue create nil", "cannot be nil", func() error { _, err := contentValues.Create(ctx, nil); return err }},
		{"ContentFieldValue create without item", "content item ID is required", func() error { _, err := contentValues.Create(ctx, &models.ContentFieldValue{FieldId: "f"}); return err }},
		{"ContentFieldValue create without field", "field ID is required", func() error {
			_, err := contentValues.Create(ctx, &models.ContentFieldValue{ContentItemId: "i"})
			return err
		}},
		{"ContentFieldValue update", "id is required", func() error { _, err := contentValues.Update(ctx, "", nil); return err }},
		{"ContentFieldValue delete", "id is required", func() error { return contentValues.Delete(ctx, "") }},
	})
}

func TestElementAndCustomRepositoriesRejectInvalidArgsWithoutQuerying(t *testing.T) {
	ctx := context.Background()
	db, fake := newFakeDB(t)
	customElements := &CustomElementRepository{db: db}
	elementTypes := &CustomElementTypeRepository{db: db}
	elements := &ElementRepository{db: db}
	comments := &ElementCommentRepository{db: db}

	assertRepoValidationCases(t, fake, []repoValidationCase{
		{"CustomElement list", "userID is required", func() error { _, err := customElements.List(ctx, "", nil); return err }},
		{"CustomElement get without id", "id is required", func() error { _, err := customElements.GetByID(ctx, "", "u"); return err }},
		{"CustomElement get without user", "userID is required", func() error { _, err := customElements.GetByID(ctx, "id", ""); return err }},
		{"CustomElement create nil", "customElement is required", func() error { _, err := customElements.Create(ctx, nil); return err }},
		{"CustomElement create without user", "userId is required", func() error { _, err := customElements.Create(ctx, &models.CustomElement{Name: "n"}); return err }},
		{"CustomElement create without name", "name is required", func() error { _, err := customElements.Create(ctx, &models.CustomElement{UserId: "u"}); return err }},
		{"CustomElement update without id", "id is required", func() error { _, err := customElements.Update(ctx, "", "u", nil); return err }},
		{"CustomElement update without user", "userID is required", func() error { _, err := customElements.Update(ctx, "id", "", nil); return err }},
		{"CustomElement delete without id", "id is required", func() error { return customElements.Delete(ctx, "", "u") }},
		{"CustomElement delete without user", "userID is required", func() error { return customElements.Delete(ctx, "id", "") }},
		{"CustomElement duplicate without id", "id is required", func() error { _, err := customElements.Duplicate(ctx, "", "u", "n"); return err }},
		{"CustomElement duplicate without user", "userID is required", func() error { _, err := customElements.Duplicate(ctx, "id", "", "n"); return err }},
		{"CustomElement duplicate without name", "newName is required", func() error { _, err := customElements.Duplicate(ctx, "id", "u", ""); return err }},
		{"CustomElementType get by id", "id is required", func() error { _, err := elementTypes.GetByID(ctx, ""); return err }},
		{"CustomElementType get by name", "name is required", func() error { _, err := elementTypes.GetByName(ctx, ""); return err }},
		{"CustomElementType create nil", "customElementType is required", func() error { _, err := elementTypes.Create(ctx, nil); return err }},
		{"CustomElementType create without name", "name is required", func() error { _, err := elementTypes.Create(ctx, &models.CustomElementType{}); return err }},
		{"CustomElementType update", "id is required", func() error { _, err := elementTypes.Update(ctx, "", nil); return err }},
		{"CustomElementType delete", "id is required", func() error { return elementTypes.Delete(ctx, "") }},
		{"Element list by project", "projectID is required", func() error { _, err := elements.List(ctx, ""); return err }},
		{"Element get", "elementID is required", func() error { _, err := elements.GetByID(ctx, ""); return err }},
		{"Element get with relations", "elementID is required", func() error { _, err := elements.GetWithRelations(ctx, ""); return err }},
		{"Element by page", "pageID is required", func() error { _, err := elements.ListByPage(ctx, ""); return err }},
		{"Element by pages without project", "projectID is required", func() error { _, err := elements.ListByPageIDs(ctx, "", []string{"p"}); return err }},
		{"Element by pages without pages", "pageIDs is required", func() error { _, err := elements.ListByPageIDs(ctx, "proj", nil); return err }},
		{"Element children", "parentID is required", func() error { _, err := elements.ListChildren(ctx, ""); return err }},
		{"Element roots", "projectID is required", func() error { _, err := elements.ListRoots(ctx, ""); return err }},
		{"Element create nil", "cannot be nil", func() error { return elements.Create(ctx, nil) }},
		{"Element create without type", "type is required", func() error { return elements.Create(ctx, &models.Element{Id: "e"}) }},
		{"Element update nil", "cannot be nil", func() error { return elements.Update(ctx, nil) }},
		{"Element update without id", "element ID is required", func() error { return elements.Update(ctx, &models.Element{}) }},
		{"Element update workflows", "elementID is required", func() error { return elements.UpdateEventWorkflows(ctx, "", nil) }},
		{"Element delete", "elementID is required", func() error { return elements.Delete(ctx, "") }},
		{"Element delete by page", "pageID is required", func() error { return elements.DeleteByPage(ctx, "") }},
		{"Element delete by project", "projectID is required", func() error { return elements.DeleteByProject(ctx, "") }},
		{"Element count by project", "projectID is required", func() error { _, err := elements.CountByProject(ctx, ""); return err }},
		{"Element replace", "projectID is required", func() error { return elements.Replace(ctx, "", nil) }},
		{"Comment create nil", "cannot be nil", func() error { _, err := comments.Create(ctx, nil); return err }},
		{"Comment get", "comment id is required", func() error { _, err := comments.GetByID(ctx, ""); return err }},
		{"Comment list", "elementId is required", func() error { _, err := comments.List(ctx, "", nil); return err }},
		{"Comment update without id", "comment id is required", func() error { _, err := comments.Update(ctx, "", map[string]any{"a": 1}); return err }},
		{"Comment update without changes", "no updates provided", func() error { _, err := comments.Update(ctx, "id", nil); return err }},
		{"Comment delete", "comment id is required", func() error { return comments.Delete(ctx, "") }},
		{"Comment by author", "authorId is required", func() error { _, err := comments.ListByAuthor(ctx, "", 1, 0); return err }},
		{"Comment count", "elementId is required", func() error { _, err := comments.Count(ctx, ""); return err }},
		{"Comment toggle", "comment id is required", func() error { _, err := comments.ToggleResolved(ctx, ""); return err }},
		{"Comment delete by element", "elementId is required", func() error { return comments.DeleteByElement(ctx, "") }},
		{"Comment by project", "projectId is required", func() error { _, err := comments.ListByProject(ctx, "", 1, 0); return err }},
		{"Comment count by project", "projectId is required", func() error { _, err := comments.CountByProject(ctx, ""); return err }},
	})
}

func TestMediaInvitationNotificationRepositoriesRejectInvalidArgs(t *testing.T) {
	ctx := context.Background()
	db, fake := newFakeDB(t)
	images := &ImageRepository{db: db}
	invitations := &InvitationRepository{db: db}
	notifications := &NotificationRepository{db: db}

	assertRepoValidationCases(t, fake, []repoValidationCase{
		{"Image list", "userID is required", func() error { _, err := images.ListByUser(ctx, ""); return err }},
		{"Image get without id", "imageID is required", func() error { _, err := images.GetByID(ctx, "", "u"); return err }},
		{"Image get without user", "userID is required", func() error { _, err := images.GetByID(ctx, "i", ""); return err }},
		{"Image delete without id", "imageID is required", func() error { return images.Delete(ctx, "", "u") }},
		{"Image delete without user", "userID is required", func() error { return images.Delete(ctx, "i", "") }},
		{"Image soft delete without id", "imageID is required", func() error { return images.SoftDelete(ctx, "", "u") }},
		{"Image soft delete without user", "userID is required", func() error { return images.SoftDelete(ctx, "i", "") }},
		{"Invitation create nil", "cannot be nil", func() error { _, err := invitations.Create(ctx, nil); return err }},
		{"Invitation by project", "projectID is required", func() error { _, err := invitations.ListByProject(ctx, ""); return err }},
		{"Invitation by id", "id is required", func() error { _, err := invitations.GetByID(ctx, ""); return err }},
		{"Invitation by token", "token is required", func() error { _, err := invitations.GetByToken(ctx, ""); return err }},
		{"Invitation accept without token", "token and userID are required", func() error { return invitations.Accept(ctx, "", "u") }},
		{"Invitation accept without user", "token and userID are required", func() error { return invitations.Accept(ctx, "t", "") }},
		{"Invitation delete", "id is required", func() error { return invitations.Delete(ctx, "") }},
		{"Invitation update status", "id is required", func() error { return invitations.UpdateStatus(ctx, "", models.InvitationStatusExpired) }},
		{"Invitation cancel", "id is required", func() error { return invitations.Cancel(ctx, "") }},
		{"Invitation pending by project", "projectID is required", func() error { _, err := invitations.ListPendingByProject(ctx, ""); return err }},
		{"Notification list", "userID is required", func() error { _, err := notifications.ListForUser(ctx, "", models.NotificationFilter{}); return err }},
		{"Notification create without user", "required", func() error { return notifications.Create(ctx, &models.Notification{Type: "alert", Title: "t"}) }},
		{"Notification create without type", "required", func() error { return notifications.Create(ctx, &models.Notification{UserId: "u", Title: "t"}) }},
		{"Notification create without title", "required", func() error { return notifications.Create(ctx, &models.Notification{UserId: "u", Type: "alert"}) }},
	})
}

func TestPageAndProjectRepositoriesRejectInvalidArgs(t *testing.T) {
	ctx := context.Background()
	db, fake := newFakeDB(t)
	pages := &PageRepository{db: db}
	projects := &ProjectRepository{db: db}

	assertRepoValidationCases(t, fake, []repoValidationCase{
		{"Page list", "projectID is required", func() error { _, err := pages.ListByProject(ctx, ""); return err }},
		{"Page get", "pageID is required", func() error { _, err := pages.GetByID(ctx, "", "proj"); return err }},
		{"Page create nil", "cannot be nil", func() error { return pages.Create(ctx, nil) }},
		{"Page create without name", "page name is required", func() error { return pages.Create(ctx, &models.Page{ProjectId: "p", Type: "page"}) }},
		{"Page create without project", "project ID is required", func() error { return pages.Create(ctx, &models.Page{Name: "n", Type: "page"}) }},
		{"Page create without type", "page type is required", func() error { return pages.Create(ctx, &models.Page{Name: "n", ProjectId: "p"}) }},
		{"Page update nil", "cannot be nil", func() error { return pages.Update(ctx, nil) }},
		{"Page update without id", "page ID is required", func() error { return pages.Update(ctx, &models.Page{}) }},
		{"Page update fields without id", "pageID is required", func() error { return pages.UpdateFields(ctx, "", map[string]any{"a": 1}) }},
		{"Page update fields without changes", "no updates provided", func() error { return pages.UpdateFields(ctx, "p", nil) }},
		{"Page delete", "pageID is required", func() error { return pages.Delete(ctx, "") }},
		{"Page delete by project without page", "required", func() error { return pages.DeleteForUser(ctx, "", "p", "u") }},
		{"Page delete by project without project", "required", func() error { return pages.DeleteForUser(ctx, "pg", "", "u") }},
		{"Page delete by project without user", "required", func() error { return pages.DeleteForUser(ctx, "pg", "p", "") }},
		{"Page hard delete", "pageID is required", func() error { return pages.HardDelete(ctx, "") }},
		{"Page restore", "pageID is required", func() error { return pages.Restore(ctx, "") }},
		{"Page exists without page", "required", func() error { _, err := pages.ExistsInProject(ctx, "", "p"); return err }},
		{"Page exists without project", "required", func() error { _, err := pages.ExistsInProject(ctx, "pg", ""); return err }},
		{"Page count", "projectID is required", func() error { _, err := pages.CountByProject(ctx, ""); return err }},
		{"Project get without project", "required", func() error { _, err := projects.GetByID(ctx, "", "u"); return err }},
		{"Project get without user", "required", func() error { _, err := projects.GetByID(ctx, "p", ""); return err }},
		{"Project access without project", "required", func() error { _, err := projects.GetWithAccess(ctx, "", "u"); return err }},
		{"Project access without user", "required", func() error { _, err := projects.GetWithAccess(ctx, "p", ""); return err }},
		{"Project public get", "projectID is required", func() error { _, err := projects.GetPublicByID(ctx, ""); return err }},
		{"Project list", "userID is required", func() error { _, err := projects.ListByUser(ctx, ""); return err }},
		{"Project collaborator list", "userID is required", func() error { _, err := projects.ListByCollaborator(ctx, ""); return err }},
		{"Project pages without project", "required", func() error { _, err := projects.ListPages(ctx, "", "u"); return err }},
		{"Project pages without user", "required", func() error { _, err := projects.ListPages(ctx, "p", ""); return err }},
		{"Project create nil", "cannot be nil", func() error { return projects.Create(ctx, nil) }},
		{"Project update without project", "required", func() error { _, err := projects.Update(ctx, "", "u", map[string]any{"a": 1}); return err }},
		{"Project update without user", "required", func() error { _, err := projects.Update(ctx, "p", "", map[string]any{"a": 1}); return err }},
		{"Project update without changes", "no updates provided", func() error { _, err := projects.Update(ctx, "p", "u", nil); return err }},
		{"Project delete without project", "required", func() error { return projects.Delete(ctx, "", "u") }},
		{"Project delete without user", "required", func() error { return projects.Delete(ctx, "p", "") }},
		{"Project hard delete without project", "required", func() error { return projects.HardDelete(ctx, "", "u") }},
		{"Project hard delete without user", "required", func() error { return projects.HardDelete(ctx, "p", "") }},
		{"Project restore without project", "required", func() error { return projects.Restore(ctx, "", "u") }},
		{"Project restore without user", "required", func() error { return projects.Restore(ctx, "p", "") }},
		{"Project exists without project", "required", func() error { _, err := projects.ExistsForUser(ctx, "", "u"); return err }},
		{"Project exists without user", "required", func() error { _, err := projects.ExistsForUser(ctx, "p", ""); return err }},
		{"Project lock without project", "required", func() error { _, err := projects.GetWithLock(ctx, "", "u"); return err }},
		{"Project lock without user", "required", func() error { _, err := projects.GetWithLock(ctx, "p", ""); return err }},
		{"Project with page nil project", "required", func() error { return projects.CreateWithPage(ctx, nil, &models.Page{}) }},
		{"Project with page nil page", "required", func() error { return projects.CreateWithPage(ctx, &models.Project{}, nil) }},
		{"Project with page without id", "required", func() error {
			return projects.CreateWithPage(ctx, &models.Project{Name: "n", OwnerId: "o"}, &models.Page{})
		}},
		{"Project with page without name", "required", func() error {
			return projects.CreateWithPage(ctx, &models.Project{ID: "i", OwnerId: "o"}, &models.Page{})
		}},
		{"Project with page without owner", "required", func() error { return projects.CreateWithPage(ctx, &models.Project{ID: "i", Name: "n"}, &models.Page{}) }},
		{"Project access check without project", "required", func() error { _, err := projects.GetAccess(ctx, "", "u"); return err }},
		{"Project access check without user", "required", func() error { _, err := projects.GetAccess(ctx, "p", ""); return err }},
	})
}

func TestSnapshotAndUserRepositoriesRejectInvalidArgs(t *testing.T) {
	ctx := context.Background()
	db, fake := newFakeDB(t)
	snapshots := &SnapshotRepository{db: db}
	users := &UserRepository{db: db}

	assertRepoValidationCases(t, fake, []repoValidationCase{
		{"Snapshot save nil", "cannot be nil", func() error { return snapshots.Save(ctx, "p", nil) }},
		{"Snapshot save without project", "projectID is required", func() error { return snapshots.Save(ctx, "", &models.Snapshot{Type: SnapshotTypeWorking}) }},
		{"Snapshot save without type", "type is required", func() error { return snapshots.Save(ctx, "p", &models.Snapshot{}) }},
		{"Snapshot save invalid type", ErrInvalidSnapshotType.Error(), func() error { return snapshots.Save(ctx, "p", &models.Snapshot{Type: "draft"}) }},
		{"Snapshot list", "projectID is required", func() error { _, err := snapshots.ListByProject(ctx, ""); return err }},
		{"Snapshot list by type without project", "projectID is required", func() error { _, err := snapshots.ListByProjectAndType(ctx, "", SnapshotTypeVersion); return err }},
		{"Snapshot list by type without type", "type is required", func() error { _, err := snapshots.ListByProjectAndType(ctx, "p", ""); return err }},
		{"Snapshot list by invalid type", ErrInvalidSnapshotType.Error(), func() error { _, err := snapshots.ListByProjectAndType(ctx, "p", "draft"); return err }},
		{"Snapshot get", "snapshotID is required", func() error { _, err := snapshots.GetByID(ctx, ""); return err }},
		{"Snapshot latest working", "projectID is required", func() error { _, err := snapshots.GetLatestWorking(ctx, ""); return err }},
		{"Snapshot latest published", "projectID is required", func() error { _, err := snapshots.GetLatestPublished(ctx, ""); return err }},
		{"Snapshot delete", "snapshotID is required", func() error { return snapshots.Delete(ctx, "") }},
		{"Snapshot delete by project", "projectID is required", func() error { return snapshots.DeleteByProject(ctx, "") }},
		{"Snapshot delete old", "projectID is required", func() error { return snapshots.DeleteOlderThan(ctx, "", time.Hour) }},
		{"Snapshot count", "projectID is required", func() error { _, err := snapshots.CountByProject(ctx, ""); return err }},
		{"Snapshot count by type without project", "projectID is required", func() error { _, err := snapshots.CountByType(ctx, "", SnapshotTypeVersion); return err }},
		{"Snapshot count by type without type", "type is required", func() error { _, err := snapshots.CountByType(ctx, "p", ""); return err }},
		{"Snapshot count by invalid type", ErrInvalidSnapshotType.Error(), func() error { _, err := snapshots.CountByType(ctx, "p", "draft"); return err }},
		{"Snapshot exists", "snapshotID is required", func() error { _, err := snapshots.ExistsByID(ctx, ""); return err }},
		{"User by id", "userID is required", func() error { _, err := users.GetByID(ctx, ""); return err }},
		{"User by email", "email is required", func() error { _, err := users.GetByEmail(ctx, ""); return err }},
		{"User by username", "username is required", func() error { _, err := users.GetByUsername(ctx, ""); return err }},
	})
}

// Some lookups are deliberately lenient: an empty argument is a valid, empty answer.
func TestRepositoriesTreatEmptyInputAsEmptyResult(t *testing.T) {
	ctx := context.Background()
	db, fake := newFakeDB(t)

	found, err := (&UserRepository{db: db}).Search(ctx, "")
	if err != nil || found == nil || len(found) != 0 {
		t.Fatalf("Search(\"\") = %#v, %v; want empty non-nil slice", found, err)
	}
	got, err := (&ElementRepository{db: db}).ListByIDs(ctx, nil)
	if err != nil || got == nil || len(got) != 0 {
		t.Fatalf("ListByIDs(nil) = %#v, %v; want empty non-nil slice", got, err)
	}
	if n := fake.queryCount(); n != 0 {
		t.Fatalf("no statement should run, ran %d", n)
	}
}

func TestIsValidSnapshotType(t *testing.T) {
	for typ, want := range map[string]bool{
		SnapshotTypeWorking: true, SnapshotTypeVersion: true, SnapshotTypePublished: true,
		"": false, "Working": false, "draft": false, "working ": false,
	} {
		if got := isValidSnapshotType(typ); got != want {
			t.Errorf("isValidSnapshotType(%q) = %v, want %v", typ, got, want)
		}
	}
}

func TestNewRepositoriesWiresEveryRepository(t *testing.T) {
	db, fake := newFakeDB(t)
	if fake == nil {
		t.Fatal("missing fake db")
	}
	r := NewRepositories(db)
	checks := map[string]bool{
		"AgentToken": r.AgentTokenRepository != nil, "Element": r.ElementRepository != nil,
		"ElementComment": r.ElementCommentRepository != nil, "User": r.UserRepository != nil,
		"Project": r.ProjectRepository != nil, "Snapshot": r.SnapshotRepository != nil,
		"Page": r.PageRepository != nil, "ContentType": r.ContentTypeRepository != nil,
		"ContentField": r.ContentFieldRepository != nil, "ContentItem": r.ContentItemRepository != nil,
		"ContentFieldValue": r.ContentFieldValueRepository != nil, "Image": r.ImageRepository != nil,
		"CustomElement": r.CustomElementRepository != nil, "CustomElementType": r.CustomElementTypeRepository != nil,
		"Invitation": r.InvitationRepository != nil, "Collaborator": r.CollaboratorRepository != nil,
		"EventWorkflow": r.EventWorkflowRepository != nil, "ElementEventWorkflow": r.ElementEventWorkflowRepository != nil,
		"Notification": r.NotificationRepository != nil,
	}
	for name, ok := range checks {
		if !ok {
			t.Errorf("%sRepository is not wired", name)
		}
	}
}
