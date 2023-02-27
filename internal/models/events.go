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
)

type Status string

const (
	InProcess   Status = "in_process"
	Completed   Status = "completed"
	Interrupted Status = "interrupted"
	Canceled    Status = "canceled"
	Waiting     Status = "waiting"
)
