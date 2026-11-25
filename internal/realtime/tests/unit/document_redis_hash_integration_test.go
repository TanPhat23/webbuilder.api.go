package unit

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"sync"
	"testing"

	"github.com/redis/go-redis/v9"

	"my-go-app/internal/realtime/domain/entity"
	domainRepo "my-go-app/internal/realtime/domain/repository"
	domainUsecase "my-go-app/internal/realtime/domain/usecase"
	"my-go-app/internal/realtime/repository"
	"my-go-app/internal/realtime/usecase"
	elementutils "my-go-app/internal/realtime/utils/element"
)

// These tests run the page document cache against a real Redis and are skipped unless
// WEBBUILDER_TEST_REDIS_ADDR is set (for example "localhost:6379").

func testRedis(t *testing.T) (*redis.Client, string) {
	t.Helper()
	addr := os.Getenv("WEBBUILDER_TEST_REDIS_ADDR")
	if addr == "" {
		t.Skip("set WEBBUILDER_TEST_REDIS_ADDR to run against a real Redis")
	}
	c := redis.NewClient(&redis.Options{Addr: addr})
	t.Cleanup(func() { _ = c.Close() })
	project := fmt.Sprintf("it-%s-%d", t.Name(), os.Getpid())
	t.Cleanup(func() {
		ctx := context.Background()
		iter := c.Scan(ctx, 0, "page:"+project+":*", 100).Iterator()
		for iter.Next(ctx) {
			c.Del(ctx, iter.Val())
		}
		c.SRem(ctx, "dirty_pages", project+":"+testPage)
	})
	return c, project
}

func uid(n int) string { return fmt.Sprintf("%08x-0000-4000-8000-%012x", n, n) }

// stack wires the real element/document use cases over the given document repository, with one
// element in the "database" for the cache-miss load.
func stack(docs domainRepo.DocumentRepository) domainUsecase.ElementUseCase {
	repo := &domainRepo.Repository{
		DocumentRepo: docs,
		ElementRepo: &memElementRepo{elements: []entity.Element{
			{ID: uid(1), Type: "Section", PageID: testPage, OrderKey: "a0", Order: 0},
		}},
	}
	return usecase.NewElementUseCase(repo, usecase.NewDocumentUseCase(repo), nil)
}

// normalized strips what the two layouts legitimately represent differently (a nil versus an
// empty children list) so documents can be compared structurally.
func normalized(t *testing.T, els []entity.Element) any {
	t.Helper()
	b, err := json.Marshal(els)
	if err != nil {
		t.Fatal(err)
	}
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatal(err)
	}
	var strip func(any) any
	strip = func(n any) any {
		switch x := n.(type) {
		case []any:
			for i := range x {
				x[i] = strip(x[i])
			}
		case map[string]any:
			if kids, ok := x["elements"].([]any); ok && len(kids) == 0 {
				delete(x, "elements")
			}
			if x["elements"] == nil {
				delete(x, "elements")
			}
			for k, c := range x {
				x[k] = strip(c)
			}
		}
		return n
	}
	return strip(v)
}

// The Redis hash layout must behave exactly like the in-memory reference for any sequence of
// operations: same errors, same sequence numbers, same document.
func TestRedisDocumentRepo_MatchesReferenceUnderRandomOperations(t *testing.T) {
	c, project := testRedis(t)
	ctx := context.Background()
	ref := stack(newMemDocRepo())
	got := stack(repository.NewDocumentRepository(c))

	applied := map[string]int{}
	for seed := int64(1); seed <= 5; seed++ {
		runRandomOpsSeed(t, ctx, ref, got, project, seed, applied)
	}
	t.Logf("operations applied by kind: %v", applied)
	for _, op := range []string{"create", "patch", "move", "remove", "update"} {
		if applied[op] < 10 {
			t.Errorf("only %d %s operations succeeded; the random run does not exercise them enough", applied[op], op)
		}
	}
}

func runRandomOpsSeed(t *testing.T, ctx context.Context, ref, got domainUsecase.ElementUseCase, project string, seed int64, applied map[string]int) {
	t.Helper()
	rng := rand.New(rand.NewSource(seed))
	page := fmt.Sprintf("%s-%d", testPage, seed)
	next := 100 * int(seed)
	for step := 0; step < 200; step++ {
		cur, err := ref.ListElementsByPage(ctx, project, page)
		if err != nil {
			t.Fatal(err)
		}
		op, nextID, seqRef, errRef, seqGot, errGot := executeRandomComparisonOp(ctx, rng, ref, got, project, page, cur, next, step)
		next = nextID
		assertRandomOpStepParity(t, ctx, ref, got, project, page, seed, step, op, seqRef, seqGot, errRef, errGot, applied)
	}
}

func pickRandomTargetID(rng *rand.Rand, els []entity.Element) string {
	flat := elementutils.Flatten(els)
	if len(flat) == 0 || rng.Intn(12) == 0 {
		return uid(999999)
	}
	return flat[rng.Intn(len(flat))].ID
}

func executeRandomComparisonOp(
	ctx context.Context,
	rng *rand.Rand,
	ref, got domainUsecase.ElementUseCase,
	project, page string,
	cur []entity.Element,
	next, step int,
) (op string, nextID int, seqRef int64, errRef error, seqGot int64, errGot error) {
	kind := rng.Intn(100)
	switch {
	case kind < 35:
		next++
		sr, er, sg, eg := executeRandomCreateOp(ctx, rng, ref, got, project, page, cur, next)
		return "create", next, sr, er, sg, eg
	case kind < 60:
		id := pickRandomTargetID(rng, cur)
		props := map[string]any{"content": fmt.Sprintf("c%d", step)}
		styles := map[string]any{fmt.Sprintf("k%d", rng.Intn(4)): step, "gone": nil}
		sr, er := ref.PatchElement(ctx, project, page, id, props, styles, map[string]any{"s": map[string]any{"n": step}})
		sg, eg := got.PatchElement(ctx, project, page, id, props, styles, map[string]any{"s": map[string]any{"n": step}})
		return "patch", next, sr, er, sg, eg
	case kind < 82:
		id, parent := pickRandomTargetID(rng, cur), pickRandomTargetID(rng, cur)
		if rng.Intn(3) == 0 {
			parent = ""
		}
		idx := rng.Intn(4)
		_, sr, er := ref.MoveElement(ctx, project, page, id, parent, idx)
		_, sg, eg := got.MoveElement(ctx, project, page, id, parent, idx)
		return "move", next, sr, er, sg, eg
	case kind < 90:
		id := pickRandomTargetID(rng, cur)
		sr, er := ref.DeleteElement(ctx, project, page, id)
		sg, eg := got.DeleteElement(ctx, project, page, id)
		return "remove", next, sr, er, sg, eg
	default:
		sr, er, sg, eg := executeRandomUpdateOp(ctx, rng, ref, got, project, page, cur, step)
		return "update", next, sr, er, sg, eg
	}
}

func executeRandomCreateOp(
	ctx context.Context,
	rng *rand.Rand,
	ref, got domainUsecase.ElementUseCase,
	project, page string,
	cur []entity.Element,
	idNum int,
) (int64, error, int64, error) {
	parent := pickRandomTargetID(rng, cur)
	if rng.Intn(3) == 0 {
		parent = ""
	}
	pos := rng.Intn(4)
	mk := func() *entity.Element {
		el := &entity.Element{ID: uid(idNum), Type: "Text", PageID: page, Content: strPtr("new")}
		if parent != "" {
			el.ParentID = &parent
		}
		return el
	}
	var p *int
	if rng.Intn(2) == 0 {
		p = &pos
	}
	sr, er := ref.CreateElementAt(ctx, project, mk(), p)
	sg, eg := got.CreateElementAt(ctx, project, mk(), p)
	return sr, er, sg, eg
}

func executeRandomUpdateOp(
	ctx context.Context,
	rng *rand.Rand,
	ref, got domainUsecase.ElementUseCase,
	project, page string,
	cur []entity.Element,
	step int,
) (int64, error, int64, error) {
	id := pickRandomTargetID(rng, cur)
	el := entity.Element{ID: id, Type: "Text", PageID: page, Content: strPtr(fmt.Sprintf("u%d", step))}
	if old := elementutils.FindByID(cur, id); old != nil {
		el.ParentID, el.OrderKey, el.Order = old.ParentID, old.OrderKey, old.Order
		if rng.Intn(2) == 0 {
			el.Elements = old.Elements
		}
	}
	e1, e2 := el, el
	sr, er := ref.UpdateElement(ctx, project, page, &e1)
	sg, eg := got.UpdateElement(ctx, project, page, &e2)
	return sr, er, sg, eg
}

func assertRandomOpDocsMatch(t *testing.T, ctx context.Context, ref, got domainUsecase.ElementUseCase, project, page string, seed int64, step int, op string) {
	t.Helper()
	a, err := ref.ListElementsByPage(ctx, project, page)
	if err != nil {
		t.Fatal(err)
	}
	b, err := got.ListElementsByPage(ctx, project, page)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(normalized(t, a), normalized(t, b)) {
		ja, jErrA := json.MarshalIndent(normalized(t, a), "", " ")
		if jErrA != nil {
			t.Fatal(jErrA)
		}
		jb, jErrB := json.MarshalIndent(normalized(t, b), "", " ")
		if jErrB != nil {
			t.Fatal(jErrB)
		}
		t.Fatalf("seed %d step %d %s: documents differ\nreference: %s\nredis: %s", seed, step, op, ja, jb)
	}
}

func assertRandomOpStepParity(
	t *testing.T,
	ctx context.Context,
	ref, got domainUsecase.ElementUseCase,
	project, page string,
	seed int64,
	step int,
	op string,
	seqRef, seqGot int64,
	errRef, errGot error,
	applied map[string]int,
) {
	t.Helper()
	if (errRef == nil) != (errGot == nil) {
		t.Fatalf("seed %d step %d %s: reference err=%v, redis err=%v", seed, step, op, errRef, errGot)
	}
	if errRef == nil {
		applied[op]++
	}
	if seqRef != seqGot {
		t.Fatalf("seed %d step %d %s: seq reference=%d redis=%d", seed, step, op, seqRef, seqGot)
	}
	assertRandomOpDocsMatch(t, ctx, ref, got, project, page, seed, step, op)
}

func seedConcurrentElementsTree(t *testing.T, ctx context.Context, repo domainRepo.DocumentRepository, project string, count int) {
	t.Helper()
	tree := make([]entity.Element, count)
	for i := range tree {
		tree[i] = entity.Element{ID: uid(i + 1), Type: "Text", PageID: testPage, OrderKey: fmt.Sprintf("a%d", i), Order: i}
	}
	if err := repo.SetDocument(ctx, project, testPage, tree); err != nil {
		t.Fatal(err)
	}
}

func runConcurrentPatchesAndCreates(t *testing.T, ctx context.Context, repo domainRepo.DocumentRepository, project string, elements, perElement, creates int) {
	t.Helper()
	incr := func(el *entity.Element) error {
		n, _ := el.Settings["n"].(float64)
		el.Settings = map[string]any{"n": n + 1}
		return nil
	}
	var wg sync.WaitGroup
	for i := 0; i < elements; i++ {
		for j := 0; j < perElement; j++ {
			wg.Add(1)
			go func(id string) {
				defer wg.Done()
				if _, err := repo.PatchDocumentElement(ctx, project, testPage, id, nil, incr); err != nil {
					t.Errorf("patch: %v", err)
				}
			}(uid(i + 1))
		}
	}
	for k := 0; k < creates; k++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, err := repo.UpdateDocument(ctx, project, testPage, nil, func(els []entity.Element) ([]entity.Element, error) {
				return append(els, entity.Element{ID: uid(1000 + idx), Type: "Text", PageID: testPage, OrderKey: fmt.Sprintf("b%02d", idx)}), nil
			})
			if err != nil {
				t.Errorf("update: %v", err)
			}
		}(k)
	}
	wg.Wait()
}

// Concurrent increments of one element and of different elements all land, and every successful
// operation takes exactly one sequence number.
func TestRedisDocumentRepo_ConcurrentPatchesAreNotLost(t *testing.T) {
	c, project := testRedis(t)
	ctx := context.Background()
	repo := repository.NewDocumentRepository(c)

	const elements, perElement, creates = 6, 25, 10
	seedConcurrentElementsTree(t, ctx, repo, project, elements)
	runConcurrentPatchesAndCreates(t, ctx, repo, project, elements, perElement, creates)

	els, seq, err := repo.GetDocumentWithSeq(ctx, project, testPage)
	if err != nil {
		t.Fatal(err)
	}
	if want := int64(elements*perElement + creates); seq != want {
		t.Fatalf("seq = %d, want %d", seq, want)
	}
	if len(els) != elements+creates {
		t.Fatalf("elements = %d, want %d", len(els), elements+creates)
	}
	for _, el := range els {
		if el.ID > uid(elements) {
			continue
		}
		if n, _ := el.Settings["n"].(float64); int(n) != perElement {
			t.Errorf("element %s counted %v, want %d (lost updates)", el.ID, n, perElement)
		}
	}
}

func TestRedisDocumentRepo_PatchMissingElementIsNotFound(t *testing.T) {
	c, project := testRedis(t)
	ctx := context.Background()
	repo := repository.NewDocumentRepository(c)
	if err := repo.SetDocument(ctx, project, testPage, []entity.Element{{ID: uid(1), Type: "Text", PageID: testPage}}); err != nil {
		t.Fatal(err)
	}
	seq, err := repo.PatchDocumentElement(ctx, project, testPage, uid(2), nil, func(*entity.Element) error { return nil })
	if err == nil || seq != 0 {
		t.Fatalf("want a not-found error and no sequence, got seq=%d err=%v", seq, err)
	}
	els, s, seqErr := repo.GetDocumentWithSeq(ctx, project, testPage)
	if seqErr != nil || s != 0 || len(els) != 1 {
		t.Fatalf("failed patch mutated doc/seq: els=%v seq=%d err=%v", els, s, seqErr)
	}
}

func mustGetDoc(t *testing.T, ctx context.Context, repo domainRepo.DocumentRepository, project string) []entity.Element {
	t.Helper()
	els, err := repo.GetDocument(ctx, project, testPage)
	if err != nil {
		t.Fatalf("GetDocument: %v", err)
	}
	return els
}

func TestRedisDocumentRepo_EmptyDocumentIsACacheHit(t *testing.T) {
	c, project := testRedis(t)
	ctx := context.Background()
	repo := repository.NewDocumentRepository(c)

	if els := mustGetDoc(t, ctx, repo, project); els != nil {
		t.Fatalf("missing page must read as nil: els=%v", els)
	}
	if err := repo.SetDocument(ctx, project, testPage, []entity.Element{}); err != nil {
		t.Fatal(err)
	}
	if els := mustGetDoc(t, ctx, repo, project); els == nil || len(els) != 0 {
		t.Fatalf("an empty page must read as empty non-nil slice, got %#v", els)
	}
	if err := repo.SetDocumentIfAbsent(ctx, project, testPage, []entity.Element{{ID: uid(1), Type: "Text", PageID: testPage}}); err != nil {
		t.Fatal(err)
	}
	if els := mustGetDoc(t, ctx, repo, project); len(els) != 0 {
		t.Fatalf("SetDocumentIfAbsent overwrote a cached page: %v", els)
	}
}

func seedLegacyCachedPage(t *testing.T, ctx context.Context, c *redis.Client, project string) string {
	t.Helper()
	legacy := []entity.Element{{
		ID: uid(1), Type: "Section", PageID: testPage, OrderKey: "a0",
		Elements: []entity.Element{{ID: uid(2), Type: "Text", PageID: testPage, OrderKey: "a0", ParentID: strPtr(uid(1))}},
	}}
	blob, err := json.Marshal(legacy)
	if err != nil {
		t.Fatal(err)
	}
	key := "page:" + project + ":" + testPage
	if err := c.Set(ctx, key, blob, 0).Err(); err != nil {
		t.Fatal(err)
	}
	if err := c.Set(ctx, key+":seq", 7, 0).Err(); err != nil {
		t.Fatal(err)
	}
	return key
}

func assertLegacyDocReadable(t *testing.T, ctx context.Context, repo domainRepo.DocumentRepository, project string) {
	t.Helper()
	els, seq, err := repo.GetDocumentWithSeq(ctx, project, testPage)
	if err != nil || seq != 7 {
		t.Fatalf("legacy read: seq=%d els=%v err=%v", seq, els, err)
	}
	if len(els) != 1 || len(els[0].Elements) != 1 {
		t.Fatalf("legacy elements unexpected: %+v", els)
	}
	if err := repo.SetDocumentIfAbsent(ctx, project, testPage, nil); err != nil {
		t.Fatal(err)
	}
	if els := mustGetDoc(t, ctx, repo, project); len(els) != 1 {
		t.Fatalf("a cache fill shadowed the legacy page: %v", els)
	}
}

// Pages cached by the previous single-string layout stay readable and are migrated by the next
// write, without losing the sequence number.
func TestRedisDocumentRepo_ReadsAndMigratesLegacyLayout(t *testing.T) {
	c, project := testRedis(t)
	ctx := context.Background()
	repo := repository.NewDocumentRepository(c)
	key := seedLegacyCachedPage(t, ctx, c, project)

	assertLegacyDocReadable(t, ctx, repo, project)

	seq, err := repo.PatchDocumentElement(ctx, project, testPage, uid(2), nil, func(el *entity.Element) error {
		el.Content = strPtr("patched")
		return nil
	})
	if err != nil || seq != 8 {
		t.Fatalf("patch of a legacy page: seq=%d err=%v", seq, err)
	}
	if n, err := c.Exists(ctx, key).Result(); err != nil || n != 0 {
		t.Fatalf("legacy key not removed after migration: n=%d err=%v", n, err)
	}
	els := mustGetDoc(t, ctx, repo, project)
	if len(els) != 1 || len(els[0].Elements) != 1 {
		t.Fatalf("migrated document wrong: %+v", els)
	}
	if els[0].Elements[0].Content == nil || *els[0].Elements[0].Content != "patched" {
		t.Fatalf("migrated content wrong: %+v", els[0].Elements[0])
	}
}
