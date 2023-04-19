package repository

import (
	"Kurajj/internal/models"
	zlog "Kurajj/pkg/logger"
	"context"
	"fmt"
	"github.com/google/uuid"
)

func NewHelpEvent(config AWSConfig, DBConnector *Connector) *HelpEvent {
	return &HelpEvent{Connector: DBConnector, Filer: NewFile(config)}
}

type HelpEvent struct {
	*Connector
	Filer
}

func (h *HelpEvent) UpdateNeeds(ctx context.Context, needs ...models.Need) error {
	tx := h.DB.Begin()
	for _, need := range needs {
		err := tx.Model(&need).Updates(need).WithContext(ctx).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

func (h *HelpEvent) GetHelpEventByTransactionID(ctx context.Context, transactionID models.ID) (models.HelpEvent, error) {
	transaction := models.Transaction{}
	err := h.DB.First(&transaction, "id = ?", transactionID).WithContext(ctx).Error
	fmt.Printf("%+v", transaction)
	if err != nil {
		return models.HelpEvent{}, err
	}

	event := models.HelpEvent{}
	err = h.DB.First(&event, "id = ?", transaction.EventID).WithContext(ctx).Error
	fmt.Printf("%+v", event)
	return event, err
}

func (h *HelpEvent) CreateNeed(ctx context.Context, need models.Need) (uint, error) {
	err := h.DB.Create(&need).WithContext(ctx).Error
	return need.ID, err
}

func (h *HelpEvent) GetHelpEventNeeds(ctx context.Context, eventID models.ID) ([]models.Need, error) {
	needs := make([]models.Need, 0)
	err := h.DB.Where("help_event_id = ?", eventID).Where("transaction_id IS NULL").Find(&needs).WithContext(ctx).Error
	return needs, err
}
func (h *HelpEvent) GetEventByID(ctx context.Context, id models.ID) (models.HelpEvent, error) {
	event := models.HelpEvent{
		TransactionNeeds: map[models.ID][]models.Need{},
	}
	err := h.DB.First(&event, id).WithContext(ctx).Error
	if err != nil {
		return models.HelpEvent{}, err
	}
	eventNeeds := make([]models.Need, 0)
	err = h.DB.Where("help_event_id = ?", id).Where("transaction_id IS NULL").Find(&eventNeeds).WithContext(ctx).Error
	if err != nil {
		return models.HelpEvent{}, err
	}
	event.Needs = eventNeeds
	transactions := make([]models.Transaction, 0)
	err = h.DB.Where("event_type = ?", models.HelpEventType).Where("event_id = ?", id).Find(&transactions).WithContext(ctx).Error
	if err != nil {
		return models.HelpEvent{}, err
	}
	event.Transactions = transactions
	tags := make([]models.Tag, 0)
	err = h.DB.Where("event_type = ?", models.HelpEventType).Where("event_id = ?", id).Find(&tags).WithContext(ctx).Error
	if err != nil {
		return models.HelpEvent{}, err
	}
	event.Tags = tags
	comments := make([]models.Comment, 0)
	err = h.DB.Where("event_type = ?", models.HelpEventType).Where("event_id = ?", id).Find(&comments).WithContext(ctx).Error
	event.Comments = comments
	if err != nil {
		return models.HelpEvent{}, err
	}
	user := models.User{}
	err = h.DB.First(&user, "id = ?", event.CreatedBy).WithContext(ctx).Error
	event.User = user
	if err != nil {
		return models.HelpEvent{}, err
	}
	for _, t := range transactions {
		transactionNeeds := make([]models.Need, 0)
		err = h.DB.Where("transaction_id = ?", t.ID).Find(&transactionNeeds).WithContext(ctx).Error
		if err != nil {
			return models.HelpEvent{}, err
		}
		event.TransactionNeeds[models.ID(t.ID)] = transactionNeeds
	}
	fmt.Printf("%+v", event)
	return event, err
}

func (h *HelpEvent) CreateEvent(ctx context.Context, event *models.HelpEvent) (uint, error) {
	tx := h.DB.Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}

	if event.File != nil {
		fileName, err := uuid.NewUUID()
		if err != nil {
			tx.Commit()
			return 0, err
		}
		filePath, err := h.Filer.Upload(ctx, fmt.Sprintf("%s.%s", fileName.String(), event.FileType), event.File)
		if err != nil {
			zlog.Log.Error(err, "could not upload file")
			return 0, err
		}
		event.ImagePath = filePath
	}

	if err := tx.Create(event).WithContext(ctx).Error; err != nil {
		tx.Rollback()
		return 0, err
	}
	if len(event.Tags) != 0 {
		if err := tx.Create(event.Tags).WithContext(ctx).Error; err != nil {
			tx.Rollback()
			return 0, err
		}
	}

	return event.ID, tx.Commit().Error
}
