package service

import (
	"Kurajj/internal/models"
	"context"
)

func NewTransactionNotification(repo Repositorier) *TransactionNotification {
	return &TransactionNotification{repo: repo}
}

type TransactionNotification struct {
	repo Repositorier
}

func (t *TransactionNotification) Read(ctx context.Context, ids []uint) error {
	return t.repo.ReadNotifications(ctx, ids)
}

func (t *TransactionNotification) GetUserNotifications(ctx context.Context, userID uint) ([]models.TransactionNotification, error) {
	return t.repo.GetByMember(ctx, userID)
}
