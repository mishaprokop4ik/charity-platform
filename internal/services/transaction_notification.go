package service

import (
	"Kurajj/internal/models"
	"Kurajj/internal/repository"
	"context"
)

type TransactionNotifier interface {
	Read(ctx context.Context, id []uint) error
	GetUserNotifications(ctx context.Context, userID uint) ([]models.TransactionNotification, error)
}

type TransactionNotification struct {
	repo *repository.Repository
}

func (t *TransactionNotification) Read(ctx context.Context, ids []uint) error {
	//TODO implement me
	panic("implement me")
}

func (t *TransactionNotification) GetUserNotifications(ctx context.Context, userID uint) ([]models.TransactionNotification, error) {
	//TODO implement me
	panic("implement me")
}
