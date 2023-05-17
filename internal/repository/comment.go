package repository

import (
	"Kurajj/internal/models"
	"context"
	"github.com/samber/lo"
)

type Comment struct {
	DBConnector *Connector
}

func (c *Comment) WriteComment(ctx context.Context, comment models.Comment) (uint, error) {
	err := c.DBConnector.
		DB.Create(&comment).
		WithContext(ctx).
		Error

	return comment.ID, err
}

func NewComment(DBConnector *Connector) *Comment {
	return &Comment{DBConnector: DBConnector}
}

func (c *Comment) GetAllCommentsInEvent(ctx context.Context, eventID uint, eventType models.EventType) ([]models.Comment, error) {
	comments := make([]models.Comment, 0)
	err := c.DBConnector.DB.
		Where("event_id = ?", eventID).
		Where("event_type = ?", eventType).
		Where("is_deleted = ?", false).
		Find(&comments).
		WithContext(ctx).
		Error
	return comments, err
}

func (c *Comment) GetCommentByID(ctx context.Context, id uint) (models.Comment, error) {
	comment := models.Comment{}
	err := c.DBConnector.DB.
		Where("id = ?", id).
		Where("is_deleted = ?", false).
		First(&comment).
		WithContext(ctx).
		Error

	return comment, err
}

func (c *Comment) UpdateComment(ctx context.Context, id uint, toUpdate map[string]any) error {
	return c.DBConnector.DB.
		Model(&models.Comment{}).
		Select(lo.Keys(toUpdate)).
		Where("id = ?", id).
		Updates(toUpdate).
		WithContext(ctx).
		Error
}

func (c *Comment) DeleteComment(ctx context.Context, id uint) error {
	tx := c.DBConnector.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	oldComment := models.Comment{}
	err := tx.Where("id = ?", id).First(&oldComment).WithContext(ctx).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	oldComment.IsDeleted = true

	err = tx.Save(&oldComment).WithContext(ctx).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
