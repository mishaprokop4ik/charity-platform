package models

import "time"

type TransactionResponse struct {
	ID              uint `gorm:"primaryKey"`
	CreatorID       uint
	CompetitionDate time.Time
	EventID         uint
	EventType       EventType
	Status          Status
}
