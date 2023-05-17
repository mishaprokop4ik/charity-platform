package service

import (
	"Kurajj/internal/models"
	"context"
)

func NewTag(repo Repositorier) *Tag {
	return &Tag{
		repo: repo,
	}
}

type Tag struct {
	repo Repositorier
}

func (t *Tag) GetTagsByEvent(ctx context.Context, eventID uint, eventType models.EventType) ([]models.Tag, error) {
	return t.repo.GetTagsByEvent(ctx, eventID, eventType)
}

func (t *Tag) UpsertTags(ctx context.Context, eventID uint, eventType models.EventType, tags []models.Tag) error {
	return t.repo.UpsertTags(ctx, eventType, eventID, tags)
}
