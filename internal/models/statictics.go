package models

import (
	"encoding/json"
	"time"
)

type DefaultStatistics struct {
	Requests                                         []Request `json:"requests"`
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

func (s DefaultStatistics) Bytes() []byte {
	bytes, _ := json.Marshal(s)
	return bytes
}

type ProposalEventStatistics struct {
	DefaultStatistics
}

type HelpEventStatistics struct {
	DefaultStatistics
}

type GlobalStatistics struct {
	DefaultStatistics
}

type Request struct {
	Date          string `json:"date"`
	RequestsCount int    `json:"requestsCount"`
}
