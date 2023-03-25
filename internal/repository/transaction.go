package repository

import (
	"Kurajj/internal/models"
	"context"
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

	rootEvent := models.ProposalEvent{}

	err = t.DBConnector.DB.Where("id = ?", transaction.EventID).First(&rootEvent).WithContext(ctx).Error
	if err != nil {
		return models.Transaction{}, err
	}

	responderInfo := models.User{}
	err = t.DBConnector.DB.Where("id = ?", rootEvent.AuthorID).First(&responderInfo).WithContext(ctx).Error
	if err != nil {
		return models.Transaction{}, err
	}

	transaction.Responder = responderInfo

	return transaction, nil
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
	if err := t.DBConnector.DB.
		Model(&models.Transaction{}).
		Select(lo.Keys(toUpdate)).
		Where("id = ?", id).
		Updates(toUpdate).
		WithContext(ctx).
		Error; err != nil {
		tx.Rollback()
		return err
	}

	eventType, ok := toUpdate["event_type"]
	if ok && eventType != models.ProposalEventType {
		return tx.Commit().Error
	}

	if status, ok := toUpdate["transaction_status"]; ok && !lo.Contains([]models.TransactionStatus{
		models.Completed,
		models.Interrupted,
		models.Canceled,
	}, status.(models.TransactionStatus)) {
		return tx.Commit().Error
	}

	eventID, ok := toUpdate["event_id"]
	if !ok {
		return tx.Commit().Error
	}

	err := t.DBConnector.DB.Model(&models.ProposalEvent{}).
		Where("event_id = ?", eventID).
		Update("remaining_helps", gorm.Expr("remaining_helps + ?", 1)).
		WithContext(ctx).
		Error
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (t *Transaction) CreateTransaction(ctx context.Context, transaction models.Transaction) (uint, error) {
	// TODO delete this comments after debug
	//var count int64
	//err := t.DBConnector.DB.Model(models.Transaction{}).
	//	Where("creator_id = ?", transaction.CreatorID).
	//	Where("event_id = ?", transaction.EventID).Count(&count).
	//	WithContext(ctx).
	//	Error
	//if err != nil {
	//	return 0, err
	//}
	//if count != 0 {
	//	return 0, fmt.Errorf("user cannot create more than one transaction per one event")
	//}
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
