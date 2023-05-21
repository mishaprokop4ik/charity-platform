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

type getHelpEventPagination struct {
	resp models.HelpEventsWithPagination
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

type getProposalEvent struct {
	proposalEvent models.ProposalEventGetResponse
	err           error
}

type getHelpEvent struct {
	helpEvent models.HelpEvent
	err       error
}

type helpEventsResponse struct {
	events []models.HelpEvent
	err    error
}

type complaintsResponse struct {
	complaints []models.ComplaintsResponse
	err        error
}

type getProposalStatistics struct {
	statistics models.ProposalEventStatistics
	err        error
}

type getGlobalStatistics struct {
	statistics models.GlobalStatistics
	err        error
}

type geHelpStatistics struct {
	statistics models.HelpEventStatistics
	err        error
}
