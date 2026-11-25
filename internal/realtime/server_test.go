package realtime

import (
	"testing"

	"my-go-app/internal/realtime/config"
)

func TestValidateRuntimeConfig_DevModeGuards(t *testing.T) {
	t.Setenv("ENV", "")
	if err := validateRuntimeConfig(&config.Config{DevMode: true}); err != nil {
		t.Fatalf("dev mode with empty ENV and empty clerk key must be allowed: %v", err)
	}

	for _, env := range []string{"production", "PRODUCTION", "prod", "staging"} {
		t.Setenv("ENV", env)
		if err := validateRuntimeConfig(&config.Config{DevMode: true}); err == nil {
			t.Fatalf("dev mode must be rejected when ENV=%q", env)
		}
	}

	t.Setenv("ENV", "")
	if err := validateRuntimeConfig(&config.Config{DevMode: true, ClerkSecretKey: "sk_live_abc"}); err == nil {
		t.Fatal("dev mode must be rejected with sk_live_ ClerkSecretKey")
	}

	t.Setenv("ENV", "production")
	if err := validateRuntimeConfig(&config.Config{DevMode: false, ClerkSecretKey: "sk_live_abc"}); err != nil {
		t.Fatalf("non-dev mode in production must be allowed: %v", err)
	}
}
