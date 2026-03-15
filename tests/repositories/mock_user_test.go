package repositories_test

import (
	"context"
	"errors"
	"testing"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/tests/testutil"
)

func TestMockUserRepository_DefaultsReturnSentinel(t *testing.T) {
	repo := &testutil.MockUserRepository{}
	ctx := context.Background()

	_, err := repo.GetUserByID(ctx, "uid-1")
	if !errors.Is(err, repositories.ErrUserNotFound) {
		t.Errorf("GetUserByID default: want ErrUserNotFound, got %v", err)
	}

	_, err = repo.GetUserByEmail(ctx, "test@example.com")
	if !errors.Is(err, repositories.ErrUserNotFound) {
		t.Errorf("GetUserByEmail default: want ErrUserNotFound, got %v", err)
	}

	_, err = repo.GetUserByUsername(ctx, "someuser")
	if !errors.Is(err, repositories.ErrUserNotFound) {
		t.Errorf("GetUserByUsername default: want ErrUserNotFound, got %v", err)
	}
}

func TestMockUserRepository_GetUserByIDFnOverridesDefault(t *testing.T) {
	want := &models.User{Id: "uid-1", Email: "a@b.com"}
	repo := &testutil.MockUserRepository{
		GetUserByIDFn: func(_ context.Context, userID string) (*models.User, error) {
			if userID == "uid-1" {
				return want, nil
			}
			return nil, repositories.ErrUserNotFound
		},
	}

	got, err := repo.GetUserByID(context.Background(), "uid-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Id != want.Id {
		t.Errorf("Id: got %q, want %q", got.Id, want.Id)
	}

	_, err = repo.GetUserByID(context.Background(), "uid-unknown")
	if !errors.Is(err, repositories.ErrUserNotFound) {
		t.Errorf("unknown ID: want ErrUserNotFound, got %v", err)
	}
}

func TestMockUserRepository_GetUserByEmailFnOverridesDefault(t *testing.T) {
	want := &models.User{Id: "uid-1", Email: "a@b.com"}
	repo := &testutil.MockUserRepository{
		GetUserByEmailFn: func(_ context.Context, email string) (*models.User, error) {
			if email == "a@b.com" {
				return want, nil
			}
			return nil, repositories.ErrUserNotFound
		},
	}

	got, err := repo.GetUserByEmail(context.Background(), "a@b.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Email != want.Email {
		t.Errorf("Email: got %q, want %q", got.Email, want.Email)
	}

	_, err = repo.GetUserByEmail(context.Background(), "other@b.com")
	if !errors.Is(err, repositories.ErrUserNotFound) {
		t.Errorf("unknown email: want ErrUserNotFound, got %v", err)
	}
}

func TestMockUserRepository_GetUserByUsernameFnOverridesDefault(t *testing.T) {
	want := &models.User{Id: "uid-1", Email: "alice@example.com"}
	repo := &testutil.MockUserRepository{
		GetUserByUsernameFn: func(_ context.Context, username string) (*models.User, error) {
			if username == "alice" {
				return want, nil
			}
			return nil, repositories.ErrUserNotFound
		},
	}

	got, err := repo.GetUserByUsername(context.Background(), "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Id != want.Id {
		t.Errorf("Id: got %q, want %q", got.Id, want.Id)
	}

	_, err = repo.GetUserByUsername(context.Background(), "bob")
	if !errors.Is(err, repositories.ErrUserNotFound) {
		t.Errorf("unknown username: want ErrUserNotFound, got %v", err)
	}
}

func TestMockUserRepository_SearchUsersDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockUserRepository{}
	users, err := repo.SearchUsers(context.Background(), "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != 0 {
		t.Errorf("expected empty slice, got %d users", len(users))
	}
}

func TestMockUserRepository_SearchUsersFnReturnsResults(t *testing.T) {
	want := []models.User{{Id: "u1", Email: "alice@example.com"}, {Id: "u2", Email: "alicia@example.com"}}
	repo := &testutil.MockUserRepository{
		SearchUsersFn: func(_ context.Context, _ string) ([]models.User, error) {
			return want, nil
		},
	}

	got, err := repo.SearchUsers(context.Background(), "ali")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != len(want) {
		t.Errorf("expected %d users, got %d", len(want), len(got))
	}
}

func TestMockUserRepository_SearchUsersFnFiltersOnQuery(t *testing.T) {
	all := []models.User{
		{Id: "u1", Email: "alice@example.com"},
		{Id: "u2", Email: "bob@example.com"},
		{Id: "u3", Email: "alicia@example.com"},
	}
	repo := &testutil.MockUserRepository{
		SearchUsersFn: func(_ context.Context, query string) ([]models.User, error) {
			var out []models.User
			for _, u := range all {
				if len(u.Email) >= len(query) && u.Email[:len(query)] == query {
					out = append(out, u)
				}
			}
			return out, nil
		},
	}

	got, err := repo.SearchUsers(context.Background(), "ali")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 results for 'ali', got %d", len(got))
	}
}