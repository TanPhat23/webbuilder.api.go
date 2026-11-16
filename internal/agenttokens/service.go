package agenttokens

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/lucsky/cuid"

	"my-go-app/internal/models"
)

const MaxActiveTokens = 10

var (
	ErrTooManyTokens = errors.New("too many active tokens: revoke one first")
	ErrTokenNotFound = errors.New("token not found")
	ErrBadName       = errors.New("name must be 1 to 100 characters")
)

type Store interface {
	Create(ctx context.Context, t *models.AgentToken) error
	ListActiveByUser(ctx context.Context, userID string) ([]models.AgentToken, error)
	Revoke(ctx context.Context, userID, id string) (bool, error)
}

type Service struct {
	store Store
	now   func() time.Time
}

func NewService(store Store) *Service { return &Service{store: store, now: time.Now} }

func (s *Service) Create(ctx context.Context, userID, name string) (string, models.AgentToken, error) {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 100 {
		return "", models.AgentToken{}, ErrBadName
	}
	active, err := s.store.ListActiveByUser(ctx, userID)
	if err != nil {
		return "", models.AgentToken{}, err
	}
	if len(active) >= MaxActiveTokens {
		return "", models.AgentToken{}, ErrTooManyTokens
	}
	secret, prefix, hash, err := NewSecret()
	if err != nil {
		return "", models.AgentToken{}, err
	}
	tok := models.AgentToken{Id: cuid.New(), UserId: userID, Name: name, Prefix: prefix, TokenHash: hash, CreatedAt: s.now()}
	if err := s.store.Create(ctx, &tok); err != nil {
		return "", models.AgentToken{}, err
	}
	return secret, tok, nil
}

func (s *Service) List(ctx context.Context, userID string) ([]models.AgentToken, error) {
	return s.store.ListActiveByUser(ctx, userID)
}

func (s *Service) Revoke(ctx context.Context, userID, id string) error {
	ok, err := s.store.Revoke(ctx, userID, id)
	if err != nil {
		return err
	}
	if !ok {
		return ErrTokenNotFound
	}
	return nil
}
