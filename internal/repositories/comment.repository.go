package repositories

import (
	"context"
	"errors"
	"fmt"
	"my-go-app/internal/models"
	"time"

	"gorm.io/gorm"
)

var (
	ErrCommentNotFound     = errors.New("comment not found")
	ErrCommentUnauthorized = errors.New("unauthorized to modify comment")
	ErrReactionNotFound    = errors.New("reaction not found")
)

type CommentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) CommentRepositoryInterface {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) CreateComment(ctx context.Context, comment models.Comment) (*models.Comment, error) {
	if err := gorm.G[models.Comment](r.db).Create(ctx, &comment); err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}
	return r.GetCommentByID(ctx, comment.Id)
}

func (r *CommentRepository) GetCommentByID(ctx context.Context, id string) (*models.Comment, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}

	comment, err := gorm.G[models.Comment](r.db).
		Where(`"Id" = ? AND "DeletedAt" IS NULL`, id).
		First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCommentNotFound
		}
		return nil, fmt.Errorf("failed to get comment: %w", err)
	}

	// Load author
	if author, err := gorm.G[models.User](r.db).Where(`"Id" = ?`, comment.AuthorId).First(ctx); err == nil {
		comment.Author = &author
	}

	// Load reactions
	reactions, _ := gorm.G[models.CommentReaction](r.db).Where(`"CommentId" = ?`, comment.Id).Find(ctx)
	comment.Reactions = reactions

	return &comment, nil
}

func (r *CommentRepository) GetComments(ctx context.Context, filter models.CommentFilter) ([]models.Comment, int64, error) {
	query := gorm.G[models.Comment](r.db).
		Where(`"DeletedAt" IS NULL`)

	if filter.ItemId != "" {
		query = query.Where(`"ItemId" = ?`, filter.ItemId)
	}
	if filter.AuthorId != "" {
		query = query.Where(`"AuthorId" = ?`, filter.AuthorId)
	}
	if filter.Status != "" {
		query = query.Where(`"Status" = ?`, filter.Status)
	}
	if filter.ParentId != nil {
		if *filter.ParentId == "" {
			query = query.Where(`"ParentId" IS NULL`)
		} else {
			query = query.Where(`"ParentId" = ?`, *filter.ParentId)
		}
	}

	total, err := query.Count(ctx, "*")
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count comments: %w", err)
	}

	sortBy := "CreatedAt"
	if filter.SortBy == "updatedAt" {
		sortBy = "UpdatedAt"
	}
	sortOrder := "DESC"
	if filter.SortOrder == "asc" {
		sortOrder = "ASC"
	}
	query = query.Order(`"` + sortBy + `" ` + sortOrder)

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	comments, err := query.Find(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get comments: %w", err)
	}

	if len(comments) == 0 {
		return comments, total, nil
	}

	// Bulk-load authors and reactions.
	commentIDs := make([]string, len(comments))
	authorIDs := make([]string, len(comments))
	for i, c := range comments {
		commentIDs[i] = c.Id
		authorIDs[i] = c.AuthorId
	}

	authors, _ := gorm.G[models.User](r.db).Where(`"Id" IN ?`, authorIDs).Find(ctx)
	authorMap := make(map[string]*models.User, len(authors))
	for i := range authors {
		authorMap[authors[i].Id] = &authors[i]
	}

	reactions, _ := gorm.G[models.CommentReaction](r.db).Where(`"CommentId" IN ?`, commentIDs).Find(ctx)
	reactionMap := make(map[string][]models.CommentReaction)
	for _, rxn := range reactions {
		reactionMap[rxn.CommentId] = append(reactionMap[rxn.CommentId], rxn)
	}

	for i := range comments {
		if a, ok := authorMap[comments[i].AuthorId]; ok {
			comments[i].Author = a
		}
		if r2, ok := reactionMap[comments[i].Id]; ok {
			comments[i].Reactions = r2
		}

		// Load replies for top-level comments.
		if filter.ParentId == nil || *filter.ParentId == "" {
			replies, _ := gorm.G[models.Comment](r.db).
				Where(`"ParentId" = ? AND "DeletedAt" IS NULL`, comments[i].Id).
				Order(`"CreatedAt" ASC`).
				Find(ctx)
			for j := range replies {
				if a, ok := authorMap[replies[j].AuthorId]; ok {
					replies[j].Author = a
				}
				if r2, ok := reactionMap[replies[j].Id]; ok {
					replies[j].Reactions = r2
				}
			}
			comments[i].Replies = replies
		}
	}

	return comments, total, nil
}

func (r *CommentRepository) UpdateComment(ctx context.Context, id string, userID string, updates map[string]any) (*models.Comment, error) {
	if id == "" || userID == "" {
		return nil, errors.New("id and userID are required")
	}

	// First, ensure the comment exists and is not deleted.
	if _, err := r.GetCommentByID(ctx, id); err != nil {
		if errors.Is(err, ErrCommentNotFound) {
			return nil, ErrCommentNotFound
		}
		return nil, fmt.Errorf("failed to verify comment existence: %w", err)
	}

	count, err := gorm.G[models.Comment](r.db).
		Where(`"Id" = ? AND "AuthorId" = ? AND "DeletedAt" IS NULL`, id, userID).
		Count(ctx, "*")
	if err != nil {
		return nil, fmt.Errorf("failed to verify comment ownership: %w", err)
	}
	if count == 0 {
		return nil, ErrCommentUnauthorized
	}

	if _, hasContent := updates["Content"]; hasContent {
		updates["Edited"] = true
	}
	updates["UpdatedAt"] = time.Now()

	if err := r.db.WithContext(ctx).
		Model(&models.Comment{}).
		Where(`"Id" = ?`, id).
		Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update comment: %w", err)
	}

	return r.GetCommentByID(ctx, id)
}

func (r *CommentRepository) DeleteComment(ctx context.Context, id string, userID string) error {
	if id == "" || userID == "" {
		return errors.New("id and userID are required")
	}

	rows, err := gorm.G[models.Comment](r.db).
		Where(`"Id" = ? AND "AuthorId" = ? AND "DeletedAt" IS NULL`, id, userID).
		Update(ctx, `"DeletedAt"`, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}
	if rows == 0 {
		return ErrCommentUnauthorized
	}
	return nil
}

func (r *CommentRepository) CreateReaction(ctx context.Context, reaction models.CommentReaction) (*models.CommentReaction, error) {
	existing, err := gorm.G[models.CommentReaction](r.db).
		Where(`"CommentId" = ? AND "UserId" = ? AND "Type" = ?`,
			reaction.CommentId, reaction.UserId, reaction.Type).
		First(ctx)

	if err == nil {
		return &existing, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check reaction: %w", err)
	}

	if err := gorm.G[models.CommentReaction](r.db).Create(ctx, &reaction); err != nil {
		return nil, fmt.Errorf("failed to create reaction: %w", err)
	}
	return &reaction, nil
}

func (r *CommentRepository) DeleteReaction(ctx context.Context, commentID string, userID string, reactionType string) error {
	rows, err := gorm.G[models.CommentReaction](r.db).
		Where(`"CommentId" = ? AND "UserId" = ? AND "Type" = ?`, commentID, userID, reactionType).
		Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete reaction: %w", err)
	}
	if rows == 0 {
		return ErrReactionNotFound
	}
	return nil
}

func (r *CommentRepository) GetReactionsByCommentID(ctx context.Context, commentID string) ([]models.CommentReaction, error) {
	if commentID == "" {
		return nil, errors.New("commentID is required")
	}
	reactions, err := gorm.G[models.CommentReaction](r.db).Where(`"CommentId" = ?`, commentID).Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get reactions: %w", err)
	}
	return reactions, nil
}

func (r *CommentRepository) GetReactionSummary(ctx context.Context, commentID string) ([]models.ReactionSummary, error) {
	if commentID == "" {
		return nil, errors.New("commentID is required")
	}
	var summary []models.ReactionSummary
	err := r.db.WithContext(ctx).
		Model(&models.CommentReaction{}).
		Select(`"Type" as type, COUNT(*) as count`).
		Where(`"CommentId" = ?`, commentID).
		Group(`"Type"`).
		Scan(&summary).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get reaction summary: %w", err)
	}
	return summary, nil
}

func (r *CommentRepository) GetCommentCountByItemID(ctx context.Context, itemID string) (int64, error) {
	if itemID == "" {
		return 0, errors.New("itemID is required")
	}
	count, err := gorm.G[models.Comment](r.db).
		Where(`"ItemId" = ? AND "DeletedAt" IS NULL`, itemID).
		Count(ctx, "*")
	if err != nil {
		return 0, fmt.Errorf("failed to count comments: %w", err)
	}
	return count, nil
}

func (r *CommentRepository) ModerateComment(ctx context.Context, id string, status string) error {
	if id == "" {
		return errors.New("id is required")
	}
	rows, err := gorm.G[models.Comment](r.db).
		Where(`"Id" = ? AND "DeletedAt" IS NULL`, id).
		Update(ctx, "Status", status)
	if err != nil {
		return fmt.Errorf("failed to moderate comment: %w", err)
	}
	if rows == 0 {
		return ErrCommentNotFound
	}
	return nil
}
