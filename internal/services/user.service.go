package services

import (
	"context"
	"errors"
	"strings"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
)

type UserService struct {
	userRepo repositories.UserRepositoryInterface
}

func NewUserService(userRepo repositories.UserRepositoryInterface) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (s *UserService) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, errors.New("userId is required")
	}

	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user does not exist")
	}

	return user, nil
}

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if strings.TrimSpace(email) == "" {
		return nil, errors.New("email is required")
	}

	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user does not exist")
	}

	return user, nil
}

func (s *UserService) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	if strings.TrimSpace(username) == "" {
		return nil, errors.New("username is required")
	}

	user, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user does not exist")
	}

	return user, nil
}

func (s *UserService) SearchUsers(ctx context.Context, query string) ([]models.User, error) {
	if strings.TrimSpace(query) == "" {
		return nil, errors.New("query is required")
	}

	return s.userRepo.SearchUsers(ctx, query)
}