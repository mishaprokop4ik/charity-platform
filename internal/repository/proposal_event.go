package repository

import (
	"Kurajj/internal/models"
	"context"
	"errors"
	"fmt"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

type ProposalEventer interface {
	proposalEventCRUDer
	GetUserProposalEvents(ctx context.Context, userID uint) ([]models.ProposalEvent, error)
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
	err := p.DBConnector.DB.Delete(&models.ProposalEvent{}).Where("id = ?", id).WithContext(ctx).Error
	return err
}

func (p *ProposalEvent) GetUserProposalEvents(ctx context.Context, userID uint) ([]models.ProposalEvent, error) {
	events := []models.ProposalEvent{}
	resp := p.DBConnector.DB.
		Find(&events).
		Where("author_id = ?", userID).
		WithContext(ctx)

	if errors.Is(resp.Error, gorm.ErrRecordNotFound) {
		return []models.ProposalEvent{}, fmt.Errorf("could not get any user proposal events")
	}

	return events, resp.Error
}

func NewProposalEvent(DBConnector *Connector) *ProposalEvent {
	return &ProposalEvent{DBConnector: DBConnector}
}
