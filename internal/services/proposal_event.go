package service

import (
	"Kurajj/internal/models"
	"Kurajj/internal/repository"
	"context"
	"fmt"
	"github.com/google/uuid"
	"io"
)

type ProposalEventer interface {
	ProposalEventCRUDer
	Response(ctx context.Context, proposalEventID, responderID uint, comment string) error
	Accept(ctx context.Context, transactionID uint) error
	UpdateStatus(ctx context.Context, status models.TransactionStatus, transactionID, userID uint, file io.Reader, fileType string) error
	GetUserProposalEvents(ctx context.Context, userID uint) ([]models.ProposalEvent, error)
	GetProposalEventBySearch(ctx context.Context, search models.ProposalEventSearchInternal) (models.ProposalEventPagination, error)
}

type ProposalEventCRUDer interface {
	CreateEvent(ctx context.Context, event models.ProposalEvent) (uint, error)
	GetEvent(ctx context.Context, id uint) (models.ProposalEvent, error)
	GetEvents(ctx context.Context) ([]models.ProposalEvent, error)
	UpdateEvent(ctx context.Context, event models.ProposalEvent) error
	DeleteEvent(ctx context.Context, id uint) error
}

func NewProposalEvent(repo *repository.Repository) *ProposalEvent {
	return &ProposalEvent{
		repo: repo, Transaction: *NewTransaction(repo)}
}

type ProposalEvent struct {
	Transaction
	repo *repository.Repository
}

func (p *ProposalEvent) GetProposalEventBySearch(ctx context.Context, search models.ProposalEventSearchInternal) (models.ProposalEventPagination, error) {
	return p.repo.ProposalEvent.GetEventsWithSearchAndSort(ctx, search)
}

func (p *ProposalEvent) UpdateStatus(ctx context.Context, status models.TransactionStatus, transactionID, userID uint, file io.Reader, fileType string) error {
	transaction, err := p.GetTransactionByID(ctx, transactionID)
	if err != nil {
		return err
	}

	if transaction.CreatorID == userID {
		transaction.ResponderStatus = status
		transaction.ReceiverStatus = status
	}

	if status == models.Completed {
		fileUniqueID, err := uuid.NewUUID()
		if err != nil {
			return err
		}
		fileName := fmt.Sprintf("%s.%s", fileUniqueID.String(), fileType)
		filePath, err := p.repo.File.Upload(ctx, fileName, file)
		if err != nil {
			return err
		}
		transaction.ReportURL = filePath
	}

	return p.UpdateTransaction(ctx, transaction)
}

func (p *ProposalEvent) Response(ctx context.Context, proposalEventID, responderID uint, comment string) error {
	//transaction, err := p.repo.ProposalEvent.GetEvent(ctx, proposalEventID)
	//if err != nil {
	//	return err
	//}
	//if transaction.AuthorID == responderID {
	//	return fmt.Errorf("event creator cannot response his/her own events")
	//}
	_, err := p.CreateTransaction(ctx, models.Transaction{
		CreatorID:       responderID,
		EventID:         proposalEventID,
		Comment:         comment,
		EventType:       models.ProposalEventType,
		ReceiverStatus:  models.Waiting,
		ResponderStatus: models.NotStarted,
	})

	return err
}

func (p *ProposalEvent) Accept(ctx context.Context, transactionID uint) error {
	return p.UpdateTransaction(ctx, models.Transaction{
		ID:             transactionID,
		ReceiverStatus: models.InProcess,
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
	return p.repo.ProposalEvent.UpdateEvent(ctx, event)
}

func (p *ProposalEvent) DeleteEvent(ctx context.Context, id uint) error {
	return p.repo.ProposalEvent.DeleteEvent(ctx, id)
}
