package repositories

import (
	"context"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"

	"my-go-app/internal/models"
)

// userHasArg reports whether any recorded statement was sent with the given bind value.
func userHasArg(f *fakeDB, want any) bool {
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

var userBoom = errors.New("boom")

var userCols = []string{"Id", "Email", "FirstName"}

func userRow(id, email, first string) []driver.Value {
	return []driver.Value{id, email, first}
}

func TestUserRepositoryLookups(t *testing.T) {
	ctx := context.Background()
	lookups := []struct {
		name   string
		arg    any
		errMsg string
		call   func(r *UserRepository) (*models.User, error)
	}{
		{"by id", "u1", "failed to get user by ID", func(r *UserRepository) (*models.User, error) { return r.GetUserByID(ctx, "u1") }},
		{"by email", "a@b.c", "failed to get user by email", func(r *UserRepository) (*models.User, error) { return r.GetUserByEmail(ctx, "a@b.c") }},
		{"by username", "%ann%", "failed to get user by username", func(r *UserRepository) (*models.User, error) { return r.GetUserByUsername(ctx, "ann") }},
	}
	for _, tc := range lookups {
		t.Run(tc.name+" found", func(t *testing.T) {
			db, f := newFakeDB(t)
			f.on("SELECT", userCols, userRow("u1", "a@b.c", "Ann"))
			got, err := tc.call(&UserRepository{db: db})
			if err != nil || got.Id != "u1" || got.Email != "a@b.c" || got.FirstName == nil || *got.FirstName != "Ann" {
				t.Fatalf("got %+v, %v", got, err)
			}
			if !userHasArg(f, tc.arg) {
				t.Fatalf("expected bind %v: %v", tc.arg, f.args)
			}
		})
		t.Run(tc.name+" missing", func(t *testing.T) {
			db, _ := newFakeDB(t)
			got, err := tc.call(&UserRepository{db: db})
			if got != nil || !errors.Is(err, ErrUserNotFound) {
				t.Fatalf("got %+v, %v", got, err)
			}
		})
		t.Run(tc.name+" database error", func(t *testing.T) {
			db, f := newFakeDB(t)
			f.onErr("SELECT", userBoom)
			got, err := tc.call(&UserRepository{db: db})
			if got != nil || !errors.Is(err, userBoom) || errors.Is(err, ErrUserNotFound) || !strings.HasPrefix(err.Error(), tc.errMsg) {
				t.Fatalf("got %+v, %v", got, err)
			}
		})
	}
}

func TestUserRepositoryGetUserByUsernameMatchesEitherName(t *testing.T) {
	db, f := newFakeDB(t)
	if _, err := (&UserRepository{db: db}).GetUserByUsername(context.Background(), "ann"); !errors.Is(err, ErrUserNotFound) {
		t.Fatal(err)
	}
	if !f.sawQuery("FirstName", "LastName", "ILIKE") {
		t.Fatalf("expected first/last name search: %v", f.queries)
	}
}

func TestUserRepositorySearchUsers(t *testing.T) {
	ctx := context.Background()
	t.Run("matches are returned and capped", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on("SELECT", userCols, userRow("u1", "ann@x.io", "Ann"), userRow("u2", "andy@x.io", "Andy"))
		got, err := (&UserRepository{db: db}).SearchUsers(ctx, "an")
		if err != nil || len(got) != 2 || got[1].Id != "u2" {
			t.Fatalf("got %+v, %v", got, err)
		}
		if !userHasArg(f, "an%") || !userHasArg(f, "%an%") {
			t.Fatalf("email prefix and substring binds expected: %v", f.args)
		}
		q := f.lastQuery()
		if !strings.Contains(q, "LIMIT") || strings.Contains(q, "Logic") {
			t.Fatalf("unexpected query %s", q)
		}
	})
	t.Run("database error", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onErr("SELECT", userBoom)
		got, err := (&UserRepository{db: db}).SearchUsers(ctx, "an")
		if got != nil || !errors.Is(err, userBoom) || !strings.HasPrefix(err.Error(), "failed to search users") {
			t.Fatalf("got %+v, %v", got, err)
		}
	})
}
