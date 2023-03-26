package models

import (
	"encoding/json"
	"time"
)

type NotificationResponse struct {
	ID        uint      `json:"id"`
	Text      string    `json:"text"`
	IsRead    bool      `json:"isRead"`
	EventID   uint      `json:"eventID"`
	EventName string    `json:"eventName"`
	NewStatus string    `json:"newStatus"`
	CreatedAt time.Time `json:"createdAt"`
}

type Notifications struct {
	Notifications []NotificationResponse `json:"notifications"`
}

func (n Notifications) Bytes() []byte {
	bytes, _ := json.Marshal(n)
	return bytes
}
