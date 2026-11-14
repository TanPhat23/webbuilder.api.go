package repositories

import (
	"context"
	"errors"
	"fmt"

	"my-go-app/internal/models"

	"gorm.io/gorm"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepositoryInterface {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	if userID == "" {
		return nil, errors.New("userID is required")
	}

	user, err := gorm.G[models.User](r.db).
		Where(`"Id" = ?`, userID).
		First(ctx)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if email == "" {
		return nil, errors.New("email is required")
	}

	user, err := gorm.G[models.User](r.db).
		Where(`"Email" = ?`, email).
		First(ctx)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

// GetUserByUsername searches by FirstName or LastName, returning the user whose
// first name or last name matches the provided value.
func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	if username == "" {
		return nil, errors.New("username is required")
	}

	user, err := gorm.G[models.User](r.db).
		Where(`"FirstName" ILIKE ? OR "LastName" ILIKE ?`, "%"+username+"%", "%"+username+"%").
		First(ctx)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) SearchUsers(ctx context.Context, query string) ([]models.User, error) {
	if query == "" {
		return []models.User{}, nil
	}

	users, err := gorm.G[models.User](r.db).
		Select(`"Id", "Email", "FirstName", "LastName", "ImageUrl", "CreatedAt", "UpdatedAt"`).
		Where(
			`"Email" ILIKE ? OR "FirstName" ILIKE ? OR "LastName" ILIKE ? OR "Id" ILIKE ?`,
			query+"%", "%"+query+"%", "%"+query+"%", "%"+query+"%",
		).
		Limit(20).
		Find(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	return users, nil
}
