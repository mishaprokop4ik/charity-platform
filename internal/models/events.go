package models

import "encoding/json"

type EventType string

const (
	ProposalEventType EventType = "proposal-event"
	HelpEventType               = "help"
)

type EventStatus string

const (
	Active   EventStatus = "active"
	InActive EventStatus = "inactive"
	Done     EventStatus = "done"
	Blocked  EventStatus = "blocked"
)

type TransactionStatus string

const (
	InProcess         TransactionStatus = "in_progress"
	Waiting           TransactionStatus = "waiting"
	WaitingForApprove TransactionStatus = "waiting_for_approve"
	Completed         TransactionStatus = "completed"
	Interrupted       TransactionStatus = "interrupted"
	Canceled          TransactionStatus = "canceled"
	NotStarted        TransactionStatus = "not_started"
	Accepted          TransactionStatus = "accepted"
	Aborted           TransactionStatus = "aborted"
)

type File struct {
	Path string `json:"path"`
}

func (f File) Bytes() []byte {
	bytes, _ := json.Marshal(f)
	return bytes
}
