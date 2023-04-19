package models

import "time"

type HelpEventTransactionResponse struct {
	TransactionID         uint              `json:"transactionID"`
	Needs                 []NeedResponse    `json:"needs"`
	CompetitionDate       string            `json:"competitionDate"`
	IsApproved            bool              `json:"isApproved"`
	CompletionPercentages float64           `json:"completionPercentages"`
	CreatorID             uint              `json:"creatorID"`
	Receiver              UserShortInfo     `json:"receiver"`
	CreationDate          time.Time         `json:"creationDate"`
	EventID               uint              `json:"eventID"`
	EventType             EventType         `json:"eventType"`
	Responder             UserShortInfo     `json:"responder"`
	Comment               string            `json:"comment"`
	TransactionStatus     TransactionStatus `json:"transactionStatus"`
	ResponderStatus       TransactionStatus `json:"responderStatus"`
	ReportURL             string            `json:"reportURL"`
}

type HelpEventTransaction struct {
	Needs                                    []Need
	Received                                 int
	ReceivedTotal                            int
	CompetitionDate                          time.Time
	CompletionPercentages                    int
	HelpEventCreatorID, TransactionCreatorID uint
	TransactionID                            *uint
	HelpEventID                              *uint
	TransactionStatus                        TransactionStatus
	ResponderStatus                          TransactionStatus
	EventCreator                             bool
}
