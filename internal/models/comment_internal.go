package models

import (
	"database/sql"
	"time"
)

type Comment struct {
	ID           uint         `gorm:"primaryKey" gorm:"column:id"`
	EventID      uint         `gorm:"column:event_id"`
	EventType    EventType    `gorm:"column:event_type"`
	Text         string       `gorm:"column:text"`
	UserID       uint         `gorm:"column:user_id"`
	CreationDate time.Time    `gorm:"column:creation_date"`
	IsUpdated    bool         `gorm:"column:is_updated"`
	UpdateTime   sql.NullTime `gorm:"column:update_time"`
	IsDeleted    bool         `gorm:"column:is_deleted"`
}
