package models

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestAgentTokenTableName(t *testing.T) {
	if got := (AgentToken{}).TableName(); got != `public."AgentToken"` {
		t.Fatalf("table name = %s", got)
	}
}

func TestAgentTokenJSONNeverExposesHash(t *testing.T) {
	used := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	b, err := json.Marshal(AgentToken{Id: "t1", UserId: "u1", Name: "ci", Prefix: "wb_abc", TokenHash: "deadbeef", LastUsedAt: &used})
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if strings.Contains(s, "deadbeef") || strings.Contains(strings.ToLower(s), "tokenhash") {
		t.Fatalf("hash leaked: %s", s)
	}
	if strings.Contains(s, "revokedAt") {
		t.Fatalf("nil RevokedAt must be omitted: %s", s)
	}
	var got map[string]any
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if got["id"] != "t1" || got["userId"] != "u1" || got["name"] != "ci" || got["prefix"] != "wb_abc" || got["lastUsedAt"] != "2026-01-02T03:04:05Z" {
		t.Fatalf("unexpected json: %s", s)
	}
}
