package models

//CREATE TABLE IF NOT EXISTS tag (
//    id bigserial PRIMARY KEY,
//    title varchar(255),
//    event_id bigint,
//    event_type event
//);
//
//CREATE TABLE IF NOT EXISTS tag_value (
//    id bigserial PRIMARY KEY,
//    tag_id bigint,
//    value varchar(255),
//    CONSTRAINT tag_id FOREIGN KEY(tag_id) REFERENCES tag(id)
//);

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
