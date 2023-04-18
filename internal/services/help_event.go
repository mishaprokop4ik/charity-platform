package service

import (
	"Kurajj/internal/models"
	"Kurajj/internal/repository"
	"context"
)

type HelpEvent struct {
	*Transaction
	repo *repository.Repository
}

func NewHelpEvent(r *repository.Repository) *HelpEvent {
	return &HelpEvent{repo: r, Transaction: NewTransaction(r)}
}

func (h *HelpEvent) CreateHelpEvent(ctx context.Context, event *models.HelpEvent) (uint, error) {
	return h.repo.CreateEvent(ctx, event)
}
