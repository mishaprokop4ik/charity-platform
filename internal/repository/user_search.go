package repository

import (
	"Kurajj/internal/models"
	"context"
	"errors"
	"gorm.io/gorm"
)

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
	err := tx.Create(&searchValue).WithContext(ctx).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, value := range searchValue.Values {
		value.SearchID = searchValue.ID
		err = tx.Create(&value).WithContext(ctx).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

func (t *UserSearch) UpsertUserTags(ctx context.Context, userID uint, searchValues []models.MemberSearch) error {
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

	return tx.Commit().Error
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
