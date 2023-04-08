package handlers

import (
	"Kurajj/internal/models"
)

type idResponse struct {
	err error
	id  int
}

type errResponse struct {
	err error
}

type refreshTokenResponse struct {
	tokens models.Tokens
	err    error
}

type getProposalEvents struct {
	proposalEvents models.ProposalEvents
	err            error
}

type getProposalEventPagination struct {
	resp models.ProposalEventsWithPagination
	err  error
}

type getTagsResponse struct {
	tags []models.Tag
	err  error
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

type getProposalEvent struct {
	proposalEvent models.ProposalEventGetResponse
	err           error
}

type getProposalStatistics struct {
	statistics models.ProposalEventStatistics
	err        error
}
