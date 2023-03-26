package models

import "time"

type TransactionNotification struct {
	ID            uint              `gorm:"column:id"`
	EventType     EventType         `gorm:"column:event_type"`
	EventID       uint              `gorm:"column:event_id"`
	Action        TransactionAction `gorm:"column:action"`
	TransactionID uint              `gorm:"column:transaction_id"`
	NewStatus     TransactionStatus `gorm:"column:new_status"`
	IsRead        bool              `gorm:"column:is_read"`
	CreationTime  time.Time         `gorm:"column:creation_time"`
	MemberID      uint              `gorm:"column:member_id"`
	EventTitle    string            `gorm:"-"`
}

type TransactionAction string

const (
	Created TransactionAction = "created"
	Updated TransactionAction = "updated"
)

func (TransactionNotification) TableName() string {
	return "notification"
}
