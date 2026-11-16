package agenttokens

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"my-go-app/internal/models"
)

type memStore struct{ tokens []models.AgentToken }

func (m *memStore) Create(_ context.Context, t *models.AgentToken) error {
	m.tokens = append(m.tokens, *t)
	return nil
}
func (m *memStore) ListActiveByUser(_ context.Context, u string) ([]models.AgentToken, error) {
	var out []models.AgentToken
	for _, t := range m.tokens {
		if t.UserId == u && t.RevokedAt == nil {
			out = append(out, t)
		}
	}
	return out, nil
}
func (m *memStore) Revoke(_ context.Context, u, id string) (bool, error) {
	for i := range m.tokens {
		if m.tokens[i].Id == id && m.tokens[i].UserId == u && m.tokens[i].RevokedAt == nil {
			now := time.Now()
			m.tokens[i].RevokedAt = &now
			return true, nil
		}
	}
	return false, nil
}

// The MCP service has the same vector in its own test. If these ever differ,
// tokens created here stop verifying there.
func TestHashSecretMatchesTheMCPServicesHash(t *testing.T) {
	const want = "a1069d5e4e69aed3fcf506b1202976b27308b6a2ae9fad0f2ee38af34ecc4604"
	if got := HashSecret("wbk_fixture"); got != want {
		t.Fatalf("HashSecret(wbk_fixture) = %s, want %s", got, want)
	}
}

func TestSecretShapeAndHash(t *testing.T) {
	secret, prefix, hash, err := NewSecret()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(secret, TokenPrefix) || len(secret) != len(TokenPrefix)+43 {
		t.Fatalf("secret %q", secret)
	}
	if !strings.HasPrefix(secret, prefix) || len(prefix) != len(TokenPrefix)+4 {
		t.Fatalf("prefix %q", prefix)
	}
	if hash != HashSecret(secret) || len(hash) != 64 || strings.Contains(hash, secret) {
		t.Fatalf("hash %q", hash)
	}
	other, _, _, _ := NewSecret()
	if other == secret {
		t.Fatal("secrets must differ")
	}
}

func TestCreateStoresOnlyTheHash(t *testing.T) {
	store := &memStore{}
	secret, tok, err := NewService(store).Create(context.Background(), "user-1", "  My laptop ")
	if err != nil {
		t.Fatal(err)
	}
	if tok.Name != "My laptop" || tok.UserId != "user-1" {
		t.Fatalf("token %+v", tok)
	}
	if store.tokens[0].TokenHash != HashSecret(secret) || store.tokens[0].TokenHash == secret {
		t.Fatal("only the hash may be stored")
	}
}

func TestCreateRejectsBadNamesAndTooManyTokens(t *testing.T) {
	svc := NewService(&memStore{})
	for _, name := range []string{"", "   ", strings.Repeat("x", 101)} {
		if _, _, err := svc.Create(context.Background(), "u", name); !errors.Is(err, ErrBadName) {
			t.Fatalf("name %q: %v", name, err)
		}
	}
	for i := 0; i < MaxActiveTokens; i++ {
		if _, _, err := svc.Create(context.Background(), "u", "t"); err != nil {
			t.Fatal(err)
		}
	}
	if _, _, err := svc.Create(context.Background(), "u", "one too many"); !errors.Is(err, ErrTooManyTokens) {
		t.Fatalf("got %v", err)
	}
}

func TestRevokeIsOwnerScopedAndRevokedTokensLeaveTheList(t *testing.T) {
	svc := NewService(&memStore{})
	_, tok, _ := svc.Create(context.Background(), "owner", "cli")
	if err := svc.Revoke(context.Background(), "someone-else", tok.Id); !errors.Is(err, ErrTokenNotFound) {
		t.Fatalf("got %v", err)
	}
	if err := svc.Revoke(context.Background(), "owner", tok.Id); err != nil {
		t.Fatal(err)
	}
	if list, _ := svc.List(context.Background(), "owner"); len(list) != 0 {
		t.Fatalf("revoked token still listed: %+v", list)
	}
	if err := svc.Revoke(context.Background(), "owner", tok.Id); !errors.Is(err, ErrTokenNotFound) {
		t.Fatalf("revoking twice: %v", err)
	}
}
