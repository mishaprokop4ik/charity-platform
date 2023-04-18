package repository

import (
	"Kurajj/internal/models"
	zlog "Kurajj/pkg/logger"
	"context"
	"fmt"
	"github.com/google/uuid"
)

type HelpEvent struct {
	*Connector
	Filer
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

	if err := tx.Create(event.Tags).WithContext(ctx).Error; err != nil {
		tx.Rollback()
		return 0, err
	}

	return event.ID, tx.Commit().Error
}

func NewHelpEvent(config AWSConfig, DBConnector *Connector) *ProposalEvent {
	return &ProposalEvent{DBConnector: DBConnector, Filer: NewFile(config)}
}
