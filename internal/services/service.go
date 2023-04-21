package service

import (
	"Kurajj/configs"
	"Kurajj/internal/models"
	"Kurajj/internal/repository"
	"context"
	"io"
)

type HelpEventer interface {
	CreateHelpEvent(ctx context.Context, event *models.HelpEvent) (uint, error)
	GetHelpEventByID(ctx context.Context, id models.ID) (models.HelpEvent, error)
	GetHelpEventByTransactionID(ctx context.Context, transactionID models.ID) (models.HelpEvent, error)
	CreateRequest(ctx context.Context, userID models.ID, transactionInfo models.TransactionAcceptCreateRequest) (uint, error)
	UpdateTransactionStatus(ctx context.Context, transaction models.HelpEventTransaction, file io.Reader, fileType string) error
	GetUserHelpEvents(ctx context.Context, userID models.ID) ([]models.HelpEvent, error)
	GetHelpEventBySearch(ctx context.Context, search models.HelpSearchInternal) (models.HelpEventPagination, error)
	UpdateEvent(ctx context.Context, event models.HelpEvent) error
	GetStatistics(ctx context.Context, fromStart int, creatorID uint) (models.HelpEventStatistics, error)
}

type Service struct {
	Authentication          Authenticator
	Admin                   AdminCRUDer
	ProposalEvent           ProposalEventer
	Transaction             Transactioner
	Comment                 Commenter
	Tag                     Tagger
	UserSearchValue         UserSearcher
	TransactionNotification TransactionNotifier
	HelpEvent               HelpEventer
}

func New(repo *repository.Repository,
	authConfig *configs.AuthenticationConfig,
	emailConfig *configs.Email,
) *Service {
	return &Service{
		Authentication:          NewAuthentication(repo, authConfig, emailConfig),
		Admin:                   NewAdmin(repo, authConfig, emailConfig),
		ProposalEvent:           NewProposalEvent(repo),
		Transaction:             NewTransaction(repo),
		Comment:                 NewComment(repo),
		Tag:                     NewTag(repo),
		UserSearchValue:         NewUserSearch(repo),
		TransactionNotification: NewTransactionNotification(repo),
		HelpEvent:               NewHelpEvent(repo),
	}
}
