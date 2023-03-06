package service

import (
	"Kurajj/internal/models"
	"context"
)

// Searcher is an interface for search in all events
type Searcher interface {
	SearchByName(ctx context.Context, name string) (models.Eventer, error)
	SearchByTags(ctx context.Context, tags []models.Tag) (models.Eventer, error)
}
