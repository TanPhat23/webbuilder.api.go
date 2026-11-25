package repositories

import (
	"context"
	"errors"
	"fmt"
	"my-go-app/internal/models"
	"time"

	"github.com/lucsky/cuid"
	"gorm.io/gorm"
)

var ErrWebhookNotFound = errors.New("webhook not found")

type WebhookRepository struct {
	db *gorm.DB
}

func NewWebhookRepository(db *gorm.DB) WebhookRepositoryInterface {
	return &WebhookRepository{db: db}
}

func (r *WebhookRepository) Create(ctx context.Context, webhook *models.ProjectWebhook) (*models.ProjectWebhook, error) {
	if webhook == nil {
		return nil, errors.New("webhook cannot be nil")
	}
	if webhook.Id == "" {
		webhook.Id = cuid.New()
	}
	now := time.Now()
	webhook.CreatedAt = now
	webhook.UpdatedAt = now
	if err := r.db.WithContext(ctx).Create(webhook).Error; err != nil {
		return nil, fmt.Errorf("failed to create webhook: %w", err)
	}
	return webhook, nil
}

func (r *WebhookRepository) ListByProject(ctx context.Context, projectID string) ([]models.ProjectWebhook, error) {
	var out []models.ProjectWebhook
	err := r.db.WithContext(ctx).
		Where(`"ProjectId" = ?`, projectID).
		Order(`"CreatedAt" DESC`).
		Find(&out).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list webhooks: %w", err)
	}
	return out, nil
}

func (r *WebhookRepository) GetByID(ctx context.Context, id string) (*models.ProjectWebhook, error) {
	var out models.ProjectWebhook
	err := r.db.WithContext(ctx).Where(`"Id" = ?`, id).First(&out).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWebhookNotFound
		}
		return nil, fmt.Errorf("failed to get webhook: %w", err)
	}
	return &out, nil
}

func (r *WebhookRepository) Update(ctx context.Context, id string, updates map[string]any) (*models.ProjectWebhook, error) {
	updates["UpdatedAt"] = time.Now()
	res := r.db.WithContext(ctx).
		Model(&models.ProjectWebhook{}).
		Where(`"Id" = ?`, id).
		Updates(updates)
	if res.Error != nil {
		return nil, fmt.Errorf("failed to update webhook: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return nil, ErrWebhookNotFound
	}
	return r.GetByID(ctx, id)
}

func (r *WebhookRepository) Delete(ctx context.Context, id string) error {
	res := r.db.WithContext(ctx).Where(`"Id" = ?`, id).Delete(&models.ProjectWebhook{})
	if res.Error != nil {
		return fmt.Errorf("failed to delete webhook: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrWebhookNotFound
	}
	return nil
}

func (r *WebhookRepository) ListMatchingForEvent(ctx context.Context, projectID, contentTypeID, event string) ([]models.ProjectWebhook, error) {
	var all []models.ProjectWebhook
	err := r.db.WithContext(ctx).
		Where(`"ProjectId" = ? AND "Enabled" = true AND ("ContentTypeId" IS NULL OR "ContentTypeId" = ?)`, projectID, contentTypeID).
		Find(&all).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query matching webhooks: %w", err)
	}
	matched := make([]models.ProjectWebhook, 0, len(all))
	for _, wh := range all {
		if wh.SubscribesTo(event) {
			matched = append(matched, wh)
		}
	}
	return matched, nil
}

func (r *WebhookRepository) CreateDelivery(ctx context.Context, delivery *models.WebhookDelivery) (*models.WebhookDelivery, error) {
	if delivery.Id == "" {
		delivery.Id = cuid.New()
	}
	if delivery.CreatedAt.IsZero() {
		delivery.CreatedAt = time.Now()
	}
	if err := r.db.WithContext(ctx).Create(delivery).Error; err != nil {
		return nil, fmt.Errorf("failed to create webhook delivery: %w", err)
	}
	return delivery, nil
}

func (r *WebhookRepository) GetDeliveryByID(ctx context.Context, id string) (*models.WebhookDelivery, error) {
	var out models.WebhookDelivery
	err := r.db.WithContext(ctx).Where(`"Id" = ?`, id).First(&out).Error
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *WebhookRepository) UpdateDelivery(ctx context.Context, id string, updates map[string]any) error {
	return r.db.WithContext(ctx).
		Model(&models.WebhookDelivery{}).
		Where(`"Id" = ?`, id).
		Updates(updates).Error
}

func (r *WebhookRepository) ListDeliveriesByWebhook(ctx context.Context, webhookID string, limit int) ([]models.WebhookDelivery, error) {
	if limit <= 0 || limit > 100 {
		limit = 25
	}
	var out []models.WebhookDelivery
	err := r.db.WithContext(ctx).
		Where(`"WebhookId" = ?`, webhookID).
		Order(`"CreatedAt" DESC`).
		Limit(limit).
		Find(&out).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list webhook deliveries: %w", err)
	}
	return out, nil
}

func (r *WebhookRepository) ListPendingRetries(ctx context.Context, now time.Time, limit int) ([]models.WebhookDelivery, error) {
	if limit <= 0 {
		limit = 20
	}
	var out []models.WebhookDelivery
	err := r.db.WithContext(ctx).
		Where(`"Status" = ? AND "NextRetryAt" IS NOT NULL AND "NextRetryAt" <= ?`, models.WebhookDeliveryStatusPending, now).
		Order(`"NextRetryAt" ASC`).
		Limit(limit).
		Find(&out).Error
	if err != nil {
		return nil, err
	}
	return out, nil
}
