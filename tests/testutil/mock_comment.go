package testutil

import (
	"context"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
)

type MockCommentRepository struct {
	CreateCommentFn           func(ctx context.Context, comment models.Comment) (*models.Comment, error)
	GetCommentByIDFn          func(ctx context.Context, id string) (*models.Comment, error)
	GetCommentsFn             func(ctx context.Context, filter models.CommentFilter) ([]models.Comment, int64, error)
	UpdateCommentFn           func(ctx context.Context, id, userID string, updates map[string]any) (*models.Comment, error)
	DeleteCommentFn           func(ctx context.Context, id, userID string) error
	CreateReactionFn          func(ctx context.Context, reaction models.CommentReaction) (*models.CommentReaction, error)
	DeleteReactionFn          func(ctx context.Context, commentID, userID, reactionType string) error
	GetReactionsByCommentIDFn func(ctx context.Context, commentID string) ([]models.CommentReaction, error)
	GetReactionSummaryFn      func(ctx context.Context, commentID string) ([]models.ReactionSummary, error)
	GetCommentCountByItemIDFn func(ctx context.Context, itemID string) (int64, error)
	ModerateCommentFn         func(ctx context.Context, id, status string) error
}

func (m *MockCommentRepository) CreateComment(ctx context.Context, comment models.Comment) (*models.Comment, error) {
	if m.CreateCommentFn != nil {
		return m.CreateCommentFn(ctx, comment)
	}
	return &comment, nil
}

func (m *MockCommentRepository) GetCommentByID(ctx context.Context, id string) (*models.Comment, error) {
	if m.GetCommentByIDFn != nil {
		return m.GetCommentByIDFn(ctx, id)
	}
	return nil, repositories.ErrCommentNotFound
}

func (m *MockCommentRepository) GetComments(ctx context.Context, filter models.CommentFilter) ([]models.Comment, int64, error) {
	if m.GetCommentsFn != nil {
		return m.GetCommentsFn(ctx, filter)
	}
	return []models.Comment{}, 0, nil
}

func (m *MockCommentRepository) UpdateComment(ctx context.Context, id, userID string, updates map[string]any) (*models.Comment, error) {
	if m.UpdateCommentFn != nil {
		return m.UpdateCommentFn(ctx, id, userID, updates)
	}
	return nil, repositories.ErrCommentUnauthorized
}

func (m *MockCommentRepository) DeleteComment(ctx context.Context, id, userID string) error {
	if m.DeleteCommentFn != nil {
		return m.DeleteCommentFn(ctx, id, userID)
	}
	return nil
}

func (m *MockCommentRepository) CreateReaction(ctx context.Context, reaction models.CommentReaction) (*models.CommentReaction, error) {
	if m.CreateReactionFn != nil {
		return m.CreateReactionFn(ctx, reaction)
	}
	return &reaction, nil
}

func (m *MockCommentRepository) DeleteReaction(ctx context.Context, commentID, userID, reactionType string) error {
	if m.DeleteReactionFn != nil {
		return m.DeleteReactionFn(ctx, commentID, userID, reactionType)
	}
	return nil
}

func (m *MockCommentRepository) GetReactionsByCommentID(ctx context.Context, commentID string) ([]models.CommentReaction, error) {
	if m.GetReactionsByCommentIDFn != nil {
		return m.GetReactionsByCommentIDFn(ctx, commentID)
	}
	return []models.CommentReaction{}, nil
}

func (m *MockCommentRepository) GetReactionSummary(ctx context.Context, commentID string) ([]models.ReactionSummary, error) {
	if m.GetReactionSummaryFn != nil {
		return m.GetReactionSummaryFn(ctx, commentID)
	}
	return []models.ReactionSummary{}, nil
}

func (m *MockCommentRepository) GetCommentCountByItemID(ctx context.Context, itemID string) (int64, error) {
	if m.GetCommentCountByItemIDFn != nil {
		return m.GetCommentCountByItemIDFn(ctx, itemID)
	}
	return 0, nil
}

func (m *MockCommentRepository) ModerateComment(ctx context.Context, id, status string) error {
	if m.ModerateCommentFn != nil {
		return m.ModerateCommentFn(ctx, id, status)
	}
	return nil
}