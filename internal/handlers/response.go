package handlers

import "Kurajj/internal/models"

type idResponse struct {
	err error
	id  int
}

type errResponse struct {
	err error
}

type getProposalEvents struct {
	proposalEvents models.ProposalEvents
	err            error
}

type commentsResponse struct {
	comments []models.Comment
	err      error
}

type transactionsResponse struct {
	transactions []models.Transaction
	err          error
}

type proposalEventsResponse struct {
	events    []models.ProposalEvent
	respError error
}
