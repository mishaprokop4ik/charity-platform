package service

import (
	"Kurajj/internal/models"
	"Kurajj/internal/repository"
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"io"
	"time"
)

type ProposalEventer interface {
	ProposalEventCRUDer
	Response(ctx context.Context, proposalEventID, responderID uint, comment string) error
	Accept(ctx context.Context, request models.AcceptRequest) error
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
		repo: repo, Transaction: NewTransaction(repo)}
}

type ProposalEvent struct {
	*Transaction
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

	if transaction.TransactionStatus == models.Completed || transaction.TransactionStatus == models.Aborted {
		return fmt.Errorf("transaction cannot be changed when it it in %s state", transaction.TransactionStatus)
	}

	transaction.TransactionStatus = status
	transaction.ResponderStatus = status

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

	if status == models.Completed || status == models.Canceled || status == models.Interrupted {
		transaction.CompetitionDate = sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		}
	}

	err = p.UpdateTransaction(ctx, transaction)
	if err != nil {
		return err
	}

	err = p.createNotification(ctx, models.TransactionNotification{
		EventType:     models.ProposalEventType,
		EventID:       transaction.EventID,
		Action:        models.Updated,
		TransactionID: transactionID,
		NewStatus:     status,
		IsRead:        false,
		CreationTime:  time.Now(),
		MemberID:      transaction.CreatorID,
	})

	if err != nil {
		return err
	}

	return nil
}

func (p *ProposalEvent) Response(ctx context.Context, proposalEventID, responderID uint, comment string) error {
	proposalEvent, err := p.repo.ProposalEvent.GetEvent(ctx, proposalEventID)
	if err != nil {
		return err
	}
	if proposalEvent.AuthorID == responderID {
		return fmt.Errorf("event creator cannot response his/her own events")
	}
	//TODO remove after debug
	//for _, transaction := range proposalEvent.Transactions {
	//	if transaction.CreatorID == responderID && lo.Contains([]models.TransactionStatus{
	//		models.Accepted,
	//		models.InProcess,
	//		models.Waiting,
	//	}, transaction.TransactionStatus) {
	//		return fmt.Errorf("user already has a transaction in this event")
	//	}
	//}
	id, err := p.CreateTransaction(ctx, models.Transaction{
		CreatorID:         responderID,
		EventID:           proposalEventID,
		Comment:           comment,
		EventType:         models.ProposalEventType,
		CreationDate:      time.Now(),
		TransactionStatus: models.Waiting,
		ResponderStatus:   models.NotStarted,
	})
	if err != nil {
		return err
	}

	err = p.createNotification(ctx, models.TransactionNotification{
		EventType:     models.ProposalEventType,
		EventID:       proposalEventID,
		Action:        models.Created,
		TransactionID: id,
		IsRead:        false,
		CreationTime:  time.Now(),
		MemberID:      proposalEvent.AuthorID,
	})

	return err
}

func (p *ProposalEvent) Accept(ctx context.Context, request models.AcceptRequest) error {
	status := models.Canceled
	if request.Accept {
		status = models.Accepted
	}
	err := p.UpdateTransaction(ctx, models.Transaction{
		ID:                request.TransactionID,
		TransactionStatus: status,
	})
	if err != nil {
		return err
	}

	transaction, err := p.GetTransactionByID(ctx, request.TransactionID)
	if err != nil {
		return err
	}

	err = p.createNotification(ctx, models.TransactionNotification{
		EventType:     models.ProposalEventType,
		EventID:       transaction.EventID,
		Action:        models.Updated,
		TransactionID: request.TransactionID,
		NewStatus:     status,
		IsRead:        false,
		CreationTime:  time.Now(),
		MemberID:      request.MemberID,
	})

	return err
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

func (p *ProposalEvent) createNotification(ctx context.Context, notification models.TransactionNotification) error {
	_, err := p.repo.TransactionNotification.Create(ctx, notification)
	return err
}
