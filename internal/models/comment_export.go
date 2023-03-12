package models

import (
	"encoding/json"
	"io"
	"time"
)

type CommentResponse struct {
	ID           uint      `json:"id,omitempty"`
	Text         string    `json:"text,omitempty"`
	CreationDate time.Time `json:"creationDate"`
	IsUpdated    bool      `json:"isUpdated,omitempty"`
	UpdateTime   string    `json:"updateTime"`
	UserShortInfo
}

type Comments struct {
	Comments []CommentResponse `json:"comments,omitempty"`
}

func (c Comments) Bytes() []byte {
	bytes, _ := json.Marshal(c)
	return bytes
}

type CommentCreateRequest struct {
	Text    string `json:"text,omitempty"`
	EventID uint   `json:"eventId,omitempty"`
}

type CommentUpdateRequest struct {
	ID   uint   `json:"id,omitempty"`
	Text string `json:"text,omitempty"`
}

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
