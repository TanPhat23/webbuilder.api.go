package repositories

import (
	"context"
	"errors"
	"fmt"
	"my-go-app/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ElementCommentRepository struct {
	db *gorm.DB
}

func NewElementCommentRepository(db *gorm.DB) ElementCommentRepositoryInterface {
	return &ElementCommentRepository{
		db: db,
	}
}

// CreateElementComment creates a new element comment
func (r *ElementCommentRepository) CreateElementComment(ctx context.Context, comment *models.ElementComment) (*models.ElementComment, error) {
	if comment == nil {
		return nil, errors.New("comment cannot be nil")
	}

	if comment.Content == "" {
		return nil, errors.New("comment content cannot be empty")
	}

	if comment.ElementId == "" {
		return nil, errors.New("elementId is required")
	}

	if comment.AuthorId == "" {
		return nil, errors.New("authorId is required")
	}

	// Generate ID if not provided
	if comment.Id == "" {
		comment.Id = uuid.NewString()
	}

	if err := r.db.WithContext(ctx).Create(comment).Error; err != nil {
		return nil, fmt.Errorf("failed to create element comment: %w", err)
	}

	// Fetch the created comment with relations
	return r.GetElementCommentByID(ctx, comment.Id)
}

// GetElementCommentByID retrieves a single element comment by ID
func (r *ElementCommentRepository) GetElementCommentByID(ctx context.Context, id string) (*models.ElementComment, error) {
	if id == "" {
		return nil, errors.New("comment id is required")
	}

	var comment models.ElementComment

	err := r.db.WithContext(ctx).
		Preload("Author").
		Preload("Element").
		Where("\"Id\" = ? AND \"DeletedAt\" IS NULL", id).
		First(&comment).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("element comment not found")
		}
		return nil, fmt.Errorf("failed to get element comment: %w", err)
	}

	return &comment, nil
}

// GetElementComments retrieves comments for an element with filtering and pagination
func (r *ElementCommentRepository) GetElementComments(ctx context.Context, elementID string, filter *models.ElementCommentFilter) ([]models.ElementComment, error) {
	if elementID == "" {
		return nil, errors.New("elementId is required")
	}

	var comments []models.ElementComment

	query := r.db.WithContext(ctx).
		Preload("Author").
		Where("\"ElementId\" = ? AND \"DeletedAt\" IS NULL", elementID)

	// Apply additional filters
	if filter != nil {
		if filter.AuthorId != "" {
			query = query.Where("\"AuthorId\" = ?", filter.AuthorId)
		}

		if filter.Resolved != nil {
			query = query.Where("\"Resolved\" = ?", *filter.Resolved)
		}

		// Sorting
		if filter.SortBy != "" {
			order := "ASC"
			if filter.SortOrder == "DESC" {
				order = "DESC"
			}
			query = query.Order(fmt.Sprintf("\"%s\" %s", filter.SortBy, order))
		} else {
			query = query.Order("\"CreatedAt\" DESC")
		}

		// Pagination
		if filter.Limit > 0 {
			query = query.Limit(filter.Limit)
		}

		if filter.Offset > 0 {
			query = query.Offset(filter.Offset)
		}
	} else {
		query = query.Order("\"CreatedAt\" DESC")
	}

	if err := query.Find(&comments).Error; err != nil {
		return nil, fmt.Errorf("failed to get element comments: %w", err)
	}

	return comments, nil
}

// UpdateElementComment updates an existing element comment
func (r *ElementCommentRepository) UpdateElementComment(ctx context.Context, id string, updates map[string]interface{}) (*models.ElementComment, error) {
	if id == "" {
		return nil, errors.New("comment id is required")
	}

	if len(updates) == 0 {
		return nil, errors.New("no updates provided")
	}

	// Verify comment exists
	comment, err := r.GetElementCommentByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update the comment
	if err := r.db.WithContext(ctx).
		Model(comment).
		Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update element comment: %w", err)
	}

	// Fetch updated comment with relations
	return r.GetElementCommentByID(ctx, id)
}

// DeleteElementComment soft deletes an element comment
func (r *ElementCommentRepository) DeleteElementComment(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("comment id is required")
	}

	// Verify comment exists
	_, err := r.GetElementCommentByID(ctx, id)
	if err != nil {
		return err
	}

	if err := r.db.WithContext(ctx).
		Model(&models.ElementComment{}).
		Where("\"Id\" = ?", id).
		Update("\"DeletedAt\"", gorm.Expr("CURRENT_TIMESTAMP")).Error; err != nil {
		return fmt.Errorf("failed to delete element comment: %w", err)
	}

	return nil
}

// GetElementCommentsByAuthorID retrieves all comments by a specific author
func (r *ElementCommentRepository) GetElementCommentsByAuthorID(ctx context.Context, authorID string, limit int, offset int) ([]models.ElementComment, error) {
	if authorID == "" {
		return nil, errors.New("authorId is required")
	}

	var comments []models.ElementComment

	query := r.db.WithContext(ctx).
		Preload("Author").
		Preload("Element").
		Where("\"AuthorId\" = ? AND \"DeletedAt\" IS NULL", authorID).
		Order("\"CreatedAt\" DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&comments).Error; err != nil {
		return nil, fmt.Errorf("failed to get comments by author: %w", err)
	}

	return comments, nil
}

// CountElementComments counts comments for an element
func (r *ElementCommentRepository) CountElementComments(ctx context.Context, elementID string) (int64, error) {
	if elementID == "" {
		return 0, errors.New("elementId is required")
	}

	var count int64

	if err := r.db.WithContext(ctx).
		Model(&models.ElementComment{}).
		Where("\"ElementId\" = ? AND \"DeletedAt\" IS NULL", elementID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count element comments: %w", err)
	}

	return count, nil
}

// ToggleResolvedStatus toggles the resolved status of a comment
func (r *ElementCommentRepository) ToggleResolvedStatus(ctx context.Context, id string) (*models.ElementComment, error) {
	if id == "" {
		return nil, errors.New("comment id is required")
	}

	comment, err := r.GetElementCommentByID(ctx, id)
	if err != nil {
		return nil, err
	}

	newStatus := !comment.Resolved

	if err := r.db.WithContext(ctx).
		Model(comment).
		Update("\"Resolved\"", newStatus).Error; err != nil {
		return nil, fmt.Errorf("failed to toggle resolved status: %w", err)
	}

	return r.GetElementCommentByID(ctx, id)
}

// DeleteElementCommentsByElementID deletes all comments for an element (cascade delete)
func (r *ElementCommentRepository) DeleteElementCommentsByElementID(ctx context.Context, elementID string) error {
	if elementID == "" {
		return errors.New("elementId is required")
	}

	if err := r.db.WithContext(ctx).
		Model(&models.ElementComment{}).
		Where("\"ElementId\" = ?", elementID).
		Update("\"DeletedAt\"", gorm.Expr("CURRENT_TIMESTAMP")).Error; err != nil {
		return fmt.Errorf("failed to delete comments for element: %w", err)
	}

	return nil
}
