package models

import (
	"encoding/json"
	"net/http"
	"time"
)

type TransactionResponse struct {
	ID              uint `gorm:"primaryKey"`
	CreatorID       uint
	CompetitionDate time.Time
	EventID         uint
	EventType       EventType
	Status          Status
}

type StatusExport struct {
	Status Status `json:"status,omitempty"`
}

func UnmarshalStatusExport(r *http.Request) (StatusExport, error) {
	s := StatusExport{}
	err := json.NewDecoder(r.Body).Decode(&s)
	return s, err
}
