package models

import (
	"encoding/json"
	"time"
)

type Project struct {
	ID          string           `gorm:"primaryKey;column:Id;type:varchar(255)" json:"id"`
	Name        string           `gorm:"column:Name;type:varchar(255);not null" json:"name"`
	Description *string          `gorm:"column:Description;type:text" json:"description,omitempty"`
	Styles      *json.RawMessage `gorm:"column:Styles;type:jsonb" json:"styles,omitempty"`
	Header      *json.RawMessage `gorm:"column:Header;type:jsonb" json:"header,omitempty"`
	Published   bool             `gorm:"column:Published;not null;default:false" json:"published"`
	Subdomain   *string          `gorm:"column:Subdomain;type:varchar(255)" json:"subdomain,omitempty"`
	OwnerId     string           `gorm:"column:OwnerId;type:varchar(255);not null" json:"ownerId"`
	CreatedAt   time.Time        `gorm:"column:CreatedAt" json:"createdAt,omitempty"`
	UpdatedAt   time.Time        `gorm:"column:UpdatedAt" json:"updatedAt,omitempty"`
	DeletedAt   *time.Time       `gorm:"column:DeletedAt" json:"deletedAt,omitempty"`

	// Relations
	Elements        []Element        `gorm:"foreignKey:ProjectId;references:Id;constraint:OnDelete:Cascade" json:"elements,omitempty"`
	Pages           []Page           `gorm:"foreignKey:ProjectId;references:Id;constraint:OnDelete:Cascade" json:"pages,omitempty"`
	Owner           User             `gorm:"foreignKey:OwnerId" json:"owner,omitempty"`
	Snapshots       []Snapshot       `gorm:"foreignKey:ProjectId;references:Id;constraint:OnDelete:Cascade" json:"snapshots,omitempty"`
	MarketplaceItem *MarketplaceItem `gorm:"foreignKey:ProjectId" json:"marketplaceItem,omitempty"`
	Collaborators   []Collaborator   `gorm:"foreignKey:ProjectId;references:Id;constraint:OnDelete:Cascade" json:"collaborators,omitempty"`
	Invitations     []Invitation     `gorm:"foreignKey:ProjectId;references:Id;constraint:OnDelete:Cascade" json:"invitations,omitempty"`
}

func (Project) TableName() string {
	return `public."Project"`
}
