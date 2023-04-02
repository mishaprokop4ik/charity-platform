package repository

import (
	"Kurajj/internal/models"
	"context"
	"errors"
	"gorm.io/gorm"
)

type Tagger interface {
	UpsertTags(ctx context.Context, eventType models.EventType, eventID uint, tags []models.Tag) error
	GetTagsByEvent(ctx context.Context, eventID uint, eventType models.EventType) ([]models.Tag, error)
	DeleteAllTagsByEvent(ctx context.Context, eventID uint, eventType models.EventType) error
	CreateTag(ctx context.Context, tag models.Tag) error
}

type Tag struct {
	DBConnector *Connector
}

func (t *Tag) CreateTag(ctx context.Context, tag models.Tag) error {
	tx := t.DBConnector.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	err := t.DBConnector.DB.Create(&tag).WithContext(ctx).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, tagValue := range tag.Values {
		tagValue.TagID = tag.ID
		err = t.DBConnector.DB.Create(&tagValue).WithContext(ctx).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return nil
}

func (t *Tag) UpsertTags(ctx context.Context, eventType models.EventType, eventID uint, tags []models.Tag) error {
	// TODO add real upsert
	tx := t.DBConnector.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	err := t.DeleteAllTagsByEvent(ctx, eventID, eventType)
	if err != nil {
		tx.Rollback()
		return err
	}
	for _, tag := range tags {
		if tag.Title == "location" {
			err = tx.Model(&models.Address{}).
				Where("event_type = ?", eventType).
				Where("event_id = ?", eventID).
				Delete(&models.Address{}).
				WithContext(ctx).
				Error
			if err != nil && !errors.Is(gorm.ErrRecordNotFound, err) {
				tx.Rollback()
				return err
			}
			if len(tag.Values) >= models.DecodedAddressLength {
				location := models.Address{
					Region:       tag.Values[0].Value,
					City:         tag.Values[1].Value,
					District:     tag.Values[2].Value,
					HomeLocation: tag.Values[3].Value,
					EventType:    eventType,
					EventID:      eventID,
				}
				err = tx.Create(&location).WithContext(ctx).Error
				if err != nil {
					return err
				}
			}

			continue
		}
		err = t.CreateTag(ctx, tag)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

func (t *Tag) GetTagsByEvent(ctx context.Context, eventID uint, eventType models.EventType) ([]models.Tag, error) {
	tags := []models.Tag{}
	err := t.DBConnector.
		DB.
		Where("event_type = ?", eventType).
		Where("event_id = ?", eventID).
		Find(&tags).
		WithContext(ctx).
		Error
	if err != nil {
		return nil, err
	}
	for i, tag := range tags {
		tagValues := []models.TagValue{}
		err = t.DBConnector.
			DB.
			Where("tag_id = ?", tag.ID).
			Find(&tagValues).
			WithContext(ctx).
			Error
		if err != nil {
			return nil, err
		}
		tags[i].Values = tagValues
	}
	return tags, nil
}

func (t *Tag) DeleteAllTagsByEvent(ctx context.Context, eventID uint, eventType models.EventType) error {
	err := t.DBConnector.DB.
		Where("event_type = ?", eventType).
		Where("event_id = ?", eventID).
		Delete(&models.Tag{}).
		WithContext(ctx).
		Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return nil
}

func NewTag(DBConnector *Connector) *Tag {
	return &Tag{DBConnector: DBConnector}
}
