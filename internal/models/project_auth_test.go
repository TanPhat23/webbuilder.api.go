package models

import (
	"encoding/base64"
	"encoding/json"
	"testing"
)

func makeFakeJWT(role string) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"role":"` + role + `","iss":"supabase"}`))
	return header + "." + payload + ".sig"
}

func TestValidateProjectAuthConfig(t *testing.T) {
	anonJWT := makeFakeJWT("anon")
	serviceJWT := makeFakeJWT("service_role")

	t.Run("accepts valid supabase config and normalizes defaults", func(t *testing.T) {
		raw := json.RawMessage(`{"provider":"supabase","supabaseUrl":"https://AbCd1234.supabase.co/","supabaseAnonKey":"` + anonJWT + `"}`)
		got, err := ValidateProjectAuthConfig(raw)
		if err != nil || got == nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var cfg ProjectAuthConfig
		if err := json.Unmarshal(*got, &cfg); err != nil {
			t.Fatal(err)
		}
		if cfg.SupabaseURL != "https://abcd1234.supabase.co" || cfg.LoginPath != "/login" || cfg.AfterLoginPath != "/" {
			t.Fatalf("unexpected normalized config: %+v", cfg)
		}
	})

	t.Run("null clears auth config", func(t *testing.T) {
		got, err := ValidateProjectAuthConfig(json.RawMessage(`null`))
		if err != nil || got != nil {
			t.Fatalf("got=%v err=%v, want nil, nil", got, err)
		}
	})

	cases := map[string]string{
		"non-https":          `{"provider":"supabase","supabaseUrl":"http://demo.supabase.co","supabaseAnonKey":"sb_publishable_123"}`,
		"ssrf localhost":     `{"provider":"supabase","supabaseUrl":"https://localhost","supabaseAnonKey":"sb_publishable_123"}`,
		"ssrf metadata ip":   `{"provider":"supabase","supabaseUrl":"https://169.254.169.254","supabaseAnonKey":"sb_publishable_123"}`,
		"ssrf arbitrary":     `{"provider":"supabase","supabaseUrl":"https://evil.example.com","supabaseAnonKey":"sb_publishable_123"}`,
		"port rejected":      `{"provider":"supabase","supabaseUrl":"https://demo.supabase.co:8443","supabaseAnonKey":"sb_publishable_123"}`,
		"path rejected":      `{"provider":"supabase","supabaseUrl":"https://demo.supabase.co/extra","supabaseAnonKey":"sb_publishable_123"}`,
		"sb_secret rejected": `{"provider":"supabase","supabaseUrl":"https://demo.supabase.co","supabaseAnonKey":"sb_secret_abc123"}`,
		"service_role jwt":   `{"provider":"supabase","supabaseUrl":"https://demo.supabase.co","supabaseAnonKey":"` + serviceJWT + `"}`,
		"open redirect":      `{"provider":"supabase","supabaseUrl":"https://demo.supabase.co","supabaseAnonKey":"sb_publishable_123","loginPath":"//evil.com"}`,
		"unknown provider":   `{"provider":"custom","supabaseUrl":"https://demo.supabase.co","supabaseAnonKey":"sb_publishable_123"}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := ValidateProjectAuthConfig(json.RawMessage(body)); err == nil {
				t.Fatalf("expected %s to be rejected", name)
			}
		})
	}
}
