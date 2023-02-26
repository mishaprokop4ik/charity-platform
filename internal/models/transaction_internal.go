package models

import (
	"database/sql"
	"reflect"
	"strings"
	"time"
)

type Transaction struct {
	ID                uint         `gorm:"primaryKey"`
	CreatorID         uint         `gorm:"column:creator_id"`
	CompetitionDate   sql.NullTime `gorm:"column:competition_date"`
	EventID           uint         `gorm:"column:event_id"`
	Comment           string       `gorm:"column:last_comment"`
	EventType         EventType    `gorm:"column:event_type"`
	TransactionStatus Status       `gorm:"column:transaction_status"`
	ResponderStatus   Status       `gorm:"column:responder_status"`
}

func (t Transaction) GetValuesToUpdate() map[string]any {
	getProposalEventTag := func(f reflect.StructField, tagName string) string {
		tag := strings.Split(f.Tag.Get(tagName), ":")
		if len(tag) != 2 {
			return ""
		}
		return tag[1]
	}

	isTimeZero := func(t any) bool {
		timeValue, ok := t.(time.Time)
		if !timeValue.IsZero() || !ok {
			return false
		}
		return true
	}

	updateValues := make(map[string]any)

	proposalEvent := reflect.TypeOf(t)
	proposalEventFields := reflect.ValueOf(t)
	proposalEventFieldsCount := proposalEvent.NumField()
	for i := 0; i < proposalEventFieldsCount; i++ {
		field := proposalEvent.Field(i)
		value := proposalEventFields.Field(i).Interface()
		fieldName := getProposalEventTag(field, "gorm")
		if !proposalEventFields.Field(i).IsZero() &&
			!isTimeZero(proposalEventFields.Field(i).Interface()) &&
			fieldName != "" {
			updateValues[fieldName] = value
		}
	}

	return updateValues
}
