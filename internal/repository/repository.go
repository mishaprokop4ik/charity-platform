package repository

import (
	"Kurajj/internal/models"
	"context"
)

type HelpEventer interface {
	CreateEvent(ctx context.Context, event *models.HelpEvent) (uint, error)
	GetEventByID(ctx context.Context, id models.ID) (models.HelpEvent, error)
	GetHelpEventNeeds(ctx context.Context, eventID models.ID) ([]models.Need, error)
	CreateNeed(ctx context.Context, need models.Need) (uint, error)
	UpdateNeeds(ctx context.Context, needs ...models.Need) error
	GetHelpEventByTransactionID(ctx context.Context, transactionID models.ID) (models.HelpEvent, error)
	UpdateHelpEvent(ctx context.Context, event models.HelpEvent) error
	GetUserHelpEvents(ctx context.Context, userID models.ID) ([]models.HelpEvent, error)
	GetEventsWithSearchAndSort(ctx context.Context,
		searchValues models.HelpSearchInternal) (models.HelpEventPagination, error)
}

type Repository struct {
	User                    Userer
	Admin                   adminCRUDer
	ProposalEvent           ProposalEventer
	Transaction             Transactioner
	Comment                 Commenter
	Tag                     Tagger
	UserSearchValue         UserSearcher
	File                    Filer
	TransactionNotification Notifier
	HelpEvent               HelpEventer
}

func New(dbConnector *Connector, config AWSConfig) *Repository {
	return &Repository{
		User:                    NewUser(config, dbConnector),
		Admin:                   NewAdmin(dbConnector),
		ProposalEvent:           NewProposalEvent(config, dbConnector),
		Transaction:             NewTransaction(dbConnector),
		Comment:                 NewComment(dbConnector),
		Tag:                     NewTag(dbConnector),
		UserSearchValue:         NewUserSearch(dbConnector),
		File:                    NewFile(config),
		TransactionNotification: NewTransactionNotification(dbConnector),
		HelpEvent:               NewHelpEvent(config, dbConnector),
	}
}
