package services

import (
	"context"
	"errors"
	"testing"

	"my-go-app/internal/models"
	"my-go-app/internal/services"
	test "my-go-app/test/mockrepo"
)

func TestGetUserByID_Success(t *testing.T) {
	mock := test.NewMockUserRepo()
	mock.SetGetUserByID(func(ctx context.Context, id string) (*models.User, error) {
		return &models.User{Id: id, Email: "test@example.com"}, nil
	})
	service := services.NewUserService(mock)

	user, err := service.GetUserByID(context.Background(), "123")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if user == nil || user.Id != "123" {
		t.Errorf("expected user with ID 123, got %v", user)
	}
}

func TestGetUserByID_EmptyID(t *testing.T) {
	mock := test.NewMockUserRepo()
	service := services.NewUserService(mock)

	user, err := service.GetUserByID(context.Background(), "")
	if err == nil {
		t.Errorf("expected error for empty ID")
	}
	if user != nil {
		t.Errorf("expected nil user, got %v", user)
	}
	if err.Error() != "userId is required" {
		t.Errorf("expected 'userId is required', got %v", err.Error())
	}
}

func TestGetUserByID_UserNotFound(t *testing.T) {
	mock := test.NewMockUserRepo()
	mock.SetGetUserByID(func(ctx context.Context, id string) (*models.User, error) {
		return nil, nil
	})
	service := services.NewUserService(mock)

	user, err := service.GetUserByID(context.Background(), "123")
	if err == nil {
		t.Errorf("expected error for user not found")
	}
	if user != nil {
		t.Errorf("expected nil user, got %v", user)
	}
	if err.Error() != "user does not exist" {
		t.Errorf("expected 'user does not exist', got %v", err.Error())
	}
}

func TestGetUserByID_RepoError(t *testing.T) {
	mock := test.NewMockUserRepo()
	mock.SetGetUserByID(func(ctx context.Context, id string) (*models.User, error) {
		return nil, errors.New("repo error")
	})
	service := services.NewUserService(mock)

	user, err := service.GetUserByID(context.Background(), "123")
	if err == nil {
		t.Errorf("expected repo error")
	}
	if user != nil {
		t.Errorf("expected nil user, got %v", user)
	}
	if err.Error() != "repo error" {
		t.Errorf("expected 'repo error', got %v", err.Error())
	}
}

func TestGetUserByEmail_Success(t *testing.T) {
	mock := test.NewMockUserRepo()
	mock.SetGetUserByEmail(func(ctx context.Context, email string) (*models.User, error) {
		return &models.User{Id: "123", Email: email}, nil
	})
	service := services.NewUserService(mock)

	user, err := service.GetUserByEmail(context.Background(), "test@example.com")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if user == nil || user.Email != "test@example.com" {
		t.Errorf("expected user with email test@example.com, got %v", user)
	}
}

func TestGetUserByEmail_EmptyEmail(t *testing.T) {
	mock := test.NewMockUserRepo()
	service := services.NewUserService(mock)

	user, err := service.GetUserByEmail(context.Background(), "")
	if err == nil {
		t.Errorf("expected error for empty email")
	}
	if user != nil {
		t.Errorf("expected nil user, got %v", user)
	}
	if err.Error() != "email is required" {
		t.Errorf("expected 'email is required', got %v", err.Error())
	}
}

func TestGetUserByEmail_UserNotFound(t *testing.T) {
	mock := test.NewMockUserRepo()
	mock.SetGetUserByEmail(func(ctx context.Context, email string) (*models.User, error) {
		return nil, nil
	})
	service := services.NewUserService(mock)

	user, err := service.GetUserByEmail(context.Background(), "test@example.com")
	if err == nil {
		t.Errorf("expected error for user not found")
	}
	if user != nil {
		t.Errorf("expected nil user, got %v", user)
	}
	if err.Error() != "user does not exist" {
		t.Errorf("expected 'user does not exist', got %v", err.Error())
	}
}

func TestGetUserByEmail_RepoError(t *testing.T) {
	mock := test.NewMockUserRepo()
	mock.SetGetUserByEmail(func(ctx context.Context, email string) (*models.User, error) {
		return nil, errors.New("repo error")
	})
	service := services.NewUserService(mock)

	user, err := service.GetUserByEmail(context.Background(), "test@example.com")
	if err == nil {
		t.Errorf("expected repo error")
	}
	if user != nil {
		t.Errorf("expected nil user, got %v", user)
	}
	if err.Error() != "repo error" {
		t.Errorf("expected 'repo error', got %v", err.Error())
	}
}

func TestGetUserByUsername_Success(t *testing.T) {
	mock := test.NewMockUserRepo()
	mock.SetGetUserByUsername(func(ctx context.Context, username string) (*models.User, error) {
		return &models.User{Id: "123", Email: "test@example.com"}, nil
	})
	service := services.NewUserService(mock)

	user, err := service.GetUserByUsername(context.Background(), "testuser")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if user == nil {
		t.Errorf("expected user, got nil")
	}
}

func TestGetUserByUsername_EmptyUsername(t *testing.T) {
	mock := test.NewMockUserRepo()
	service := services.NewUserService(mock)

	user, err := service.GetUserByUsername(context.Background(), "")
	if err == nil {
		t.Errorf("expected error for empty username")
	}
	if user != nil {
		t.Errorf("expected nil user, got %v", user)
	}
	if err.Error() != "username is required" {
		t.Errorf("expected 'username is required', got %v", err.Error())
	}
}

func TestGetUserByUsername_UserNotFound(t *testing.T) {
	mock := test.NewMockUserRepo()
	mock.SetGetUserByUsername(func(ctx context.Context, username string) (*models.User, error) {
		return nil, nil
	})
	service := services.NewUserService(mock)

	user, err := service.GetUserByUsername(context.Background(), "testuser")
	if err == nil {
		t.Errorf("expected error for user not found")
	}
	if user != nil {
		t.Errorf("expected nil user, got %v", user)
	}
	if err.Error() != "user does not exist" {
		t.Errorf("expected 'user does not exist', got %v", err.Error())
	}
}

func TestGetUserByUsername_RepoError(t *testing.T) {
	mock := test.NewMockUserRepo()
	mock.SetGetUserByUsername(func(ctx context.Context, username string) (*models.User, error) {
		return nil, errors.New("repo error")
	})
	service := services.NewUserService(mock)

	user, err := service.GetUserByUsername(context.Background(), "testuser")
	if err == nil {
		t.Errorf("expected repo error")
	}
	if user != nil {
		t.Errorf("expected nil user, got %v", user)
	}
	if err.Error() != "repo error" {
		t.Errorf("expected 'repo error', got %v", err.Error())
	}
}

func TestGetUserByID_WhitespaceID(t *testing.T) {
	mock := test.NewMockUserRepo()
	service := services.NewUserService(mock)

	user, err := service.GetUserByID(context.Background(), "   ")
	if err == nil {
		t.Errorf("expected error for whitespace-only ID")
	}
	if user != nil {
		t.Errorf("expected nil user, got %v", user)
	}
	if err.Error() != "userId is required" {
		t.Errorf("expected 'userId is required', got %v", err.Error())
	}
}

func TestGetUserByEmail_WhitespaceEmail(t *testing.T) {
	mock := test.NewMockUserRepo()
	service := services.NewUserService(mock)

	user, err := service.GetUserByEmail(context.Background(), "   \t\n  ")
	if err == nil {
		t.Errorf("expected error for whitespace-only email")
	}
	if user != nil {
		t.Errorf("expected nil user, got %v", user)
	}
	if err.Error() != "email is required" {
		t.Errorf("expected 'email is required', got %v", err.Error())
	}
}

func TestGetUserByUsername_WhitespaceUsername(t *testing.T) {
	mock := test.NewMockUserRepo()
	service := services.NewUserService(mock)

	user, err := service.GetUserByUsername(context.Background(), "  \n  ")
	if err == nil {
		t.Errorf("expected error for whitespace-only username")
	}
	if user != nil {
		t.Errorf("expected nil user, got %v", user)
	}
	if err.Error() != "username is required" {
		t.Errorf("expected 'username is required', got %v", err.Error())
	}
}

func TestSearchUsers_Success(t *testing.T) {
	users := []models.User{
		{Id: "1", Email: "john@example.com"},
		{Id: "2", Email: "jane@example.com"},
	}
	mock := test.NewMockUserRepo()
	mock.SetSearchUsers(func(ctx context.Context, query string) ([]models.User, error) {
		return users, nil
	})
	service := services.NewUserService(mock)

	result, err := service.SearchUsers(context.Background(), "john")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 users, got %d", len(result))
	}
}

func TestSearchUsers_EmptyQuery(t *testing.T) {
	mock := test.NewMockUserRepo()
	service := services.NewUserService(mock)

	result, err := service.SearchUsers(context.Background(), "")
	if err == nil {
		t.Errorf("expected error for empty query")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
	if err.Error() != "query is required" {
		t.Errorf("expected 'query is required', got %v", err.Error())
	}
}

func TestSearchUsers_WhitespaceQuery(t *testing.T) {
	mock := test.NewMockUserRepo()
	service := services.NewUserService(mock)

	result, err := service.SearchUsers(context.Background(), "   \t  ")
	if err == nil {
		t.Errorf("expected error for whitespace-only query")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
	if err.Error() != "query is required" {
		t.Errorf("expected 'query is required', got %v", err.Error())
	}
}

func TestSearchUsers_RepositoryError(t *testing.T) {
	mock := test.NewMockUserRepo()
	mock.SetSearchUsers(func(ctx context.Context, query string) ([]models.User, error) {
		return nil, errors.New("repo error")
	})
	service := services.NewUserService(mock)

	result, err := service.SearchUsers(context.Background(), "john")
	if err == nil {
		t.Errorf("expected repo error")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
	if err.Error() != "repo error" {
		t.Errorf("expected 'repo error', got %v", err.Error())
	}
}

func TestSearchUsers_EmptyResult(t *testing.T) {
	mock := test.NewMockUserRepo()
	mock.SetSearchUsers(func(ctx context.Context, query string) ([]models.User, error) {
		return []models.User{}, nil
	})
	service := services.NewUserService(mock)

	result, err := service.SearchUsers(context.Background(), "nonexistent")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 users, got %d", len(result))
	}
}
