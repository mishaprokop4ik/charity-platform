package service

import (
	"Kurajj/internal/models"
	"context"
	"database/sql"
	"time"
)

type Comment struct {
	repo Repositorier
}

func (c *Comment) WriteComment(ctx context.Context, comment models.Comment) (uint, error) {
	return c.repo.WriteComment(ctx, comment)
}

func (c *Comment) GetAllCommentsInEvent(ctx context.Context, eventID uint, eventType models.EventType) ([]models.Comment, error) {
	return c.repo.GetAllCommentsInEvent(ctx, eventID, eventType)
}

func (c *Comment) GetCommentByID(ctx context.Context, id uint) (models.Comment, error) {
	return c.repo.GetCommentByID(ctx, id)
}

func (c *Comment) UpdateComment(ctx context.Context, comment models.Comment) error {
	comment.IsUpdated = true
	comment.UpdatedAt = sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}

	return c.repo.UpdateComment(ctx, comment.ID, comment.GetValuesToUpdate())
}

func (c *Comment) DeleteComment(ctx context.Context, id uint) error {
	return c.repo.DeleteComment(ctx, id)
}

func NewComment(repo Repositorier) *Comment {
	return &Comment{repo: repo}
}
