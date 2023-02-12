package service

import (
	"Kurajj/internal/models"
	"Kurajj/internal/repository"
	"context"
)

type ProposalEventer interface {
	ProposalEventCRUDer
	GetUserProposalEvents(ctx context.Context, userID uint) ([]models.ProposalEvent, error)
}

type ProposalEventCRUDer interface {
	CreateEvent(ctx context.Context, event models.ProposalEvent) (uint, error)
	GetEvent(ctx context.Context, id uint) (models.ProposalEvent, error)
	GetEvents(ctx context.Context) ([]models.ProposalEvent, error)
	UpdateEvent(ctx context.Context, event models.ProposalEvent) error
	DeleteEvent(ctx context.Context, id uint) error
}

type ProposalEvent struct {
	repo *repository.Repository
}

func (p *ProposalEvent) GetUserProposalEvents(ctx context.Context, userID uint) ([]models.ProposalEvent, error) {
	return p.repo.ProposalEvent.GetUserProposalEvents(ctx, userID)
}

func NewProposalEvent(repo *repository.Repository) *ProposalEvent {
	return &ProposalEvent{repo: repo}
}

func (p *ProposalEvent) CreateEvent(ctx context.Context, event models.ProposalEvent) (uint, error) {
	return p.repo.ProposalEvent.CreateEvent(ctx, event)
}

func (p *ProposalEvent) GetEvent(ctx context.Context, id uint) (models.ProposalEvent, error) {
	return p.repo.ProposalEvent.GetEvent(ctx, id)
}

func (p *ProposalEvent) GetEvents(ctx context.Context) ([]models.ProposalEvent, error) {
	return p.repo.ProposalEvent.GetEvents(ctx)
}

func (p *ProposalEvent) UpdateEvent(ctx context.Context, event models.ProposalEvent) error {

	return p.repo.ProposalEvent.UpdateEvent(ctx, event.ID, event.GetValuesToUpdate())
}

func (p *ProposalEvent) DeleteEvent(ctx context.Context, id uint) error {
	return p.repo.ProposalEvent.DeleteEvent(ctx, id)
}
