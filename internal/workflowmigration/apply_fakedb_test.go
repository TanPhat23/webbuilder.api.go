package workflowmigration

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"strings"
	"testing"
)

// In-memory database/sql driver that answers exactly the statements issued by
// planInTx, Apply and DryRun, so the persistence logic runs without PostgreSQL.

type applyFakeProject struct {
	id    string
	logic []byte
}
type applyFakePage struct {
	id, project string
	logic       []byte
}
type applyFakeElement struct {
	id       string
	page     any // string or nil
	settings []byte
}
type applyFakeWorkflow struct {
	id, project string
	enabled     bool
	handlers    []byte
}
type applyFakeLink struct{ id, workflow, element, event string }

type applyFake struct {
	projects  []applyFakeProject
	pages     []applyFakePage
	elements  []applyFakeElement
	workflows []applyFakeWorkflow
	links     []applyFakeLink

	failQuery map[string]error // substring of statement -> error
	failExec  map[string]error
	rowsErr   map[string]error // substring -> error reported by rows.Err
	nullFirst map[string]bool  // substring -> first column returned as NULL
	hook      func(f *applyFake, q string)
	beginErr  error
	commitErr error
	beginOpts driver.TxOptions
	commits   int
	rollbacks int
	execs     []string
}

func (f *applyFake) db(t *testing.T) *sql.DB {
	t.Helper()
	db := sql.OpenDB(applyFakeConnector{f})
	t.Cleanup(func() { db.Close() })
	return db
}

func (f *applyFake) element(id string) *applyFakeElement {
	for i := range f.elements {
		if f.elements[i].id == id {
			return &f.elements[i]
		}
	}
	return nil
}

func (f *applyFake) page(id string) *applyFakePage {
	for i := range f.pages {
		if f.pages[i].id == id {
			return &f.pages[i]
		}
	}
	return nil
}

func (f *applyFake) project(id string) *applyFakeProject {
	for i := range f.projects {
		if f.projects[i].id == id {
			return &f.projects[i]
		}
	}
	return nil
}

type applyFakeConnector struct{ f *applyFake }

func (c applyFakeConnector) Connect(context.Context) (driver.Conn, error) {
	return &applyFakeConn{c.f}, nil
}
func (c applyFakeConnector) Driver() driver.Driver { return applyFakeDriver{c.f} }

type applyFakeDriver struct{ f *applyFake }

func (d applyFakeDriver) Open(string) (driver.Conn, error) { return &applyFakeConn{d.f}, nil }

type applyFakeConn struct{ f *applyFake }

func (c *applyFakeConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare unsupported")
}
func (c *applyFakeConn) Close() error { return nil }
func (c *applyFakeConn) Begin() (driver.Tx, error) {
	return c.BeginTx(context.Background(), driver.TxOptions{})
}
func (c *applyFakeConn) BeginTx(_ context.Context, opts driver.TxOptions) (driver.Tx, error) {
	if c.f.beginErr != nil {
		return nil, c.f.beginErr
	}
	c.f.beginOpts = opts
	return applyFakeTx{c.f}, nil
}

type applyFakeTx struct{ f *applyFake }

func (t applyFakeTx) Commit() error {
	if t.f.commitErr != nil {
		return t.f.commitErr
	}
	t.f.commits++
	return nil
}
func (t applyFakeTx) Rollback() error { t.f.rollbacks++; return nil }

func (c *applyFakeConn) ExecContext(_ context.Context, q string, args []driver.NamedValue) (driver.Result, error) {
	f := c.f
	f.execs = append(f.execs, q)
	for sub, err := range f.failExec {
		if strings.Contains(q, sub) {
			return nil, err
		}
	}
	switch {
	case strings.HasPrefix(q, "LOCK TABLE"):
	case strings.HasPrefix(q, `UPDATE "Element"`):
		if e := f.element(args[1].Value.(string)); e != nil {
			e.settings = []byte(args[0].Value.(string))
		}
	case strings.HasPrefix(q, `UPDATE "Page"`):
		if p := f.page(args[1].Value.(string)); p != nil {
			p.logic = []byte(args[0].Value.(string))
		}
	case strings.HasPrefix(q, `UPDATE "Project"`):
		if p := f.project(args[1].Value.(string)); p != nil {
			p.logic = []byte(args[0].Value.(string))
		}
	default:
		return nil, errors.New("fake: unexpected exec " + q)
	}
	return driver.RowsAffected(1), nil
}

func (c *applyFakeConn) QueryContext(_ context.Context, q string, args []driver.NamedValue) (driver.Rows, error) {
	f := c.f
	if f.hook != nil {
		f.hook(f, q)
	}
	for sub, err := range f.failQuery {
		if strings.Contains(q, sub) {
			return nil, err
		}
	}
	var cols []string
	var data [][]driver.Value
	switch q {
	case `SELECT "Id", "Logic" FROM "Project"`:
		cols = []string{"Id", "Logic"}
		for _, p := range f.projects {
			data = append(data, []driver.Value{p.id, p.logic})
		}
	case `SELECT "Id", "ProjectId", "Logic" FROM "Page"`:
		cols = []string{"Id", "ProjectId", "Logic"}
		for _, p := range f.pages {
			data = append(data, []driver.Value{p.id, p.project, p.logic})
		}
	case `SELECT "Id", "PageId", "Settings" FROM "Element"`:
		cols = []string{"Id", "PageId", "Settings"}
		for _, e := range f.elements {
			data = append(data, []driver.Value{e.id, e.page, e.settings})
		}
	case `SELECT "Id", "ProjectId", "Enabled", "Handlers" FROM "EventWorkflow"`:
		cols = []string{"Id", "ProjectId", "Enabled", "Handlers"}
		for _, w := range f.workflows {
			data = append(data, []driver.Value{w.id, w.project, w.enabled, w.handlers})
		}
	case `SELECT "Id", "WorkflowId", "ElementId", "EventName" FROM "ElementEventWorkflow"`:
		cols = []string{"Id", "WorkflowId", "ElementId", "EventName"}
		for _, l := range f.links {
			data = append(data, []driver.Value{l.id, l.workflow, l.element, l.event})
		}
	case `SELECT "Settings" FROM "Element" WHERE "Id"=$1`:
		cols = []string{"Settings"}
		if e := f.element(args[0].Value.(string)); e != nil {
			data = append(data, []driver.Value{e.settings})
		}
	case `SELECT "Logic" FROM "Page" WHERE "Id"=$1`:
		cols = []string{"Logic"}
		if p := f.page(args[0].Value.(string)); p != nil {
			data = append(data, []driver.Value{p.logic})
		}
	case `SELECT "Logic" FROM "Project" WHERE "Id"=$1`:
		cols = []string{"Logic"}
		if p := f.project(args[0].Value.(string)); p != nil {
			data = append(data, []driver.Value{p.logic})
		}
	default:
		return nil, errors.New("fake: unexpected query " + q)
	}
	rows := &applyFakeRows{cols: cols, data: data}
	for sub, err := range f.rowsErr {
		if strings.Contains(q, sub) {
			rows.err = err
		}
	}
	for sub := range f.nullFirst {
		if strings.Contains(q, sub) {
			for _, row := range rows.data {
				row[0] = nil
			}
		}
	}
	return rows, nil
}

type applyFakeRows struct {
	cols []string
	data [][]driver.Value
	pos  int
	err  error
}

func (r *applyFakeRows) Columns() []string { return r.cols }
func (r *applyFakeRows) Close() error      { return nil }
func (r *applyFakeRows) Next(dest []driver.Value) error {
	if r.pos >= len(r.data) {
		if r.err != nil {
			return r.err
		}
		return io.EOF
	}
	copy(dest, r.data[r.pos])
	r.pos++
	return nil
}

const applyFakeResetHandlers = `[{"config":{"type":"resetForm","formElementId":"f"}}]`

// applyFakeSeed: workflow w, linked to two elements of one page.
func applyFakeSeed() *applyFake {
	return &applyFake{
		projects:  []applyFakeProject{{id: "p"}},
		pages:     []applyFakePage{{id: "page", project: "p"}},
		elements:  []applyFakeElement{{id: "e1", page: "page", settings: []byte(`{"keep":true}`)}, {id: "e2", page: "page"}},
		workflows: []applyFakeWorkflow{{id: "w", project: "p", enabled: true, handlers: []byte(applyFakeResetHandlers)}},
		links:     []applyFakeLink{{"a", "w", "e1", "onClick"}, {"b", "w", "e2", "onClick"}},
	}
}
