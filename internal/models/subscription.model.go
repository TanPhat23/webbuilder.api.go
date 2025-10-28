package models

import (
	"time"
)

type Subscription struct {
	Id            string    `gorm:"primaryKey;column:Id;type:varchar(255)" json:"id"`
	UserId        string    `gorm:"column:UserId;type:varchar(255);not null;index" json:"userId"`
	PlanId        string    `gorm:"column:PlanId;type:varchar(255);not null" json:"planId"`
	BillingPeriod string    `gorm:"column:BillingPeriod;type:varchar(50);not null" json:"billingPeriod"`
	Status        string    `gorm:"column:Status;type:varchar(50);not null;default:'active';index" json:"status"`
	StartDate     time.Time `gorm:"column:StartDate;default:CURRENT_TIMESTAMP" json:"startDate,omitempty"`
	EndDate       *time.Time `gorm:"column:EndDate" json:"endDate,omitempty"`
	Amount        float64   `gorm:"column:Amount;type:decimal(10,2);not null" json:"amount"`
	Currency      string    `gorm:"column:Currency;type:varchar(3);not null;default:'USD'" json:"currency"`
	CreatedAt     time.Time `gorm:"column:CreatedAt;default:CURRENT_TIMESTAMP" json:"createdAt,omitempty"`
	UpdatedAt     time.Time `gorm:"column:UpdatedAt;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"updatedAt,omitempty"`
	User          User      `gorm:"foreignKey:UserId" json:"user,omitempty"`
}

func (Subscription) TableName() string {
	return `public."Subscription"`
}
