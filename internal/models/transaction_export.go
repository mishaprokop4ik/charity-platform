package models

import (
	"encoding/json"
	"io"
	"time"
)

type TransactionResponse struct {
	ID                uint              `json:"id,omitempty"`
	CreatorID         uint              `json:"creatorID,omitempty"`
	CompetitionDate   time.Time         `json:"competitionDate"`
	EventID           uint              `json:"eventID,omitempty"`
	EventType         EventType         `json:"eventType,omitempty"`
	Comment           string            `json:"comment"`
	TransactionStatus TransactionStatus `json:"transactionStatus,omitempty"`
	ResponderStatus   TransactionStatus `json:"responderStatus,omitempty"`
}

type StatusExport struct {
	Status TransactionStatus `json:"status,omitempty"`
}

type TransactionsExport struct {
	Transactions []TransactionResponse `json:"transactions,omitempty"`
}

func (t TransactionsExport) Bytes() []byte {
	bytes, _ := json.Marshal(t)
	return bytes
}

func UnmarshalStatusExport(r *io.ReadCloser) (StatusExport, error) {
	s := StatusExport{}
	err := json.NewDecoder(*r).Decode(&s)
	return s, err
}
