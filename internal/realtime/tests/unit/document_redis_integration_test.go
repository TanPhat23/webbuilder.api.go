package unit

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/redis/go-redis/v9"

	"my-go-app/internal/realtime/domain/entity"
	"my-go-app/internal/realtime/repository"
)

// Cross-instance optimistic concurrency check against a real Redis. Skipped unless
// WEBBUILDER_TEST_REDIS_ADDR (e.g. "localhost:6379") is set, because no Redis fake is a
// dependency of this module. It uses two independent clients to stand in for two server instances.
func TestRedisDocumentRepo_UpdateDocumentIsAtomicAcrossInstances(t *testing.T) {
	addr := os.Getenv("WEBBUILDER_TEST_REDIS_ADDR")
	if addr == "" {
		t.Skip("set WEBBUILDER_TEST_REDIS_ADDR to run against a real Redis")
	}
	ctx := context.Background()
	c1 := redis.NewClient(&redis.Options{Addr: addr})
	c2 := redis.NewClient(&redis.Options{Addr: addr})
	defer c1.Close()
	defer c2.Close()

	project, page := "it-proj", fmt.Sprintf("it-page-%d", os.Getpid())
	defer c1.Del(ctx, "page:"+project+":"+page, "page:"+project+":"+page+":seq")

	r1 := repository.NewDocumentRepository(c1)
	r2 := repository.NewDocumentRepository(c2)

	const perInstance = 40
	var wg sync.WaitGroup
	for _, r := range []int{1, 2} {
		repo := r1
		if r == 2 {
			repo = r2
		}
		for i := 0; i < perInstance; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				id := fmt.Sprintf("e-%d-%d", r, i)
				if _, err := repo.UpdateDocument(ctx, project, page, nil, func(els []entity.Element) ([]entity.Element, error) {
					return append(els, entity.Element{ID: id}), nil
				}); err != nil {
					t.Errorf("UpdateDocument: %v", err)
				}
			}()
		}
	}
	wg.Wait()

	els, seq, err := r1.GetDocumentWithSeq(ctx, project, page)
	if err != nil {
		t.Fatal(err)
	}
	if len(els) != 2*perInstance || seq != 2*perInstance {
		t.Fatalf("lost updates: elements=%d seq=%d want %d", len(els), seq, 2*perInstance)
	}
}
