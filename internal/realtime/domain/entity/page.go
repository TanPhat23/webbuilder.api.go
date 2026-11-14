package entity

import (
	"time"

	"gorm.io/gorm"
)

// Page represents a page within a project.
// This matches the Prisma Page model.
type Page struct {
	ID        string         `json:"id" gorm:"primaryKey;column:Id"`
	Name      string         `json:"name" gorm:"column:Name;type:varchar(255);uniqueIndex:ux_project_page_name"`
	Type      string         `json:"type" gorm:"column:Type;type:varchar(50);index"`
	Styles    map[string]any `json:"styles" gorm:"column:Styles;serializer:json;type:jsonb"`
	ProjectID string         `json:"projectId" gorm:"column:ProjectId;index;uniqueIndex:ux_project_page_name"`
	CreatedAt time.Time      `json:"createdAt" gorm:"column:CreatedAt"`
	UpdatedAt time.Time      `json:"updatedAt" gorm:"column:UpdatedAt"`
	DeletedAt gorm.DeletedAt `json:"deletedAt,omitempty" gorm:"column:DeletedAt;index"`

	Elements []Element `json:"elements,omitempty" gorm:"foreignKey:PageID;references:ID"`
	Project  *Project  `json:"project,omitempty" gorm:"foreignKey:ProjectID;references:ID"`
}

// TableName sets the database table name for Page.
func (Page) TableName() string {
	return "Page"
}

type DirtyPage struct {
	ProjectID string `json:"projectId"`
	PageID    string `json:"pageId"`
}

// User represents a user in the system.
// This matches the Prisma User model.
type User struct {
	ID        string    `json:"id" gorm:"primaryKey;column:Id"`
	Email     string    `json:"email" gorm:"column:Email;type:varchar(255)"`
	FirstName *string   `json:"firstName,omitempty" gorm:"column:FirstName;type:varchar(100)"`
	LastName  *string   `json:"lastName,omitempty" gorm:"column:LastName;type:varchar(100)"`
	ImageUrl  *string   `json:"imageUrl,omitempty" gorm:"column:ImageUrl;type:varchar(2048)"`
	CreatedAt time.Time `json:"createdAt" gorm:"column:CreatedAt;type:timestamptz(6)"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"column:UpdatedAt;type:timestamptz(6)"`
}

// TableName sets the database table name for User.
func (User) TableName() string {
	return "User"
}

// Placeholder types to satisfy relations in other entities.
// These are kept here for compilation until they are moved to their own files.

type Collaborator struct {
	ID        string `json:"id" gorm:"primaryKey;column:Id"`
	ProjectID string `json:"projectId" gorm:"column:ProjectId;index"`
	UserID    string `json:"userId" gorm:"column:UserId;index"`
}

func (Collaborator) TableName() string {
	return "Collaborator"
}

type Invitation struct {
	ID        string `json:"id" gorm:"primaryKey;column:Id"`
	ProjectID string `json:"projectId" gorm:"column:ProjectId;index"`
	Email     string `json:"email" gorm:"column:Email;type:varchar(255)"`
}

func (Invitation) TableName() string {
	return "Invitation"
}

type MarketplaceItem struct {
	ID        string `json:"id" gorm:"primaryKey;column:Id"`
	ProjectID string `json:"projectId,omitempty" gorm:"column:ProjectId;index"`
}

func (MarketplaceItem) TableName() string {
	return "MarketplaceItem"
}
