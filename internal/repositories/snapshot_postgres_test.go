package repositories

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"my-go-app/internal/models"
)

// Snapshot rows have a CreatedAt column but no UpdatedAt. Saving one used to
// fail with `column "UpdatedAt" of relation "Snapshot" does not exist` because
// the model embedded the audit fields that include UpdatedAt.
// UUID_GORM_TEST_DSN must identify a disposable database with the migrations.
func TestPostgresSnapshotSaveMatchesTheTable(t *testing.T) {
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
	ownerID, projectID := "owner-"+suffix, "p-"+suffix
	if err := db.Exec(`INSERT INTO "User" ("Id", "Email", "CreatedAt", "UpdatedAt") VALUES (?, ?, NOW(), NOW())`, ownerID, ownerID+"@example.invalid").Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Exec(`DELETE FROM "User" WHERE "Id" = ?`, ownerID) })
	t.Cleanup(func() { db.Exec(`DELETE FROM "Project" WHERE "Id" = ?`, projectID) })
	if err := NewProjectRepository(db).CreateProjectWithPage(ctx, &models.Project{ID: projectID, Name: "Project", OwnerId: ownerID},
		&models.Page{Id: "page-" + suffix, Name: "Home", Type: "page", Styles: json.RawMessage(`{}`)}); err != nil {
		t.Fatal(err)
	}

	repo := NewSnapshotRepository(db)
	save := func(id, typ, name string) {
		t.Helper()
		s := &models.Snapshot{Id: id, Name: name, Type: typ, Timestamp: 1, Elements: json.RawMessage(`[]`)}
		if err := repo.SaveSnapshot(ctx, projectID, s); err != nil {
			t.Fatalf("saving a %s snapshot: %v", typ, err)
		}
	}
	save("w1-"+suffix, SnapshotTypeWorking, "first")  // creates
	save("w2-"+suffix, SnapshotTypeWorking, "second") // updates the existing working snapshot
	save("v1-"+suffix, SnapshotTypeVersion, "v1")

	var n int64
	db.Model(&models.Snapshot{}).Where(`"ProjectId" = ?`, projectID).Count(&n)
	if n != 2 {
		t.Fatalf("want one working snapshot (updated in place) and one version, got %d rows", n)
	}
}
