package service

import (
	"Kurajj/internal/models"
	"Kurajj/internal/repository"
	"context"
	"fmt"
)

type Transactioner interface {
	UpdateTransaction(ctx context.Context, transaction models.Transaction) error
	GetCurrentEventTransactions(ctx context.Context,
		eventID uint,
		eventType models.EventType) ([]models.Transaction, error)
	UpdateAllNotFinishedTransactions(ctx context.Context, eventID uint, eventType models.EventType, newStatus models.TransactionStatus) error
	GetAllEventTransactions(ctx context.Context, eventID uint, eventType models.EventType) ([]models.Transaction, error)
	CreateTransaction(ctx context.Context, transaction models.Transaction) (uint, error)
	GetTransactionByID(ctx context.Context, id uint) (models.Transaction, error)
}

type Transaction struct {
	repo *repository.Repository
}

func (t *Transaction) GetTransactionByID(ctx context.Context, id uint) (models.Transaction, error) {
	return t.repo.Transaction.GetTransactionByID(ctx, id)
}

func (t *Transaction) CreateTransaction(ctx context.Context, transaction models.Transaction) (uint, error) {
	return t.repo.Transaction.CreateTransaction(ctx, transaction)
}

func (t *Transaction) UpdateTransaction(ctx context.Context, transaction models.Transaction) error {
	fmt.Println(transaction)
	if transaction.ID != 0 {
		return t.repo.Transaction.UpdateTransactionByID(ctx, transaction.ID, transaction.GetValuesToUpdate())
	}

	return t.repo.Transaction.UpdateTransactionByEvent(ctx, transaction.EventID,
		transaction.EventType,
		transaction.GetValuesToUpdate())
}

func (t *Transaction) GetCurrentEventTransactions(ctx context.Context,
	eventID uint,
	eventType models.EventType) ([]models.Transaction, error) {
	return t.repo.Transaction.GetCurrentEventTransactions(ctx, eventID, eventType)
}

func (t *Transaction) UpdateAllNotFinishedTransactions(ctx context.Context,
	eventID uint,
	eventType models.EventType,
	newStatus models.TransactionStatus) error {
	return t.repo.Transaction.UpdateAllNotFinishedTransactions(ctx, eventID, eventType, newStatus)
}

func (t *Transaction) GetAllEventTransactions(ctx context.Context,
	eventID uint,
	eventType models.EventType) ([]models.Transaction, error) {
	return t.repo.Transaction.GetAllEventTransactions(ctx, eventID, eventType)
}

func NewTransaction(repo *repository.Repository) *Transaction {
	return &Transaction{repo: repo}
}
