package models

import (
	"time"
)

type Image struct {
	ImageId   string     `gorm:"column:ImageId;primaryKey" json:"imageId"`
	ImageLink string     `gorm:"column:ImageLink;not null;default:''" json:"imageLink"`
	ImageName *string    `gorm:"column:ImageName" json:"imageName"`
	UserId    string     `gorm:"column:UserId;not null" json:"userId"`
	CreatedAt time.Time  `gorm:"column:CreatedAt;type:timestamp(6);not null" json:"createdAt"`
	DeletedAt *time.Time `gorm:"column:DeletedAt;type:timestamp(6)" json:"deletedAt,omitempty"`
	UpdatedAt time.Time  `gorm:"column:UpdatedAt;type:timestamp(6);not null" json:"updatedAt"`
}

func (Image) TableName() string {
	return "Image"
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
