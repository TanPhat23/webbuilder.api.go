package repositories

import (
	"context"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"
	"time"

	"my-go-app/internal/models"
)

// snapshotHasArg reports whether any recorded statement was sent with the given bind value.
func snapshotHasArg(f *fakeDB, want any) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, args := range f.args {
		for _, a := range args {
			if a.Value == want {
				return true
			}
		}
	}
	return false
}

var snapshotBoom = errors.New("boom")

var snapshotCols = []string{"Id", "ProjectId", "Name", "Type"}

func snapshotRow(id, typ string) []driver.Value { return []driver.Value{id, "p1", "snap " + id, typ} }

const (
	snapshotProjectFrom  = `FROM "public"."Project"`
	snapshotSnapshotFrom = `FROM "public"."Snapshot"`
)

func snapshotWithProject(f *fakeDB) {
	f.on(snapshotProjectFrom, []string{"Id"}, []driver.Value{"p1"})
}

func TestSnapshotRepositorySaveSnapshotValidation(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	repo := &SnapshotRepository{db: db}
	if err := repo.SaveSnapshot(ctx, "p1", &models.Snapshot{Type: "bogus"}); !errors.Is(err, ErrInvalidSnapshotType) {
		t.Fatalf("unknown type: %v", err)
	}
	if f.queryCount() != 0 {
		t.Fatalf("invalid type must not query: %v", f.queries)
	}
	err := repo.SaveSnapshot(ctx, "p1", &models.Snapshot{Type: SnapshotTypeVersion})
	if !errors.Is(err, ErrProjectNotFound) {
		t.Fatalf("missing project: %v", err)
	}
	if f.sawQuery("INSERT") {
		t.Fatalf("must not insert for a missing project: %v", f.queries)
	}

	db, f = newFakeDB(t)
	f.onErr(snapshotProjectFrom, snapshotBoom)
	err = (&SnapshotRepository{db: db}).SaveSnapshot(ctx, "p1", &models.Snapshot{Type: SnapshotTypeVersion})
	if !errors.Is(err, snapshotBoom) || !strings.HasPrefix(err.Error(), "failed to validate project existence") {
		t.Fatal(err)
	}
}

func TestSnapshotRepositorySaveSnapshotCreatesNonWorkingSnapshots(t *testing.T) {
	ctx := context.Background()
	for _, typ := range []string{SnapshotTypeVersion, SnapshotTypePublished} {
		t.Run(typ, func(t *testing.T) {
			db, f := newFakeDB(t)
			snapshotWithProject(f)
			s := &models.Snapshot{Id: "s1", Type: typ, Name: "n"}
			if err := (&SnapshotRepository{db: db}).SaveSnapshot(ctx, "p1", s); err != nil {
				t.Fatal(err)
			}
			if s.ProjectId != "p1" || s.CreatedAt.IsZero() {
				t.Fatalf("not linked/stamped: %+v", s)
			}
			if !f.sawQuery("INSERT INTO", "Snapshot") || f.sawQuery("UPDATE") || f.sawQuery(snapshotSnapshotFrom) {
				t.Fatalf("expected a plain insert: %v", f.queries)
			}
		})
	}
	t.Run("keeps a preset CreatedAt", func(t *testing.T) {
		stamp := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)
		s := &models.Snapshot{Id: "s1", Type: SnapshotTypeVersion, CreatedAt: stamp}
		db, f := newFakeDB(t)
		snapshotWithProject(f)
		if err := (&SnapshotRepository{db: db}).SaveSnapshot(ctx, "p1", s); err != nil || !s.CreatedAt.Equal(stamp) {
			t.Fatalf("got %v, %v", s.CreatedAt, err)
		}
	})
	t.Run("insert error", func(t *testing.T) {
		db, f := newFakeDB(t)
		snapshotWithProject(f)
		f.onErr("INSERT", snapshotBoom)
		err := (&SnapshotRepository{db: db}).SaveSnapshot(ctx, "p1", &models.Snapshot{Type: SnapshotTypeVersion})
		if !errors.Is(err, snapshotBoom) || !strings.HasPrefix(err.Error(), "failed to create snapshot") {
			t.Fatal(err)
		}
	})
}

func TestSnapshotRepositorySaveSnapshotWorking(t *testing.T) {
	ctx := context.Background()
	t.Run("first working snapshot is inserted", func(t *testing.T) {
		db, f := newFakeDB(t)
		snapshotWithProject(f)
		s := &models.Snapshot{Id: "new", Type: SnapshotTypeWorking}
		if err := (&SnapshotRepository{db: db}).SaveSnapshot(ctx, "p1", s); err != nil {
			t.Fatal(err)
		}
		if !f.sawQuery("INSERT INTO", "Snapshot") || f.sawQuery("UPDATE") || s.Id != "new" {
			t.Fatalf("expected insert: %v %+v", f.queries, s)
		}
	})
	t.Run("existing working snapshot is updated in place", func(t *testing.T) {
		db, f := newFakeDB(t)
		snapshotWithProject(f)
		created := time.Date(2021, 5, 6, 7, 8, 9, 0, time.UTC)
		f.on(snapshotSnapshotFrom, append(snapshotCols, "CreatedAt"), append(snapshotRow("old", SnapshotTypeWorking), created))
		f.onExec("UPDATE", 1)
		s := &models.Snapshot{Id: "new", Type: SnapshotTypeWorking, Name: "renamed"}
		if err := (&SnapshotRepository{db: db}).SaveSnapshot(ctx, "p1", s); err != nil {
			t.Fatal(err)
		}
		if s.Id != "old" || !s.CreatedAt.Equal(created) {
			t.Fatalf("existing identity not adopted: %+v", s)
		}
		if f.sawQuery("INSERT") || !f.sawQuery("UPDATE", "Snapshot") || !snapshotHasArg(f, "old") || !snapshotHasArg(f, "renamed") {
			t.Fatalf("expected update of the old row: %v %v", f.queries, f.args)
		}
	})
	t.Run("lookup error", func(t *testing.T) {
		db, f := newFakeDB(t)
		snapshotWithProject(f)
		f.onErr(snapshotSnapshotFrom, snapshotBoom)
		err := (&SnapshotRepository{db: db}).SaveSnapshot(ctx, "p1", &models.Snapshot{Type: SnapshotTypeWorking})
		if !errors.Is(err, snapshotBoom) || !strings.HasPrefix(err.Error(), "failed to check existing working snapshot") {
			t.Fatal(err)
		}
	})
	t.Run("update error", func(t *testing.T) {
		db, f := newFakeDB(t)
		snapshotWithProject(f)
		f.on(snapshotSnapshotFrom, snapshotCols, snapshotRow("old", SnapshotTypeWorking))
		f.onErr("UPDATE", snapshotBoom)
		err := (&SnapshotRepository{db: db}).SaveSnapshot(ctx, "p1", &models.Snapshot{Type: SnapshotTypeWorking})
		if !errors.Is(err, snapshotBoom) || !strings.HasPrefix(err.Error(), "failed to update working snapshot") {
			t.Fatal(err)
		}
	})
}

func TestSnapshotRepositoryReads(t *testing.T) {
	ctx := context.Background()
	t.Run("by project", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on("SELECT", snapshotCols, snapshotRow("a", "working"), snapshotRow("b", "version"))
		got, err := (&SnapshotRepository{db: db}).GetSnapshotsByProjectID(ctx, "p1")
		if err != nil || len(got) != 2 || got[0].Id != "a" || got[1].Type != "version" {
			t.Fatalf("got %+v, %v", got, err)
		}
		if !f.sawQuery("ORDER BY", "CreatedAt", "DESC") || !snapshotHasArg(f, "p1") {
			t.Fatalf("unexpected query: %v", f.queries)
		}
		db, f = newFakeDB(t)
		f.onErr("SELECT", snapshotBoom)
		got, err = (&SnapshotRepository{db: db}).GetSnapshotsByProjectID(ctx, "p1")
		if got != nil || !errors.Is(err, snapshotBoom) || !strings.HasPrefix(err.Error(), "failed to get snapshots by project ID") {
			t.Fatalf("got %+v, %v", got, err)
		}
	})
	t.Run("by project and type", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on("SELECT", snapshotCols, snapshotRow("a", "version"))
		got, err := (&SnapshotRepository{db: db}).GetSnapshotsByProjectIDAndType(ctx, "p1", "version")
		if err != nil || len(got) != 1 || got[0].Id != "a" || !snapshotHasArg(f, "version") {
			t.Fatalf("got %+v, %v", got, err)
		}
		db, f = newFakeDB(t)
		f.onErr("SELECT", snapshotBoom)
		_, err = (&SnapshotRepository{db: db}).GetSnapshotsByProjectIDAndType(ctx, "p1", "version")
		if !errors.Is(err, snapshotBoom) || !strings.HasPrefix(err.Error(), "failed to get snapshots by project ID and type") {
			t.Fatal(err)
		}
		if _, err = (&SnapshotRepository{db: db}).GetSnapshotsByProjectIDAndType(ctx, "p1", "bogus"); !errors.Is(err, ErrInvalidSnapshotType) {
			t.Fatal(err)
		}
	})
	singles := []struct {
		name   string
		errMsg string
		want   string
		call   func(r *SnapshotRepository) (*models.Snapshot, error)
	}{
		{"by id", "failed to get snapshot", "", func(r *SnapshotRepository) (*models.Snapshot, error) { return r.GetSnapshotByID(ctx, "a") }},
		{"latest working", "failed to get latest working snapshot", SnapshotTypeWorking, func(r *SnapshotRepository) (*models.Snapshot, error) {
			return r.GetLatestWorkingSnapshot(ctx, "p1")
		}},
		{"latest published", "failed to get latest published snapshot", SnapshotTypePublished, func(r *SnapshotRepository) (*models.Snapshot, error) {
			return r.GetLatestPublishedSnapshot(ctx, "p1")
		}},
	}
	for _, tc := range singles {
		t.Run(tc.name, func(t *testing.T) {
			db, f := newFakeDB(t)
			f.on("SELECT", snapshotCols, snapshotRow("a", "working"))
			got, err := tc.call(&SnapshotRepository{db: db})
			if err != nil || got.Id != "a" || got.ProjectId != "p1" {
				t.Fatalf("got %+v, %v", got, err)
			}
			if tc.want != "" && !snapshotHasArg(f, tc.want) {
				t.Fatalf("type %q not bound: %v", tc.want, f.args)
			}
		})
		t.Run(tc.name+" missing", func(t *testing.T) {
			db, _ := newFakeDB(t)
			if got, err := tc.call(&SnapshotRepository{db: db}); got != nil || !errors.Is(err, ErrSnapshotNotFound) {
				t.Fatalf("got %+v, %v", got, err)
			}
		})
		t.Run(tc.name+" error", func(t *testing.T) {
			db, f := newFakeDB(t)
			f.onErr("SELECT", snapshotBoom)
			_, err := tc.call(&SnapshotRepository{db: db})
			if !errors.Is(err, snapshotBoom) || errors.Is(err, ErrSnapshotNotFound) || !strings.HasPrefix(err.Error(), tc.errMsg) {
				t.Fatal(err)
			}
		})
	}
}

func TestSnapshotRepositoryDeletes(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name     string
		affected int64
		dbErr    error
		wantErr  error
		prefix   string
	}{
		{"deleted", 1, nil, nil, ""},
		{"missing", 0, nil, ErrSnapshotNotFound, ""},
		{"database error", 0, snapshotBoom, snapshotBoom, "failed to delete snapshot"},
	}
	for _, tc := range tests {
		t.Run("one "+tc.name, func(t *testing.T) {
			db, f := newFakeDB(t)
			if tc.dbErr != nil {
				f.onErr("DELETE", tc.dbErr)
			} else {
				f.onExec("DELETE", tc.affected)
			}
			err := (&SnapshotRepository{db: db}).DeleteSnapshot(ctx, "a")
			if !errors.Is(err, tc.wantErr) || (tc.prefix != "" && !strings.HasPrefix(err.Error(), tc.prefix)) {
				t.Fatalf("want %v, got %v", tc.wantErr, err)
			}
			if !snapshotHasArg(f, "a") {
				t.Fatalf("id not bound: %v", f.args)
			}
		})
	}
	t.Run("all for a project, even when none exist", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onExec("DELETE", 0)
		if err := (&SnapshotRepository{db: db}).DeleteSnapshotsByProjectID(ctx, "p1"); err != nil {
			t.Fatal(err)
		}
		if !f.sawQuery("DELETE FROM", "Snapshot", "ProjectId") || !snapshotHasArg(f, "p1") {
			t.Fatalf("unexpected: %v", f.queries)
		}
		db, f = newFakeDB(t)
		f.onErr("DELETE", snapshotBoom)
		err := (&SnapshotRepository{db: db}).DeleteSnapshotsByProjectID(ctx, "p1")
		if !errors.Is(err, snapshotBoom) || !strings.HasPrefix(err.Error(), "failed to delete snapshots by project ID") {
			t.Fatal(err)
		}
	})
	t.Run("old snapshots spare working ones and respect the cutoff", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onExec("DELETE", 2)
		before := time.Now().Add(-time.Hour)
		if err := (&SnapshotRepository{db: db}).DeleteOldSnapshots(ctx, "p1", time.Hour); err != nil {
			t.Fatal(err)
		}
		if !f.sawQuery("DELETE FROM", "Type", "CreatedAt") || !snapshotHasArg(f, SnapshotTypeWorking) {
			t.Fatalf("unexpected: %v %v", f.queries, f.args)
		}
		found := false
		for _, a := range f.args[0] {
			if ts, ok := a.Value.(time.Time); ok && !ts.Before(before) && time.Since(ts) < time.Hour+time.Minute {
				found = true
			}
		}
		if !found {
			t.Fatalf("cutoff about one hour ago not bound: %v", f.args)
		}
		db, f = newFakeDB(t)
		f.onErr("DELETE", snapshotBoom)
		err := (&SnapshotRepository{db: db}).DeleteOldSnapshots(ctx, "p1", time.Hour)
		if !errors.Is(err, snapshotBoom) || !strings.HasPrefix(err.Error(), "failed to delete old snapshots") {
			t.Fatal(err)
		}
	})
}

func TestSnapshotRepositoryCounts(t *testing.T) {
	ctx := context.Background()
	t.Run("by project", func(t *testing.T) {
		db, f := newFakeDB(t)
		cols, rows := countRow(4)
		f.on("count(", cols, rows...)
		n, err := (&SnapshotRepository{db: db}).CountSnapshotsByProjectID(ctx, "p1")
		if err != nil || n != 4 || !snapshotHasArg(f, "p1") {
			t.Fatalf("got %d, %v", n, err)
		}
		db, f = newFakeDB(t)
		f.onErr("count(", snapshotBoom)
		n, err = (&SnapshotRepository{db: db}).CountSnapshotsByProjectID(ctx, "p1")
		if n != 0 || !errors.Is(err, snapshotBoom) || !strings.HasPrefix(err.Error(), "failed to count snapshots") {
			t.Fatalf("got %d, %v", n, err)
		}
	})
	t.Run("by type", func(t *testing.T) {
		db, f := newFakeDB(t)
		cols, rows := countRow(2)
		f.on("count(", cols, rows...)
		n, err := (&SnapshotRepository{db: db}).CountSnapshotsByType(ctx, "p1", "published")
		if err != nil || n != 2 || !snapshotHasArg(f, "published") {
			t.Fatalf("got %d, %v", n, err)
		}
		db, f = newFakeDB(t)
		f.onErr("count(", snapshotBoom)
		_, err = (&SnapshotRepository{db: db}).CountSnapshotsByType(ctx, "p1", "published")
		if !errors.Is(err, snapshotBoom) || !strings.HasPrefix(err.Error(), "failed to count snapshots by type") {
			t.Fatal(err)
		}
		if _, err = (&SnapshotRepository{db: db}).CountSnapshotsByType(ctx, "p1", "bogus"); !errors.Is(err, ErrInvalidSnapshotType) {
			t.Fatal(err)
		}
	})
	t.Run("exists", func(t *testing.T) {
		for _, tc := range []struct {
			n    int64
			want bool
		}{{0, false}, {1, true}} {
			db, f := newFakeDB(t)
			cols, rows := countRow(tc.n)
			f.on("count(", cols, rows...)
			got, err := (&SnapshotRepository{db: db}).ExistsByID(ctx, "a")
			if err != nil || got != tc.want {
				t.Fatalf("count %d: %v %v", tc.n, got, err)
			}
		}
		db, f := newFakeDB(t)
		f.onErr("count(", snapshotBoom)
		got, err := (&SnapshotRepository{db: db}).ExistsByID(ctx, "a")
		if got || !errors.Is(err, snapshotBoom) || !strings.HasPrefix(err.Error(), "failed to check snapshot existence") {
			t.Fatalf("%v %v", got, err)
		}
	})
}
