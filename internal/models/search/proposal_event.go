package search

import (
	"Kurajj/internal/models"
	"encoding/json"
	"net/http"
	"strings"
)

type AllEventsSearch struct {
	Name        string              `json:"name,omitempty"`
	Tags        []models.TagRequest `json:"tags,omitempty"`
	SortField   string              `json:"sortField,omitempty"`
	StatusState models.EventStatus  `json:"statusState,omitempty"`
}

func UnmarshalAllEventsSearch(r *http.Request) (AllEventsSearch, error) {
	search := AllEventsSearch{}
	err := json.NewDecoder(r.Body).Decode(&search)
	return search, err
}

func (s AllEventsSearch) GetSearchValues() models.ProposalEventSearchInternal {
	tags := s.convertTagsRequestToInternal()
	name := strings.ToLower(s.Name)
	return models.ProposalEventSearchInternal{
		Name:      &name,
		Tags:      &tags,
		SortField: s.SortField,
	}
}

func (s AllEventsSearch) convertTagsRequestToInternal() []models.Tag {
	tags := make([]models.Tag, len(s.Tags))
	for i, tag := range s.Tags {
		tags[i] = models.Tag{
			ID:        tag.ID,
			Title:     strings.ToLower(tag.Title),
			EventID:   tag.EventID,
			EventType: tag.EventType,
			Values:    s.getTagFromStrings(tag.ID, tag.Values...),
		}
	}
	return tags
}

func (s AllEventsSearch) getTagFromStrings(tagID uint, values ...string) []models.TagValue {
	tagValues := make([]models.TagValue, len(values))
	for i, value := range values {
		tagValues[i] = models.TagValue{
			TagID: tagID,
			Value: strings.ToLower(value),
		}
	}

	return tagValues
}

type OwnEventsSearch struct {
	Name        string             `json:"name,omitempty"`
	StatusState models.EventStatus `json:"statusState,omitempty"`
	SortField   string             `json:"sortField,omitempty"`
}

type TakingPartEventsSearch struct {
	Name        string             `json:"name,omitempty"`
	StatusState models.EventStatus `json:"statusState,omitempty"`
	SortField   string             `json:"sortField,omitempty"`
}
