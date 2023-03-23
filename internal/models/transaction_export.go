package models

import (
	"encoding/json"
	"io"
	"time"
)

type TransactionResponse struct {
	ID              uint              `json:"id"`
	CreatorID       uint              `json:"creatorID"`
	Creator         UserShortInfo     `json:"creator"`
	CreationDate    time.Time         `json:"creationDate"`
	CompetitionDate time.Time         `json:"competitionDate"`
	EventID         uint              `json:"eventID"`
	EventType       EventType         `json:"eventType"`
	Responder       UserShortInfo     `json:"responder"`
	Comment         string            `json:"comment"`
	ReceiverStatus  TransactionStatus `json:"receiverStatus"`
	ResponderStatus TransactionStatus `json:"responderStatus"`
}

type StatusExport struct {
	Status TransactionStatus `json:"status"`
}

type TransactionsExport struct {
	Transactions []TransactionResponse `json:"transactions"`
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
