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

func NewTransactionNotification(repo *repository.Repository) *TransactionNotification {
	return &TransactionNotification{repo: repo}
}

func (t *TransactionNotification) Read(ctx context.Context, ids []uint) error {
	return t.repo.TransactionNotification.ReadNotifications(ctx, ids)
}

func (t *TransactionNotification) GetUserNotifications(ctx context.Context, userID uint) ([]models.TransactionNotification, error) {
	return t.repo.TransactionNotification.GetByMember(ctx, userID)
}
