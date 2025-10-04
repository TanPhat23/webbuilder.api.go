package models

type ContentField struct {
	ContentTypeId string              `gorm:"column:ContentTypeId;type:varchar(255);not null;uniqueIndex:contentTypeId_name" json:"contentTypeId"`
	Id            string              `gorm:"primaryKey;column:Id;type:varchar(255)" json:"id"`
	Name          string              `gorm:"column:Name;type:varchar(255);not null;uniqueIndex:contentTypeId_name" json:"name"`
	Required      bool                `gorm:"column:Required;not null;default:false" json:"required"`
	Type          string              `gorm:"column:Type;type:varchar(255);not null" json:"type"`
	ContentType   ContentType         `gorm:"foreignKey:ContentTypeId;references:Id" json:"contentType,omitempty"`
	Values        []ContentFieldValue `gorm:"foreignKey:FieldId" json:"values,omitempty"`
}

func (ContentField) TableName() string {
	return `public."ContentField"`
}
