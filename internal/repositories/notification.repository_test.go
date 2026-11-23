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

// notificationHasArg reports whether any recorded statement was sent with the given bind value.
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

var notificationBoom = errors.New("boom")

var notificationCols = []string{"Id", "UserId", "Type", "Title", "Description", "Read"}

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

func TestNotificationRepositoryListForUser(t *testing.T) {
	ctx := context.Background()
	t.Run("returns notifications with stats and filter counts", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on("ORDER BY", notificationCols, notificationRow("n1", false), notificationRow("n2", true))
		notificationCounts(f, 10, 4, 2, 1, 2, 3, 4)
		got, err := (&NotificationRepository{db: db}).ListForUser(ctx, "u1", models.NotificationFilter{})
		if err != nil {
			t.Fatal(err)
		}
		if len(got.Notifications) != 2 || got.Notifications[0].Id != "n1" || !got.Notifications[1].Read {
			t.Fatalf("unexpected notifications %+v", got.Notifications)
		}
		if got.Stats.Total != 10 || got.Stats.Unread != 4 || got.Stats.ReadToday != 2 {
			t.Fatalf("stats %+v", got.Stats)
		}
		fc := got.FilterCounts
		if fc.Messages != 1 || fc.Followers != 2 || fc.Alerts != 3 || fc.Updates != 4 || fc.All != 10 || fc.Unread != 4 {
			t.Fatalf("filter counts %+v", fc)
		}
		for _, typ := range []string{"message", "user", "alert", "settings"} {
			if !notificationHasArg(f, typ) {
				t.Fatalf("count for type %q missing: %v", typ, f.args)
			}
		}
	})
	t.Run("empty result is an empty slice", func(t *testing.T) {
		db, _ := newFakeDB(t)
		got, err := (&NotificationRepository{db: db}).ListForUser(ctx, "u1", models.NotificationFilter{Filter: "all"})
		if err != nil || got.Notifications == nil || len(got.Notifications) != 0 {
			t.Fatalf("got %+v, %v", got, err)
		}
	})
	t.Run("unread filter", func(t *testing.T) {
		db, f := newFakeDB(t)
		if _, err := (&NotificationRepository{db: db}).ListForUser(ctx, "u1", models.NotificationFilter{Filter: "unread"}); err != nil {
			t.Fatal(err)
		}
		if !f.sawQuery("ORDER BY", "Read", "false") {
			t.Fatalf("list query should filter unread: %v", f.queries)
		}
	})
	t.Run("type filter", func(t *testing.T) {
		db, f := newFakeDB(t)
		if _, err := (&NotificationRepository{db: db}).ListForUser(ctx, "u1", models.NotificationFilter{Filter: "settings"}); err != nil {
			t.Fatal(err)
		}
		if !f.sawQuery("ORDER BY", "Type") || !notificationHasArg(f, "settings") {
			t.Fatalf("list query should filter by type: %v", f.queries)
		}
	})
	t.Run("search matches title and description", func(t *testing.T) {
		db, f := newFakeDB(t)
		if _, err := (&NotificationRepository{db: db}).ListForUser(ctx, "u1", models.NotificationFilter{Search: "  hello "}); err != nil {
			t.Fatal(err)
		}
		if !f.sawQuery("ILIKE", "Title", "Description") || !notificationHasArg(f, "%hello%") {
			t.Fatalf("trimmed search missing: %v %v", f.queries, f.args)
		}
	})
	t.Run("blank search is ignored", func(t *testing.T) {
		db, f := newFakeDB(t)
		if _, err := (&NotificationRepository{db: db}).ListForUser(ctx, "u1", models.NotificationFilter{Search: "   "}); err != nil {
			t.Fatal(err)
		}
		if f.sawQuery("ILIKE") {
			t.Fatalf("blank search should not filter: %v", f.queries)
		}
	})
	t.Run("limit is clamped to the default", func(t *testing.T) {
		for _, limit := range []int{0, -3, 500} {
			db, f := newFakeDB(t)
			if _, err := (&NotificationRepository{db: db}).ListForUser(ctx, "u1", models.NotificationFilter{Limit: limit}); err != nil {
				t.Fatal(err)
			}
			if !notificationHasArg(f, int64(50)) && !f.sawQuery("LIMIT 50") {
				t.Fatalf("limit %d should clamp to 50: %v %v", limit, f.queries, f.args)
			}
		}
	})
	t.Run("a smaller limit is honoured", func(t *testing.T) {
		db, f := newFakeDB(t)
		if _, err := (&NotificationRepository{db: db}).ListForUser(ctx, "u1", models.NotificationFilter{Limit: 7}); err != nil {
			t.Fatal(err)
		}
		if !notificationHasArg(f, int64(7)) && !f.sawQuery("LIMIT 7") {
			t.Fatalf("limit 7 missing: %v %v", f.queries, f.args)
		}
	})
	t.Run("read-today count is bound to the start of the local day in UTC", func(t *testing.T) {
		db, f := newFakeDB(t)
		if _, err := (&NotificationRepository{db: db}).ListForUser(ctx, "u1", models.NotificationFilter{}); err != nil {
			t.Fatal(err)
		}
		y, m, d := time.Now().Date()
		want := time.Date(y, m, d, 0, 0, 0, 0, time.Local).UTC()
		found := false
		f.mu.Lock()
		for _, args := range f.args {
			for _, a := range args {
				if ts, ok := a.Value.(time.Time); ok && ts.Equal(want) {
					found = true
				}
			}
		}
		f.mu.Unlock()
		if !found {
			t.Fatalf("start of day %v not bound", want)
		}
	})
	t.Run("list query error", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onErr("ORDER BY", notificationBoom)
		got, err := (&NotificationRepository{db: db}).ListForUser(ctx, "u1", models.NotificationFilter{})
		if got != nil || !errors.Is(err, notificationBoom) || !strings.HasPrefix(err.Error(), "list notifications") {
			t.Fatalf("got %+v, %v", got, err)
		}
	})
	t.Run("count query error", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onErr("count(", notificationBoom)
		got, err := (&NotificationRepository{db: db}).ListForUser(ctx, "u1", models.NotificationFilter{})
		if got != nil || !errors.Is(err, notificationBoom) || !strings.HasPrefix(err.Error(), "count notifications") {
			t.Fatalf("got %+v, %v", got, err)
		}
	})
}

func TestNotificationRepositorySetRead(t *testing.T) {
	ctx := context.Background()
	t.Run("updates and returns the notification", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on("SELECT", notificationCols, notificationRow("n1", false))
		got, err := (&NotificationRepository{db: db}).SetRead(ctx, "u1", "n1", true)
		if err != nil {
			t.Fatal(err)
		}
		if !got.Read || got.Id != "n1" || time.Since(got.UpdatedAt) > time.Minute {
			t.Fatalf("unexpected notification %+v", got)
		}
		if !f.sawQuery("UPDATE", "Notification", "Read") || !notificationHasArg(f, true) {
			t.Fatalf("update missing: %v", f.queries)
		}
	})
	t.Run("can mark unread", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on("SELECT", notificationCols, notificationRow("n1", true))
		got, err := (&NotificationRepository{db: db}).SetRead(ctx, "u1", "n1", false)
		if err != nil || got.Read {
			t.Fatalf("got %+v, %v", got, err)
		}
	})
	t.Run("not found for this user", func(t *testing.T) {
		db, f := newFakeDB(t)
		got, err := (&NotificationRepository{db: db}).SetRead(ctx, "u1", "n1", true)
		if got != nil || !errors.Is(err, ErrNotificationNotFound) {
			t.Fatalf("got %+v, %v", got, err)
		}
		if f.sawQuery("UPDATE") {
			t.Fatal("must not update a missing notification")
		}
	})
	t.Run("lookup error", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onErr("SELECT", notificationBoom)
		_, err := (&NotificationRepository{db: db}).SetRead(ctx, "u1", "n1", true)
		if !errors.Is(err, notificationBoom) || errors.Is(err, ErrNotificationNotFound) || !strings.HasPrefix(err.Error(), "get notification") {
			t.Fatal(err)
		}
	})
	t.Run("update error", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on("SELECT", notificationCols, notificationRow("n1", false))
		f.onErr("UPDATE", notificationBoom)
		got, err := (&NotificationRepository{db: db}).SetRead(ctx, "u1", "n1", true)
		if got != nil || !errors.Is(err, notificationBoom) || !strings.HasPrefix(err.Error(), "update notification") {
			t.Fatalf("got %+v, %v", got, err)
		}
	})
}

func TestNotificationRepositoryMarkAllRead(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.onExec("UPDATE", 3)
	if err := (&NotificationRepository{db: db}).MarkAllRead(ctx, "u1"); err != nil {
		t.Fatal(err)
	}
	if !f.sawQuery("UPDATE", "Notification", "Read", "UserId") || !notificationHasArg(f, "u1") || !notificationHasArg(f, true) {
		t.Fatalf("unexpected statements: %v %v", f.queries, f.args)
	}

	db, f = newFakeDB(t)
	f.onErr("UPDATE", notificationBoom)
	if err := (&NotificationRepository{db: db}).MarkAllRead(ctx, "u1"); !errors.Is(err, notificationBoom) {
		t.Fatal(err)
	}
}

func TestNotificationRepositoryCreate(t *testing.T) {
	ctx := context.Background()
	t.Run("stamps id and timestamps", func(t *testing.T) {
		db, f := newFakeDB(t)
		n := &models.Notification{UserId: "u1", Type: "alert", Title: "T", Description: "d"}
		if err := (&NotificationRepository{db: db}).Create(ctx, n); err != nil {
			t.Fatal(err)
		}
		if n.Id == "" || n.CreatedAt.IsZero() || !n.CreatedAt.Equal(n.UpdatedAt) || n.CreatedAt.Location() != time.UTC {
			t.Fatalf("not stamped: %+v", n)
		}
		if !f.sawQuery("INSERT INTO", "Notification") || !notificationHasArg(f, n.Id) {
			t.Fatalf("missing insert: %v", f.queries)
		}
	})
	t.Run("keeps a caller supplied id", func(t *testing.T) {
		db, _ := newFakeDB(t)
		n := &models.Notification{Id: "fixed", UserId: "u1", Type: "alert", Title: "T"}
		if err := (&NotificationRepository{db: db}).Create(ctx, n); err != nil || n.Id != "fixed" {
			t.Fatalf("got %+v, %v", n, err)
		}
	})
	t.Run("requires user, type and title", func(t *testing.T) {
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
	})
	t.Run("database error is returned", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onErr("INSERT", notificationBoom)
		err := (&NotificationRepository{db: db}).Create(ctx, &models.Notification{UserId: "u1", Type: "alert", Title: "T"})
		if !errors.Is(err, notificationBoom) {
			t.Fatal(err)
		}
	})
}
