package models

type EventType string

const (
	ProposalEventType EventType = "proposal-event"
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
	InProcess   TransactionStatus = "in_process"
	Completed   TransactionStatus = "completed"
	Interrupted TransactionStatus = "interrupted"
	Canceled    TransactionStatus = "canceled"
	NotStarted  TransactionStatus = "not_started"
)

type Eventer interface {
	Serialize() ([]byte, error)
}
