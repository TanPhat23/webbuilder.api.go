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
	notificationBoom = errors.New("boom")
	notificationCols = []string{"Id", "UserId", "Type", "Title", "Description", "Read"}
)

func notificationHasArg(f *fakeDB, want any) bool {
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

func notificationRow(id string, read bool) []driver.Value {
	return []driver.Value{id, "u1", "alert", "Title " + id, "desc", read}
}

// notificationCounts answers the seven count queries ListForUser issues, in order.
func notificationCounts(f *fakeDB, values ...int64) {
	for _, v := range values {
		cols, rows := countRow(v)
		f.on("count(", cols, rows...).once()
	}
}

func assertNotificationNoErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func assertNotificationErrPrefix(t *testing.T, err, want error, prefix string) {
	t.Helper()
	if !errors.Is(err, want) {
		t.Fatalf("want %v, got %v", want, err)
	}
	if prefix != "" && !strings.HasPrefix(err.Error(), prefix) {
		t.Fatalf("want prefix %q, got %v", prefix, err)
	}
}

func TestNotificationRepositoryListForUserWithStats(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.on("ORDER BY", notificationCols, notificationRow("n1", false), notificationRow("n2", true))
	notificationCounts(f, 10, 4, 2, 1, 2, 3, 4)
	got, err := (&NotificationRepository{db: db}).ListForUser(ctx, "u1", models.NotificationFilter{})
	assertNotificationNoErr(t, err)
	assertListedNotificationStats(t, f, got)
}

func assertListedNotificationStats(t *testing.T, f *fakeDB, got *models.NotificationList) {
	t.Helper()
	if len(got.Notifications) != 2 || got.Notifications[0].Id != "n1" || !got.Notifications[1].Read {
		t.Fatalf("unexpected notifications %+v", got.Notifications)
	}
	if got.Stats.Total != 10 || got.Stats.Unread != 4 || got.Stats.ReadToday != 2 {
		t.Fatalf("stats %+v", got.Stats)
	}
	fc := got.FilterCounts
	if fc.Messages != 1 || fc.Followers != 2 || fc.Alerts != 3 {
		t.Fatalf("filter counts %+v", fc)
	}
	if fc.Updates != 4 || fc.All != 10 || fc.Unread != 4 {
		t.Fatalf("filter counts %+v", fc)
	}
	for _, typ := range []string{"message", "user", "alert", "settings"} {
		if !notificationHasArg(f, typ) {
			t.Fatalf("count for type %q missing: %v", typ, f.args)
		}
	}
}

func TestNotificationRepositoryListFiltersAndEmpty(t *testing.T) {
	ctx := context.Background()
	dbEmpty, _ := newFakeDB(t)
	emptyGot, err := (&NotificationRepository{db: dbEmpty}).ListForUser(ctx, "u1", models.NotificationFilter{Filter: "all"})
	assertNotificationNoErr(t, err)
	if emptyGot.Notifications == nil || len(emptyGot.Notifications) != 0 {
		t.Fatalf("got %+v", emptyGot)
	}

	dbUnread, fUnread := newFakeDB(t)
	_, err = (&NotificationRepository{db: dbUnread}).ListForUser(ctx, "u1", models.NotificationFilter{Filter: "unread"})
	assertNotificationNoErr(t, err)
	if !fUnread.sawQuery("ORDER BY", "Read", "false") {
		t.Fatalf("list query should filter unread: %v", fUnread.queries)
	}

	dbType, fType := newFakeDB(t)
	_, err = (&NotificationRepository{db: dbType}).ListForUser(ctx, "u1", models.NotificationFilter{Filter: "settings"})
	assertNotificationNoErr(t, err)
	if !fType.sawQuery("ORDER BY", "Type") || !notificationHasArg(fType, "settings") {
		t.Fatalf("list query should filter by type: %v", fType.queries)
	}
}

func TestNotificationRepositoryListSearch(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	_, err := (&NotificationRepository{db: db}).ListForUser(ctx, "u1", models.NotificationFilter{Search: "  hello "})
	assertNotificationNoErr(t, err)
	if !f.sawQuery("ILIKE", "Title", "Description") || !notificationHasArg(f, "%hello%") {
		t.Fatalf("trimmed search missing: %v %v", f.queries, f.args)
	}

	dbBlank, fBlank := newFakeDB(t)
	_, err = (&NotificationRepository{db: dbBlank}).ListForUser(ctx, "u1", models.NotificationFilter{Search: "   "})
	assertNotificationNoErr(t, err)
	if fBlank.sawQuery("ILIKE") {
		t.Fatalf("blank search should not filter: %v", fBlank.queries)
	}
}

func TestNotificationRepositoryListLimits(t *testing.T) {
	ctx := context.Background()
	for _, limit := range []int{0, -3, 500} {
		db, f := newFakeDB(t)
		_, err := (&NotificationRepository{db: db}).ListForUser(ctx, "u1", models.NotificationFilter{Limit: limit})
		assertNotificationNoErr(t, err)
		if !notificationHasArg(f, int64(50)) && !f.sawQuery("LIMIT 50") {
			t.Fatalf("limit %d should clamp to 50: %v %v", limit, f.queries, f.args)
		}
	}
	db, f := newFakeDB(t)
	_, err := (&NotificationRepository{db: db}).ListForUser(ctx, "u1", models.NotificationFilter{Limit: 7})
	assertNotificationNoErr(t, err)
	if !notificationHasArg(f, int64(7)) && !f.sawQuery("LIMIT 7") {
		t.Fatalf("limit 7 missing: %v %v", f.queries, f.args)
	}
}

func TestNotificationRepositoryListReadTodayAndErrors(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	_, err := (&NotificationRepository{db: db}).ListForUser(ctx, "u1", models.NotificationFilter{})
	assertNotificationNoErr(t, err)
	y, m, d := time.Now().Date()
	want := time.Date(y, m, d, 0, 0, 0, 0, time.Local).UTC()
	if !hasBoundTime(f, want) {
		t.Fatalf("start of day %v not bound", want)
	}

	dbErr1, fErr1 := newFakeDB(t)
	fErr1.onErr("ORDER BY", notificationBoom)
	got1, err := (&NotificationRepository{db: dbErr1}).ListForUser(ctx, "u1", models.NotificationFilter{})
	if got1 != nil {
		t.Fatalf("expected nil, got %+v", got1)
	}
	assertNotificationErrPrefix(t, err, notificationBoom, "list notifications")

	dbErr2, fErr2 := newFakeDB(t)
	fErr2.onErr("count(", notificationBoom)
	got2, err := (&NotificationRepository{db: dbErr2}).ListForUser(ctx, "u1", models.NotificationFilter{})
	if got2 != nil {
		t.Fatalf("expected nil, got %+v", got2)
	}
	assertNotificationErrPrefix(t, err, notificationBoom, "count notifications")
}

func hasBoundTime(f *fakeDB, want time.Time) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, args := range f.args {
		for _, a := range args {
			if ts, ok := a.Value.(time.Time); ok && ts.Equal(want) {
				return true
			}
		}
	}
	return false
}

func TestNotificationRepositorySetRead(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.on("SELECT", notificationCols, notificationRow("n1", false))
	got, err := (&NotificationRepository{db: db}).SetRead(ctx, "u1", "n1", true)
	assertNotificationNoErr(t, err)
	if !got.Read || got.Id != "n1" || time.Since(got.UpdatedAt) > time.Minute {
		t.Fatalf("unexpected notification %+v", got)
	}
	if !f.sawQuery("UPDATE", "Notification", "Read") || !notificationHasArg(f, true) {
		t.Fatalf("update missing: %v", f.queries)
	}

	dbUnread, fUnread := newFakeDB(t)
	fUnread.on("SELECT", notificationCols, notificationRow("n1", true))
	unreadGot, err := (&NotificationRepository{db: dbUnread}).SetRead(ctx, "u1", "n1", false)
	assertNotificationNoErr(t, err)
	if unreadGot.Read {
		t.Fatalf("got %+v", unreadGot)
	}
}

func TestNotificationRepositorySetReadErrors(t *testing.T) {
	ctx := context.Background()
	dbMiss, fMiss := newFakeDB(t)
	got, err := (&NotificationRepository{db: dbMiss}).SetRead(ctx, "u1", "n1", true)
	if got != nil || fMiss.sawQuery("UPDATE") {
		t.Fatalf("got %+v, queries=%v", got, fMiss.queries)
	}
	assertNotificationErrPrefix(t, err, ErrNotificationNotFound, "")

	dbLookup, fLookup := newFakeDB(t)
	fLookup.onErr("SELECT", notificationBoom)
	_, err = (&NotificationRepository{db: dbLookup}).SetRead(ctx, "u1", "n1", true)
	if errors.Is(err, ErrNotificationNotFound) {
		t.Fatalf("unexpected not-found: %v", err)
	}
	assertNotificationErrPrefix(t, err, notificationBoom, "get notification")

	dbUpd, fUpd := newFakeDB(t)
	fUpd.on("SELECT", notificationCols, notificationRow("n1", false))
	fUpd.onErr("UPDATE", notificationBoom)
	gotUpd, err := (&NotificationRepository{db: dbUpd}).SetRead(ctx, "u1", "n1", true)
	if gotUpd != nil {
		t.Fatalf("expected nil, got %+v", gotUpd)
	}
	assertNotificationErrPrefix(t, err, notificationBoom, "update notification")
}

func TestNotificationRepositoryMarkAllRead(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.onExec("UPDATE", 3)
	assertNotificationNoErr(t, (&NotificationRepository{db: db}).MarkAllRead(ctx, "u1"))
	if !f.sawQuery("UPDATE", "Notification", "Read", "UserId") || !notificationHasArg(f, "u1") || !notificationHasArg(f, true) {
		t.Fatalf("unexpected statements: %v %v", f.queries, f.args)
	}

	dbErr, fErr := newFakeDB(t)
	fErr.onErr("UPDATE", notificationBoom)
	assertNotificationErrPrefix(t, (&NotificationRepository{db: dbErr}).MarkAllRead(ctx, "u1"), notificationBoom, "")
}

func TestNotificationRepositoryCreate(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	n := &models.Notification{UserId: "u1", Type: "alert", Title: "T", Description: "d"}
	assertNotificationNoErr(t, (&NotificationRepository{db: db}).Create(ctx, n))
	if n.Id == "" || n.CreatedAt.IsZero() {
		t.Fatalf("not stamped: %+v", n)
	}
	if !n.CreatedAt.Equal(n.UpdatedAt) || n.CreatedAt.Location() != time.UTC {
		t.Fatalf("bad timestamps: %+v", n)
	}
	if !f.sawQuery("INSERT INTO", "Notification") || !notificationHasArg(f, n.Id) {
		t.Fatalf("missing insert: %v", f.queries)
	}

	dbFixed, _ := newFakeDB(t)
	fixed := &models.Notification{Id: "fixed", UserId: "u1", Type: "alert", Title: "T"}
	assertNotificationNoErr(t, (&NotificationRepository{db: dbFixed}).Create(ctx, fixed))
	if fixed.Id != "fixed" {
		t.Fatalf("got %+v", fixed)
	}
}

func TestNotificationRepositoryCreateValidationAndError(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	repo := &NotificationRepository{db: db}
	for _, n := range []*models.Notification{
		{Type: "alert", Title: "T"},
		{UserId: "u1", Title: "T"},
		{UserId: "u1", Type: "alert"},
	} {
		if err := repo.Create(ctx, n); err == nil || err.Error() != "userId, type and title are required" {
			t.Fatalf("%+v: %v", n, err)
		}
	}
	if f.queryCount() != 0 {
		t.Fatalf("invalid notifications must not reach the database: %v", f.queries)
	}

	dbErr, fErr := newFakeDB(t)
	fErr.onErr("INSERT", notificationBoom)
	err := (&NotificationRepository{db: dbErr}).Create(ctx, &models.Notification{UserId: "u1", Type: "alert", Title: "T"})
	assertNotificationErrPrefix(t, err, notificationBoom, "")
}
