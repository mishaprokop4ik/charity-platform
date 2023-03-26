package repository

import (
	"Kurajj/internal/models"
	"context"
)

type TransactionNotification struct {
	DBConnector *Connector
}

const transactionLimit = 20

func (t *TransactionNotification) Create(ctx context.Context, notification models.TransactionNotification) (uint, error) {
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
		Where("member_id = ?", userID).
		Select(&notifications).Limit(transactionLimit).
		Order("is_read").
		Order("creation_date").
		WithContext(ctx).
		Error

	return notifications, err
}

func NewTransactionNotification(connector *Connector) *TransactionNotification {
	return &TransactionNotification{DBConnector: connector}
}

type Notifier interface {
	Create(ctx context.Context, notification models.TransactionNotification) (uint, error)
	Update(ctx context.Context, newNotification models.TransactionNotification) error
	GetByMember(ctx context.Context, userID uint) ([]models.TransactionNotification, error)
}
