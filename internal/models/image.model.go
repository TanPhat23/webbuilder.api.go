package models

import (
	"time"
)

type Image struct {
	AuditFields
	ImageId   string  `gorm:"primaryKey;column:ImageId;type:varchar(255)" json:"image_id"`
	ImageLink string  `gorm:"column:ImageLink;type:text;not null;default:''" json:"image_link"`
	UserId    string  `gorm:"column:UserId;type:varchar(255);not null" json:"user_id"`
	ImageName *string `gorm:"column:ImageName;type:varchar(255)" json:"image_name,omitempty"`
}

func (Image) TableName() string {
	return `public."Image"`
}

type CreateImageRequest struct {
	ImageName *string `json:"image_name"`
}

type ImageUploadResponse struct {
	ImageId   string    `json:"image_id"`
	ImageLink string    `json:"image_link"`
	ImageName *string   `json:"image_name"`
	CreatedAt time.Time `json:"created_at"`
}