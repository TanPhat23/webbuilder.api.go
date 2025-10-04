package models

import (
	"time"
)

type ContentItem struct {
	ContentTypeId string              `gorm:"column:ContentTypeId;type:varchar(255);not null" json:"contentTypeId"`
	CreatedAt     time.Time           `gorm:"column:CreatedAt" json:"createdAt,omitempty"`
	Id            string              `gorm:"primaryKey;column:Id;type:varchar(255)" json:"id"`
	Published     bool                `gorm:"column:Published;not null;default:false" json:"published"`
	Slug          string              `gorm:"column:Slug;type:varchar(255);unique;not null" json:"slug"`
	Title         string              `gorm:"column:Title;type:varchar(255);not null" json:"title"`
	UpdatedAt     time.Time           `gorm:"column:UpdatedAt" json:"updatedAt,omitempty"`
	FieldValues   []ContentFieldValue `gorm:"foreignKey:ContentItemId" json:"fieldValues,omitempty"`
	ContentType   ContentType         `gorm:"foreignKey:ContentTypeId;references:Id" json:"contentType,omitempty"`
}

func (ContentItem) TableName() string {
	return `public."ContentItem"`
}
