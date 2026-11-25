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

func (s *UserService) GetByID(ctx context.Context, userID string) (*models.User, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, errors.New("userId is required")
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user does not exist")
	}

	return user, nil
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	if strings.TrimSpace(email) == "" {
		return nil, errors.New("email is required")
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user does not exist")
	}

	return user, nil
}

func (s *UserService) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	if strings.TrimSpace(username) == "" {
		return nil, errors.New("username is required")
	}

	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user does not exist")
	}

	return user, nil
}

func (s *UserService) Search(ctx context.Context, query string) ([]models.User, error) {
	if strings.TrimSpace(query) == "" {
		return nil, errors.New("query is required")
	}

	return s.userRepo.Search(ctx, query)
}
