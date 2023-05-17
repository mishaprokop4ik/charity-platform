package repository

import (
	"Kurajj/internal/models"
	zlog "Kurajj/pkg/logger"
	"context"
	"errors"
	"gorm.io/gorm"
)

type TransactionNotification struct {
	DBConnector *Connector
}

func (t *TransactionNotification) ReadNotifications(ctx context.Context, ids []uint) error {
	return t.DBConnector.DB.Model(&models.TransactionNotification{}).Where("id IN (?)", ids).UpdateColumn("is_read", true).WithContext(ctx).Error
}

func (t *TransactionNotification) GetByID(ctx context.Context, id uint) (models.TransactionNotification, error) {
	notification := models.TransactionNotification{}
	err := t.DBConnector.DB.Where("id = ?", id).First(&notification).WithContext(ctx).Error
	return notification, err
}

const transactionLimit = 20

func (t *TransactionNotification) CreateNotification(ctx context.Context, notification models.TransactionNotification) (uint, error) {
	err := t.DBConnector.DB.Create(&notification).WithContext(ctx).Error
	return notification.ID, err
}

func (t *TransactionNotification) Update(ctx context.Context, newNotification models.TransactionNotification) error {
	err := t.DBConnector.DB.Save(&newNotification).WithContext(ctx).Error
	return err
}

func (t *TransactionNotification) GetByMember(ctx context.Context, userID uint) ([]models.TransactionNotification, error) {
	notifications := []models.TransactionNotification{}
	err := t.DBConnector.DB.
		Order("is_read ASC, creation_time DESC").
		Limit(transactionLimit).
		Where("member_id = ?", userID).
		Find(&notifications).
		WithContext(ctx).
		Error

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		zlog.Log.Error(err, "could not find any of TransactionNotification")
		return nil, err
	}

	for i, notification := range notifications {
		transaction := models.Transaction{}
		err = t.DBConnector.DB.Where("id = ?", notification.TransactionID).First(&transaction).WithContext(ctx).Error
		if err != nil {
			zlog.Log.Error(err, "could not find any of Transaction")
			return nil, err
		}
		switch notification.EventType {
		case models.ProposalEventType:
			event := models.ProposalEvent{}
			err = t.DBConnector.DB.Where("id = ?", transaction.EventID).First(&event).WithContext(ctx).Error
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, err
			}

			notifications[i].EventTitle = event.Title
		case models.HelpEventType:
			event := models.HelpEvent{}
			err = t.DBConnector.DB.Where("id = ?", transaction.EventID).First(&event).WithContext(ctx).Error
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, err
			}

			notifications[i].EventTitle = event.Title
		}
	}

	return notifications, err
}

func NewTransactionNotification(connector *Connector) *TransactionNotification {
	return &TransactionNotification{DBConnector: connector}
}
