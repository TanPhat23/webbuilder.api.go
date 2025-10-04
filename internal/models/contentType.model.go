package models

import (
	"time"
)

type ContentType struct {
	CreatedAt   time.Time           `gorm:"column:CreatedAt" json:"createdAt,omitempty"`
	Description *string             `gorm:"column:Description;type:text" json:"description,omitempty"`
	Id          string              `gorm:"primaryKey;column:Id;type:varchar(255)" json:"id"`
	Name        string              `gorm:"column:Name;type:varchar(255);unique;not null" json:"name"`
	UpdatedAt   time.Time           `gorm:"column:UpdatedAt" json:"updatedAt,omitempty"`
	Fields      []ContentField      `gorm:"foreignKey:ContentTypeId" json:"fields,omitempty"`
	Items       []ContentItem       `gorm:"foreignKey:ContentTypeId" json:"items,omitempty"`
}

func (ContentType) TableName() string {
	return `public."ContentType"`
}
