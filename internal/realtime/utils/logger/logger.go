package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// Logger provides a minimal logging interface for injection and testing.
type Logger interface {
	Info(args ...any)
	Infof(format string, args ...any)
	Error(args ...any)
	Errorf(format string, args ...any)
	Debug(args ...any)
	Debugf(format string, args ...any)
	Fatal(args ...any)
	Fatalf(format string, args ...any)
}

type MyLogger struct {
	l *log.Logger
}

func New(out io.Writer, prefix string, flag int) *MyLogger {
	if out == nil {
		out = os.Stdout
	}
	return &MyLogger{l: log.New(out, prefix, flag)}
}

func Default() *MyLogger {
	return New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
}

var DefaultLogger Logger = Default()

const callerDepth = 4

func (l *MyLogger) output(msg string) {
	_ = l.l.Output(callerDepth, msg)
}

func formatArgs(args ...any) string {
	return strings.TrimRight(fmt.Sprintln(args...), "\n")
}

func (l *MyLogger) Info(args ...any) {
	l.output("[INFO] " + formatArgs(args...))
}

func (l *MyLogger) Infof(format string, args ...any) {
	l.output("[INFO] " + fmt.Sprintf(format, args...))
}

func (l *MyLogger) Error(args ...any) {
	l.output("[ERROR] " + formatArgs(args...))
}

func (l *MyLogger) Errorf(format string, args ...any) {
	l.output("[ERROR] " + fmt.Sprintf(format, args...))
}

func (l *MyLogger) Debug(args ...any) {
	l.output("[DEBUG] " + formatArgs(args...))
}

func (l *MyLogger) Debugf(format string, args ...any) {
	l.output("[DEBUG] " + fmt.Sprintf(format, args...))
}

func (l *MyLogger) Fatal(args ...any) {
	l.output("[FATAL] " + formatArgs(args...))
	os.Exit(1)
}

func (l *MyLogger) Fatalf(format string, args ...any) {
	l.output("[FATAL] " + fmt.Sprintf(format, args...))
	os.Exit(1)
}
