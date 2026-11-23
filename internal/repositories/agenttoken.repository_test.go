package repositories

import (
	"context"
	"database/sql/driver"
	"errors"
	"testing"
	"time"

	"my-go-app/internal/models"
)

var errAgentTokenBoom = errors.New("agent token boom")

func TestAgentTokenCreateInsertsAndPropagatesErrors(t *testing.T) {
	db, f := newFakeDB(t)
	f.onExec("INSERT INTO", 1)
	repo := NewAgentTokenRepository(db)
	tok := &models.AgentToken{Id: "t1", UserId: "u1", Name: "cli", Prefix: "wb_", TokenHash: "h"}
	if err := repo.Create(context.Background(), tok); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if !f.sawQuery("INSERT INTO", "AgentToken") {
		t.Errorf("no insert into AgentToken: %q", f.lastQuery())
	}
	if tok.CreatedAt.IsZero() {
		t.Error("CreatedAt should be populated on create")
	}

	db2, f2 := newFakeDB(t)
	f2.onErr("INSERT INTO", errAgentTokenBoom)
	if err := NewAgentTokenRepository(db2).Create(context.Background(), tok); !errors.Is(err, errAgentTokenBoom) {
		t.Errorf("Create error = %v, want %v", err, errAgentTokenBoom)
	}
}

func TestAgentTokenListActiveByUser(t *testing.T) {
	db, f := newFakeDB(t)
	now := time.Now()
	f.on("FROM \"public\".\"AgentToken\"", []string{"Id", "UserId", "Name", "CreatedAt"},
		[]driver.Value{"t2", "u1", "newer", now},
		[]driver.Value{"t1", "u1", "older", now.Add(-time.Hour)})
	got, err := NewAgentTokenRepository(db).ListActiveByUser(context.Background(), "u1")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Id != "t2" || got[1].Name != "older" {
		t.Fatalf("unexpected tokens: %+v", got)
	}
	if !f.sawQuery("\"RevokedAt\" IS NULL", "ORDER BY \"CreatedAt\" DESC") {
		t.Errorf("query should filter revoked tokens and sort newest first: %q", f.lastQuery())
	}
	if args := f.args[0]; len(args) == 0 || args[0].Value != "u1" {
		t.Errorf("user id not bound: %+v", args)
	}

	db2, f2 := newFakeDB(t)
	f2.onErr("SELECT", errAgentTokenBoom)
	if _, err := NewAgentTokenRepository(db2).ListActiveByUser(context.Background(), "u1"); !errors.Is(err, errAgentTokenBoom) {
		t.Errorf("error = %v", err)
	}
}

func TestAgentTokenRevoke(t *testing.T) {
	db, f := newFakeDB(t)
	f.onExec("UPDATE", 1).once()
	repo := NewAgentTokenRepository(db)
	ok, err := repo.Revoke(context.Background(), "u1", "t1")
	if err != nil || !ok {
		t.Fatalf("Revoke = %v, %v; want true, nil", ok, err)
	}
	if !f.sawQuery("UPDATE", "\"RevokedAt\"", "\"RevokedAt\" IS NULL") {
		t.Errorf("unexpected update: %q", f.lastQuery())
	}

	// Nothing matched (unknown, other user's or already revoked token).
	f.onExec("UPDATE", 0)
	ok, err = repo.Revoke(context.Background(), "u1", "missing")
	if err != nil || ok {
		t.Errorf("Revoke missing = %v, %v; want false, nil", ok, err)
	}

	db2, f2 := newFakeDB(t)
	f2.onErr("UPDATE", errAgentTokenBoom)
	ok, err = NewAgentTokenRepository(db2).Revoke(context.Background(), "u1", "t1")
	if ok || !errors.Is(err, errAgentTokenBoom) {
		t.Errorf("Revoke error = %v, %v", ok, err)
	}
}
