package service

import (
	"Kurajj/internal/models"
	"Kurajj/internal/repository"
	"context"
	"database/sql"
	"fmt"
	"github.com/samber/lo"
	"time"
)

func NewHelpEvent(r *repository.Repository) *HelpEvent {
	return &HelpEvent{repo: r, Transaction: NewTransaction(r)}
}

type HelpEvent struct {
	*Transaction
	repo *repository.Repository
}

func (h *HelpEvent) GetHelpEventBySearch(ctx context.Context, search models.HelpSearchInternal) (models.HelpEventPagination, error) {
	return h.repo.HelpEvent.GetEventsWithSearchAndSort(ctx, search)
}

func (h *HelpEvent) GetUserHelpEvents(ctx context.Context, userID models.ID) ([]models.HelpEvent, error) {
	return h.repo.HelpEvent.GetUserHelpEvents(ctx, userID)
}

func (h *HelpEvent) GetHelpEventByTransactionID(ctx context.Context, transactionID models.ID) (models.HelpEvent, error) {
	return h.repo.HelpEvent.GetHelpEventByTransactionID(ctx, transactionID)
}

func (h *HelpEvent) UpdateTransactionStatus(ctx context.Context, transaction models.HelpEventTransaction) error {
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
		for i := range eventNeeds {
			transactionNeed, transactionNeedIndex, _ := lo.FindIndexOf(transaction.Needs, func(n models.Need) bool {
				return n.Title == eventNeeds[i].Title
			})
			if transactionNeedIndex == -1 {
				return fmt.Errorf("cannot find need with %s title", eventNeeds[i].Title)
			}
			eventNeeds[i].ReceivedTotal = transactionNeed.ReceivedTotal
			//transaction.Needs[transactionNeedIndex].ReceivedTotal = transactionNeed.ReceivedTotal
			//transaction.Needs[transactionNeedIndex].Received = transactionNeed.ReceivedTotal
		}
		err = h.updateNeeds(ctx, eventNeeds...)
		if err != nil {
			return err
		}
		if transaction.TransactionStatus == models.Completed {
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
		return n.Amount == n.ReceivedTotal
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
	//for _, transaction := range proposalEvent.Transactions {
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
