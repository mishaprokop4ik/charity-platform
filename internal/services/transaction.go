package service

import (
	"Kurajj/internal/models"
	"context"
	"fmt"
	"time"
)

type Transaction struct {
	repo Repositorier
}

func (t *Transaction) GetGlobalStatistics(ctx context.Context, fromStart int) (models.GlobalStatistics, error) {
	currentTransactions, err := t.getCurrentMonthTransactions(ctx, fromStart)
	if err != nil {
		return models.GlobalStatistics{}, err
	}

	previousTransactions, err := t.getPreviousMonthTransactions(ctx, fromStart)
	if err != nil {
		return models.GlobalStatistics{}, err
	}

	statistics := t.generateStatistics(currentTransactions, previousTransactions)
	return statistics, nil
}

func (t *Transaction) getCurrentMonthTransactions(ctx context.Context, fromStart int) ([]models.Transaction, error) {
	currentMonthTo := time.Now()
	currentMonthFrom := currentMonthTo.AddDate(0, 0, -fromStart)

	currentTransactions, err := t.repo.GetGlobalStatistics(ctx, currentMonthFrom, currentMonthTo)
	if err != nil {
		return nil, err
	}

	return currentTransactions, nil
}

func (t *Transaction) getPreviousMonthTransactions(ctx context.Context, fromStart int) ([]models.Transaction, error) {
	previousMonthTo := time.Now().AddDate(0, 0, -fromStart)
	previousMonthFrom := previousMonthTo.AddDate(0, 0, -fromStart)
	previousTransactions, err := t.repo.GetGlobalStatistics(ctx, previousMonthFrom, previousMonthTo)
	if err != nil {
		return nil, err
	}
	return previousTransactions, nil
}

func (t *Transaction) generateStatistics(currentTransactions, previousTransactions []models.Transaction) models.GlobalStatistics {
	statistics := models.GlobalStatistics{}
	requests := make([]models.Request, 28)
	currentMonthTo := time.Now()
	currentMonthFrom := currentMonthTo.AddDate(0, 0, -28)
	for i := 1; i <= 28; i++ {
		currenntDateResponse := currentMonthFrom.AddDate(0, 0, i)
		requests[i-1] = models.Request{
			Date:          fmt.Sprintf("%s %d", currenntDateResponse.Month().String(), currenntDateResponse.Day()),
			RequestsCount: t.getRequestsCount(currentTransactions, currentMonthFrom.AddDate(0, 0, i)),
		}
	}
	statistics.Requests = requests

	statistics.StartDate = currentMonthFrom
	statistics.EndDate = currentMonthTo
	t.generateTransactionSubStatistics(&statistics, currentTransactions, previousTransactions)
	return statistics
}

func (t *Transaction) getRequestsCount(transactions []models.Transaction, from time.Time) int {
	count := 0
	for _, transaction := range transactions {
		if transaction.CreationDate.Day() == from.Day() {
			count += 1
		}
	}

	return count
}

func (t *Transaction) generateTransactionSubStatistics(statistics *models.GlobalStatistics, currentTransactions, previousTransactions []models.Transaction) {
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

func (t *Transaction) GetTransactionByID(ctx context.Context, id uint) (models.Transaction, error) {
	return t.repo.GetTransactionByID(ctx, id)
}

func (t *Transaction) CreateTransaction(ctx context.Context, transaction models.Transaction) (uint, error) {
	return t.repo.CreateTransaction(ctx, transaction)
}

func (t *Transaction) UpdateTransaction(ctx context.Context, transaction models.Transaction) error {
	if transaction.ID != 0 {
		return t.repo.UpdateTransactionByID(ctx, transaction.ID, transaction.GetValuesToUpdate())
	}

	return t.repo.UpdateTransactionByEvent(ctx, transaction.EventID,
		transaction.EventType,
		transaction.GetValuesToUpdate())
}

func (t *Transaction) GetCurrentEventTransactions(ctx context.Context,
	eventID uint,
	eventType models.EventType) ([]models.Transaction, error) {
	return t.repo.GetCurrentEventTransactions(ctx, eventID, eventType)
}

func (t *Transaction) UpdateAllNotFinishedTransactions(ctx context.Context,
	eventID uint,
	eventType models.EventType,
	newStatus models.TransactionStatus) error {
	return t.repo.UpdateAllNotFinishedTransactions(ctx, eventID, eventType, newStatus)
}

func (t *Transaction) GetAllEventTransactions(ctx context.Context,
	eventID uint,
	eventType models.EventType) ([]models.Transaction, error) {
	return t.repo.GetAllEventTransactions(ctx, eventID, eventType)
}

func NewTransaction(repo Repositorier) *Transaction {
	return &Transaction{repo: repo}
}
