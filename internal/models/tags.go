package models

type Tag struct {
	ID        uint      `gorm:"primaryKey"`
	Title     string    `gorm:"column:title"`
	EventID   uint      `gorm:"column:event_id"`
	EventType EventType `gorm:"column:event_type"`
	Values    []TagValue
}

type TagRequest struct {
	ID        uint      `json:"id,omitempty"`
	Title     string    `json:"title,omitempty"`
	EventID   uint      `json:"eventID,omitempty"`
	EventType EventType `json:"eventType,omitempty"`
	Values    []string  `json:"values,omitempty"`
}

type TagRequestCreate struct {
	Title  string   `json:"title,omitempty"`
	Values []string `json:"values,omitempty"`
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
	Value string `json:"value,omitempty"`
}

type TagResponse struct {
	ID     uint               `json:"id,omitempty"`
	Title  string             `json:"title,omitempty"`
	Values []TagValueResponse `json:"values,omitempty"`
}
