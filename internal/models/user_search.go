package models

import (
	"encoding/json"
	"io"
)

func UnmarshalSearchValuesGroupCreateRequest(r *io.ReadCloser) (MemberSearchValuesRequestCreate, error) {
	tags := MemberSearchValuesRequestCreate{}
	err := json.NewDecoder(*r).Decode(&tags)
	return tags, err
}

type MemberSearch struct {
	ID     uint          `gorm:"primaryKey"`
	Title  string        `gorm:"column:title"`
	UserID uint          `gorm:"column:member_id"`
	Values []SearchValue `gorm:"-"`
}

func (s MemberSearch) Response() SearchValueResponse {
	resp := SearchValueResponse{
		Values: make([]MemberSearchValueResponse, len(s.Values)),
	}
	resp.Title = s.Title
	resp.ID = s.ID
	for i, searchValue := range s.Values {
		resp.Values[i] = MemberSearchValueResponse{
			ID:    searchValue.ID,
			Value: searchValue.Value,
		}
	}

	return resp
}

func (MemberSearch) TableName() string {
	return "member_search"
}

func (s MemberSearch) GetTagValuesResponse() []MemberSearchValueResponse {
	values := make([]MemberSearchValueResponse, len(s.Values))
	for i := range values {
		values[i].ID = s.Values[i].ID
		values[i].Value = s.Values[i].Value
	}

	return values
}

type SearchValueResponse struct {
	ID     uint                        `json:"id"`
	Title  string                      `json:"title"`
	Values []MemberSearchValueResponse `json:"values"`
}

type SearchValuesResponse struct {
	Values []SearchValueResponse `json:"searchValues"`
}

func (s SearchValuesResponse) Bytes() []byte {
	bytes, _ := json.Marshal(s)
	return bytes
}

type MemberSearchRequestCreate struct {
	Title  string   `json:"title"`
	Values []string `json:"values"`
}

type MemberSearchValuesRequestCreate struct {
	Tags []MemberSearchRequestCreate `json:"tags"`
}

func (t MemberSearchValuesRequestCreate) Internal() []MemberSearch {
	search := make([]MemberSearch, len(t.Tags))
	for i := 0; i < len(t.Tags); i++ {

		searchValues := make([]SearchValue, len(t.Tags[i].Values))
		for j, value := range t.Tags[i].Values {
			searchValues[j] = SearchValue{
				Value: value,
			}
		}

		search[i] = MemberSearch{
			Title:  t.Tags[i].Title,
			Values: searchValues,
		}
	}
	return search
}

type SearchValue struct {
	ID       uint   `gorm:"primaryKey"`
	SearchID uint   `gorm:"column:member_search_id"`
	Value    string `gorm:"column:value"`
}

func (SearchValue) TableName() string {
	return "member_search_value"
}

type MemberSearchValueResponse struct {
	ID    uint   `json:"id"`
	Value string `json:"value"`
}
