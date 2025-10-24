package models

import (
	"encoding/json"
	"time"
)

type CustomElementType struct {
	Id          string    `gorm:"primaryKey;column:Id;type:varchar(255)" json:"id"`
	Name        string    `gorm:"column:Name;type:varchar(255);not null;uniqueIndex" json:"name"`
	Description *string   `gorm:"column:Description;type:text" json:"description,omitempty"`
	Category    *string   `gorm:"column:Category;type:varchar(100)" json:"category,omitempty"`
	Icon        *string   `gorm:"column:Icon;type:varchar(255)" json:"icon,omitempty"`
	CreatedAt   time.Time `gorm:"column:CreatedAt;type:timestamp;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"column:UpdatedAt;type:timestamp;default:CURRENT_TIMESTAMP" json:"updatedAt"`
}

func (CustomElementType) TableName() string {
	return `public."CustomElementType"`
}

type CustomElement struct {
	Id          string          `gorm:"primaryKey;column:Id;type:varchar(255)" json:"id"`
	Name        string          `gorm:"column:Name;type:varchar(255);not null" json:"name"`
	TypeId      *string         `gorm:"column:TypeId;type:varchar(255);index" json:"typeId,omitempty"`
	Description *string         `gorm:"column:Description;type:text" json:"description,omitempty"`
	Category    *string         `gorm:"column:Category;type:varchar(100)" json:"category,omitempty"`
	Icon        *string         `gorm:"column:Icon;type:varchar(255)" json:"icon,omitempty"`
	Thumbnail   *string         `gorm:"column:Thumbnail;type:varchar(255)" json:"thumbnail,omitempty"`
	Structure   json.RawMessage `gorm:"column:Structure;type:jsonb;not null" json:"structure"`
	DefaultProps json.RawMessage `gorm:"column:DefaultProps;type:jsonb" json:"defaultProps,omitempty"`
	Tags        *string         `gorm:"column:Tags;type:varchar(500)" json:"tags,omitempty"`
	UserId      string          `gorm:"column:UserId;type:varchar(255);not null;index" json:"userId"`
	IsPublic    bool            `gorm:"column:IsPublic;type:boolean;default:false" json:"isPublic"`
	Version     string          `gorm:"column:Version;type:varchar(20);default:'1.0.0'" json:"version"`
	CreatedAt   time.Time       `gorm:"column:CreatedAt;type:timestamp;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt   time.Time       `gorm:"column:UpdatedAt;type:timestamp;default:CURRENT_TIMESTAMP" json:"updatedAt"`
}

func (CustomElement) TableName() string {
	return `public."CustomElement"`
}
