package search

import (
	"Kurajj/internal/models"
	"encoding/json"
	"io"
	"strings"
)

type AllEventsSearch struct {
	Name        string               `json:"name,omitempty"`
	Tags        []models.TagRequest  `json:"tags,omitempty"`
	SortField   string               `json:"sortField,omitempty"`
	TakingPart  bool                 `json:"takingPart,omitempty"`
	StatusState []models.EventStatus `json:"statusStates,omitempty"`
	Location    models.Location      `json:"location,omitempty"`
}

func UnmarshalAllEventsSearch(r *io.ReadCloser) (AllEventsSearch, error) {
	search := AllEventsSearch{}
	err := json.NewDecoder(*r).Decode(&search)
	return search, err
}

func (s AllEventsSearch) GetSearchValues() models.ProposalEventSearchInternal {
	tags := s.convertTagsRequestToInternal()
	name := strings.ToLower(s.Name)
	return models.ProposalEventSearchInternal{
		Name:       &name,
		Tags:       &tags,
		TakingPart: &s.TakingPart,
		State:      s.StatusState,
		SortField:  s.SortField,
		Location: &models.Location{
			Country:  strings.ToLower(s.Location.Country),
			Area:     strings.ToLower(s.Location.Area),
			City:     strings.ToLower(s.Location.City),
			District: strings.ToLower(s.Location.District),
			Street:   strings.ToLower(s.Location.Street),
			Home:     strings.ToLower(s.Location.Home),
		},
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
