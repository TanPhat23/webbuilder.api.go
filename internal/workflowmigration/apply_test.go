package workflowmigration

import (
	"context"
	"encoding/json"
	"testing"
)

func TestApplyIsAtomicAndIdempotent(t *testing.T) {
	db := migrationTestDB(t)
	_, err := db.Exec(`INSERT INTO "Project" VALUES ('p',NULL); INSERT INTO "Page" VALUES ('page','p',NULL);
 INSERT INTO "Element" VALUES ('e1','page','{"keep":true}'),('e2','page',NULL);
 INSERT INTO "EventWorkflow" VALUES ('w','p',true,'[{"config":{"type":"resetForm","formElementId":"f"}}]');
 INSERT INTO "ElementEventWorkflow" VALUES ('a','e1','w','onClick'),('b','e2','w','onClick');`)
	if err != nil {
		t.Fatal(err)
	}
	report, err := Apply(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Issues) != 0 || len(report.Interactions) != 2 || len(report.Flows) != 1 {
		t.Fatalf("unexpected report: %#v", report)
	}
	var raw []byte
	if err := db.QueryRow(`SELECT "Settings" FROM "Element" WHERE "Id"='e1'`).Scan(&raw); err != nil {
		t.Fatal(err)
	}
	var settings map[string]any
	if err := json.Unmarshal(raw, &settings); err != nil {
		t.Fatal(err)
	}
	if settings["keep"] != true || settings["migratedWorkflowLinks"].(map[string]any)["a"] != true || len(settings["interactions"].(map[string]any)) != 1 {
		t.Fatalf("wrong settings: %s", raw)
	}
	report, err = Apply(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Issues) != 0 || len(report.Interactions) != 0 || len(report.Flows) != 0 {
		t.Fatalf("not idempotent: %#v", report)
	}
}

func TestApplyDoesNotWriteWhenAnyWorkflowIsUnsupported(t *testing.T) {
	db := migrationTestDB(t)
	_, err := db.Exec(`INSERT INTO "Project" VALUES ('p',NULL); INSERT INTO "Page" VALUES ('page','p',NULL);
 INSERT INTO "Element" VALUES ('e1','page',NULL),('e2','page',NULL);
 INSERT INTO "EventWorkflow" VALUES ('w1','p',true,'[{"config":{"type":"resetForm","formElementId":"f"}}]'),('w2','p',true,'[{"config":{"type":"unknown"}}]');
 INSERT INTO "ElementEventWorkflow" VALUES ('a','e1','w1','onClick'),('b','e2','w2','onClick');`)
	if err != nil {
		t.Fatal(err)
	}
	report, err := Apply(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Issues) == 0 {
		t.Fatal("missing unsupported workflow")
	}
	var unchanged bool
	if err := db.QueryRow(`SELECT "Settings" IS NULL FROM "Element" WHERE "Id"='e1'`).Scan(&unchanged); err != nil || !unchanged {
		t.Fatalf("partial migration: %v %v", unchanged, err)
	}
}

func TestApplyRollsBackElementWritesWhenFlowWriteFails(t *testing.T) {
	db := migrationTestDB(t)
	_, err := db.Exec(`INSERT INTO "Project" VALUES ('p',NULL); INSERT INTO "Page" VALUES ('page','p',NULL);
 INSERT INTO "Element" VALUES ('e1','page',NULL),('e2','page',NULL);
 INSERT INTO "EventWorkflow" VALUES ('w','p',true,'[{"config":{"type":"resetForm","formElementId":"f"}}]');
 INSERT INTO "ElementEventWorkflow" VALUES ('a','e1','w','onClick'),('b','e2','w','onClick');
 ALTER TABLE "Page" ADD CONSTRAINT reject_logic CHECK ("Logic" IS NULL);`)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Apply(context.Background(), db); err == nil {
		t.Fatal("expected flow write failure")
	}
	var untouched bool
	if err := db.QueryRow(`SELECT bool_and("Settings" IS NULL) FROM "Element"`).Scan(&untouched); err != nil || !untouched {
		t.Fatalf("element writes were not rolled back: %v %v", untouched, err)
	}
}
