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
