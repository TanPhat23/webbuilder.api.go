package repositories

import (
	"context"
	"errors"

	"my-go-app/internal/models"

	"gorm.io/gorm"
)

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

	var user models.User
	err := r.db.WithContext(ctx).
		Where("\"Id\" = ?", userID).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if email == "" {
		return nil, errors.New("email is required")
	}

	var user models.User
	err := r.db.WithContext(ctx).
		Where("\"Email\" = ?", email).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	if username == "" {
		return nil, errors.New("username is required")
	}

	var user models.User
	err := r.db.WithContext(ctx).
		Where("\"Id\" = ?", username).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) SearchUsers(ctx context.Context, query string) ([]models.User, error) {
	if query == "" {
		return []models.User{}, nil
	}

	var users []models.User
	err := r.db.WithContext(ctx).
		Where("\"Email\" ILIKE ? OR \"FirstName\" ILIKE ? OR \"LastName\" ILIKE ? OR \"Id\" ILIKE ?",
			"%"+query+"%", "%"+query+"%", "%"+query+"%", "%"+query+"%").
		Limit(20). // Limit results for performance
		Find(&users).Error

	if err != nil {
		return nil, err
	}

	return users, nil
}
