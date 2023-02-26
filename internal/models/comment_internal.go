package models

import (
	"database/sql"
	"reflect"
	"strings"
	"time"
)

type Comment struct {
	ID           uint         `gorm:"primaryKey" gorm:"column:id"`
	EventID      uint         `gorm:"column:event_id"`
	EventType    EventType    `gorm:"column:event_type"`
	Text         string       `gorm:"column:text"`
	UserID       uint         `gorm:"column:user_id"`
	CreationDate time.Time    `gorm:"column:creation_date"`
	IsUpdated    bool         `gorm:"column:is_updated"`
	UpdateTime   sql.NullTime `gorm:"column:update_time"`
	IsDeleted    bool         `gorm:"column:is_deleted"`
}

func (c Comment) GetValuesToUpdate() map[string]any {
	getCommentTag := func(f reflect.StructField, tagName string) string {
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

	proposalEvent := reflect.TypeOf(c)
	proposalEventFields := reflect.ValueOf(c)
	proposalEventFieldsCount := proposalEvent.NumField()
	for i := 0; i < proposalEventFieldsCount; i++ {
		field := proposalEvent.Field(i)
		value := proposalEventFields.Field(i).Interface()
		fieldName := getCommentTag(field, "gorm")
		if !proposalEventFields.Field(i).IsZero() &&
			!isTimeZero(proposalEventFields.Field(i).Interface()) &&
			fieldName != "" {
			updateValues[fieldName] = value
		}
	}

	return updateValues
}
