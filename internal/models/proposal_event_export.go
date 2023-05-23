package models

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/url"
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
	EndDate               time.Time    `json:"endDate"`
	MaxConcurrentRequests int          `json:"maxConcurrentRequests"`
	FileBytes             []byte       `json:"fileBytes"`
	FileType              string       `json:"fileType"`
	FilePath              string       `json:"imagePath"`
	Tags                  []TagRequest `json:"tags"`
}

func (p *ProposalEventRequestCreate) TagsInternal() []Tag {
	tags := make([]Tag, len(p.Tags))
	for i, tag := range p.Tags {
		tagValues := make([]TagValue, 0)
		for _, tagValue := range tag.Values {
			if tagValue != "" {
				tagValues = append(tagValues, TagValue{
					Value: tagValue,
				})
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
			if len(t.Values[0]) != 0 {
				location.Region = t.Values[0]
			}
			if len(t.Values[1]) != 0 {
				location.City = t.Values[1]
			}
			if len(t.Values[2]) != 0 {
				location.District = t.Values[2]
			}
			if len(t.Values[3]) != 0 {
				location.HomeLocation = t.Values[3]
			}
			location.EventType = ProposalEventType
			p.Tags = append(p.Tags[:i], p.Tags[i+1:]...)
		}
	}

	event := ProposalEvent{
		AuthorID:              userID,
		Title:                 p.Title,
		Description:           p.Description,
		Location:              location,
		CreationDate:          time.Now(),
		EndDate:               p.EndDate,
		Status:                Active,
		ImagePath:             p.FilePath,
		MaxConcurrentRequests: uint(p.MaxConcurrentRequests),
		RemainingHelps:        p.MaxConcurrentRequests,
		Tags:                  p.TagsInternal(),
	}

	_, err := url.ParseRequestURI(event.ImagePath)
	if (len(p.FileBytes) == 0 || p.FileType == "") && err != nil {
		event.ImagePath = defaultImagePath
	} else if len(p.FileBytes) != 0 && p.FileType != "" {
		event.FileType = p.FileType
		event.File = bytes.NewBuffer(p.FileBytes)
	}
	return event
}

type ProposalEventGetResponse struct {
	ID                    uint                  `json:"id"`
	Title                 string                `json:"title"`
	Description           string                `json:"description"`
	CreationDate          string                `json:"creationDate"`
	EndDate               time.Time             `json:"endDate"`
	MaxConcurrentRequests uint                  `json:"maxConcurrentRequests"`
	AvailableHelps        uint                  `json:"availableHelps"`
	CompetitionDate       string                `json:"competitionDate"`
	Status                EventStatus           `json:"status"`
	Image                 string                `json:"imageURL"`
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

type ProposalEventsItems struct {
	ProposalEvents []ProposalEventGetResponse `json:"items"`
}

func (l ProposalEvents) Bytes() []byte {
	bytes, _ := json.Marshal(l)
	return bytes
}

type ProposalEventsWithPagination struct {
	ProposalEventsItems
	Pagination
}

func (l ProposalEventsWithPagination) Bytes() []byte {
	bytes, _ := json.Marshal(l)
	return bytes
}

type ProposalEventRequestUpdate struct {
	ID                    uint        `json:"id"`
	Title                 string      `json:"title"`
	Description           string      `json:"description"`
	CompetitionDate       time.Time   `json:"competitionDate"`
	Status                EventStatus `json:"status"`
	EndDate               time.Time   `json:"endDate"`
	FileBytes             []byte      `json:"fileBytes"`
	FileType              string      `json:"fileType"`
	MaxConcurrentRequests int         `json:"maxConcurrentRequests"`
}

func (p *ProposalEventRequestUpdate) Internal() ProposalEvent {
	event := ProposalEvent{
		ID:          p.ID,
		Title:       p.Title,
		Description: p.Description,
		Status:      p.Status,
		EndDate:     p.EndDate,
		CompetitionDate: sql.NullTime{
			Time: p.CompetitionDate,
		},
		MaxConcurrentRequests: uint(p.MaxConcurrentRequests),
	}
	if len(p.FileBytes) != 0 {
		event.File = bytes.NewReader(p.FileBytes)
		event.FileType = p.FileType
	}

	return event
}

func GetProposalEvents(events ...ProposalEvent) ProposalEvents {
	responseEvents := ProposalEvents{
		ProposalEvents: make([]ProposalEventGetResponse, 0),
	}
	for _, e := range events {
		responseEvents.ProposalEvents = append(responseEvents.ProposalEvents, GetProposalEvent(e))
	}

	return responseEvents
}

func GetProposalEventItems(events ...ProposalEvent) ProposalEventsItems {
	responseEvents := ProposalEventsItems{
		ProposalEvents: make([]ProposalEventGetResponse, 0),
	}
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
	homeLocation := ""
	if event.Location.Street != "" {
		homeLocation = event.Location.Street
	}
	if event.Location.HomeLocation != "" && event.Location.Street != "" {
		homeLocation = homeLocation + " " + event.Location.HomeLocation
	} else if event.Location.HomeLocation != "" && event.Location.Street == "" {
		homeLocation = event.Location.HomeLocation
	}
	tags = append(tags, TagResponse{
		ID:    event.Location.ID,
		Title: "location",
		Values: []string{
			event.Location.Region,
			event.Location.City,
			event.Location.District,
			homeLocation,
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
			ReportURL:         t.ReportURL,
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
		EndDate:               event.EndDate,
		AvailableHelps:        uint(event.RemainingHelps),
		User: UserShortInfo{
			ID:              event.AuthorID,
			Username:        event.User.FullName,
			ProfileImageURL: event.User.AvatarImagePath,
			PhoneNumber:     Telephone(event.User.Telephone),
		},
		Image:        event.ImagePath,
		Comments:     comments,
		Transactions: transactions,
		Tags:         tags,
		Status:       event.Status,
	}
}
