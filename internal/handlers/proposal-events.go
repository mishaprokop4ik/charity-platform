package handlers

import (
	"Kurajj/internal/models"
	"net/http"
)

func (h Handler) CreateProposalEvent(w http.ResponseWriter, r *http.Request) {
	proposalEvent, err := models.CreateProposalEventFromRequest(r)

}
