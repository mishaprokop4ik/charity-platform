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

type TransactionAcceptCreateRequest struct {
	ID      int    `json:"id"`
	Comment string `json:"comment"`
}

func UnmarshalTransactionAcceptCreateRequest(b *io.ReadCloser) (TransactionAcceptCreateRequest, error) {
	r := TransactionAcceptCreateRequest{}
	err := json.NewDecoder(*b).Decode(&r)
	return r, err
}

type TransactionAcceptRequest struct {
	IsAccepted bool `json:"isAccepted"`
}

func UnmarshalTransactionAcceptRequest(b *io.ReadCloser) (TransactionAcceptRequest, error) {
	r := TransactionAcceptRequest{}
	err := json.NewDecoder(*b).Decode(&r)
	return r, err
}

type AcceptRequest struct {
	Accept        bool
	TransactionID uint
}

type CommentUpdateRequest struct {
	ID   uint   `json:"id"`
	Text string `json:"text"`
}
