package models

import "encoding/json"

type Element struct {
	Id             string          `gorm:"primaryKey;column:Id;type:varchar(255)" json:"id"`
	Type           string          `gorm:"column:Type;type:varchar(32)" json:"type"`
	Content        *string         `gorm:"column:Content;type:text" json:"content,omitempty"`
	Name           *string         `gorm:"column:Name;type:varchar(255)" json:"name,omitempty"`
	Styles         json.RawMessage `gorm:"column:Styles;type:jsonb" json:"styles,omitempty"`
	TailwindStyles *string         `gorm:"column:TailwindStyles;type:varchar(255)" json:"tailwindStyles,omitempty"`
	Src            *string         `gorm:"column:Src;type:varchar(255)" json:"src,omitempty"`
	Href           *string         `gorm:"column:Href;type:varchar(255)" json:"href,omitempty"`
	ParentId       *string         `gorm:"column:ParentId;type:varchar(255)" json:"parentId,omitempty"`
	PageId         *string         `gorm:"column:PageId;type:varchar(255)" json:"pageId,omitempty"`
	ProjectId      string          `gorm:"column:ProjectId;type:varchar(255)" json:"projectId"`
	Order          int             `gorm:"column:Order;default:0" json:"order"`
	Settings       *json.RawMessage `gorm:"-" json:"settings,omitempty"`

	// Relations
	Comments []ElementComment `gorm:"foreignKey:ElementId;references:Id" json:"comments,omitempty"`
}

type EditorElement interface {
	GetElement() *Element
	GetType() string
}

func (e *Element) GetElement() *Element {
	return e
}

func (e *Element) GetType() string {
	return e.Type
}

type SectionElement struct {
	*Element
	Elements []any `json:"elements" db:"-"`
}

type FrameElement struct {
	*Element
	Elements []any `json:"elements" db:"-"`
}

type CarouselElement struct {
	*Element
	Elements []any `json:"elements" db:"-"`
}

type ButtonElement struct {
	*Element
}

type InputElement struct {
	*Element
}

type ListElement struct {
	*Element
	Elements []any `json:"elements" db:"-"`
}

type SelectElement struct {
	*Element
	Elements []any `json:"elements" db:"-"`
}

type FormElement struct {
	*Element
	Elements []any `json:"elements" db:"-"`
}

type TextElement struct {
	*Element
}

type DataLoaderElement struct {
	*Element
	Elements []any `json:"elements" db:"-"`
}

type CMSContentListElement struct {
	*Element
	Elements []any `json:"elements" db:"-"`
}

type CMSContentItemElement struct {
	*Element
	Elements []any `json:"elements" db:"-"`
}

type CMSContentGridElement struct {
	*Element
	Elements []any `json:"elements" db:"-"`
}

func (Element) TableName() string {
	return `public."Element"`
}
