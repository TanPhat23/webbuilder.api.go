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
		rng := rand.New(rand.NewSource(seed))
		page := fmt.Sprintf("%s-%d", testPage, seed)
		next := 100 * int(seed)
		pick := func(els []entity.Element) string {
			flat := elementutils.FlattenElements(els)
			if len(flat) == 0 || rng.Intn(12) == 0 {
				return uid(999999) // unknown id: both stacks must fail the same way
			}
			return flat[rng.Intn(len(flat))].ID
		}

		for step := 0; step < 200; step++ {
			cur, err := ref.GetElementsByPageID(ctx, project, page)
			if err != nil {
				t.Fatal(err)
			}
			var op string
			var errRef, errGot error
			var seqRef, seqGot int64
			// Weighted so the tree keeps growing: removes would otherwise empty it and starve the rest.
			kind := rng.Intn(100)
			switch {
			case kind < 35: // create
				op = "create"
				next++
				parent := pick(cur)
				if rng.Intn(3) == 0 {
					parent = ""
				}
				pos := rng.Intn(4)
				mk := func() *entity.Element {
					el := &entity.Element{ID: uid(next), Type: "Text", PageID: page, Content: strPtr("new")}
					if parent != "" {
						el.ParentID = &parent
					}
					return el
				}
				var p *int
				if rng.Intn(2) == 0 {
					p = &pos
				}
				seqRef, errRef = ref.CreateElementAt(ctx, project, mk(), p)
				seqGot, errGot = got.CreateElementAt(ctx, project, mk(), p)
			case kind < 60: // patch
				op = "patch"
				id := pick(cur)
				props := map[string]any{"content": fmt.Sprintf("c%d", step)}
				styles := map[string]any{fmt.Sprintf("k%d", rng.Intn(4)): step, "gone": nil}
				seqRef, errRef = ref.PatchElement(ctx, project, page, id, props, styles, map[string]any{"s": map[string]any{"n": step}})
				seqGot, errGot = got.PatchElement(ctx, project, page, id, props, styles, map[string]any{"s": map[string]any{"n": step}})
			case kind < 82: // move
				op = "move"
				id, parent := pick(cur), pick(cur)
				if rng.Intn(3) == 0 {
					parent = ""
				}
				idx := rng.Intn(4)
				_, seqRef, errRef = ref.MoveElement(ctx, project, page, id, parent, idx)
				_, seqGot, errGot = got.MoveElement(ctx, project, page, id, parent, idx)
			case kind < 90: // remove
				op = "remove"
				id := pick(cur)
				seqRef, errRef = ref.RemoveElement(ctx, project, page, id)
				seqGot, errGot = got.RemoveElement(ctx, project, page, id)
			default: // update (full replace)
				op = "update"
				id := pick(cur)
				el := entity.Element{ID: id, Type: "Text", PageID: page, Content: strPtr(fmt.Sprintf("u%d", step))}
				if old := elementutils.FindByID(cur, id); old != nil {
					el.ParentID, el.OrderKey, el.Order = old.ParentID, old.OrderKey, old.Order
					if rng.Intn(2) == 0 {
						el.Elements = old.Elements
					}
				}
				e1, e2 := el, el
				seqRef, errRef = ref.UpdateElement(ctx, project, page, &e1)
				seqGot, errGot = got.UpdateElement(ctx, project, page, &e2)
			}

			if (errRef == nil) != (errGot == nil) {
				t.Fatalf("seed %d step %d %s: reference err=%v, redis err=%v", seed, step, op, errRef, errGot)
			}
			if errRef == nil {
				applied[op]++
			}
			if seqRef != seqGot {
				t.Fatalf("seed %d step %d %s: seq reference=%d redis=%d", seed, step, op, seqRef, seqGot)
			}
			a, _ := ref.GetElementsByPageID(ctx, project, page)
			b, err := got.GetElementsByPageID(ctx, project, page)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(normalized(t, a), normalized(t, b)) {
				ja, _ := json.MarshalIndent(normalized(t, a), "", " ")
				jb, _ := json.MarshalIndent(normalized(t, b), "", " ")
				t.Fatalf("seed %d step %d %s: documents differ\nreference: %s\nredis: %s", seed, step, op, ja, jb)
			}
		}
	}
	t.Logf("operations applied by kind: %v", applied)
	for _, op := range []string{"create", "patch", "move", "remove", "update"} {
		if applied[op] < 10 {
			t.Errorf("only %d %s operations succeeded; the random run does not exercise them enough", applied[op], op)
		}
	}
}

// Concurrent increments of one element and of different elements all land, and every successful
// operation takes exactly one sequence number.
func TestRedisDocumentRepo_ConcurrentPatchesAreNotLost(t *testing.T) {
	c, project := testRedis(t)
	ctx := context.Background()
	repo := repository.NewDocumentRepository(c)

	const elements, perElement = 6, 25
	tree := make([]entity.Element, elements)
	for i := range tree {
		tree[i] = entity.Element{ID: uid(i + 1), Type: "Text", PageID: testPage, OrderKey: fmt.Sprintf("a%d", i), Order: i}
	}
	if err := repo.SetDocument(ctx, project, testPage, tree); err != nil {
		t.Fatal(err)
	}

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
	// Structural updates interleaved with the patches.
	const creates = 10
	for k := 0; k < creates; k++ {
		wg.Add(1)
		go func(k int) {
			defer wg.Done()
			_, err := repo.UpdateDocument(ctx, project, testPage, nil, func(els []entity.Element) ([]entity.Element, error) {
				return append(els, entity.Element{ID: uid(1000 + k), Type: "Text", PageID: testPage, OrderKey: fmt.Sprintf("b%02d", k)}), nil
			})
			if err != nil {
				t.Errorf("update: %v", err)
			}
		}(k)
	}
	wg.Wait()

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
		if el.ID > uid(elements) { // a created element
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
	if _, s, _ := repo.GetDocumentWithSeq(ctx, project, testPage); s != 0 {
		t.Fatalf("a failed patch consumed sequence %d", s)
	}
}

func TestRedisDocumentRepo_EmptyDocumentIsACacheHit(t *testing.T) {
	c, project := testRedis(t)
	ctx := context.Background()
	repo := repository.NewDocumentRepository(c)

	if els, _ := repo.GetDocument(ctx, project, testPage); els != nil {
		t.Fatalf("a missing page must read as nil, got %v", els)
	}
	if err := repo.SetDocument(ctx, project, testPage, []entity.Element{}); err != nil {
		t.Fatal(err)
	}
	if els, _ := repo.GetDocument(ctx, project, testPage); els == nil || len(els) != 0 {
		t.Fatalf("an empty page must read as an empty non-nil slice, got %#v", els)
	}
	// SetDocumentIfAbsent must not touch a page that is already cached.
	if err := repo.SetDocumentIfAbsent(ctx, project, testPage, []entity.Element{{ID: uid(1), Type: "Text", PageID: testPage}}); err != nil {
		t.Fatal(err)
	}
	if els, _ := repo.GetDocument(ctx, project, testPage); len(els) != 0 {
		t.Fatalf("SetDocumentIfAbsent overwrote a cached page: %v", els)
	}
}

// Pages cached by the previous single-string layout stay readable and are migrated by the next
// write, without losing the sequence number.
func TestRedisDocumentRepo_ReadsAndMigratesLegacyLayout(t *testing.T) {
	c, project := testRedis(t)
	ctx := context.Background()
	repo := repository.NewDocumentRepository(c)

	legacy := []entity.Element{{
		ID: uid(1), Type: "Section", PageID: testPage, OrderKey: "a0",
		Elements: []entity.Element{{ID: uid(2), Type: "Text", PageID: testPage, OrderKey: "a0", ParentID: strPtr(uid(1))}},
	}}
	blob, _ := json.Marshal(legacy)
	key := "page:" + project + ":" + testPage
	c.Set(ctx, key, blob, 0)
	c.Set(ctx, key+":seq", 7, 0)

	els, seq, err := repo.GetDocumentWithSeq(ctx, project, testPage)
	if err != nil || seq != 7 || len(els) != 1 || len(els[0].Elements) != 1 {
		t.Fatalf("legacy read: seq=%d els=%v err=%v", seq, els, err)
	}
	if err := repo.SetDocumentIfAbsent(ctx, project, testPage, nil); err != nil {
		t.Fatal(err)
	}
	if els, _ := repo.GetDocument(ctx, project, testPage); len(els) != 1 {
		t.Fatalf("a cache fill shadowed the legacy page: %v", els)
	}

	seq, err = repo.PatchDocumentElement(ctx, project, testPage, uid(2), nil, func(el *entity.Element) error {
		el.Content = strPtr("patched")
		return nil
	})
	if err != nil || seq != 8 {
		t.Fatalf("patch of a legacy page: seq=%d err=%v", seq, err)
	}
	if n, _ := c.Exists(ctx, key).Result(); n != 0 {
		t.Fatal("legacy key was not removed after migration")
	}
	els, _ = repo.GetDocument(ctx, project, testPage)
	if len(els) != 1 || len(els[0].Elements) != 1 || els[0].Elements[0].Content == nil || *els[0].Elements[0].Content != "patched" {
		t.Fatalf("migrated document wrong: %+v", els)
	}
}
