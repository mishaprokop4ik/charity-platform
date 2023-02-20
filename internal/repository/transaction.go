package repository

import (
	"Kurajj/internal/models"
	"context"
	"github.com/samber/lo"
)

type Transactioner interface {
	UpdateTransaction(ctx context.Context, eventID uint, eventType models.EventType, toUpdate map[string]any) error
	GetCurrentEventTransactions(ctx context.Context, eventID uint, eventType models.EventType) ([]models.Transaction, error)
	UpdateAllNotFinishedTransactions(ctx context.Context, eventID uint, eventType models.EventType, newStatus models.Status) error
	GetAllEventTransactions(ctx context.Context, eventID uint, eventType models.EventType) ([]models.Transaction, error)
}

type Transaction struct {
	DBConnector *Connector
}

func (t *Transaction) UpdateTransaction(ctx context.Context, eventID uint, eventType models.EventType, toUpdate map[string]any) error {
	return t.DBConnector.DB.
		Select(lo.Keys(toUpdate)).
		Where("event_id = ?", eventID).
		Where("event_type = ?", eventType).
		Updates(toUpdate).
		WithContext(ctx).
		Error
}

func (t *Transaction) GetCurrentEventTransactions(ctx context.Context, eventID uint, eventType models.EventType) ([]models.Transaction, error) {
	transactions := make([]models.Transaction, 0)
	err := t.DBConnector.DB.
		Where("event_id = ?", eventID).
		Where("event_type = ?", eventType).
		Not("status IN ?", models.Completed, models.Interrupted, models.Canceled).
		Find(&transactions).WithContext(ctx).
		Error

	return transactions, err
}

func (t *Transaction) UpdateAllNotFinishedTransactions(ctx context.Context, eventID uint, eventType models.EventType, newStatus models.Status) error {
	return t.DBConnector.DB.
		Where("event_id = ?", eventID).
		Where("event_type = ?", eventType).
		Update("status = ?", newStatus).
		WithContext(ctx).
		Error
}

func (t *Transaction) GetAllEventTransactions(ctx context.Context, eventID uint, eventType models.EventType) ([]models.Transaction, error) {
	transactions := make([]models.Transaction, 0)
	err := t.DBConnector.DB.
		Where("event_id = ?", eventID).
		Where("event_type = ?", eventType).
		Find(&transactions).WithContext(ctx).
		Error

	return transactions, err
}

func NewTransaction(DBConnector *Connector) *Transaction {
	return &Transaction{DBConnector: DBConnector}
}
