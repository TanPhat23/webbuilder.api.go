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

var (
	snapshotBoom = errors.New("boom")
	snapshotCols = []string{"Id", "ProjectId", "Name", "Type"}
)

const (
	snapshotProjectFrom  = `FROM "public"."Project"`
	snapshotSnapshotFrom = `FROM "public"."Snapshot"`
)

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

func snapshotRow(id, typ string) []driver.Value { return []driver.Value{id, "p1", "snap " + id, typ} }

func snapshotWithProject(f *fakeDB) {
	f.on(snapshotProjectFrom, []string{"Id"}, []driver.Value{"p1"})
}

func assertSnapshotErr(t *testing.T, err, target error, prefix string) {
	t.Helper()
	if !errors.Is(err, target) {
		t.Fatalf("want %v, got %v", target, err)
	}
	if prefix != "" && !strings.HasPrefix(err.Error(), prefix) {
		t.Fatalf("want prefix %q, got %v", prefix, err)
	}
}

func TestSnapshotRepositorySaveSnapshotValidation(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	repo := &SnapshotRepository{db: db}
	assertSnapshotErr(t, repo.SaveSnapshot(ctx, "p1", &models.Snapshot{Type: "bogus"}), ErrInvalidSnapshotType, "")
	if f.queryCount() != 0 {
		t.Fatalf("invalid type must not query: %v", f.queries)
	}
	assertSnapshotErr(t, repo.SaveSnapshot(ctx, "p1", &models.Snapshot{Type: SnapshotTypeVersion}), ErrProjectNotFound, "")
	if f.sawQuery("INSERT") {
		t.Fatalf("must not insert for a missing project: %v", f.queries)
	}

	dbErr, fErr := newFakeDB(t)
	fErr.onErr(snapshotProjectFrom, snapshotBoom)
	err := (&SnapshotRepository{db: dbErr}).SaveSnapshot(ctx, "p1", &models.Snapshot{Type: SnapshotTypeVersion})
	assertSnapshotErr(t, err, snapshotBoom, "failed to validate project existence")
}

func TestSnapshotRepositorySaveSnapshotCreatesNonWorkingSnapshots(t *testing.T) {
	ctx := context.Background()
	for _, typ := range []string{SnapshotTypeVersion, SnapshotTypePublished} {
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
	}
}

func TestSnapshotRepositorySaveSnapshotTimestampAndError(t *testing.T) {
	ctx := context.Background()
	stamp := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)
	s := &models.Snapshot{Id: "s1", Type: SnapshotTypeVersion, CreatedAt: stamp}
	db, f := newFakeDB(t)
	snapshotWithProject(f)
	if err := (&SnapshotRepository{db: db}).SaveSnapshot(ctx, "p1", s); err != nil || !s.CreatedAt.Equal(stamp) {
		t.Fatalf("got %v, %v", s.CreatedAt, err)
	}

	dbErr, fErr := newFakeDB(t)
	snapshotWithProject(fErr)
	fErr.onErr("INSERT", snapshotBoom)
	assertSnapshotErr(t, (&SnapshotRepository{db: dbErr}).SaveSnapshot(ctx, "p1", &models.Snapshot{Type: SnapshotTypeVersion}), snapshotBoom, "failed to create snapshot")
}

func TestSnapshotRepositorySaveSnapshotWorkingCreate(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	snapshotWithProject(f)
	s := &models.Snapshot{Id: "new", Type: SnapshotTypeWorking}
	if err := (&SnapshotRepository{db: db}).SaveSnapshot(ctx, "p1", s); err != nil {
		t.Fatal(err)
	}
	if !f.sawQuery("INSERT INTO", "Snapshot") || f.sawQuery("UPDATE") {
		t.Fatalf("expected insert: %v %+v", f.queries, s)
	}
	if s.Id != "new" {
		t.Fatalf("expected id new, got %q", s.Id)
	}
}

func TestSnapshotRepositorySaveSnapshotWorkingUpdate(t *testing.T) {
	ctx := context.Background()
	dbUpd, fUpd := newFakeDB(t)
	snapshotWithProject(fUpd)
	created := time.Date(2021, 5, 6, 7, 8, 9, 0, time.UTC)
	fUpd.on(snapshotSnapshotFrom, append(snapshotCols, "CreatedAt"), append(snapshotRow("old", SnapshotTypeWorking), created))
	fUpd.onExec("UPDATE", 1)
	upd := &models.Snapshot{Id: "new", Type: SnapshotTypeWorking, Name: "renamed"}
	if err := (&SnapshotRepository{db: dbUpd}).SaveSnapshot(ctx, "p1", upd); err != nil {
		t.Fatal(err)
	}
	if upd.Id != "old" || !upd.CreatedAt.Equal(created) {
		t.Fatalf("expected update of old row: %+v", upd)
	}
	if fUpd.sawQuery("INSERT") || !fUpd.sawQuery("UPDATE", "Snapshot") {
		t.Fatalf("expected update query: %v", fUpd.queries)
	}
	if !snapshotHasArg(fUpd, "old") || !snapshotHasArg(fUpd, "renamed") {
		t.Fatalf("expected old/renamed args: %v", fUpd.args)
	}
}

func TestSnapshotRepositorySaveSnapshotWorkingErrors(t *testing.T) {
	ctx := context.Background()
	db1, f1 := newFakeDB(t)
	snapshotWithProject(f1)
	f1.onErr(snapshotSnapshotFrom, snapshotBoom)
	assertSnapshotErr(t, (&SnapshotRepository{db: db1}).SaveSnapshot(ctx, "p1", &models.Snapshot{Type: SnapshotTypeWorking}), snapshotBoom, "failed to check existing working snapshot")

	db2, f2 := newFakeDB(t)
	snapshotWithProject(f2)
	f2.on(snapshotSnapshotFrom, snapshotCols, snapshotRow("old", SnapshotTypeWorking))
	f2.onErr("UPDATE", snapshotBoom)
	assertSnapshotErr(t, (&SnapshotRepository{db: db2}).SaveSnapshot(ctx, "p1", &models.Snapshot{Type: SnapshotTypeWorking}), snapshotBoom, "failed to update working snapshot")
}

func TestSnapshotRepositoryReadLists(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.on("SELECT", snapshotCols, snapshotRow("a", "working"), snapshotRow("b", "version"))
	got, err := (&SnapshotRepository{db: db}).GetSnapshotsByProjectID(ctx, "p1")
	if err != nil || len(got) != 2 || got[0].Id != "a" || got[1].Type != "version" {
		t.Fatalf("got %+v, %v", got, err)
	}
	if !f.sawQuery("ORDER BY", "CreatedAt", "DESC") || !snapshotHasArg(f, "p1") {
		t.Fatalf("unexpected query: %v", f.queries)
	}

	dbErr, fErr := newFakeDB(t)
	fErr.onErr("SELECT", snapshotBoom)
	_, err = (&SnapshotRepository{db: dbErr}).GetSnapshotsByProjectID(ctx, "p1")
	assertSnapshotErr(t, err, snapshotBoom, "failed to get snapshots by project ID")
}

func TestSnapshotRepositoryReadByType(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.on("SELECT", snapshotCols, snapshotRow("a", "version"))
	got, err := (&SnapshotRepository{db: db}).GetSnapshotsByProjectIDAndType(ctx, "p1", "version")
	if err != nil || len(got) != 1 {
		t.Fatalf("got %+v, %v", got, err)
	}
	if got[0].Id != "a" || !snapshotHasArg(f, "version") {
		t.Fatalf("got %+v, args=%v", got, f.args)
	}

	dbErr, fErr := newFakeDB(t)
	fErr.onErr("SELECT", snapshotBoom)
	_, err = (&SnapshotRepository{db: dbErr}).GetSnapshotsByProjectIDAndType(ctx, "p1", "version")
	assertSnapshotErr(t, err, snapshotBoom, "failed to get snapshots by project ID and type")
	_, err = (&SnapshotRepository{db: dbErr}).GetSnapshotsByProjectIDAndType(ctx, "p1", "bogus")
	assertSnapshotErr(t, err, ErrInvalidSnapshotType, "")
}

type snapshotSingleCase struct {
	errMsg string
	want   string
	call   func(r *SnapshotRepository) (*models.Snapshot, error)
}

func snapshotSingleCases(ctx context.Context) []snapshotSingleCase {
	return []snapshotSingleCase{
		{"failed to get snapshot", "", func(r *SnapshotRepository) (*models.Snapshot, error) { return r.GetSnapshotByID(ctx, "a") }},
		{"failed to get latest working snapshot", SnapshotTypeWorking, func(r *SnapshotRepository) (*models.Snapshot, error) { return r.GetLatestWorkingSnapshot(ctx, "p1") }},
		{"failed to get latest published snapshot", SnapshotTypePublished, func(r *SnapshotRepository) (*models.Snapshot, error) {
			return r.GetLatestPublishedSnapshot(ctx, "p1")
		}},
	}
}

func assertSingleSnapshotLookupCase(t *testing.T, tc snapshotSingleCase) {
	t.Helper()
	db, f := newFakeDB(t)
	f.on("SELECT", snapshotCols, snapshotRow("a", "working"))
	got, err := tc.call(&SnapshotRepository{db: db})
	if err != nil || got.Id != "a" || got.ProjectId != "p1" {
		t.Fatalf("got %+v, %v", got, err)
	}
	if tc.want != "" && !snapshotHasArg(f, tc.want) {
		t.Fatalf("want arg %q in %v", tc.want, f.args)
	}

	dbMiss, missFake := newFakeDB(t)
	if missFake == nil {
		t.Fatal("missing fake db")
	}
	if gotMiss, err := tc.call(&SnapshotRepository{db: dbMiss}); gotMiss != nil || !errors.Is(err, ErrSnapshotNotFound) {
		t.Fatalf("got %+v, %v", gotMiss, err)
	}

	dbErr, fErr := newFakeDB(t)
	fErr.onErr("SELECT", snapshotBoom)
	_, err = tc.call(&SnapshotRepository{db: dbErr})
	assertSnapshotErr(t, err, snapshotBoom, tc.errMsg)
}

func TestSnapshotRepositorySingleLookups(t *testing.T) {
	ctx := context.Background()
	for _, tc := range snapshotSingleCases(ctx) {
		assertSingleSnapshotLookupCase(t, tc)
	}
}

func TestSnapshotRepositoryDeleteOneAndByProject(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		affected int64
		dbErr    error
		wantErr  error
		prefix   string
	}{
		{1, nil, nil, ""},
		{0, nil, ErrSnapshotNotFound, ""},
		{0, snapshotBoom, snapshotBoom, "failed to delete snapshot"},
	}
	for _, tc := range tests {
		db, f := newFakeDB(t)
		if tc.dbErr != nil {
			f.onErr("DELETE", tc.dbErr)
		} else {
			f.onExec("DELETE", tc.affected)
		}
		assertSnapshotErr(t, (&SnapshotRepository{db: db}).DeleteSnapshot(ctx, "a"), tc.wantErr, tc.prefix)
		if !snapshotHasArg(f, "a") {
			t.Fatalf("id not bound: %v", f.args)
		}
	}

	dbProj, fProj := newFakeDB(t)
	fProj.onExec("DELETE", 0)
	if err := (&SnapshotRepository{db: dbProj}).DeleteSnapshotsByProjectID(ctx, "p1"); err != nil {
		t.Fatal(err)
	}
	if !fProj.sawQuery("DELETE FROM", "Snapshot", "ProjectId") || !snapshotHasArg(fProj, "p1") {
		t.Fatalf("unexpected: %v", fProj.queries)
	}

	dbErr, fErr := newFakeDB(t)
	fErr.onErr("DELETE", snapshotBoom)
	assertSnapshotErr(t, (&SnapshotRepository{db: dbErr}).DeleteSnapshotsByProjectID(ctx, "p1"), snapshotBoom, "failed to delete snapshots by project ID")
}

func TestSnapshotRepositoryDeleteOld(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.onExec("DELETE", 2)
	before := time.Now().Add(-time.Hour)
	if err := (&SnapshotRepository{db: db}).DeleteOldSnapshots(ctx, "p1", time.Hour); err != nil {
		t.Fatal(err)
	}
	if !f.sawQuery("DELETE FROM", "Type", "CreatedAt") || !snapshotHasArg(f, SnapshotTypeWorking) {
		t.Fatalf("unexpected: %v %v", f.queries, f.args)
	}
	assertRecentCutoffArg(t, f.args[0], before)

	dbErr, fErr := newFakeDB(t)
	fErr.onErr("DELETE", snapshotBoom)
	assertSnapshotErr(t, (&SnapshotRepository{db: dbErr}).DeleteOldSnapshots(ctx, "p1", time.Hour), snapshotBoom, "failed to delete old snapshots")
}

func assertRecentCutoffArg(t *testing.T, args []driver.NamedValue, before time.Time) {
	t.Helper()
	for _, a := range args {
		if ts, ok := a.Value.(time.Time); ok && !ts.Before(before) && time.Since(ts) < time.Hour+time.Minute {
			return
		}
	}
	t.Fatalf("cutoff about one hour ago not bound: %v", args)
}

func TestSnapshotRepositoryCounts(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	cols, rows := countRow(4)
	f.on("count(", cols, rows...)
	n, err := (&SnapshotRepository{db: db}).CountSnapshotsByProjectID(ctx, "p1")
	if err != nil || n != 4 || !snapshotHasArg(f, "p1") {
		t.Fatalf("got %d, %v", n, err)
	}

	dbErr, fErr := newFakeDB(t)
	fErr.onErr("count(", snapshotBoom)
	_, err = (&SnapshotRepository{db: dbErr}).CountSnapshotsByProjectID(ctx, "p1")
	assertSnapshotErr(t, err, snapshotBoom, "failed to count snapshots")

	dbType, fType := newFakeDB(t)
	cols, rows = countRow(2)
	fType.on("count(", cols, rows...)
	n, err = (&SnapshotRepository{db: dbType}).CountSnapshotsByType(ctx, "p1", "published")
	if err != nil || n != 2 || !snapshotHasArg(fType, "published") {
		t.Fatalf("got %d, %v", n, err)
	}

	_, err = (&SnapshotRepository{db: dbErr}).CountSnapshotsByType(ctx, "p1", "published")
	assertSnapshotErr(t, err, snapshotBoom, "failed to count snapshots by type")
	_, err = (&SnapshotRepository{db: dbErr}).CountSnapshotsByType(ctx, "p1", "bogus")
	assertSnapshotErr(t, err, ErrInvalidSnapshotType, "")
}

func TestSnapshotRepositoryExistsByID(t *testing.T) {
	ctx := context.Background()
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
	if got {
		t.Fatal("expected false on error")
	}
	assertSnapshotErr(t, err, snapshotBoom, "failed to check snapshot existence")
}
