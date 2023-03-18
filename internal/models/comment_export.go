package models

import (
	"encoding/json"
	"io"
	"time"
)

func UnmarshalCommentUpdateRequest(r *io.ReadCloser) (CommentUpdateRequest, error) {
	c := CommentUpdateRequest{}
	err := json.NewDecoder(*r).Decode(&c)
	return c, err
}

func UnmarshalCommentCreateRequest(r *io.ReadCloser) (CommentCreateRequest, error) {
	c := CommentCreateRequest{}
	err := json.NewDecoder(*r).Decode(&c)
	return c, err
}

type CommentResponse struct {
	ID           uint      `json:"id"`
	Text         string    `json:"text"`
	CreationDate time.Time `json:"creationDate"`
	IsUpdated    bool      `json:"isUpdated"`
	UpdateTime   string    `json:"updateTime"`
	UserShortInfo
}

type Comments struct {
	Comments []CommentResponse `json:"comments"`
}

func (c Comments) Bytes() []byte {
	bytes, _ := json.Marshal(c)
	return bytes
}

type CommentCreateRequest struct {
	Text    string `json:"text"`
	EventID uint   `json:"eventId"`
}

type CommentUpdateRequest struct {
	ID   uint   `json:"id"`
	Text string `json:"text"`
}
