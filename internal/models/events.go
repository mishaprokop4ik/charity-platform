package models

type EventType string

const (
	ProposalEventType EventType = "proposal-event"
)

type TransactionStatus string

const (
	Finished    TransactionStatus = "finished"
	InProcess   TransactionStatus = "in_process"
	Completed   TransactionStatus = "completed"
	Interrupted TransactionStatus = "interrupted"
	Canceled    TransactionStatus = "canceled"
)
