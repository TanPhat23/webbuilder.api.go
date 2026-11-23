package database

import (
	"net"
	"strings"
	"testing"
)

// closedPort returns a loopback port nothing is listening on, so connecting fails fast.
func closedPort(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	_, port, _ := net.SplitHostPort(l.Addr().String())
	if err := l.Close(); err != nil {
		t.Fatal(err)
	}
	return port
}

func TestNewPostgresConnectionBuildsDSNFromFields(t *testing.T) {
	port := closedPort(t)
	db, err := NewPostgresConnection(Config{
		Host: "127.0.0.1", Port: port, User: "covuser", Password: "covpw", DBName: "covdb", SSLMode: "disable",
	})
	if err == nil {
		t.Fatalf("expected connection error, got db=%v", db)
	}
	if db != nil {
		t.Fatal("db must be nil on error")
	}
	// The driver reports the effective connection target, proving every field reached the DSN.
	msg := err.Error()
	for _, want := range []string{"127.0.0.1", port, "covuser", "covdb"} {
		if !strings.Contains(msg, want) {
			t.Errorf("error %q does not mention %q", msg, want)
		}
	}
	if strings.Contains(msg, "covpw") {
		t.Errorf("error leaks the password: %q", msg)
	}
}

func TestNewPostgresConnectionURLTakesPriority(t *testing.T) {
	port := closedPort(t)
	db, err := NewPostgresConnection(Config{
		URL:  "postgres://urluser:urlpw@127.0.0.1:" + port + "/urldb?sslmode=disable&connect_timeout=1",
		Host: "should-not-be-used.invalid", Port: "1", User: "fielduser", DBName: "fielddb",
	})
	if err == nil {
		t.Fatalf("expected connection error, got db=%v", db)
	}
	msg := err.Error()
	if !strings.Contains(msg, "urluser") || !strings.Contains(msg, "urldb") {
		t.Errorf("URL not used: %q", msg)
	}
	if strings.Contains(msg, "fielduser") || strings.Contains(msg, "should-not-be-used") {
		t.Errorf("individual fields leaked into the connection when URL was set: %q", msg)
	}
}

func TestNewPostgresConnectionInvalidURL(t *testing.T) {
	db, err := NewPostgresConnection(Config{URL: "postgres://user@host:notaport/db"})
	if err == nil || db != nil {
		t.Fatalf("expected parse error and nil db, got db=%v err=%v", db, err)
	}
}
