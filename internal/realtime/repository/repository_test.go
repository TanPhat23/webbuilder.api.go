package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	domainEntity "my-go-app/internal/realtime/domain/entity"
	domainErrors "my-go-app/internal/realtime/domain/errors"
)

type rule struct {
	contains  string
	cols      []string
	rows      [][]driver.Value
	affected  int64
	isExec    bool
	err       error
	remaining int
}

type scriptDB struct {
	mu    sync.Mutex
	rules []*rule
}

func (s *scriptDB) on(contains string, cols []string, rows ...[]driver.Value) *rule {
	r := &rule{contains: contains, cols: cols, rows: rows, remaining: -1}
	s.mu.Lock()
	s.rules = append(s.rules, r)
	s.mu.Unlock()
	return r
}

func (s *scriptDB) onExec(contains string, affected int64) *rule {
	r := &rule{contains: contains, affected: affected, isExec: true, remaining: -1}
	s.mu.Lock()
	s.rules = append(s.rules, r)
	s.mu.Unlock()
	return r
}

func (s *scriptDB) match(query string, exec bool) *rule {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := len(s.rules) - 1; i >= 0; i-- {
		r := s.rules[i]
		if r.remaining == 0 || r.isExec != exec || !strings.Contains(query, r.contains) {
			continue
		}
		if r.remaining > 0 {
			r.remaining--
		}
		return r
	}
	return nil
}

type sConn struct{ db *scriptDB }
type sDriver struct{}
type sConnector struct{ db *scriptDB }
type sTx struct{}
type sRows struct {
	cols []string
	rows [][]driver.Value
	pos  int
}

func (c sConnector) Connect(context.Context) (driver.Conn, error) { return &sConn{db: c.db}, nil }
func (sConnector) Driver() driver.Driver                         { return sDriver{} }
func (sDriver) Open(string) (driver.Conn, error)                 { return nil, errors.New("unused") }
func (*sConn) Prepare(string) (driver.Stmt, error)               { return nil, driver.ErrSkip }
func (*sConn) Close() error                                      { return nil }
func (*sConn) Begin() (driver.Tx, error)                         { return &sTx{}, nil }
func (*sConn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	return &sTx{}, nil
}
func (*sTx) Commit() error   { return nil }
func (*sTx) Rollback() error { return nil }
func (r *sRows) Columns() []string {
	return r.cols
}
func (*sRows) Close() error { return nil }
func (r *sRows) Next(dest []driver.Value) error {
	if r.pos >= len(r.rows) {
		return io.EOF
	}
	copy(dest, r.rows[r.pos])
	r.pos++
	return nil
}

func (c *sConn) QueryContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	r := c.db.match(query, false)
	if r == nil {
		return &sRows{}, nil
	}
	if r.err != nil {
		return nil, r.err
	}
	return &sRows{cols: r.cols, rows: r.rows}, nil
}

func (c *sConn) ExecContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Result, error) {
	r := c.db.match(query, true)
	if r == nil {
		return driver.RowsAffected(1), nil
	}
	if r.err != nil {
		return nil, r.err
	}
	return driver.RowsAffected(r.affected), nil
}

func newScriptGorm(t *testing.T) (*gorm.DB, *scriptDB) {
	t.Helper()
	s := &scriptDB{}
	sqlDB := sql.OpenDB(sConnector{db: s})
	t.Cleanup(func() { _ = sqlDB.Close() })
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
		DisableAutomaticPing: true,
		Logger:               gormlogger.Discard,
	})
	if err != nil {
		t.Fatal(err)
	}
	return db, s
}

func TestDocumentRepositoryRedisAndLuaLifecycle(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	ctx := context.Background()
	docRepo := NewDocumentRepository(rdb)

	// Cache miss
	if got, err := docRepo.GetDocument(ctx, "p1", "pg1"); err != nil || got != nil {
		t.Fatalf("cache miss got=%v err=%v", got, err)
	}

	els := []domainEntity.Element{
		{
			ID:     "e1",
			Type:   "Frame",
			PageID: "pg1",
			Elements: []domainEntity.Element{
				{ID: "e2", Type: "Text", PageID: "pg1"},
			},
		},
	}

	if err := docRepo.SetDocumentIfAbsent(ctx, "p1", "pg1", els); err != nil {
		t.Fatalf("SetDocumentIfAbsent: %v", err)
	}
	// Second SetDocumentIfAbsent should keep cached value untouched
	if err := docRepo.SetDocumentIfAbsent(ctx, "p1", "pg1", []domainEntity.Element{{ID: "other", Type: "Text", PageID: "pg1"}}); err != nil {
		t.Fatalf("SetDocumentIfAbsent second: %v", err)
	}

	got, seq, err := docRepo.GetDocumentWithSeq(ctx, "p1", "pg1")
	if err != nil || seq != 0 || len(got) != 1 || len(got[0].Elements) != 1 {
		t.Fatalf("GetDocumentWithSeq got=%+v seq=%d err=%v", got, seq, err)
	}

	// UpdateDocument whole tree
	seq, err = docRepo.UpdateDocument(ctx, "p1", "pg1", nil, func(cur []domainEntity.Element) ([]domainEntity.Element, error) {
		return append(cur, domainEntity.Element{ID: "e3", Type: "Text", PageID: "pg1"}), nil
	})
	if err != nil || seq != 1 {
		t.Fatalf("UpdateDocument seq=%d err=%v", seq, err)
	}

	// PatchDocumentElement on single element
	seq, err = docRepo.PatchDocumentElement(ctx, "p1", "pg1", "e2", nil, func(target *domainEntity.Element) error {
		txt := "Patched"
		target.Content = &txt
		return nil
	})
	if err != nil || seq != 2 {
		t.Fatalf("PatchDocumentElement seq=%d err=%v", seq, err)
	}

	// Missing element patch when page cached returns NOT_FOUND
	if _, err := docRepo.PatchDocumentElement(ctx, "p1", "pg1", "missing", nil, func(*domainEntity.Element) error { return nil }); !domainErrors.IsCode(err, domainErrors.CodeNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}

	// Dirty pages set + pop
	if err := docRepo.MarkPageDirty(ctx, "p2", "pg2"); err != nil {
		t.Fatalf("MarkPageDirty: %v", err)
	}
	dirty, err := docRepo.PopDirtyPages(ctx, 10)
	if err != nil || len(dirty) < 1 {
		t.Fatalf("PopDirtyPages dirty=%v err=%v", dirty, err)
	}

	// Overwrite with SetDocument, verify, then delete
	if err := docRepo.SetDocument(ctx, "p1", "pg1", []domainEntity.Element{{ID: "e9", Type: "Text", PageID: "pg1"}}); err != nil {
		t.Fatalf("SetDocument: %v", err)
	}
	if err := docRepo.DeleteDocument(ctx, "p1", "pg1"); err != nil {
		t.Fatalf("DeleteDocument: %v", err)
	}

	// Legacy single-key JSON fallback + auto-migration
	legacyJSON := `[{"id":"leg-1","type":"Text","pageId":"pg-leg"}]`
	_ = mr.Set("page:p1:pg-leg", legacyJSON)
	seq, err = docRepo.PatchDocumentElement(ctx, "p1", "pg-leg", "leg-1", nil, func(target *domainEntity.Element) error {
		txt := "migrated"
		target.Content = &txt
		return nil
	})
	if err != nil || seq != 1 {
		t.Fatalf("legacy migration seq=%d err=%v", seq, err)
	}
}

func TestGormRepositoriesAndLogicRepository(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	db, s := newScriptGorm(t)
	repos := NewRepository(db, rdb)
	ctx := context.Background()

	el := &domainEntity.Element{ID: "el-1", Type: "Text", PageID: "pg-1"}
	if err := repos.ElementRepo.CreateElement(ctx, "proj-1", "pg-1", el); err != nil {
		t.Fatalf("CreateElement: %v", err)
	}
	if err := repos.ElementRepo.CreateElements(ctx, "proj-1", "pg-1", []domainEntity.Element{*el}); err != nil {
		t.Fatalf("CreateElements: %v", err)
	}
	s.onExec(`UPDATE "Element"`, 1)
	if err := repos.ElementRepo.UpdateElement(ctx, "proj-1", "pg-1", el); err != nil {
		t.Fatalf("UpdateElement: %v", err)
	}
	if err := repos.ElementRepo.UpdateElements(ctx, "proj-1", "pg-1", []domainEntity.Element{*el}); err != nil {
		t.Fatalf("UpdateElements: %v", err)
	}
	if err := repos.ElementRepo.UpsertElements(ctx, "proj-1", "pg-1", []domainEntity.Element{*el}); err != nil {
		t.Fatalf("UpsertElements: %v", err)
	}
	if err := repos.ElementRepo.DeleteElement(ctx, "proj-1", "pg-1", "el-1"); err != nil {
		t.Fatalf("DeleteElement: %v", err)
	}
	s.on(`FROM "Element"`, []string{"Id", "Type", "PageId"}, []driver.Value{"el-1", "Text", "pg-1"})
	if got, err := repos.ElementRepo.GetElementByID(ctx, "proj-1", "pg-1", "el-1"); err != nil || got == nil {
		t.Fatalf("GetElementByID got=%v err=%v", got, err)
	}
	if list, err := repos.ElementRepo.GetElementsByPageID(ctx, "proj-1", "pg-1"); err != nil || len(list) == 0 {
		t.Fatalf("GetElementsByPageID list=%v err=%v", list, err)
	}
	if list, err := repos.ElementRepo.GetElementsByProjectID(ctx, "proj-1"); err != nil || len(list) == 0 {
		t.Fatalf("GetElementsByProjectID list=%v err=%v", list, err)
	}

	// Snapshot repository
	if err := repos.SnapshotRepo.SaveSnapshot(ctx, "proj-1", "pg-1", "snap", domainEntity.SnapshotTypes.Version, []domainEntity.Element{*el}); err != nil {
		t.Fatalf("SaveSnapshot: %v", err)
	}
	s.on(`FROM "Snapshot"`, []string{"Id", "Name"}, []driver.Value{"s1", "snap"})
	if snap, err := repos.SnapshotRepo.GetSnapshot(ctx, "proj-1", "pg-1", domainEntity.SnapshotTypes.Version); err != nil || snap == nil {
		t.Fatalf("GetSnapshot snap=%v err=%v", snap, err)
	}
	if err := repos.SnapshotRepo.DeleteSnapshot(ctx, "pg-1", "proj-1", domainEntity.SnapshotTypes.Version); err != nil {
		t.Fatalf("DeleteSnapshot: %v", err)
	}

	// Workflow & ElementWorkflow repositories
	wf := &domainEntity.EventWorkflow{ID: "wf-1", Name: "Click"}
	if err := repos.WorkflowRepo.CreateWorkflow(ctx, "proj-1", wf); err != nil {
		t.Fatalf("CreateWorkflow: %v", err)
	}
	if err := repos.WorkflowRepo.UpdateWorkflow(ctx, "proj-1", wf); err != nil {
		t.Fatalf("UpdateWorkflow: %v", err)
	}
	s.on(`FROM "EventWorkflow"`, []string{"Id", "Name"}, []driver.Value{"wf-1", "Click"})
	if got, err := repos.WorkflowRepo.GetWorkflowByID(ctx, "proj-1", "wf-1"); err != nil || got == nil {
		t.Fatalf("GetWorkflowByID got=%v err=%v", got, err)
	}
	if list, err := repos.WorkflowRepo.GetWorkflowsByProjectID(ctx, "proj-1"); err != nil || len(list) == 0 {
		t.Fatalf("GetWorkflowsByProjectID err=%v", err)
	}
	if err := repos.WorkflowRepo.DeleteWorkflow(ctx, "proj-1", "wf-1"); err != nil {
		t.Fatalf("DeleteWorkflow: %v", err)
	}

	s.on(`SELECT count`, []string{"count"}, []driver.Value{int64(1)})
	ewf := &domainEntity.ElementEventWorkflow{ID: "ewf-1", ElementID: "el-1", WorkflowID: "wf-1", EventName: "click"}
	if err := repos.ElementWorkflowRepo.CreateElementWorkflow(ctx, "proj-1", ewf); err != nil {
		t.Fatalf("CreateElementWorkflow: %v", err)
	}
	if err := repos.ElementWorkflowRepo.UpdateElementWorkflow(ctx, "proj-1", ewf); err != nil {
		t.Fatalf("UpdateElementWorkflow: %v", err)
	}
	s.on(`FROM "ElementEventWorkflow"`, []string{"Id", "ElementId"}, []driver.Value{"ewf-1", "el-1"})
	if got, err := repos.ElementWorkflowRepo.GetElementWorkflowByElementID(ctx, "proj-1", "el-1"); err != nil || got == nil {
		t.Fatalf("GetElementWorkflowByElementID got=%v err=%v", got, err)
	}
	if list, err := repos.ElementWorkflowRepo.GetElementWorkflowsByProjectID(ctx, "proj-1"); err != nil || len(list) == 0 {
		t.Fatalf("GetElementWorkflowsByProjectID err=%v", err)
	}
	if list, err := repos.ElementWorkflowRepo.GetElementWorkflowsByPageID(ctx, "proj-1", "pg-1"); err != nil || len(list) == 0 {
		t.Fatalf("GetElementWorkflowsByPageID err=%v", err)
	}
	if err := repos.ElementWorkflowRepo.DeleteElementWorkflow(ctx, "proj-1", "el-1"); err != nil {
		t.Fatalf("DeleteElementWorkflow: %v", err)
	}

	// ProjectRepository
	proj := &domainEntity.Project{ID: "proj-1", Name: "Demo", OwnerID: "u1"}
	if err := repos.ProjectRepo.CreateProject(proj); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	s.on(`FROM "Project"`, []string{"id", "name"}, []driver.Value{"proj-1", "Demo"})
	if gotProj, err := repos.ProjectRepo.GetProjectByID("proj-1"); err != nil || gotProj == nil {
		t.Fatalf("GetProjectByID got=%v err=%v", gotProj, err)
	}
	if err := repos.ProjectRepo.UpdateProject(proj); err != nil {
		t.Fatalf("UpdateProject: %v", err)
	}
	if err := repos.ProjectRepo.DeleteProject("proj-1"); err != nil {
		t.Fatalf("DeleteProject: %v", err)
	}

	// LogicRepository Access / Apply / Move / Get validation & happy paths
	if r, w, err := repos.LogicRepo.Access(ctx, "", "u1"); err != nil || r || w {
		t.Fatalf("empty Access r=%v w=%v err=%v", r, w, err)
	}
	s.on(`SELECT ("OwnerId"`, []string{"owner", "read", "write"}, []driver.Value{true, false, false})
	if r, w, err := repos.LogicRepo.Access(ctx, "proj-1", "u1"); err != nil || !r || !w {
		t.Fatalf("owner Access r=%v w=%v err=%v", r, w, err)
	}

	if _, _, err := repos.LogicRepo.Move(ctx, "proj-1", "pg-1", "u1", "bad", "x", "project"); !errors.Is(err, domainErrors.ErrValidation) {
		t.Fatalf("expected validation error on bad kind, got %v", err)
	}

	s.on(`SELECT "Logic", "OwnerId" FROM "Project"`, []string{"Logic", "OwnerId"}, []driver.Value{[]byte(`{"v":1,"variables":{},"flows":{}}`), "u1"})
	s.on(`SELECT "Id", "Logic" FROM "Page"`, []string{"Id", "Logic"}, []driver.Value{"pg-1", []byte(`{"v":1,"variables":{},"flows":{}}`)})
	s.on(`UPDATE "Project"`, []string{"LogicRevision"}, []driver.Value{int64(2)})
	s.on(`UPDATE "Page"`, []string{"LogicRevision"}, []driver.Value{int64(3)})

	revDoc, err := repos.LogicRepo.Apply(ctx, "proj-1", "pg-1", "u1", map[string]any{
		"variables": map[string]any{
			"count": map[string]any{"type": "number", "default": float64(1)},
		},
	})
	if err != nil || revDoc.Revision != 3 {
		t.Fatalf("Logic.Apply rev=%d err=%v", revDoc.Revision, err)
	}

	s.on(`SELECT "Logic", "LogicRevision", "OwnerId"`, []string{"Logic", "LogicRevision", "OwnerId"}, []driver.Value{[]byte(`null`), int64(2), "u1"})
	s.on(`SELECT "Logic", "LogicRevision" FROM "Page"`, []string{"Logic", "LogicRevision"}, []driver.Value{[]byte(`null`), int64(3)})
	pDoc, pgDoc, err := repos.LogicRepo.Get(ctx, "proj-1", "pg-1", "u1")
	if err != nil || pDoc.Revision != 2 || pgDoc.Revision != 3 {
		t.Fatalf("Logic.Get err=%v p=%d pg=%d", err, pDoc.Revision, pgDoc.Revision)
	}

	s2db, s2 := newScriptGorm(t)
	lr := NewLogicRepository(s2db)
	s2.on(`SELECT "Logic", "OwnerId" FROM "Project"`, []string{"Logic", "OwnerId"}, []driver.Value{[]byte(`{"v":1,"variables":{},"flows":{}}`), "u1"})
	s2.on(`SELECT "Id", "Logic" FROM "Page"`, []string{"Id", "Logic"}, []driver.Value{"pg-1", []byte(`{"v":1,"variables":{"x":{"type":"string","default":"hi"}},"flows":{}}`)})
	s2.on(`UPDATE "Project"`, []string{"LogicRevision"}, []driver.Value{int64(4)})
	s2.on(`UPDATE "Page"`, []string{"LogicRevision"}, []driver.Value{int64(5)})
	projMoved, pageMoved, err := lr.Move(ctx, "proj-1", "pg-1", "u1", "variable", "x", "project")
	if err != nil || projMoved.Revision != 4 || pageMoved.Revision != 5 {
		t.Fatalf("Logic.Move err=%v p=%d pg=%d", err, projMoved.Revision, pageMoved.Revision)
	}
}
