package models

import "time"

type Project struct {
	ID          string         `gorm:"primaryKey;column:Id;type:uuid" json:"id"`
	Name        string     `gorm:"column:Name;type:varchar(255);not null" json:"name"`
	Description *string    `gorm:"column:Description;type:text" json:"description,omitempty"`
	Styles      string     `gorm:"column:Styles;type:text" json:"styles"`
	CustomStyles *string    `gorm:"column:CustomStyles;type:text" json:"customStyles,omitempty"`
	Published   bool       `gorm:"column:Published;not null;default:false" json:"published"`
	Subdomain   *string    `gorm:"column:Subdomain;type:varchar(255)" json:"subdomain,omitempty"`
	OwnerId     string     `gorm:"column:OwnerId;type:varchar(255);not null" json:"ownerId"`
	CreatedAt   time.Time  `gorm:"column:CreatedAt" json:"createdAt,omitempty"`
	UpdatedAt   time.Time  `gorm:"column:UpdatedAt" json:"updatedAt,omitempty"`
	DeletedAt   *time.Time `gorm:"column:DeletedAt" json:"deletedAt,omitempty"`
}

// Table name is managed by repositories; removed GetTable method from model.
