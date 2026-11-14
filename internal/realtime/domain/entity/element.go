package entity

type Element struct {
	ID             string         `json:"id" gorm:"primaryKey;column:Id" validate:"omitempty,uuid"`
	Name           *string        `json:"name,omitempty" gorm:"column:Name;type:varchar(255)"`
	Type           string         `json:"type" gorm:"column:Type;type:varchar(32)" validate:"required"`
	Content        *string        `json:"content,omitempty" gorm:"column:Content;type:varchar(2048)"`
	Href           *string        `json:"href,omitempty" gorm:"column:Href;type:varchar(2048)"`
	Src            *string        `json:"src,omitempty" gorm:"column:Src;type:varchar(2048)"`
	Styles         map[string]any `json:"styles" gorm:"column:Styles;serializer:json;type:jsonb"`
	Settings       map[string]any `json:"settings,omitempty" gorm:"column:Settings;serializer:json;type:jsonb"`
	TailwindStyles *string        `json:"tailwindStyles,omitempty" gorm:"column:TailwindStyles;type:text"`
	Order          int            `json:"order" gorm:"column:Order;default:0;index:idx_page_order,priority:2" validate:"gte=0"`
	OrderKey       string         `json:"orderKey,omitempty" gorm:"column:OrderKey"`
	IsLocked       bool           `json:"isLocked" gorm:"column:IsLocked;default:false"`
	ParentID       *string        `json:"parentId,omitempty" gorm:"column:ParentId;index" validate:"omitempty,uuid"`
	PageID         string         `json:"pageId" gorm:"column:PageId;index;index:idx_page_order,priority:1" validate:"required"`

	// Relations
	Page           *Page                  `json:"page,omitempty" gorm:"foreignKey:PageID;references:ID"`
	Parent         *Element               `json:"parent,omitempty" gorm:"foreignKey:ParentID;references:ID"`
	Elements       []Element              `json:"elements" gorm:"foreignKey:ParentID;references:ID"`
	Comments       []ElementComment       `json:"comments,omitempty" gorm:"foreignKey:ElementID"`
	EventWorkflows []ElementEventWorkflow `json:"eventWorkflows,omitempty" gorm:"foreignKey:ElementID"`
}

func (Element) TableName() string {
	return "Element"
}
