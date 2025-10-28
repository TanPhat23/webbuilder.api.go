package models

import (
	"time"

)

type MarketplaceItem struct {
	Id           string     `gorm:"primaryKey;column:Id;type:varchar(255)" json:"id"`
	Title        string     `gorm:"column:Title;type:varchar(255);not null" json:"title"`
	Description  string     `gorm:"column:Description;type:text;not null" json:"description"`
	Preview      *string    `gorm:"column:Preview;type:text" json:"preview,omitempty"`
	TemplateType string     `gorm:"column:TemplateType;type:varchar(50);not null;default:'block'" json:"templateType"`
	Featured     bool       `gorm:"column:Featured;not null;default:false" json:"featured"`
	PageCount    *int       `gorm:"column:PageCount;type:int" json:"pageCount,omitempty"`
	Downloads    int        `gorm:"column:Downloads;not null;default:0" json:"downloads"`
	Likes        int        `gorm:"column:Likes;not null;default:0" json:"likes"`
	AuthorId     string     `gorm:"column:AuthorId;type:varchar(255);not null" json:"authorId"`
	AuthorName   string     `gorm:"column:AuthorName;type:varchar(255);not null" json:"authorName"`
	Verified     bool       `gorm:"column:Verified;not null;default:false" json:"verified"`
	ProjectId    *string    `gorm:"column:ProjectId;type:varchar(255)" json:"projectId,omitempty"`
	CreatedAt    time.Time  `gorm:"column:CreatedAt" json:"createdAt,omitempty"`
	UpdatedAt    time.Time  `gorm:"column:UpdatedAt" json:"updatedAt,omitempty"`
	DeletedAt    *time.Time `gorm:"column:DeletedAt" json:"deletedAt,omitempty"`
	Tags         []Tag      `gorm:"many2many:MarketplaceItemTag;foreignKey:Id;joinForeignKey:ItemId;References:Id;joinReferences:TagId" json:"tags,omitempty"`
	Categories   []Category `gorm:"many2many:MarketplaceItemCategory;foreignKey:Id;joinForeignKey:ItemId;References:Id;joinReferences:CategoryId" json:"categories,omitempty"`
}

func (MarketplaceItem) TableName() string {
	return `public."MarketplaceItem"`
}

type Category struct {
	Id   string `gorm:"primaryKey;column:Id;type:varchar(255)" json:"id"`
	Name string `gorm:"column:Name;type:varchar(255);not null;unique" json:"name"`
}

func (Category) TableName() string {
	return `public."Category"`
}

type Tag struct {
	Id   string `gorm:"primaryKey;column:Id;type:varchar(255)" json:"id"`
	Name string `gorm:"column:Name;type:varchar(255);not null;unique" json:"name"`
}

func (Tag) TableName() string {
	return `public."Tag"`
}

type MarketplaceItemTag struct {
	ItemId string `gorm:"primaryKey;column:ItemId;type:varchar(255)" json:"itemId"`
	TagId  string `gorm:"primaryKey;column:TagId;type:varchar(255)" json:"tagId"`
}

func (MarketplaceItemTag) TableName() string {
	return `public."MarketplaceItemTag"`
}

type MarketplaceItemCategory struct {
	ItemId     string `gorm:"primaryKey;column:ItemId;type:varchar(255)" json:"itemId"`
	CategoryId string `gorm:"primaryKey;column:CategoryId;type:varchar(255)" json:"categoryId"`
}

func (MarketplaceItemCategory) TableName() string {
	return `public."MarketplaceItemCategory"`
}

// Request/Response DTOs
type CreateMarketplaceItemRequest struct {
	Title        string   `json:"title" validate:"required"`
	Description  string   `json:"description" validate:"required"`
	Preview      *string  `json:"preview"`
	TemplateType string   `json:"templateType"`
	PageCount    *int     `json:"pageCount"`
	ProjectId    *string  `json:"projectId"`
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
	ProjectId    *string  `json:"projectId"`
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
	ProjectId    *string    `json:"projectId,omitempty"`
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
