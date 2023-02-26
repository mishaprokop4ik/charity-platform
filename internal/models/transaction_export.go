package models

import (
	"encoding/json"
	"net/http"
	"time"
)

type TransactionResponse struct {
	ID                uint      `json:"id,omitempty"`
	CreatorID         uint      `json:"creatorID,omitempty"`
	CompetitionDate   time.Time `json:"competitionDate"`
	EventID           uint      `json:"eventID,omitempty"`
	EventType         EventType `json:"eventType,omitempty"`
	Comment           string    `json:"comment"`
	TransactionStatus Status    `json:"transactionStatus,omitempty"`
	ResponderStatus   Status    `json:"responderStatus,omitempty"`
}

type StatusExport struct {
	Status Status `json:"status,omitempty"`
}

type TransactionsExport struct {
	Transactions []TransactionResponse `json:"transactions,omitempty"`
}

func (t TransactionsExport) Bytes() []byte {
	bytes, _ := json.Marshal(t)
	return bytes
}

func UnmarshalStatusExport(r *http.Request) (StatusExport, error) {
	s := StatusExport{}
	err := json.NewDecoder(r.Body).Decode(&s)
	return s, err
}
