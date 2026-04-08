package models

type ContentItem struct {
	FieldValues []ContentFieldValue `gorm:"foreignKey:ContentItemId" json:"fieldValues,omitempty"`
	ContentType *ContentType        `gorm:"foreignKey:ContentTypeId;references:Id" json:"contentType,omitempty"`

	Id            string `gorm:"primaryKey;column:Id;type:varchar(255)" json:"id"`
	ContentTypeId string `gorm:"column:ContentTypeId;type:varchar(255);not null" json:"contentTypeId"`
	Slug          string `gorm:"column:Slug;type:varchar(255);unique;not null" json:"slug"`
	Title         string `gorm:"column:Title;type:varchar(255);not null" json:"title"`

	Published bool `gorm:"column:Published;not null;default:false" json:"published"`

	AuditFields
}

func (ContentItem) TableName() string {
	return `public."ContentItem"`
}