package models

import (
	"database/sql"
	"gorm.io/gorm"
)

type Transaction struct {
	gorm.Model
	CreatorID       uint
	CompetitionDate sql.NullTime
	EventID         uint
	EventType       EventType
	Status          TransactionStatus
}
