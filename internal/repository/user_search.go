package repository

import (
	"Kurajj/internal/models"
	"context"
	"errors"
	"gorm.io/gorm"
)

type UserSearcher interface {
	UpsertTags(ctx context.Context, userID uint, searchValues []models.MemberSearch) error
}

type UserSearch struct {
	DBConnector *Connector
}

func (t *UserSearch) createSearchValue(ctx context.Context, searchValue models.MemberSearch) error {
	tx := t.DBConnector.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	err := t.DBConnector.DB.Create(&searchValue).WithContext(ctx).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, value := range searchValue.Values {
		value.SearchID = searchValue.ID
		err = t.DBConnector.DB.Create(&value).WithContext(ctx).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return nil
}

func (t *UserSearch) UpsertTags(ctx context.Context, userID uint, searchValues []models.MemberSearch) error {
	tx := t.DBConnector.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	err := t.deleteAllUserSearchValues(ctx, userID)
	if err != nil {
		tx.Rollback()
		return err
	}
	for _, searchValue := range searchValues {
		searchValue.UserID = userID
		err = t.createSearchValue(ctx, searchValue)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return nil
}

func (t *UserSearch) deleteAllUserSearchValues(ctx context.Context, userID uint) error {
	err := t.DBConnector.DB.
		Where("member_id = ?", userID).
		Delete(&models.MemberSearch{}).
		WithContext(ctx).
		Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return nil
}

func NewUserSearch(DBConnector *Connector) *UserSearch {
	return &UserSearch{DBConnector: DBConnector}
}
