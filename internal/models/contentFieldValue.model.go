package models

type ContentFieldValue struct {
	ContentItemId string      `gorm:"column:ContentItemId;type:varchar(255);not null;uniqueIndex:contentItemId_fieldId" json:"contentItemId"`
	FieldId       string      `gorm:"column:FieldId;type:varchar(255);not null;uniqueIndex:contentItemId_fieldId" json:"fieldId"`
	Id            string      `gorm:"primaryKey;column:Id;type:varchar(255)" json:"id"`
	Value         *string     `gorm:"column:Value;type:text" json:"value,omitempty"`
	ContentItem   ContentItem `gorm:"foreignKey:ContentItemId;references:Id" json:"contentItem,omitempty"`
	Field         ContentField `gorm:"foreignKey:FieldId;references:Id" json:"field,omitempty"`
}

func (ContentFieldValue) TableName() string {
	return `public."ContentFieldValue"`
}
