package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	domainEntity "my-go-app/internal/realtime/domain/entity"
	domainErrors "my-go-app/internal/realtime/domain/errors"
	domainRepo "my-go-app/internal/realtime/domain/repository"
	elementutils "my-go-app/internal/realtime/utils/element"
)

const dirtyPageKey = "dirty_pages"

const (
	// seqTTL outlives the document TTL so the page sequence survives a document cache expiry.
	// It is refreshed on every successful mutation.
	seqTTL = 7 * 24 * time.Hour
	// maxUpdateRetries bounds optimistic-concurrency retries of an update.
	maxUpdateRetries = 20
	// structField is the hash field that marks a page as cached and holds its structure
	// version. It is bumped by every whole-document update (create, move, remove, ...) so two
	// of those can never interleave, while per-element patches leave it alone and only
	// conflict with a writer of the same element. Element ids never collide with it.
	structField = "_sv"
)

// Page documents are stored as one Redis hash per page: the field is the element id and the value
// is that element as flat JSON (no nested children; the tree comes from ParentId and the sibling
// order keys). A patch therefore reads and writes one small value instead of the whole page.
//
// Concurrency is optimistic and server-side: each write is a Lua script that checks the exact
// previous value of every field it touches (and the structure version for whole-document updates)
// and applies atomically, so it is safe across server instances.
//
// Pages cached before this layout existed live in a single JSON string under the legacy key; they
// are still read, and are migrated into the hash by the next whole-document update.
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

// storedDoc is a page as read from Redis.
type storedDoc struct {
	raw     map[string]string // element id -> flat element JSON
	sv      string            // structure version; "" when the page is not in the hash
	present bool              // false on a cache miss
	legacy  bool              // read from the pre-hash single-string key
}

// applyDocumentScript applies a whole-document diff if nothing it read has changed.
// KEYS: hash, seq, dirty set, legacy key
// ARGV: hashTTL, seqTTL, dirtyMember, expectedSv, nSets, nDels, (id, old, new)*nSets, (id, old)*nDels
// Returns the new sequence number, or -1 when another writer got there first.
var applyDocumentScript = redis.NewScript(`
local sv = redis.call('HGET', KEYS[1], '` + structField + `')
if sv == false then sv = '' end
if sv ~= ARGV[4] then return -1 end
local nSets, nDels = tonumber(ARGV[5]), tonumber(ARGV[6])
local p = 7
for i = 1, nSets do
  local cur = redis.call('HGET', KEYS[1], ARGV[p])
  if cur == false then cur = '' end
  if cur ~= ARGV[p + 1] then return -1 end
  p = p + 3
end
for i = 1, nDels do
  local cur = redis.call('HGET', KEYS[1], ARGV[p])
  if cur == false then cur = '' end
  if cur ~= ARGV[p + 1] then return -1 end
  p = p + 2
end
p = 7
for i = 1, nSets do
  redis.call('HSET', KEYS[1], ARGV[p], ARGV[p + 2])
  p = p + 3
end
for i = 1, nDels do
  redis.call('HDEL', KEYS[1], ARGV[p])
  p = p + 2
end
redis.call('HSET', KEYS[1], '` + structField + `', tostring((tonumber(sv) or 0) + 1))
redis.call('EXPIRE', KEYS[1], ARGV[1])
redis.call('DEL', KEYS[4])
local seq = redis.call('INCR', KEYS[2])
redis.call('EXPIRE', KEYS[2], ARGV[2])
redis.call('SADD', KEYS[3], ARGV[3])
return seq
`)

// patchElementScript replaces one element if it still holds the value that was read.
// KEYS: hash, seq, dirty set
// ARGV: hashTTL, seqTTL, dirtyMember, id, old, new
var patchElementScript = redis.NewScript(`
if redis.call('HGET', KEYS[1], ARGV[4]) ~= ARGV[5] then return -1 end
redis.call('HSET', KEYS[1], ARGV[4], ARGV[6])
redis.call('EXPIRE', KEYS[1], ARGV[1])
local seq = redis.call('INCR', KEYS[2])
redis.call('EXPIRE', KEYS[2], ARGV[2])
redis.call('SADD', KEYS[3], ARGV[3])
return seq
`)

// setIfAbsentScript fills the cache unless the page is already cached in either layout.
// KEYS: hash, legacy key
// ARGV: hashTTL, (id, json)*
var setIfAbsentScript = redis.NewScript(`
if redis.call('EXISTS', KEYS[1]) == 1 or redis.call('EXISTS', KEYS[2]) == 1 then return 0 end
redis.call('HSET', KEYS[1], '` + structField + `', '0')
for i = 2, #ARGV, 2 do
  redis.call('HSET', KEYS[1], ARGV[i], ARGV[i + 1])
end
redis.call('EXPIRE', KEYS[1], ARGV[1])
return 1
`)

// SetDocument stores the list of elements for a page in Redis, replacing any cached copy.
func (r *documentRepository) SetDocument(ctx context.Context, projectID string, pageID string, elements []domainEntity.Element) error {
	raw, err := flattenForStorage(elements)
	if err != nil {
		return err
	}
	fields := make(map[string]any, len(raw)+1)
	for id, v := range raw {
		fields[id] = v
	}
	// A fresh version, so a whole-document update that read the replaced page cannot apply.
	fields[structField] = strconv.FormatInt(time.Now().UnixNano(), 10)

	hashKey := r.hashKey(projectID, pageID)
	if _, err := r.client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Del(ctx, hashKey, r.docKey(projectID, pageID))
		pipe.HSet(ctx, hashKey, fields)
		pipe.Expire(ctx, hashKey, r.ttl)
		return nil
	}); err != nil {
		return fmt.Errorf("failed to set document in redis: %w", err)
	}
	return nil
}

// GetDocument retrieves the list of elements for a page from Redis
func (r *documentRepository) GetDocument(ctx context.Context, projectID string, pageID string) ([]domainEntity.Element, error) {
	doc, err := r.readDoc(ctx, projectID, pageID)
	if err != nil {
		return nil, err
	}
	if !doc.present {
		return nil, nil
	}
	return doc.tree()
}

// DeleteDocument removes the page elements from Redis
func (r *documentRepository) DeleteDocument(ctx context.Context, projectID string, pageID string) error {
	if err := r.client.Del(ctx, r.hashKey(projectID, pageID), r.docKey(projectID, pageID)).Err(); err != nil {
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

// SetDocumentIfAbsent stores the elements only if the page is not cached yet.
func (r *documentRepository) SetDocumentIfAbsent(ctx context.Context, projectID string, pageID string, elements []domainEntity.Element) error {
	raw, err := flattenForStorage(elements)
	if err != nil {
		return err
	}
	args := make([]any, 0, 1+2*len(raw))
	args = append(args, int64(r.ttl/time.Second))
	for id, v := range raw {
		args = append(args, id, v)
	}
	keys := []string{r.hashKey(projectID, pageID), r.docKey(projectID, pageID)}
	if err := setIfAbsentScript.Run(ctx, r.client, keys, args...).Err(); err != nil {
		return fmt.Errorf("failed to set value in redis: %w", err)
	}
	return nil
}

// GetDocumentWithSeq reads the document and its sequence number in one MULTI/EXEC so the pair
// is a consistent snapshot (mutations write both atomically).
func (r *documentRepository) GetDocumentWithSeq(ctx context.Context, projectID string, pageID string) ([]domainEntity.Element, int64, error) {
	hashVals, seq, err := r.readHashAndSeq(ctx, projectID, pageID)
	if err != nil {
		return nil, 0, err
	}
	doc, err := r.resolveStoredDoc(ctx, projectID, pageID, hashVals)
	if err != nil {
		return nil, 0, err
	}
	if !doc.present {
		return nil, seq, nil
	}
	elements, err := doc.tree()
	if err != nil {
		return nil, 0, err
	}
	return elements, seq, nil
}

func (r *documentRepository) readHashAndSeq(ctx context.Context, projectID, pageID string) (map[string]string, int64, error) {
	pipe := r.client.TxPipeline()
	hashCmd := pipe.HGetAll(ctx, r.hashKey(projectID, pageID))
	seqCmd := pipe.Get(ctx, r.seqKey(projectID, pageID))
	if _, err := pipe.Exec(ctx); err != nil && !errors.Is(err, redis.Nil) {
		return nil, 0, fmt.Errorf("failed to read document snapshot from redis: %w", err)
	}
	seq, err := seqCmd.Int64()
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, 0, fmt.Errorf("failed to read page seq: %w", err)
	}
	return hashCmd.Val(), seq, nil
}

func (r *documentRepository) resolveStoredDoc(ctx context.Context, projectID, pageID string, hashVals map[string]string) (*storedDoc, error) {
	if doc := docFromHash(hashVals); doc.present {
		return doc, nil
	}
	return r.readJSONBlob(ctx, projectID, pageID)
}

// UpdateDocument atomically applies mutate to the current document (loading it with load on a
// cache miss): it reads the page, runs mutate on the tree, and writes only the elements that
// changed. If another writer touched anything it read in between, the whole read-modify-write
// is retried against fresh data. The sequence number is incremented only by the successful
// write, so failed or aborted attempts never consume one.
func (r *documentRepository) UpdateDocument(ctx context.Context, projectID string, pageID string, load domainRepo.DocumentLoadFunc, mutate domainRepo.DocumentMutateFunc) (int64, error) {
	for attempt := 0; attempt < maxUpdateRetries; attempt++ {
		seq, applied, err := r.tryUpdateDocumentOnce(ctx, projectID, pageID, load, mutate)
		if err != nil {
			return 0, err
		}
		if applied {
			return seq, nil
		}
		if err := backoff(ctx, attempt); err != nil {
			return 0, err
		}
	}
	return 0, fmt.Errorf("document update for page %s aborted after %d conflicting attempts", pageID, maxUpdateRetries)
}

func (r *documentRepository) tryUpdateDocumentOnce(ctx context.Context, projectID, pageID string, load domainRepo.DocumentLoadFunc, mutate domainRepo.DocumentMutateFunc) (int64, bool, error) {
	doc, elements, err := r.loadDocElements(ctx, projectID, pageID, load)
	if err != nil {
		return 0, false, err
	}
	updated, err := mutate(elements)
	if err != nil {
		return 0, false, err
	}
	next, err := flattenForStorage(updated)
	if err != nil {
		return 0, false, err
	}
	return r.applyDiff(ctx, projectID, pageID, doc, next)
}

func (r *documentRepository) loadDocElements(ctx context.Context, projectID, pageID string, load domainRepo.DocumentLoadFunc) (*storedDoc, []domainEntity.Element, error) {
	doc, err := r.readDoc(ctx, projectID, pageID)
	if err != nil {
		return nil, nil, err
	}
	if doc.present {
		elements, err := doc.tree()
		return doc, elements, err
	}
	if load == nil {
		return doc, nil, nil
	}
	elements, err := load(ctx)
	return doc, elements, err
}

// PatchDocumentElement atomically applies mutate to one element. It reads and writes only that
// element, so edits to different elements of the same page never conflict. It falls back to a
// whole-document update when the page is not in the hash yet (cache miss or legacy layout).
func (r *documentRepository) PatchDocumentElement(ctx context.Context, projectID string, pageID string, elementID string, load domainRepo.DocumentLoadFunc, mutate domainRepo.ElementMutateFunc) (int64, error) {
	hashKey := r.hashKey(projectID, pageID)
	for attempt := 0; attempt < maxUpdateRetries; attempt++ {
		seq, applied, err := r.tryPatchElementOnce(ctx, hashKey, projectID, pageID, elementID, load, mutate)
		if err != nil {
			return 0, err
		}
		if applied {
			return seq, nil
		}
		if err := backoff(ctx, attempt); err != nil {
			return 0, err
		}
	}
	return 0, fmt.Errorf("element %s update aborted after %d conflicting attempts", elementID, maxUpdateRetries)
}

func (r *documentRepository) tryPatchElementOnce(ctx context.Context, hashKey, projectID, pageID, elementID string, load domainRepo.DocumentLoadFunc, mutate domainRepo.ElementMutateFunc) (int64, bool, error) {
	prev, err := r.client.HGet(ctx, hashKey, elementID).Result()
	if errors.Is(err, redis.Nil) {
		seq, err := r.handleMissingCachedElement(ctx, hashKey, projectID, pageID, elementID, load, mutate)
		return seq, err == nil, err
	}
	if err != nil {
		return 0, false, fmt.Errorf("failed to read element from redis: %w", err)
	}
	next, err := marshalMutatedElement(prev, mutate)
	if err != nil {
		return 0, false, err
	}
	res, err := patchElementScript.Run(ctx, r.client,
		[]string{hashKey, r.seqKey(projectID, pageID), dirtyPageKey},
		int64(r.ttl/time.Second), int64(seqTTL/time.Second), dirtyMember(projectID, pageID), elementID, prev, next,
	).Int64()
	if err != nil {
		return 0, false, fmt.Errorf("failed to patch element: %w", err)
	}
	return res, res >= 0, nil
}

func (r *documentRepository) handleMissingCachedElement(ctx context.Context, hashKey, projectID, pageID, elementID string, load domainRepo.DocumentLoadFunc, mutate domainRepo.ElementMutateFunc) (int64, error) {
	cached, err := r.client.HExists(ctx, hashKey, structField).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to check cached document: %w", err)
	}
	if cached {
		return 0, errElementNotFound
	}
	return r.patchViaDocument(ctx, projectID, pageID, elementID, load, mutate)
}

func marshalMutatedElement(prev string, mutate domainRepo.ElementMutateFunc) (string, error) {
	var el domainEntity.Element
	if err := json.Unmarshal([]byte(prev), &el); err != nil {
		return "", fmt.Errorf("failed to unmarshal element: %w", err)
	}
	if err := mutate(&el); err != nil {
		return "", err
	}
	el.Elements = nil
	next, err := json.Marshal(el)
	if err != nil {
		return "", fmt.Errorf("failed to marshal element: %w", err)
	}
	return string(next), nil
}

var errElementNotFound = domainErrors.New(domainErrors.CodeNotFound, "element not found in cache", nil)

func (r *documentRepository) patchViaDocument(ctx context.Context, projectID, pageID, elementID string, load domainRepo.DocumentLoadFunc, mutate domainRepo.ElementMutateFunc) (int64, error) {
	return r.UpdateDocument(ctx, projectID, pageID, load, func(elements []domainEntity.Element) ([]domainEntity.Element, error) {
		target := elementutils.FindByID(elements, elementID)
		if target == nil {
			return nil, errElementNotFound
		}
		if err := mutate(target); err != nil {
			return nil, err
		}
		return elements, nil
	})
}

// applyDiff writes the difference between the page as read and its next version. applied is false
// when another writer changed something that was read, in which case nothing was written.
func (r *documentRepository) applyDiff(ctx context.Context, projectID, pageID string, doc *storedDoc, next map[string]string) (seq int64, applied bool, err error) {
	prev, expectedSv := previousDocState(doc)
	sets, nSets, dels, nDels := diffDocumentEntries(prev, next)
	args := append([]any{
		int64(r.ttl / time.Second), int64(seqTTL / time.Second), dirtyMember(projectID, pageID), expectedSv, nSets, nDels,
	}, append(sets, dels...)...)

	res, err := applyDocumentScript.Run(ctx, r.client,
		[]string{r.hashKey(projectID, pageID), r.seqKey(projectID, pageID), dirtyPageKey, r.docKey(projectID, pageID)},
		args...,
	).Int64()
	if err != nil {
		return 0, false, fmt.Errorf("failed to update document: %w", err)
	}
	return res, res >= 0, nil
}

// A legacy page has no hash yet, so every element is written as new.
func previousDocState(doc *storedDoc) (map[string]string, string) {
	if doc.legacy {
		return nil, ""
	}
	return doc.raw, doc.sv
}

func diffDocumentEntries(prev, next map[string]string) (sets []any, nSets int, dels []any, nDels int) {
	for id, v := range next {
		if old, ok := prev[id]; !ok || old != v {
			sets = append(sets, id, old, v)
			nSets++
		}
	}
	for id, old := range prev {
		if _, ok := next[id]; !ok {
			dels = append(dels, id, old)
			nDels++
		}
	}
	return sets, nSets, dels, nDels
}

// readDoc reads a page from the hash, falling back to the legacy single-string layout.
func (r *documentRepository) readDoc(ctx context.Context, projectID, pageID string) (*storedDoc, error) {
	vals, err := r.client.HGetAll(ctx, r.hashKey(projectID, pageID)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get value from redis: %w", err)
	}
	if doc := docFromHash(vals); doc.present {
		return doc, nil
	}
	return r.readJSONBlob(ctx, projectID, pageID)
}

func docFromHash(vals map[string]string) *storedDoc {
	sv, ok := vals[structField]
	if !ok {
		return &storedDoc{}
	}
	delete(vals, structField)
	return &storedDoc{raw: vals, sv: sv, present: true}
}

func (r *documentRepository) readJSONBlob(ctx context.Context, projectID, pageID string) (*storedDoc, error) {
	val, err := r.client.Get(ctx, r.docKey(projectID, pageID)).Result()
	if errors.Is(err, redis.Nil) {
		return &storedDoc{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get value from redis: %w", err)
	}
	var tree []domainEntity.Element
	if err := json.Unmarshal([]byte(val), &tree); err != nil {
		return nil, fmt.Errorf("failed to unmarshal elements: %w", err)
	}
	raw, err := flattenForStorage(tree)
	if err != nil {
		return nil, err
	}
	return &storedDoc{raw: raw, present: true, legacy: true}, nil
}

// tree rebuilds the nested document. A cached page with no elements is an empty, non-nil slice;
// nil is reserved for a cache miss.
func (d *storedDoc) tree() ([]domainEntity.Element, error) {
	flat := make([]domainEntity.Element, 0, len(d.raw))
	for _, v := range d.raw {
		var el domainEntity.Element
		if err := json.Unmarshal([]byte(v), &el); err != nil {
			return nil, fmt.Errorf("failed to unmarshal element: %w", err)
		}
		flat = append(flat, el)
	}
	tree := elementutils.BuildTree(flat)
	if tree == nil {
		tree = []domainEntity.Element{}
	}
	return tree, nil
}

// flattenForStorage turns a tree into element id -> flat JSON. Nested elements take their
// parent's id as ParentID so the tree can be rebuilt from the flat form.
func flattenForStorage(tree []domainEntity.Element) (map[string]string, error) {
	out := make(map[string]string)
	var walk func(els []domainEntity.Element, parent *string) error
	walk = func(els []domainEntity.Element, parent *string) error {
		for _, el := range els {
			children := el.Elements
			el.Elements = nil
			if parent != nil {
				el.ParentID = parent
			}
			b, err := json.Marshal(el)
			if err != nil {
				return fmt.Errorf("failed to marshal element: %w", err)
			}
			out[el.ID] = string(b)
			id := el.ID
			if err := walk(children, &id); err != nil {
				return err
			}
		}
		return nil
	}
	if err := walk(tree, nil); err != nil {
		return nil, err
	}
	return out, nil
}

// backoff waits briefly before a retry; it grows with the attempt number and is jittered so
// competing writers do not retry in lockstep.
func backoff(ctx context.Context, attempt int) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(time.Duration(1+rand.Intn(5<<min(attempt, 4))) * time.Millisecond):
		return nil
	}
}

func dirtyMember(projectID, pageID string) string {
	return fmt.Sprintf("%s:%s", projectID, pageID)
}

func (r *documentRepository) seqKey(projectID string, pageID string) string {
	return r.docKey(projectID, pageID) + ":seq"
}

func (r *documentRepository) hashKey(projectID string, pageID string) string {
	return r.docKey(projectID, pageID) + ":els"
}

// docKey is the legacy single-string key; it also prefixes the hash and sequence keys.
func (r *documentRepository) docKey(projectID string, pageID string) string {
	return fmt.Sprintf("page:%s:%s", projectID, pageID)
}
