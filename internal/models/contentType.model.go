package models

import (
	"time"
)

type ContentType struct {
	Fields  []ContentField `gorm:"foreignKey:ContentTypeId" json:"fields,omitempty"`
	Items   []ContentItem  `gorm:"foreignKey:ContentTypeId" json:"items,omitempty"`
	Project *Project       `gorm:"foreignKey:ProjectId;references:ID" json:"-"`

	Id          string  `gorm:"primaryKey;column:Id;type:varchar(255)" json:"id"`
	ProjectId   *string `gorm:"column:ProjectId;type:varchar(255);index" json:"projectId,omitempty"`
	Name        string  `gorm:"column:Name;type:varchar(255);not null" json:"name"`
	Description *string `gorm:"column:Description;type:text" json:"description,omitempty"`

	CreatedAt time.Time `gorm:"column:CreatedAt" json:"-"`
	UpdatedAt time.Time `gorm:"column:UpdatedAt" json:"-"`
}

func (ContentType) TableName() string {
	return `public."ContentType"`
}