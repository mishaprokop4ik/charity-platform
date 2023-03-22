package search

import (
	"Kurajj/internal/models"
	"encoding/json"
	"io"
	"strings"
)

func UnmarshalAllEventsSearch(r *io.ReadCloser) (AllEventsSearch, error) {
	search := AllEventsSearch{}
	err := json.NewDecoder(*r).Decode(&search)
	return search, err
}

type AllEventsSearch struct {
	Name        string              `json:"name"`
	Tags        []models.TagRequest `json:"tags"`
	SortField   string              `json:"sortField"`
	Order       models.Order        `json:"order"`
	TakingPart  bool                `json:"takingPart"`
	StatusState models.EventStatus  `json:"statusStates"`
	PageNumber  int                 `json:"pageNumber"`
	PageSize    int                 `json:"pageSize"`
}

func (s AllEventsSearch) Internal() models.ProposalEventSearchInternal {
	tags := s.convertTagsRequestToInternal()
	name := strings.ToLower(s.Name)
	if s.Order == "" {
		s.Order = models.AscendingOrder
	}
	location := models.Address{}
	for i, tag := range tags {
		if strings.ToLower(tag.Title) == "location" ||
			strings.ToLower(tag.Title) == "place" &&
				len(tag.Values) >= models.DecodedAddressLength {
			if len(tag.Values[0].Value) != 0 {
				location.Region = tag.Values[0].Value
			}
			if len(tag.Values[1].Value) != 0 {
				location.City = tag.Values[1].Value
			}
			if len(tag.Values[2].Value) != 0 {
				location.District = tag.Values[2].Value
			}
			if len(tag.Values[3].Value) != 0 {
				location.HomeLocation = tag.Values[3].Value
			}
			if len(tag.Values[0].Value) != 0 &&
				len(tag.Values[1].Value) != 0 &&
				len(tag.Values[2].Value) != 0 &&
				len(tag.Values[3].Value) != 0 {
				tags = append(tags[:i], tags[i+1:]...)
			}
		}
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
		Pagination: models.PaginationRequest{
			PageSize:   s.PageSize,
			PageNumber: s.PageNumber,
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
