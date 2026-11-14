package unit

import (
	"net/url"
	"os"
	"testing"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"my-go-app/internal/realtime/domain/entity"
	"my-go-app/internal/realtime/pkg/database"
)

// Every test uses its own schema; existing databases and other tests are untouched.
func legacyRealtimeDatabase(t *testing.T) (string, *gorm.DB) {
	t.Helper()
	dsn := os.Getenv("WEBBUILDER_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set WEBBUILDER_TEST_DATABASE_URL to run PostgreSQL schema regressions")
	}
	admin, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	schema := "ws_regression_" + uuid.NewString()[:8]
	if err := admin.Exec(`CREATE SCHEMA "` + schema + `"`).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := admin.Exec(`DROP SCHEMA "` + schema + `" CASCADE`).Error; err != nil {
			t.Error(err)
		}
		sqlDB, _ := admin.DB()
		_ = sqlDB.Close()
	})
	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	query := parsed.Query()
	query.Set("search_path", schema)
	parsed.RawQuery = query.Encode()
	testDSN := parsed.String()
	db, err := gorm.Open(postgres.Open(testDSN), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { sqlDB, _ := db.DB(); _ = sqlDB.Close() })
	// "OrderKey" as goose's 00001_baseline.sql creates it: nullable, no default.
	if err := db.Exec(`CREATE TABLE "Element" (
  "Id" text PRIMARY KEY, "Name" text, "Type" text NOT NULL, "Content" text,
  "Href" text, "Src" text, "Styles" jsonb, "Settings" jsonb, "TailwindStyles" text,
  "Order" integer NOT NULL DEFAULT 0, "IsLocked" boolean NOT NULL DEFAULT false,
  "ParentId" text, "PageId" text NOT NULL, "OrderKey" text
 )`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO "Element" ("Id", "Type", "Content", "PageId") VALUES (?, 'Text', 'saved', 'page')`, testElemID).Error; err != nil {
		t.Fatal(err)
	}
	return testDSN, db
}

// Goose owns the schema: connecting must not run DDL, and a NULL "OrderKey"
// (rows the API inserted without one) loads as unordered.
func TestPostgresStartupDoesNotChangeSchema(t *testing.T) {
	dsn, existing := legacyRealtimeDatabase(t)
	db, err := database.NewPostgresConnection(database.Config{URL: dsn})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	var element entity.Element
	if err := db.First(&element, `"Id" = ?`, testElemID).Error; err != nil {
		t.Fatalf("elements with a NULL order key must load: %v", err)
	}
	if element.Content == nil || *element.Content != "saved" || element.OrderKey != "" {
		t.Fatalf("element changed: %+v", element)
	}
	var nullable string
	if err := existing.Raw(`SELECT is_nullable FROM information_schema.columns WHERE table_schema = current_schema() AND table_name = 'Element' AND column_name = 'OrderKey'`).Scan(&nullable).Error; err != nil {
		t.Fatal(err)
	}
	if nullable != "YES" {
		t.Fatalf("startup altered the goose-owned OrderKey column (is_nullable=%q)", nullable)
	}
	if err := existing.Exec(`ALTER TABLE "Element" DROP COLUMN "OrderKey"`).Error; err != nil {
		t.Fatal(err)
	}
	again, err := database.NewPostgresConnection(database.Config{URL: dsn})
	if err != nil {
		t.Fatal(err)
	}
	againSQL, _ := again.DB()
	defer againSQL.Close()
	if again.Migrator().HasColumn("Element", "OrderKey") {
		t.Fatal("startup added a column; schema changes belong in goose migrations")
	}
}
