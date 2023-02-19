package models

import (
	"database/sql"
	"time"
)

type Transaction struct {
	ID              uint `gorm:"primaryKey"`
	CreatorID       uint
	CompetitionDate sql.NullTime
	EventID         uint
	EventType       EventType
	Status          TransactionStatus
}

type TransactionResponse struct {
	ID              uint `gorm:"primaryKey"`
	CreatorID       uint
	CompetitionDate time.Time
	EventID         uint
	EventType       EventType
	Status          TransactionStatus
}
