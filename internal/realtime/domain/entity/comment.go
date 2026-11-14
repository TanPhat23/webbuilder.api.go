package entity

import (
	"time"

	"gorm.io/gorm"
)

// ElementComment represents a comment left on a specific UI element in the builder.
// This matches the Prisma ElementComment model.
type ElementComment struct {
	ID        string         `json:"id" gorm:"primaryKey;column:Id"`
	Content   string         `json:"content" gorm:"column:Content;type:text"`
	AuthorID  string         `json:"authorId" gorm:"column:AuthorId;index"`
	ElementID string         `json:"elementId" gorm:"column:ElementId;index;index:idx_element_resolved"`
	Resolved  bool           `json:"resolved" gorm:"column:Resolved;default:false;index;index:idx_element_resolved"`
	CreatedAt time.Time      `json:"createdAt" gorm:"column:CreatedAt"`
	UpdatedAt time.Time      `json:"updatedAt" gorm:"column:UpdatedAt"`
	DeletedAt gorm.DeletedAt `json:"deletedAt,omitempty" gorm:"column:DeletedAt;index"`

	Author  *User    `json:"author,omitempty" gorm:"foreignKey:AuthorID;references:ID"`
	Element *Element `json:"element,omitempty" gorm:"foreignKey:ElementID;references:ID"`
}

// TableName sets the database table name for ElementComment.
func (ElementComment) TableName() string {
	return "ElementComment"
}

// Comment represents a general comment system (e.g., for Marketplace items).
// This matches the Prisma Comment model.
type Comment struct {
	ID        string         `json:"id" gorm:"primaryKey;column:id"`
	Content   string         `json:"content" gorm:"column:content;type:text"`
	AuthorID  string         `json:"authorId" gorm:"column:author_id;index"`
	ItemID    string         `json:"itemId" gorm:"column:item_id;index"`
	ParentID  *string        `json:"parentId,omitempty" gorm:"column:parent_id;index"`
	Status    string         `json:"status" gorm:"column:status;type:varchar(50);default:'published';index"`
	Edited    bool           `json:"edited" gorm:"column:edited;default:false"`
	CreatedAt time.Time      `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt time.Time      `json:"updatedAt" gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `json:"deletedAt,omitempty" gorm:"column:deleted_at;index"`

	Author  *User     `json:"author,omitempty" gorm:"foreignKey:AuthorID;references:ID"`
	Parent  *Comment  `json:"parent,omitempty" gorm:"foreignKey:ParentID;references:ID"`
	Replies []Comment `json:"replies,omitempty" gorm:"foreignKey:ParentID"`
}

// TableName sets the database table name for Comment.
func (Comment) TableName() string {
	return "Comment"
}
