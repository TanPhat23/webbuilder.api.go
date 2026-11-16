// Package agenttokens creates, lists and revokes the personal access tokens
// that Claude Code and Codex CLI present to the MCP service.
package agenttokens

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

const TokenPrefix = "wbk_"

// NewSecret returns a new token secret, the short prefix shown in the token
// list, and the hash that is stored.
func NewSecret() (secret, prefix, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", "", err
	}
	secret = TokenPrefix + base64.RawURLEncoding.EncodeToString(b)
	return secret, secret[:len(TokenPrefix)+4], HashSecret(secret), nil
}

func HashSecret(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}
