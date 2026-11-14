package unit

import (
	"context"
	"encoding/json"
	"sync"

	"my-go-app/internal/realtime/domain/entity"
	domainRepo "my-go-app/internal/realtime/domain/repository"
)

// memDocRepo is an in-memory DocumentRepository. UpdateDocument runs the whole
// read-modify-write under one lock and bumps seq only on success, which is the contract the
// Redis implementation provides with WATCH/MULTI/EXEC. Documents are stored as JSON so callers
// never share memory with the "database".
type memDocRepo struct {
	mu    sync.Mutex
	docs  map[string][]byte
	seqs  map[string]int64
	dirty map[string]int
}

func newMemDocRepo() *memDocRepo {
	return &memDocRepo{docs: map[string][]byte{}, seqs: map[string]int64{}, dirty: map[string]int{}}
}

func memKey(projectID, pageID string) string { return projectID + ":" + pageID }

func (m *memDocRepo) decode(key string) ([]entity.Element, bool) {
	raw, ok := m.docs[key]
	if !ok {
		return nil, false
	}
	var els []entity.Element
	_ = json.Unmarshal(raw, &els)
	return els, true
}

func (m *memDocRepo) SetDocument(ctx context.Context, projectID, pageID string, elements []entity.Element) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	raw, _ := json.Marshal(elements)
	m.docs[memKey(projectID, pageID)] = raw
	return nil
}

func (m *memDocRepo) SetDocumentIfAbsent(ctx context.Context, projectID, pageID string, elements []entity.Element) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	k := memKey(projectID, pageID)
	if _, ok := m.docs[k]; !ok {
		raw, _ := json.Marshal(elements)
		m.docs[k] = raw
	}
	return nil
}

func (m *memDocRepo) GetDocument(ctx context.Context, projectID, pageID string) ([]entity.Element, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	els, _ := m.decode(memKey(projectID, pageID))
	return els, nil
}

func (m *memDocRepo) GetDocumentWithSeq(ctx context.Context, projectID, pageID string) ([]entity.Element, int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	k := memKey(projectID, pageID)
	els, _ := m.decode(k)
	return els, m.seqs[k], nil
}

func (m *memDocRepo) DeleteDocument(ctx context.Context, projectID, pageID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.docs, memKey(projectID, pageID))
	return nil
}

func (m *memDocRepo) MarkPageDirty(ctx context.Context, projectID, pageID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.dirty[memKey(projectID, pageID)]++
	return nil
}

func (m *memDocRepo) PopDirtyPages(ctx context.Context, limit int) ([]entity.DirtyPage, error) {
	return nil, nil
}

func (m *memDocRepo) UpdateDocument(ctx context.Context, projectID, pageID string, load domainRepo.DocumentLoadFunc, mutate domainRepo.DocumentMutateFunc) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	k := memKey(projectID, pageID)

	els, ok := m.decode(k)
	if !ok && load != nil {
		var err error
		if els, err = load(ctx); err != nil {
			return 0, err
		}
	}
	updated, err := mutate(els)
	if err != nil {
		return 0, err
	}
	raw, _ := json.Marshal(updated)
	m.docs[k] = raw
	m.seqs[k]++
	m.dirty[k]++
	return m.seqs[k], nil
}

func (m *memDocRepo) seq(projectID, pageID string) int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.seqs[memKey(projectID, pageID)]
}

func (m *memDocRepo) snapshot(projectID, pageID string) []entity.Element {
	m.mu.Lock()
	defer m.mu.Unlock()
	els, _ := m.decode(memKey(projectID, pageID))
	return els
}

// memElementRepo only serves the cache-miss load path.
type memElementRepo struct {
	domainRepo.ElementRepository
	elements []entity.Element
}

func (r *memElementRepo) GetElementsByPageID(ctx context.Context, projectID, pageID string) ([]entity.Element, error) {
	return r.elements, nil
}

func (r *memElementRepo) GetElementsByProjectID(ctx context.Context, projectID string) ([]entity.Element, error) {
	return r.elements, nil
}
