package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestCommandRequiresExplicitMigrationURL(t *testing.T) {
	var out, errs bytes.Buffer
	code := run(nil, func(string) string { return "" }, &out, &errs)
	if code == 0 || !strings.Contains(errs.String(), "WORKFLOW_MIGRATION_DATABASE_URL") {
		t.Fatalf("expected explicit URL requirement: %d %s", code, errs.String())
	}
}
func TestCommandRejectsApply(t *testing.T) {
	var out, errs bytes.Buffer
	code := run([]string{"--apply"}, func(string) string { return "" }, &out, &errs)
	if code == 0 {
		t.Fatal("apply must not be accepted by the read-only command")
	}
}

func TestApplyRequiresOfflineLocalDatabase(t *testing.T) {
	for _, tc := range []struct {
		args []string
		dsn  string
	}{
		{[]string{"--apply"}, "postgres://u:p@localhost/db"},
		{[]string{"--apply", "--offline"}, "postgres://u:p@remote.example/db"},
	} {
		var out, errs bytes.Buffer
		if code := run(tc.args, func(string) string { return tc.dsn }, &out, &errs); code == 0 || errs.Len() == 0 {
			t.Fatalf("unsafe invocation accepted: %#v", tc)
		}
	}
}

func TestApplyAcceptsOfflineLocalArguments(t *testing.T) {
	var out, errs bytes.Buffer
	code := run([]string{"--apply", "--offline"}, func(string) string { return "postgres://postgres:postgres@127.0.0.1:1/webbuilder?connect_timeout=1" }, &out, &errs)
	if code != 1 {
		t.Fatalf("expected connection failure after accepting arguments, got %d: %s", code, errs.String())
	}
}

func TestApplyRejectsURLWithRemoteHostOverride(t *testing.T) {
	var out, errs bytes.Buffer
	code := run([]string{"--apply", "--offline"}, func(string) string {
		return "postgres://postgres:postgres@localhost/webbuilder?host=production.example"
	}, &out, &errs)
	if code != 2 {
		t.Fatalf("remote host override not rejected as arguments: %d %s", code, errs.String())
	}
}
