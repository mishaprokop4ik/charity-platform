package service

import (
	"Kurajj/internal/models"
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"github.com/samber/lo"
	"io"
	"time"
	_ "time/tzdata"
)

func NewHelpEvent(r Repositorier) HelpEventer {
	helpEventService := &HelpEvent{repo: r, Transaction: NewTransaction(r), MaxEventsPerUser: 5}
	helpEventCron := cron.New()
	helpEventCron.AddFunc("@every 1m", helpEventService.provisionEvents)
	helpEventCron.Start()
	return helpEventService
}

type HelpEvent struct {
	*Transaction
	repo             Repositorier
	MaxEventsPerUser uint
}

func (h *HelpEvent) GetHelpEventStatistics(ctx context.Context, fromStart int, creatorID uint) (models.HelpEventStatistics, error) {
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

func (h *HelpEvent) provisionEvents() {
	ctx := context.Background()
	events, err := h.repo.GetAllHelpEvents(ctx)
	if err != nil {
		fmt.Println(err)
	}
	//loc, err := time.LoadLocation("Europe/Kiev")
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	currentTime := time.Now().Add(3 * time.Hour)
	for _, e := range events {
		endDate := e.EndDate
		if e.Status == models.Active && currentTime.After(endDate) {
			e.Status = models.Done
			err = h.repo.UpdateHelpEvent(ctx, e)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func (h *HelpEvent) getCurrentMonthTransactions(ctx context.Context, fromStart int, creatorID uint) ([]models.Transaction, error) {
	currentMonthTo := time.Now()
	currentMonthFrom := currentMonthTo.AddDate(0, 0, -fromStart)

	currentTransactions, err := h.repo.GetHelpEventStatistics(ctx, creatorID, currentMonthFrom, currentMonthTo)
	if err != nil {
		return nil, err
	}

	return currentTransactions, nil
}

func (h *HelpEvent) getPreviousMonthTransactions(ctx context.Context, fromStart int, creatorID uint) ([]models.Transaction, error) {
	previousMonthTo := time.Now().AddDate(0, 0, -fromStart)
	previousMonthFrom := previousMonthTo.AddDate(0, 0, -fromStart)
	previousTransactions, err := h.repo.GetHelpEventStatistics(ctx, creatorID, previousMonthFrom, previousMonthTo)
	if err != nil {
		return nil, err
	}
	return previousTransactions, nil
}

func (h *HelpEvent) generateStatistics(currentTransactions, previousTransactions []models.Transaction) models.HelpEventStatistics {
	statistics := models.HelpEventStatistics{}
	requests := make([]models.Request, 28)
	currentMonthTo := time.Now()
	currentMonthFrom := currentMonthTo.AddDate(0, 0, -28)
	for i := 1; i <= 28; i++ {
		currentDateResponse := currentMonthFrom.AddDate(0, 0, i)
		requests[i-1] = models.Request{
			Date:          fmt.Sprintf("%s %d", currentDateResponse.Month().String(), currentDateResponse.Day()),
			RequestsCount: h.getRequestsCount(currentTransactions, currentMonthFrom.AddDate(0, 0, i)),
		}
	}
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

func (h *HelpEvent) UpdateHelpEvent(ctx context.Context, event models.HelpEvent) error {
	return h.repo.UpdateHelpEvent(ctx, event)
}

func (h *HelpEvent) GetHelpEventBySearch(ctx context.Context, search models.HelpSearchInternal) (models.HelpEventPagination, error) {
	events, err := h.repo.GetHelpEventsWithSearchAndSort(ctx, search)
	if err != nil {
		return models.HelpEventPagination{}, err
	}
	for i := range events.Events {
		events.Events[i].CalculateCompletionPercentages()
	}
	return events, nil
}

func (h *HelpEvent) GetUserHelpEvents(ctx context.Context, userID models.ID) ([]models.HelpEvent, error) {
	events, err := h.repo.GetUserHelpEvents(ctx, userID)
	if err != nil {
		return nil, err
	}
	for i := range events {
		events[i].CalculateCompletionPercentages()
	}
	return events, nil
}

func (h *HelpEvent) GetHelpEventByTransactionID(ctx context.Context, transactionID models.ID) (models.HelpEvent, error) {
	return h.repo.GetHelpEventByTransactionID(ctx, transactionID)
}

func (h *HelpEvent) UpdateTransactionStatus(ctx context.Context, transaction models.HelpEventTransaction,
	file io.Reader, fileType, createdFilePath string) error {
	oldTransaction, err := h.GetTransactionByID(ctx, *transaction.TransactionID)
	if err != nil {
		return err
	}
	var notificationReceiver uint
	notificationStatus := models.TransactionStatus("")
	if transaction.EventCreator {
		notificationStatus = transaction.TransactionStatus
		oldTransaction.UpdateStatus(!transaction.EventCreator, transaction.TransactionStatus)
		oldTransaction.CompetitionDate = sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		}
		notificationReceiver = oldTransaction.CreatorID
		if transaction.TransactionStatus == models.Completed {
			eventNeeds, err := h.repo.GetHelpEventNeeds(ctx, models.ID(*transaction.HelpEventID))
			if err != nil {
				return err
			}
			transactionNeeds, err := h.repo.GetTransactionNeeds(ctx, models.ID(*transaction.TransactionID))
			if err != nil {
				return err
			}
			for i, eventNeed := range eventNeeds {
				transactionNeed, transactionNeedIndex, _ := lo.FindIndexOf(transactionNeeds, func(n models.Need) bool {
					return n.Title == eventNeed.Title && n.Unit == eventNeed.Unit && n.Amount == eventNeed.Amount
				})
				if transactionNeedIndex == -1 {
					continue
				}
				eventNeeds[i].ReceivedTotal += transactionNeed.Received
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
		notificationStatus = transaction.ResponderStatus
		helpEvent, err := h.repo.GetEventByID(ctx, models.ID(oldTransaction.EventID))
		if err != nil {
			return err
		}
		notificationReceiver = helpEvent.CreatedBy
		oldTransaction.UpdateStatus(!transaction.EventCreator, transaction.ResponderStatus)
		err = h.updateNeeds(ctx, transaction.Needs...)
		if err != nil {
			return err
		}

		if transaction.ResponderStatus == models.Completed {
			if createdFilePath != "" {
				oldTransaction.ReportURL = createdFilePath
			} else {
				fileUniqueID, err := uuid.NewUUID()
				if err != nil {
					return err
				}
				fileName := fmt.Sprintf("%s.%s", fileUniqueID.String(), fileType)
				filePath, err := h.repo.Upload(ctx, fileName, file)
				if err != nil {
					return err
				}
				oldTransaction.ReportURL = filePath
			}
		}
	}

	err = h.UpdateTransaction(ctx, oldTransaction)
	if err != nil {
		return err
	}

	err = h.createNotification(ctx, models.TransactionNotification{
		EventType:     models.HelpEventType,
		EventID:       oldTransaction.EventID,
		Action:        models.Updated,
		TransactionID: oldTransaction.ID,
		IsRead:        false,
		NewStatus:     notificationStatus,
		CreationTime:  time.Now(),
		MemberID:      notificationReceiver,
	})

	return err
}

func (h *HelpEvent) CompleteHelpEvent(ctx context.Context, helpEventID uint, eventNeeds []models.Need) error {
	allNeedsCompleted := lo.CountBy(eventNeeds, func(n models.Need) bool {
		return n.Amount == n.ReceivedTotal
	}) == len(eventNeeds)

	if allNeedsCompleted {
		oldHelpEvent, err := h.repo.GetEventByID(ctx, models.ID(helpEventID))
		if err != nil {
			return err
		}

		oldHelpEvent.Status = models.Done

		err = h.repo.UpdateHelpEvent(ctx, oldHelpEvent)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *HelpEvent) CreateHelpEvent(ctx context.Context, event *models.HelpEvent) (uint, error) {
	userID := models.ID(event.CreatedBy)
	events, err := h.GetUserHelpEvents(ctx, userID)
	if err != nil {
		return 0, err
	}
	currentEventsCount := uint(0)
	for _, event := range events {
		if event.Status == models.Active {
			currentEventsCount += 1
		}

		if currentEventsCount >= h.MaxEventsPerUser {
			return 0, fmt.Errorf("user cannot create more than %d events", h.MaxEventsPerUser)
		}

	}
	return h.repo.CreateEvent(ctx, event)
}

func (h *HelpEvent) GetHelpEventByID(ctx context.Context, id models.ID) (models.HelpEvent, error) {
	helpEvent, err := h.repo.GetEventByID(ctx, id)
	if err != nil {
		return models.HelpEvent{}, err
	}

	helpEvent.CalculateCompletionPercentages()
	return helpEvent, nil
}

func (h *HelpEvent) CreateRequest(ctx context.Context, userID models.ID, transactionInfo models.TransactionAcceptCreateRequest) (uint, error) {
	helpEvent, err := h.repo.GetEventByID(ctx, models.ID(transactionInfo.ID))
	if err != nil {
		return 0, err
	}
	if models.ID(helpEvent.CreatedBy) == userID {
		return 0, fmt.Errorf("event creator cannot response his/her own events")
	}
	for _, transaction := range helpEvent.Transactions {
		if models.ID(transaction.CreatorID) == userID && lo.Contains([]models.TransactionStatus{
			models.Accepted,
			models.InProcess,
			models.Waiting,
		}, transaction.TransactionStatus) {
			return 0, fmt.Errorf("user already has transaction in this event")
		}
	}
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

	helpEventNeeds, err := h.repo.GetHelpEventNeeds(ctx, models.ID(transactionInfo.ID))
	if err != nil {
		return 0, err
	}

	for i := range helpEventNeeds {
		helpEventNeeds[i].TransactionID = &transactionID
		helpEventNeeds[i].ID = 0
		helpEventNeeds[i].Received = 0
		_, err := h.repo.CreateNeed(ctx, helpEventNeeds[i])
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
	_, err := h.repo.CreateNotification(ctx, notification)
	return err
}

func (h *HelpEvent) updateNeeds(ctx context.Context, needs ...models.Need) error {
	return h.repo.UpdateNeeds(ctx, needs...)
}
