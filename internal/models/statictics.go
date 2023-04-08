package models

import (
	"encoding/json"
	"time"
)

type ProposalEventStatistics struct {
	Requests                                         []Request `json:"requests,omitempty"`
	StartDate                                        time.Time `json:"startDate"`
	EndDate                                          time.Time `json:"endDate"`
	TransactionsCount                                uint      `json:"transactionsCount"`
	TransactionsCountCompareWithPreviousMonth        int       `json:"transactionsCountCompare"`
	DoneTransactionsCount                            uint      `json:"doneTransactionsCount"`
	DoneTransactionsCountCompareWithPreviousMonth    int       `json:"doneTransactionsCountCompare"`
	CanceledTransactionCount                         uint      `json:"canceledTransactionCount"`
	CanceledTransactionCountCompareWithPreviousMonth int       `json:"canceledTransactionCountCompare"`
	AbortedTransactionsCount                         uint      `json:"abortedTransactionsCount"`
	AbortedTransactionsCountCompareWithPreviousMonth int       `json:"abortedTransactionsCountCompare"`
}

func (s ProposalEventStatistics) Bytes() []byte {
	bytes, _ := json.Marshal(s)
	return bytes
}

type Request struct {
	DayNumber     int8 `json:"dayNumber"`
	RequestsCount int  `json:"requestsCount"`
}