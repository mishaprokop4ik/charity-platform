package service

import (
	"Kurajj/internal/models"
	"Kurajj/internal/repository"
	"context"
)

type Comment struct {
	repo *repository.Repository
}

type Commenter interface {
	GetAllCommentsInEvent(ctx context.Context, eventID uint, eventType models.EventType) ([]models.Comment, error)
	GetCommentByID(ctx context.Context, id uint) (models.Comment, error)
	UpdateComment(ctx context.Context, comment models.Comment) error
	DeleteComment(ctx context.Context, id uint) error
}
