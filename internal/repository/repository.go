package repository

import (
	"Kurajj/internal/models"
	"context"
)

type HelpEventer interface {
	CreateEvent(ctx context.Context, event *models.HelpEvent) (uint, error)
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
	HelpEventer
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
	}
}
