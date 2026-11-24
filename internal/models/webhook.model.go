package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

const (
	WebhookEventItemCreated = "item.created"
	WebhookEventItemUpdated = "item.updated"
	WebhookEventItemDeleted = "item.deleted"
	WebhookEventTest        = "webhook.test"

	WebhookDeliveryStatusPending   = "pending"
	WebhookDeliveryStatusSucceeded = "succeeded"
	WebhookDeliveryStatusFailed    = "failed"
)

func IsValidWebhookEvent(event string) bool {
	switch event {
	case WebhookEventItemCreated, WebhookEventItemUpdated, WebhookEventItemDeleted:
		return true
	default:
		return false
	}
}

type StringSlice []string

func (s StringSlice) Value() (driver.Value, error) {
	if s == nil {
		return "[]", nil
	}
	raw, err := json.Marshal([]string(s))
	if err != nil {
		return nil, err
	}
	return string(raw), nil
}

func (s *StringSlice) Scan(value any) error {
	if value == nil {
		*s = StringSlice{}
		return nil
	}
	var raw []byte
	switch v := value.(type) {
	case []byte:
		raw = v
	case string:
		raw = []byte(v)
	default:
		return fmt.Errorf("unsupported StringSlice scan type %T", value)
	}
	var out []string
	if err := json.Unmarshal(raw, &out); err != nil {
		return err
	}
	*s = StringSlice(out)
	return nil
}

type ProjectWebhook struct {
	Id            string      `gorm:"primaryKey;column:Id;type:varchar(255)" json:"id"`
	ProjectId     string      `gorm:"column:ProjectId;type:varchar(255);not null;index" json:"projectId"`
	ContentTypeId *string     `gorm:"column:ContentTypeId;type:varchar(255)" json:"contentTypeId,omitempty"`
	Name          string      `gorm:"column:Name;type:varchar(255);not null" json:"name"`
	Url           string      `gorm:"column:Url;type:text;not null" json:"url"`
	Secret        string      `gorm:"column:Secret;type:varchar(255);not null" json:"-"`
	Events        StringSlice `gorm:"column:Events;type:jsonb;not null" json:"events"`
	Enabled       bool        `gorm:"column:Enabled;not null;default:true" json:"enabled"`
	CreatedAt     time.Time   `gorm:"column:CreatedAt" json:"createdAt"`
	UpdatedAt     time.Time   `gorm:"column:UpdatedAt" json:"updatedAt"`
}

func (ProjectWebhook) TableName() string {
	return `public."ProjectWebhook"`
}

func (w ProjectWebhook) SubscribesTo(event string) bool {
	for _, ev := range w.Events {
		if ev == event {
			return true
		}
	}
	return false
}

type WebhookDelivery struct {
	Id           string          `gorm:"primaryKey;column:Id;type:varchar(255)" json:"id"`
	WebhookId    string          `gorm:"column:WebhookId;type:varchar(255);not null;index" json:"webhookId"`
	ProjectId    string          `gorm:"column:ProjectId;type:varchar(255);not null" json:"projectId"`
	Event        string          `gorm:"column:Event;type:varchar(64);not null" json:"event"`
	Payload      json.RawMessage `gorm:"column:Payload;type:jsonb;not null" json:"payload"`
	Status       string          `gorm:"column:Status;type:varchar(32);not null;default:pending" json:"status"`
	HttpStatus   *int            `gorm:"column:HttpStatus" json:"httpStatus,omitempty"`
	ResponseBody *string         `gorm:"column:ResponseBody;type:text" json:"responseBody,omitempty"`
	ErrorMessage *string         `gorm:"column:ErrorMessage;type:text" json:"errorMessage,omitempty"`
	Attempts     int             `gorm:"column:Attempts;not null;default:0" json:"attempts"`
	NextRetryAt  *time.Time      `gorm:"column:NextRetryAt" json:"nextRetryAt,omitempty"`
	CreatedAt    time.Time       `gorm:"column:CreatedAt" json:"createdAt"`
	CompletedAt  *time.Time      `gorm:"column:CompletedAt" json:"completedAt,omitempty"`
}

func (WebhookDelivery) TableName() string {
	return `public."WebhookDelivery"`
}
