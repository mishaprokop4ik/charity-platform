package repository

import (
	"Kurajj/internal/models"
	"context"
	"github.com/samber/lo"
)

type Transactioner interface {
	UpdateTransactionByEvent(ctx context.Context, eventID uint, eventType models.EventType, toUpdate map[string]any) error
	UpdateTransactionByID(ctx context.Context, id uint, toUpdate map[string]any) error
	GetCurrentEventTransactions(ctx context.Context,
		eventID uint,
		eventType models.EventType) ([]models.Transaction, error)
	UpdateAllNotFinishedTransactions(ctx context.Context, eventID uint, eventType models.EventType, newStatus models.TransactionStatus) error
	GetAllEventTransactions(ctx context.Context, eventID uint, eventType models.EventType) ([]models.Transaction, error)
	CreateTransaction(ctx context.Context, transaction models.Transaction) (uint, error)
	GetTransactionByID(ctx context.Context, id uint) (models.Transaction, error)
}

type Transaction struct {
	DBConnector *Connector
}

func (t *Transaction) GetTransactionByID(ctx context.Context, id uint) (models.Transaction, error) {
	transaction := models.Transaction{}
	err := t.DBConnector.DB.Where("id = ?", id).First(&transaction).WithContext(ctx).Error
	return transaction, err
}

func (t *Transaction) UpdateTransactionByEvent(ctx context.Context, eventID uint, eventType models.EventType, toUpdate map[string]any) error {
	return t.DBConnector.DB.
		Select(lo.Keys(toUpdate)).
		Model(&models.Transaction{}).
		Where("event_id = ?", eventID).
		Where("event_type = ?", eventType).
		Updates(toUpdate).
		WithContext(ctx).
		Error
}

func (t *Transaction) UpdateTransactionByID(ctx context.Context, id uint, toUpdate map[string]any) error {
	return t.DBConnector.DB.
		Model(&models.Transaction{}).
		Select(lo.Keys(toUpdate)).
		Where("id = ?", id).
		Updates(toUpdate).
		WithContext(ctx).
		Error
}

func (t *Transaction) CreateTransaction(ctx context.Context, transaction models.Transaction) (uint, error) {
	err := t.DBConnector.DB.Create(&transaction).WithContext(ctx).Error
	return transaction.ID, err
}

func (t *Transaction) GetCurrentEventTransactions(ctx context.Context, eventID uint, eventType models.EventType) ([]models.Transaction, error) {
	transactions := make([]models.Transaction, 0)
	err := t.DBConnector.DB.
		Where("event_id = ?", eventID).
		Where("event_type = ?", eventType).
		Not("status IN (?)", []models.TransactionStatus{models.Completed, models.Interrupted, models.Canceled}).
		Find(&transactions).WithContext(ctx).
		Error

	return transactions, err
}

func (t *Transaction) UpdateAllNotFinishedTransactions(ctx context.Context, eventID uint, eventType models.EventType, newStatus models.TransactionStatus) error {
	return t.DBConnector.DB.
		Model(&models.Transaction{}).
		Where("event_id = ?", eventID).
		Where("event_type = ?", eventType).
		Not("status IN (?)", []models.TransactionStatus{models.Completed, models.Interrupted, models.Canceled}).
		Update("status", newStatus).
		WithContext(ctx).
		Error
}

func (t *Transaction) GetAllEventTransactions(ctx context.Context,
	eventID uint,
	eventType models.EventType) ([]models.Transaction, error) {
	transactions := []models.Transaction{}
	err := t.DBConnector.DB.
		Find(&transactions).
		Where("event_id = ?", eventID).
		Where("event_type = ?", eventType).
		WithContext(ctx).
		Error

	return transactions, err
}

func NewTransaction(DBConnector *Connector) *Transaction {
	return &Transaction{DBConnector: DBConnector}
}
