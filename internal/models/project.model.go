package models

import (
	"encoding/json"
)

type Project struct {
	Pages           []Page           `gorm:"foreignKey:ProjectId;references:Id;constraint:OnDelete:Cascade" json:"pages,omitempty"`
	Snapshots       []Snapshot       `gorm:"foreignKey:ProjectId;references:Id;constraint:OnDelete:Cascade" json:"snapshots,omitempty"`
	Collaborators   []Collaborator   `gorm:"foreignKey:ProjectId;references:Id;constraint:OnDelete:Cascade" json:"collaborators,omitempty"`
	Invitations     []Invitation     `gorm:"foreignKey:ProjectId;references:Id;constraint:OnDelete:Cascade" json:"invitations,omitempty"`

	ID      string `gorm:"primaryKey;column:Id;type:varchar(255)" json:"id"`
	Name    string `gorm:"column:Name;type:varchar(255);not null" json:"name"`
	OwnerId string `gorm:"column:OwnerId;type:varchar(255);not null" json:"ownerId"`

	Owner           User             `gorm:"foreignKey:OwnerId" json:"owner,omitempty"`
	MarketplaceItem *MarketplaceItem `gorm:"foreignKey:ProjectId" json:"marketplaceItem,omitempty"`
	Styles          *json.RawMessage `gorm:"column:Styles;type:jsonb" json:"styles,omitempty"`
	Header          *json.RawMessage `gorm:"column:Header;type:jsonb" json:"header,omitempty"`
	Description     *string          `gorm:"column:Description;type:text" json:"description,omitempty"`
	Subdomain       *string          `gorm:"column:Subdomain;type:varchar(255)" json:"subdomain,omitempty"`

	Published bool `gorm:"column:Published;not null;default:false" json:"published"`

	*AuditFields
}

func (Project) TableName() string {
	return `public."Project"`
}