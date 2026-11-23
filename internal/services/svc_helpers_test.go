package services

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

// errBoom is the sentinel returned by fake repositories to prove that the
// service wraps or forwards repository failures instead of swallowing them.
var errBoom = errors.New("boom")

func strp(s string) *string { return &s }

func rawp(s string) *json.RawMessage {
	r := json.RawMessage(s)
	return &r
}

func boolp(b bool) *bool { return &b }

// wantErr fails unless err is non-nil and contains substr.
func wantErr(t *testing.T, err error, substr string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error containing %q, got nil", substr)
	}
	if !strings.Contains(err.Error(), substr) {
		t.Fatalf("expected error containing %q, got %q", substr, err.Error())
	}
}

func wantNoErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// wantCalls fails unless the recorded calls equal want, in order.
func wantCalls(t *testing.T, got []string, want ...string) {
	t.Helper()
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("calls = %v, want %v", got, want)
	}
}
