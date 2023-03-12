package models

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

type ProposalEventRequestCreate struct {
	Title                 string       `json:"title,omitempty"`
	Description           string       `json:"description,omitempty"`
	MaxConcurrentRequests int          `json:"maxConcurrentRequests,omitempty"`
	Location              Address      `json:"-"`
	Tags                  []TagRequest `json:"tags,omitempty"`
}

func (Address) TableName() string {
	return "location"
}

func UnmarshalProposalEventCreate(r *io.ReadCloser) (ProposalEventRequestCreate, error) {
	e := ProposalEventRequestCreate{}
	err := json.NewDecoder(*r).Decode(&e)
	if err != nil {
		return ProposalEventRequestCreate{}, err
	}
	for i, tag := range e.Tags {
		if tag.Title == "location" {
			if len(tag.Values) != 4 {
				return ProposalEventRequestCreate{}, fmt.Errorf("location tag is incorrect")
			}
			locationValues := tag.Values
			e.Location = Address{
				Region:       locationValues[0],
				City:         locationValues[1],
				District:     locationValues[2],
				HomeLocation: locationValues[3],
			}
			e.Tags = append(e.Tags[:i], e.Tags[i+1:]...)
		}
	}
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
	User                  UserShortInfo         `json:"authorInfo,omitempty"`
	Category              string                `json:"category,omitempty"`
	Comments              []CommentResponse     `json:"comments,omitempty"`
	Transactions          []TransactionResponse `json:"transactions,omitempty"`
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
			UserShortInfo: UserShortInfo{
				ID: comment.UserID,
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
		User: UserShortInfo{
			ID:              event.AuthorID,
			Username:        event.User.FullName,
			ProfileImageURL: event.User.AvatarImagePath,
			PhoneNumber:     Telephone(event.User.Telephone),
		},
		Category:     event.Category,
		Comments:     comments,
		Transactions: transactions,
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
