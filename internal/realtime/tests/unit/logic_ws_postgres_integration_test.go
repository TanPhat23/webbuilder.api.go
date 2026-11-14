package unit

import (
	"testing"
	"time"

	"my-go-app/internal/realtime/repository"
	"my-go-app/internal/realtime/usecase"
)

// Exercise the websocket handler and the actual PostgreSQL logic repository together.
func TestLogicWSPostgresBroadcastAndReconnect(t *testing.T) {
	_, db := legacyRealtimeDatabase(t)
	for _, query := range []string{
		`CREATE TABLE "Project" ("Id" text PRIMARY KEY, "OwnerId" text NOT NULL, "DeletedAt" timestamp, "Logic" jsonb, "LogicRevision" bigint NOT NULL DEFAULT 0)`,
		`CREATE TABLE "Collaborator" ("ProjectId" text NOT NULL, "UserId" text NOT NULL, "Role" text NOT NULL)`,
		`CREATE TABLE "Page" ("Id" text PRIMARY KEY, "ProjectId" text NOT NULL, "Logic" jsonb, "LogicRevision" bigint NOT NULL DEFAULT 0)`,
		`INSERT INTO "Project" ("Id", "OwnerId") VALUES ('proj', 'editor')`,
		`INSERT INTO "Collaborator" ("ProjectId", "UserId", "Role") VALUES ('proj', 'viewer', 'viewer')`,
		`INSERT INTO "Page" ("Id", "ProjectId") VALUES ('page', 'proj'), ('page2', 'proj')`,
	} {
		if err := db.Exec(query).Error; err != nil {
			t.Fatal(err)
		}
	}
	url := startLogicServer(t, usecase.NewLogicUseCase(repository.NewLogicRepository(db)))
	editor, ok := dialLogic(t, url, "editor", "page")
	if !ok {
		t.Fatal("editor could not join")
	}
	viewer, ok := dialLogic(t, url, "viewer", "page2")
	if !ok {
		t.Fatal("viewer could not join")
	}
	editor.raw("logic:update", "page", map[string]any{"scope": "project", "patch": map[string]any{"variables": map[string]any{"count": variable("number")}}}, "write")
	for name, client := range map[string]*logicClient{"editor": editor, "viewer": viewer} {
		msg, got := client.next("logic:update", 3*time.Second)
		if !got || decode[logicChanged](t, msg).Logic["variables"].(map[string]any)["count"] == nil {
			t.Fatalf("%s missed update: %s", name, msg.Payload)
		}
	}
	viewer.raw("logic:update", "page2", map[string]any{"scope": "project", "patch": map[string]any{"variables": map[string]any{"blocked": variable("number")}}}, "refused")
	if msg, got := viewer.next("error", 3*time.Second); !got || msg.ClientOpID != "refused" {
		t.Fatalf("viewer write was not rejected: %+v", msg)
	}
	rejoined, ok := dialLogic(t, url, "editor", "page")
	if !ok {
		t.Fatal("reconnecting editor could not join")
	}
	msg, got := rejoined.next("logic:sync", 3*time.Second)
	if !got {
		t.Fatal("missing persisted logic snapshot")
	}
	snapshot := decode[struct {
		Project map[string]any `json:"project"`
	}](t, msg)
	vars := snapshot.Project["variables"].(map[string]any)
	if vars["count"] == nil || vars["blocked"] != nil {
		t.Fatalf("unexpected persisted variables: %v", vars)
	}
}
