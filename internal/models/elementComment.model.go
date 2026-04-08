package models

type ElementComment struct {
	Author  *User    `gorm:"foreignKey:AuthorId;references:Id" json:"author,omitempty"`
	Element *Element `gorm:"foreignKey:ElementId;references:Id" json:"element,omitempty"`

	AuditFields

	Id        string `gorm:"primaryKey;column:Id;type:varchar(255)" json:"id"`
	Content   string `gorm:"column:Content;type:text;not null" json:"content"`
	AuthorId  string `gorm:"column:AuthorId;type:varchar(255);not null" json:"authorId"`
	ElementId string `gorm:"column:ElementId;type:varchar(255);not null" json:"elementId"`
	Resolved  bool   `gorm:"column:Resolved;not null;default:false" json:"resolved"`
}

func (ElementComment) TableName() string {
	return `public."ElementComment"`
}

type CreateElementCommentRequest struct {
	Content   string `json:"content" validate:"required"`
	ElementId string `json:"elementId" validate:"required"`
}

type UpdateElementCommentRequest struct {
	Content  *string `json:"content"`
	Resolved *bool   `json:"resolved"`
}

type ElementCommentResponse struct {
	Id        string         `json:"id"`
	Content   string         `json:"content"`
	AuthorId  string         `json:"authorId"`
	ElementId string         `json:"elementId"`
	CreatedAt string         `json:"createdAt"`
	UpdatedAt string         `json:"updatedAt"`
	Resolved  bool           `json:"resolved"`
	Author    *CommentAuthor `json:"author,omitempty"`
}

type ElementCommentFilter struct {
	ElementId string
	AuthorId  string
	Resolved  *bool
	Limit     int
	Offset    int
	SortBy    string
	SortOrder string
}