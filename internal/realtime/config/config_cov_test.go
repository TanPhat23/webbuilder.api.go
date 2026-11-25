package config

import (
	"reflect"
	"testing"
)

func clearConfigEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{
		"SERVER_PORT", "DATABASE_URL", "DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSLMODE",
		"REDIS_URL", "REDIS_HOST", "REDIS_PORT", "REDIS_PASSWORD", "REDIS_DB", "DEV_MODE", "CLERK_SECRET_KEY",
		"ALLOWED_ORIGINS", "ENV",
	} {
		// t.Setenv registers restoration of the original value; Unsetenv then makes the key absent.
		t.Setenv(k, "")
		unsetEnv(t, k)
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	clearConfigEnv(t)
	t.Chdir(t.TempDir()) // no .env file to pick up

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	want := Config{
		ServerPort: "8080",
		Database: DatabaseConfig{
			Host: "localhost", Port: "5432", User: "postgres", Password: "postgres", Name: "webbuilder", SSLMode: "disable",
		},
		Redis:          RedisConfig{Host: "localhost", Port: "6379"},
		DevMode:        false,
		AllowedOrigins: []string{"http://localhost:3000", "https://webbuilderv2.vercel.app"},
	}
	if !reflect.DeepEqual(*cfg, want) {
		t.Fatalf("defaults mismatch:\n got %+v\nwant %+v", *cfg, want)
	}
}

func TestLoadConfigProductionEmptyAllowedOriginsByDefault(t *testing.T) {
	clearConfigEnv(t)
	t.Chdir(t.TempDir())
	t.Setenv("ENV", "production")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(cfg.AllowedOrigins) != 0 {
		t.Fatalf("expected empty AllowedOrigins in production without ALLOWED_ORIGINS, got %v", cfg.AllowedOrigins)
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	clearConfigEnv(t)
	t.Chdir(t.TempDir())
	env := map[string]string{
		"SERVER_PORT": "9000", "DATABASE_URL": "postgres://u:p@h/db", "DB_HOST": "dbh", "DB_PORT": "6543",
		"DB_USER": "alice", "DB_PASSWORD": "pw", "DB_NAME": "wb", "DB_SSLMODE": "require",
		"REDIS_URL": "redis://r:1/2", "REDIS_HOST": "rh", "REDIS_PORT": "7000", "REDIS_PASSWORD": "rpw", "REDIS_DB": "3",
		"DEV_MODE": "true", "CLERK_SECRET_KEY": "sk_test",
		"ALLOWED_ORIGINS": "https://app.example.com, https://admin.example.com",
	}
	for k, v := range env {
		t.Setenv(k, v)
	}
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	want := Config{
		ServerPort: "9000",
		Database: DatabaseConfig{
			URL: "postgres://u:p@h/db", Host: "dbh", Port: "6543", User: "alice", Password: "pw", Name: "wb", SSLMode: "require",
		},
		Redis:          RedisConfig{URL: "redis://r:1/2", Host: "rh", Port: "7000", Password: "rpw", DB: 3},
		DevMode:        true,
		ClerkSecretKey: "sk_test",
		AllowedOrigins: []string{"https://app.example.com", "https://admin.example.com"},
	}
	if !reflect.DeepEqual(*cfg, want) {
		t.Fatalf("env mismatch:\n got %+v\nwant %+v", *cfg, want)
	}
}

func TestGetEnv(t *testing.T) {
	t.Setenv("COV_SET", "value")
	t.Setenv("COV_EMPTY", "")
	unsetEnv(t, "COV_UNSET")
	tests := []struct{ key, fallback, want string }{
		{"COV_SET", "fb", "value"},
		{"COV_EMPTY", "fb", ""}, // set-but-empty wins over the fallback
		{"COV_UNSET", "fb", "fb"},
	}
	for _, tt := range tests {
		if got := envOrDefault(tt.key, tt.fallback); got != tt.want {
			t.Errorf("envOrDefault(%q)=%q want %q", tt.key, got, tt.want)
		}
	}
}

func TestGetEnvAsInt(t *testing.T) {
	unsetEnv(t, "COV_INT_UNSET")
	tests := []struct {
		name, key, val string
		set            bool
		want           int
	}{
		{"valid", "COV_INT", "42", true, 42},
		{"negative", "COV_INT", "-7", true, -7},
		{"invalid falls back", "COV_INT", "abc", true, 5},
		{"empty falls back", "COV_INT", "", true, 5},
		{"unset falls back", "COV_INT_UNSET", "", false, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.set {
				t.Setenv(tt.key, tt.val)
			}
			if got := envIntOrDefault(tt.key, 5); got != tt.want {
				t.Fatalf("got %d want %d", got, tt.want)
			}
		})
	}
}

func TestGetEnvAsBool(t *testing.T) {
	unsetEnv(t, "COV_BOOL_UNSET")
	tests := []struct {
		name, key, val string
		set            bool
		fallback, want bool
	}{
		{"true", "COV_BOOL", "true", true, false, true},
		{"1", "COV_BOOL", "1", true, false, true},
		{"false overrides true fallback", "COV_BOOL", "false", true, true, false},
		{"invalid falls back true", "COV_BOOL", "maybe", true, true, true},
		{"invalid falls back false", "COV_BOOL", "maybe", true, false, false},
		{"unset falls back", "COV_BOOL_UNSET", "", false, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.set {
				t.Setenv(tt.key, tt.val)
			}
			if got := envBoolOrDefault(tt.key, tt.fallback); got != tt.want {
				t.Fatalf("got %v want %v", got, tt.want)
			}
		})
	}
}
