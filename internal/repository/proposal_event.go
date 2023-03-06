package repository

import (
	"Kurajj/internal/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/samber/lo"
	"gorm.io/gorm"
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
	query := p.DBConnector.DB.Order(searchValues.SortField).Where("status IN (?)", searchValues.State)
	if searchValues.GetOwn != nil {
		if *searchValues.GetOwn {
			query = query.Where("author_id = ?", searchValues.SearcherID)
		} else {
			query = query.Not("author_id = ?", searchValues.SearcherID)
		}
	}

	if searchValues.Name != nil && *searchValues.Name != "" {
		query = query.Where("title = ?", *searchValues.Name)
	}

	if searchValues.TakingPart != nil {
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

	if searchValues.Tags != nil {
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
	err := query.Find(&events).WithContext(ctx).Error
	return events, err
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

	return newSearchValues
}

func (p *ProposalEvent) CreateEvent(ctx context.Context, event models.ProposalEvent) (uint, error) {
	err := p.DBConnector.DB.
		Create(&event).
		WithContext(ctx).
		Error

	return event.ID, err
}

func (p *ProposalEvent) GetEvent(ctx context.Context, id uint) (models.ProposalEvent, error) {
	event := models.ProposalEvent{}
	resp := p.DBConnector.DB.
		Where("id = ?", id).
		First(&event).
		Where("is_deleted = ?", false).
		WithContext(ctx)

	if errors.Is(resp.Error, gorm.ErrRecordNotFound) {
		return models.ProposalEvent{}, fmt.Errorf("cound not get proposal event by id %d", id)
	}

	return event, resp.Error
}

func (p *ProposalEvent) GetEvents(ctx context.Context) ([]models.ProposalEvent, error) {
	events := []models.ProposalEvent{}
	resp := p.DBConnector.DB.
		Find(&events).
		Where("is_deleted = ?", false).
		WithContext(ctx)

	if errors.Is(resp.Error, gorm.ErrRecordNotFound) {
		return []models.ProposalEvent{}, fmt.Errorf("could not get any proposal events")
	}

	return events, resp.Error
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
		Not("status IN ?", models.Completed, models.Interrupted, models.Canceled).
		Update("status = ?", models.Canceled).
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
		Update("is_deleted = ?", true).
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
		Find(&events).
		Where("author_id = ?", userID).
		Where("is_deleted", false).
		WithContext(ctx)

	if errors.Is(resp.Error, gorm.ErrRecordNotFound) {
		return []models.ProposalEvent{}, fmt.Errorf("could not get any user proposal events")
	}

	return events, resp.Error
}

func NewProposalEvent(DBConnector *Connector) *ProposalEvent {
	return &ProposalEvent{DBConnector: DBConnector}
}
