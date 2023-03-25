package models

import (
	"database/sql"
	"reflect"
	"strings"
	"time"
)

type Transaction struct {
	ID                uint              `gorm:"primaryKey"`
	CreatorID         uint              `gorm:"column:creator_id"`
	Creator           User              `gorm:"-"`
	Responder         User              `gorm:"-"`
	CompetitionDate   sql.NullTime      `gorm:"column:completion_date"`
	EventID           uint              `gorm:"column:event_id"`
	Comment           string            `gorm:"column:comment"`
	CreationDate      time.Time         `gorm:"column:creation_date"`
	EventType         EventType         `gorm:"column:event_type"`
	TransactionStatus TransactionStatus `gorm:"column:transaction_status"`
	ResponderStatus   TransactionStatus `gorm:"column:responder_status"`
	ReportURL         string            `gorm:"column:report_url"`
}

func (Transaction) TableName() string {
	return "transaction"
}

func (t Transaction) GetValuesToUpdate() map[string]any {
	getTransactionTag := func(f reflect.StructField, tagName string) string {
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
		fieldName := getTransactionTag(field, "gorm")
		if !proposalEventFields.Field(i).IsZero() &&
			!isTimeZero(proposalEventFields.Field(i).Interface()) &&
			fieldName != "" {
			updateValues[fieldName] = value
		}
	}

	return updateValues
}
