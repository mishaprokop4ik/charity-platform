package models

import (
	"database/sql"
	"reflect"
	"strings"
	"time"
)

type ProposalEvent struct {
	ID              uint   `gorm:"primaryKey"`
	Title           string `gorm:"column:title"`
	Description     string
	CreationDate    time.Time
	CompetitionDate sql.NullTime `gorm:"column:completion_date"`
	AuthorID        uint         `gorm:"column:author_id"`
	Category        string       `gorm:"column:category"`
	Comments        []Comment
	Transactions    []Transaction
}

func (p ProposalEvent) GetValuesToUpdate() map[string]any {
	getProposalEventTag := func(f reflect.StructField, tagName string) string {
		tag := strings.Split(f.Tag.Get(tagName), ":")
		if len(tag) != 2 {
			return ""
		}
		return tag[1]
	}
	updateValues := make(map[string]any)

	proposalEvent := reflect.TypeOf(p)
	proposalEventFields := reflect.ValueOf(p)
	proposalEventFieldsCount := proposalEvent.NumField()
	for i := 0; i < proposalEventFieldsCount; i++ {
		field := proposalEvent.Field(i)
		value := proposalEventFields.Field(i).Interface()
		if !proposalEventFields.Field(i).IsZero() {
			updateValues[getProposalEventTag(field, "gorm")] = value
		}
	}

	return updateValues
}

type Comment struct {
	ID           uint `gorm:"primaryKey"`
	EventID      uint
	EventType    EventType
	Text         string
	UserID       uint
	CreationDate time.Time
}
