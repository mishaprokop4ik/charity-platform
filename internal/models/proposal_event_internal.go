package models

import (
	"database/sql"
	"encoding/json"
	"io"
	"reflect"
	"strings"
	"time"
)

type Order string

var (
	AscendingOrder Order = "ASC"
	DecreaseOrder  Order = "DESC"
)

type ProposalEvent struct {
	ID                    uint          `gorm:"primaryKey"`
	Title                 string        `gorm:"column:title"`
	Description           string        `gorm:"column:description"`
	CreationDate          time.Time     `gorm:"column:creation_date"`
	CompetitionDate       sql.NullTime  `gorm:"column:competition_date"`
	AuthorID              uint          `gorm:"column:author_id"`
	EndDate               time.Time     `gorm:"column:end_date"`
	Status                EventStatus   `gorm:"column:status"`
	MaxConcurrentRequests uint          `gorm:"column:max_concurrent_requests"`
	RemainingHelps        int           `gorm:"column:remaining_helps"`
	IsDeleted             bool          `gorm:"column:is_deleted"`
	ImagePath             string        `gorm:"column:image_path"`
	Banned                bool          `gorm:"column:is_banned"`
	FileType              string        `gorm:"-"`
	File                  io.Reader     `gorm:"-"`
	Comments              []Comment     `gorm:"-"`
	Transactions          []Transaction `gorm:"-"`
	Tags                  []Tag         `gorm:"-"`
	Location              Address       `gorm:"-"`
	User                  User          `gorm:"-"`
}

func (p ProposalEvent) TableName() string {
	return "propositional_event"
}

func (p ProposalEvent) GetValuesToUpdate() map[string]any {
	getProposalEventTag := func(f reflect.StructField, tagName string) string {
		tag := strings.Split(f.Tag.Get(tagName), ":")
		if len(tag) != 2 {
			return ""
		}
		return tag[1]
	}

	isTimeZero := func(t any) bool {
		timeValue, ok := t.(time.Time)
		if !timeValue.IsZero() || !ok {
			return false
		}
		return true
	}

	updateValues := make(map[string]any)

	proposalEvent := reflect.TypeOf(p)
	proposalEventFields := reflect.ValueOf(p)
	proposalEventFieldsCount := proposalEvent.NumField()
	for i := 0; i < proposalEventFieldsCount; i++ {
		field := proposalEvent.Field(i)
		value := proposalEventFields.Field(i).Interface()
		fieldName := getProposalEventTag(field, "gorm")
		if !proposalEventFields.Field(i).IsZero() &&
			!isTimeZero(proposalEventFields.Field(i).Interface()) &&
			fieldName != "" {
			updateValues[fieldName] = value
		}
	}

	return updateValues
}

type ProposalEventsInternal struct {
	ProposalEvents []ProposalEventGetResponse `json:"proposalEvents"`
}

func (p ProposalEventsInternal) Serialize() ([]byte, error) {
	decodedEvent, err := json.Marshal(p)
	return decodedEvent, err
}

type ProposalEventSearchInternal struct {
	Name             *string
	Tags             *[]Tag
	SortField        string
	Order            *Order
	SearcherID       *uint
	State            []EventStatus
	TakingPart       *bool
	Location         *Address
	Pagination       PaginationRequest
	AllowTitleSearch *bool
}

func (i ProposalEventSearchInternal) GetTagsValues() []string {
	if i.Tags == nil {
		return []string{}
	}

	values := make([]string, 0)
	for _, tag := range *i.Tags {
		tagValues := tag.Values
		for _, tagValue := range tagValues {
			if tagValue.Value != "" {
				values = append(values, tagValue.Value)
			}
		}
	}

	return values
}

func (i ProposalEventSearchInternal) GetTagsTitle() []string {
	if i.Tags == nil {
		return []string{}
	}

	titles := make([]string, 0)
	for _, tag := range *i.Tags {
		tagTitle := tag.Title
		titles = append(titles, tagTitle)
	}

	return titles
}

type PaginationRequest struct {
	PageNumber int
	PageSize   int
}

type Pagination struct {
	TotalRecords int64 `json:"totalRecords"`
	TotalPage    int   `json:"totalPageCount"`
	Offset       int   `json:"offset"`
	Limit        int   `json:"limit"`
	Page         int   `json:"pageNumber"`
	PrevPage     int   `json:"previousPage"`
	NextPage     int   `json:"nextPage"`
}

type ProposalEventPagination struct {
	Events     []ProposalEvent
	Pagination Pagination
}
