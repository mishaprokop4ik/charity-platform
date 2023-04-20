package models

import "encoding/json"

type HelpSearchInternal struct {
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

func (i HelpSearchInternal) GetTagsValues() []string {
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

func (i HelpSearchInternal) GetTagsTitle() []string {
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

type HelpEventPagination struct {
	Events     []HelpEvent
	Pagination Pagination
}

func (l HelpEventsWithPagination) Bytes() []byte {
	bytes, _ := json.Marshal(l)
	return bytes
}

func GetHelpEventItems(events ...HelpEvent) HelpEventsItems {
	helpEvents := HelpEventsItems{
		HelpEvents: make([]HelpEventResponse, len(events)),
	}
	for i := range events {
		helpEvents.HelpEvents[i] = events[i].Response()
	}

	return helpEvents
}
