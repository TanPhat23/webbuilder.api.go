package repositories

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// fakeDB is a scripted database/sql driver. It lets the repository tests run
// without PostgreSQL: every statement is recorded, and each one is answered by
// the first matching rule (or an empty result when no rule matches). The real
// SQL behaviour is covered by the *_postgres_test.go files; this fake covers
// argument validation, query building and error mapping.
type fakeDB struct {
	mu      sync.Mutex
	rules   []*fakeRule
	queries []string
	args    [][]driver.NamedValue
	begins  int
	commits int
	aborts  int
}

type fakeRule struct {
	contains string
	cols     []string
	rows     [][]driver.Value
	affected int64
	err      error
	// remaining is the number of times the rule may still fire; <0 is unlimited.
	remaining int
}

// on registers a rule that answers statements containing the given text. It
// fires on every matching statement until replaced by an earlier rule.
func (f *fakeDB) on(contains string, cols []string, rows ...[]driver.Value) *fakeRule {
	r := &fakeRule{contains: contains, cols: cols, rows: rows, remaining: -1}
	f.mu.Lock()
	f.rules = append(f.rules, r)
	f.mu.Unlock()
	return r
}

// onExec registers a rule for a write statement reporting the given rows affected.
func (f *fakeDB) onExec(contains string, affected int64) *fakeRule {
	r := &fakeRule{contains: contains, affected: affected, remaining: -1}
	f.mu.Lock()
	f.rules = append(f.rules, r)
	f.mu.Unlock()
	return r
}

// onErr registers a rule that makes matching statements fail with err.
func (f *fakeDB) onErr(contains string, err error) *fakeRule {
	r := &fakeRule{contains: contains, err: err, remaining: -1}
	f.mu.Lock()
	f.rules = append(f.rules, r)
	f.mu.Unlock()
	return r
}

// once limits the rule to a single use, so a later rule can answer the next statement.
func (r *fakeRule) once() *fakeRule { r.remaining = 1; return r }

func (f *fakeDB) lastQuery() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.queries) == 0 {
		return ""
	}
	return f.queries[len(f.queries)-1]
}

func (f *fakeDB) queryCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.queries)
}

// sawQuery reports whether any recorded statement contains all the fragments.
func (f *fakeDB) sawQuery(fragments ...string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
next:
	for _, q := range f.queries {
		for _, frag := range fragments {
			if !strings.Contains(q, frag) {
				continue next
			}
		}
		return true
	}
	return false
}

func (f *fakeDB) match(query string, args []driver.NamedValue) *fakeRule {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.queries = append(f.queries, query)
	f.args = append(f.args, args)
	for _, r := range f.rules {
		if r.remaining == 0 || !strings.Contains(query, r.contains) {
			continue
		}
		if r.remaining > 0 {
			r.remaining--
		}
		return r
	}
	return nil
}

type fakeConnector struct{ db *fakeDB }

func (c fakeConnector) Connect(context.Context) (driver.Conn, error) { return &fakeConn{db: c.db}, nil }
func (c fakeConnector) Driver() driver.Driver                        { return fakeDriver{} }

type fakeDriver struct{}

func (fakeDriver) Open(string) (driver.Conn, error) { return nil, errors.New("use the connector") }

type fakeConn struct{ db *fakeDB }

func (*fakeConn) Prepare(string) (driver.Stmt, error) { return nil, driver.ErrSkip }
func (*fakeConn) Close() error                        { return nil }
func (c *fakeConn) Begin() (driver.Tx, error)         { return c.BeginTx(context.Background(), driver.TxOptions{}) }
func (c *fakeConn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	c.db.mu.Lock()
	c.db.begins++
	c.db.mu.Unlock()
	return &fakeTx{db: c.db}, nil
}

func (c *fakeConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	r := c.db.match(query, args)
	if r == nil {
		return &fakeRows{}, nil
	}
	if r.err != nil {
		return nil, r.err
	}
	return &fakeRows{cols: r.cols, rows: r.rows}, nil
}

func (c *fakeConn) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	r := c.db.match(query, args)
	if r == nil {
		return driver.RowsAffected(0), nil
	}
	if r.err != nil {
		return nil, r.err
	}
	return driver.RowsAffected(r.affected), nil
}

type fakeTx struct{ db *fakeDB }

func (t *fakeTx) Commit() error {
	t.db.mu.Lock()
	t.db.commits++
	t.db.mu.Unlock()
	return nil
}

func (t *fakeTx) Rollback() error {
	t.db.mu.Lock()
	t.db.aborts++
	t.db.mu.Unlock()
	return nil
}

type fakeRows struct {
	cols []string
	rows [][]driver.Value
	pos  int
}

func (r *fakeRows) Columns() []string { return r.cols }
func (r *fakeRows) Close() error      { return nil }
func (r *fakeRows) Next(dest []driver.Value) error {
	if r.pos >= len(r.rows) {
		return io.EOF
	}
	copy(dest, r.rows[r.pos])
	r.pos++
	return nil
}

// newFakeDB returns a gorm handle (PostgreSQL dialect, no connection) backed by a scripted fakeDB.
func newFakeDB(t testing.TB) (*gorm.DB, *fakeDB) {
	t.Helper()
	f := &fakeDB{}
	sqlDB := sql.OpenDB(fakeConnector{db: f})
	t.Cleanup(func() { sqlDB.Close() })
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{DisableAutomaticPing: true, Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	return db, f
}

// countRow is the single-row answer to a SELECT count(*) query.
func countRow(n int64) ([]string, [][]driver.Value) {
	return []string{"count"}, [][]driver.Value{{n}}
}
