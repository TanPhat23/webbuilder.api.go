package redis

import (
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
)

func TestNewRedisClientHostPort(t *testing.T) {
	mr := miniredis.RunT(t)
	host, port := mr.Host(), mr.Port()

	c, err := NewRedisClient(Config{Host: host, Port: port})
	if err != nil {
		t.Fatalf("NewRedisClient: %v", err)
	}
	defer c.Close()
	if got := c.Options().Addr; got != host+":"+port {
		t.Fatalf("addr=%q", got)
	}
	if err := c.Set(t.Context(), "k", "v", 0).Err(); err != nil {
		t.Fatal(err)
	}
	if got, _ := mr.Get("k"); got != "v" {
		t.Fatalf("server did not receive write, got %q", got)
	}
}

func TestNewRedisClientURLTakesPriority(t *testing.T) {
	mr := miniredis.RunT(t)
	// Host/Port point at a dead address; a working URL must win.
	c, err := NewRedisClient(Config{URL: "redis://" + mr.Addr() + "/2", Host: "127.0.0.1", Port: "1"})
	if err != nil {
		t.Fatalf("NewRedisClient: %v", err)
	}
	defer c.Close()
	if c.Options().DB != 2 {
		t.Fatalf("db from URL not applied: %d", c.Options().DB)
	}
}

func TestNewRedisClientPasswordAndDB(t *testing.T) {
	mr := miniredis.RunT(t)
	mr.RequireAuth("secret")

	if _, err := NewRedisClient(Config{Host: mr.Host(), Port: mr.Port(), Password: "wrong"}); err == nil {
		t.Fatal("wrong password accepted")
	}
	c, err := NewRedisClient(Config{Host: mr.Host(), Port: mr.Port(), Password: "secret", DB: 4})
	if err != nil {
		t.Fatalf("NewRedisClient: %v", err)
	}
	defer c.Close()
	if err := c.Set(t.Context(), "k", "v", 0).Err(); err != nil {
		t.Fatal(err)
	}
	if !mr.DB(4).Exists("k") {
		t.Fatal("key not written to selected DB 4")
	}
}

func TestNewRedisClientErrors(t *testing.T) {
	mr := miniredis.RunT(t)
	port := mr.Port()
	mr.Close() // nothing listens on it any more

	tests := []struct {
		name string
		cfg  Config
		want string
	}{
		{"bad url", Config{URL: "http://not-redis"}, "failed to parse redis url"},
		{"unreachable", Config{Host: "127.0.0.1", Port: port}, "failed to connect to redis"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewRedisClient(tt.cfg)
			if err == nil {
				c.Close()
				t.Fatal("expected error")
			}
			if c != nil {
				t.Fatal("client must be nil on error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("err=%q want substring %q", err, tt.want)
			}
		})
	}
}
