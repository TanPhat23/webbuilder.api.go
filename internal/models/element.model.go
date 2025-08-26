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
	PageId					*string 			`gorm:"column:PageId;type:varchar(255)" json:"pageId,omitempty"`
	ProjectId      string          `gorm:"column:ProjectId;type:varchar(255)" json:"projectId"`
	Order          int             `gorm:"column:Order;default:0" json:"order"`
	Settings       *json.RawMessage  `gorm:"-" json:"settings,omitempty"`
}

type SectionElement struct {
    Element
    Elements []any `json:"elements" db:"-"`
}

type FrameElement struct {
    Element
    Elements []any `json:"elements" db:"-"`
}

type CarouselElement struct {
    Element
    Elements []any `json:"elements" db:"-"`
}

type ButtonElement struct {
    Element
    ButtonType    string        `json:"buttonType" db:"button_type"`
    FrameElement  *FrameElement `json:"element,omitempty" db:"-"`
}

type InputElement struct {
    Element
}

type ListElement struct {
    Element
    Elements []any `json:"elements" db:"-"`
}

type SelectElement struct {
    Element
    Elements []any            `json:"elements" db:"-"`
}


type FormElement struct {
    Element
    Elements []any           `json:"elements" db:"-"`
}

func (e Element) TableName() string {
	return `public."Element"`
}

type EditorElement interface {
    GetElement() Element
    GetType() string
}

func (e Element) GetElement() Element {
    return e
}

func (e Element) GetType() string {
    return e.Type
}

func (fe FormElement) GetType() string {
    return fe.Type
}

func (fe FrameElement) GetElement() Element {
    return fe.Element
}

func (fe FrameElement) GetType() string {
    return fe.Type
}

func (ce CarouselElement) GetElement() Element {
    return ce.Element
}

func (ce CarouselElement) GetType() string {
    return ce.Type
}

func (be ButtonElement) GetElement() Element {
    return be.Element
}

func (be ButtonElement) GetType() string {
    return be.Type
}

func (ie InputElement) GetElement() Element {
    return ie.Element
}

func (ie InputElement) GetType() string {
    return ie.Type
}

func (le ListElement) GetElement() Element {
    return le.Element
}

func (le ListElement) GetType() string {
    return le.Type
}

func (se SelectElement) GetElement() Element {
    return se.Element
}

func (se SelectElement) GetType() string {
    return se.Type
}

func (fe FormElement) GetElement() Element {
    return fe.Element
}
