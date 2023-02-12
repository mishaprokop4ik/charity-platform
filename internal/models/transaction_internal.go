package models

import "database/sql"

type Transaction struct {
	ID              uint `gorm:"primaryKey"`
	CreatorID       uint
	CompetitionDate sql.NullTime
	EventID         uint
	EventType       EventType
	Status          TransactionStatus
}
