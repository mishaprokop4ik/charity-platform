package service

import (
	"Kurajj/internal/models"
	"Kurajj/internal/repository"
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"io"
	"time"
)

func NewHelpEvent(r *repository.Repository) *HelpEvent {
	return &HelpEvent{repo: r, Transaction: NewTransaction(r)}
}

type HelpEvent struct {
	*Transaction
	repo *repository.Repository
}

func (h *HelpEvent) GetStatistics(ctx context.Context, fromStart int, creatorID uint) (models.HelpEventStatistics, error) {
	currentTransactions, err := h.getCurrentMonthTransactions(ctx, fromStart, creatorID)
	if err != nil {
		return models.HelpEventStatistics{}, err
	}

	previousTransactions, err := h.getPreviousMonthTransactions(ctx, fromStart, creatorID)
	if err != nil {
		return models.HelpEventStatistics{}, err
	}

	statistics := h.generateStatistics(currentTransactions, previousTransactions)
	return statistics, nil
}

func (h *HelpEvent) getCurrentMonthTransactions(ctx context.Context, fromStart int, creatorID uint) ([]models.Transaction, error) {
	currentMonthTo := time.Now()
	currentMonthFrom := currentMonthTo.AddDate(0, 0, int(-fromStart))

	currentTransactions, err := h.repo.HelpEvent.GetStatistics(ctx, creatorID, currentMonthFrom, currentMonthTo)
	if err != nil {
		return nil, err
	}

	return currentTransactions, nil
}

func (h *HelpEvent) getPreviousMonthTransactions(ctx context.Context, fromStart int, creatorID uint) ([]models.Transaction, error) {
	previousMonthTo := time.Now().AddDate(0, 0, int(-fromStart))
	previousMonthFrom := previousMonthTo.AddDate(0, 0, int(-fromStart))
	previousTransactions, err := h.repo.HelpEvent.GetStatistics(ctx, creatorID, previousMonthFrom, previousMonthTo)
	if err != nil {
		return nil, err
	}
	return previousTransactions, nil
}

func (h *HelpEvent) generateStatistics(currentTransactions, previousTransactions []models.Transaction) models.HelpEventStatistics {
	statistics := models.HelpEventStatistics{}
	requests := make([]models.Request, 28)
	currentMonthTo := time.Now()
	currentMonthFrom := currentMonthTo.AddDate(0, 0, int(-28))
	for i := 1; i <= 28; i++ {
		currenntDateResponse := currentMonthFrom.AddDate(0, 0, i)
		requests[i-1] = models.Request{
			Date:          fmt.Sprintf("%s %d", currenntDateResponse.Month().String(), currenntDateResponse.Day()),
			RequestsCount: h.getRequestsCount(currentTransactions, currentMonthFrom.AddDate(0, 0, i)),
		}
	}
	fmt.Println(requests)
	statistics.Requests = requests

	statistics.StartDate = currentMonthFrom
	statistics.EndDate = currentMonthTo
	h.generateTransactionSubStatistics(&statistics, currentTransactions, previousTransactions)
	return statistics
}

func (h *HelpEvent) generateTransactionSubStatistics(statistics *models.HelpEventStatistics, currentTransactions, previousTransactions []models.Transaction) {
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

func (h *HelpEvent) getRequestsCount(transactions []models.Transaction, from time.Time) int {
	count := 0
	for _, transaction := range transactions {
		if transaction.CreationDate.Day() == from.Day() {
			count += 1
		}
	}

	return count
}

func (h *HelpEvent) UpdateEvent(ctx context.Context, event models.HelpEvent) error {
	return h.repo.HelpEvent.UpdateHelpEvent(ctx, event)
}

func (h *HelpEvent) GetHelpEventBySearch(ctx context.Context, search models.HelpSearchInternal) (models.HelpEventPagination, error) {
	events, err := h.repo.HelpEvent.GetEventsWithSearchAndSort(ctx, search)
	if err != nil {
		return models.HelpEventPagination{}, err
	}
	for i := range events.Events {
		events.Events[i].CalculateCompletionPercentages()
		events.Events[i].CalculateTransactionsCompletionPercentages()
	}
	return events, nil
}

func (h *HelpEvent) GetUserHelpEvents(ctx context.Context, userID models.ID) ([]models.HelpEvent, error) {
	events, err := h.repo.HelpEvent.GetUserHelpEvents(ctx, userID)
	if err != nil {
		return nil, err
	}
	for i := range events {
		events[i].CalculateCompletionPercentages()
		events[i].CalculateTransactionsCompletionPercentages()
	}
	return events, nil
}

func (h *HelpEvent) GetHelpEventByTransactionID(ctx context.Context, transactionID models.ID) (models.HelpEvent, error) {
	return h.repo.HelpEvent.GetHelpEventByTransactionID(ctx, transactionID)
}

func (h *HelpEvent) UpdateTransactionStatus(ctx context.Context, transaction models.HelpEventTransaction,
	file io.Reader, fileType string) error {
	oldTransaction, err := h.GetTransactionByID(ctx, *transaction.TransactionID)
	if err != nil {
		return err
	}
	var notificationReceiver uint
	if transaction.EventCreator {
		oldTransaction.UpdateStatus(!transaction.EventCreator, transaction.TransactionStatus)
		notificationReceiver = transaction.TransactionCreatorID
		oldTransaction.CompetitionDate = sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		}
		eventNeeds, err := h.repo.HelpEvent.GetHelpEventNeeds(ctx, models.ID(*transaction.HelpEventID))
		if err != nil {
			return err
		}
		if transaction.TransactionStatus == models.Completed {
			transactionNeeds, err := h.repo.HelpEvent.GetTransactionNeeds(ctx, models.ID(*transaction.TransactionID))
			if err != nil {
				return err
			}
			for i := range eventNeeds {
				transactionNeed, transactionNeedIndex, _ := lo.FindIndexOf(transactionNeeds, func(n models.Need) bool {
					return n.Title == eventNeeds[i].Title && n.Unit == eventNeeds[i].Unit && n.Amount == eventNeeds[i].Amount
				})
				if transactionNeedIndex == -1 {
					continue
				}
				eventNeeds[i].ReceivedTotal = transactionNeed.Received + eventNeeds[i].ReceivedTotal
			}
			err = h.updateNeeds(ctx, eventNeeds...)
			if err != nil {
				return err
			}
			err = h.CompleteHelpEvent(ctx, *transaction.HelpEventID, eventNeeds)
			if err != nil {
				return err
			}
		}
	} else {
		notificationReceiver = transaction.HelpEventCreatorID
		oldTransaction.UpdateStatus(!transaction.EventCreator, transaction.ResponderStatus)
		err = h.updateNeeds(ctx, transaction.Needs...)
		if err != nil {
			return err
		}

		if transaction.ResponderStatus == models.Completed {
			fileUniqueID, err := uuid.NewUUID()
			if err != nil {
				return err
			}
			fileName := fmt.Sprintf("%s.%s", fileUniqueID.String(), fileType)
			filePath, err := h.repo.File.Upload(ctx, fileName, file)
			if err != nil {
				return err
			}
			oldTransaction.ReportURL = filePath
		}
	}

	err = h.UpdateTransaction(ctx, oldTransaction)
	if err != nil {
		return err
	}

	err = h.createNotification(ctx, models.TransactionNotification{
		EventType:     models.HelpEventType,
		EventID:       *transaction.HelpEventID,
		Action:        models.Updated,
		TransactionID: oldTransaction.ID,
		IsRead:        false,
		CreationTime:  time.Now(),
		MemberID:      notificationReceiver,
	})

	return err
}

func (h *HelpEvent) CompleteHelpEvent(ctx context.Context, helpEventID uint, eventNeeds []models.Need) error {
	allNeedsCompleted := lo.CountBy(eventNeeds, func(n models.Need) bool {
		return n.Amount == n.Received
	}) == len(eventNeeds)

	if allNeedsCompleted {
		oldHelpEvent, err := h.repo.HelpEvent.GetEventByID(ctx, models.ID(helpEventID))
		if err != nil {
			return err
		}

		oldHelpEvent.Status = models.Done

		err = h.repo.HelpEvent.UpdateHelpEvent(ctx, oldHelpEvent)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *HelpEvent) CreateHelpEvent(ctx context.Context, event *models.HelpEvent) (uint, error) {
	return h.repo.HelpEvent.CreateEvent(ctx, event)
}

func (h *HelpEvent) GetHelpEventByID(ctx context.Context, id models.ID) (models.HelpEvent, error) {
	helpEvent, err := h.repo.HelpEvent.GetEventByID(ctx, id)
	if err != nil {
		return models.HelpEvent{}, err
	}

	helpEvent.CalculateTransactionsCompletionPercentages()

	helpEvent.CalculateCompletionPercentages()
	return helpEvent, nil
}

func (h *HelpEvent) CreateRequest(ctx context.Context, userID models.ID, transactionInfo models.TransactionAcceptCreateRequest) (uint, error) {
	helpEvent, err := h.repo.HelpEvent.GetEventByID(ctx, models.ID(transactionInfo.ID))
	if err != nil {
		return 0, err
	}
	//if models.ID(helpEvent.CreatedBy) == userID {
	//	return fmt.Errorf("event creator cannot response his/her own events")
	//}
	//TODO remove after debug
	//for _, transaction := range helpEvent.Transactions {
	//	if transaction.CreatorID == responderID && lo.Contains([]models.TransactionStatus{
	//		models.Accepted,
	//		models.InProcess,
	//		models.Waiting,
	//	}, transaction.TransactionStatus) {
	//		return fmt.Errorf("user already has transaction in this event")
	//	}
	//}
	transactionID, err := h.CreateTransaction(ctx, models.Transaction{
		CreatorID:         uint(userID),
		EventID:           uint(transactionInfo.ID),
		Comment:           transactionInfo.Comment,
		EventType:         models.HelpEventType,
		CreationDate:      time.Now(),
		TransactionStatus: models.Waiting,
		ResponderStatus:   models.NotStarted,
	})
	if err != nil {
		return 0, err
	}

	helpEventNeeds, err := h.repo.HelpEvent.GetHelpEventNeeds(ctx, models.ID(transactionInfo.ID))
	if err != nil {
		return 0, err
	}

	for i := range helpEventNeeds {
		helpEventNeeds[i].TransactionID = &transactionID
		helpEventNeeds[i].ID = 0
		helpEventNeeds[i].Received = 0
		_, err := h.repo.HelpEvent.CreateNeed(ctx, helpEventNeeds[i])
		if err != nil {
			return 0, err
		}
	}

	err = h.createNotification(ctx, models.TransactionNotification{
		EventType:     models.HelpEventType,
		EventID:       uint(transactionInfo.ID),
		Action:        models.Created,
		TransactionID: transactionID,
		IsRead:        false,
		CreationTime:  time.Now(),
		MemberID:      helpEvent.CreatedBy,
	})

	return transactionID, err
}

func (h *HelpEvent) createNotification(ctx context.Context, notification models.TransactionNotification) error {
	_, err := h.repo.TransactionNotification.Create(ctx, notification)
	return err
}

func (h *HelpEvent) updateNeeds(ctx context.Context, needs ...models.Need) error {
	return h.repo.HelpEvent.UpdateNeeds(ctx, needs...)
}
