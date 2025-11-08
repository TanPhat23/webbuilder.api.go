package repositories

import (
	"fmt"
	"my-go-app/internal/models"
	"time"

	"gorm.io/gorm"
)

type CommentRepository struct {
	DB *gorm.DB
}

func NewCommentRepository(db *gorm.DB) *CommentRepository {
	return &CommentRepository{
		DB: db,
	}
}

// CreateComment creates a new comment
func (r *CommentRepository) CreateComment(comment models.Comment) (*models.Comment, error) {
	if err := r.DB.Create(&comment).Error; err != nil {
		return nil, err
	}

	return r.GetCommentByID(comment.Id)
}

// GetCommentByID retrieves a comment by its ID with author and reactions
func (r *CommentRepository) GetCommentByID(id string) (*models.Comment, error) {
	var comment models.Comment
	err := r.DB.Where(`"Id" = ? AND "DeletedAt" IS NULL`, id).
		First(&comment).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	// Load author information
	var author models.User
	if err := r.DB.Where(`"Id" = ?`, comment.AuthorId).First(&author).Error; err == nil {
		comment.Author = &author
	}

	// Load reactions
	var reactions []models.CommentReaction
	r.DB.Where(`"CommentId" = ?`, comment.Id).Find(&reactions)
	comment.Reactions = reactions

	return &comment, nil
}

// GetComments retrieves comments with filtering and pagination
func (r *CommentRepository) GetComments(filter models.CommentFilter) ([]models.Comment, int64, error) {
	var comments []models.Comment
	var total int64

	query := r.DB.Model(&models.Comment{}).
		Where(`"DeletedAt" IS NULL`)

	// Apply filters
	if filter.ItemId != "" {
		query = query.Where(`"ItemId" = ?`, filter.ItemId)
	}

	if filter.AuthorId != "" {
		query = query.Where(`"AuthorId" = ?`, filter.AuthorId)
	}

	if filter.Status != "" {
		query = query.Where(`"Status" = ?`, filter.Status)
	}

	// Filter by parent (top-level comments or replies)
	if filter.ParentId != nil {
		if *filter.ParentId == "" {
			query = query.Where(`"ParentId" IS NULL`)
		} else {
			query = query.Where(`"ParentId" = ?`, *filter.ParentId)
		}
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	sortBy := "CreatedAt"
	if filter.SortBy != "" {
		switch filter.SortBy {
		case "createdAt":
			sortBy = "CreatedAt"
		case "updatedAt":
			sortBy = "UpdatedAt"
		}
	}

	sortOrder := "DESC"
	if filter.SortOrder == "asc" {
		sortOrder = "ASC"
	}

	query = query.Order(`"` + sortBy + `" ` + sortOrder)

	// Apply pagination
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	// Execute query
	if err := query.Find(&comments).Error; err != nil {
		return nil, 0, err
	}

	// Load authors and reactions for all comments
	if len(comments) > 0 {
		commentIds := make([]string, len(comments))
		authorIds := make([]string, len(comments))
		for i, comment := range comments {
			commentIds[i] = comment.Id
			authorIds[i] = comment.AuthorId
		}

		// Load all authors
		var authors []models.User
		r.DB.Where(`"Id" IN ?`, authorIds).Find(&authors)
		authorMap := make(map[string]*models.User)
		for i := range authors {
			authorMap[authors[i].Id] = &authors[i]
		}

		// Load all reactions
		var reactions []models.CommentReaction
		r.DB.Where(`"CommentId" IN ?`, commentIds).Find(&reactions)
		reactionMap := make(map[string][]models.CommentReaction)
		for _, reaction := range reactions {
			reactionMap[reaction.CommentId] = append(reactionMap[reaction.CommentId], reaction)
		}

		// Load replies count
		type ReplyCount struct {
			ParentId string
			Count    int64
		}
		var replyCounts []ReplyCount
		r.DB.Model(&models.Comment{}).
			Select(`"ParentId", COUNT(*) as count`).
			Where(`"ParentId" IN ? AND "DeletedAt" IS NULL`, commentIds).
			Group(`"ParentId"`).
			Scan(&replyCounts)
		replyCountMap := make(map[string]int64)
		for _, rc := range replyCounts {
			replyCountMap[rc.ParentId] = rc.Count
		}

		// Assign to comments
		for i := range comments {
			if author, ok := authorMap[comments[i].AuthorId]; ok {
				comments[i].Author = author
			}
			if reactions, ok := reactionMap[comments[i].Id]; ok {
				comments[i].Reactions = reactions
			}
			// Load replies if requested
			if filter.ParentId == nil || (filter.ParentId != nil && *filter.ParentId == "") {
				// This is a top-level comment, optionally load replies
				var replies []models.Comment
				r.DB.Where(`"ParentId" = ? AND "DeletedAt" IS NULL`, comments[i].Id).
					Order(`"CreatedAt" ASC`).
					Find(&replies)

				// Load authors for replies
				for j := range replies {
					if author, ok := authorMap[replies[j].AuthorId]; ok {
						replies[j].Author = author
					}
					if reactions, ok := reactionMap[replies[j].Id]; ok {
						replies[j].Reactions = reactions
					}
				}
				comments[i].Replies = replies
			}
		}
	}

	return comments, total, nil
}

// UpdateComment updates a comment
func (r *CommentRepository) UpdateComment(id string, userId string, updates map[string]any) (*models.Comment, error) {
	// Verify ownership
	var count int64
	if err := r.DB.Model(&models.Comment{}).
		Where(`"Id" = ? AND "AuthorId" = ? AND "DeletedAt" IS NULL`, id, userId).
		Count(&count).Error; err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, fmt.Errorf("comment not found or unauthorized")
	}

	// Mark as edited if content is being updated
	if _, hasContent := updates["Content"]; hasContent {
		updates["Edited"] = true
	}

	updates["UpdatedAt"] = time.Now()

	if err := r.DB.Model(&models.Comment{}).
		Where(`"Id" = ?`, id).
		Updates(updates).Error; err != nil {
		return nil, err
	}

	return r.GetCommentByID(id)
}

// DeleteComment soft deletes a comment
func (r *CommentRepository) DeleteComment(id string, userId string) error {
	result := r.DB.Model(&models.Comment{}).
		Where(`"Id" = ? AND "AuthorId" = ? AND "DeletedAt" IS NULL`, id, userId).
		Update("DeletedAt", time.Now())
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// CreateReaction creates or updates a reaction
func (r *CommentRepository) CreateReaction(reaction models.CommentReaction) (*models.CommentReaction, error) {
	// Check if reaction already exists
	var existing models.CommentReaction
	err := r.DB.Where(`"CommentId" = ? AND "UserId" = ? AND "Type" = ?`,
		reaction.CommentId, reaction.UserId, reaction.Type).
		First(&existing).Error

	if err == nil {
		// Reaction already exists
		return &existing, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// Create new reaction
	if err := r.DB.Create(&reaction).Error; err != nil {
		return nil, err
	}

	return &reaction, nil
}

// DeleteReaction deletes a reaction
func (r *CommentRepository) DeleteReaction(commentId string, userId string, reactionType string) error {
	result := r.DB.Where(`"CommentId" = ? AND "UserId" = ? AND "Type" = ?`,
		commentId, userId, reactionType).
		Delete(&models.CommentReaction{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// GetReactionsByCommentID retrieves all reactions for a comment
func (r *CommentRepository) GetReactionsByCommentID(commentId string) ([]models.CommentReaction, error) {
	var reactions []models.CommentReaction
	if err := r.DB.Where(`"CommentId" = ?`, commentId).Find(&reactions).Error; err != nil {
		return nil, err
	}
	return reactions, nil
}

// GetReactionSummary retrieves reaction counts grouped by type
func (r *CommentRepository) GetReactionSummary(commentId string) ([]models.ReactionSummary, error) {
	var summary []models.ReactionSummary
	err := r.DB.Model(&models.CommentReaction{}).
		Select(`"Type" as type, COUNT(*) as count`).
		Where(`"CommentId" = ?`, commentId).
		Group(`"Type"`).
		Scan(&summary).Error
	if err != nil {
		return nil, err
	}
	return summary, nil
}

// GetCommentCountByItemID returns the number of comments for a marketplace item
func (r *CommentRepository) GetCommentCountByItemID(itemId string) (int64, error) {
	var count int64
	err := r.DB.Model(&models.Comment{}).
		Where(`"ItemId" = ? AND "DeletedAt" IS NULL`, itemId).
		Count(&count).Error
	return count, err
}

// ModerateComment updates the status of a comment (for admin/moderation)
func (r *CommentRepository) ModerateComment(id string, status string) error {
	result := r.DB.Model(&models.Comment{}).
		Where(`"Id" = ? AND "DeletedAt" IS NULL`, id).
		Update("Status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
