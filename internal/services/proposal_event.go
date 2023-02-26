package service

import (
	"Kurajj/internal/models"
	"Kurajj/internal/repository"
	"context"
)

type ProposalEventer interface {
	ProposalEventCRUDer
	Response(ctx context.Context, proposalEventID, responderID uint) error
	Accept(ctx context.Context, transactionID uint) error
	UpdateStatus(ctx context.Context, status models.Status, id uint) error
	GetUserProposalEvents(ctx context.Context, userID uint) ([]models.ProposalEvent, error)
}

type ProposalEventCRUDer interface {
	CreateEvent(ctx context.Context, event models.ProposalEvent) (uint, error)
	GetEvent(ctx context.Context, id uint) (models.ProposalEvent, error)
	GetEvents(ctx context.Context) ([]models.ProposalEvent, error)
	UpdateEvent(ctx context.Context, event models.ProposalEvent) error
	DeleteEvent(ctx context.Context, id uint) error
}

func NewProposalEvent(repo *repository.Repository) *ProposalEvent {
	return &ProposalEvent{repo: repo}
}

type ProposalEvent struct {
	Transaction
	repo *repository.Repository
}

func (p *ProposalEvent) Response(ctx context.Context, proposalEventID, responderID uint) error {
	_, err := p.CreateTransaction(ctx, models.Transaction{
		CreatorID: responderID,
		EventID:   proposalEventID,
		EventType: models.ProposalEventType,
		Status:    models.Waiting,
	})

	return err
}

func (p *ProposalEvent) Accept(ctx context.Context, transactionID uint) error {
	return p.UpdateTransaction(ctx, models.Transaction{
		ID:     transactionID,
		Status: models.InProcess,
	})
}

func (p *ProposalEvent) GetUserProposalEvents(ctx context.Context, userID uint) ([]models.ProposalEvent, error) {
	return p.repo.ProposalEvent.GetUserProposalEvents(ctx, userID)
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
	//TODO add transaction login
	err := p.UpdateAllNotFinishedTransactions(ctx, id, models.ProposalEventType, models.Canceled)
	if err != nil {
		return err
	}
	return p.repo.ProposalEvent.DeleteEvent(ctx, id)
}
