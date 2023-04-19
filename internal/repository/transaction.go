package repository

import (
	"Kurajj/internal/models"
	zlog "Kurajj/pkg/logger"
	"context"
	"fmt"
	"github.com/samber/lo"
	"gorm.io/gorm"
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
	if err != nil {
		return models.Transaction{}, err
	}

	return t.updateTransactionUsers(ctx, transaction)
}

func (t *Transaction) updateTransactionUsers(ctx context.Context, transaction models.Transaction) (models.Transaction, error) {
	creatorInfo := models.User{}
	err := t.DBConnector.DB.Where("id = ?", transaction.CreatorID).First(&creatorInfo).WithContext(ctx).Error
	if err != nil {
		return models.Transaction{}, err
	}

	transaction.Creator = creatorInfo
	var authorID uint
	switch transaction.EventType {
	case models.ProposalEventType:
		authorID, err = t.getProposalEventCreator(ctx, transaction.EventID)
	case models.HelpEventType:
		authorID, err = t.getHelpEventCreator(ctx, transaction.EventID)
	default:
		return models.Transaction{}, fmt.Errorf("unexpected event type %s", transaction.EventType)
	}

	responderInfo := models.User{}
	err = t.DBConnector.DB.Where("id = ?", authorID).First(&responderInfo).WithContext(ctx).Error
	if err != nil {
		return models.Transaction{}, err
	}

	transaction.Responder = responderInfo

	return transaction, nil
}

func (t *Transaction) getProposalEventCreator(ctx context.Context, eventID uint) (uint, error) {
	rootEvent := models.ProposalEvent{}

	err := t.DBConnector.DB.Where("id = ?", eventID).First(&rootEvent).WithContext(ctx).Error
	if err != nil {
		return 0, err
	}

	return rootEvent.AuthorID, nil
}

func (t *Transaction) getHelpEventCreator(ctx context.Context, eventID uint) (uint, error) {
	rootEvent := models.HelpEvent{}

	err := t.DBConnector.DB.Where("id = ?", eventID).First(&rootEvent).WithContext(ctx).Error
	if err != nil {
		return 0, err
	}

	return rootEvent.CreatedBy, nil
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
	tx := t.DBConnector.DB.Begin()
	transaction := models.Transaction{}
	err := tx.Where("id = ?", id).First(&transaction).WithContext(ctx).Error
	if err != nil {
		zlog.Log.Error(err, "could not get transaction when updating remaining helps")
		return tx.Commit().Error
	}
	if err := tx.
		Model(&models.Transaction{}).
		Select(lo.Keys(toUpdate)).
		Where("id = ?", id).
		Updates(toUpdate).
		WithContext(ctx).
		Error; err != nil {
		tx.Rollback()
		return err
	}

	status := toUpdate["transaction_status"]

	if lo.Contains([]models.TransactionStatus{
		models.Completed,
		models.Interrupted,
		models.Canceled,
	}, status.(models.TransactionStatus)) && status != transaction.TransactionStatus {
		switch transaction.EventType {
		case models.ProposalEventType:
			err = tx.Model(&models.ProposalEvent{}).
				Where("id = ?", transaction.EventID).
				Update("remaining_helps", gorm.Expr("remaining_helps + ?", 1)).
				WithContext(ctx).
				Error
			if err != nil {
				tx.Rollback()
				return err
			}
		}
	} else if status == models.Accepted && transaction.TransactionStatus != status {
		switch transaction.EventType {
		case models.ProposalEventType:
			err = tx.Model(&models.ProposalEvent{}).
				Where("id = ?", transaction.EventID).
				Update("remaining_helps", gorm.Expr("remaining_helps - ?", 1)).
				WithContext(ctx).
				Error
			if err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return tx.Commit().Error
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
	if err != nil {
		return []models.Transaction{}, err
	}
	for i, transaction := range transactions {
		newTransaction, err := t.updateTransactionUsers(ctx, transaction)
		if err != nil {
			return nil, err
		}
		transactions[i] = newTransaction
	}

	return transactions, nil
}

func NewTransaction(DBConnector *Connector) *Transaction {
	return &Transaction{DBConnector: DBConnector}
}
