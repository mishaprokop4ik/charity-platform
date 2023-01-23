package models

import (
	"encoding/json"
	"github.com/go-playground/validator/v10"
	"net/http"
)

type ProposalEvent struct {
}

func (e *ProposalEvent) Validate() error {
	validate := validator.New()
	return validate.Struct(e)
}

// CreateProposalEventFromRequest gets JSON body from HTTP request and creates and validates new ProposalEvent instance.
func CreateProposalEventFromRequest(r *http.Request) (*ProposalEvent, error) {
	proposalEvent := &ProposalEvent{}
	if err := json.NewDecoder(r.Body).Decode(proposalEvent); err != nil {
		return &ProposalEvent{}, err
	}

	if err := proposalEvent.Validate(); err != nil {
		return &ProposalEvent{}, err
	}

	return proposalEvent, nil
}
