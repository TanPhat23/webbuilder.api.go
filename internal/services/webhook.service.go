package services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/lucsky/cuid"

	"my-go-app/internal/authz"
	"my-go-app/internal/dto"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
)

const (
	maxWebhookAttempts      = 3
	maxWebhookResponseBytes = 2048
)

var webhookRetryBackoffs = []time.Duration{
	5 * time.Second,
	30 * time.Second,
	120 * time.Second,
}

func SignWebhookPayload(secret string, timestamp int64, rawBody []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(strconv.FormatInt(timestamp, 10)))
	mac.Write([]byte("."))
	mac.Write(rawBody)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func generateWebhookSecret() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return "whsec_" + hex.EncodeToString(buf), nil
}

type WebhookService struct {
	webhookRepo     repositories.WebhookRepositoryInterface
	projectRepo     repositories.ProjectRepositoryInterface
	contentTypeRepo repositories.ContentTypeRepositoryInterface
	httpClient      *http.Client
	allowLocal      bool
	queue           chan string
	now             func() time.Time
}

// ValidateWebhookEnv refuses ALLOW_LOCAL_WEBHOOKS=true in production environments.
func ValidateWebhookEnv() error {
	if strings.EqualFold(os.Getenv("ALLOW_LOCAL_WEBHOOKS"), "true") {
		env := strings.ToLower(strings.TrimSpace(os.Getenv("ENV")))
		if env == "production" || env == "prod" {
			return errors.New("ALLOW_LOCAL_WEBHOOKS cannot be enabled when ENV is production")
		}
	}
	return nil
}

func NewWebhookService(
	webhookRepo repositories.WebhookRepositoryInterface,
	projectRepo repositories.ProjectRepositoryInterface,
	contentTypeRepo repositories.ContentTypeRepositoryInterface,
) *WebhookService {
	allowLocal := strings.EqualFold(os.Getenv("ALLOW_LOCAL_WEBHOOKS"), "true")
	if err := ValidateWebhookEnv(); err != nil {
		log.Printf("WARNING: %v; disabling local webhooks", err)
		allowLocal = false
	}
	return NewWebhookServiceWithOptions(webhookRepo, projectRepo, contentTypeRepo, allowLocal, 10*time.Second)
}

func NewWebhookServiceWithOptions(
	webhookRepo repositories.WebhookRepositoryInterface,
	projectRepo repositories.ProjectRepositoryInterface,
	contentTypeRepo repositories.ContentTypeRepositoryInterface,
	allowLocal bool,
	timeout time.Duration,
) *WebhookService {
	return &WebhookService{
		webhookRepo:     webhookRepo,
		projectRepo:     projectRepo,
		contentTypeRepo: contentTypeRepo,
		httpClient:      newWebhookHTTPClient(allowLocal, timeout),
		allowLocal:      allowLocal,
		queue:           make(chan string, 256),
		now:             time.Now,
	}
}

func (s *WebhookService) StartWorkers(ctx context.Context, workerCount int) {
	if workerCount <= 0 {
		workerCount = 4
	}
	for i := 0; i < workerCount; i++ {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case deliveryID := <-s.queue:
					_ = s.DeliverByID(context.Background(), deliveryID)
				}
			}
		}()
	}
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.processPendingRetries(ctx)
			}
		}
	}()
}

func (s *WebhookService) processPendingRetries(ctx context.Context) {
	if s.webhookRepo == nil {
		return
	}
	pending, err := s.webhookRepo.ListPendingRetries(ctx, s.now(), 20)
	if err != nil {
		return
	}
	for _, d := range pending {
		_ = s.DeliverByID(ctx, d.Id)
	}
}

func (s *WebhookService) ListByProjectForUser(ctx context.Context, projectID, userID string) ([]dto.WebhookResponse, error) {
	if _, err := authorizeProjectPerm(ctx, s.projectRepo, projectID, userID, authz.PermCMSView); err != nil {
		return nil, err
	}
	webhooks, err := s.webhookRepo.ListByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.WebhookResponse, 0, len(webhooks))
	for _, wh := range webhooks {
		out = append(out, dto.ToWebhookResponse(wh, false))
	}
	return out, nil
}

func (s *WebhookService) CreateForUser(ctx context.Context, projectID, userID string, req *dto.CreateWebhookRequest) (*dto.WebhookResponse, error) {
	if _, err := authorizeProjectPerm(ctx, s.projectRepo, projectID, userID, authz.PermCMSEdit); err != nil {
		return nil, err
	}
	if err := ValidateWebhookURL(req.Url, s.allowLocal); err != nil {
		return nil, err
	}
	if err := validateWebhookEvents(req.Events); err != nil {
		return nil, err
	}

	var contentTypeID *string
	if req.ContentTypeId != nil && strings.TrimSpace(*req.ContentTypeId) != "" && s.contentTypeRepo != nil {
		ct, err := s.contentTypeRepo.GetContentTypeByProjectAndSlugOrID(ctx, projectID, strings.TrimSpace(*req.ContentTypeId))
		if err != nil {
			return nil, err
		}
		idCopy := ct.Id
		contentTypeID = &idCopy
	}

	secret, err := generateWebhookSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to generate webhook secret: %w", err)
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	wh := &models.ProjectWebhook{
		Id:            cuid.New(),
		ProjectId:     projectID,
		ContentTypeId: contentTypeID,
		Name:          strings.TrimSpace(req.Name),
		Url:           strings.TrimSpace(req.Url),
		Secret:        secret,
		Events:        models.StringSlice(req.Events),
		Enabled:       enabled,
	}
	created, err := s.webhookRepo.Create(ctx, wh)
	if err != nil {
		return nil, err
	}
	resp := dto.ToWebhookResponse(*created, true)
	return &resp, nil
}

func (s *WebhookService) UpdateForUser(ctx context.Context, projectID, webhookID, userID string, req *dto.UpdateWebhookRequest) (*dto.WebhookResponse, error) {
	if _, err := authorizeProjectPerm(ctx, s.projectRepo, projectID, userID, authz.PermCMSEdit); err != nil {
		return nil, err
	}
	existing, err := s.webhookRepo.GetByID(ctx, webhookID)
	if err != nil {
		return nil, err
	}
	if existing.ProjectId != projectID {
		return nil, repositories.ErrWebhookNotFound
	}

	updates := map[string]any{}
	if req.Name != nil {
		updates["Name"] = strings.TrimSpace(*req.Name)
	}
	if req.Url != nil {
		if err := ValidateWebhookURL(*req.Url, s.allowLocal); err != nil {
			return nil, err
		}
		updates["Url"] = strings.TrimSpace(*req.Url)
	}
	if req.Events != nil {
		if err := validateWebhookEvents(*req.Events); err != nil {
			return nil, err
		}
		updates["Events"] = models.StringSlice(*req.Events)
	}
	if req.Enabled != nil {
		updates["Enabled"] = *req.Enabled
	}
	if req.ContentTypeId != nil {
		trimmed := strings.TrimSpace(*req.ContentTypeId)
		if trimmed == "" {
			updates["ContentTypeId"] = nil
		} else if s.contentTypeRepo != nil {
			ct, err := s.contentTypeRepo.GetContentTypeByProjectAndSlugOrID(ctx, projectID, trimmed)
			if err != nil {
				return nil, err
			}
			updates["ContentTypeId"] = ct.Id
		}
	}
	if len(updates) == 0 {
		return nil, errors.New("at least one field must be updated")
	}

	updated, err := s.webhookRepo.Update(ctx, webhookID, updates)
	if err != nil {
		return nil, err
	}
	resp := dto.ToWebhookResponse(*updated, false)
	return &resp, nil
}

func (s *WebhookService) DeleteForUser(ctx context.Context, projectID, webhookID, userID string) error {
	if _, err := authorizeProjectPerm(ctx, s.projectRepo, projectID, userID, authz.PermCMSEdit); err != nil {
		return err
	}
	existing, err := s.webhookRepo.GetByID(ctx, webhookID)
	if err != nil {
		return err
	}
	if existing.ProjectId != projectID {
		return repositories.ErrWebhookNotFound
	}
	return s.webhookRepo.Delete(ctx, webhookID)
}

func (s *WebhookService) ListDeliveriesForUser(ctx context.Context, projectID, webhookID, userID string, limit int) ([]models.WebhookDelivery, error) {
	if _, err := authorizeProjectPerm(ctx, s.projectRepo, projectID, userID, authz.PermCMSView); err != nil {
		return nil, err
	}
	existing, err := s.webhookRepo.GetByID(ctx, webhookID)
	if err != nil {
		return nil, err
	}
	if existing.ProjectId != projectID {
		return nil, repositories.ErrWebhookNotFound
	}
	return s.webhookRepo.ListDeliveriesByWebhook(ctx, webhookID, limit)
}

func (s *WebhookService) SendTestWebhookForUser(ctx context.Context, projectID, webhookID, userID string) (*models.WebhookDelivery, error) {
	if _, err := authorizeProjectPerm(ctx, s.projectRepo, projectID, userID, authz.PermCMSEdit); err != nil {
		return nil, err
	}
	wh, err := s.webhookRepo.GetByID(ctx, webhookID)
	if err != nil {
		return nil, err
	}
	if wh.ProjectId != projectID {
		return nil, repositories.ErrWebhookNotFound
	}

	deliveryID := cuid.New()
	payloadMap := map[string]any{
		"id":        deliveryID,
		"event":     models.WebhookEventTest,
		"timestamp": s.now().UTC().Format(time.RFC3339),
		"projectId": projectID,
		"test":      true,
	}
	raw, err := json.Marshal(payloadMap)
	if err != nil {
		return nil, err
	}
	delivery := &models.WebhookDelivery{
		Id:        deliveryID,
		WebhookId: wh.Id,
		ProjectId: projectID,
		Event:     models.WebhookEventTest,
		Payload:   raw,
		Status:    models.WebhookDeliveryStatusPending,
	}
	created, err := s.webhookRepo.CreateDelivery(ctx, delivery)
	if err != nil {
		return nil, err
	}
	_ = s.DeliverByID(ctx, created.Id)
	latest, err := s.webhookRepo.GetDeliveryByID(ctx, created.Id)
	if err != nil {
		return created, nil
	}
	return latest, nil
}

func (s *WebhookService) EmitItemEvent(ctx context.Context, projectID string, ct *models.ContentType, event string, item *models.ContentItem) {
	if s.webhookRepo == nil || ct == nil || item == nil || projectID == "" {
		return
	}
	webhooks, err := s.webhookRepo.FindMatchingForEvent(ctx, projectID, ct.Id, event)
	if err != nil || len(webhooks) == 0 {
		return
	}

	flattened := FlattenContentItem(item)
	for _, wh := range webhooks {
		deliveryID := cuid.New()
		payloadMap := map[string]any{
			"id":        deliveryID,
			"event":     event,
			"timestamp": s.now().UTC().Format(time.RFC3339),
			"projectId": projectID,
			"collection": map[string]any{
				"id":   ct.Id,
				"slug": ct.Slug,
				"name": ct.Name,
			},
			"item": flattened,
		}
		raw, err := json.Marshal(payloadMap)
		if err != nil {
			continue
		}
		delivery := &models.WebhookDelivery{
			Id:        deliveryID,
			WebhookId: wh.Id,
			ProjectId: projectID,
			Event:     event,
			Payload:   raw,
			Status:    models.WebhookDeliveryStatusPending,
		}
		created, err := s.webhookRepo.CreateDelivery(ctx, delivery)
		if err != nil {
			continue
		}
		select {
		case s.queue <- created.Id:
		default:
			// Queue is full; leave pending in DB for the retry poller instead of spawning unbounded goroutines.
			next := s.now()
			_ = s.webhookRepo.UpdateDelivery(ctx, created.Id, map[string]any{
				"NextRetryAt": next,
			})
		}
	}
}

func (s *WebhookService) DeliverByID(ctx context.Context, deliveryID string) error {
	delivery, err := s.webhookRepo.GetDeliveryByID(ctx, deliveryID)
	if err != nil {
		return err
	}
	wh, err := s.webhookRepo.GetByID(ctx, delivery.WebhookId)
	if err != nil {
		return err
	}

	attempt := delivery.Attempts + 1
	now := s.now()
	ts := now.Unix()
	sig := SignWebhookPayload(wh.Secret, ts, delivery.Payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, wh.Url, bytes.NewReader(delivery.Payload))
	if err != nil {
		return s.recordDeliveryFailure(ctx, delivery.Id, attempt, nil, "", err.Error(), false)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "WebBuilder-Webhooks/1.0")
	req.Header.Set("X-WebBuilder-Event", delivery.Event)
	req.Header.Set("X-WebBuilder-Delivery-Id", delivery.Id)
	req.Header.Set("X-WebBuilder-Timestamp", strconv.FormatInt(ts, 10))
	req.Header.Set("X-WebBuilder-Signature", sig)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return s.recordDeliveryFailure(ctx, delivery.Id, attempt, nil, "", err.Error(), true)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, maxWebhookResponseBytes))
	bodyStr := string(bodyBytes)
	statusCode := resp.StatusCode

	if statusCode >= 200 && statusCode < 300 {
		completed := s.now()
		return s.webhookRepo.UpdateDelivery(ctx, delivery.Id, map[string]any{
			"Status":       models.WebhookDeliveryStatusSucceeded,
			"HttpStatus":   statusCode,
			"ResponseBody": bodyStr,
			"ErrorMessage": nil,
			"Attempts":     attempt,
			"NextRetryAt":  nil,
			"CompletedAt":  completed,
		})
	}

	retryable := statusCode >= 500 || statusCode == http.StatusTooManyRequests
	return s.recordDeliveryFailure(ctx, delivery.Id, attempt, &statusCode, bodyStr, fmt.Sprintf("HTTP %d", statusCode), retryable)
}

func (s *WebhookService) recordDeliveryFailure(ctx context.Context, deliveryID string, attempt int, httpStatus *int, responseBody, errMsg string, retryable bool) error {
	now := s.now()
	updates := map[string]any{
		"Attempts":     attempt,
		"HttpStatus":   httpStatus,
		"ResponseBody": responseBody,
		"ErrorMessage": errMsg,
	}
	if retryable && attempt < maxWebhookAttempts {
		idx := attempt - 1
		if idx >= len(webhookRetryBackoffs) {
			idx = len(webhookRetryBackoffs) - 1
		}
		next := now.Add(webhookRetryBackoffs[idx])
		updates["Status"] = models.WebhookDeliveryStatusPending
		updates["NextRetryAt"] = next
	} else {
		updates["Status"] = models.WebhookDeliveryStatusFailed
		updates["NextRetryAt"] = nil
		updates["CompletedAt"] = now
	}
	return s.webhookRepo.UpdateDelivery(ctx, deliveryID, updates)
}

func validateWebhookEvents(events []string) error {
	if len(events) == 0 {
		return errors.New("at least one webhook event is required")
	}
	for _, ev := range events {
		if !models.IsValidWebhookEvent(ev) {
			return fmt.Errorf("invalid webhook event %q", ev)
		}
	}
	return nil
}
