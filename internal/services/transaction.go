package service

import (
	"Kurajj/internal/models"
	"context"
)

type Transaction struct {
	repo Repositorier
}

func (t *Transaction) GetTransactionByID(ctx context.Context, id uint) (models.Transaction, error) {
	return t.repo.GetTransactionByID(ctx, id)
}

func (t *Transaction) CreateTransaction(ctx context.Context, transaction models.Transaction) (uint, error) {
	return t.repo.CreateTransaction(ctx, transaction)
}

func (t *Transaction) UpdateTransaction(ctx context.Context, transaction models.Transaction) error {
	if transaction.ID != 0 {
		return t.repo.UpdateTransactionByID(ctx, transaction.ID, transaction.GetValuesToUpdate())
	}

	return t.repo.UpdateTransactionByEvent(ctx, transaction.EventID,
		transaction.EventType,
		transaction.GetValuesToUpdate())
}

func (t *Transaction) GetCurrentEventTransactions(ctx context.Context,
	eventID uint,
	eventType models.EventType) ([]models.Transaction, error) {
	return t.repo.GetCurrentEventTransactions(ctx, eventID, eventType)
}

func (t *Transaction) UpdateAllNotFinishedTransactions(ctx context.Context,
	eventID uint,
	eventType models.EventType,
	newStatus models.TransactionStatus) error {
	return t.repo.UpdateAllNotFinishedTransactions(ctx, eventID, eventType, newStatus)
}

func (t *Transaction) GetAllEventTransactions(ctx context.Context,
	eventID uint,
	eventType models.EventType) ([]models.Transaction, error) {
	return t.repo.GetAllEventTransactions(ctx, eventID, eventType)
}

func NewTransaction(repo Repositorier) *Transaction {
	return &Transaction{repo: repo}
}
