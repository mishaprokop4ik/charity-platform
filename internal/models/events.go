package models

type EventType string

const (
	ProposalEventType EventType = "proposal-event"
)

type Status string

const (
	InProcess   Status = "in_process"
	Completed   Status = "completed"
	Interrupted Status = "interrupted"
	Canceled    Status = "canceled"
)
