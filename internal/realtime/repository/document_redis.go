package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	domainEntity "my-go-app/internal/realtime/domain/entity"
	domainRepo "my-go-app/internal/realtime/domain/repository"
)

const dirtyPageKey = "dirty_pages"

const (
	// seqTTL outlives the document TTL so the page sequence survives a document cache expiry.
	// It is refreshed on every successful mutation.
	seqTTL = 7 * 24 * time.Hour
	// maxUpdateRetries bounds optimistic-concurrency retries of UpdateDocument.
	maxUpdateRetries = 20
)

type documentRepository struct {
	client *redis.Client
	ttl    time.Duration
}

// NewDocumentRepository creates a new instance of DocumentRepository using Redis as storage
func NewDocumentRepository(client *redis.Client) domainRepo.DocumentRepository {
	return &documentRepository{
		client: client,
		ttl:    24 * time.Hour,
	}
}

// SetDocument stores the list of elements for a page in Redis
func (r *documentRepository) SetDocument(ctx context.Context, projectID string, pageID string, elements []domainEntity.Element) error {
	key := r.getKey(projectID, pageID)

	data, err := json.Marshal(elements)
	if err != nil {
		return fmt.Errorf("failed to marshal elements: %w", err)
	}

	if err := r.client.Set(ctx, key, data, r.ttl).Err(); err != nil {
		return fmt.Errorf("failed to set value in redis: %w", err)
	}

	return nil
}

// GetDocument retrieves the list of elements for a page from Redis
func (r *documentRepository) GetDocument(ctx context.Context, projectID string, pageID string) ([]domainEntity.Element, error) {
	key := r.getKey(projectID, pageID)

	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get value from redis: %w", err)
	}

	var elements []domainEntity.Element
	if err := json.Unmarshal([]byte(val), &elements); err != nil {
		return nil, fmt.Errorf("failed to unmarshal elements: %w", err)
	}
	// A stored JSON null is an empty document; nil is reserved for a cache miss.
	if elements == nil {
		elements = []domainEntity.Element{}
	}

	return elements, nil
}

// DeleteDocument removes the page elements from Redis
func (r *documentRepository) DeleteDocument(ctx context.Context, projectID string, pageID string) error {
	key := r.getKey(projectID, pageID)

	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete key from redis: %w", err)
	}

	return nil
}

func (r *documentRepository) MarkPageDirty(ctx context.Context, projectID string, pageID string) error {
	val := fmt.Sprintf("%s:%s", projectID, pageID)
	added, err := r.client.SAdd(ctx, dirtyPageKey, val).Result()
	if err != nil {
		// Debug log for troubleshooting SAdd failures.
		fmt.Printf("[DEBUG] documentRepository.MarkPageDirty: SAdd failed key=%s member=%s err=%v\n", dirtyPageKey, val, err)
		return err
	}
	// Debug log showing SAdd result (0 = already present, >0 = new member added)
	fmt.Printf("[DEBUG] documentRepository.MarkPageDirty: SAdd key=%s member=%s added=%d\n", dirtyPageKey, val, added)
	return nil
}

func (r *documentRepository) PopDirtyPages(ctx context.Context, limit int) ([]domainEntity.DirtyPage, error) {
	results, err := r.client.SPopN(ctx, dirtyPageKey, int64(limit)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	var dirtyPages []domainEntity.DirtyPage
	for _, result := range results {
		parts := strings.SplitN(result, ":", 2)
		if len(parts) == 2 {
			dirtyPages = append(dirtyPages, domainEntity.DirtyPage{
				ProjectID: parts[0],
				PageID:    parts[1],
			})
		}
	}
	return dirtyPages, nil
}

// SetDocumentIfAbsent stores the elements only if the page is not cached yet (SET NX).
func (r *documentRepository) SetDocumentIfAbsent(ctx context.Context, projectID string, pageID string, elements []domainEntity.Element) error {
	data, err := json.Marshal(elements)
	if err != nil {
		return fmt.Errorf("failed to marshal elements: %w", err)
	}
	if err := r.client.SetNX(ctx, r.getKey(projectID, pageID), data, r.ttl).Err(); err != nil {
		return fmt.Errorf("failed to set value in redis: %w", err)
	}
	return nil
}

// GetDocumentWithSeq reads the document and its sequence number in one MULTI/EXEC so the pair
// is a consistent snapshot (mutations write both in a single transaction).
func (r *documentRepository) GetDocumentWithSeq(ctx context.Context, projectID string, pageID string) ([]domainEntity.Element, int64, error) {
	key := r.getKey(projectID, pageID)

	pipe := r.client.TxPipeline()
	docCmd := pipe.Get(ctx, key)
	seqCmd := pipe.Get(ctx, r.getSeqKey(projectID, pageID))
	if _, err := pipe.Exec(ctx); err != nil && !errors.Is(err, redis.Nil) {
		return nil, 0, fmt.Errorf("failed to read document snapshot from redis: %w", err)
	}

	var seq int64
	if v, err := seqCmd.Int64(); err == nil {
		seq = v
	} else if !errors.Is(err, redis.Nil) {
		return nil, 0, fmt.Errorf("failed to read page seq: %w", err)
	}

	val, err := docCmd.Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, seq, nil
		}
		return nil, 0, fmt.Errorf("failed to get value from redis: %w", err)
	}
	var elements []domainEntity.Element
	if err := json.Unmarshal([]byte(val), &elements); err != nil {
		return nil, 0, fmt.Errorf("failed to unmarshal elements: %w", err)
	}
	return elements, seq, nil
}

// UpdateDocument performs an optimistic read-modify-write on the page document:
// WATCH the document key, read, mutate, then MULTI { SET doc, INCR seq, SADD dirty } EXEC.
// If another writer (on any server instance) touches the key in between, EXEC aborts and the
// whole read-modify-write is retried against fresh data. The sequence number is incremented
// only inside the successful transaction, so failed or aborted attempts never consume one.
func (r *documentRepository) UpdateDocument(ctx context.Context, projectID string, pageID string, load domainRepo.DocumentLoadFunc, mutate domainRepo.DocumentMutateFunc) (int64, error) {
	key := r.getKey(projectID, pageID)
	seqKey := r.getSeqKey(projectID, pageID)
	dirtyMember := fmt.Sprintf("%s:%s", projectID, pageID)

	var seq int64
	txf := func(tx *redis.Tx) error {
		var elements []domainEntity.Element
		val, err := tx.Get(ctx, key).Result()
		switch {
		case errors.Is(err, redis.Nil):
			if load != nil {
				if elements, err = load(ctx); err != nil {
					return err
				}
			}
		case err != nil:
			return fmt.Errorf("failed to get value from redis: %w", err)
		default:
			if err := json.Unmarshal([]byte(val), &elements); err != nil {
				return fmt.Errorf("failed to unmarshal elements: %w", err)
			}
		}

		updated, err := mutate(elements)
		if err != nil {
			return err
		}
		data, err := json.Marshal(updated)
		if err != nil {
			return fmt.Errorf("failed to marshal elements: %w", err)
		}

		var incr *redis.IntCmd
		if _, err := tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.Set(ctx, key, data, r.ttl)
			incr = pipe.Incr(ctx, seqKey)
			pipe.Expire(ctx, seqKey, seqTTL)
			pipe.SAdd(ctx, dirtyPageKey, dirtyMember)
			return nil
		}); err != nil {
			return err
		}
		seq = incr.Val()
		return nil
	}

	for attempt := 0; attempt < maxUpdateRetries; attempt++ {
		err := r.client.Watch(ctx, txf, key)
		if err == nil {
			return seq, nil
		}
		if !errors.Is(err, redis.TxFailedErr) {
			return 0, err
		}
		// Lost the race against another writer: back off briefly and retry on fresh data.
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-time.After(time.Duration(1+rand.Intn(5<<min(attempt, 4))) * time.Millisecond):
		}
	}
	return 0, fmt.Errorf("document update for page %s aborted after %d conflicting attempts", pageID, maxUpdateRetries)
}

func (r *documentRepository) getSeqKey(projectID string, pageID string) string {
	return r.getKey(projectID, pageID) + ":seq"
}

func (r *documentRepository) getKey(projectID string, pageID string) string {
	return fmt.Sprintf("page:%s:%s", projectID, pageID)
}
