package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"my-go-app/internal/models"
)

// UUID_GORM_TEST_DSN must identify a disposable database with the baseline migration.
func TestPostgresProjectRepositoryParity(t *testing.T) {
	dsn := os.Getenv("UUID_GORM_TEST_DSN")
	if dsn == "" {
		t.Skip("requires dedicated migrated PostgreSQL database")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { sqlDB.Close() })
	ctx := context.Background()
	suffix := uuid.NewString()
	ownerID := "owner-" + suffix
	if err := db.Exec(`INSERT INTO "User" ("Id", "Email", "CreatedAt", "UpdatedAt") VALUES (?, ?, NOW(), NOW())`, ownerID, ownerID+"@example.invalid").Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Exec(`DELETE FROM "User" WHERE "Id" = ?`, ownerID) })
	projectID := "p-" + suffix
	pageID := "page-" + suffix
	t.Cleanup(func() { db.Exec(`DELETE FROM "Project" WHERE "Id" IN (?, ?)`, projectID, "rollback-"+suffix) })
	project := &models.Project{ID: projectID, Name: "Project", OwnerId: ownerID}
	page := &models.Page{Id: pageID, Name: "Home", Type: "page", Styles: json.RawMessage(`{}`)}
	repo := NewProjectRepository(db)
	if err := repo.CreateProjectWithPage(ctx, project, page); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.GetProjectByID(ctx, projectID, "other"); !errors.Is(err, ErrProjectNotFound) {
		t.Fatalf("owner isolation: %v", err)
	}
	if got, err := repo.GetProjectByID(ctx, projectID, ownerID); err != nil || got.ID != projectID {
		t.Fatalf("owner lookup: %+v %v", got, err)
	}
	if err := db.Exec(`UPDATE "Project" SET "Published" = true, "Description" = 'present' WHERE "Id" = ?`, projectID).Error; err != nil {
		t.Fatal(err)
	}
	updated, err := repo.UpdateProject(ctx, projectID, ownerID, map[string]any{"Published": false, "Description": nil})
	if err != nil || updated.Published || updated.Description != nil {
		t.Fatalf("zero/null update: %+v %v", updated, err)
	}
	if _, err := repo.UpdateProject(ctx, projectID, "other", map[string]any{"Published": true}); !errors.Is(err, ErrProjectUnauthorized) {
		t.Fatalf("update ownership: %v", err)
	}
	pages := NewPageRepository(db)
	if _, err := pages.GetPageByID(ctx, pageID, "other"); !errors.Is(err, ErrPageNotFound) {
		t.Fatalf("project isolation: %v", err)
	}
	if _, err := pages.GetPageByID(ctx, pageID, projectID); err != nil {
		t.Fatal(err)
	}
	elementRepo := NewElementRepository(db)
	elementID := "element-" + suffix
	if err := elementRepo.ReplaceElements(ctx, projectID, []models.EditorElement{&models.Element{Id: elementID, Type: "Text", PageId: &pageID}}); err != nil {
		t.Fatal(err)
	}
	element, err := elementRepo.GetElementWithRelations(ctx, elementID)
	if err != nil || element.Page == nil || element.Page.Id != pageID {
		t.Fatalf("element batch insert/preload: %+v %v", element, err)
	}
	if err := repo.HardDeleteProject(ctx, projectID, "other"); !errors.Is(err, ErrProjectUnauthorized) {
		t.Fatalf("hard delete owner isolation: %v", err)
	}
	snapshotID := "snapshot-" + suffix
	if err := db.Exec(`INSERT INTO "Snapshot" ("Id", "ProjectId", "Elements", "Timestamp", "Name", "Type") VALUES (?, ?, '[]', 1, 'Draft', 'working')`, snapshotID, projectID).Error; err != nil {
		t.Fatal(err)
	}
	snapshots := &SnapshotRepository{db: db}
	foundSnapshots, err := snapshots.GetSnapshotsByProjectID(ctx, projectID)
	if err != nil || len(foundSnapshots) != 1 || foundSnapshots[0].Id != snapshotID {
		t.Fatalf("snapshot read: %+v %v", foundSnapshots, err)
	}
	if n, err := snapshots.CountSnapshotsByProjectID(ctx, projectID); err != nil || n != 1 {
		t.Fatalf("snapshot count: %d %v", n, err)
	}
	if err := snapshots.DeleteSnapshot(ctx, snapshotID); err != nil {
		t.Fatal(err)
	}
	if err := snapshots.DeleteSnapshot(ctx, snapshotID); !errors.Is(err, ErrSnapshotNotFound) {
		t.Fatalf("snapshot delete missing: %v", err)
	}
	contentID := "content-" + suffix
	if err := db.Exec(`INSERT INTO "ContentType" ("Id", "Name", "UpdatedAt") VALUES (?, ?, NOW())`, contentID, contentID).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Exec(`DELETE FROM "ContentType" WHERE "Id" = ?`, contentID) })
	fields := NewContentFieldRepository(db)
	field, err := fields.CreateContentField(ctx, &models.ContentField{ContentTypeId: contentID, Name: "Field", Type: "text"})
	if err != nil {
		t.Fatal(err)
	}
	if found, err := fields.GetContentFieldByID(ctx, field.Id); err != nil || found.ContentTypeId != contentID {
		t.Fatalf("content field read: %+v %v", found, err)
	}
	if _, err := fields.CreateContentField(ctx, &models.ContentField{ContentTypeId: contentID, Name: "Field", Type: "text"}); !errors.Is(err, ErrContentFieldDuplicate) {
		t.Fatalf("duplicate field: %v", err)
	}
	notifications := NewNotificationRepository(db)
	note := &models.Notification{UserId: ownerID, Type: "alerts", Title: "Test", Description: "content"}
	if err := notifications.Create(ctx, note); err != nil {
		t.Fatal(err)
	}
	if _, err := notifications.SetRead(ctx, "other", note.Id, true); !errors.Is(err, ErrNotificationNotFound) {
		t.Fatalf("notification owner isolation: %v", err)
	}
	if _, err := notifications.SetRead(ctx, ownerID, note.Id, true); err != nil {
		t.Fatal(err)
	}
	list, err := notifications.ListForUser(ctx, ownerID, models.NotificationFilter{})
	if err != nil || len(list.Notifications) != 1 || !list.Notifications[0].Read {
		t.Fatalf("notification read: %+v %v", list, err)
	}
	bad := &models.Project{ID: "rollback-" + suffix, Name: "Rollback", OwnerId: ownerID}
	if err := repo.CreateProjectWithPage(ctx, bad, &models.Page{Id: pageID, Name: "Duplicate", Type: "page", Styles: json.RawMessage(`{}`)}); err == nil {
		t.Fatal("expected duplicate page failure")
	}
	if _, err := repo.GetProjectByID(ctx, bad.ID, ownerID); !errors.Is(err, ErrProjectNotFound) {
		t.Fatalf("project not rolled back: %v", err)
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := repo.GetProjectByID(cancelled, projectID, ownerID); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled query: %v", err)
	}
}
