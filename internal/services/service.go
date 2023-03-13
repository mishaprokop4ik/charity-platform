package service

import (
	"Kurajj/internal/config"
	"Kurajj/internal/repository"
)

type Service struct {
	Authentication  Authenticator
	Admin           AdminCRUDer
	ProposalEvent   ProposalEventer
	Transaction     Transactioner
	Comment         Commenter
	Tag             Tagger
	UserSearchValue UserSearcher
}

func New(repo *repository.Repository,
	authConfig *config.AuthenticationConfig,
	emailConfig *config.Email) *Service {
	return &Service{
		Authentication:  NewAuthentication(repo, authConfig, emailConfig),
		Admin:           NewAdmin(repo, authConfig, emailConfig),
		ProposalEvent:   NewProposalEvent(repo),
		Transaction:     NewTransaction(repo),
		Comment:         NewComment(repo),
		Tag:             NewTag(repo),
		UserSearchValue: NewUserSearch(repo),
	}
}
