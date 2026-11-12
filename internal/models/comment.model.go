package models

import (
	"time"
)

// Comment represents a comment on a marketplace item
type Comment struct {
	// Relations/Slices
	Replies   []Comment         `gorm:"foreignKey:ParentId;references:Id" json:"replies,omitempty"`
	Reactions []CommentReaction `gorm:"foreignKey:CommentId;references:Id" json:"reactions,omitempty"`
	Author    *User             `gorm:"foreignKey:AuthorId;references:Id" json:"author,omitempty"`
	Item      *MarketplaceItem  `gorm:"foreignKey:ItemId;references:Id" json:"item,omitempty"`
	Parent    *Comment          `gorm:"foreignKey:ParentId;references:Id" json:"parent,omitempty"`

	// Embedded Audit Fields (16-byte types: time.Time)
	AuditFields

	// Strings
	Id       string  `gorm:"primaryKey;column:Id;type:varchar(255)" json:"id"`
	AuthorId string  `gorm:"column:AuthorId;type:varchar(255);not null" json:"authorId"`
	ItemId   string  `gorm:"column:ItemId;type:varchar(255);not null" json:"itemId"`
	Status   string  `gorm:"column:Status;type:varchar(50);not null;default:'published'" json:"status"`
	Content  string  `gorm:"column:Content;type:text;not null" json:"content"`
	ParentId *string `gorm:"column:ParentId;type:varchar(255)" json:"parentId,omitempty"`

	// Bools
	Edited bool `gorm:"column:Edited;not null;default:false" json:"edited"`
}

func (Comment) TableName() string {
	return `public."Comment"`
}

// CommentReaction represents a user's reaction to a comment
type CommentReaction struct {
	// Relations/Pointers
	Comment *Comment `gorm:"foreignKey:CommentId;references:Id" json:"comment,omitempty"`
	User    *User    `gorm:"foreignKey:UserId;references:Id" json:"user,omitempty"`

	// 16-byte type (time.Time)
	CreatedAt time.Time `gorm:"column:CreatedAt" json:"createdAt"`

	// Strings
	Id        string `gorm:"primaryKey;column:Id;type:varchar(255)" json:"id"`
	CommentId string `gorm:"column:CommentId;type:varchar(255);not null" json:"commentId"`
	UserId    string `gorm:"column:UserId;type:varchar(255);not null" json:"userId"`
	Type      string `gorm:"column:Type;type:varchar(50);not null" json:"type"`
}

func (CommentReaction) TableName() string {
	return `public."CommentReaction"`
}

// Request/Response DTOs - Ordered by size for optimal alignment
type CreateCommentRequest struct {
	Content  string  `json:"content" validate:"required,min=1,max=5000"`
	ItemId   string  `json:"itemId" validate:"required,min=1"`
	ParentId *string `json:"parentId" validate:"omitempty,min=1"`
}

type UpdateCommentRequest struct {
	Content *string `json:"content" validate:"omitempty,min=1,max=5000"`
	Status  *string `json:"status" validate:"omitempty,oneof=published archived spam flagged"`
}

type CommentResponse struct {
	Id        string            `json:"id"`
	Content   string            `json:"content"`
	AuthorId  string            `json:"authorId"`
	ItemId    string            `json:"itemId"`
	ParentId  *string           `json:"parentId,omitempty"`
	Status    string            `json:"status"`
	Edited    bool              `json:"edited"`
	CreatedAt time.Time         `json:"createdAt"`
	UpdatedAt time.Time         `json:"updatedAt"`
	Author    *CommentAuthor    `json:"author,omitempty"`
	Replies   []CommentResponse `json:"replies,omitempty"`
	Reactions []ReactionSummary `json:"reactions,omitempty"`
}

type CommentAuthor struct {
	Id        string  `json:"id"`
	FirstName *string `json:"firstName,omitempty"`
	LastName  *string `json:"lastName,omitempty"`
	Email     string  `json:"email"`
	ImageUrl  *string `json:"imageUrl,omitempty"`
}

type ReactionSummary struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

type CreateReactionRequest struct {
	Type string `json:"type" validate:"required,min=1,max=50,oneof=like love haha wow sad angry"`
}

type CommentFilter struct {
	ItemId    string
	AuthorId  string
	Status    string
	ParentId  *string
	Limit     int
	Offset    int
	SortBy    string
	SortOrder string
}
