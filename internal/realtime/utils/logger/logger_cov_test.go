package logger

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestLoggerLevelsAndFormatting(t *testing.T) {
	tests := []struct {
		name string
		call func(l *MyLogger)
		want string
	}{
		{"Info", func(l *MyLogger) { l.Info("a", 1, "b") }, "[INFO] a 1 b\n"},
		{"Infof", func(l *MyLogger) { l.Infof("n=%d s=%s", 7, "x") }, "[INFO] n=7 s=x\n"},
		{"Error", func(l *MyLogger) { l.Error("boom", 2) }, "[ERROR] boom 2\n"},
		{"Errorf", func(l *MyLogger) { l.Errorf("bad %v", "thing") }, "[ERROR] bad thing\n"},
		{"Debug", func(l *MyLogger) { l.Debug("dbg") }, "[DEBUG] dbg\n"},
		{"Debugf", func(l *MyLogger) { l.Debugf("%03d", 5) }, "[DEBUG] 005\n"},
		{"Info no args", func(l *MyLogger) { l.Info() }, "[INFO] \n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			tt.call(New(&buf, "", 0))
			if buf.String() != tt.want {
				t.Fatalf("got %q want %q", buf.String(), tt.want)
			}
		})
	}
}

func TestLoggerPrefix(t *testing.T) {
	var buf bytes.Buffer
	New(&buf, "app: ", 0).Info("hi")
	if buf.String() != "app: [INFO] hi\n" {
		t.Fatalf("got %q", buf.String())
	}
}

func TestFormatArgsTrimsOnlyTrailingNewlines(t *testing.T) {
	if got := formatArgs("a", "b\n"); got != "a b" {
		t.Fatalf("got %q", got)
	}
	if got := formatArgs("x\ny"); got != "x\ny" {
		t.Fatalf("inner newline must be preserved, got %q", got)
	}
}

func TestNewNilWriterFallsBackToStdout(t *testing.T) {
	l := New(nil, "", 0)
	if l == nil || l.l == nil {
		t.Fatal("nil logger")
	}
	if l.l.Writer() != os.Stdout {
		t.Fatal("nil writer must default to os.Stdout")
	}
}

func TestDefault(t *testing.T) {
	l := Default()
	if l.l.Writer() != os.Stdout {
		t.Fatal("Default must write to stdout")
	}
	if l.l.Flags() != log.LstdFlags|log.Lshortfile {
		t.Fatalf("flags=%d", l.l.Flags())
	}
	if DefaultLogger == nil {
		t.Fatal("DefaultLogger unset")
	}
}

func TestLoggerSatisfiesInterface(t *testing.T) {
	var _ Logger = (*MyLogger)(nil)
}

func TestFatalExitsWithStatusOne(t *testing.T) {
	if mode := os.Getenv("COV_LOGGER_FATAL"); mode != "" {
		l := New(os.Stderr, "", 0)
		if mode == "f" {
			l.Fatalf("dying %d", 1)
		} else {
			l.Fatal("dying", 1)
		}
		return // unreachable when Fatal exits
	}
	for _, mode := range []string{"p", "f"} {
		cmd := exec.Command(os.Args[0], "-test.run=^TestFatalExitsWithStatusOne$")
		cmd.Env = append(os.Environ(), "COV_LOGGER_FATAL="+mode)
		out, err := cmd.CombinedOutput()
		ee, ok := err.(*exec.ExitError)
		if !ok || ee.ExitCode() != 1 {
			t.Fatalf("mode %s: want exit status 1, got err=%v out=%s", mode, err, out)
		}
		if !strings.Contains(string(out), "[FATAL] dying 1") {
			t.Fatalf("mode %s: missing fatal message: %s", mode, out)
		}
	}
}
