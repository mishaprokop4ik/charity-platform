package models

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

func UnmarshalProposalEventUpdate(r *io.ReadCloser) (ProposalEventRequestUpdate, error) {
	e := ProposalEventRequestUpdate{}
	err := json.NewDecoder(*r).Decode(&e)
	return e, err
}

func UnmarshalProposalEventCreate(r *io.ReadCloser) (ProposalEventRequestCreate, error) {
	e := ProposalEventRequestCreate{}
	err := json.NewDecoder(*r).Decode(&e)
	return e, err
}

type ProposalEventRequestCreate struct {
	Title                 string       `json:"title"`
	Description           string       `json:"description"`
	MaxConcurrentRequests int          `json:"maxConcurrentRequests"`
	Tags                  []TagRequest `json:"tags"`
}

func (p *ProposalEventRequestCreate) TagsInternal() []Tag {
	tags := make([]Tag, len(p.Tags))
	for i, tag := range p.Tags {
		tagValues := make([]TagValue, len(tag.Values))
		for _, tagValue := range tag.Values {
			tagValues[i] = TagValue{
				Value: tagValue,
			}
		}
		tags[i] = Tag{
			Title:     tag.Title,
			EventType: ProposalEventType,
			Values:    tagValues,
		}
	}
	return tags
}

func (p *ProposalEventRequestCreate) InternalValue(userID uint) ProposalEvent {
	location := Address{}
	for i, t := range p.Tags {
		if t.Title == "location" && len(t.Values) >= DecodedAddressLength {
			location.Region = t.Values[0]
			location.City = t.Values[1]
			location.District = t.Values[2]
			location.HomeLocation = t.Values[3]
			location.EventType = ProposalEventType
			p.Tags = append(p.Tags[:i], p.Tags[i+1:]...)
		}
	}
	return ProposalEvent{
		AuthorID:              userID,
		Title:                 p.Title,
		Description:           p.Description,
		Location:              location,
		CreationDate:          time.Now(),
		Status:                Active,
		MaxConcurrentRequests: uint(p.MaxConcurrentRequests),
		RemainingHelps:        p.MaxConcurrentRequests,
		Tags:                  p.TagsInternal(),
	}
}

type ProposalEventGetResponse struct {
	ID                    uint                  `json:"id"`
	Title                 string                `json:"title"`
	Description           string                `json:"description"`
	CreationDate          string                `json:"creationDate"`
	MaxConcurrentRequests uint                  `json:"maxConcurrentRequests"`
	AvailableHelps        uint                  `json:"availableHelps"`
	CompetitionDate       string                `json:"competitionDate"`
	Status                EventStatus           `json:"status"`
	User                  UserShortInfo         `json:"authorInfo"`
	Comments              []CommentResponse     `json:"comments"`
	Transactions          []TransactionResponse `json:"transactions"`
	Tags                  []TagResponse         `json:"tags"`
}

func (p ProposalEventGetResponse) Bytes() []byte {
	bytes, _ := json.Marshal(p)
	return bytes
}

type ProposalEvents struct {
	ProposalEvents []ProposalEventGetResponse `json:"proposalEvents"`
}

func (l ProposalEvents) Bytes() []byte {
	bytes, _ := json.Marshal(l)
	return bytes
}

type ProposalEventsWithPagination struct {
	ProposalEvents
	Pagination Pagination `json:"pagination"`
}

func (l ProposalEventsWithPagination) Bytes() []byte {
	bytes, _ := json.Marshal(l)
	return bytes
}

type ProposalEventRequestUpdate struct {
	ID                    uint      `json:"id"`
	Title                 string    `json:"title"`
	Description           string    `json:"description"`
	CompetitionDate       time.Time `json:"competitionDate"`
	MaxConcurrentRequests int       `json:"maxConcurrentRequests"`
}

func GetProposalEvents(events ...ProposalEvent) ProposalEvents {
	responseEvents := ProposalEvents{}
	for _, e := range events {
		responseEvents.ProposalEvents = append(responseEvents.ProposalEvents, GetProposalEvent(e))
	}
	return responseEvents
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
			ID:            comment.ID,
			Text:          comment.Text,
			CreationDate:  comment.CreationDate,
			IsUpdated:     comment.IsUpdated,
			UpdateTime:    updatedTime,
			UserShortInfo: comment.UserValues,
		}
	}

	tags := make([]TagResponse, len(event.Tags))
	for i, tag := range event.Tags {
		tags[i] = TagResponse{
			ID:     tag.ID,
			Title:  tag.Title,
			Values: tag.GetTagValuesResponse(),
		}
	}

	tags = append(tags, TagResponse{
		ID:    event.Location.ID,
		Title: "location",
		Values: []string{
			event.Location.Region,
			event.Location.City,
			event.Location.District,
			fmt.Sprintf("%s %s", event.Location.Street, event.Location.HomeLocation),
		},
	})

	transactions := make([]TransactionResponse, len(event.Transactions))
	for i, t := range event.Transactions {
		transaction := TransactionResponse{
			ID:                t.ID,
			CreatorID:         t.CreatorID,
			EventID:           t.EventID,
			Comment:           t.Comment,
			EventType:         t.EventType,
			CreationDate:      t.CreationDate,
			TransactionStatus: t.TransactionStatus,
			ResponderStatus:   t.ResponderStatus,
			Creator:           t.Creator.ToShortInfo(),
			Responder:         t.Responder.ToShortInfo(),
		}
		if t.CompetitionDate.Valid && !t.CompetitionDate.Time.IsZero() {
			transaction.CompetitionDate = t.CompetitionDate.Time
		}
		transactions[i] = transaction
	}
	return ProposalEventGetResponse{
		ID:                    event.ID,
		Title:                 event.Title,
		Description:           event.Description,
		CreationDate:          event.CreationDate.String(),
		CompetitionDate:       completionDate,
		MaxConcurrentRequests: event.MaxConcurrentRequests,
		AvailableHelps:        uint(event.RemainingHelps),
		User: UserShortInfo{
			ID:              event.AuthorID,
			Username:        event.User.FullName,
			ProfileImageURL: event.User.AvatarImagePath,
			PhoneNumber:     Telephone(event.User.Telephone),
		},
		Comments:     comments,
		Transactions: transactions,
		Tags:         tags,
		Status:       event.Status,
	}
}
