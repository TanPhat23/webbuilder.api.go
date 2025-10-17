package models

import (
	"time"
)

type MarketplaceItem struct {
	Id           string     `gorm:"column:Id;primaryKey" json:"id"`
	Title        string     `gorm:"column:Title;not null" json:"title"`
	Description  string     `gorm:"column:Description;not null" json:"description"`
	Preview      *string    `gorm:"column:Preview" json:"preview"`
	TemplateType string     `gorm:"column:TemplateType;not null;default:'block'" json:"templateType"`
	Featured     bool       `gorm:"column:Featured;not null;default:false" json:"featured"`
	PageCount    *int       `gorm:"column:PageCount" json:"pageCount"`
	Downloads    int        `gorm:"column:Downloads;not null;default:0" json:"downloads"`
	Likes        int        `gorm:"column:Likes;not null;default:0" json:"likes"`
	AuthorId     string     `gorm:"column:AuthorId;not null" json:"authorId"`
	AuthorName   string     `gorm:"column:AuthorName;not null" json:"authorName"`
	Verified     bool       `gorm:"column:Verified;not null;default:false" json:"verified"`
	CreatedAt    time.Time  `gorm:"column:CreatedAt;type:timestamp(6);not null" json:"createdAt"`
	UpdatedAt    time.Time  `gorm:"column:UpdatedAt;type:timestamp(6);not null" json:"updatedAt"`
	DeletedAt    *time.Time `gorm:"column:DeletedAt;type:timestamp(6)" json:"deletedAt,omitempty"`
	Tags         []Tag      `gorm:"many2many:MarketplaceItemTag;joinForeignKey:ItemId;joinReferences:TagId" json:"tags,omitempty"`
	Categories   []Category `gorm:"many2many:MarketplaceItemCategory;joinForeignKey:ItemId;joinReferences:CategoryId" json:"categories,omitempty"`
}

func (MarketplaceItem) TableName() string {
	return "MarketplaceItem"
}

type Category struct {
	Id   string `gorm:"column:Id;primaryKey" json:"id"`
	Name string `gorm:"column:Name;not null;unique" json:"name"`
}

func (Category) TableName() string {
	return "Category"
}

type Tag struct {
	Id   string `gorm:"column:Id;primaryKey" json:"id"`
	Name string `gorm:"column:Name;not null;unique" json:"name"`
}

func (Tag) TableName() string {
	return "Tag"
}

type MarketplaceItemTag struct {
	ItemId string `gorm:"column:ItemId;primaryKey" json:"itemId"`
	TagId  string `gorm:"column:TagId;primaryKey" json:"tagId"`
}

func (MarketplaceItemTag) TableName() string {
	return "MarketplaceItemTag"
}

type MarketplaceItemCategory struct {
	ItemId     string `gorm:"column:ItemId;primaryKey" json:"itemId"`
	CategoryId string `gorm:"column:CategoryId;primaryKey" json:"categoryId"`
}

func (MarketplaceItemCategory) TableName() string {
	return "MarketplaceItemCategory"
}

// Request/Response DTOs
type CreateMarketplaceItemRequest struct {
	Title        string   `json:"title" validate:"required"`
	Description  string   `json:"description" validate:"required"`
	Preview      *string  `json:"preview"`
	TemplateType string   `json:"templateType"`
	PageCount    *int     `json:"pageCount"`
	TagIds       []string `json:"tagIds"`
	CategoryIds  []string `json:"categoryIds"`
}

type UpdateMarketplaceItemRequest struct {
	Title        *string  `json:"title"`
	Description  *string  `json:"description"`
	Preview      *string  `json:"preview"`
	TemplateType *string  `json:"templateType"`
	Featured     *bool    `json:"featured"`
	PageCount    *int     `json:"pageCount"`
	TagIds       []string `json:"tagIds"`
	CategoryIds  []string `json:"categoryIds"`
}

type MarketplaceItemResponse struct {
	Id           string     `json:"id"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	Preview      *string    `json:"preview"`
	TemplateType string     `json:"templateType"`
	Featured     bool       `json:"featured"`
	PageCount    *int       `json:"pageCount"`
	Downloads    int        `json:"downloads"`
	Likes        int        `json:"likes"`
	AuthorId     string     `json:"authorId"`
	AuthorName   string     `json:"authorName"`
	Verified     bool       `json:"verified"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
	Tags         []Tag      `json:"tags"`
	Categories   []Category `json:"categories"`
}

type CreateCategoryRequest struct {
	Name string `json:"name" validate:"required"`
}

type CreateTagRequest struct {
	Name string `json:"name" validate:"required"`
}
