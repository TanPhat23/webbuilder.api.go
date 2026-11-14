package workflowmigration

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func migrationTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("WORKFLOW_MIGRATION_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("WORKFLOW_MIGRATION_TEST_DATABASE_URL not set")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	schema := "migration_test_" + strings.ReplaceAll(time.Now().Format("150405.000000000"), ".", "")
	if _, err = db.Exec(`CREATE SCHEMA ` + schema); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`SET search_path TO ` + schema); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Exec(`DROP SCHEMA ` + schema + ` CASCADE`); db.Close() })
	_, err = db.Exec(`CREATE TABLE "Project" ("Id" text PRIMARY KEY, "Logic" jsonb);
 CREATE TABLE "Page" ("Id" text PRIMARY KEY, "ProjectId" text, "Logic" jsonb);
 CREATE TABLE "Element" ("Id" text PRIMARY KEY, "PageId" text, "Settings" jsonb);
 CREATE TABLE "EventWorkflow" ("Id" text PRIMARY KEY, "ProjectId" text, "Enabled" boolean, "Handlers" jsonb);
 CREATE TABLE "ElementEventWorkflow" ("Id" text PRIMARY KEY, "ElementId" text, "WorkflowId" text, "EventName" text);`)
	if err != nil {
		t.Fatal(err)
	}
	return db
}
func TestDryRunReportsConflictsWithoutWriting(t *testing.T) {
	db := migrationTestDB(t)
	_, err := db.Exec(`INSERT INTO "Project" VALUES ('p', NULL); INSERT INTO "Page" VALUES ('page','p',NULL);
 INSERT INTO "Element" VALUES ('e','page','{"interactions":{"ix_ca978112":{"on":"click","do":[]}}}'),('e2','page',NULL);
 INSERT INTO "EventWorkflow" VALUES ('w','p',true,'[{"config":{"type":"resetForm","formElementId":"f"}}]'),('unsupported','p',true,'[{"config":{"type":"unknown"}}]');
 INSERT INTO "ElementEventWorkflow" VALUES ('a','e','w','onClick'),('b','e2','unsupported','onClick');`)
	if err != nil {
		t.Fatal(err)
	}
	report, err := DryRun(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Issues) != 2 {
		t.Fatalf("expected both conflict and unsupported reports: %#v", report)
	}
	if len(report.Interactions) != 0 || len(report.Flows) != 0 {
		t.Fatalf("conflicting plans must not be offered: %#v", report)
	}
	var settings []byte
	if err := db.QueryRow(`SELECT "Settings" FROM "Element" WHERE "Id"='e'`).Scan(&settings); err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(settings, &got); err != nil {
		t.Fatal(err)
	}
	interactions := got["interactions"].(map[string]any)
	if len(interactions) != 1 {
		t.Fatalf("dry run wrote data: %s", settings)
	}
}
func TestDryRunDetectsFlowCollisionAcrossProjectPages(t *testing.T) {
	db := migrationTestDB(t)
	flow := stableName("flow_", "w", 16)
	_, err := db.Exec(`INSERT INTO "Project" VALUES ('p',NULL); INSERT INTO "Page" VALUES ('page','p',NULL), ('other','p',NULL);
 INSERT INTO "Element" VALUES ('e1','page',NULL),('e2','page',NULL);
 INSERT INTO "EventWorkflow" VALUES ('w','p',true,'[{"config":{"type":"resetForm","formElementId":"f"}}]');
 INSERT INTO "ElementEventWorkflow" VALUES ('a','e1','w','onClick'),('b','e2','w','onClick');`)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE "Page" SET "Logic"=jsonb_build_object('v',1,'variables','{}'::jsonb,'flows',jsonb_build_object($1::text,jsonb_build_object('do','[]'::jsonb))) WHERE "Id"='other'`, flow); err != nil {
		t.Fatal(err)
	}
	report, err := DryRun(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Issues) != 1 || len(report.Interactions) != 0 || len(report.Flows) != 0 {
		t.Fatalf("flow collision missed: %#v", report)
	}
}

func TestDryRunReportsCollidingProposedInteractionIDs(t *testing.T) {
	db := migrationTestDB(t)
	_, err := db.Exec(`INSERT INTO "Project" VALUES ('p',NULL); INSERT INTO "Page" VALUES ('page','p',NULL);
 INSERT INTO "Element" VALUES ('e','page',NULL);
 INSERT INTO "EventWorkflow" VALUES ('w1','p',true,'[{"config":{"type":"resetForm","formElementId":"f"}}]'),('w2','p',true,'[{"config":{"type":"resetForm","formElementId":"f"}}]');
 INSERT INTO "ElementEventWorkflow" VALUES ('69235','e','w1','onClick'),('95303','e','w2','onClick');`)
	if err != nil {
		t.Fatal(err)
	}
	report, err := DryRun(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Issues) != 2 || len(report.Interactions) != 0 {
		t.Fatalf("planned ID collision missed: %#v", report)
	}
}

func TestDryRunRejectsMalformedExistingLogic(t *testing.T) {
	db := migrationTestDB(t)
	_, err := db.Exec(`INSERT INTO "Project" VALUES ('p','{"v":1,"flows":{}}'); INSERT INTO "Page" VALUES ('page','p',NULL);
 INSERT INTO "Element" VALUES ('e1','page',NULL),('e2','page',NULL);
 INSERT INTO "EventWorkflow" VALUES ('w','p',true,'[{"config":{"type":"resetForm","formElementId":"f"}}]');
 INSERT INTO "ElementEventWorkflow" VALUES ('a','e1','w','onClick'),('b','e2','w','onClick');`)
	if err != nil {
		t.Fatal(err)
	}
	report, err := DryRun(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Issues) != 1 || len(report.Flows) != 0 {
		t.Fatalf("malformed existing logic accepted: %#v", report)
	}
}
