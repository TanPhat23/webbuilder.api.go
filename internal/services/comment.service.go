package services

import (
	"context"
	"errors"
	"fmt"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
)

type CommentService struct {
	commentRepo     repositories.CommentRepositoryInterface
	marketplaceRepo repositories.MarketplaceRepositoryInterface
}

func NewCommentService(
	commentRepo repositories.CommentRepositoryInterface,
	marketplaceRepo repositories.MarketplaceRepositoryInterface,
) *CommentService {
	return &CommentService{
		commentRepo:     commentRepo,
		marketplaceRepo: marketplaceRepo,
	}
}

func (s *CommentService) CreateComment(ctx context.Context, comment *models.Comment) (*models.Comment, error) {
	if comment == nil {
		return nil, errors.New("comment cannot be nil")
	}
	if comment.ItemId == "" {
		return nil, errors.New("itemId is required")
	}
	if comment.AuthorId == "" {
		return nil, errors.New("authorId is required")
	}
	if comment.Content == "" {
		return nil, errors.New("comment content cannot be empty")
	}

	item, err := s.marketplaceRepo.GetMarketplaceItemByID(comment.ItemId)
	if err != nil {
		return nil, fmt.Errorf("marketplace item not found: %w", err)
	}
	if item == nil {
		return nil, errors.New("marketplace item does not exist")
	}

	if comment.Status == "" {
		comment.Status = "published"
	}

	return s.commentRepo.CreateComment(ctx, *comment)
}

func (s *CommentService) GetCommentByID(ctx context.Context, id string) (*models.Comment, error) {
	if id == "" {
		return nil, errors.New("comment id is required")
	}

	return s.commentRepo.GetCommentByID(ctx, id)
}

func (s *CommentService) GetComments(ctx context.Context, itemID string) ([]models.Comment, error) {
	if itemID == "" {
		return nil, errors.New("itemId is required")
	}

	filter := models.CommentFilter{
		ItemId: itemID,
		Status: "published",
	}

	comments, _, err := s.commentRepo.GetComments(ctx, filter)
	return comments, err
}

func (s *CommentService) GetCommentsByItemID(ctx context.Context, itemID string, filter models.CommentFilter) ([]models.Comment, int64, error) {
	if itemID == "" {
		return nil, 0, errors.New("itemId is required")
	}

	filter.ItemId = itemID
	if filter.Status == "" {
		filter.Status = "published"
	}
	return s.commentRepo.GetComments(ctx, filter)
}

func (s *CommentService) UpdateComment(ctx context.Context, id string, userID string, updates map[string]any) (*models.Comment, error) {
	if id == "" {
		return nil, errors.New("comment id is required")
	}
	if userID == "" {
		return nil, errors.New("userID is required")
	}
	if len(updates) == 0 {
		return nil, errors.New("no updates provided")
	}

	comment, err := s.GetCommentByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve comment: %w", err)
	}
	if comment == nil {
		return nil, errors.New("comment does not exist")
	}

	if comment.AuthorId != userID {
		return nil, errors.New("unauthorized: user is not the comment author")
	}

	if content, ok := updates["Content"]; ok {
		if value, ok := content.(string); ok && value == "" {
			return nil, errors.New("content cannot be empty")
		}
	}

	return s.commentRepo.UpdateComment(ctx, id, userID, updates)
}

func (s *CommentService) DeleteComment(ctx context.Context, id string, userID string) error {
	if id == "" {
		return errors.New("comment id is required")
	}
	if userID == "" {
		return errors.New("userID is required")
	}

	comment, err := s.GetCommentByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to retrieve comment: %w", err)
	}
	if comment == nil {
		return errors.New("comment does not exist")
	}

	if comment.AuthorId != userID {
		return errors.New("unauthorized: user is not the comment author")
	}

	return s.commentRepo.DeleteComment(ctx, id, userID)
}

func (s *CommentService) CreateReaction(ctx context.Context, reaction *models.CommentReaction) (*models.CommentReaction, error) {
	if reaction == nil {
		return nil, errors.New("reaction cannot be nil")
	}
	if reaction.CommentId == "" {
		return nil, errors.New("commentId is required")
	}
	if reaction.UserId == "" {
		return nil, errors.New("userId is required")
	}
	if reaction.Type == "" {
		return nil, errors.New("reaction type is required")
	}

	if _, err := s.GetCommentByID(ctx, reaction.CommentId); err != nil {
		return nil, fmt.Errorf("failed to verify comment: %w", err)
	}

	return s.commentRepo.CreateReaction(ctx, *reaction)
}

func (s *CommentService) DeleteReaction(ctx context.Context, commentID string, userID string, reactionType string) error {
	if commentID == "" {
		return errors.New("commentId is required")
	}
	if userID == "" {
		return errors.New("userId is required")
	}
	if reactionType == "" {
		return errors.New("reaction type is required")
	}

	if _, err := s.GetCommentByID(ctx, commentID); err != nil {
		return fmt.Errorf("failed to verify comment: %w", err)
	}

	return s.commentRepo.DeleteReaction(ctx, commentID, userID, reactionType)
}

func (s *CommentService) GetReactionsByCommentID(ctx context.Context, commentID string) ([]models.CommentReaction, error) {
	if commentID == "" {
		return nil, errors.New("commentId is required")
	}

	if _, err := s.GetCommentByID(ctx, commentID); err != nil {
		return nil, fmt.Errorf("failed to verify comment: %w", err)
	}

	return s.commentRepo.GetReactionsByCommentID(ctx, commentID)
}

func (s *CommentService) GetReactionSummary(ctx context.Context, commentID string) (map[string]int, error) {
	if commentID == "" {
		return nil, errors.New("commentId is required")
	}

	if _, err := s.GetCommentByID(ctx, commentID); err != nil {
		return nil, fmt.Errorf("failed to verify comment: %w", err)
	}

	summary, err := s.commentRepo.GetReactionSummary(ctx, commentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reaction summary: %w", err)
	}

	result := make(map[string]int)
	for _, item := range summary {
		result[item.Type] = item.Count
	}
	return result, nil
}

func (s *CommentService) GetCommentCountByItemID(ctx context.Context, itemID string) (int64, error) {
	if itemID == "" {
		return 0, errors.New("itemId is required")
	}

	return s.commentRepo.GetCommentCountByItemID(ctx, itemID)
}

func (s *CommentService) ModerateComment(ctx context.Context, id string, status string) error {
	if id == "" {
		return errors.New("comment id is required")
	}
	if status == "" {
		return errors.New("status is required")
	}

	validStatuses := map[string]bool{
		"pending":  true,
		"published": true,
		"flagged":   true,
		"deleted":   true,
	}
	if !validStatuses[status] {
		return fmt.Errorf("invalid status: %s", status)
	}

	comment, err := s.GetCommentByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to retrieve comment: %w", err)
	}
	if comment == nil {
		return errors.New("comment does not exist")
	}

	return s.commentRepo.ModerateComment(ctx, id, status)
}