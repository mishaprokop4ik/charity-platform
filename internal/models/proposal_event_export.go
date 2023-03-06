package models

import (
	"encoding/json"
	"io"
	"time"
)

type ProposalEventRequestCreate struct {
	Title                 string   `json:"title,omitempty"`
	Description           string   `json:"description,omitempty"`
	MaxConcurrentRequests int      `json:"maxConcurrentRequests,omitempty"`
	Location              Location `json:"location,omitempty"`
}

type Location struct {
	Country  string `json:"country,omitempty"`
	Area     string `json:"area,omitempty"`
	City     string `json:"city,omitempty"`
	District string `json:"district,omitempty"`
	Street   string `json:"street,omitempty"`
	Home     string `json:"home,omitempty"`
}

func (Location) TableName() string {
	return "location"
}

func UnmarshalProposalEventCreate(r *io.ReadCloser) (ProposalEventRequestCreate, error) {
	e := ProposalEventRequestCreate{}
	err := json.NewDecoder(*r).Decode(&e)
	return e, err
}

type ProposalEventGetResponse struct {
	ID                    uint                  `json:"id,omitempty"`
	Title                 string                `json:"title,omitempty"`
	Description           string                `json:"description,omitempty"`
	CreationDate          string                `json:"creationDate,omitempty"`
	MaxConcurrentRequests uint                  `json:"maxConcurrentRequests,omitempty"`
	AvailableHelps        uint                  `json:"availableHelps,omitempty"`
	CompetitionDate       string                `json:"competitionDate,omitempty"`
	Status                EventStatus           `json:"status,omitempty"`
	AuthorID              uint                  `json:"authorID,omitempty"`
	Category              string                `json:"category,omitempty"`
	Comments              []CommentResponse     `json:"comments,omitempty"`
	Transactions          []TransactionResponse `json:"transactions,omitempty"`
	Location              Location              `json:"location,omitempty"`
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
	comments := make([]CommentResponse, len(event.Comments))
	for i, comment := range event.Comments {
		updatedTime := ""
		if comment.UpdatedAt.Valid {
			updatedTime = comment.UpdatedAt.Time.String()
		}
		comments[i] = CommentResponse{
			ID:           comment.ID,
			Text:         comment.Text,
			CreationDate: comment.CreationDate,
			IsUpdated:    comment.IsUpdated,
			UpdateTime:   updatedTime,
			UserComment: UserComment{
				AuthorID: comment.UserID,
			},
		}
	}

	transactions := make([]TransactionResponse, len(event.Transactions))
	for i, t := range event.Transactions {
		transaction := TransactionResponse{
			ID:                t.ID,
			CreatorID:         t.CreatorID,
			EventID:           t.EventID,
			Comment:           t.Comment,
			EventType:         t.EventType,
			TransactionStatus: t.TransactionStatus,
			ResponderStatus:   t.ResponderStatus,
		}
		if t.CompetitionDate.Valid {
			transaction.CompetitionDate = t.CompetitionDate.Time
		}
		transactions[i] = transaction
	}
	return ProposalEventGetResponse{
		ID:              event.ID,
		Title:           event.Title,
		Description:     event.Description,
		CreationDate:    event.CreationDate.String(),
		CompetitionDate: completionDate,
		AuthorID:        event.AuthorID,
		Category:        event.Category,
		Comments:        comments,
		Transactions:    transactions,
		Location:        event.Location,
	}
}

type ProposalEvents struct {
	ProposalEvents []ProposalEventGetResponse `json:"proposalEvents,omitempty"`
}

func GetProposalEvents(events ...ProposalEvent) ProposalEvents {
	responseEvents := ProposalEvents{}
	for _, e := range events {
		responseEvents.ProposalEvents = append(responseEvents.ProposalEvents, GetProposalEvent(e))
	}
	return responseEvents
}

func (l ProposalEvents) Bytes() []byte {
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

func UnmarshalProposalEventUpdate(r *io.ReadCloser) (ProposalEventRequestUpdate, error) {
	e := ProposalEventRequestUpdate{}
	err := json.NewDecoder(*r).Decode(&e)
	return e, err
}
