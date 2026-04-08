package models

import (
	"time"
)

type ContentType struct {
	Fields []ContentField `gorm:"foreignKey:ContentTypeId" json:"fields,omitempty"`
	Items  []ContentItem  `gorm:"foreignKey:ContentTypeId" json:"items,omitempty"`

	Id          string  `gorm:"primaryKey;column:Id;type:varchar(255)" json:"id"`
	Name        string  `gorm:"column:Name;type:varchar(255);unique;not null" json:"name"`
	Description *string `gorm:"column:Description;type:text" json:"description,omitempty"`

	CreatedAt time.Time `gorm:"column:CreatedAt" json:"-"`
	UpdatedAt time.Time `gorm:"column:UpdatedAt" json:"-"`
}

func (ContentType) TableName() string {
	return `public."ContentType"`
}