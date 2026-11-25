package services

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/lucsky/cuid"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
)

var (
	ErrCollectionPublicReadDisabled   = errors.New("public read is disabled for this collection")
	ErrCollectionPublicCreateDisabled = errors.New("public form submissions are disabled for this collection")
	ErrCollectionRateLimited          = errors.New("rate limit exceeded")
)

type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string {
	return "validation failed"
}

type PublicItemsPage struct {
	Items  []map[string]any `json:"items"`
	Total  int              `json:"total"`
	Limit  int              `json:"limit"`
	Offset int              `json:"offset"`
}

type PublicSubmitResult struct {
	ID        string    `json:"id"`
	Slug      string    `json:"slug,omitempty"`
	Published bool      `json:"published,omitempty"`
	Accepted  bool      `json:"accepted"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
}

// ItemEventEmitter is implemented by WebhookService in Task 4.
type ItemEventEmitter interface {
	EmitItemEvent(ctx context.Context, projectID string, ct *models.ContentType, event string, item *models.ContentItem)
}

const (
	maxPublicCollectionOffset = 10000
	maxRateLimitBuckets       = 50000
	maxPublicFieldLen         = 16 * 1024
	maxPublicRichTextFieldLen = 100 * 1024
	rateLimitPruneInterval    = 1024
)

type slidingWindowLimiter struct {
	mu      sync.Mutex
	buckets map[string][]time.Time
	ops     uint64
}

func newSlidingWindowLimiter() *slidingWindowLimiter {
	return &slidingWindowLimiter{buckets: make(map[string][]time.Time)}
}

func (l *slidingWindowLimiter) evictExpiredLocked(cutoff time.Time) {
	for k, hits := range l.buckets {
		kept := hits[:0]
		for _, ts := range hits {
			if ts.After(cutoff) {
				kept = append(kept, ts)
			}
		}
		if len(kept) == 0 {
			delete(l.buckets, k)
		} else {
			l.buckets[k] = kept
		}
	}
}

func (l *slidingWindowLimiter) Allow(key string, limit int, window time.Duration, now time.Time) bool {
	if limit <= 0 {
		limit = 30
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	cutoff := now.Add(-window)
	l.ops++
	if l.ops%rateLimitPruneInterval == 0 {
		l.evictExpiredLocked(cutoff)
	}

	hits := l.buckets[key]
	kept := hits[:0]
	for _, ts := range hits {
		if ts.After(cutoff) {
			kept = append(kept, ts)
		}
	}
	if len(kept) >= limit {
		l.buckets[key] = kept
		return false
	}
	if len(kept) == 0 {
		delete(l.buckets, key)
	}

	if _, exists := l.buckets[key]; !exists && len(l.buckets) >= maxRateLimitBuckets {
		l.evictExpiredLocked(cutoff)
		if len(l.buckets) >= maxRateLimitBuckets {
			for k := range l.buckets {
				delete(l.buckets, k)
				if len(l.buckets) < maxRateLimitBuckets {
					break
				}
			}
		}
	}

	l.buckets[key] = append(kept, now)
	return true
}

type PublicCollectionService struct {
	contentTypeRepo repositories.ContentTypeRepositoryInterface
	contentItemRepo repositories.ContentItemRepositoryInterface
	limiter         *slidingWindowLimiter
	emitter         ItemEventEmitter
	now             func() time.Time
}

func NewPublicCollectionService(
	contentTypeRepo repositories.ContentTypeRepositoryInterface,
	contentItemRepo repositories.ContentItemRepositoryInterface,
) *PublicCollectionService {
	return &PublicCollectionService{
		contentTypeRepo: contentTypeRepo,
		contentItemRepo: contentItemRepo,
		limiter:         newSlidingWindowLimiter(),
		now:             time.Now,
	}
}

func (s *PublicCollectionService) WithEventEmitter(emitter ItemEventEmitter) *PublicCollectionService {
	s.emitter = emitter
	return s
}

func (s *PublicCollectionService) ListPublicItems(ctx context.Context, projectID, collectionSlugOrID string, limit, offset int, sortBy, sortOrder string) (*PublicItemsPage, error) {
	ct, err := s.contentTypeRepo.GetByProjectAndSlugOrID(ctx, projectID, collectionSlugOrID)
	if err != nil {
		return nil, err
	}
	if !ct.PublicRead {
		return nil, ErrCollectionPublicReadDisabled
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	if offset > maxPublicCollectionOffset {
		offset = maxPublicCollectionOffset
	}

	items, err := s.contentItemRepo.ListPublic(ctx, ct.Id, limit+offset, sortBy, sortOrder)
	if err != nil {
		return nil, err
	}
	total := len(items)
	if offset >= len(items) {
		items = []models.ContentItem{}
	} else {
		items = items[offset:]
		if len(items) > limit {
			items = items[:limit]
		}
	}

	flattened := make([]map[string]any, 0, len(items))
	for i := range items {
		flattened = append(flattened, FlattenContentItem(&items[i]))
	}
	return &PublicItemsPage{
		Items:  flattened,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

func (s *PublicCollectionService) SubmitPublicItem(ctx context.Context, projectID, collectionSlugOrID, clientIP string, raw map[string]any) (*PublicSubmitResult, error) {
	ct, err := s.contentTypeRepo.GetByProjectAndSlugOrID(ctx, projectID, collectionSlugOrID)
	if err != nil {
		return nil, err
	}
	if !ct.PublicCreate {
		return nil, ErrCollectionPublicCreateDisabled
	}

	limitKey := ct.Id + ":" + clientIP
	if !s.limiter.Allow(limitKey, ct.RateLimitPerMin, time.Minute, s.now()) {
		return nil, ErrCollectionRateLimited
	}

	// Silent honeypot check: accept without writing or firing webhooks.
	if hp, ok := raw["_hp"].(string); ok && strings.TrimSpace(hp) != "" {
		return &PublicSubmitResult{ID: "ok", Accepted: true}, nil
	}

	normalized, fieldErrs := ValidateSubmissionFields(ct.Fields, raw)
	if len(fieldErrs) > 0 {
		return nil, &ValidationError{Fields: fieldErrs}
	}

	itemID := cuid.New()
	title := resolveSubmissionTitle(raw, ct.Fields, normalized)
	slug := resolveSubmissionSlug(raw, ct.Slug, itemID)

	fieldValues := make([]models.ContentFieldValue, 0, len(normalized))
	for _, field := range ct.Fields {
		val, ok := normalized[field.Name]
		if !ok {
			continue
		}
		v := val
		fieldValues = append(fieldValues, models.ContentFieldValue{
			FieldId: field.Id,
			Value:   &v,
		})
	}

	item := &models.ContentItem{
		Id:            itemID,
		ContentTypeId: ct.Id,
		Title:         title,
		Slug:          slug,
		Published:     ct.DefaultPublished,
	}

	created, err := s.contentItemRepo.Create(ctx, item, fieldValues)
	if err != nil {
		return nil, err
	}

	if s.emitter != nil {
		// Ensure Field.Name is populated on FieldValues for webhook payload flattening.
		fieldByID := make(map[string]models.ContentField, len(ct.Fields))
		for _, f := range ct.Fields {
			fieldByID[f.Id] = f
		}
		for i := range created.FieldValues {
			if created.FieldValues[i].Field == nil || created.FieldValues[i].Field.Name == "" {
				if f, ok := fieldByID[created.FieldValues[i].FieldId]; ok {
					fCopy := f
					created.FieldValues[i].Field = &fCopy
				}
			}
		}
		s.emitter.EmitItemEvent(ctx, projectID, ct, models.WebhookEventItemCreated, created)
	}

	return &PublicSubmitResult{
		ID:        created.Id,
		Slug:      created.Slug,
		Published: created.Published,
		Accepted:  true,
		CreatedAt: created.CreatedAt,
	}, nil
}

func ValidateSubmissionFields(fields []models.ContentField, input map[string]any) (map[string]string, map[string]string) {
	normalized := make(map[string]string, len(fields))
	errs := make(map[string]string)

	for _, f := range fields {
		rawVal, exists := input[f.Name]
		strVal, empty := stringifySubmissionValue(rawVal)
		if !exists || empty {
			if f.Required {
				errs[f.Name] = "this field is required"
			}
			continue
		}
		fieldType := strings.ToLower(strings.TrimSpace(f.Type))
		maxLen := maxPublicFieldLen
		if fieldType == "richtext" {
			maxLen = maxPublicRichTextFieldLen
		}
		if len(strVal) > maxLen {
			errs[f.Name] = fmt.Sprintf("must be at most %d bytes", maxLen)
			continue
		}
		if errMsg := validateFieldType(fieldType, strVal); errMsg != "" {
			errs[f.Name] = errMsg
			continue
		}
		normalized[f.Name] = strVal
	}
	return normalized, errs
}

func stringifySubmissionValue(v any) (string, bool) {
	if v == nil {
		return "", true
	}
	switch val := v.(type) {
	case string:
		trimmed := strings.TrimSpace(val)
		return trimmed, trimmed == ""
	case bool:
		return strconv.FormatBool(val), false
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64), false
	case int:
		return strconv.Itoa(val), false
	default:
		s := strings.TrimSpace(fmt.Sprint(val))
		return s, s == ""
	}
}

func validateFieldType(fieldType, val string) string {
	switch fieldType {
	case "email":
		addr, err := mail.ParseAddress(val)
		if err != nil || addr.Address != val || !strings.Contains(val, "@") {
			return "must be a valid email address"
		}
	case "number", "integer", "float":
		if _, err := strconv.ParseFloat(val, 64); err != nil {
			return "must be a valid number"
		}
	case "boolean", "bool":
		if _, err := strconv.ParseBool(strings.ToLower(val)); err != nil {
			return "must be true or false"
		}
	case "url":
		u, err := url.ParseRequestURI(val)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return "must be a valid http or https URL"
		}
	case "date", "datetime":
		if _, err := time.Parse(time.RFC3339, val); err != nil {
			if _, err2 := time.Parse("2006-01-02", val); err2 != nil {
				return "must be a valid ISO-8601 date"
			}
		}
	}
	return ""
}

func resolveSubmissionTitle(raw map[string]any, fields []models.ContentField, normalized map[string]string) string {
	if explicit, ok := raw["title"].(string); ok && strings.TrimSpace(explicit) != "" {
		return truncateString(strings.TrimSpace(explicit), 120)
	}
	for _, f := range fields {
		t := strings.ToLower(strings.TrimSpace(f.Type))
		if t == "text" || t == "string" || t == "email" || t == "" {
			if v, ok := normalized[f.Name]; ok && v != "" {
				return truncateString(v, 120)
			}
		}
	}
	return "Submission"
}

func resolveSubmissionSlug(raw map[string]any, collectionSlug, itemID string) string {
	if explicit, ok := raw["slug"].(string); ok && strings.TrimSpace(explicit) != "" {
		return models.SlugifyCollectionName(explicit, itemID)
	}
	suffix := itemID
	if len(suffix) > 8 {
		suffix = suffix[len(suffix)-8:]
	}
	base := collectionSlug
	if base == "" {
		base = "item"
	}
	return base + "-" + suffix
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

func FlattenContentItem(item *models.ContentItem) map[string]any {
	flattened := map[string]any{
		"contentTypeId": item.ContentTypeId,
		"createdAt":     item.CreatedAt,
		"id":            item.Id,
		"published":     item.Published,
		"slug":          item.Slug,
		"title":         item.Title,
		"updatedAt":     item.UpdatedAt,
	}
	for _, fv := range item.FieldValues {
		if fv.Field != nil && fv.Field.Name != "" {
			flattened[fv.Field.Name] = fv.Value
		}
	}
	return flattened
}
