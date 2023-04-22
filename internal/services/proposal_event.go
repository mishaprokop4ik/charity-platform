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
	GetStatistics(ctx context.Context, fromStart int, creatorID uint) (models.ProposalEventStatistics, error)
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

func (p *ProposalEvent) GetStatistics(ctx context.Context, fromStart int, creatorID uint) (models.ProposalEventStatistics, error) {
	currentTransactions, err := p.getCurrentMonthTransactions(ctx, fromStart, creatorID)
	if err != nil {
		return models.ProposalEventStatistics{}, err
	}

	previousTransactions, err := p.getPreviousMonthTransactions(ctx, fromStart, creatorID)
	if err != nil {
		return models.ProposalEventStatistics{}, err
	}

	statistics := p.generateStatistics(currentTransactions, previousTransactions)
	return statistics, nil
}

func (p *ProposalEvent) getCurrentMonthTransactions(ctx context.Context, fromStart int, creatorID uint) ([]models.Transaction, error) {
	currentMonthTo := time.Now()
	currentMonthFrom := currentMonthTo.AddDate(0, 0, int(-fromStart))

	currentTransactions, err := p.repo.ProposalEvent.GetStatistics(ctx, creatorID, currentMonthFrom, currentMonthTo)
	if err != nil {
		return nil, err
	}

	return currentTransactions, nil
}

func (p *ProposalEvent) getPreviousMonthTransactions(ctx context.Context, fromStart int, creatorID uint) ([]models.Transaction, error) {
	previousMonthTo := time.Now().AddDate(0, 0, int(-fromStart))
	previousMonthFrom := previousMonthTo.AddDate(0, 0, int(-fromStart))
	previousTransactions, err := p.repo.ProposalEvent.GetStatistics(ctx, creatorID, previousMonthFrom, previousMonthTo)
	if err != nil {
		return nil, err
	}
	return previousTransactions, nil
}

func (p *ProposalEvent) generateStatistics(currentTransactions, previousTransactions []models.Transaction) models.ProposalEventStatistics {
	statistics := models.ProposalEventStatistics{}
	requests := make([]models.Request, 28)
	currentMonthTo := time.Now()
	currentMonthFrom := currentMonthTo.AddDate(0, 0, -28)
	for i := 1; i <= 28; i++ {
		currenntDateResponse := currentMonthFrom.AddDate(0, 0, i)
		requests[i-1] = models.Request{
			Date:          fmt.Sprintf("%s %d", currenntDateResponse.Month().String(), currenntDateResponse.Day()),
			RequestsCount: p.getRequestsCount(currentTransactions, currentMonthFrom.AddDate(0, 0, i)),
		}
	}
	statistics.Requests = requests

	statistics.StartDate = currentMonthFrom
	statistics.EndDate = currentMonthTo
	p.generateTransactionSubStatistics(&statistics, currentTransactions, previousTransactions)
	return statistics
}

func (p *ProposalEvent) generateTransactionSubStatistics(statistics *models.ProposalEventStatistics, currentTransactions, previousTransactions []models.Transaction) {
	statistics.TransactionsCount = uint(len(currentTransactions))

	if len(currentTransactions) != 0 && len(previousTransactions) != 0 {
		statistics.TransactionsCountCompareWithPreviousMonth = compareTwoNumberInPercentage(len(currentTransactions), len(previousTransactions))
	} else if len(previousTransactions) == 0 {
		statistics.TransactionsCountCompareWithPreviousMonth = len(currentTransactions) * 100
	}

	doneTransactionsCount := len(getTransactionsByStatus(currentTransactions, models.Completed))
	previousDoneTransactionsCount := len(getTransactionsByStatus(previousTransactions, models.Completed))
	statistics.DoneTransactionsCount = uint(doneTransactionsCount)
	if doneTransactionsCount != 0 && previousDoneTransactionsCount != 0 {
		statistics.DoneTransactionsCountCompareWithPreviousMonth = compareTwoNumberInPercentage(doneTransactionsCount, previousDoneTransactionsCount)
	} else if previousDoneTransactionsCount == 0 {
		statistics.DoneTransactionsCountCompareWithPreviousMonth = doneTransactionsCount * 100
	}

	canceledTransactionsCount := len(getTransactionsByStatus(currentTransactions, models.Canceled))
	previousCanceledTransactionsCount := len(getTransactionsByStatus(previousTransactions, models.Canceled))
	statistics.CanceledTransactionCount = uint(canceledTransactionsCount)
	if canceledTransactionsCount != 0 && previousCanceledTransactionsCount != 0 {
		statistics.CanceledTransactionCountCompareWithPreviousMonth = compareTwoNumberInPercentage(canceledTransactionsCount, previousCanceledTransactionsCount)
	} else if previousCanceledTransactionsCount == 0 {
		statistics.CanceledTransactionCountCompareWithPreviousMonth = canceledTransactionsCount * 100
	}

	abortedTransactionsCount := len(getTransactionsByStatus(currentTransactions, models.Aborted))
	previousAbortedTransactionsCount := len(getTransactionsByStatus(previousTransactions, models.Aborted))
	statistics.AbortedTransactionsCount = uint(abortedTransactionsCount)
	if abortedTransactionsCount != 0 && previousAbortedTransactionsCount != 0 {
		statistics.AbortedTransactionsCountCompareWithPreviousMonth = compareTwoNumberInPercentage(abortedTransactionsCount, previousAbortedTransactionsCount)
	} else if previousAbortedTransactionsCount == 0 {
		statistics.AbortedTransactionsCountCompareWithPreviousMonth = abortedTransactionsCount * 100
	}
}

func (p *ProposalEvent) getRequestsCount(transactions []models.Transaction, from time.Time) int {
	count := 0
	for _, transaction := range transactions {
		if transaction.CreationDate.Day() == from.Day() {
			count += 1
		}
	}

	return count
}

func compareTwoNumberInPercentage(x, y int) int {
	return (x/y)*100 - 100
}

func getTransactionsByStatus(transactions []models.Transaction, transactionStatus models.TransactionStatus) []models.Transaction {
	newTransactions := []models.Transaction{}
	for i := range transactions {
		if transactions[i].TransactionStatus == transactionStatus ||
			transactions[i].ResponderStatus == transactionStatus {
			newTransactions = append(newTransactions, transactions[i])
		}
	}
	return newTransactions
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
		proposalEvent, err := p.repo.ProposalEvent.GetEvent(ctx, transaction.EventID)
		if err != nil {
			return err
		}
		proposalEvent.RemainingHelps = proposalEvent.RemainingHelps + 1
		err = p.repo.ProposalEvent.UpdateEvent(ctx, proposalEvent)
		if err != nil {
			return err
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

	err = p.repo.ProposalEvent.UpdateEvent(ctx, models.ProposalEvent{
		ID:             proposalEvent.ID,
		RemainingHelps: proposalEvent.RemainingHelps - 1,
	})
	//TODO remove after debug
	//for _, transaction := range proposalEvent.Transactions {
	//	if transaction.CreatorID == responderID && lo.Contains([]models.TransactionStatus{
	//		models.Accepted,
	//		models.InProcess,
	//		models.Waiting,
	//	}, transaction.TransactionStatus) {
	//		return fmt.Errorf("user already has transaction in this event")
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
		MemberID:      transaction.CreatorID,
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

func (p *ProposalEvent) UpdateEvent(ctx context.Context, newEvent models.ProposalEvent) error {
	oldEvent, err := p.repo.ProposalEvent.GetEvent(ctx, newEvent.ID)
	if err != nil {
		return err
	}
	if newEvent.MaxConcurrentRequests-oldEvent.MaxConcurrentRequests != 0 {
		newEvent.RemainingHelps = p.calculateRemainingHelps(oldEvent, newEvent)
	}
	return p.repo.ProposalEvent.UpdateEvent(ctx, newEvent)
}

func (p *ProposalEvent) calculateRemainingHelps(oldEvent, newEvent models.ProposalEvent) int {
	return int(newEvent.MaxConcurrentRequests-oldEvent.MaxConcurrentRequests) + oldEvent.RemainingHelps
}

func (p *ProposalEvent) DeleteEvent(ctx context.Context, id uint) error {
	return p.repo.ProposalEvent.DeleteEvent(ctx, id)
}

func (p *ProposalEvent) createNotification(ctx context.Context, notification models.TransactionNotification) error {
	_, err := p.repo.TransactionNotification.Create(ctx, notification)
	return err
}
