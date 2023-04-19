package models

type HelpEventTransactionResponse struct {
	TransactionID         uint           `json:"transactionID"`
	Needs                 []NeedResponse `json:"needs"`
	CompetitionDate       string         `json:"competitionDate"`
	IsApproved            bool           `json:"isApproved"`
	CompletionPercentages float64        `json:"completionPercentages"`
}

type HelpEventTransaction struct {
	Needs                                    []Need
	Received                                 int
	ReceivedTotal                            int
	CompetitionDate                          string
	CompletionPercentages                    int
	HelpEventCreatorID, TransactionCreatorID uint
	TransactionID                            *uint
	HelpEventID                              *uint
	TransactionStatus                        TransactionStatus
	ResponderStatus                          TransactionStatus
	EventCreator                             bool
}
