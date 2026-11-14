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
		`CREATE TABLE "Page" ("Id" text PRIMARY KEY, "ProjectId" text NOT NULL, "Logic" jsonb, "LogicRevision" bigint NOT NULL DEFAULT 0)`,
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
		`CREATE TABLE "Page" ("Id" text PRIMARY KEY, "ProjectId" text NOT NULL, "Logic" jsonb, "LogicRevision" bigint NOT NULL DEFAULT 0)`,
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
