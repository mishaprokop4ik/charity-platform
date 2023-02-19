package models

import (
	"encoding/json"
	"net/http"
	"time"
)

type DescriptionField struct {
	Name string `json:"name,omitempty"`
	Text string `json:"text,omitempty"`
}

type ProposalEventRequestCreate struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

func UnmarshalProposalEventCreate(r *http.Request) (ProposalEventRequestCreate, error) {
	e := ProposalEventRequestCreate{}
	err := json.NewDecoder(r.Body).Decode(&e)
	return e, err
}

type ProposalEventGetResponse struct {
	ID              uint                  `json:"id,omitempty"`
	Title           string                `json:"title,omitempty"`
	Description     string                `json:"description,omitempty"`
	CreationDate    string                `json:"creationDate,omitempty"`
	CompetitionDate string                `json:"competitionDate,omitempty"`
	AuthorID        uint                  `json:"authorID,omitempty"`
	Category        string                `json:"category,omitempty"`
	Comments        []Comment             `json:"comments,omitempty"`
	Transactions    []TransactionResponse `json:"transactions,omitempty"`
}

func (p ProposalEventGetResponse) Bytes() []byte {
	bytes, _ := json.Marshal(p)
	return bytes
}

func GetProposalEvent(event ProposalEvent) ProposalEventGetResponse {
	completionDate := ""
	if event.CompetitionDate.Valid {
		completionDate = event.CompetitionDate.Time.String()
	}
	return ProposalEventGetResponse{
		ID:              event.ID,
		Title:           event.Title,
		Description:     event.Description,
		CreationDate:    event.CreationDate.String(),
		CompetitionDate: completionDate,
		AuthorID:        event.AuthorID,
		Category:        event.Category,
		//Comments:        event.Comments,
		//Transactions:    event.Transactions,
	}
}

type ProposalEventList struct {
	ProposalEvents []ProposalEventGetResponse `json:"proposalEvents,omitempty"`
}

func GetProposalEvents(events ...ProposalEvent) ProposalEventList {
	list := ProposalEventList{}
	for _, e := range events {
		list.ProposalEvents = append(list.ProposalEvents, GetProposalEvent(e))
	}
	return list
}

func (l ProposalEventList) Bytes() []byte {
	bytes, _ := json.Marshal(l)
	return bytes
}

type ProposalEventRequestUpdate struct {
	ID              uint      `json:"id,omitempty"`
	Title           string    `json:"title,omitempty"`
	Description     string    `json:"description,omitempty"`
	CompetitionDate time.Time `json:"competitionDate,omitempty"`
	Category        string    `json:"category,omitempty"`
}

func UnmarshalProposalEventUpdate(r *http.Request) (ProposalEventRequestUpdate, error) {
	e := ProposalEventRequestUpdate{}
	err := json.NewDecoder(r.Body).Decode(&e)
	return e, err
}
