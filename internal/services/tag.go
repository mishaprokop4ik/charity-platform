package service

import (
	"Kurajj/internal/models"
	"Kurajj/internal/repository"
	"context"
)

type Tagger interface {
	UpsertTags(ctx context.Context, eventID uint, eventType models.EventType, tags []models.Tag) error
	GetTagsByEvent(ctx context.Context, eventID uint, eventType models.EventType) ([]models.Tag, error)
}

type Tag struct {
	repo *repository.Repository
}

func (t *Tag) GetTagsByEvent(ctx context.Context, eventID uint, eventType models.EventType) ([]models.Tag, error) {
	return t.repo.Tag.GetTagsByEvent(ctx, eventID, eventType)
}

func (t *Tag) UpsertTags(ctx context.Context, eventID uint, eventType models.EventType, tags []models.Tag) error {
	return t.repo.Tag.UpsertTags(ctx, eventType, eventID, tags)
}

func NewTag(repo *repository.Repository) *Tag {
	return &Tag{
		repo: repo,
	}
}
