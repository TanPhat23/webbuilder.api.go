package models

import (
	"time"
)

type Image struct {
	ImageId   string     `gorm:"primaryKey;column:ImageId;type:varchar(255)" json:"imageId"`
	ImageLink string     `gorm:"column:ImageLink;type:text;not null;default:''" json:"imageLink"`
	ImageName *string    `gorm:"column:ImageName;type:varchar(255)" json:"imageName,omitempty"`
	UserId    string     `gorm:"column:UserId;type:varchar(255);not null" json:"userId"`
	CreatedAt time.Time  `gorm:"column:CreatedAt" json:"createdAt,omitempty"`
	DeletedAt *time.Time `gorm:"column:DeletedAt" json:"deletedAt,omitempty"`
	UpdatedAt time.Time  `gorm:"column:UpdatedAt" json:"updatedAt,omitempty"`
}

func (Image) TableName() string {
	return `public."Image"`
}

type CreateImageRequest struct {
	ImageName *string `json:"imageName"`
}

type ImageUploadResponse struct {
	ImageId   string    `json:"imageId"`
	ImageLink string    `json:"imageLink"`
	ImageName *string   `json:"imageName"`
	CreatedAt time.Time `json:"createdAt"`
}
