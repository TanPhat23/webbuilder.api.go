package models

type ContentField struct {
	Values      []ContentFieldValue `gorm:"foreignKey:FieldId" json:"values,omitempty"`
	ContentType *ContentType        `gorm:"foreignKey:ContentTypeId;references:Id" json:"contentType,omitempty"`

	Id            string `gorm:"primaryKey;column:Id;type:varchar(255)" json:"id"`
	ContentTypeId string `gorm:"column:ContentTypeId;type:varchar(255);not null;uniqueIndex:contentTypeId_name" json:"contentTypeId"`
	Name          string `gorm:"column:Name;type:varchar(255);not null;uniqueIndex:contentTypeId_name" json:"name"`
	Type          string `gorm:"column:Type;type:varchar(255);not null" json:"type"`

	Required bool `gorm:"column:Required;not null;default:false" json:"required"`
}

func (ContentField) TableName() string {
	return `public."ContentField"`
}