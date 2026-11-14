package unit

import (
	"bytes"
	"strings"
	"testing"

	"my-go-app/internal/realtime/utils/logger"
)

func TestMyLogger_Basic(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(&buf, "", 0) // deterministic: no flags

	l.Info("hello", "world")
	l.Infof("value=%d", 42)
	l.Error("oops")
	l.Errorf("err=%s", "boom")
	l.Debug("dbg")
	l.Debugf("dbg=%s", "ok")

	out := buf.String()

	checks := []string{
		"[INFO] hello world",
		"[INFO] value=42",
		"[ERROR] oops",
		"[ERROR] err=boom",
		"[DEBUG] dbg",
		"[DEBUG] dbg=ok",
	}

	for _, c := range checks {
		if !strings.Contains(out, c) {
			t.Fatalf("expected %q in output, got: %q", c, out)
		}
	}
}

func TestNewWithNilWriterAndDefaults(t *testing.T) {
	// New with nil writer should return a usable logger (writes to stdout in that case).
	l := logger.New(nil, "", 0)
	if l == nil {
		t.Fatal("expected non-nil logger from New(nil, ...)")
	}

	// Default logger should also be non-nil and satisfy the Logger interface.
	if logger.Default() == nil {
		t.Fatal("Default() returned nil")
	}
	if logger.DefaultLogger == nil {
		t.Fatal("DefaultLogger is nil")
	}
}
