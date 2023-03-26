package models

import (
	"encoding/json"
	"fmt"
	"time"
)

type NotificationResponse struct {
	ID         uint      `json:"id"`
	Text       string    `json:"text"`
	IsRead     bool      `json:"isRead"`
	EventID    uint      `json:"eventID"`
	EventTitle string    `json:"eventTitle"`
	NewStatus  string    `json:"newStatus"`
	CreatedAt  time.Time `json:"createdAt"`
}

type Notifications struct {
	Notifications []NotificationResponse `json:"notifications"`
}

func (n Notifications) Bytes() []byte {
	bytes, _ := json.Marshal(n)
	return bytes
}

func GenerateNotificationResponses(notifications []TransactionNotification) []NotificationResponse {
	newNotifications := make([]NotificationResponse, len(notifications))
	for i := range notifications {
		newNotifications[i] = GenerateNotificationResponse(notifications[i])
	}

	return newNotifications
}

func GenerateNotificationResponse(notification TransactionNotification) NotificationResponse {
	notificationType := ""
	switch notification.EventType {
	case ProposalEventType:
		notificationType = "Propositional Event"
	}
	text := ""
	switch notification.Action {
	case Created:
		text = fmt.Sprintf("%s transaction was created in %s event.", notificationType, notification.EventTitle)
	case Updated:
		text = fmt.Sprintf("%s status changed to %s in %s event.", notificationType, notification.NewStatus, notification.EventTitle)
	}
	return NotificationResponse{
		ID:         notification.ID,
		Text:       text,
		IsRead:     notification.IsRead,
		EventID:    notification.EventID,
		EventTitle: notification.EventTitle,
		NewStatus:  string(notification.NewStatus),
		CreatedAt:  notification.CreationTime,
	}
}
