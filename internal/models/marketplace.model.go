package models

import (
	"time"
)

type MarketplaceItem struct {
	Tags         []Tag      `gorm:"many2many:MarketplaceItemTag;foreignKey:Id;joinForeignKey:ItemId;References:Id;joinReferences:TagId" json:"tags,omitempty"`
	Categories   []Category `gorm:"many2many:MarketplaceItemCategory;foreignKey:Id;joinForeignKey:ItemId;References:Id;joinReferences:CategoryId" json:"categories,omitempty"`
	CreatedAt    time.Time  `gorm:"column:CreatedAt" json:"createdAt,omitempty"`
	UpdatedAt    time.Time  `gorm:"column:UpdatedAt" json:"updatedAt,omitempty"`
	DeletedAt    *time.Time `gorm:"column:DeletedAt" json:"deletedAt,omitempty"`
	Id           string     `gorm:"primaryKey;column:Id;type:varchar(255)" json:"id"`
	ProjectId    *string    `gorm:"column:ProjectId;type:varchar(255)" json:"projectId,omitempty"`
	Title        string     `gorm:"column:Title;type:varchar(255);not null" json:"title"`
	Description  string     `gorm:"column:Description;type:text;not null" json:"description"`
	TemplateType string     `gorm:"column:TemplateType;type:varchar(50);not null;default:'block'" json:"templateType"`
	AuthorId     string     `gorm:"column:AuthorId;type:varchar(255);not null" json:"authorId"`
	AuthorName   string     `gorm:"column:AuthorName;type:varchar(255);not null" json:"authorName"`
	Preview      *string    `gorm:"column:Preview;type:text" json:"preview,omitempty"`
	PageCount    *int       `gorm:"column:PageCount;type:int" json:"pageCount,omitempty"`
	Downloads    int        `gorm:"column:Downloads;not null;default:0" json:"downloads"`
	Likes        int        `gorm:"column:Likes;not null;default:0" json:"likes"`
	Featured     bool       `gorm:"column:Featured;not null;default:false" json:"featured"`
	Verified     bool       `gorm:"column:Verified;not null;default:false" json:"verified"`
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

// Request DTOs - Ordered by size for optimal alignment (24-byte slices → 16-byte strings → 8-byte pointers)
type CreateMarketplaceItemRequest struct {
	TagIds       []string `json:"tagIds"       validate:"omitempty,dive,min=1"`
	CategoryIds  []string `json:"categoryIds"  validate:"omitempty,dive,min=1"`
	Title        string   `json:"title"        validate:"required,min=1,max=255"`
	Description  string   `json:"description"  validate:"required,min=1,max=5000"`
	TemplateType string   `json:"templateType" validate:"required,min=1,max=50,oneof=block page template section"`
	Preview      *string  `json:"preview"      validate:"omitempty,min=1,max=5000"`
	ProjectId    *string  `json:"projectId"    validate:"omitempty,min=1"`
	PageCount    *int     `json:"pageCount"    validate:"omitempty,gte=1,lte=1000"`
}

type UpdateMarketplaceItemRequest struct {
	TagIds       []string `json:"tagIds"       validate:"omitempty,dive,min=1"`
	CategoryIds  []string `json:"categoryIds"  validate:"omitempty,dive,min=1"`
	Title        *string  `json:"title"        validate:"omitempty,min=1,max=255"`
	Description  *string  `json:"description"  validate:"omitempty,min=1,max=5000"`
	Preview      *string  `json:"preview"      validate:"omitempty,min=1,max=5000"`
	TemplateType *string  `json:"templateType" validate:"omitempty,min=1,max=50,oneof=block page template section"`
	ProjectId    *string  `json:"projectId"    validate:"omitempty,min=1"`
	Featured     *bool    `json:"featured"`
	PageCount    *int     `json:"pageCount"    validate:"omitempty,gte=1,lte=1000"`
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
	Name string `json:"name" validate:"required,min=1,max=100"`
}

type CreateTagRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}
