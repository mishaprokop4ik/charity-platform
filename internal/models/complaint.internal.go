package models

import "time"

type Complaint struct {
	ID           ID        `gorm:"id"`
	Description  string    `gorm:"description"`
	EventType    EventType `gorm:"event_type"`
	CreatedBy    ID        `gorm:"created_by"`
	EventID      ID        `gorm:"event_id"`
	CreationDate time.Time `gorm:"creation_date"`
}

func (c Complaint) TableName() string {
	return "complaints"

}

type ComplaintsResponse struct {
	EventID        ID                  `json:"eventID"`
	EventType      EventType           `json:"eventType"`
	CreationDate   time.Time           `json:"creationDate"`
	CreatorEventID int                 `json:"creatorID"`
	Complaints     []ComplaintResponse `json:"complaints"`
}
type ComplaintResponse struct {
	Description  string    `json:"description"`
	CreationDate time.Time `json:"creationDate"`
}
