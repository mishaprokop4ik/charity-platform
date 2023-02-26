package service

import (
	"Kurajj/internal/models"
	"Kurajj/internal/repository"
	"context"
	"database/sql"
	"time"
)

type Comment struct {
	repo *repository.Repository
}

func (c *Comment) WriteComment(ctx context.Context, comment models.Comment) (uint, error) {
	return c.repo.Comment.WriteComment(ctx, comment)
}

func (c *Comment) GetAllCommentsInEvent(ctx context.Context, eventID uint, eventType models.EventType) ([]models.Comment, error) {
	return c.repo.Comment.GetAllCommentsInEvent(ctx, eventID, eventType)
}

func (c *Comment) GetCommentByID(ctx context.Context, id uint) (models.Comment, error) {
	return c.repo.Comment.GetCommentByID(ctx, id)
}

func (c *Comment) UpdateComment(ctx context.Context, comment models.Comment) error {
	comment.IsUpdated = true
	comment.UpdatedAt = sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}

	return c.repo.Comment.UpdateComment(ctx, comment.ID, comment.GetValuesToUpdate())
}

func (c *Comment) DeleteComment(ctx context.Context, id uint) error {
	return c.repo.Comment.DeleteComment(ctx, id)
}

func NewComment(repo *repository.Repository) *Comment {
	return &Comment{repo: repo}
}

type Commenter interface {
	GetAllCommentsInEvent(ctx context.Context, eventID uint, eventType models.EventType) ([]models.Comment, error)
	GetCommentByID(ctx context.Context, id uint) (models.Comment, error)
	UpdateComment(ctx context.Context, comment models.Comment) error
	DeleteComment(ctx context.Context, id uint) error
	WriteComment(ctx context.Context, comment models.Comment) (uint, error)
}
