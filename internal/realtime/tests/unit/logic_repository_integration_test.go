package unit

import (
	"context"
	"sync"
	"testing"

	"my-go-app/internal/realtime/repository"
)

func TestLogicRepositoryConcurrentNamesAndIsolation(t *testing.T) {
	_, db := legacyRealtimeDatabase(t)
	for _, sql := range []string{
		`CREATE TABLE "Project" ("Id" text PRIMARY KEY, "OwnerId" text NOT NULL, "DeletedAt" timestamp, "Logic" jsonb, "LogicRevision" bigint NOT NULL DEFAULT 0)`,
		`CREATE TABLE "Collaborator" ("ProjectId" text NOT NULL, "UserId" text NOT NULL, "Role" text NOT NULL)`,
		`CREATE TABLE "Page" ("Id" text PRIMARY KEY, "ProjectId" text NOT NULL, "DeletedAt" timestamp, "Logic" jsonb, "LogicRevision" bigint NOT NULL DEFAULT 0)`,
		`INSERT INTO "Project" ("Id", "OwnerId") VALUES ('proj', 'owner')`,
		`INSERT INTO "Collaborator" ("ProjectId", "UserId", "Role") VALUES ('proj', 'viewer', 'viewer'), ('proj', 'editor', 'editor')`,
		`INSERT INTO "Page" ("Id", "ProjectId") VALUES ('p1', 'proj'), ('p2', 'proj')`,
	} {
		if err := db.Exec(sql).Error; err != nil {
			t.Fatal(err)
		}
	}
	repo := repository.NewLogicRepository(db)
	for _, tc := range []struct {
		user        string
		read, write bool
	}{{"owner", true, true}, {"editor", true, true}, {"viewer", true, false}, {"stranger", false, false}} {
		read, write, err := repo.Access(context.Background(), "proj", tc.user)
		if err != nil || read != tc.read || write != tc.write {
			t.Fatalf("%s: access %v/%v: %v", tc.user, read, write, err)
		}
	}
	patch := map[string]any{"variables": map[string]any{"count": map[string]any{"type": "number", "default": float64(1)}}}
	var wg sync.WaitGroup
	errors := make(chan error, 2)
	for _, id := range []string{"p1", "p2"} {
		wg.Add(1)
		go func(page string) {
			defer wg.Done()
			_, err := repo.Apply(context.Background(), "proj", page, "editor", patch)
			errors <- err
		}(id)
	}
	wg.Wait()
	close(errors)
	successes := 0
	for err := range errors {
		if err == nil {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("expected one successful conflicting update, got %d", successes)
	}
	project, page, err := repo.Get(context.Background(), "proj", "p1", "viewer")
	if err != nil {
		t.Fatal(err)
	}
	if project.Doc["v"] != float64(1) || page.Doc["v"] != float64(1) {
		t.Fatal("snapshot did not include both documents")
	}
	if _, _, err := repo.Get(context.Background(), "other", "p1", "stranger"); err == nil {
		t.Fatal("cross-project read accepted")
	}
	if _, err := repo.Apply(context.Background(), "proj", "p1", "viewer", patch); err == nil {
		t.Fatal("viewer can write")
	}
	if _, _, err := repo.Get(context.Background(), "proj", "p1", "stranger"); err == nil {
		t.Fatal("stranger can read")
	}
	if _, err := repo.Apply(context.Background(), "other", "p1", "stranger", patch); err == nil {
		t.Fatal("cross-project update accepted")
	}
}

// Each committed write bumps that document's revision; reads report it. The
// revision is what lets a client drop a sync older than a broadcast it has.
func TestLogicRepositoryRevisionsIncreasePerDocument(t *testing.T) {
	_, db := legacyRealtimeDatabase(t)
	for _, sql := range []string{
		`CREATE TABLE "Project" ("Id" text PRIMARY KEY, "OwnerId" text NOT NULL, "DeletedAt" timestamp, "Logic" jsonb, "LogicRevision" bigint NOT NULL DEFAULT 0)`,
		`CREATE TABLE "Collaborator" ("ProjectId" text NOT NULL, "UserId" text NOT NULL, "Role" text NOT NULL)`,
		`CREATE TABLE "Page" ("Id" text PRIMARY KEY, "ProjectId" text NOT NULL, "DeletedAt" timestamp, "Logic" jsonb, "LogicRevision" bigint NOT NULL DEFAULT 0)`,
		`INSERT INTO "Project" ("Id", "OwnerId") VALUES ('proj', 'owner')`,
		`INSERT INTO "Page" ("Id", "ProjectId") VALUES ('p1', 'proj')`,
	} {
		if err := db.Exec(sql).Error; err != nil {
			t.Fatal(err)
		}
	}
	repo := repository.NewLogicRepository(db)
	ctx := context.Background()
	project, page, err := repo.Get(ctx, "proj", "p1", "owner")
	if err != nil || project.Revision != 0 || page.Revision != 0 {
		t.Fatalf("fresh revisions: %d %d %v", project.Revision, page.Revision, err)
	}
	var last int64
	for _, name := range []string{"a", "b", "c"} {
		got, err := repo.Apply(ctx, "proj", "", "owner", map[string]any{"variables": map[string]any{name: map[string]any{"type": "number", "default": float64(0)}}})
		if err != nil {
			t.Fatal(err)
		}
		if got.Revision <= last {
			t.Fatalf("project revision did not increase: %d after %d", got.Revision, last)
		}
		last = got.Revision
	}
	// An incomplete draft step is stored like the editor creates it.
	got, err := repo.Apply(ctx, "proj", "p1", "owner", map[string]any{"flows": map[string]any{"f": map[string]any{"do": []any{map[string]any{"show": ""}}}}})
	if err != nil || got.Revision != 1 {
		t.Fatalf("page revision: %+v %v", got, err)
	}
	if _, err := repo.Apply(ctx, "proj", "", "owner", map[string]any{"variables": map[string]any{"bad": map[string]any{"type": "number", "default": "x"}}}); err == nil {
		t.Fatal("invalid patch accepted")
	}
	project, page, err = repo.Get(ctx, "proj", "p1", "owner")
	if err != nil || project.Revision != last || page.Revision != 1 {
		t.Fatalf("a rejected write must not bump the revision: project=%d (want %d) page=%d err=%v", project.Revision, last, page.Revision, err)
	}
}

func TestLogicRepositoryIgnoresDeletedPages(t *testing.T) {
	_, db := legacyRealtimeDatabase(t)
	for _, query := range []string{
		`CREATE TABLE "Project" ("Id" text PRIMARY KEY, "OwnerId" text NOT NULL, "DeletedAt" timestamp, "Logic" jsonb, "LogicRevision" bigint NOT NULL DEFAULT 0)`,
		`CREATE TABLE "Collaborator" ("ProjectId" text NOT NULL, "UserId" text NOT NULL, "Role" text NOT NULL)`,
		`CREATE TABLE "Page" ("Id" text PRIMARY KEY, "ProjectId" text NOT NULL, "DeletedAt" timestamp, "Logic" jsonb, "LogicRevision" bigint NOT NULL DEFAULT 0)`,
		`INSERT INTO "Project" ("Id", "OwnerId") VALUES ('proj', 'owner')`,
		`INSERT INTO "Page" ("Id", "ProjectId", "DeletedAt", "Logic") VALUES ('deleted', 'proj', now(), '{"v":1,"variables":{"used":{"type":"number","default":0}},"flows":{}}')`,
	} {
		if err := db.Exec(query).Error; err != nil {
			t.Fatal(err)
		}
	}
	repo := repository.NewLogicRepository(db)
	patch := map[string]any{"variables": map[string]any{"used": map[string]any{"type": "number", "default": float64(0)}}}
	if _, err := repo.Apply(context.Background(), "proj", "", "owner", patch); err != nil {
		t.Fatalf("deleted page blocked name: %v", err)
	}
	if _, err := repo.Apply(context.Background(), "proj", "deleted", "owner", patch); err == nil {
		t.Fatal("updated deleted page")
	}
	if _, _, err := repo.Get(context.Background(), "proj", "deleted", "owner"); err == nil {
		t.Fatal("read deleted page")
	}
}

func TestLogicRepositoryMoveIsAtomicAndAuthorized(t *testing.T) {
	_, db := legacyRealtimeDatabase(t)
	for _, query := range []string{
		`CREATE TABLE "Project" ("Id" text PRIMARY KEY, "OwnerId" text NOT NULL, "DeletedAt" timestamp, "Logic" jsonb, "LogicRevision" bigint NOT NULL DEFAULT 0)`,
		`CREATE TABLE "Collaborator" ("ProjectId" text NOT NULL, "UserId" text NOT NULL, "Role" text NOT NULL)`,
		`CREATE TABLE "Page" ("Id" text PRIMARY KEY, "ProjectId" text NOT NULL, "DeletedAt" timestamp, "Logic" jsonb, "LogicRevision" bigint NOT NULL DEFAULT 0)`,
		`INSERT INTO "Project" ("Id", "OwnerId", "Logic") VALUES ('proj', 'owner', '{"v":1,"variables":{"count":{"type":"number","default":2}},"flows":{}}')`,
		`INSERT INTO "Page" ("Id", "ProjectId") VALUES ('p1', 'proj')`,
		`INSERT INTO "Collaborator" VALUES ('proj', 'viewer', 'viewer')`,
	} {
		if err := db.Exec(query).Error; err != nil {
			t.Fatal(err)
		}
	}
	repo := repository.NewLogicRepository(db)
	ctx := context.Background()
	if _, _, err := repo.Move(ctx, "proj", "p1", "viewer", "variable", "count", "page"); err == nil {
		t.Fatal("viewer moved logic")
	}
	project, page, err := repo.Move(ctx, "proj", "p1", "owner", "variable", "count", "page")
	if err != nil {
		t.Fatal(err)
	}
	if project.Revision != 1 || page.Revision != 1 {
		t.Fatalf("revisions: %d %d", project.Revision, page.Revision)
	}
	if _, ok := project.Doc["variables"].(map[string]any)["count"]; ok {
		t.Fatal("source still contains variable")
	}
	if _, ok := page.Doc["variables"].(map[string]any)["count"]; !ok {
		t.Fatal("destination lost variable")
	}
	if _, _, err := repo.Move(ctx, "proj", "p1", "owner", "variable", "count", "page"); err == nil {
		t.Fatal("missing source accepted")
	}
	gotProject, gotPage, err := repo.Get(ctx, "proj", "p1", "owner")
	if err != nil || gotProject.Revision != 1 || gotPage.Revision != 1 {
		t.Fatalf("failed move changed revisions: %+v %+v %v", gotProject, gotPage, err)
	}
	if _, _, err := repo.Move(ctx, "proj", "p1", "owner", "variable", "count", "project"); err != nil {
		t.Fatalf("move back: %v", err)
	}
	// A failure after the project update must roll back both documents.
	if err := db.Exec(`CREATE FUNCTION fail_page_update() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'injected failure'; END $$`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`CREATE TRIGGER fail_page_update BEFORE UPDATE ON "Page" FOR EACH ROW EXECUTE FUNCTION fail_page_update()`).Error; err != nil {
		t.Fatal(err)
	}
	if _, _, err := repo.Move(ctx, "proj", "p1", "owner", "variable", "count", "page"); err == nil {
		t.Fatal("move succeeded despite page write failure")
	}
	gotProject, gotPage, err = repo.Get(ctx, "proj", "p1", "owner")
	if err != nil || gotProject.Revision != 2 || gotPage.Revision != 2 {
		t.Fatalf("failed transaction changed revisions: %+v %+v %v", gotProject, gotPage, err)
	}
	if _, ok := gotProject.Doc["variables"].(map[string]any)["count"]; !ok {
		t.Fatal("failed transaction lost source definition")
	}
}
