package search

import (
	"Kurajj/internal/models"
	"encoding/json"
	"io"
	"strings"
)

type AllEventsSearch struct {
	Name        string              `json:"name,omitempty"`
	Tags        []models.TagRequest `json:"tags,omitempty"`
	SortField   string              `json:"sortField,omitempty"`
	Order       models.Order        `json:"order,omitempty"`
	TakingPart  bool                `json:"takingPart,omitempty"`
	StatusState models.EventStatus  `json:"statusStates,omitempty"`
}

func UnmarshalAllEventsSearch(r *io.ReadCloser) (AllEventsSearch, error) {
	search := AllEventsSearch{}
	err := json.NewDecoder(*r).Decode(&search)
	return search, err
}

func (s AllEventsSearch) GetSearchValues() models.ProposalEventSearchInternal {
	tags := s.convertTagsRequestToInternal()
	name := strings.ToLower(s.Name)
	if s.Order == "" {
		s.Order = models.AscendingOrder
	}
	location := models.Address{}
	for i, tag := range tags {
		if strings.ToLower(tag.Title) == "location" ||
			strings.ToLower(tag.Title) == "place" &&
				len(tag.Values) > 4 {
			location.Region = tag.Values[0].Value
			location.City = tag.Values[1].Value
			location.District = tag.Values[2].Value
			location.HomeLocation = tag.Values[3].Value
		}
		tag.Values = append(tag.Values[:i], tag.Values[i+1:]...)
	}
	if s.StatusState == "" {
		s.StatusState = models.Active
	}
	return models.ProposalEventSearchInternal{
		Name:       &name,
		Tags:       &tags,
		TakingPart: &s.TakingPart,
		State: []models.EventStatus{
			s.StatusState,
		},
		Order:     &s.Order,
		SortField: s.SortField,
		Location:  &location,
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
