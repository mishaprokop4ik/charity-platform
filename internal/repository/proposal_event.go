package repository

import (
	"Kurajj/internal/models"
	zlog "Kurajj/pkg/logger"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/samber/lo"
	"gorm.io/gorm"
	"strings"
	"time"
)

const defaultSortField = "creation_date"

type ProposalEventer interface {
	proposalEventCRUDer
	GetUserProposalEvents(ctx context.Context, userID uint) ([]models.ProposalEvent, error)
	GetEventsWithSearchAndSort(ctx context.Context,
		searchValues models.ProposalEventSearchInternal) ([]models.ProposalEvent, error)
}

type proposalEventCRUDer interface {
	CreateEvent(ctx context.Context, event models.ProposalEvent) (uint, error)
	GetEvent(ctx context.Context, id uint) (models.ProposalEvent, error)
	GetEvents(ctx context.Context) ([]models.ProposalEvent, error)
	UpdateEvent(ctx context.Context, id uint, toUpdate map[string]any) error
	DeleteEvent(ctx context.Context, id uint) error
}

type ProposalEvent struct {
	DBConnector *Connector
}

func (p *ProposalEvent) GetEventsWithSearchAndSort(ctx context.Context,
	searchValues models.ProposalEventSearchInternal) ([]models.ProposalEvent, error) {
	events := []models.ProposalEvent{}
	searchValues = p.removeEmptySearchValues(searchValues)
	query := p.DBConnector.DB.
		Order(fmt.Sprintf("%s %s", searchValues.SortField, strings.ToUpper(string(*searchValues.Order)))).
		Where("status IN (?)", searchValues.State)

	if searchValues.GetOwn != nil && searchValues.SearcherID != nil {
		fmt.Println("1")
		if *searchValues.GetOwn {
			query = query.Where("author_id = ?", searchValues.SearcherID)
		} else {
			query = query.Not("author_id = ?", searchValues.SearcherID)
		}
	}

	if searchValues.Name != nil && *searchValues.Name != "" {
		fmt.Println("2")
		query = query.Where("title = ?", *searchValues.Name)
	}

	if searchValues.TakingPart != nil {
		fmt.Println("3")
		if *searchValues.TakingPart {
			query = query.Joins("JOIN transaction ON transaction.event_id = propositional_event.id").
				Where("transaction.creator_id = ?", searchValues.SearcherID).
				Distinct("propositional_event.*")
		} else {
			query = query.Joins("JOIN transaction ON transaction.event_id = propositional_event.id").
				Not("transaction.creator_id = ?", searchValues.SearcherID).
				Distinct("propositional_event.*")
		}
	}

	if searchValues.Tags != nil && len(*searchValues.Tags) > 0 {
		fmt.Println("4")
		if len(searchValues.GetTagsValues()) == 0 {
			subQuery := p.DBConnector.DB.Table("tag").Select("event_id").
				Where("LOWER(title) IN (?) AND event_type = ?", searchValues.GetTagsTitle(), models.ProposalEventType)

			query = query.Where("id IN (?)", subQuery)
		} else {
			subQuery := p.DBConnector.DB.Table("tag").Select("event_id").
				Joins("JOIN tag_value ON tag.id = tag_value.tag_id").
				Where("LOWER(tag.title) IN (?) AND LOWER(tag_value.value) IN (?) AND tag.event_type = ?",
					searchValues.GetTagsTitle(), searchValues.GetTagsValues(), models.ProposalEventType)

			query = query.Where("id IN (?)", subQuery)
		}
	}
	if searchValues.Location != nil {
		fmt.Println("5")
		location := *searchValues.Location
		subQuery := p.DBConnector.DB.Table("location").Select("event_id").
			Where("event_type = ?", models.ProposalEventType)

		if location.Region != "" {
			subQuery = subQuery.Where("LOWER(region) LIKE ?", "%"+location.Region+"%")
		}
		if location.City != "" {
			subQuery = subQuery.Where("LOWER(city) LIKE ?", "%"+location.City+"%")
		}
		if location.District != "" {
			subQuery = subQuery.Where("LOWER(district) LIKE ?", "%"+location.District+"%")
		}
		if location.HomeLocation != "" {
			subQuery = subQuery.Where("LOWER(home) LIKE ?", "%"+location.HomeLocation+"%")
		}

		query = query.Where("id IN (?)", subQuery)
	}

	err := query.Find(&events).WithContext(ctx).Error
	if err != nil {
		return nil, err
	}

	events, err = p.insertUserInProposalEvents(ctx, events)
	if err != nil {
		return nil, err
	}

	for i, event := range events {
		updatedEvent, err := p.updateMissingEventData(ctx, event)
		if err != nil {
			return nil, err
		}
		events[i] = updatedEvent
	}

	return events, nil
}

func (p *ProposalEvent) removeEmptySearchValues(searchValues models.ProposalEventSearchInternal) models.ProposalEventSearchInternal {
	boolRef := func(b bool) *bool {
		return &b
	}
	newSearchValues := models.ProposalEventSearchInternal{}
	if searchValues.SortField == "" {
		newSearchValues.SortField = defaultSortField
	} else {
		newSearchValues.SortField = searchValues.SortField
	}

	if searchValues.Name != nil {
		newSearchValues.Name = searchValues.Name
	}

	if searchValues.Tags != nil {
		newSearchValues.Tags = searchValues.Tags
	} else {
		newSearchValues.Tags = nil
	}

	if searchValues.GetOwn == nil {
		newSearchValues.GetOwn = boolRef(false)
	} else {
		newSearchValues.GetOwn = searchValues.GetOwn
	}

	if searchValues.TakingPart == nil {
		newSearchValues.GetOwn = boolRef(false)
	} else {
		newSearchValues.GetOwn = searchValues.TakingPart
	}

	if searchValues.State == nil {
		newSearchValues.State = []models.EventStatus{
			models.Active,
			models.Done,
		}
	} else {
		newSearchValues.State = searchValues.State
	}

	if searchValues.Location == nil {
		newSearchValues.Location = nil
	} else {
		newSearchValues.Location = searchValues.Location
	}
	if newSearchValues.Order == nil {
		newSearchValues.Order = &models.AscendingOrder
	} else {
		newSearchValues.Order = searchValues.Order
	}
	return newSearchValues
}

func (p *ProposalEvent) CreateEvent(ctx context.Context, event models.ProposalEvent) (uint, error) {
	tx := p.DBConnector.DB.Begin()
	err := p.DBConnector.DB.
		Create(&event).
		WithContext(ctx).
		Error

	if err != nil {
		tx.Rollback()
		return 0, err
	}

	if !event.Location.IsEmpty() {
		event.Location.EventID = event.ID
		err = p.DBConnector.DB.
			Create(&event.Location).
			WithContext(ctx).
			Error

		if err != nil {
			tx.Rollback()
			return 0, err
		}
	}

	for _, tag := range event.Tags {
		tag.EventID = event.ID
		err = p.DBConnector.DB.Create(&tag).WithContext(ctx).Error
		if err != nil {
			tx.Rollback()
			return 0, err
		}
		for _, tagValue := range tag.Values {
			tagValue.TagID = tag.ID
			err = p.DBConnector.DB.Create(&tagValue).WithContext(ctx).Error
			if err != nil {
				tx.Rollback()
				return 0, err
			}
		}
	}

	return event.ID, tx.Commit().Error
}

func (p *ProposalEvent) GetEvent(ctx context.Context, id uint) (models.ProposalEvent, error) {
	event := models.ProposalEvent{}
	err := p.DBConnector.DB.
		Where("id = ?", id).
		First(&event).
		Where("is_deleted = ?", false).
		WithContext(ctx).
		Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.ProposalEvent{}, fmt.Errorf("cound not get proposal event by id %d", id)
	} else if err != nil {
		return models.ProposalEvent{}, err
	}

	event, err = p.updateMissingEventData(ctx, event)
	if err != nil {
		return models.ProposalEvent{}, nil
	}

	events, err := p.insertUserInProposalEvents(ctx, []models.ProposalEvent{
		event,
	})
	if err != nil {
		return models.ProposalEvent{}, err
	}
	return events[0], nil
}

func (p *ProposalEvent) GetEvents(ctx context.Context) ([]models.ProposalEvent, error) {
	events := []models.ProposalEvent{}
	resp := p.DBConnector.DB.
		Where("is_deleted = ?", false).
		Find(&events).
		WithContext(ctx)

	if errors.Is(resp.Error, gorm.ErrRecordNotFound) {
		return []models.ProposalEvent{}, fmt.Errorf("could not get any proposal events")
	} else if resp.Error != nil {
		return nil, resp.Error
	}

	events, err := p.insertUserInProposalEvents(ctx, events)
	if err != nil {
		return nil, err
	}

	for i, event := range events {
		newEvent, err := p.updateMissingEventData(ctx, event)
		if err != nil {
			return nil, err
		}
		events[i] = newEvent
	}

	return events, nil
}

func (p *ProposalEvent) UpdateEvent(ctx context.Context, id uint, toUpdate map[string]any) error {
	return p.DBConnector.DB.
		Model(&models.ProposalEvent{}).
		Select(lo.Keys(toUpdate)).
		Where("id = ?", id).
		Updates(toUpdate).
		WithContext(ctx).
		Error
}

func (p *ProposalEvent) DeleteEvent(ctx context.Context, id uint) error {
	tx := p.DBConnector.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	oldProposalEvent := &models.ProposalEvent{}
	err := p.DBConnector.DB.Where("id = ?", id).WithContext(ctx).First(oldProposalEvent).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	oldProposalEvent.CompetitionDate = sql.NullTime{Time: time.Now(), Valid: true}
	oldProposalEvent.IsDeleted = true
	err = p.DBConnector.DB.Where("id = ?", id).Updates(oldProposalEvent).WithContext(ctx).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	err = p.DBConnector.DB.
		Model(&models.Transaction{}).
		Where("event_id = ?", id).
		Where("event_type = ?", models.ProposalEventType).
		Not("status IN (?)", []models.TransactionStatus{models.Completed, models.Interrupted, models.Canceled}).
		Update("status", models.Canceled).
		WithContext(ctx).
		Error
	if err != nil {
		tx.Rollback()
		return err
	}
	err = p.DBConnector.DB.
		Model(&models.Comment{}).
		Where("event_id = ?", id).
		Where("event_type = ?", models.ProposalEventType).
		Update("is_deleted", true).
		WithContext(ctx).
		Error
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (p *ProposalEvent) GetUserProposalEvents(ctx context.Context, userID uint) ([]models.ProposalEvent, error) {
	events := []models.ProposalEvent{}
	resp := p.DBConnector.DB.
		Where("author_id = ?", userID).
		Where("is_deleted", false).
		Find(&events).
		WithContext(ctx)

	if errors.Is(resp.Error, gorm.ErrRecordNotFound) {
		return []models.ProposalEvent{}, fmt.Errorf("could not get any user proposal events")
	} else if resp.Error != nil {
		return nil, resp.Error
	}

	events, err := p.insertUserInProposalEvents(ctx, events)
	if err != nil {
		return nil, err
	}

	for i, event := range events {
		newEvent, err := p.updateMissingEventData(ctx, event)
		if err != nil {
			return nil, err
		}
		events[i] = newEvent
	}

	return events, nil
}

func (p *ProposalEvent) insertUserInProposalEvents(ctx context.Context, events []models.ProposalEvent) ([]models.ProposalEvent, error) {
	for i, event := range events {
		memberID := event.AuthorID
		member := models.User{}
		err := p.DBConnector.DB.Where("id = ?", memberID).First(&member).WithContext(ctx).Error
		if err != nil {
			return []models.ProposalEvent{}, err
		}
		events[i].User = member
	}
	fmt.Println(events)
	return events, nil
}

func (p *ProposalEvent) getProposalEventTransactions(ctx context.Context, eventID uint) ([]models.Transaction, error) {
	transactions := []models.Transaction{}
	err := p.DBConnector.DB.
		Find(&transactions).
		Where("event_id = ?", eventID).
		Where("event_type = ?", models.ProposalEventType).
		WithContext(ctx).
		Error
	if err != nil {
		return []models.Transaction{}, err
	}
	for i, transaction := range transactions {
		newTransaction, err := p.updateTransactionUsers(ctx, transaction)
		if err != nil {
			return nil, err
		}
		transactions[i] = newTransaction
	}
	return transactions, err
}

func (p *ProposalEvent) updateTransactionUsers(ctx context.Context, transaction models.Transaction) (models.Transaction, error) {
	creatorInfo := models.User{}
	err := p.DBConnector.DB.Where("id = ?", transaction.CreatorID).First(&creatorInfo).WithContext(ctx).Error
	if err != nil {
		return models.Transaction{}, err
	}

	transaction.Creator = creatorInfo

	rootEvent := models.ProposalEvent{}

	err = p.DBConnector.DB.Where("id = ?", transaction.EventID).First(&rootEvent).WithContext(ctx).Error
	if err != nil {
		return models.Transaction{}, err
	}

	responderInfo := models.User{}
	err = p.DBConnector.DB.Where("id = ?", rootEvent.AuthorID).First(&responderInfo).WithContext(ctx).Error
	if err != nil {
		return models.Transaction{}, err
	}

	transaction.Responder = responderInfo

	return transaction, nil
}

func (p *ProposalEvent) getProposalEventComments(ctx context.Context, eventID uint) ([]models.Comment, error) {
	comments := []models.Comment{}
	err := p.DBConnector.DB.
		Where("event_type = ?", models.ProposalEventType).
		Where("event_id = ?", eventID).
		Where("is_deleted = ?", false).
		Find(&comments).
		WithContext(ctx).
		Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return comments, nil
	}
	return comments, err
}

func (p *ProposalEvent) updateCommentUsers(ctx context.Context, comment models.Comment) (models.Comment, error) {
	creatorInfo := models.User{}
	err := p.DBConnector.DB.Where("id = ?", comment.UserID).First(&creatorInfo).WithContext(ctx).Error
	if err != nil {
		return models.Comment{}, err
	}

	comment.UserValues = models.UserShortInfo{
		ID:              creatorInfo.ID,
		Username:        creatorInfo.FullName,
		ProfileImageURL: creatorInfo.AvatarImagePath,
		PhoneNumber:     models.Telephone(creatorInfo.Telephone),
	}

	return comment, nil
}

func (p *ProposalEvent) updateMissingEventData(ctx context.Context, proposalEvent models.ProposalEvent) (models.ProposalEvent, error) {
	comments, err := p.getProposalEventComments(ctx, proposalEvent.ID)
	if err != nil {
		return models.ProposalEvent{}, err
	}
	proposalEvent.Comments = comments
	transactions, err := p.getProposalEventTransactions(ctx, proposalEvent.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return models.ProposalEvent{}, err
	}
	proposalEvent.Transactions = transactions
	location := models.Address{}
	err = p.DBConnector.DB.
		Where("event_type = ?", models.ProposalEventType).
		Where("event_id = ?", proposalEvent.ID).
		First(&location).
		WithContext(ctx).
		Error

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return models.ProposalEvent{}, err
	}
	proposalEvent.Location = location

	tags := []models.Tag{}
	err = p.DBConnector.DB.
		Where("event_type = ?", models.ProposalEventType).
		Where("event_id = ?", proposalEvent.ID).
		Find(&tags).
		WithContext(ctx).
		Error

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return models.ProposalEvent{}, err
	}
	proposalEvent.Tags = tags

	for i, tag := range proposalEvent.Tags {
		tagValues := []models.TagValue{}
		err = p.DBConnector.DB.Where("tag_id = ?", tag.ID).
			Find(&tagValues).
			WithContext(ctx).
			Error

		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			zlog.Log.Error(err, "got error while getting tag values")
			continue
		}
		proposalEvent.Tags[i].Values = tagValues
		tag.Values = tagValues
	}

	return proposalEvent, nil
}

func NewProposalEvent(DBConnector *Connector) *ProposalEvent {
	return &ProposalEvent{DBConnector: DBConnector}
}
