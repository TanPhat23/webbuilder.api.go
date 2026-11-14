package unit

import (
	"context"
	"testing"

	"my-go-app/internal/realtime/domain/entity"
	"my-go-app/internal/realtime/repository"
)

func TestPersistPageReconcilesElementsAndScopesIDs(t *testing.T) {
	_, db := legacyRealtimeDatabase(t)
	repo := repository.NewElementRepository(db)
	ctx := context.Background()
	parent := "parent"
	elements := []entity.Element{{ID: parent, Type: "Frame"}, {ID: "child", Type: "Text", ParentID: &parent}}
	if err := repo.UpdateElements(ctx, "proj", "page", elements); err != nil {
		t.Fatal(err)
	}
	var count int64
	if err := db.Table(`"Element"`).Where(`"PageId" = ? AND "Id" IN ?`, "page", []string{"parent", "child"}).Count(&count).Error; err != nil || count != 2 {
		t.Fatalf("elements not stored on page: %d %v", count, err)
	}
	if err := repo.UpdateElements(ctx, "proj", "page", elements[:1]); err != nil {
		t.Fatal(err)
	}
	if err := db.Table(`"Element"`).Where(`"Id" = ?`, "child").Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("removed element retained: %d %v", count, err)
	}
	if err := db.Exec(`INSERT INTO "Element" ("Id", "Type", "PageId") VALUES ('other', 'Text', 'other-page')`).Error; err != nil {
		t.Fatal(err)
	}
	if err := repo.UpdateElements(ctx, "proj", "page", []entity.Element{{ID: "other", Type: "Text"}}); err == nil {
		t.Fatal("must reject an element ID belonging to another page")
	}
	var page string
	if err := db.Table(`"Element"`).Select(`"PageId"`).Where(`"Id" = 'other'`).Scan(&page).Error; err != nil || page != "other-page" {
		t.Fatalf("other page changed: %q %v", page, err)
	}
}
