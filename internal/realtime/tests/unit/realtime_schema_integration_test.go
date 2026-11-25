package unit

import (
	"context"
	"encoding/json"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"my-go-app/internal/realtime/domain/entity"
	"my-go-app/internal/realtime/pkg/database"
	"my-go-app/internal/realtime/repository"
)

func closeSQLHandle(t *testing.T, db *gorm.DB) {
	t.Helper()
	sqlDB, err := db.DB()
	if err != nil {
		t.Errorf("db.DB: %v", err)
		return
	}
	if err := sqlDB.Close(); err != nil {
		t.Errorf("close sqlDB: %v", err)
	}
}

func openSchemaDB(t *testing.T, dsn string) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func seedLegacyElementTable(t *testing.T, db *gorm.DB) {
	t.Helper()
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
}

// Every test uses its own schema; existing databases and other tests are untouched.
func legacyRealtimeDatabase(t *testing.T) (string, *gorm.DB) {
	t.Helper()
	dsn := os.Getenv("WEBBUILDER_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set WEBBUILDER_TEST_DATABASE_URL to run PostgreSQL schema regressions")
	}
	admin := openSchemaDB(t, dsn)
	schema := "ws_regression_" + uuid.NewString()[:8]
	if err := admin.Exec(`CREATE SCHEMA "` + schema + `"`).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := admin.Exec(`DROP SCHEMA "` + schema + `" CASCADE`).Error; err != nil {
			t.Error(err)
		}
		closeSQLHandle(t, admin)
	})
	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	query := parsed.Query()
	query.Set("search_path", schema)
	parsed.RawQuery = query.Encode()
	testDSN := parsed.String()
	db := openSchemaDB(t, testDSN)
	t.Cleanup(func() { closeSQLHandle(t, db) })
	seedLegacyElementTable(t, db)
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
	t.Cleanup(func() { closeSQLHandle(t, db) })
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
	t.Cleanup(func() { closeSQLHandle(t, again) })
	if again.Migrator().HasColumn("Element", "OrderKey") {
		t.Fatal("startup added a column; schema changes belong in goose migrations")
	}
}

func TestElementUpsertPersistsChangedOrderKey(t *testing.T) {
	_, db := legacyRealtimeDatabase(t)
	if err := db.Exec(`CREATE TABLE "Page" ("Id" text PRIMARY KEY, "ProjectId" text, "DeletedAt" timestamptz);
		INSERT INTO "Page" ("Id", "ProjectId") VALUES ('page', 'project')`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`UPDATE "Element" SET "OrderKey" = 'a0'`).Error; err != nil {
		t.Fatal(err)
	}
	repo := repository.NewElementRepository(db)
	content := "saved"
	if err := repo.UpsertElements(context.Background(), "project", "page", []entity.Element{{ID: testElemID, Type: "Text", Content: &content, OrderKey: "a1"}}); err != nil {
		t.Fatal(err)
	}
	var stored entity.Element
	if err := db.First(&stored, `"Id" = ?`, testElemID).Error; err != nil {
		t.Fatal(err)
	}
	if stored.OrderKey != "a1" {
		t.Fatalf("persistence lost the moved element's order key: got %q, want a1", stored.OrderKey)
	}
}

func assertCreateAndUpdateViaWS(t *testing.T, client *wsTestClient, id string) {
	t.Helper()
	client.sendRequest("element:create", map[string]any{"element": map[string]any{"id": id, "type": "Text", "content": "new", "parentId": nil}}, "create-after-join")
	created, ok := client.next("element:create", 3*time.Second)
	if !ok || created.RequestID != "create-after-join" || created.Seq != 1 {
		t.Fatalf("create not acknowledged: ok=%v message=%+v", ok, created)
	}
	client.sendRequest("element:update", map[string]any{"id": id, "content": "updated"}, "update-after-join")
	updated, ok := client.next("element:update", 3*time.Second)
	if !ok || updated.RequestID != "update-after-join" || updated.Seq != 2 {
		t.Fatalf("update not acknowledged: ok=%v message=%+v", ok, updated)
	}
}

func TestWSPageWithNullOrderKeysJoinsAndSendsElements(t *testing.T) {
	dsn, legacy := legacyRealtimeDatabase(t)
	if err := legacy.Exec(`CREATE TABLE "Page" ("Id" text PRIMARY KEY, "ProjectId" text, "DeletedAt" timestamptz);
		INSERT INTO "Page" ("Id", "ProjectId") VALUES ('page', 'proj');
		CREATE TABLE "ElementComment" ("Id" text PRIMARY KEY, "ElementId" text, "DeletedAt" timestamptz);
		CREATE TABLE "ElementEventWorkflow" ("Id" text PRIMARY KEY, "ElementId" text)`).Error; err != nil {
		t.Fatal(err)
	}
	db, err := database.NewPostgresConnection(database.Config{URL: dsn})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { closeSQLHandle(t, db) })
	endpoint, docs := startWSServerWithElements(t, repository.NewElementRepository(db))
	if docs == nil {
		t.Fatal("missing docs")
	}
	client, snapshot := dialAndJoin(t, endpoint)
	var page struct {
		Elements []entity.Element `json:"elements"`
		Seq      int64            `json:"seq"`
	}
	if err := json.Unmarshal(snapshot.Payload, &page); err != nil {
		t.Fatal(err)
	}
	if len(page.Elements) != 1 || page.Elements[0].ID != testElemID {
		t.Fatalf("legacy elements did not arrive: %s", snapshot.Payload)
	}
	if page.Elements[0].Content == nil || *page.Elements[0].Content != "saved" {
		t.Fatalf("legacy content unexpected: %s", snapshot.Payload)
	}
	id := uuid.NewString()
	assertCreateAndUpdateViaWS(t, client, id)
	client.sendRequest("sync:page", map[string]any{}, "read-back")
	snapshot, ok := client.next("sync:page", 3*time.Second)
	if !ok {
		t.Fatal("page no longer syncs after edits")
	}
	if err := json.Unmarshal(snapshot.Payload, &page); err != nil {
		t.Fatal(err)
	}
	added := findElementByID(page.Elements, id)
	if added == nil || added.Content == nil {
		t.Fatalf("synced page missing edited element: %s", snapshot.Payload)
	}
	if *added.Content != "updated" || added.OrderKey == "" {
		t.Fatalf("synced page lost edits: %+v", added)
	}
}
