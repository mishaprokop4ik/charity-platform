package models

import (
	"database/sql"
	"reflect"
	"strings"
	"time"
)

type ProposalEvent struct {
	ID                    uint          `gorm:"primaryKey"`
	Title                 string        `gorm:"column:title"`
	Description           string        `gorm:"column:description"`
	CreationDate          time.Time     `gorm:"column:creation_date"`
	CompetitionDate       sql.NullTime  `gorm:"column:competition_date"`
	AuthorID              uint          `gorm:"column:author_id"`
	Category              string        `gorm:"column:category"`
	MaxConcurrentRequests uint          `gorm:"column:max_concurrent_requests"`
	RemainingHelps        int           `gorm:"column:remaining_helps"`
	IsDeleted             bool          `gorm:"column:is_deleted"`
	Comments              []Comment     `gorm:"-"`
	Transactions          []Transaction `gorm:"-"`
}

func (p ProposalEvent) TableName() string {
	return "propositional_event"
}

func (p ProposalEvent) GetValuesToUpdate() map[string]any {
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

	proposalEvent := reflect.TypeOf(p)
	proposalEventFields := reflect.ValueOf(p)
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
