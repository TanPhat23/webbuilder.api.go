package models

import (
	"time"
)

type User struct {
	Id             string           `gorm:"primaryKey;column:Id;type:varchar(255)" json:"id"`
	Email          string           `gorm:"column:Email;type:varchar(255);not null" json:"email"`
	FirstName      *string          `gorm:"column:FirstName;type:varchar(255)" json:"firstName,omitempty"`
	LastName       *string          `gorm:"column:LastName;type:varchar(255)" json:"lastName,omitempty"`
	ImageUrl       *string          `gorm:"column:ImageUrl;type:text" json:"imageUrl,omitempty"`
	CreatedAt      time.Time        `gorm:"column:CreatedAt" json:"createdAt,omitempty"`
	UpdatedAt      time.Time        `gorm:"column:UpdatedAt" json:"updatedAt,omitempty"`
	Images         []Image          `gorm:"foreignKey:UserId" json:"images,omitempty"`
	Projects       []Project        `gorm:"foreignKey:OwnerId" json:"projects,omitempty"`
	MarketplaceItems []MarketplaceItem `gorm:"foreignKey:AuthorId" json:"marketplaceItems,omitempty"`
	Subscriptions  []Subscription   `gorm:"foreignKey:UserId" json:"subscriptions,omitempty"`
	Collaborators  []Collaborator   `gorm:"foreignKey:UserId" json:"collaborators,omitempty"`
	CustomElements []CustomElement  `gorm:"foreignKey:UserId" json:"customElements,omitempty"`
}

func (User) TableName() string {
	return `public."User"`
}
