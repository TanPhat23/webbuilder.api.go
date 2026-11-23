package services

import (
	"context"
	"errors"
	"testing"

	"my-go-app/internal/models"
)

func TestUserService_Lookups(t *testing.T) {
	ctx := context.Background()
	u := &models.User{}

	type lookup struct {
		name   string
		hook   func(r *fkUserRepo, f func(string) (*models.User, error))
		call   func(s *UserService, in string) (*models.User, error)
		method string
		reqMsg string
	}
	lookups := []lookup{
		{"ByID", func(r *fkUserRepo, f func(string) (*models.User, error)) {
			r.OnGetUserByID = func(_ context.Context, v string) (*models.User, error) { return f(v) }
		}, func(s *UserService, in string) (*models.User, error) { return s.GetUserByID(ctx, in) }, "GetUserByID", "userId is required"},
		{"ByEmail", func(r *fkUserRepo, f func(string) (*models.User, error)) {
			r.OnGetUserByEmail = func(_ context.Context, v string) (*models.User, error) { return f(v) }
		}, func(s *UserService, in string) (*models.User, error) { return s.GetUserByEmail(ctx, in) }, "GetUserByEmail", "email is required"},
		{"ByUsername", func(r *fkUserRepo, f func(string) (*models.User, error)) {
			r.OnGetUserByUsername = func(_ context.Context, v string) (*models.User, error) { return f(v) }
		}, func(s *UserService, in string) (*models.User, error) { return s.GetUserByUsername(ctx, in) }, "GetUserByUsername", "username is required"},
	}

	for _, l := range lookups {
		t.Run(l.name, func(t *testing.T) {
			t.Run("blank input rejected without repo call", func(t *testing.T) {
				repo := &fkUserRepo{}
				_, err := l.call(NewUserService(repo), "   ")
				wantErr(t, err, l.reqMsg)
				wantCalls(t, repo.Calls)
			})
			t.Run("found", func(t *testing.T) {
				repo := &fkUserRepo{}
				var got string
				l.hook(repo, func(v string) (*models.User, error) { got = v; return u, nil })
				res, err := l.call(NewUserService(repo), "x")
				wantNoErr(t, err)
				if res != u || got != "x" {
					t.Fatalf("res=%v got=%q", res, got)
				}
				wantCalls(t, repo.Calls, l.method)
			})
			t.Run("repo error propagated", func(t *testing.T) {
				repo := &fkUserRepo{}
				l.hook(repo, func(string) (*models.User, error) { return nil, errBoom })
				_, err := l.call(NewUserService(repo), "x")
				if !errors.Is(err, errBoom) {
					t.Fatalf("err = %v", err)
				}
			})
			t.Run("nil user", func(t *testing.T) {
				repo := &fkUserRepo{}
				l.hook(repo, func(string) (*models.User, error) { return nil, nil })
				res, err := l.call(NewUserService(repo), "x")
				wantErr(t, err, "user does not exist")
				if res != nil {
					t.Fatalf("res = %v", res)
				}
			})
		})
	}
}

func TestUserService_SearchUsers(t *testing.T) {
	ctx := context.Background()

	t.Run("blank query", func(t *testing.T) {
		repo := &fkUserRepo{}
		_, err := NewUserService(repo).SearchUsers(ctx, " ")
		wantErr(t, err, "query is required")
		wantCalls(t, repo.Calls)
	})
	t.Run("results forwarded", func(t *testing.T) {
		repo := &fkUserRepo{OnSearchUsers: func(_ context.Context, q string) ([]models.User, error) {
			if q != "al" {
				t.Errorf("query = %q", q)
			}
			return []models.User{{}, {}}, nil
		}}
		res, err := NewUserService(repo).SearchUsers(ctx, "al")
		wantNoErr(t, err)
		if len(res) != 2 {
			t.Fatalf("len = %d", len(res))
		}
	})
	t.Run("repo error", func(t *testing.T) {
		repo := &fkUserRepo{OnSearchUsers: func(context.Context, string) ([]models.User, error) { return nil, errBoom }}
		_, err := NewUserService(repo).SearchUsers(ctx, "al")
		if !errors.Is(err, errBoom) {
			t.Fatalf("err = %v", err)
		}
	})
}
