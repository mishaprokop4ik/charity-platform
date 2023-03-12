package models

import (
	"encoding/json"
	"io"
)

type Tag struct {
	ID        uint       `gorm:"primaryKey"`
	Title     string     `gorm:"column:title"`
	EventID   uint       `gorm:"column:event_id"`
	EventType EventType  `gorm:"column:event_type"`
	Values    []TagValue `gorm:"-"`
}

func (t Tag) GetTagValuesResponse() []TagValueResponse {
	values := make([]TagValueResponse, len(t.Values))
	for i := range values {
		values[i].ID = t.Values[i].ID
		values[i].Value = t.Values[i].Value
	}

	return values
}

type TagRequest struct {
	ID        uint      `json:"id,omitempty"`
	Title     string    `json:"title,omitempty"`
	EventID   uint      `json:"eventID,omitempty"`
	EventType EventType `json:"eventType,omitempty"`
	Values    []string  `json:"values,omitempty"`
}
type TagsResponse struct {
	ID        uint               `json:"id,omitempty"`
	Title     string             `json:"title,omitempty"`
	EventID   uint               `json:"eventID,omitempty"`
	EventType EventType          `json:"eventType,omitempty"`
	Values    []TagValueResponse `json:"values,omitempty"`
}

type Tags struct {
	Tags []TagsResponse `json:"tags"`
}

func (t Tags) Bytes() []byte {
	bytes, _ := json.Marshal(t)
	return bytes
}

type TagRequestCreate struct {
	Title  string   `json:"title,omitempty"`
	Values []string `json:"values,omitempty"`
}

type TagGroupRequestCreate struct {
	Tags      []TagRequestCreate `json:"tags"`
	EventID   uint               `json:"eventID"`
	EventType EventType          `json:"eventType"`
}

func (t TagGroupRequestCreate) Internal() []Tag {
	tags := make([]Tag, len(t.Tags))
	for i := 0; i < len(t.Tags); i++ {

		tagValues := make([]TagValue, len(t.Tags[i].Values))
		for j, value := range t.Tags[i].Values {
			tagValues[j] = TagValue{
				Value: value,
			}
		}

		tags[i] = Tag{
			Title:     t.Tags[i].Title,
			EventID:   t.EventID,
			EventType: t.EventType,
			Values:    tagValues,
		}
	}
	return tags
}

func UnmarshalTagGroupCreateRequest(r *io.ReadCloser) (TagGroupRequestCreate, error) {
	tags := TagGroupRequestCreate{}
	err := json.NewDecoder(*r).Decode(&tags)
	return tags, err
}

type TagValueRequest struct {
	ID    uint   `json:"id,omitempty"`
	TagID uint   `json:"tagID,omitempty"`
	Value string `json:"value,omitempty"`
}

func (Tag) TableName() string {
	return "tag"
}

type TagValue struct {
	ID    uint   `gorm:"primaryKey"`
	TagID uint   `gorm:"column:tag_id"`
	Value string `gorm:"column:value"`
}

func (TagValue) TableName() string {
	return "tag_value"
}

type TagValueResponse struct {
	ID    uint   `json:"id,omitempty"`
	Value string `json:"value"`
}

type TagResponse struct {
	ID     uint               `json:"id,omitempty"`
	Title  string             `json:"title,omitempty"`
	Values []TagValueResponse `json:"values,omitempty"`
}
