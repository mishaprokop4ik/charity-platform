package models

import (
	"encoding/json"
	"io"
	"time"
)

type ComplaintRequest struct {
	Description string    `json:"description"`
	EventType   EventType `json:"eventType"`
	EventID     ID        `json:"eventID"`
}

func (c *ComplaintRequest) Internal(userID int) Complaint {
	return Complaint{
		Description:  c.Description,
		EventType:    c.EventType,
		EventID:      c.EventID,
		CreatedBy:    ID(userID),
		CreationDate: time.Now(),
	}
}

func NewComplaintCreateRequest(r *io.ReadCloser) (ComplaintRequest, error) {
	c := ComplaintRequest{}
	err := json.NewDecoder(*r).Decode(&c)
	return c, err
}

type EventBan struct {
	Type EventType `json:"type"`
	ID   ID        `json:"id"`
}

func NewEventBanCreateRequest(r *io.ReadCloser) (EventBan, error) {
	c := EventBan{}
	err := json.NewDecoder(*r).Decode(&c)
	return c, err
}
