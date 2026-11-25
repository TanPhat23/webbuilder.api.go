package repositories

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"my-go-app/internal/models"
)

type AgentTokenRepository struct{ db *gorm.DB }

func NewAgentTokenRepository(db *gorm.DB) AgentTokenRepositoryInterface {
	return &AgentTokenRepository{db: db}
}

func (r *AgentTokenRepository) Create(ctx context.Context, t *models.AgentToken) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *AgentTokenRepository) ListActiveByUser(ctx context.Context, userID string) ([]models.AgentToken, error) {
	var out []models.AgentToken
	err := r.db.WithContext(ctx).
		Where(`"UserId" = ? AND "RevokedAt" IS NULL`, userID).
		Order(`"CreatedAt" DESC`).
		Find(&out).Error
	return out, err
}

// Revoke marks the user's token revoked. It reports false when there is no such
// active token for that user.
func (r *AgentTokenRepository) Revoke(ctx context.Context, userID, id string) (bool, error) {
	res := r.db.WithContext(ctx).Model(&models.AgentToken{}).
		Where(`"Id" = ? AND "UserId" = ? AND "RevokedAt" IS NULL`, id, userID).
		Update("RevokedAt", time.Now())
	return res.RowsAffected > 0, res.Error
}

func (r *AgentTokenRepository) GetActiveByHash(ctx context.Context, hash string) (*models.AgentToken, error) {
	var out models.AgentToken
	err := r.db.WithContext(ctx).
		Where(`"TokenHash" = ? AND "RevokedAt" IS NULL`, hash).
		First(&out).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &out, nil
}

func (r *AgentTokenRepository) Touch(ctx context.Context, id string, at time.Time) error {
	return r.db.WithContext(ctx).
		Model(&models.AgentToken{}).
		Where(`"Id" = ?`, id).
		Update("LastUsedAt", at).Error
}
