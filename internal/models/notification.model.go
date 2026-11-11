package models

import "time"

type Notification struct {
	Id          string    `gorm:"primaryKey;column:Id;type:varchar(255)" json:"id"`
	UserId      string    `gorm:"column:UserId;type:varchar(255);not null;index" json:"userId"`
	Type        string    `gorm:"column:Type;type:varchar(100);not null" json:"type"`
	Title       string    `gorm:"column:Title;type:varchar(500);not null" json:"title"`
	Description string    `gorm:"column:Description;type:text;not null" json:"description"`
	Read        bool      `gorm:"column:Read;not null;default:false" json:"read"`
	CreatedAt   time.Time `gorm:"column:CreatedAt" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"column:UpdatedAt" json:"updatedAt"`
}

func (Notification) TableName() string {
	return `public."Notification"`
}

// NotificationFilter narrows a user's notification list. Filter is "all",
// "unread" or a notification type ("message", "user", "alert", "settings").
type NotificationFilter struct {
	Filter string
	Search string
	Limit  int
}

type NotificationStats struct {
	Total     int64 `json:"total"`
	Unread    int64 `json:"unread"`
	ReadToday int64 `json:"readToday"`
}

type NotificationFilterCounts struct {
	All       int64 `json:"all"`
	Unread    int64 `json:"unread"`
	Messages  int64 `json:"messages"`
	Followers int64 `json:"followers"`
	Alerts    int64 `json:"alerts"`
	Updates   int64 `json:"updates"`
}

type NotificationList struct {
	Notifications []Notification           `json:"notifications"`
	Stats         NotificationStats        `json:"stats"`
	FilterCounts  NotificationFilterCounts `json:"filterCounts"`
}

// ProjectAccess is a user's role on a project (owner or collaborator).
type ProjectAccess struct {
	UserId         string           `json:"userId"`
	ProjectId      string           `json:"projectId"`
	Role           CollaboratorRole `json:"role"`
	IsOwner        bool             `json:"isOwner"`
	IsCollaborator bool             `json:"isCollaborator"`
}
